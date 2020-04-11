FROM golang:1.14.1 as builder

ARG GITHUB_REF
ARG GITHUB_SHA

ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on

WORKDIR /go/src/github.com/virtual-vgo/vvgo
COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN BIN_PATH=/ make vvgo

FROM builder as tester
CMD ["make", "test"]

FROM gcr.io/distroless/base-debian10 as vvgo
COPY --from=builder vvgo vvgo
COPY ./public /public
EXPOSE 8080
CMD ["/vvgo"]
ENTRYPOINT ["/vvgo"]
