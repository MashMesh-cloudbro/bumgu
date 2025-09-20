# Istio 설치 가이드 (Helm 사용)

이 문서는 Helm을 사용하여 Istio를 설치하는 과정을 안내합니다. 특히 Cilium CNI 환경에서 Istio를 함께 사용하는 경우를 고려하여 작성되었습니다.

## 사전 준비 사항

- `kubectl`
- `helm`

## 설치 과정

### 1. Namespace 생성

Istio 컨트롤 플레인과 인그레스 게이트웨이를 위한 Namespace를 생성합니다.

```sh
kubectl create namespace istio-system
kubectl create namespace istio-ingress
```

### 2. Istio Helm Repository 추가

```sh
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update
```

### 3. Istio Base 차트 설치

Istio의 CRD(Custom Resource Definitions)를 설치합니다.

```sh
helm install istio-base istio/base -n istio-system --wait
```

### 4. Istiod (Control Plane) 설치

#### Cilium CNI 환경에서의 고려사항

Cilium을 CNI(Container Network Interface)로 사용하는 환경에서는 Istio의 기본 트래픽 리디렉션 방식과 충돌이 발생할 수 있습니다.

-   **기본 방식 (`istio-init`):** 각 애플리케이션 파드에 `initContainer`를 추가하여 `iptables` 규칙을 설정합니다. 이 방식은 `NET_ADMIN`과 같은 높은 권한을 요구하며, eBPF를 통해 `iptables`를 우회하는 Cilium과 충돌할 수 있습니다.
-   **권장 방식 (`Istio CNI` 플러그인):** `istio-cni`는 Cilium을 대체하는 것이 아닌, 함께 동작하는 CNI **플러그인**입니다. `DaemonSet`으로 각 노드에 배포되어 파드의 네트워크 설정을 직접 처리합니다.

**Istio CNI 사용의 장점:**
-   **보안 강화:** 애플리케이션 파드에 `NET_ADMIN` 권한이 필요 없어집니다.
-   **충돌 방지:** Cilium과의 `iptables` 충돌 문제를 원천적으로 방지합니다.
-   **성능 향상:** 파드 시작 시 `initContainer` 실행 단계가 없어지므로 시작 시간이 단축됩니다.

따라서 Cilium 환경에서는 **Istio CNI를 활성화하는 것이 강력히 권장됩니다.**

#### 설치 명령어

`--set cni.enabled=true` 옵션을 추가하여 Istio CNI 플러그인을 활성화합니다.

```sh
helm install istiod istio/istiod -n istio-system --set cni.enabled=true --wait
```

만약 이미 `istiod`를 설치했다면, 아래 `helm upgrade` 명령어로 CNI를 활성화할 수 있습니다.

```sh
helm upgrade istiod istio/istiod -n istio-system --set cni.enabled=true --wait
```

### 5. Istio Ingress Gateway 설치

외부 트래픽을 서비스 메쉬 내부로 라우팅하기 위한 인그레스 게이트웨이를 설치합니다.

```sh
helm install istio-ingress istio/gateway -n istio-ingress --wait
```

### 6. 설치 확인

`istiod`, `istio-cni-node`, `istio-ingressgateway` 파드가 모두 정상적으로 실행 중인지 확인합니다.

```sh
# istiod 컨트롤 플레인 및 CNI 파드 확인
kubectl get pods -n istio-system

# istio-ingressgateway 파드 확인
kubectl get pods -n istio-ingress
```


## 사이드카 주입 활성화

Istio를 적용할 Namespace에 `istio-injection=enabled` 레이블을 추가하여 자동 사이드카 주입을 활성화합니다.

예시 (default Namespace에 적용):

```sh
kubectl label namespace default istio-injection=enabled
```
