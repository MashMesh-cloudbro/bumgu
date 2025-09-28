#!/bin/bash

helm install cilium cilium/cilium --version 1.15.7 \
   --namespace kube-system \
   --set operator.replicas=1 \
   --set cni.exclusive=false
