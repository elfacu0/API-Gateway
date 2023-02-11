FROM golang:1.20-alpine
RUN apk --update add redis
COPY config /etc
EXPOSE 6370
CMD [ "redis-server","/etc/redis.conf","--port","6370"]