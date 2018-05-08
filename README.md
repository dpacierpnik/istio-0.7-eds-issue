# Istio 0.7.1 issues with Pilot

## Summary

This project has been created to reproduce different issues with Istio Pilot, which behaves very unstable in our cluster. 

This repository contains original Istio 0.7.1 with one adjustment - Ingress Controller service configuration has been modified - 
*type* has been set to *NodePort* and *nodePort* has been set to *80*.

Test has been implemented as simple Go application which runs forever, and executes very simple test scenario, 
which consists of the following steps:

1. Create *K8s pod* with httpbin application (using *K8s deployment*)

1. Create *K8s service* for httpbin application

1. Create *K8s ingress* for httpbin application

1. Call httpbin application via *K8s ingress*. Because application may not be available immediately, 
   retry this operation every 3 seconds until either expected response will be returned (http status: 200 OK), 
   or max retries count will be exceeded (1000 attempts).
   If application will not be available after 50 minutes (3 seconds of delay * 1000 attempts = 3000 seconds = 50 minutes), 
   report warning, that the issue has been found.
   
1. When retries are finished, cleanup K8s resources (deployment, service and ingress), and repeat whole scenario again.

## Pre-requisites:

- Minikube 0.25.0 or newer installed
- Hyperkit driver
- Kubectl 1.9 or newer installed
- Golang 1.9 or newer installed

## How to run

1. Clone this repository under your GOPATH src:

   ```bash
   export ISTIO_ISSUE_PATH=$GOPATH/src/github.com/dpacierpnik
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
   minikube start --vm-driver=hyperkit --memory=4096 \
	  --extra-config=controller-manager.ClusterSigningCertFile="/var/lib/localkube/certs/ca.crt" \
	  --extra-config=controller-manager.ClusterSigningKeyFile="/var/lib/localkube/certs/ca.key" \
	  --extra-config=apiserver.Admission.PluginNames=NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota \
	  --kubernetes-version=v1.9.0 \
      --extra-config=apiserver.ServiceNodePortRange=79-36000
   ```

1. Install Istio using appropriate installation script:

   1. Istio with auth disabled (installed using istio.yaml):
   
      ```bash
      ./install-istio.sh
      ```

   1. Istio with auth enabled (installed using istio-auth.yaml):
   
      ```bash
      ./install-istio-auth.sh
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