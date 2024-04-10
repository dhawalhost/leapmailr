#! /bin/bash


# Stop and Clean up  any existing containers with the same name.
docker stop leapmailr
docker rm leapmailr
docker rmi leapmailr

# Build the Docker image
docker build --no-cache -t leapmailr .

#Run the Docker Container
docker  run -it --name leapmailr -d -p 3444:8080 leapmailr
