FROM alpine:latest

WORKDIR /app

COPY paymentApp /app

COPY app.env /app

CMD [ "./paymentApp" ]