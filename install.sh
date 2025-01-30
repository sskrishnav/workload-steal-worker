#!/bin/bash
echo "Build Docker image"
docker build -t workloadstealworker:alpha1 .

echo "Create a 'kind' cluster with name 'local'"
kind create cluster --name local

echo "Set the kubectl context to local cluster"
kubectl cluster-info --context kind-local

echo "Load Image to Kind cluster named 'local'"
kind load docker-image --name local workloadstealworker:alpha1

echo "Create 'hiro' namespace if it doesn't exist"
kubectl get namespace | grep -q "hiro" || kubectl create namespace hiro

echo "Deploying Worker Server"
kubectl apply -f deploy/deployment.yaml
