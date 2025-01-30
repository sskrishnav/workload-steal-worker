#!/bin/bash
echo "Build Docker image"
docker build -t workloadstealworker:alpha1 .

echo "Load Image to Kind cluster"
kind load docker-image workloadstealworker:alpha1

echo "Deploying Worker Server"
kubectl apply -f deployment.yaml
