#!/bin/bash

# THIS PROGRAM STARTS OR STOPS ALL THE SERVICES LISTED IN CONFIG.CFG

# RUN IT WITH
# ./run -start (to start services) or ./run -stop (to stop all running services)

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
source "$SCRIPT_DIR/config.cfg"

start(){
    # start shared services first
    docker-compose -f ../../infra/shared_docker_services/docker-compose.yaml up
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
    # stop shared services
    docker-compose -f ../../infra/shared_docker_services/docker-compose.yaml down
}

if [ "$1" == "-start" ]; then
    start
elif [ "$1" == "-stop" ]; then
    stop
else
    echo "Usage: $0 {-start|-stop}"
    exit 1
fi