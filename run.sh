#!/bin/bash

if [[ ! -d "/root/.hbtcd/config" ]]; then
  echo "Initialize /root/.hbtcd"
  cp -r /go/initial-node/* /root/.hbtcd/
fi

cd /go; ./hbtcd start $@
