FROM golang:1.24.5

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o scanner

EXPOSE 8080

RUN apt update && apt install -y iputils-ping

CMD ["./scanner"]
