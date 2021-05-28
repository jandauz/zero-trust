RELEASE_VERSION	=v0.1.0
SERVICE_NAME	?=ingest
DOCKER_REGISTRY	?=localhost:12345

.PHONY: all
all: help

.PHONY: tidy
tidy: ## Updates the go modules and vendors all dependencies
	go mod tidy
	go mod vendor

.PHONY: k3d-registry
k3d-registry: ## Prints the k3d registry
	$(eval K3D_REGISTRY_PORT=$(shell docker inspect -f='{{(index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort}}' $(shell docker ps -aqf "name=k3d-zero-trust-registry")))
	@echo
	@echo k3d-registry is:
	@echo
	@echo " localhost:${K3D_REGISTRY_PORT}"
	@echo

.PHONY: image
image: tidy ## Builds and publishes docker image
	docker build -t "$(DOCKER_REGISTRY)/$(SERVICE_NAME):$(RELEASE_VERSION)" .
	docker push "$(DOCKER_REGISTRY)/$(SERVICE_NAME):$(RELEASE_VERSION)"

.PHONY: deploy
deploy: ## Deploys prebuilt image to k8s using currently selected context
	kubectl apply -f ../deploy/pubsub/redis.yaml
	kubectl apply -f ../deploy/$(SERVICE_NAME).yaml
	kubectl rollout restart deploy $(SERVICE_NAME) -n zero-trust
	kubectl rollout status deploy $(SERVICE_NAME) -n zero-trust

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'