package testscenario

import (
	log "github.com/sirupsen/logrus"
	k8sApps "k8s.io/api/apps/v1"
	k8sCore "k8s.io/api/core/v1"
	k8sExts "k8s.io/api/extensions/v1beta1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"net/http"
	"time"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

const (
	scenarioIdLength = 8
)

type Config struct {
	Namespace                string
	HostnameFormat           string
	RetryDelay               time.Duration
	MaxRetries               int
	NumberOfResourcesPerTest int
	OperationDelay           time.Duration
	MinOkResponsesToSucceed  int
}

func Run(httpClient *http.Client, k8sInterface kubernetes.Interface, config *Config) {

	log.SetLevel(log.InfoLevel)

	scenarioNo := 0
	failureNo := 0

	// repeat foreaver
	for ; true; {

		scenarioNo++
		s := newScenario(httpClient, k8sInterface, config, scenarioNo)
		err := s.run()
		if err != nil {
			failureNo++
			log.Warnf("### Found issue! (failures count: %d / %d)", failureNo, scenarioNo)
		}
	}
}

type scenario struct {
	httpClient   *http.Client
	k8sInterface kubernetes.Interface
	config       *Config
	no           int
	id           string
}

func newScenario(httpClient *http.Client, k8sInterface kubernetes.Interface, config *Config, scenarioNo int) *scenario {

	scenarioId := generateId(scenarioIdLength)

	return &scenario{
		httpClient:   httpClient,
		k8sInterface: k8sInterface,
		config:       config,
		no:           scenarioNo,
		id:           scenarioId,
	}
}

func (s *scenario) run() error {

	log.Info("############################################################")
	log.Infof("[%d] Running scenario: '%s'...", s.no, s.id)
	tc := newCleanup(s, s.config.NumberOfResourcesPerTest)
	defer tc.run()

	var lastIngress *k8sExts.Ingress

	for i := 0; i < s.config.NumberOfResourcesPerTest; i++ {

		resourceId := fmt.Sprintf("sample-app-%s-%d", s.id, i)

		log.Infof("Scenario '%s' - creating resources: '%s'...", s.id, resourceId)

		deployment := s.createDeploymentOrExit(resourceId)
		tc.deployments[i] = deployment

		time.Sleep(s.config.OperationDelay)

		service := s.createServiceOrExit(deployment, resourceId)
		tc.services[i] = service

		time.Sleep(s.config.OperationDelay)

		ingress := s.createIngressOrExit(service, resourceId)
		tc.ingresses[i] = ingress

		time.Sleep(s.config.OperationDelay)

		lastIngress = ingress
	}

	err := s.callWithRetries(lastIngress)
	if err == nil {
		log.Infof("Scenario '%s' - SUCCEED", s.id)
	} else {
		log.Infof("Scenario '%s' - FAILED with: %v", s.id, err)
	}
	return err
}

func (s *scenario) createDeploymentOrExit(resourceId string) *k8sApps.Deployment {

	labels := s.stdLabels(resourceId)

	podAnnotations := make(map[string]string)
	podAnnotations["sidecar.istio.io/inject"] = "true"

	deployment := &k8sApps.Deployment{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      resourceId,
			Namespace: s.config.Namespace,
			Labels:    labels,
		},
		Spec: k8sApps.DeploymentSpec{
			Selector: &k8sMeta.LabelSelector{
				MatchLabels: labels,
			},
			Template: k8sCore.PodTemplateSpec{
				ObjectMeta: k8sMeta.ObjectMeta{
					Labels:      labels,
					Annotations: podAnnotations,
				},
				Spec: k8sCore.PodSpec{
					Containers: []k8sCore.Container{
						{
							Name:            "httpbin",
							Image:           "docker.io/citizenstig/httpbin",
							ImagePullPolicy: "IfNotPresent",
							Ports: []k8sCore.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8000,
								},
							},
						},
					},
				},
			},
		},
	}

	log.Debugf("Scenario '%s' - creating deployment: %+v", s.id, deployment)

	result, createErr := s.k8sInterface.AppsV1().Deployments(s.config.Namespace).Create(deployment)
	if createErr != nil {

		if errors.IsAlreadyExists(createErr) {

			log.Debugf("Scenario '%s' - deployment already exists. Trying to get is...", s.id)
			tempResult, getErr := s.k8sInterface.AppsV1().Deployments(s.config.Namespace).Get(deployment.GetName(), k8sMeta.GetOptions{})
			if getErr != nil {
				log.Fatalf("Error getting existing deployment. Root cause: %v", getErr)
			} else {
				result = tempResult
			}
		} else {
			log.Fatalf("Error creating deployment. Root cause: %v", createErr)
		}
	}

	log.Debugf("Scenario '%s' - deployment created", s.id)
	return result
}

func (s *scenario) createServiceOrExit(deployment *k8sApps.Deployment, resourceId string) *k8sCore.Service {

	labels := s.stdLabels(resourceId)

	podTmpl := deployment.Spec.Template

	selectors := make(map[string]string)
	selectors["app"] = podTmpl.ObjectMeta.Labels["app"]

	svc := &k8sCore.Service{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      resourceId,
			Namespace: s.config.Namespace,
			Labels:    labels,
		},
		Spec: k8sCore.ServiceSpec{
			Selector: selectors,
			Type:     k8sCore.ServiceTypeClusterIP,
			Ports: []k8sCore.ServicePort{
				{
					Port:       8888,
					Name:       "http",
					TargetPort: intstr.FromString(podTmpl.Spec.Containers[0].Ports[0].Name),
				},
			},
		},
	}

	log.Debugf("Scenario '%s' - creating service: %+v", s.id, svc)

	result, createErr := s.k8sInterface.CoreV1().Services(s.config.Namespace).Create(svc)
	if createErr != nil {

		if errors.IsAlreadyExists(createErr) {

			log.Debugf("Scenario '%s' - service already exists. Trying to get is...", s.id)
			tempResult, getErr := s.k8sInterface.CoreV1().Services(s.config.Namespace).Get(svc.GetName(), k8sMeta.GetOptions{})
			if getErr != nil {
				log.Fatalf("Error getting existing service. Root cause: %v", getErr)
			} else {
				result = tempResult
			}
		} else {
			log.Fatalf("Error creating service. Root cause: %v", createErr)
		}
	}

	log.Debugf("Scenario '%s' - service created", s.id)
	return result
}

func (s *scenario) createIngressOrExit(service *k8sCore.Service, resourceId string) *k8sExts.Ingress {

	labels := s.stdLabels(resourceId)

	annotations := make(map[string]string)
	annotations["kubernetes.io/ingress.class"] = "istio"

	ing := &k8sExts.Ingress{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:        resourceId,
			Namespace:   s.config.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: k8sExts.IngressSpec{
			Rules: []k8sExts.IngressRule{
				{
					Host: s.hostnameFor(resourceId),
					IngressRuleValue: k8sExts.IngressRuleValue{
						HTTP: &k8sExts.HTTPIngressRuleValue{
							Paths: []k8sExts.HTTPIngressPath{
								{
									Path: "/.*",
									Backend: k8sExts.IngressBackend{
										ServiceName: service.Name,
										ServicePort: intstr.FromInt(int(service.Spec.Ports[0].Port)),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	log.Debugf("Scenario '%s' - creating ingress: %+v", s.id, ing)

	result, createErr := s.k8sInterface.ExtensionsV1beta1().Ingresses(s.config.Namespace).Create(ing)

	if createErr != nil {

		if errors.IsAlreadyExists(createErr) {

			log.Debugf("Scenario '%s' - ingress already exists. Trying to get is...", s.id)
			tempResult, getErr := s.k8sInterface.ExtensionsV1beta1().Ingresses(s.config.Namespace).Get(ing.GetName(), k8sMeta.GetOptions{})
			if getErr != nil {
				log.Fatalf("Error getting existing ingress. Root cause: %v", getErr)
			} else {
				result = tempResult
			}
		} else {
			log.Fatalf("Error creating ingress. Root cause: %v", createErr)
		}
	}

	log.Debugf("Scenario '%s' - ingress created.", s.id)
	return result
}

type scenarioCleanup struct {
	scenario    *scenario
	deployments []*k8sApps.Deployment
	services    []*k8sCore.Service
	ingresses   []*k8sExts.Ingress
}

func newCleanup(scenario *scenario, resourcesCount int) *scenarioCleanup {

	tc := &scenarioCleanup{
		scenario: scenario,
	}
	tc.deployments = make([]*k8sApps.Deployment, resourcesCount)
	tc.services = make([]*k8sCore.Service, resourcesCount)
	tc.ingresses = make([]*k8sExts.Ingress, resourcesCount)
	return tc
}

func (c *scenarioCleanup) run() {

	log.Infof("Scenario '%s' - cleanup...", c.scenario.id)

	for _, ingress := range c.ingresses {
		delErr := c.scenario.k8sInterface.ExtensionsV1beta1().Ingresses(c.scenario.config.Namespace).Delete(ingress.Name, &k8sMeta.DeleteOptions{})
		if delErr != nil {
			log.Warnf("Scenario '%s' - can not delete ingress '%s'. Root cause: %+v", c.scenario.id, ingress.Name, delErr)
		}
		time.Sleep(c.scenario.config.OperationDelay)
	}

	for _, service := range c.services {
		delErr := c.scenario.k8sInterface.CoreV1().Services(c.scenario.config.Namespace).Delete(service.Name, &k8sMeta.DeleteOptions{})
		if delErr != nil {
			log.Warnf("Scenario '%s' - can not delete service '%s'. Root cause: %+v", c.scenario.id, service.Name, delErr)
		}
		time.Sleep(c.scenario.config.OperationDelay)
	}

	for _, deployment := range c.deployments {
		delErr := c.scenario.k8sInterface.AppsV1().Deployments(c.scenario.config.Namespace).Delete(deployment.Name, &k8sMeta.DeleteOptions{})
		if delErr != nil {
			log.Warnf("Scenario '%s' - can not delete deployment '%s'. Root cause: %+v", c.scenario.id, deployment.Name, delErr)
		}
		time.Sleep(c.scenario.config.OperationDelay)
	}

	log.Debugf("Scenario '%s' - cleanup done.", c.scenario.id)
}

func (s *scenario) stdLabels(resourceId string) map[string]string {
	labels := make(map[string]string)
	labels["scenario.id"] = s.id
	labels["scenario.no"] = strconv.Itoa(s.no)
	labels["resource.id"] = resourceId
	labels["app"] = fmt.Sprintf("sample-app-%s", resourceId)
	return labels
}

func (s *scenario) hostnameFor(resourceId string) string {
	return fmt.Sprintf(s.config.HostnameFormat, resourceId)
}

func (s *scenario) callWithRetries(ingress *k8sExts.Ingress) error {

	hostname := ingress.Spec.Rules[0].Host

	log.Debugf("Scenario '%s' - calling: '%s'", s.id, hostname)

	response, err := s.withRetries(func() (*http.Response, error) {
		return s.httpClient.Get(fmt.Sprintf("http://%s/headers", hostname))
	}, httpNotOkPredicate, s.config.MaxRetries, s.config.RetryDelay)

	if err != nil {
		return fmt.Errorf("can not call '%s'. Root cause: %v", hostname, err)
	}

	failed := httpNotOkPredicate(response)

	if failed {
		return fmt.Errorf("can not call '%s'. Response status: %s", hostname, response.Status)
	} else if s.config.MinOkResponsesToSucceed > 1 {

		nextResponse, nextErr := s.withRetries(func() (*http.Response, error) {
			return s.httpClient.Get(fmt.Sprintf("http://%s/headers", hostname))
		}, httpOkPredicate, s.config.MinOkResponsesToSucceed, s.config.RetryDelay)

		if nextErr != nil {
			return fmt.Errorf("can not call '%s'. Root cause: %v", hostname, nextErr)
		}

		failedAgain := httpNotOkPredicate(nextResponse)
		if failedAgain {
			return fmt.Errorf(" can not call '%s' AGAIN. Response status: %s", hostname, response.Status)
		}
	}
	return nil
}

func (s *scenario) withRetries(httpCall func() (*http.Response, error), shouldRetryPredicate func(*http.Response) bool,
	maxRetries int, retryDelay time.Duration) (*http.Response, error) {

	var response *http.Response
	var err error

	retry := true
	for retryNo := 0; retry; retryNo++ {

		log.Debugf("[%d / %d] Retrying...", retryNo, maxRetries)
		response, err = httpCall()

		if err != nil {
			log.Errorf("[%d / %d] Got error: %s", retryNo, maxRetries, err)
		} else if shouldRetryPredicate(response) {
			log.Infof("[%d / %d] Got response: %s", retryNo, maxRetries, response.Status)
		} else {
			log.Infof("No more retries (got expected response).")
			retry = false
		}

		if retry {

			if retryNo >= maxRetries {
				// do not retry anymore
				log.Infof("No more retries (max retries exceeded).")
				retry = false
			} else {
				time.Sleep(s.config.RetryDelay)
			}
		}
	}

	return response, err
}

func httpNotOkPredicate(response *http.Response) bool {
	return !httpOkPredicate(response)
}

func httpOkPredicate(response *http.Response) bool {
	return response.StatusCode >= 200 && response.StatusCode < 300
}

func generateId(n int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
