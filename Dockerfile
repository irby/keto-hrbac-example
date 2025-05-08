FROM golang:1.23.2

WORKDIR /app
COPY . .

RUN go mod tidy && go build -o consent-app

CMD ["./consent-app"]