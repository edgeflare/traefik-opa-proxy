# Traefik-OPA-Proxy

[Traefik `forwardAuth` middleware](https://doc.traefik.io/traefik/middlewares/http/forwardauth/) interprets 2xx response code from the auth service as an "authorization successful". Otherwise, the response from the authentication server is returned.

[Open Policy Agent (OPA)](https://www.openpolicyagent.org/) returns a 200 OK with the payload `{"allow": false}` for "authorization failed", meaning Traefik always allows client's requests even if they should be blocked.

This `traefik-opa-proxy` translates OPA's decisions into HTTP status codes: a 403 Forbidden for `{"allow": false}` and a 200 OK for `{"allow": true}`. Use this service with Traefik forwardAuth middleware instead of connecting directly to OPA. The payload sent from Traefik to OPA matches the format expected by the [OPA-Envoy plugin](https://github.com/open-policy-agent/opa-envoy-plugin), so the same policies should work with Envoy based proxies, e.g., Istio and Gloo without modification.

> ## UPDATE: This repo is archived in favor of https://github.com/edgeflare/traefikopa. It can be useful when Traefik installation can't be modified with plugin or you don't need, for example, request body for OPA policy evaluation.

## Test locally

Start the proxy in a terminal window

```sh
go mod tidy
OPA_URL=http://localhost:8181/v1/data/httpapi/authz go run .
```

In another terminal start opa

```sh
opa run --server --log-level=debug --bundle ./example
```

In a third terminal make a few HTTP requests. The responses should conform to [demo authorization policy](example/demo-authz.rego)

```sh
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:8182
# 403
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:8182/allowed
# 200
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:8182/allowed -X POST
# 403
```


## Test on Kubernetes

```sh
opa build example/demo-authz.rego
kubectl -n kube-system create configmap demo-authz-policy --from-file=bundle.tar.gz
kubectl apply -f ./example
```

See example directory for more.
