FROM alpine:latest

WORKDIR /app

COPY postsApp /app

COPY app.env /app

CMD [ "./postsApp" ]