Logs:

- App.log - logs from the httpbin application. Result of the following command:

  ```bash
  kubectl logs sample-app-lgimuwvn-29-5867cb85b5-wlfp9 httpbin > App.log
  ```

- App_pods_running.log - screenshot with running application pods.

- App_Sidecar.log - logs from the sidecar of httpbin application. Result of the following command:
                                                                
  ```bash
  kubectl logs sample-app-lgimuwvn-29-5867cb85b5-wlfp9 istio-proxy > App_Sidecar.log
  ```

- App_Sidecar_routes.log - routes from the sidecar of httpbin application. Result of the following command:
  
  ```bash
  kubectl port-forward sample-app-lgimuwvn-29-5867cb85b5-wlfp9 15000:15000 -n istio-system
  curl http://localhost:15000/routes > App_Sidecar_routes.log
  ```

- Test.log - logs from the test execution. Shows that application is not available (Ingress never knows routing). Result of the following command:

  ```bash
  go run cmd/istioissue/main.go
  ```
 
- Istio_IngressController.log - logs from the Istio Ingress Controller. Result of the following command:
  
  ```bash
  kubectl logs istio-ingress-5bb556fcbf-h9lxf -n istio-system > Istio_Ingress_Controller.log
  ```
 
- Istio_IngressController_routes.log - routes from the Istio Ingress Controller. Result of the following command:
  
  ```bash
  kubectl port-forward istio-ingress-5bb556fcbf-h9lxf 15001:15000 -n istio-system
  curl http://localhost:15001/routes > Istio_IngressController_routes.log
  ```
 
- Istio_Pilot.log - logs from the Istio Pilot. Result of the following command:
  
  ```bash
  kubectl logs istio-pilot-67d6ddbdf6-wjls9 discovery -n istio-system > Istio_Pilot.log
  ```

- Istio_pods_running.log - screenshot with running Istio pods.
