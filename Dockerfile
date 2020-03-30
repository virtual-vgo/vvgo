FROM golang:1.14.1 as builder
WORKDIR /go/src/vvgo
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/vvgo

FROM builder as tester
RUN apt-get update
RUN apt-get install -y shellcheck
CMD ["go", "test", "-v", "./..."]

FROM gcr.io/distroless/base-debian10 as vvgo
COPY --from=builder /go/bin/vvgo ./vvgo
COPY --from=builder /go/src/vvgo/public ./public
EXPOSE 8080
CMD ["/vvgo"]
