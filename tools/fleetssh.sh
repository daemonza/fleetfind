#!/usr/bin/env bash
# helper script using fleetfind to ssh directly to a container
# Coded by Werner Gillmer <werner.gillmer@gmail.com>

DOCKER_CONTAINER=$1

if [ -z $DOCKER_CONTAINER ]; then
	echo "usage : fleetssh.sh <docker container name>"
	exit 1
fi

echo "Finding host server..."
HOST=$(fleetfind list $DOCKER_CONTAINER | awk '/host/ {print $1}' | cut -d: -f2)
if [ -z $HOST ]; then
	echo "$DOCKER_CONTAINER not found on fleet"
	exit 1
fi

echo "Finding docker process..."
DOCKER_PROCESS=$(fleetfind list $DOCKER_CONTAINER | awk '/api-1/ {print $2}')
if [ -z $DOCKER_PROCESS ]; then
	echo "failed to find docker process id for $DOCKER_CONTAINER"
	exit 1
fi


echo "Starting ssh connection to docker process $DOCKER_PROCESS on $HOST"
fleetctl ssh $HOST "docker exec -i -t $DOCKER_PROCESS bash"
