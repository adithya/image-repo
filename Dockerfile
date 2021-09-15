FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY gcp-service-acc-creds.json ./
COPY gcp-private-key.pem ./
RUN go build -o /shopify-challenge

EXPOSE 8080

CMD ["/shopify-challenge"]