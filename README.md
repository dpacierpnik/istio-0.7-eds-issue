# Istio 0.7.1 issue with EDS - envoy of the app never knows about ingress

## Summary

## Pre-requisites:

- Minikube 0.25.0 or newer installed
- Kubectl 1.7.3 or newer installed
- Golang 1.9 or newer installed

## How to run

1. Clone this repository under your GOPATH src:

   ```bash
   export ISTIO_ISSUE_PATH=$GOPATH/src/dpacierpnik
   mkdir -p $ISTIO_ISSUE_PATH
   cd $ISTIO_ISSUE_PATH
   git clone $this_repo
   ```

1. Go to the issue directory:

   ```bash
   cd istio-0.7-eds-issue
   ```

1. Run minikube (0.25.0 or later):

   ```bash
   minikube start \
	  --extra-config=controller-manager.ClusterSigningCertFile="/var/lib/localkube/certs/ca.crt" \
	  --extra-config=controller-manager.ClusterSigningKeyFile="/var/lib/localkube/certs/ca.key" \
	  --extra-config=apiserver.Admission.PluginNames=NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota \
	  --kubernetes-version=v1.9.0 \
      --extra-config=apiserver.ServiceNodePortRange=79-36000
   ```

1. Install Istio:

   ```bash
   ./install-istio.sh
   ```
      
1. Run application with test scenario:

   1. Ensure dependencies, and build app:

      ```bash
      dep ensure
      go build ./...
      ```
      
   1. Run:
   
      ```bash
      go run cmd/istioissue/main.go
      ```