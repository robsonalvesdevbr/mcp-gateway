FROM tinygo/tinygo:0.37.0@sha256:d1307cd9d2bd25167dbbd5abe253f09f2036e7993474bd88df72c1cd3ace0bc0 AS build
WORKDIR /src
RUN --mount=target=. tinygo build -opt=z -scheduler=none -gc=leaking -no-debug -o /tmp/jcat main.go

FROM alpine:3.21@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c AS create-user
RUN adduser -D user

FROM scratch
COPY --from=create-user /etc/passwd /etc/passwd
COPY --from=build /tmp/jcat /
USER user
ENTRYPOINT ["/jcat"]