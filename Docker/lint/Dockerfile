ARG EASYP_BASE_VERSION

FROM --platform=${BUILDPLATFORM} easyp/base:${EASYP_BASE_VERSION}

ENV DIR_PATH="/app"
ENV CONFIG_PATH="/app/easyp.yaml"

CMD ["/easyp", "--cfg", "${CONFIG_PATH}", "lint", "-p", "${DIR_PATH}"]
