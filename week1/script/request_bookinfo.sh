#!/bin/bash

for ((i=1; i<=50; i++)); do
    curl localhost:8080/productpage;
    sleep 1;
done
