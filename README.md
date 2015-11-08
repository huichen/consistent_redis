# How to run a masterless redis cluster with consistent hashing and data redundancy

## Set consistent service environment

Follow the [instruction](https://github.com/huichen/consistent_service) to have etcd and registrator installed on your cluster.

## Build redis docker image

    docker build -t unmerged/redis -f Dockerfile .
  
## Run redis containers

Run on each of your hosts in the cluster. Prefer to have at least three redis copies running.

    docker run -d -p <your host ip>::6379 unmerged/redis
  
## Test consistent redis client

    go run main.go --endpoints=http://<your etcd endpoint ip:port> --service_name=/services/redis
