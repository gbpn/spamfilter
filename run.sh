#!/bin/bash
which docker
docker ps
docker build -t gbpn-bayes .
docker run -it -p 8080:8080 --rm bayes go run .

