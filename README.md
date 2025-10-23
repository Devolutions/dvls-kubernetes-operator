# dvls-kubernetes-operator
:warning: **This operator is a work in progress, expect breaking changes between releases** :warning:

Operator to sync Devolutions Server `Credential Entry` entries as Kubernetes Secrets

## Description
This operator uses the defined custom resource DvlsSecret which manages its own Kubernetes Secret and will keep itself up to date at a defined interval (every minute by default).
The Docker image can be found [here](https://hub.docker.com/r/devolutions/dvls-kubernetes-operator).

### Operator configuration
The following Environment Variables can be used to configure the operator :
- `DEVO_OPERATOR_DVLS_BASEURI` (required) - DVLS instance base URI
- `DEVO_OPERATOR_DVLS_APPID` (required) - DVLS Application ID
- `DEVO_OPERATOR_DVLS_APPSECRET` (required) - DVLS Application Secret
- `DEVO_OPERATOR_REQUEUE_DURATION` (optional) - Entry/Secret resync interval (default 60s). Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
- `SSL_CERT_FILE` (optional) - Path to a custom CA certificate file for DVLS servers with self-signed certificates. This is automatically set by the Helm chart when `instanceSecret.caCert` is provided.

A sample of the custom resource can be found [here](https://github.com/Devolutions/dvls-kubernetes-operator/blob/master/config/samples/dvls_v1alpha1_dvlssecret.yaml).
The entry ID can be fetched by going in the entry properties, `Advanced -> Session ID`.

### Devolutions Server configuration
We recommend creating an [Application ID](https://helpserver.devolutions.net/webinterface_applications.html?q=application) specifically to be used with the Operator that has [minimal access to a vault](https://helpserver.devolutions.net/vaults_applications.html?q=application) that only contains the secrets to be synchronized.

Only `Credential Entry` entries are supported at the moment. The available entry data will depend on the `Credential Entry` type.

### Kubernetes configuration
Since this operator uses Kubernetes Secrets, it is recommended that you follow [best practices](https://kubernetes.io/docs/concepts/security/secrets-good-practices/) surrounding secrets, especially [encryption at rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/).

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Helm Chart
A Helm Chart is available to simplify installation. Add the Devolutions Helm chart repository, create a `values.yaml` from [the default values](https://github.com/Devolutions/dvls-kubernetes-operator/blob/master/chart/values.yaml) as a baseline, and update values as necessary.

#### Required Configuration
The following values **must** be configured in your `values.yaml`:
- `controllerManager.manager.env.devoOperatorDvlsBaseuri` - Your DVLS server URL (e.g., `https://dvls.example.com`)
- `controllerManager.manager.env.devoOperatorDvlsAppid` - Application ID from your DVLS server
- `instanceSecret.secret` - Application Secret from your DVLS server

#### Optional Configuration
- `instanceSecret.caCert` - Custom CA certificate for self-signed DVLS servers (see below)
- `controllerManager.manager.env.devoOperatorRequeueDuration` - How often to sync secrets (default: `60s`)

#### Basic Example Configuration

Create a `values.yaml` file with your DVLS configuration:

```yaml
controllerManager:
  manager:
    env:
      devoOperatorDvlsAppid: "00000000-0000-0000-0000-000000000000"
      devoOperatorDvlsBaseuri: "https://dvls.example.com"
      devoOperatorRequeueDuration: "60s"

instanceSecret:
  secret: "your-app-secret-here"
```

#### Installation
```sh
helm repo add devolutions-helm-charts https://devolutions.github.io/helm-charts
helm repo update
helm install dvls-kubernetes-operator devolutions-helm-charts/dvls-kubernetes-operator --values values.yaml
```

#### Using a Custom CA Certificate

If your DVLS server uses a self-signed certificate (common in test/development environments), you need to provide the CA certificate so the operator can establish a trusted TLS connection.

**When to use this:**
- Testing with self-signed certificates
- Internal CA certificates not in the system trust store
- Development/staging environments with custom PKI

**Configuration:**

Add the CA certificate content to your `values.yaml`:

```yaml
controllerManager:
  manager:
    env:
      devoOperatorDvlsAppid: "00000000-0000-0000-0000-000000000000"
      devoOperatorDvlsBaseuri: "https://dvls.example.com"
      devoOperatorRequeueDuration: "60s"

instanceSecret:
  secret: "your-app-secret"
  # Add your CA certificate here (PEM format)
  caCert: |
    -----BEGIN CERTIFICATE-----
    MIIDXTCCAkWgAwIBAgIJAKZ...
    (your CA certificate content)
    ...
    -----END CERTIFICATE-----
```

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/dvls-kubernetes-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/dvls-kubernetes-operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

