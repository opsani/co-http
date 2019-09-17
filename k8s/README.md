# co-http/k8s

This repository contains the Kubernetes config files required to deploy the
testing app used by Optune. If you have a Kubernetes cluster running, just
`kubectl apply` each file in this repository with the typical secrets
modifications to deploy the app. If you don't have a Kubernetes cluster running,
the next section has a walkthrough on how to deploy using AWS EKS.

## Preliminaries

Install the following tools:
1. `kubectl`
2. `kubectx`
3. [`k9s`][k9s]
4. [`eksctl`][eksctl]

## Spin up a cluster

Use `eksctl` to spin up an EKS cluster: `eksctl create cluster --name testing`

## Connect to your cluster

1. `kubectx` should display the cluster name, highlighted.
2. `k9s` should start the command-line GUI you will use to monitor the cluster.

## Deploy the 3-component app

First, modify the `#changeme` tagged entries in `servo.yaml` and
`optune-auth.yaml`.

Next, execute the following command to deploy the contents of the denoted files
to the Kubernetes cluster you set up:

```
kubectl apply -f c1.yaml -f c2.yaml -f c3.yaml \
  -f optune-auth.yaml -f rbac.yaml -f servo.yaml
```

- `cX.yaml` files describe `co-http` app components.
- `optune-auth.yaml` has secrets.
- `rbac.yaml` allows servos in `Kubernetes 1.13+` to perform their cluster-level
  responsibilities.
- `servo.yaml` is the Optune servo that measures and adjusts the `co-http` app
  using guidance from the Optune.ai server.

## Appendix

Tested on `Kubernetes 1.15.2`.

[k9s]: https://github.com/derailed/k9s
[eksctl]: https://eksctl.io/introduction/installation/
