# k3d cluster setup

The following parameters can be used to configure your deployment. Define these as environment variables to set or override the default value:
```shell
CLUSTER_NAME    # default: demo
NODE_COUNT      # default: 3
```

## Prerequisites
- [k3d](https://k3d.io/#install-current-latest-release)

## Usage
Run `make` by itself to see the active configuration.

- `make cluster` to create a k3d cluster (make cluster CLUSTER_NAME=<cluster_name>)
- `make cluster-list` to list k3d clusters
- `make cluster-start` to start a previously stopped k3d cluster
- `make cluster-stop` to stop a started k3d cluster

> Note: It is recommended to shutdown current cluster before creating a new one using `make cluster-stop CLUSTER_NAME=<cluster_name>` as the exposed ports would conflict.

## Cleanup
To list previously created clusters:
```shell
make cluster-list
```

To delete any of the previously created clusters:
```shell
make cluster-down CLUSTER_NAME=<cluster_name>
```
