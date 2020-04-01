FROM golang:1.14.1 as builder
WORKDIR /go/src/vvgo
COPY . .
RUN go mod download
ARG GITHUB_REF
ARG GITHUB_SHA
RUN go generate
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/vvgo

FROM builder as tester
CMD ["go", "test", "-v", "./..."]

FROM gcr.io/distroless/base-debian10 as vvgo
COPY --from=builder /go/bin/vvgo /vvgo
COPY --from=builder /go/src/vvgo/public /public
EXPOSE 8080
CMD ["/vvgo"]
ENTRYPOINT ["/vvgo"]
