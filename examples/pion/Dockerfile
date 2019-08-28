FROM golang:latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
RUN setcap cap_net_raw=+ep /app/main
EXPOSE 8080
CMD ["./main"]
