FROM ubuntu:25.04

RUN mkdir -p /etc/api-gateway
COPY configs/config.yaml /etc/api-gateway/config.yaml

RUN useradd -r -u 2000 -s /bin/false api-gateway && \
    chown -R api-gateway /etc/api-gateway

USER api-gateway

WORKDIR /app

COPY --chown=api-gateway --chmod=500 api-gateway ./api-gateway

ENTRYPOINT ["./api-gateway", "server"]

