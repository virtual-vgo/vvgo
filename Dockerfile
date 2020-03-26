FROM golang:1.14.1
EXPOSE 8080
WORKDIR /go/src/vvgo
COPY . .

RUN go install -v ./...

CMD ["vvgo"]