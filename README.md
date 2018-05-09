# Istio 0.7.1 issues with Pilot

## Summary

This project has been created to reproduce different issues with Istio Pilot, which behaves very unstable in our cluster. 

This repository contains original Istio 0.7.1 with one adjustment - Ingress Controller service configuration has been modified - 
*type* has been set to *NodePort* and *nodePort* has been set to *80*.

## Pre-requisites:

1. Install the following tools:

   - Golang 1.9 or newer
   - Kubectl 1.9 or newer
   - Minikube 0.25.0 or newer
   - Hyperkit driver

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

1. Ensure dependencies, and build app:

   ```bash
   dep ensure
   go build ./...
   ```

1. Run minikube:

   ```bash
   minikube start --vm-driver=hyperkit --memory=4096 \
	  --extra-config=controller-manager.ClusterSigningCertFile="/var/lib/localkube/certs/ca.crt" \
	  --extra-config=controller-manager.ClusterSigningKeyFile="/var/lib/localkube/certs/ca.key" \
	  --extra-config=apiserver.Admission.PluginNames=NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota \
	  --kubernetes-version=v1.9.0 \
      --extra-config=apiserver.ServiceNodePortRange=79-36000
   ```

## Test scenarios

Test has been implemented as simple Go application which runs forever, and executes very simple test scenario.

### 01 - Failing Pilot (auth disabled)

#### Test scenario

1. Create *K8s pod* with httpbin application (using *K8s deployment*)

1. Create *K8s service* for this application deployment

1. Create *K8s ingress* for this application deployment

1. Call httpbin application via *K8s ingress*. Because application may not be available immediately, 
   retry this operation every 3 seconds until either expected response will be returned (http status: 200 OK), 
   or max retries count will be exceeded (1000 attempts).
   If application will not be available after 50 minutes (3 seconds of delay * 1000 attempts = 3000 seconds = 50 minutes), 
   report warning, that the issue has been found.
   
1. When retries are finished, cleanup all K8s resources (deployment, service and ingress), and repeat whole scenario again.

#### How to run

1. Istio with auth disabled (installed using istio.yaml):
   
   ```bash
   ./install-istio.sh
   ```

1. Run the test application:

   ```bash
   go run cmd/istioissue/main.go
   ```

#### Test result

After some time (it is random how long it takes) Pilot fails and is restarted. 
After restart applications are not available anymore (503s are returned when trying to access application via Istio controller).
In the Pilot logs there is a stack trace from goroutine with the following error:

```
fatal error: concurrent map read and map write

goroutine 1409 [running]:
runtime.throw(0x17b9212, 0x21)
	/usr/local/go/src/runtime/panic.go:605 +0x95 fp=0xc420d6a790 sp=0xc420d6a770 pc=0x42bb85
runtime.mapaccess2(0x159fc40, 0xc4202bef90, 0xc42038d480, 0xc4208999d8, 0xc420899a01)
	/usr/local/go/src/runtime/hashmap.go:413 +0x24e fp=0xc420d6a7d8 sp=0xc420d6a790 pc=0x40887e
reflect.mapaccess(0x159fc40, 0xc4202bef90, 0xc42038d480, 0xc4202bef90)
	/usr/local/go/src/runtime/hashmap.go:1218 +0x3f fp=0xc420d6a810 sp=0xc420d6a7d8 pc=0x40b19f
reflect.Value.MapIndex(0x159fc40, 0xc4202bef90, 0x15, 0x1532a00, 0xc42038d480, 0x98, 0x14cd660, 0xc420d78690, 0x16)
	/usr/local/go/src/reflect/value.go:1058 +0x126 fp=0xc420d6a898 sp=0xc420d6a810 pc=0x4ba756
fmt.(*pp).printValue(0xc420384000, 0x159fc40, 0xc4202bef90, 0x15, 0x76, 0x0)
	/usr/local/go/src/fmt/print.go:750 +0xf47 fp=0xc420d6aad8 sp=0xc420d6a898 pc=0x4d0397
fmt.(*pp).printArg(0xc420384000, 0x159fc40, 0xc4202bef90, 0xc400000076)
	/usr/local/go/src/fmt/print.go:682 +0x1e5 fp=0xc420d6ab58 sp=0xc420d6aad8 pc=0x4cec45
fmt.(*pp).doPrintf(0xc420384000, 0x17df892, 0x34, 0xc420d6ae88, 0x3, 0x3)
	/usr/local/go/src/fmt/print.go:996 +0x15a fp=0xc420d6ac88 sp=0xc420d6ab58 pc=0x4d2baa
fmt.Sprintf(0x17df892, 0x34, 0xc420d6ae88, 0x3, 0x3, 0x1532a00, 0xc42001e070)
	/usr/local/go/src/fmt/print.go:196 +0x66 fp=0xc420d6ace0 sp=0xc420d6ac88 pc=0x4cb0e6
istio.io/istio/vendor/go.uber.org/zap.(*SugaredLogger).log(0xc420066538, 0x7facad587c00, 0x17df892, 0x34, 0xc420d6ae88, 0x3, 0x3, 0x0, 0x0, 0x0)
	/workspace/go/src/istio.io/istio/vendor/go.uber.org/zap/sugar.go:230 +0x127 fp=0xc420d6ad30 sp=0xc420d6ace0 pc=0x83dfb7
istio.io/istio/vendor/go.uber.org/zap.(*SugaredLogger).Infof(0xc420066538, 0x17df892, 0x34, 0xc420d6ae88, 0x3, 0x3)
	/workspace/go/src/istio.io/istio/vendor/go.uber.org/zap/sugar.go:138 +0x83 fp=0xc420d6ad90 sp=0xc420d6ad30 pc=0x83d693
istio.io/istio/pkg/log.Infof(0x17df892, 0x34, 0xc420d6ae88, 0x3, 0x3)
	/workspace/go/src/istio.io/istio/pkg/log/log.go:174 +0x5f fp=0xc420d6add0 sp=0xc420d6ad90 pc=0xa6b0bf
istio.io/istio/pilot/pkg/proxy/envoy/v2.(*DiscoveryServer).removeEdsCon(0xc420293960, 0xc42089cf00, 0x3a, 0xc420777da0, 0x5c, 0xc420c321e0)
	/workspace/go/src/istio.io/istio/pilot/pkg/proxy/envoy/v2/eds.go:521 +0x3b2 fp=0xc420d6aec8 sp=0xc420d6add0 pc=0x128ec42
istio.io/istio/pilot/pkg/proxy/envoy/v2.(*DiscoveryServer).StreamEndpoints.func1(0xc420b08d20, 0x2226b40, 0xc420047cf0, 0xc420047d10, 0xc420e04a40, 0x10, 0xc420c321e0, 0xc420293960, 0xc420047d00)
	/workspace/go/src/istio.io/istio/pilot/pkg/proxy/envoy/v2/eds.go:304 +0x247 fp=0xc420d6af98 sp=0xc420d6aec8 pc=0x1290ca7
runtime.goexit()
	/usr/local/go/src/runtime/asm_amd64.s:2337 +0x1 fp=0xc420d6afa0 sp=0xc420d6af98 pc=0x45caf1
created by istio.io/istio/pilot/pkg/proxy/envoy/v2.(*DiscoveryServer).StreamEndpoints
	/workspace/go/src/istio.io/istio/pilot/pkg/proxy/envoy/v2/eds.go:296 +0x2a0
```

Results of the test can be found here: [issue_logs/01_istio_no_auth](./issue_logs/01_istio_no_auth)

### 02 - Failing Pilot (auth enabled)

#### Test scenario

1. Create *K8s pod* with httpbin application (using *K8s deployment*)

1. Create *K8s service* for this application deployment

1. Create *K8s ingress* for this application deployment

1. Call httpbin application via *K8s ingress*. Because application may not be available immediately, 
   retry this operation every 3 seconds until either expected response will be returned (http status: 200 OK), 
   or max retries count will be exceeded (1000 attempts).
   If application will not be available after 50 minutes (3 seconds of delay * 1000 attempts = 3000 seconds = 50 minutes), 
   report warning, that the issue has been found.
   
1. When retries are finished, cleanup all K8s resources (deployment, service and ingress), and repeat whole scenario again.

#### How to run

1. Istio with auth enabled (installed using istio-auth.yaml):

   ```bash
   ./install-istio-auth.sh
   ```

1. Run the test application:

   ```bash
   go run cmd/istioissue/main.go
   ```

#### Test result

After some time (it is random how long it takes) Pilot starts to fail and is restarted all the time.
It causes applications are not available for a long time.

Results of the test can be found here: [issue_logs/02_istio_with_auth](./issue_logs/02_istio_with_auth)

### 03 - Many resources - envoy never knows about application (auth disabled)

#### Test scenario

1. Create many *K8s pods* with httpbin application (using *K8s deployment*)

1. Create *K8s service* for each httpbin application deployment

1. Create *K8s ingress* for each httpbin application deployment

1. Call last deployed httpbin application via *K8s ingress*. Because application may not be available immediately, 
   retry this operation every 3 seconds until either expected response will be returned (http status: 200 OK), 
   or max retries count will be exceeded (1000 attempts).
   If application will not be available after 50 minutes (3 seconds of delay * 1000 attempts = 3000 seconds = 50 minutes), 
   report warning, that the issue has been found.
   
1. When retries are finished, cleanup all K8s resources (deployments, services and ingresses), and repeat whole scenario again.

#### How to run

1. Istio with auth disabled (installed using istio.yaml):
   
   ```bash
   ./install-istio.sh
   ```

1. Run the test application:

   ```bash
   go run cmd/istioissue/main.go --resources-per-test 30
   ```

#### Test result

After few runs, Istio starts to be unstable.
Application is not available at all (*Ingress Controller* returns 404s all the time).
It looks like *Ingress Controller* (*Envoy*) never knows about application. 

Results of the test can be found here: [issue_logs/istio_no_auth_many_resources](./issue_logs/istio_no_auth_many_resources)
