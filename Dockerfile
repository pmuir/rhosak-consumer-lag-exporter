#
# Copyright RedHat.
# License: MIT License see the file LICENSE
#
FROM registry.access.redhat.com/ubi8/go-toolset:latest as builder


COPY --chown=default:root / .

# install all deps, and build
RUN go build

# ---------------------------------------------------------------------------- #

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

COPY --from=builder --chown=root:root /opt/app-root/src/rhosak-consumer-lag-exporter /opt/app-root/src/rhosak-consumer-lag-exporter

EXPOSE 80

USER 1001

CMD ["/opt/app-root/src/rhosak-consumer-lag-exporter", "export", "--serve", "--host=0.0.0.0", "--port=80"]
