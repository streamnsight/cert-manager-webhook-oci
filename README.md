# ACME webhook for Oracle Cloud Infrastructure

This webhook lets `cert-manager` solve `dns01` challenges using Oracle Cloud Infrastructure DNS.

## Current target

- `cert-manager` `1.19.1`
- Oracle OCI Go SDK `v65.113.0`
- Kubernetes Go libraries `v0.34.1`
- Go `1.25`
- Oracle Linux `10` runtime image

The solver config schema and Helm values are intentionally unchanged from earlier releases:

- `compartmentOCID`
- `ociProfileSecretName`
- `groupName`
- `certManager.namespace`
- `certManager.serviceAccountName`
- `useWorkloadIdentity`
- `region`
- `ociProfileSecretNames`

## Installation

Install `cert-manager` separately using its upstream installation documentation.

### Helm chart

The chart defaults to the published OCI webhook image:

```bash
helm upgrade --install cert-manager-webhook-oci deploy/cert-manager-webhook-oci \
  --namespace cert-manager \
  --create-namespace \
  --set groupName=acme.yourcompany.com
```

If you build and publish your own image, override the image coordinates:

```bash
helm upgrade --install cert-manager-webhook-oci deploy/cert-manager-webhook-oci \
  --namespace cert-manager \
  --create-namespace \
  --set groupName=acme.yourcompany.com \
  --set image.repository=ghcr.io/your-org/cert-manager-webhook-oci \
  --set image.tag=ol10-cm1.19.1
```

To enable OKE Workload Identity, add:

```bash
--set useWorkloadIdentity=true \
--set region=us-ashburn-1
```

The webhook must be deployed into the same namespace as `cert-manager`.

### Static manifest

A rendered manifest for this target is published under:

```text
deploy/v1.19.1-1.5.0/cert-manager-webhook-oci.yaml
```

## Issuer example

Create an `Issuer` or `ClusterIssuer` using the stable `cert-manager.io/v1` API:

```yaml
apiVersion: cert-manager.io/v1
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
              compartmentOCID: ocid1.compartment.oc1..exampleuniqueID
```

## Authentication modes

The webhook supports three OCI authentication modes:

- OCI profile secret
- Instance Principal
- OKE Workload Identity

### OCI profile secret

If you keep the default secret name, create the profile secret as:

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
  privateKeyPassphrase: ""
```

The OCI principal needs at least:

```text
use dns-zones in compartment
manage dns-records in compartment
```

#### Using Instance Principal (least secure)

- Using Network Source: Limit the scope to a subnet (more secure)

In Identity, Network Source, create a Network Source targetting the VCN where the cluster is deployed, and the CIDR block of the nodepool

Create a Policy reading:

```
allow any-user to use dns-zones in compartment <compartment-name> where ALL {request.networkSource.name = '<network-source-name>'}
allow any-user to manage dns-records in compartment <compartment-name> where ALL {request.networkSource.name = '<network-source-name>'}
```

- Using a Dynamic Group: Limit the scope to a compartment (less secure)

Create 

#### Using Workload Identity (most secure)

Only available on OKE Enhanced Clusters

```
allow any-user to use dns-zones in compartment <compartment-name> where ALL {request.principal.type='workload', request.principal.namespace ='cert-manager', request.principal.service_account = 'cert-manager-webhook-oci', request.principal.cluster_id = '<oke-cluster-id>'}
allow any-user to manage dns-record in compartment <compartment-name> where ALL {request.principal.type='workload', request.principal.namespace ='cert-manager', request.principal.service_account = 'cert-manager-webhook-oci', request.principal.cluster_id = '<oke-cluster-id>'}
```

## Certificate example

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-cert
  namespace: cert-manager
spec:
  secretName: example-cert
  commonName: example.com
  dnsNames:
    - example.com
  issuerRef:
    name: letsencrypt-staging
    kind: ClusterIssuer
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

`fetch-test-binaries.sh` uses `setup-envtest` and defaults to Kubernetes `1.34.x`. Override that with `K8S_VERSION=<minor-or-patch>` if you need a different supported test control-plane version.

## Upgrade notes

- Issuer and Certificate examples now use `cert-manager.io/v1`.
- Existing solver config keys and Helm values are unchanged.
- The default image tag moved from the old OL7 / cert-manager 1.6 line to `ol10-cm1.19.1`.
- Legacy rendered manifests under `deploy/v1.6.1*` are kept for reference while the new `v1.19.1-1.5.0` manifest becomes the current target.
