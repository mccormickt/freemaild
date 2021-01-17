FROM golang:1.16-alpine AS builder
RUN apk --no-cache add ca-certificates
WORKDIR /build
COPY go.* ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 go build -o freemaild -a -ldflags '-w' cmd/main.go

FROM scratch
COPY --from=builder /build/freemaild /bin/freemaild
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/bin/freemaild"] 
CMD ["/bin/freemaild"] 