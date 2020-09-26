#!/bin/bash

# Specify docker image name
image_name="cnis"

# Build docker image
docker build -t "$image_name" . -f docker_files/Dockerfile
