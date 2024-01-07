# Build a very small docker image
FROM alpine:latest

WORKDIR /app

COPY brokerApp /app

COPY app.env /app

CMD [ "./brokerApp" ]
