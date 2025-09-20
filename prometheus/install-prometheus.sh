#!/bin/bash

helm install my-prometheus prometheus-community/kube-prometheus-stack -f values.yaml
