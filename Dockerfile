FROM node:14.5 as node
WORKDIR /wrk/
COPY package.json .
COPY package-lock.json .
RUN npm install

COPY ui ui
COPY .eslintrc.js .
COPY tsconfig.json .
COPY webpack.config.js .
RUN npx webpack --mode=production

FROM golang:1.16 as builder
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

FROM alpine:3.4 as vvgo
RUN apk add --no-cache ca-certificates apache2-utils
WORKDIR /app
COPY LICENSE .
COPY public ./public
COPY --from=node /wrk/public/dist ./public/dist
COPY --from=builder /go/src/app/vvgo ./vvgo
COPY --from=builder /go/src/app/version.json ./version.json
EXPOSE 8080
CMD ["./vvgo"]
ENTRYPOINT ["./vvgo"]
