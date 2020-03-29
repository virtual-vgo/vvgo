FROM golang:1.14.1 as builder
WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/app

FROM gcr.io/distroless/base-debian10 as artifact
COPY --from=builder /go/bin/app .
EXPOSE 8080
CMD ["/app"]
