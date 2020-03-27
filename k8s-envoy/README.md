# Deploy a simple Single Tier (single service)

* Single application (using Opsani co-http app)
* Saturation style load generation using Vegeta load generator
* Prometheus for metrics (using the provided Prometheus manifests)

## Temporary Instructions

There are a few changes still needed in the related servo plugins that are in process of being addressed. In advance of that:

1. Use the rstarmer/servo-k8s-prom-hey:latest image
2. In addition to the load/servo-config-map-hey.yaml, you will need to add an override config (use coctl) (sample in coconfig.yaml)

    ```yaml
    adjustment:
        control: {}
    measurement:
        control:
            duration: 300
            load:
                n_clients: 30
                n_requests: 10000000
                service: web
                t_limit: 300
                test_url: http://web:80
            past: 60
            warmup: 0
    optimization:
        perf: metrics['main_request_rate']
    ```

3. Do not uncomment the control and hey sections from the config-map file until the updates have been applied.

## Assumptions

* Servo and App will live in a K8s namespace that matches the Opsani APP_ID allocated to your project.
* Prometheus will reside in the opsani-monitoring namespace (which will be created) or the URL for prometheus to interact with the app under test will be provided.
* A service-account with a deployment and pod management role will be created for both default and metrics namespaces
* By default, load will be generated using the "hey" load generator app simply making requests against the application endpoint.  Advanced load generation can be accomplished in conjunction with an Opsani Customer Engineer.

Currently the Opsani API uses a TOKEN for authentication, and, assuming you've set an environment variable (CO_TOKEN is very useful for this see the coctl section below) with that TOKEN you can create the secret needed by the Servo Deployment with:

```bash
kubectl create secret generic optune-auth --from-literal=token=${CO_TOKEN}
```

There is a base Kustomize configuration in the base/ directory, which includes service, deployment and RBAC configurations as needed for a simple web app that performs differently based on available cpu and memory, the servo service along with a baseline configuration for gathering metrics from prometheus and generating load with 'hey', and an all-in-one prometheus deployment if prometheus is not already deployed into the environment.

As this environment is set up to use kustomize (which is incorporated into the kubectl client as of the 1.14 release), we can make the needed modifications to the files in the load/ directory, specficially:

1. Update the opsani-servo-account.yaml document to include both the CO_ACCOUNT (usually your coroporate domain name) and your CO_APP id.

  If you set these parameters based on the `coctl` section below, you can create this file with:

  ```bash
  cat > load/opsani-servo-account.yaml <<EOF
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
            image: rstarmer/servo-k8s-prom-hey
            args:
            - ${CO_APP}
            - '--auth-token=/etc/optune-auth/token'
            env:
            - name: OPTUNE_ACCOUNT
              value: ${CO_ACCOUNT}
  EOF
  ```

2. Update the configuration template in load/opsani-servo-config-map-hey.yaml

  You only need to modify the configurations if you are not using the default "web" application name or if you need to change the labels being used to select metrics from the appropraite pods.

  You may also need to change the load generator and metrics gathering duration(s)

  Note that the config.yaml needs to have the whole config file defined/updated due to the way Kustomize matchs parameters (the file is _one_ parameter to Kustomize)

## Launch the enviroment

Once the changes have been made, you should now be able to trigger an optimization "onboarding" test, which should produce a metric:

```bash
kubectl apply -k load/
```

You can check the servo logs:

```bash
kubectl logs -f $(kubectl get pods -o jsonpath='{.items[?(@.metadata.labels.comp=="opsani-servo")].metadata.name}')
```

And you can also review the logs in the Opsani UI[https://optune.ai]

Changes to the load generator can be made by updating the opsani-servo-config-map-hay.yaml document, and re-applying the kustomization:

```bash
kubectl apply -k load/
```

And if the optimization isn't running (e.g. you're seeing "SLEEP" responses in the servo log), you can restart the optimization with the coctl tool (see below):

```bash
coctl restart
```

## Coctl - manage config overrides and optimization start/stop/restart

We've developed a light weight python based tool to support overriding the initial servo configuration, and to update some
of the optimization settings.

The tool is available as source here[https://github.com/opsani/coctl], or as a docker image (`docker pull opsani/coctl`).

To use Coctl, you will likely want to set three environment variables:
  CO_TOKEN={your Opsani app token}
  CO_ACCOUNT={your Opsani account name}
  CO_APP={your Opsani app name}

Often it is convenient to include these environment parameters in a per-application file and source the appropriate file if working with multiple applications:

```bash
cat > ~/opsani-app.env <<EOF
export CO_TOKEN=REPLACE_WITH_YOUR_OPSANI_APP_TOKEN
export CO_ACCOUNT=REPLACE_WITH_YOUR_OPSANI_ACCOUNT
export CO_APP=REPLACE_WITH_YOUR_OPSANI_APP_ID
EOF
```

And then source this file:

```bash
source ~/opsani-app.env
```

and an alias makes life easier if you're running the docker container as an app:

```bash
alias coctl="docker run -it --rm --name coctl -v \$(pwd)/:/work/ -e CO_TOKEN=\$CO_TOKEN -e CO_DOMAIN=\$CO_DOMAIN -e CO_APP=\$CO_APP coctl:latest "
```

It is often useful to include this, and the previous "source" command in your .bashrc or .bash_profile file in order to ensure it's always available.
