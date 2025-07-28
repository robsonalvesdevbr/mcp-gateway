FROM tinygo/tinygo:0.37.0@sha256:d1307cd9d2bd25167dbbd5abe253f09f2036e7993474bd88df72c1cd3ace0bc0 AS build
WORKDIR /src
RUN --mount=target=. tinygo build -opt=z -scheduler=none -gc=leaking -no-debug -o /tmp/jcat main.go

FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1 AS create-user
RUN adduser -D user

FROM scratch
COPY --from=create-user /etc/passwd /etc/passwd
COPY --from=build /tmp/jcat /
USER user
ENTRYPOINT ["/jcat"]