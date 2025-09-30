FROM alpine:3.22

RUN mkdir -p /etc/api-gateway
COPY configs/config.yaml /etc/api-gateway/config.yaml

RUN adduser -u 2000 -S api-gateway && \
    chown -R api-gateway /etc/api-gateway

USER api-gateway

WORKDIR /app

COPY --chown=api-gateway --chmod=500 api-gateway ./api-gateway

ENTRYPOINT ["echo 'Container Running'; sleep infinity"]

