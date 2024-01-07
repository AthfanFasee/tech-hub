FROM alpine:latest

WORKDIR /app

COPY authenticationApp /app

COPY app.env /app

CMD [ "./authenticationApp" ]
