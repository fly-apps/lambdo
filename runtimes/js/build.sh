#!/usr/bin/env bash

# Requires building on x86-64 architecture
# for Fly VM's
docker build \
  -t fideloper/lambdo-js:20 \
  -f Dockerfile \
  .

# aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws/i2r3m5g4
# docker push public.ecr.aws/i2r3m5g4/runtime-js:latest
