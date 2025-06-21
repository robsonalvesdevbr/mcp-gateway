FROM golang:1.23-alpine@sha256:9a425d78a8257fc92d41ad979d38cb54005bac3fdefbdadde868e004eccbb898 AS build
WORKDIR /app
COPY main.go ./
RUN go build -o /main main.go

FROM alpine:3.22@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715
COPY --from=build /main /
ENTRYPOINT ["/main"]