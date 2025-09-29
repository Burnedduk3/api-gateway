FROM alpine:3.22

RUN mkdir -p /etc/product-service
COPY configs/product-config.yaml /etc/product-service/config.yaml

RUN adduser -u 2000 -S product-service && \
    chown -R product-service /etc/product-service

USER product-service

WORKDIR /app

COPY --chown=product-service --chmod=500 product-service ./product-service

ENTRYPOINT ["echo 'Container Running'; sleep infinity"]

