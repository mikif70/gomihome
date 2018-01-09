#!/bin/bash

docker run \
        --rm \
        -v /opt/gomihome:/data \
        --name gomihome \
        --log-opt max-size=2m \
        --log-opt max-file=5 \
        mikif70/gomihome:1.1.4 \
                -i 172.17.0.1:8086 \
                -D \
				-l /data/gomihome.log \
				-t 15m \
                $@
