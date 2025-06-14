
# syntax=docker/dockerfile:1

FROM alpine:3.22@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715 AS ca-certificates
RUN apk add ca-certificates

FROM golang:1.24-alpine3.22@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder
WORKDIR /app
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /proxy .

FROM scratch
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /proxy /
EXPOSE 8080
ENV ALLOWED_HOSTS=
ENTRYPOINT ["/proxy"]