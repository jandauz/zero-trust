DOMAIN				?=example.com
DAPR_RUNTIME		?=latest
DAPR_HA				?=true
DAPR_LOG_AS_JSON	?=true
INGRESS_NAMESPACE	?=traefik

.PHONY: all
all:
	@echo === ACTIVE CONFIGURATION ===
	@echo "DOMAIN			${DOMAIN}"
	@echo "DAPR_RUNTIME:		${DAPR_RUNTIME}"
	@echo "DAPR_HA:		${DAPR_HA}"
	@echo "DAPR_LOG_AS_JSON:	${DAPR_LOG_AS_JSON}"
	@echo "INGRESS_NAMESPACE:	${INGRESS_NAMESPACE}"
	@echo
	@echo "Export these as environment variables to change their values"
	@echo

.PHONY: certs
certs: ## Created wildcard TLS certificates using mkcert
	# Removing existing certificates ...
	rm -rf certs/$(DOMAIN)
	# Generating new certificates ...
	mkdir -p certs/$(DOMAIN)
	mkcert $(DOMAIN) "*.$(DOMAIN)"
	chmod 644 *.pem
	mv *.pem certs/$(DOMAIN)

.PHONY: dapr
dapr: dapr-install ## Install Dapr

.PHONY: dapr-install
dapr-install: ## Install and configure Dapr
	# Installing Dapr ...
	dapr init -k \
		--runtime-version $(DAPR_RUNTIME) \
		--log-as-json $(DAPR_LOG_AS_JSON) \
		--enable-ha $(DAPR_HA)
	# Waiting for everything to finish ...
	kubectl rollout status deployment/dapr-operator -n dapr-system
	kubectl rollout status deployment/dapr-dashboard -n dapr-system
	kubectl rollout status deployment/dapr-sentry -n dapr-system
	kubectl rollout status deployment/dapr-sidecar-injector -n dapr-system

.PHONY: ingress
ingress: ## Install and configure Traefik ingress
	# Add and update helm repo ...
	helm repo add traefik https://helm.traefik.io/traefik
	helm repo update
	# Configure traefik namespace ...
	sed "s/NSNAME/traefik/g" config/namespace-template.yaml > config/namespace.yaml
	kubectl apply -f config/namespace.yaml
	# Creating Dapr API token secret ...
	$(eval API_TOKEN=$(shell openssl rand -base64 32))
	kubectl create secret generic dapr-api-token --from-literal=token="$(API_TOKEN)" -n $(INGRESS_NAMESPACE)
	# Apply ingress config ...
	kubectl apply -f config/ingress-config.yaml -n $(INGRESS_NAMESPACE)
	# Installing Traefik ...
	helm install traefik traefik/traefik \
			--set deployment.replicas=2 \
			-f config/ingress-annotations.yaml \
			-n $(INGRESS_NAMESPACE)
	kubectl rollout status deployment/traefik -n $(INGRESS_NAMESPACE)
	# Creating cert secret ...
	kubectl create secret tls tls-secret \
		--key certs/$(DOMAIN)/$(DOMAIN)+1-key.pem \
		--cert certs/$(DOMAIN)/$(DOMAIN)+1.pem \
		-n $(INGRESS_NAMESPACE)
	# Apply IngressRoute ...
	sed "s/DOMAINNAME/$(DOMAIN)/g" config/ingress-template.yaml > config/ingress.yaml
	kubectl apply -f config/ingress.yaml -n $(INGRESS_NAMESPACE)

.PHONY: test
test: ## Test deployment and send Dapr API health check
	curl -v \
		-H "Content-Type: application/json" \
		"https://$(DOMAIN)/v1.0/healthz"
	@echo
	@echo === DNS CHECK ===
	@echo Ensure server certificate has:
	@echo
	@echo " subjectAltName: host '$(DOMAIN)' matched cert's '$(DOMAIN)'"
	@echo " SSL certificate verify ok"
	@echo
	@echo And that the response status code from Dapr health check was '204'
	@echo
	@echo " HTTP/2 204"
	@echo

.PHONY: token
token: ## Print Dapr API token
	$(eval API_TOKEN=$(shell kubectl get secret dapr-api-token -n $(INGRESS_NAMESPACE) -o jsonpath="{.data.token}" | base64 --decode))
	@echo
	@echo Dapr API token is:
	@echo
	@echo " ${API_TOKEN}"
	@echo

.PHONY: reload
reload: ## Reloads API to pick up new components
	kubectl rollout restart deployment/traefik -n traefik
	kubectl rollout status deployment/traefik -n traefik

.PHONY: redis
redis: ## Install Redis into the cluster
	# Add and update helm repo ...
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	kubectl create ns redis
	# Installing redis ...
	helm install redis bitnami/redis -n redis
	kubectl rollout status statefulset.apps/redis-master -n redis
	kubectl rollout status statefulset.apps/redis-replicas -n redis

.PHONY: mongo
mongo: ## Install MongoDB into the cluster
	# Add and update helm repo ...
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	kubectl create ns mongo
	# Installing mongo ...
	helm install mongo bitnami/mongodb \
		--set architecture=replicaset \
		--set auth.username=dapr \
		--set auth.database=dapr \
		--set replicaCount=3 \
		-n mongo
	kubectl rollout status statefulset.apps/mongo-mongodb -n mongo
	kubectl rollout status statefulset.apps/mongo-mongodb-arbiter -n mongo

.PHONY: namespace
namespace: ## Configure namespace (make namespace NSNAME=default)
	# Creating zero-trust namespace
	sed "s/NSNAME/${NSNAME}/g" config/namespace-template.yaml > config/namespace.yaml
	kubectl apply -f config/namespace.yaml
	$(eval REDIS_PASSWORD=$(shell kubectl get secret redis -n redis -o jsonpath="{.data.redis-password}" | base64 --decode))
	kubectl create secret generic redis --from-literal=redis-password="$(REDIS_PASSWORD)" -n ${NSNAME}
	export MONGODB_PASSWORD=$(shell kubectl get secret --namespace mongo mongo-mongodb -o jsonpath="{.data.mongodb-password}" | base64 --decode)
	kubectl create secret generic mongo-mongodb --from-literal=mongodb-password="$(MONGODB_PASSWORD)" -n ${NSNAME}
.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
