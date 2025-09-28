#!/bin/bash
# save this script as setup-kind.sh
# Create the cluster
kind create cluster --config kind-config.yaml

# Install NGINX Ingress
#kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
#kubectl wait --namespace ingress-nginx \
#  --for=condition=ready pod \
#  --selector=app.kubernetes.io/component=controller \
#  --timeout=90s

# Verify
kubectl get nodes
#kubectl get pods -n ingress-nginx
