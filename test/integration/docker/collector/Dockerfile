FROM observiq/observiq-otel-collector:1.27.0

USER root

# Copy manager.yaml into container instead of mounting as volume, because
# the collector process will re-write the manager during startup in order
# to merge runtime environment variables into it.
COPY manager.yaml /etc/otel/manager.yaml
RUN \
    chown -R otel:otel /etc/otel/manager.yaml&& \
    chown otel:otel /etc/otel/manager.yaml

USER otel
