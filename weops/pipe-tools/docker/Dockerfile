#
# build container
#
FROM --platform=linux/amd64 golang:1.19-alpine as builder
WORKDIR /

COPY oracledb_exporter-linux-amd64 /oracledb_exporter-linux-amd64

EXPOSE     9161
ENTRYPOINT ["/oracledb_exporter-linux-amd64"]
