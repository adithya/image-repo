FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
#COPY go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o /shopify-challenge

EXPOSE 8080

CMD ["/shopify-challenge"]