Logs:

- Test.log - logs from the test execution. Shows that applications are not available for a very long time. Result of the following command:

  ```bash
  go run cmd/istioissue/main.go
  ```

- Istio_Pilot_crashing_screen.png - screenshot which shows that Istio Pilot is crashing all the time. Result of the following command:

  ```bash
  kubectl get pods -n istio-system
  ```

- Istio_Pilot.log - logs from the Istio Pilot pod. Result of the following command: 

  ```bash
  kubectl logs istio-pilot-657cb5ddf7-8k2q7 -n istio-system discovery > Istio_Pilot.log
  ```

  In this log you can find a lot of stacktraces related to the Pilot failure.
