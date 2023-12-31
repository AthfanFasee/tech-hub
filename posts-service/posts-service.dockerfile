FROM alpine:latest

RUN mkdir /app

COPY authenticationApp /app

CMD [ "app/authenticationApp" ]