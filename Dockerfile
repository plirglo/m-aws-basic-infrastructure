FROM hashicorp/terraform:0.13.2 as initializer

COPY resources /resources
RUN cd /resources/terraform && terraform init

FROM hashicorp/terraform:0.13.2

ARG ARG_M_VERSION="unknown"

ENV M_WORKDIR="/workdir" \
    M_RESOURCES="/resources" \
    M_SHARED="/shared" \
    M_VERSION="$ARG_M_VERSION"

COPY --from=initializer /resources/ /resources/
COPY workdir /workdir

ARG ARG_HOST_UID=1000
ARG ARG_HOST_GID=1000

RUN apk add --update --no-cache make=4.3-r0 && \
    wget https://github.com/mikefarah/yq/releases/download/3.3.4/yq_linux_amd64 -O /usr/bin/yq && \
    chmod +x /usr/bin/yq && \
    chown -R $ARG_HOST_UID:$ARG_HOST_GID /workdir /resources

USER $ARG_HOST_UID:$ARG_HOST_GID

WORKDIR /workdir
ENTRYPOINT ["make"]
