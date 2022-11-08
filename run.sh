#!/bin/bash
which docker
docker ps
docker build -t gbpn-bayes .
docker run -p 8080:8080 -it --rm bayes go run .

