FROM node:16.13 as node
WORKDIR /wrk/
COPY package.json .
COPY yarn.lock .
RUN yarn install

COPY public public
COPY src src
COPY tsconfig.json .
RUN npm run build

FROM golang:1.17 as builder
WORKDIR /go/src/app/
ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY pkg pkg
RUN go generate ./...
RUN go build -o vvgo ./cmd/vvgo

COPY .git .git
RUN go run ./cmd/version

FROM alpine:3 as vvgo
RUN apk add --no-cache ca-certificates apache2-utils tzdata
WORKDIR /app
COPY LICENSE .
COPY public ./public
COPY --from=node /wrk/build ./build
COPY --from=builder /go/src/app/vvgo ./vvgo
COPY --from=builder /go/src/app/version.json ./version.json
EXPOSE 8080
CMD ["./vvgo"]
ENTRYPOINT ["./vvgo"]
