#!/bin/bash

# THIS PROGRAM STARTS OR STOPS ALL THE SERVICES LISTED IN CONFIG.CFG

# RUN IT WITH
# ./run -start (to start services) or ./run -stop (to stop all running services)

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
source "$SCRIPT_DIR/config.cfg"

start(){
    for service in "${services[@]}"; do
        cd "$service"
        make run/api
        cd ..
    done
}

stop(){
    for service in "${services[@]}"; do
        cd "$service"
        make down
        cd ..
    done
}

if [ "$1" == "-start" ]; then
    start
elif [ "$1" == "-stop" ]; then
    stop
else
    echo "Usage: $0 {-start|-stop}"
    exit 1
fi