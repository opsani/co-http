# Canary mode with a single tier simple web application

To enable canary mode, we need to generate an optune-auth secret as for the default co-http saturation mode example.  We'll also need to update our optimization configuration with the following data:

1. Both the canary and non-canary resource metrics for the performance function, mapping metrics and know resources (e.g. from adjustment results) back to the principal canary parameters. These are required in order for the canary style optimization to separate canary performance from production metrics.
2. A function for determining resource costs (can be static, a function, or lookup against ec2 resoruces)

An example optimization object that can be pushed via CURL to the Optune backend:

{ "optimization":
  { "mode": "canary",
  "perf": "canary_request_rate",
  "canary_inst_count": "1",
  "prod_inst_count": "3",
  "canary_inst_perf": "canary_request_rate",
  "prod_inst_perf": "main_request_rate",
  "baseline": {"inst_cost": "3"}}}

  curl 