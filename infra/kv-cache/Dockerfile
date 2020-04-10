FROM redis:5.0.8
WORKDIR /usr/local/etc/redis/
COPY redis.conf .
EXPOSE 6379
CMD [ "redis-server", "./redis.conf" ]
