FROM alpine:latest

WORKDIR /app

COPY loggerApp /app

COPY app.env /app

CMD [ "./loggerApp" ]
