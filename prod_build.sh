#!/bin/bash

# switch to root
su -;
# install docker
curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun;
# install nginx && mysql
docker pull nginx:latest;
docker pull mysql:latest;
docker pull
