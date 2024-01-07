FROM alpine:latest

WORKDIR /app

COPY mailerApp /app

COPY templates /templates

COPY app.env /app

CMD [ "./mailerApp" ]
