FROM golang:1.23-alpine@sha256:9a425d78a8257fc92d41ad979d38cb54005bac3fdefbdadde868e004eccbb898 AS build
WORKDIR /app
COPY main.go ./
RUN go build -o /main main.go

FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1
COPY --from=build /main /
ENTRYPOINT ["/main"]