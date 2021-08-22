FROM node:13.12 as node
COPY public/package.json .
COPY public/package-lock.json .
RUN npm install

FROM golang:1.16 as builder
WORKDIR /go/src/app/
ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY pkg pkg
COPY tools tools
RUN go generate ./...
RUN go build -o vvgo ./cmd/vvgo

COPY .git .git
RUN go run ./tools/version

FROM node:13.12 as parts_browser
COPY parts_browser .
RUN npm install && npm run-script build

FROM alpine:3.4 as vvgo
RUN apk add --no-cache ca-certificates apache2-utils
COPY --from=node node_modules /public/node_modules
COPY --from=builder /go/src/app/vvgo /vvgo
COPY ./public /public
COPY --from=builder /go/src/app/version.json ./version.json
COPY --from=parts_browser build ./parts_browser/build
EXPOSE 8080
CMD ["/vvgo"]
ENTRYPOINT ["/vvgo"]
