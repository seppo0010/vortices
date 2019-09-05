FROM golang:latest
RUN apt update && apt install -y libnetfilter-queue-dev iptables tcpdump
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main . && setcap cap_net_admin=+ep main && chmod a+x main
EXPOSE 8080
CMD ["./main"]
