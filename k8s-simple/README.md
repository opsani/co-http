# Deploy a simple Single Tier (single service)

* Single application (using Opsani co-http app)
* Saturation style load generation using Vegeta load generator
* Prometheus for metrics (using the provided Prometheus manifests)

## Coctl - manage config overrides and optimization start/stop/restart

While much of the control is available via the Opsani UI (optune.ai) or by an appropriately configured config.yaml document (or configmap in this case) for the Servo, there are certain times whne it's convenient to use a command line tool to update configuration overrides. We've developed a light weight python based tool to support overriding the initial servo configuration, and to update some
of the optimization settings.

The tool is available as source here[https://github.com/opsani/coctl], or as a docker image (`docker pull opsani/coctl`).

To use Coctl, you will likely want to set three environment variables:
  CO_TOKEN={your Opsani app token}
  CO_DOMAIN={your Opsani account name}
  CO_APP={your Opsani app name}

Often it is convenient to include these environment parameters in a per-application file and source the appropriate file if working with multiple applications:

```bash
cat > ~/opsani-app.env <<EOF
export CO_TOKEN=REPLACE_WITH_YOUR_OPSANI_APP_TOKEN
export CO_DOMAIN=REPLACE_WITH_YOUR_OPSANI_ACCOUNT
export CO_APP=REPLACE_WITH_YOUR_OPSANI_APP_ID
EOF
```

And then source this file:

```bash
source ~/opsani-app.env
```

and an alias makes life easier if you're running the docker container as an app:

```bash
alias coctl="docker run -it --rm --name coctl -v \$(pwd)/:/work/ -e CO_TOKEN=\$CO_TOKEN -e CO_DOMAIN=\$CO_DOMAIN -e CO_APP=\$CO_APP opsani/coctl:latest "
```

It is often useful to include this, and the previous "source" command in your .bashrc or .bash_profile file in order to ensure it's always available.

While most of the configuration is handled properly in the config.yaml embeded in the opsani-servo-config-vegata.yaml document, we do have to ensure that a) the measurement.control.duration parameter is set to an appropriate duration (300 seconds is a common first measurement target, and can be matched to the duration in the vegeta config).  And we need to ensure we are gathering a performance metric that matches with a metric being deliverd by prometheus.  We do this in the configuratin override that can be managed either via the UI, or via the API (and that via coctl).

```yaml
adjustment:
    control: {}
measurement:
    control:
        duration: 300
        load: {}
        warmup: 0
optimization:
    perf: metrics['main_request_rate']
```

## Assumptions

* Servo and App will live in a K8s namespace that matches the Opsani APP_ID allocated to your project.
* Prometheus will reside in the opsani-monitoring namespace (which will be created) or the URL for prometheus to interact with the app under test will be provided.
* A service-account with a deployment and pod management role will be created for both default and metrics namespaces
* By default, load will be generated using the "hey" load generator app simply making requests against the application endpoint.  Advanced load generation can be accomplished in conjunction with an Opsani Customer Engineer.

Currently the Opsani API uses a TOKEN for authentication, and, assuming you've set an environment variable (CO_TOKEN is very useful for this see the coctl section below) with that TOKEN you can create the secret needed by the Servo Deployment with:

```bash
kubectl create namespace ${CO_APP}
kubectl config set-context --current --namespace=${CO_APP}
kubectl create secret generic optune-auth --from-literal=token=${CO_TOKEN}
```

There is a base Kustomize configuration in the servo-base/ directory, which includes service, deployment and RBAC configurations as needed for a simple web app that performs differently based on available cpu and memory, the servo service along with a baseline configuration for gathering metrics from prometheus and generating load with 'vegeta', and an all-in-one prometheus deployment if prometheus is not already deployed into the environment.

As this environment is set up to use kustomize (which is incorporated into the kubectl client as of the 1.14 release), we can make the needed modifications to the files in the load/ directory, specficially:

1. Update the opsani-servo-account.yaml document to include both the CO_ACCOUNT (usually your coroporate domain name) and your CO_APP id.

If you set these parameters based on the `coctl` section, you can create this file with:

```bash
cat > servo-load/opsani-servo-account.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: opsani-servo

spec:
  template:
    spec:
      containers:
        - name: main
          imagePullPolicy: Always
          image: opsani/servo-k8s-prom-vegeta
          args:
          - ${CO_APP}
          - '--auth-token=/etc/optune-auth/token'
          env:
          - name: OPTUNE_ACCOUNT
            value: ${CO_DOMAIN}
EOF
```

2. Update the configuration template in servo-load/opsani-servo-config-map-vegeta.yaml

You only need to modify the configurations if you are not using the default "web" application name or if you need to change the labels being used to select metrics from the appropraite pods.

You may also need to change the load generator and metrics gathering duration(s)

Note that the config.yaml needs to have the whole config file defined/updated due to the way Kustomize matchs parameters (the file is _one_ parameter to Kustomize)

## Launch the enviroment

Firstly deploy the web service, prometheus, and if appropraite the ingress controller.  Currently the ingress includes adjustments for services with an AWS L4 load balancer.

```bash
kubectl apply -k prometheus/
kubectl apply -k web/
```

If you do want to leverage the ingress, you can deploy the ingress-aws-l4 kustomize service:

```bash
kubectl apply -k ingress-aws-l4
```

Ensure that the ingress is up and running and has a target defined:

```bash
AWS_SLB=$(kubectl get ingress web -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
cat >> servo-load/opsani-servo-config-map-vegeta.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: opsani-servo-config
data:
  config.yaml: |
    k8s:
      application:
        components:
          web-main:
            settings:
              cpu:
                min: 0.125
                max: 1.0
              replicas:
                min: 1
                max: 4
    prom:
      prometheus_endpoint: 'http://prometheus.opsani-monitoring.svc:9090'
      metrics:
        main_request_rate:
          query: sum(rate(envoy_cluster_upstream_rq_total{app="web",role="main"}[1m]))
          unit: rps
        main_p90_time:
          query: histogram_quantile(0.9,sum(rate(envoy_cluster_external_upstream_rq_time_bucket{app="web",role="main"}[1m])) by (envoy_cluster_name, le))
          unit: ms
        main_error_rate:
          query: sum(rate(envoy_cluster_external_upstream_rq_xx{app="web",envoy_response_code_class!="2",role="main"}[1m]))
          unit: rpm
        main_median_response_time:
          query: avg(histogram_quantile(0.5,rate(envoy_cluster_external_upstream_rq_time_bucket{app="web",role="main"}[1m])))
          unit: ms
    vegeta:
      rate: 30000/m
      duration: 5m
      target: GET http://${AWS_SLB:-web}:80/
      workers: 50
      max-workers: 100
      interactive: true
EOF
```

*NOTE: configmaps need to be updated in their entirety, as the map is a single "value" in the document.

Once the changes have been made, you should now be able to trigger an optimization "onboarding" test, which should produce a metric:

```bash
kubectl apply -k servo-load/
```

You can check the servo logs:

```bash
kubectl logs -f $(kubectl get pods -o jsonpath='{.items[?(@.metadata.labels.comp=="opsani-servo")].metadata.name}')
```

If you make a change to the configmap (update the document), re-apply with `kubectl apply -k servo-load`, and then re-start the servo by first deleting the current servo:

```bash
kubectl delete pod $(kubectl get pods -o jsonpath='{.items[?(@.metadata.labels.comp=="opsani-servo")].metadata.name}')
```

And you can also review the logs in the Opsani UI[<https://optune.ai>]

Changes to the load generator can be made by updating the opsani-servo-config-map-hay.yaml document, and re-applying the kustomization:

```bash
kubectl apply -k load/
```

And if the optimization isn't running (e.g. you're seeing "SLEEP" responses in the servo log), you can restart the optimization with the coctl tool (see below):

```bash
coctl restart
```

You can check the state of the optimizer as well:

```bash
coctl status
```
