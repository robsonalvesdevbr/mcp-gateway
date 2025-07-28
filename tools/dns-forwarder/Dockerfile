FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

COPY --from=coredns/coredns:1.12.2 /coredns /coredns
COPY Corefile /Corefile
COPY entrypoint.sh /entrypoint.sh

CMD ["/entrypoint.sh"]
