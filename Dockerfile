FROM ubuntu:focal

RUN apt-get update && apt-get install -y \
    mysql-client \
    less \
    lsof \
    curl \
    htop \
    && rm -rf /var/lib/apt/lists/*

ADD ./build/blog /root/blog
CMD ["./blog","run"]