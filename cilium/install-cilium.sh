#!/bin/bash

helm install cilium cilium/cilium --version 1.15.7 \
   --namespace kube-system \
   --set operator.replicas=1 \
   --set cni.exclusive=false # istio의 istio-cni 를 사용하기 위해 false로 설정합니다.
