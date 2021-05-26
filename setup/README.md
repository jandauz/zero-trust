# k8s cluster setup

An exercise in creating a zero trust deployment of Dapr on Kubernetes, configured with:
- Ingress
  - [Traefik Proxy](https://traefik.io/traefik/) for ingress controller and TLS to service mapping

## Prequisites
- 1.15+ k8s cluster. The following provides easy installation on:
  - [k3d](https://github.com/jandauz/zero-trust/tree/main/setup/k3d)
- Tools:
  - [kubectl](https://kubernetes.io/docs/tasks/tools/) to interact with the k8s cluster
  - [Helm](https://helm.sh/docs/intro/install/) to install Dapr and dependencies
  - [mkcert](https://github.com/FiloSottile/mkcert#installation) to generate wildcard certificates

## Setup
The following parameters can be used to configure your deployment. Define these as environment variables to set or override the default value:
```shell
DOMAIN            # default: example.com
DAPR_HA           # default: true
DAPR_LOG_AS_JSON  # default: true
INGRESS_NAMESPACE # default: traefik
```
> Note: This requires an existing k8s cluster with the current context (`kubectl config current-context`) set to the desired k8s cluster. List all registered contexts using `kubectl config get-contexts` and set the desired context using `kubectl config use-context <context>` if needed.

## Usage
Run `make` to display active configuration and adjust as necessary.

To deploy and configure Dapr:
- `make dapr` to install Dapr

To configure external access:
- `make certs` to create TLS certs
- `make ingress` to configure Traefik ingress (read [here](https://github.com/jandauz/zero-trust/tree/main/setup/docs/traefik.md) for more information)
- `make test` to test deployment