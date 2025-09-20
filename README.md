# Prometheus 및 Grafana 배포 가이드

이 문서는 쿠버네티스 클러스터에 `kube-prometheus-stack` Helm 차트를 사용하여 Prometheus와 Grafana를 배포하고 연동하는 과정을 안내합니다. 이 스택을 사용하면 Prometheus Operator가 설치되어 `ServiceMonitor`와 같은 CRD를 통해 선언적으로 모니터링 대상을 관리할 수 있습니다.

## 사전 요구사항

1.  **실행 중인 쿠버네티스 클러스터**: `kind`, `minikube` 또는 클라우드 기반 클러스터.
2.  **`kubectl` CLI**: 클러스터에 연결되도록 설정되어 있어야 합니다.
3.  **`helm` CLI**: Helm 3 버전이 설치되어 있어야 합니다.
4.  **`cilium`: `cilium`의 `cni.exclusive`가 `false`로 설정되어 있어야 합니다.

---

## 1단계: Helm Repository 추가

`kube-prometheus-stack` 차트가 포함된 `prometheus-community` Helm 리포지토리를 추가하고 업데이트합니다.

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

---

## 2단계: 네임스페이스 생성

Prometheus와 Grafana를 포함한 모든 모니터링 관련 리소스를 배포할 별도의 네임스페이스를 생성합니다. 여기서는 `grafana`를 사용합니다.

```bash
kubectl create namespace grafana
```

---

## 3단계: `values.yaml` 설정 파일 작성

배포를 위한 설정 파일(`prometheus/values.yaml`)을 작성합니다. 이 파일은 Prometheus가 여러 네임스페이스(`sample`, `bookinfo` 등)를 모니터링하고, Grafana를 함께 설치하도록 설정합니다.

```yaml
# prometheus/values.yaml

# Alertmanager는 이 가이드에서 비활성화합니다.
alertmanager:
  enabled: false

# Grafana 설치를 활성화합니다.
grafana:
  enabled: true
  # Grafana Admin 기본 비밀번호 설정 (실제 환경에서는 변경 권장)
  adminPassword: "admin"

# Prometheus Operator 및 Prometheus 자체에 대한 설정
prometheus:
  prometheusSpec:
    # ServiceMonitor를 찾을 네임스페이스를 지정합니다.
    # 이 목록에 있는 네임스페이스의 ServiceMonitor만 감지합니다.
    serviceMonitorNamespaceSelector:
      matchNames:
        - grafana
        - sample
        - bookinfo
        - istio-system

    # 어떤 라벨이 붙은 ServiceMonitor를 선택할지 지정합니다.
    # ServiceMonitor 리소스에는 반드시 이 라벨이 포함되어야 합니다.
    serviceMonitorSelector:
      matchLabels:
        release: my-prometheus

    # PodMonitor와 PrometheusRule에 대해서도 동일하게 네임스페이스와 라벨 셀렉터를 설정합니다.
    podMonitorNamespaceSelector:
      matchNames:
        - grafana
        - sample
        - bookinfo
        - istio-system
    podMonitorSelector:
      matchLabels:
        release: my-prometheus

    ruleNamespaceSelector:
      matchNames:
        - grafana
        - sample
        - bookinfo
        - istio-system
    ruleSelector:
      matchLabels:
        release: my-prometheus
```

---

## 4단계: Helm 차트 배포

작성한 `values.yaml` 파일을 사용하여 `grafana` 네임스페이스에 `kube-prometheus-stack`을 배포합니다. 릴리스 이름은 `my-prometheus`로 지정합니다.

```bash
helm install my-prometheus prometheus-community/kube-prometheus-stack \
  --namespace grafana \
  -f ./prometheus/values.yaml
```

---

## 5단계: 배포 확인 및 UI 접근

1.  **배포 확인**:
    잠시 후, `grafana` 네임스페이스의 파드들이 모두 `Running` 상태가 되었는지 확인합니다.
    ```bash
    kubectl -n grafana get pods
    ```

2.  **Prometheus UI 접근**:
    로컬 PC에서 Prometheus 대시보드에 접근하기 위해 포트 포워딩을 설정합니다.
    ```bash
    # 새 터미널을 열고 실행
    kubectl -n grafana port-forward svc/my-prometheus-kube-prometh-prometheus 9090:9090
    ```
    이제 웹 브라우저에서 `http://localhost:9090`으로 접속할 수 있습니다.

3.  **Grafana UI 접근**:
    로컬 PC에서 Grafana에 접근하기 위해 포트 포워딩을 설정합니다.
    ```bash
    # 새 터미널을 열고 실행
    kubectl -n grafana port-forward svc/my-prometheus-grafana 3000:80
    ```
    이제 웹 브라우저에서 `http://localhost:3000`으로 접속할 수 있습니다.
    *   **사용자명**: `admin`
    *   **비밀번호**: `values.yaml`에서 설정한 `adminPassword` 값 (예: `admin`)

---

## 6단계: Grafana와 Prometheus 연동 확인

`kube-prometheus-stack` 차트는 Grafana가 Prometheus를 데이터 소스로 사용하도록 **자동으로 설정**합니다. 이 설정이 올바르게 되었는지 확인합니다.

1.  Grafana UI에 로그인합니다.
2.  왼쪽 메뉴에서 톱니바퀴 아이콘(Configuration) > **Data Sources**로 이동합니다.
3.  **Prometheus**라는 이름의 데이터 소스가 이미 추가되어 있고, 초록색 체크 표시와 함께 "Data source is working" 메시지가 표시되는지 확인합니다.

이것으로 Prometheus와 Grafana의 배포 및 연동이 완료되었습니다. 이제 `sample-app/README.md` 가이드에 따라 애플리케이션 메트릭을 수집할 준비가 되었습니다.
