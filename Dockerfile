FROM cgr.dev/chainguard/go:latest AS builder
WORKDIR /build
COPY go.* ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 go build -o freemaild -a -ldflags '-w' cmd/main.go

FROM cgr.dev/chainguard/wolfi-base:latest AS certs
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=builder /build/freemaild /bin/freemaild
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/bin/freemaild"]
CMD ["/bin/freemaild"]
