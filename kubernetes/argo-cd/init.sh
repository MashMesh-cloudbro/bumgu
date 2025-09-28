helm dependency build 
helm template argocd . --namespace argo | kubectl apply -f -
