FROM golang:latest

WORKDIR /app

COPY ./app .

RUN GOOS=linux go build -ldflags="-w -s" -o go-stressTest .

ENTRYPOINT ["./go-stressTest"]