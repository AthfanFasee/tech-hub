FROM alpine:latest

RUN mkdir /app

COPY postsApp /app

CMD [ "app/postsApp" ]