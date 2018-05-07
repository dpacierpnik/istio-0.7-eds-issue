Logs:

- Test.log - logs from the test execution. Shows that after Pilot restart, application is not available anymore. Result of the following command:

  ```bash
  go run cmd/istioissue/main.go
  ```

- Istio_Pilot_restarted_screen.png - screenshot which shows that Istio Pilot has been restarted. Result of the following command:

  ```bash
  kubectl get pods -n istio-system
  ```

- Istio_Pilot_previous.log - logs from the Istio Pilot previous pod (before restart). Result of the following command: 

  ```bash
  kubectl logs istio-pilot-67d6ddbdf6-hqb2l -n istio-system discovery > Istio_Pilot_current.log
  ```

  In this log you can find the following error: `fatal error: concurrent map read and map write`

- Istio_Pilot_current.log - logs from the Istio Pilot current pod (after restart). Result of the following command: 

  ```bash
  kubectl logs istio-pilot-67d6ddbdf6-hqb2l -n istio-system discovery > Istio_Pilot_current.log
  ```
  
- Istio_IngressController.log - logs from the Istio Ingress Controller. Result of the following command:
  
  ```bash
  kubectl logs istio-ingress-5bb556fcbf-pgplh -n istio-system > Istio_Ingress_Controller.log
  ```

- SampleApp_working_fine.png - screenshot which shows that sample application has been successfully deployed and is running fine. Result of the following command:

  ```bash
  kubectl get pods
  ```

- SampleApp_Proxy.log - logs from the sidecar/proxy/envoy of the sample application deployed in the default namespace, and exposed via Ingress Controller. Result of the following command:

  ```bash
  kubectl logs sample-app-zqyepnnn-7f57964b45-w4xxc istio-proxy > SampleApp_Proxy.log
  ```
