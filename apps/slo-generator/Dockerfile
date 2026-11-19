# Packages fabianoflorentino/slo-generator (a fork of google/slo-generator) with
# the Prometheus backend/exporter. CI publishes this image to GHCR; the CronJob
# in deploy/slo/cronjob.yaml pins the tag.
FROM python:3.12-alpine

ARG SLO_GENERATOR_REPO=https://github.com/fabianoflorentino/slo-generator
ARG SLO_GENERATOR_REF=master

RUN apk add --no-cache git \
    && git clone --depth 1 --branch "${SLO_GENERATOR_REF}" "${SLO_GENERATOR_REPO}" /src \
    && pip install --no-cache-dir "/src[prometheus]" \
    && pip install --no-cache-dir "setuptools<81" \
    && rm -rf /src /root/.cache

ENV HOME=/tmp \
    PYTHONUNBUFFERED=1

USER 65532:65532

ENTRYPOINT ["slo-generator"]
