# `sample-app`을 위한 Istio 메트릭 수집 가이드

이 문서는 `sample-app`을 쿠버네티스에 배포한 후, Istio 서비스 메쉬의 메트릭을 수집하여 Prometheus와 Grafana에서 시각화하기 위한 전체 과정을 안내합니다.

## 사전 요구사항

1.  **Kubernetes 클러스터**: `kind`와 같은 로컬 환경 또는 클라우드 환경의 쿠버네티스 클러스터.
2.  **Istio**: 클러스터에 Istio 제어 플레인(`istiod`)이 설치되어 있어야 합니다.
3.  **Prometheus Operator**: `kube-prometheus-stack`과 같은 Prometheus Operator가 `grafana` 또는 `monitoring` 네임스페이스에 설치되어 있어야 합니다.

---

## 1단계: Istio 사이드카 주입 활성화

`sample` 네임스페이스에 배포되는 모든 파드에 Istio 프록시 사이드카가 자동으로 주입되도록 네임스페이스에 레이블을 추가합니다.

```bash
kubectl label namespace sample istio-injection=enabled --overwrite
```

---

## 2단계: Istio Ingress Gateway를 통한 서비스 노출

`kubectl port-forward`는 Istio 프록시를 우회하여 메트릭이 수집되지 않으므로, 반드시 Istio Ingress Gateway를 통해 서비스를 외부로 노출해야 합니다.

1.  **`Gateway` 및 `VirtualService` 생성**:
    `sample-app/kubernetes/gateway.yaml` 파일을 생성하여 외부 트래픽을 `sample-app` 서비스로 라우팅합니다.

    ```yaml
    # sample-app/kubernetes/gateway.yaml
    apiVersion: networking.istio.io/v1beta1
    kind: Gateway
    metadata:
      name: sample-app-gateway
      namespace: sample
    spec:
      selector:
        istio: ingressgateway # Istio의 기본 인그레스 게이트웨이 사용
      servers:
      - port:
          number: 80
          name: http
          protocol: HTTP
        hosts:
        - "*"
    ---
    apiVersion: networking.istio.io/v1beta1
    kind: VirtualService
    metadata:
      name: sample-app-vs
      namespace: sample
    spec:
      hosts:
      - "*"
      gateways:
      - sample-app-gateway
      http:
      - route:
        - destination:
            host: sample-app
            port:
              number: 80
    ```

2.  **리소스 적용**:

    ```bash
    kubectl apply -f sample-app/kubernetes/gateway.yaml
    ```

---

## 3단계: Prometheus가 메트릭을 수집하도록 설정

Prometheus가 `sample` 네임스페이스의 서비스들을 스크랩 대상으로 인식하도록 설정합니다.

1.  **서비스에 메트릭 포트 추가**:
    `ServiceMonitor`가 파드의 메트릭 엔드포인트를 찾을 수 있도록 각 서비스(`sample-app`, `mysql`)에 Istio의 메트릭 포트(`15020`)를 추가합니다.

    *   `sample-app/kubernetes/sample-app-go.yaml` 수정:
        ```yaml
        # ...
        ports:
        - name: http-app
          port: 80
          targetPort: 8081
        - name: http-envoy-prom # 메트릭 포트 추가
          port: 15020
          targetPort: 15020
        # ...
        ```
    *   `sample-app/kubernetes/mysql.yaml` 수정:
        ```yaml
        # ...
        ports:
        - name: tcp-mysql
          port: 3306
        - name: http-envoy-prom # 메트릭 포트 추가
          port: 15020
          targetPort: 15020
        # ...
        ```
    *   수정 후 리소스 다시 적용:
        ```bash
        kubectl apply -f sample-app/kubernetes/sample-app-go.yaml
        kubectl apply -f sample-app/kubernetes/mysql.yaml
        ```

2.  **`ServiceMonitor` 생성**:
    `sample-app`과 `mysql` 각각에 대한 `ServiceMonitor`를 생성합니다. `release: my-prometheus` 레이블은 Prometheus Operator 설정과 일치해야 합니다. (자신의 환경에 맞게 수정)

    *   `sample-app/kubernetes/servicemonitor.yaml`
    *   `sample-app/kubernetes/mysql-servicemonitor.yaml`
    *   리소스 적용:
        ```bash
        kubectl apply -f sample-app/kubernetes/servicemonitor.yaml
        kubectl apply -f sample-app/kubernetes/mysql-servicemonitor.yaml
        ```

3.  **Prometheus 모니터링 네임스페이스 설정**:
    Prometheus가 `sample` 네임스페이스를 감시하도록 설정합니다. `kube-prometheus-stack`의 `values.yaml`을 수정하고 Helm 업그레이드를 수행합니다.

    *   `prometheus/values.yaml` 파일 수정:
        ```yaml
        prometheus:
          prometheusSpec:
            serviceMonitorNamespaceSelector:
              matchNames:
                - grafana # 기존 모니터링 네임스페이스
                - sample  # sample 네임스페이스 추가
                - bookinfo
        ```
    *   Helm 업그레이드 실행:
        ```bash
        helm upgrade my-prometheus prometheus-community/kube-prometheus-stack -n grafana -f ./prometheus/values.yaml
        ```

### 💡 `kube-prometheus-stack`과 `ServiceMonitor`의 역할

*   **`kube-prometheus-stack` 설정 (`values.yaml`)**: Prometheus라는 **감시자**를 설정하는 것입니다. 이 설정은 Prometheus에게 "어떤 네임스페이스(`serviceMonitorNamespaceSelector`)를 감시하고, 어떤 특정 라벨(`serviceMonitorSelector`)이 붙은 `ServiceMonitor` 문서를 찾아야 하는지" 알려줍니다.
*   **`ServiceMonitor` 리소스 (`servicemonitor.yaml`)**: 각 애플리케이션이 Prometheus에게 "나를 어떻게 모니터링해야 하는지" 알려주는 **명세서**입니다. 이 명세서에는 메트릭 포트, 경로, 수집 주기 등의 정보가 담겨 있습니다.

따라서, `kube-prometheus-stack` 설정만으로는 충분하지 않으며, 모니터링할 각 애플리케이션에 대한 `ServiceMonitor` 명세서를 반드시 별도로 생성하고 적용해야 합니다.

---

## 4단계: 상세 메트릭 수집 활성화

요청/응답 크기, 상세 레이턴시 등 더 많은 메트릭을 수집하려면 `Telemetry` 리소스를 `sample` 네임스페이스에 적용해야 합니다.

1.  **`telemetry.yaml` 파일 생성**:
    ```yaml
    # sample-app/kubernetes/telemetry.yaml
    apiVersion: telemetry.istio.io/v1alpha1
    kind: Telemetry
    metadata:
      name: sample-app-full-metrics
      namespace: sample
    spec:
      metrics:
      - providers:
        - name: prometheus
    ```

2.  **리소스 적용**:
    ```bash
    kubectl apply -f sample-app/kubernetes/telemetry.yaml
    ```

---

## 5단계: 트래픽 발생 및 Grafana에서 확인

1.  **Ingress Gateway로 포트 포워딩**:
    `NodePort` 대신 포트 포워딩을 사용하면 로컬에서 쉽게 접근할 수 있습니다.
    ```bash
    kubectl -n istio-system port-forward svc/istio-ingressgateway 8080:80
    ```

2.  **트래픽 발생**:
    새 터미널에서 `curl`을 사용하여 여러 번 요청을 보냅니다.
    ```bash
    # 10초 간격으로 10번 요청 보내기
    for i in {1..10}; do curl http://localhost:8080/users; sleep 10; done
    ```

3.  **Grafana 대시보드 확인**:
    *   **HTTP 메트릭 (`sample-app`)**: Grafana에서 **"Istio Mesh Dashboard"**를 엽니다.
    *   **TCP 메트릭 (`mysql`)**: Grafana에서 **"Istio TCP Metrics"** 대시보드를 엽니다.
    *   두 대시보드 모두 상단의 **"Namespace"** 필터에서 **`sample`**을 선택해야 데이터가 올바르게 표시됩니다.

이제 모든 메트릭이 정상적으로 수집되고 Grafana 대시보드에서 시각화될 것입니다.
