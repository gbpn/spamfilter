#!/bin/bash
which docker
docker ps
docker build -t gbpn-bayes .
docker run -it --rm --name gbpn-bayes gbpn-bayes
