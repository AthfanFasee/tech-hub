FROM alpine:latest

RUN mkdir /app

COPY paymentApp /app

CMD [ "app/paymentApp" ]