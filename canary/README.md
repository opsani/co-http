# co-http/canary

This source provides a small deployment of the co-http app in a canary
structure.

1. `app.yaml`: co-http deployments `main` and `canary` with a single exposed
   service.
2. `servo.yaml`: the servo and auth
3. `monitoring.yaml`: Prometheus and Grafana
4. `debug.yaml`: Ubuntu image for `curl`, `ab`, and other probing inside the
   cluster.

## Usage

To deploy the canary test app, start up your favorite Kubernetes cluster
(recommended 3GiB Memory, 3 CPU) and execute the following command:

```
kubectl apply -f app.yaml -f servo.yaml -f monitoring.yaml
```

Tested on `Kubernetes 1.13.4` and `Kubernetes 1.15.2`

## Debugging

This section describes some useful ways to answer the questions "Is it up and
running?" and "

### Access the deployed app via localhost:8001
I find it helpful to expose the service's port to check on the app:
```
kubectl port-forward service/app 8001:80
```
After deploying the `app.yaml` and forwarding that service port,
`curl 127.0.0.1:8001` should give e.g. `busy for 1033073 us`.

### Access Prometheus via localhost:9090

First, run the following command:
```
kubectl port-forward svc/prometheus 9090
```

You should see e.g. this output:
```
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Then just point your brower at `localhost:9090` to begin querying the Prometheus
database.

### Run custom Docker images
This might be useful if you want to change some internal functionality of the
testing app, use your own app, etc.

If you want to use images in your own Docker Image Repository:
```
minikube delete
minikube start
minikube addons configure registry-creds
minikube addons enable registry-creds
```
And before you deploy, add the following tag (e.g. for ECR) to your container
`spec:` section in your `deployment.yaml`.
```
imagePullSecrets:
- name: awsecr-cred
```

## To-Do
1. Strip down the bloated `monitor.yaml`
2. Reintroduce Grafana support
3. Rename Prometheus jobs to be properly informative, audit these jobs for
   relevance.
