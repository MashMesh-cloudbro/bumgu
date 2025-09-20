#!/bin/bash

kubectl create namespace sample

echo "Waiting for namespace 'sample' to be created..."
while ! kubectl get namespace sample > /dev/null 2>&1; do
  sleep 1
done
echo "Namespace 'sample' created."

kubectl label namespaces sample istio-injection=enabled --overwrite

kubectl apply -f mysql.yaml
kubectl apply -f sample-app-go.yaml
