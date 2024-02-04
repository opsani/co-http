#!/bin/sh
export OTEL_RESOURCE_ATTRIBUTES="$OTEL_RESOURCE_ATTRIBUTES","container.id="$(sed -rn "/\/sys\/fs\/cgroup\/devices/ s#.*/(cri-containerd-)?([0-9a-f]{64})(\.scope)? .*#\2#p" /proc/self/mountinfo)
echo "Starting 'co-http $@' with the following environment variables:"
echo "OTEL_EXPORTER_OTLP_ENDPOINT=$OTEL_EXPORTER_OTLP_ENDPOINT"
echo "OTEL_RESOURCE_ATTRIBUTES=$OTEL_RESOURCE_ATTRIBUTES"
exec /co-http "$@"