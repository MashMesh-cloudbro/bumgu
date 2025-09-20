# #!/bin/bash
#
# 1. productpage 서비스에 http-metrics 포트 추가
kubectl patch service productpage -n bookinfo --type='json' -p='[{"op": "add", "path": "/spec/ports/-", "value": {"name": "http-monitoring", "port": 15020, "protocol": "TCP", "targetPort": 15020}}]'

# 2. details 서비스에 http-metrics 포트 추가
kubectl patch service details -n bookinfo --type='json' -p='[{"op": "add", "path": "/spec/ports/-", "value": {"name": "http-monitoring", "port": 15020, "protocol": "TCP", "targetPort": 15020}}]'

# 3. ratings 서비스에 http-metrics 포트 추가
kubectl patch service ratings -n bookinfo --type='json' -p='[{"op": "add", "path": "/spec/ports/-", "value": {"name": "http-monitoring", "port": 15020, "protocol": "TCP", "targetPort": 15020}}]'

# 4. reviews 서비스에 http-metrics 포트 추가
kubectl patch service reviews -n bookinfo --type='json' -p='[{"op": "add", "path": "/spec/ports/-", "value": {"name": "http-monitoring", "port": 15020, "protocol": "TCP", "targetPort": 15020}}]'
