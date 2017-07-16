FROM registry.z.xinu.tv/golang/alpine/onbuild AS base
FROM alpine

COPY --from=base /go/bin/app /bin/surfer
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
