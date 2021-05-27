# zero-trust

zero-trust is inspired by Microsoft's [Fabrikam Drone Delivery](https://github.com/mspnp/aks-fabrikam-dronedelivery) and Mark Chmarny's [Dapr Hardening Demo](https://github.com/mchmarny/dapr-demos/tree/master/hardened) and serves as an excerise in creating a zero trust deployment of a Dapr solution on Kubernetes.

## Scenario

The business case loosely follow's Microsoft's [Fabrikam Drone Delivery](https://github.com/mspnp/aks-fabrikam-dronedelivery). Zero Trust is a fictional company that provides delivery services. Users are able to schedule, track, and cancel deliveries. When a delivery is requested, the system will determine availability of couriers, determine an effective route, and schedule the delivery. The user may cancel the request as long as the delivery has not be started. While the delivery is in progress, the user can track the ETA of the delivery.

## Deploy

The following are steps to deploy this project:
- Use an existing k8s cluster or [create a new one](https://github.com/jandauz/zero-trust/tree/main/setup/k3d)
- [Setup the cluster](https://github.com/jandauz/zero-trust/tree/main/setup) 