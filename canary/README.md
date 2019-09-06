# co-http/canary

This source provides a small deployment of the co-http app in a canary
structure.

1. `app.yaml`: co-http deployments `main` and `canary` with a single exposed
   service.
2. `servo.yaml`: the servo and auth
3. `debug.yaml`: Ubuntu image for `curl`, `ab`, and other probing inside the
   cluster.

## Usage

Tested on `Kubernetes 1.13.4`

```
kubectl apply -f app.yaml -f servo.yaml
```

### Debugging

I find it helpful to expose the service's port to check on the app:
```
kubectl port-forward service/app-svc 8001:80
```
After deploying the `app.yaml` and forwarding that service port,
`curl 127.0.0.1:8001` should give e.g. `busy for 1033073 us`.
