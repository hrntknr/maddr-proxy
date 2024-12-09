#!/bin/bash -x

if ! dpkg -s docker.io &>/dev/null; then
  apt-get update
  apt-get install -y docker.io
  systemctl enable docker
  systemctl start docker
fi

if ! docker inspect maddr-proxy &>/dev/null; then
  docker run -d --name maddr-proxy --network host --privileged --restart always ghcr.io/hrntknr/maddr-proxy:latest --password "$PASSWORD" --setup-route
fi
