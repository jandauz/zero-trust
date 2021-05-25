CLUSTER_NAME		?=demo
NODE_COUNT			?=3

.PHONY: all
all:
		@echo === ACTIVE CONFIGURATION ===
		@echo "CLUSTER_NAME:		${CLUSTER_NAME}"
		@echo "NODE_COUNT:		${NODE_COUNT}"
		@echo
		@echo "Export these as environment variables to change their values"
		@echo


.PHONY: cluster
cluster: ## Create k3d cluster
		k3d cluster create $(CLUSTER_NAME) \
				--api-port 6550 \
				-p "80:80@loadbalancer" \
				-p "443:443@loadbalancer" \
				--agents $(NODE_COUNT) \
				--registry-create \
				--k3s-server-arg "--no-deploy=traefik"
		kubectl config use-context k3d-$(CLUSTER_NAME)

.PHONY: cluster-list
cluster-list: ## List k3d clusters
		k3d cluster list

.PHONY: cluster-start
cluster-start: ## Start k3d cluster
		k3d cluster start $(CLUSTER_NAME)

.PHONY: cluster-stop
cluster-stop: ## Stop k3d cluster
		k3d cluster stop $(CLUSTER_NAME)

.PHONY: cluster-down
cluster-down: ## Delete k3d cluster
		k3d cluster delete $(CLUSTER_NAME)

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'