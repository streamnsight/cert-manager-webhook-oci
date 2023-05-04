package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/dns"
)

// GroupName is used to identify the company or business unit that created this webhook.
// This name will need to be referenced in each Issuer's `webhook` stanza to inform
// cert-manager of where to send ChallengePayload resources in order to solve the DNS01 challenge.
// This group name should be **unique**, hence using your own company's domain here is recommended.
var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName,
		&ociDNSProviderSolver{},
	)
}

// ociDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type ociDNSProviderSolver struct {
	// If a Kubernetes 'clientset' is needed, you must:
	// 1. uncomment the additional `client` field in this structure below
	// 2. uncomment the "k8s.io/client-go/kubernetes" import at the top of the file
	// 3. uncomment the relevant code in the Initialize method below
	// 4. ensure your webhook's service account has the required RBAC role
	//    assigned to it for interacting with the Kubernetes APIs you need.
	client *kubernetes.Clientset
}

// ociDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type ociDNSProviderConfig struct {
	// Change the two fields below according to the format of the configuration
	// to be decoded.
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.

	CompartmentOCID     string `json:"compartmentOCID"`
	OCIProfileSecretRef string `json:"ociProfileSecretName"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.
func (c *ociDNSProviderSolver) Name() string {
	return "oci"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *ociDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	klog.V(6).Infof("call function Present: namespace=%s, zone=%s, fqdn=%s", ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return fmt.Errorf("unable to load config: %v", err)
	}

	ociDNSClient, err := c.ociDNSClient(&cfg, ch.ResourceNamespace)
	if err != nil {
		return fmt.Errorf("unable to initialize ociDNSClient: %v", err)
	}

	ctx := context.Background()

	_, err = ociDNSClient.PatchZoneRecords(ctx, patchRequest(&cfg, ch, dns.RecordOperationOperationAdd))
	if err != nil {
		return fmt.Errorf("can not create TXT record: %v", err)
	}
	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *ociDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	klog.V(6).Infof("call function CleanUp: namespace=%s, zone=%s, fqdn=%s", ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return fmt.Errorf("unable to load config: %v", err)
	}

	ociDNSClient, err := c.ociDNSClient(&cfg, ch.ResourceNamespace)
	if err != nil {
		return fmt.Errorf("unable to initialize ociDNSClient: %v", err)
	}

	ctx := context.Background()

	_, err = ociDNSClient.PatchZoneRecords(ctx, patchRequest(&cfg, ch, dns.RecordOperationOperationRemove))
	if err != nil {
		return fmt.Errorf("can not delete TXT record: %v", err)
	}
	return nil
}

func patchRequest(cfg *ociDNSProviderConfig, ch *v1alpha1.ChallengeRequest, operation dns.RecordOperationOperationEnum) dns.PatchZoneRecordsRequest {
	domain := strings.TrimSuffix(ch.ResolvedFQDN, ".")
	rtype := "TXT"
	ttl := 60

	return dns.PatchZoneRecordsRequest{
		CompartmentId: &cfg.CompartmentOCID,
		ZoneNameOrId:  &ch.ResolvedZone,

		PatchZoneRecordsDetails: dns.PatchZoneRecordsDetails{
			Items: []dns.RecordOperation{
				dns.RecordOperation{
					Domain:    &domain,
					Rtype:     &rtype,
					Rdata:     &ch.Key,
					Ttl:       &ttl,
					Operation: operation,
				},
			},
		},
		RequestMetadata: getRequestMetadataWithDefaultRetryPolicy(),
	}
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *ociDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	k8sClient, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.client = k8sClient

	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (ociDNSProviderConfig, error) {
	cfg := ociDNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("cannot unmarshal raw JSON: %v", err)
	}

	return cfg, nil
}

// ociDNSClient is a helper function to initialize a DNS client from the oci-sdk
func (c *ociDNSProviderSolver) ociDNSClient(cfg *ociDNSProviderConfig, namespace string) (*dns.DnsClient, error) {
	secretName := cfg.OCIProfileSecretRef
	klog.V(6).Infof("Trying to load oci profile from secret `%s` in namespace `%s`", secretName, namespace)
	ctx := context.Background()

	// Need to handle Instance Principal case
	var userConfigIsValid bool
	var userConfigErrorMsg string
	userConfigIsValid = true
	userConfigErrorMsg = ""
	sec, err := c.client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})

	if err != nil {
		userConfigIsValid = false
		userConfigErrorMsg = fmt.Sprintf("%s\nunable to get secret `%s/%s`; %v", userConfigErrorMsg, secretName, namespace, err)
	}

	tenancy, err := stringFromSecretData(&sec.Data, "tenancy")
	if err != nil {
		userConfigIsValid = false
		userConfigErrorMsg = fmt.Sprintf("%s\nunable to get tenancy from secret `%s/%s`; %v", userConfigErrorMsg, secretName, namespace, err)
	}

	user, err := stringFromSecretData(&sec.Data, "user")
	if err != nil {
		userConfigIsValid = false
		userConfigErrorMsg = fmt.Sprintf("%s\nunable to get user from secret `%s/%s`; %v", userConfigErrorMsg, secretName, namespace, err)
	}

	region, err := stringFromSecretData(&sec.Data, "region")
	if err != nil {
		userConfigIsValid = false
		userConfigErrorMsg = fmt.Sprintf("%s\nunable to get region from secret `%s/%s`; %v", userConfigErrorMsg, secretName, namespace, err)
	}

	fingerprint, err := stringFromSecretData(&sec.Data, "fingerprint")
	if err != nil {
		userConfigIsValid = false
		userConfigErrorMsg = fmt.Sprintf("%s\nunable to get fingerprint from secret `%s/%s`; %v", userConfigErrorMsg, secretName, namespace, err)
	}

	privateKey, err := stringFromSecretData(&sec.Data, "privateKey")
	if err != nil {
		userConfigIsValid = false
		userConfigErrorMsg = fmt.Sprintf("%s\nunable to get privateKey from secret `%s/%s`; %v", userConfigErrorMsg, secretName, namespace, err)
	}

	privateKeyPassphrase, err := stringFromSecretData(&sec.Data, "privateKeyPassphrase")
	if err != nil {
		userConfigIsValid = false
		userConfigErrorMsg = fmt.Sprintf("%s\nunable to get privateKeyPassphrase from secret `%s/%s`; %v", userConfigErrorMsg, secretName, namespace, err)
	}

	var configProvider common.ConfigurationProvider
	if os.Getenv("OCI_USE_WORKLOAD_IDENTITY") == "true" {
		configProvider, err = auth.OkeWorkloadIdentityConfigurationProvider()
		if err != nil {
			return nil, fmt.Errorf("unable to authenticate with Workload Identity; %v", err)
		}
	} else if userConfigIsValid {
		configProvider = common.NewRawConfigurationProvider(tenancy, user, region, fingerprint, privateKey, &privateKeyPassphrase)
	} else {
		klog.V(6).Infof("user config not valid: %s", userConfigErrorMsg)
		klog.V(6).Infof("trying Instance Principal auth")
		configProvider, err = auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, fmt.Errorf("unable to authenticate with Instance Principal; %v", err)
		}
	}

	dnsClient, err := dns.NewDnsClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &dnsClient, nil
}

func stringFromSecretData(secretData *map[string][]byte, key string) (string, error) {
	bytes, ok := (*secretData)[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret data", key)
	}
	return string(bytes), nil
}

func getRequestMetadataWithDefaultRetryPolicy() common.RequestMetadata {
	return common.RequestMetadata{
		RetryPolicy: getDefaultRetryPolicy(),
	}
}

func getDefaultRetryPolicy() *common.RetryPolicy {
	// how many times to do the retry
	attempts := uint(10)

	// retry for all non-200 status code
	retryOnAllNon200ResponseCodes := func(r common.OCIOperationResponse) bool {
		response := r.Response.HTTPResponse()
		retry := !((r.Error == nil && 199 < response.StatusCode && response.StatusCode < 300) || response.StatusCode == 401)
		if retry {
			klog.V(6).Infof("request %s %s responded %s; retrying...", response.Request.Method, response.Request.URL.String(), response.Status)
		}
		return retry
	}
	return getExponentialBackoffRetryPolicy(attempts, retryOnAllNon200ResponseCodes)
}

func getExponentialBackoffRetryPolicy(n uint, fn func(r common.OCIOperationResponse) bool) *common.RetryPolicy {
	// the duration between each retry operation, you might want to waite longer each time the retry fails
	exponentialBackoff := func(r common.OCIOperationResponse) time.Duration {
		response := r.Response.HTTPResponse()
		duration := time.Duration(math.Pow(float64(2), float64(r.AttemptNumber-1))) * time.Second
		klog.V(6).Infof("backing off %s to retry %s %s after %d attempts", duration, response.Request.Method, response.Request.URL.String(), r.AttemptNumber)
		return duration
	}
	policy := common.NewRetryPolicy(n, fn, exponentialBackoff)
	return &policy
}
