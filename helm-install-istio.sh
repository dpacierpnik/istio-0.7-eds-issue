#!/usr/bin/env bash

pushd istio-0.7.1/install/kubernetes/

kubectl apply -f istio.yaml

./webhook-create-signed-cert.sh \
    --service istio-sidecar-injector \
    --namespace istio-system \
    --secret sidecar-injector-certs

kubectl apply -f istio-sidecar-injector-configmap-release.yaml

cat istio-sidecar-injector.yaml | \
     ./webhook-patch-ca-bundle.sh > \
     istio-sidecar-injector-with-ca-bundle.yaml

kubectl apply -f istio-sidecar-injector-with-ca-bundle.yaml

kubectl -n istio-system get deployment -listio=sidecar-injector

kubectl label namespace default istio-injection=enabled

popd
