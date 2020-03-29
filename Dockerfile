FROM golang:1.14.1 as builder
WORKDIR /go/src/vvgo
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -v -o vvgo

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/vvgo .
EXPOSE 8080
CMD ["/vvgo"]