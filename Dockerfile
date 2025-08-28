FROM golang:1.24.5


WORKDIR /app
COPY . .
COPY config.json /app/config.json


RUN mkdir -p /app/data
RUN go build -o scanner

CMD ["./scanner"]
