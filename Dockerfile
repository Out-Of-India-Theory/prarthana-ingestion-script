FROM golang:1.21-alpine

ARG TOKEN
ENV GOPRIVATE="github.com/Out-Of-India-Theory"

RUN apk add git

RUN go env -w GOPRIVATE="github.com/Out-Of-India-Theory" \
    && git config --global url."https://oit-devops:${TOKEN}@github.com".insteadOf "https://github.com"

COPY . /go/src/github.com/Out-Of-India-Theory/prarthana-automated-script

WORKDIR /go/src/github.com/Out-Of-India-Theory/prarthana-automated-script

RUN echo $GOPRIVATE

RUN go mod tidy \
    && go mod download

RUN GOOS=linux GOARCH=arm64 go build -o main .

EXPOSE 8080

CMD ["./main"]