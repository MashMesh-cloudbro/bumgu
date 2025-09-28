#!/bin/bash

helm upgrade --install cilium cilium/cilium --version 1.15.7 \
   --namespace kube-system \
   --set operator.replicas=1 \
   --set cni.exclusive=false \
   --set hubble.enabled=true \
   --set hubble.relay.enabled=true \
   --set hubble.ui.enabled=true \
   --set prometheus.enabled=true \
   --set hubble.metrics.enabled="{dns,drop,tcp,flow,port-distribution,icmp,http}"
