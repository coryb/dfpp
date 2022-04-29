FROM --platform=$TARGETPLATFORM scratch AS base
ARG TARGETOS TARGETARCH
COPY ./dist/dfpp-${TARGETOS}-$TARGETARCH /bin/dfpp

FROM base
MAINTAINER Cory Bennett <docker@corybennett.org> https://github.com/coryb/dfpp
COPY docker-root/ /
WORKDIR /root
ENTRYPOINT ["/bin/dfpp"]
