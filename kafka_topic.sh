#!/bin/bash

# wait kafka to be ready
while ! nc -z kafka:9092; do
    sleep 1
done


# create a topic with proper config
kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists \
  --topic user-created \
  --partitions 3 \
  --replication-factor 1 \
  --config cleanup.policy=compact \
  --config retention.ms=604800000  # 7 days


kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists \
  --topic user-updated \
  --partitions 3 \
  --replication-factor 1