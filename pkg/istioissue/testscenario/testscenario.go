package testscenario

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	k8sApps "k8s.io/api/apps/v1"
	k8sCore "k8s.io/api/core/v1"
	k8sExts "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"math/rand"
	"net/http"
	"time"

	istioConfigApi "github.com/dpacierpnik/istio-0.7-eds-issue/pkg/apis/config.istio.io/v1alpha2"
	istioConfig "github.com/dpacierpnik/istio-0.7-eds-issue/pkg/clients/config.istio.io/clientset/versioned"
)

const (
	testIdLength   = 8
	istioNamespace = "istio-system"
)

type HttpInterface interface {
	Get(url string) (resp *http.Response, err error)
}

type TestScenario struct {
	httpClient           *http.Client
	k8sInterface         kubernetes.Interface
	istioConfigInterface istioConfig.Interface
	namespace            string
	hostnameFormat       string
	retrySleep           time.Duration
	maxRetries           int
}

func New(httpClient *http.Client, k8sInterface kubernetes.Interface, istioConfigInterface istioConfig.Interface, namespace string, hostnameFormat string, retrySleep time.Duration, maxRetries int) *TestScenario {
	return &TestScenario{
		httpClient:           httpClient,
		k8sInterface:         k8sInterface,
		istioConfigInterface: istioConfigInterface,
		namespace:            namespace,
		hostnameFormat:       hostnameFormat,
		retrySleep:           retrySleep,
		maxRetries:           maxRetries,
	}
}

func (s *TestScenario) Run() {

	log.SetLevel(log.InfoLevel)

	testsCounter := 0
	failuresCounter := 0

	// repeat foreaver
	for true {

		testsCounter++
		t := newTest(s, testsCounter)

		err := t.run()

		if err != nil {
			failuresCounter++
			log.Warnf("### Found issue! (failures count: %d / %d)", failuresCounter, testsCounter)
		}
	}
}

type test struct {
	testNo   int
	testId   string
	scenario *TestScenario
}

type testCleanup struct {
	test       *test
	deployment *k8sApps.Deployment
	service    *k8sCore.Service
	ingress    *k8sExts.Ingress
	jwtRule    *istioConfigApi.Rule
}

func newTest(s *TestScenario, testNo int) *test {

	testId := generateTestId(testIdLength)

	return &test{
		testNo:   testNo,
		testId:   testId,
		scenario: s,
	}
}

func (t *test) run() error {

	log.Infof("[%d] Running test: '%s'", t.testNo, t.testId)
	tc := testCleanup{test: t}
	defer tc.run()

	deployment := t.createDeploymentOrExit()
	tc.deployment = deployment

	service := t.createServiceOrExit(deployment)
	tc.service = service

	ingress := t.createIngressOrExit(service)
	tc.ingress = ingress

	jwtRule := t.createJwtRuleOrExit()
	tc.jwtRule = jwtRule

	err := t.callWithRetries(ingress)

	log.Infof("Test '%s' failed: %v", t.testId, err)
	log.Info("############################################################")
	return err
}

func (c *testCleanup) run() {

	if c.jwtRule != nil {
		delErr := c.test.scenario.istioConfigInterface.ConfigV1alpha2().Rules(istioNamespace).Delete(c.jwtRule.Name, &k8sMeta.DeleteOptions{})
		if delErr != nil {
			log.Warnf("Test '%s' - can not delete rule '%s'. Root cause: %+v", c.test.testId, c.jwtRule.Name, delErr)
		}
	}

	if c.ingress != nil {
		delErr := c.test.scenario.k8sInterface.ExtensionsV1beta1().Ingresses(c.test.scenario.namespace).Delete(c.ingress.Name, &k8sMeta.DeleteOptions{})
		if delErr != nil {
			log.Warnf("Test '%s' - can not delete ingress '%s'. Root cause: %+v", c.test.testId, c.ingress.Name, delErr)
		}
	}

	if c.service != nil {
		delErr := c.test.scenario.k8sInterface.CoreV1().Services(c.test.scenario.namespace).Delete(c.service.Name, &k8sMeta.DeleteOptions{})
		if delErr != nil {
			log.Warnf("Test '%s' - can not delete service '%s'. Root cause: %+v", c.test.testId, c.service.Name, delErr)
		}
	}

	if c.deployment != nil {
		delErr := c.test.scenario.k8sInterface.AppsV1().Deployments(c.test.scenario.namespace).Delete(c.deployment.Name, &k8sMeta.DeleteOptions{})
		if delErr != nil {
			log.Warnf("Test '%s' - can not delete deployment '%s'. Root cause: %+v", c.test.testId, c.deployment.Name, delErr)
		}
	}
}

func (t *test) createDeploymentOrExit() *k8sApps.Deployment {

	labels := stdLabels(t.testId)

	podAnnotations := make(map[string]string)
	podAnnotations["sidecar.istio.io/inject"] = "true"

	deployment := &k8sApps.Deployment{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      fmt.Sprintf("sample-app-%s", t.testId),
			Namespace: t.scenario.namespace,
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
							Name:            fmt.Sprintf("sample-app-cont-%s", t.testId),
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
					ImagePullSecrets: []k8sCore.LocalObjectReference{
						{
							Name: "team-sf-artifactory-user",
						},
					},
				},
			},
		},
	}

	log.Debugf("Test '%s' - creating deployment: %+v", t.testId, deployment)

	result, createErr := t.scenario.k8sInterface.AppsV1().Deployments(t.scenario.namespace).Create(deployment)
	if createErr != nil {

		if errors.IsAlreadyExists(createErr) {

			log.Debugf("Test '%s' - deployment already exists. Trying to get it...", t.testId)
			tempResult, getErr := t.scenario.k8sInterface.AppsV1().Deployments(t.scenario.namespace).Get(deployment.GetName(), k8sMeta.GetOptions{})
			if getErr != nil {
				log.Fatalf("Error getting existing deployment. Root cause: %v", getErr)
			} else {
				result = tempResult
			}
		} else {
			log.Fatalf("Error creating deployment. Root cause: %v", createErr)
		}
	}

	log.Debugf("Test '%s' - deployment created", t.testId)
	return result
}

func (t *test) createServiceOrExit(deployment *k8sApps.Deployment) *k8sCore.Service {

	labels := stdLabels(t.testId)

	podTmpl := deployment.Spec.Template

	selectors := make(map[string]string)
	selectors["app"] = podTmpl.ObjectMeta.Labels["app"]

	svc := &k8sCore.Service{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      fmt.Sprintf("sample-app-%s", t.testId),
			Namespace: t.scenario.namespace,
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

	log.Debugf("Test '%s' - creating service: %+v", t.testId, svc)

	result, createErr := t.scenario.k8sInterface.CoreV1().Services(t.scenario.namespace).Create(svc)
	if createErr != nil {

		if errors.IsAlreadyExists(createErr) {

			log.Debugf("Test '%s' - service already exists. Trying to get it...", t.testId)
			tempResult, getErr := t.scenario.k8sInterface.CoreV1().Services(t.scenario.namespace).Get(svc.GetName(), k8sMeta.GetOptions{})
			if getErr != nil {
				log.Fatalf("Error getting existing service. Root cause: %v", getErr)
			} else {
				result = tempResult
			}
		} else {
			log.Fatalf("Error creating service. Root cause: %v", createErr)
		}
	}

	log.Debugf("Test '%s' - service created", t.testId)
	return result
}

func (t *test) createIngressOrExit(service *k8sCore.Service) *k8sExts.Ingress {

	labels := stdLabels(t.testId)

	annotations := make(map[string]string)
	annotations["kubernetes.io/ingress.class"] = "istio"

	ing := &k8sExts.Ingress{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:        fmt.Sprintf("sample-app-%s", t.testId),
			Namespace:   t.scenario.namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: k8sExts.IngressSpec{
			Rules: []k8sExts.IngressRule{
				{
					Host: t.hostnameFor(t.testId),
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

	log.Debugf("Test '%s' - creating ingress: %+v", t.testId, ing)

	result, createErr := t.scenario.k8sInterface.ExtensionsV1beta1().Ingresses(t.scenario.namespace).Create(ing)

	if createErr != nil {

		if errors.IsAlreadyExists(createErr) {

			log.Debugf("Test '%s' - ingress already exists. Trying to get it...", t.testId)
			tempResult, getErr := t.scenario.k8sInterface.ExtensionsV1beta1().Ingresses(t.scenario.namespace).Get(ing.GetName(), k8sMeta.GetOptions{})
			if getErr != nil {
				log.Fatalf("Error getting existing ingress. Root cause: %v", getErr)
			} else {
				result = tempResult
			}
		} else {
			log.Fatalf("Error creating ingress. Root cause: %v", createErr)
		}
	}

	log.Debugf("Test '%s' - ingress created.", t.testId)
	return result
}

func (t *test) createJwtRuleOrExit() *istioConfigApi.Rule {

	objectMetadata := k8sMeta.ObjectMeta{
		Name:      "test-jwt-rule",
		Namespace: "istio-system",
	}

	spec := &istioConfigApi.RuleSpec{
		Match: fmt.Sprintf(`destination.service == "%s.%s.svc.cluster.local"`, fmt.Sprintf("sample-app-%s", t.testId), t.scenario.namespace),
		Actions: []*istioConfigApi.Action{
			{
				Handler:   "handler.jwt",
				Instances: []string{"jwt.auth.istio-system"},
			},
		},
	}

	rule := &istioConfigApi.Rule{
		ObjectMeta: objectMetadata,
		Spec:       spec,
	}

	result, createErr := t.scenario.istioConfigInterface.ConfigV1alpha2().Rules(istioNamespace).Create(rule)

	if createErr != nil {

		if errors.IsAlreadyExists(createErr) {

			log.Debugf("Test '%s' - JWT rule already exists. Trying to get it...", t.testId)
			tempResult, getErr := t.scenario.istioConfigInterface.ConfigV1alpha2().Rules(istioNamespace).Get(rule.Name, k8sMeta.GetOptions{})
			if getErr != nil {
				log.Fatalf("Error getting existing JWT rule. Root cause: %v", getErr)
			} else {
				result = tempResult
			}
		} else {
			log.Fatalf("Error creating JWT rule. Root cause: %v", createErr)
		}
	}

	log.Debugf("Test '%s' - JWT rule created.", t.testId)
	return result
}

func (t *test) hostnameFor(testId string) string {
	return fmt.Sprintf(t.scenario.hostnameFormat, testId)
}

func (t *test) callWithRetries(ingress *k8sExts.Ingress) error {

	hostname := ingress.Spec.Rules[0].Host

	log.Debugf("Test '%s' - calling: '%s'", t.testId, hostname)

	response, err := t.withRetries(func() (*http.Response, error) {
		return t.scenario.httpClient.Get(fmt.Sprintf("http://%s/headers", hostname))
	}, httpForbiddenPredicate)

	if err != nil {
		return fmt.Errorf("can not call '%s'. Root cause: %v", hostname, err)
	}

	failed := httpForbiddenPredicate(response)

	if failed {
		return fmt.Errorf("can not call '%s'. Response status: %s", hostname, response.Status)
	}
	return nil
}

func (t *test) withRetries(httpCall func() (*http.Response, error), shouldRetryPredicate func(*http.Response) bool) (*http.Response, error) {

	var response *http.Response
	var err error

	retry := true
	for retryNo := 0; retry; retryNo++ {

		log.Debugf("[%d / %d] Retrying...", retryNo, t.scenario.maxRetries)
		response, err = httpCall()

		if err != nil {
			log.Errorf("[%d / %d] Got error: %s", retryNo, t.scenario.maxRetries, err)
		} else if shouldRetryPredicate(response) {
			log.Errorf("[%d / %d] Got response: %s", retryNo, t.scenario.maxRetries, response.Status)
		} else {
			log.Infof("No more retries (got expected response).")
			retry = false
		}

		if retry {

			if retryNo >= t.scenario.maxRetries {
				// do not retry anymore
				log.Infof("No more retries (max retries exceeded).")
				retry = false
			} else {
				time.Sleep(t.scenario.retrySleep)
			}
		}
	}

	return response, err
}

func httpNotOkPredicate(response *http.Response) bool {
	return response.StatusCode < 200 || response.StatusCode > 299
}

func httpForbiddenPredicate(response *http.Response) bool {
	return response.StatusCode != 403
}

func generateTestId(n int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func stdLabels(testId string) map[string]string {
	labels := make(map[string]string)
	labels["app"] = fmt.Sprintf("sample-app-%s", testId)
	return labels
}
