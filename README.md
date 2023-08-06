# ACME webhook for Oracle Cloud Infrastructure

This solver can be used when you want to use cert-manager with Oracle Cloud Infrastructure as a DNS provider.

## Changes from upstream

The [original code](https://gitlab.com/dn13/cert-manager-webhook-oci) is not under active maintenance, these are some minor fixes to make it compatible with recent Kubernetes and Cert Manager:

* [Fixes by Tyler Lawson](https://gitlab.com/lawsontyler/cert-manager-webhook-oci) to support newer OCI APIs
* Fixes to RBAC to support newer versions of Kubernetes
* GitHub actions to build the container image and Helm repo.
* Update to v65 OCI SDK, and implement Instance Principal auth

## Requirements

- [go](https://golang.org/) >= 1.13.0 *only for development*
- [helm](https://helm.sh/) >= v3.0.0
- [kubernetes](https://kubernetes.io/) >= v1.14.0
- [cert-manager](https://cert-manager.io/) >= 1.0

## Installation

### cert-manager

Follow the [instructions](https://cert-manager.io/docs/installation/) using the cert-manager documentation to install it within your cluster.

### Webhook

<!-- #### Using Public Helm Chart

```bash
helm repo add cert-manager-webhook-oci https://streamnsight.github.io/cert-manager-webhook-oci
helm install --namespace cert-manager cert-manager-webhook-oci cert-manager-webhook-oci/cert-manager-webhook-oci
``` -->

#### From local checkout

Build and publish the container image to your registry

Install:

```bash
helm install --namespace cert-manager \
--set image.repository=ghcr.io/giovannicandido/cert-manager-webhook-oci \
--set image.tag=build-pipeline \
--set groupName=acme.youcompany.com \
 cert-manager-webhook-oci deploy/cert-manager-webhook-oci
```

**Note**: The kubernetes resources used to install the Webhook should be deployed within the same namespace as the cert-manager.

To uninstall the webhook run
```bash
helm uninstall --namespace cert-manager cert-manager-webhook-oci
```

## Issuer

Create a `ClusterIssuer` or `Issuer` resource as following:
```yaml
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: mail@example.com # REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
      - dns01:
          webhook:
            groupName: acme.yourcompany.com
            solverName: oci
            config:
              ociProfileSecretName: oci-profile
              compartmentOCID: ocid-of-compartment-to-use
```

### Credentials
In order to access the Oracle Cloud Infrastructure API, the webhook needs an OCI profile configuration.

If you choose another name for the secret than `oci-profile`, ensure you modify the value of `ociProfileSecretName` in the `[Cluster]Issuer`.

The secret for the example above will look like this:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oci-profile
type: Opaque
stringData:
  tenancy: "your tenancy ocid"
  user: "your user ocid"
  region: "your region"
  fingerprint: "your key fingerprint"
  privateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    ...KEY DATA HERE...
    -----END RSA PRIVATE KEY-----
  privateKeyPassphrase: "private keys passphrase or empty string if none"
```

### Create a certificate

Finally you can create certificates, for example:

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: example-cert
  namespace: cert-manager
spec:
  commonName: example.com
  dnsNames:
    - example.com
  issuerRef:
    name: letsencrypt-staging
    kind: ClusterIssuer
  secretName: example-cert
```

## Development

### Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating a
DNS01 webhook.**

First, create an oracle cloud infrastructure account and ensure you have a DNS zone set up.
Next, create config files based on the `*.sample` files in the `testdata/oci` directory.

You can then run the test suite with:

```bash
# first install necessary binaries (only required once)
./scripts/fetch-test-binaries.sh
# then run the tests
TEST_ZONE_NAME=example.com. make verify
```


