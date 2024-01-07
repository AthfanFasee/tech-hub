FROM alpine:latest

WORKDIR /app

COPY listenerApp /app

COPY app.env /app

CMD [ "./listenerApp" ]
