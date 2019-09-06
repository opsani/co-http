# co-http/canary

This source provides a small deployment of the co-http app in a canary
structure.

1. `debug.yaml`: Ubuntu image for `curl`, `ab`, and other probing inside the
   cluster.
2. `app.yaml`: co-http deployments `main` and `canary` with a single exposed
   service.
3. `servo.yaml`: the servo and auth

## Usage

Tested on `Kubernetes 1.13.4`

```
kubectl apply -f debug.yaml -f app.yaml -f servo.yaml
```
