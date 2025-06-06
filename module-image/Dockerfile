FROM scratch AS suffix-windows
ENV BINARY_SUFFIX=".exe"

FROM scratch AS suffix-darwin
ENV BINARY_SUFFIX=""

FROM scratch AS suffix-linux
ENV BINARY_SUFFIX=""

FROM suffix-$TARGETOS
ARG TARGETARCH
ARG TARGETOS
COPY dist/${TARGETOS}_$TARGETARCH/docker-mcp$BINARY_SUFFIX /cli-plugins/mcp/$TARGETOS/docker-mcp$BINARY_SUFFIX
COPY module-metadata.json /