ARG EASYP_BASE_VERSION

FROM easyp/base:${EASYP_BASE_VERSION}

ENV DIR_PATH="/app"
ENV CONFIG_PATH="/app/easyp.yaml"

CMD ["/easyp", "--cfg", "${EASYP_BASE_VERSION}", "lint", "-p", "${DIR_PATH}"]
