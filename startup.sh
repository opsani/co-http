#!/bin/sh
if [[ -n "$OTEL_RESOURCE_ATTRIBUTES" ]]; then
  OTEL_RESOURCE_ATTRIBUTES="$OTEL_RESOURCE_ATTRIBUTES,"
fi
export OTEL_RESOURCE_ATTRIBUTES="${OTEL_RESOURCE_ATTRIBUTES}container.id="$(sed -rn "/\/sys\/fs\/cgroup\/devices/ s#.*/(cri-containerd-)?([0-9a-f]{64})(\.scope)? .*#\2#p" /proc/self/mountinfo)
echo "Starting 'co-http $@' with the following environment variables:"
echo "OTEL_EXPORTER_OTLP_ENDPOINT=$OTEL_EXPORTER_OTLP_ENDPOINT"
echo "OTEL_EXPORTER_OTLP_INSECURE=$OTEL_EXPORTER_OTLP_INSECURE"
echo "OTEL_RESOURCE_ATTRIBUTES=$OTEL_RESOURCE_ATTRIBUTES"
echo "OTEL_TRACES_EXPORTER=$OTEL_TRACES_EXPORTER"
echo "OTEL_SERVICE_NAME=$OTEL_SERVICE_NAME"
exec /co-http "$@"