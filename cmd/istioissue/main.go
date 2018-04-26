package main

import (
	"context"
	"os/exec"
	"strings"
	"net"
	"net/http"
	"time"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"crypto/tls"
	"path/filepath"
	"os"
	"regexp"
	"github.com/dpacierpnik/istio-0.7-eds-issue/pkg/istioissue/testscenario"
)

const (
	namespace                   = "default"
	hostnameFormat              = "sample-app-%s.test.local"
	hostnamePattern             = "sample-app-(.*).test.local"
	maxRetries                  = 1000
	retrySleep                  = 3 * time.Second
	ingressControllerServiceURL = "istio-ingress.istio-system.svc.cluster.local"
)

func main() {

	httpClient := newHttpClientOrExit()

	kubeConfig := defaultConfigOrExit()
	k8sInterface := k8sInterfaceOrExit(kubeConfig)

	ts := testscenario.New(
		httpClient,
		k8sInterface,
		namespace,
		hostnameFormat,
		retrySleep,
		maxRetries)

	ts.Run()
}

func newHttpClientOrExit() *http.Client {

	hostnameRegexp, err := regexp.Compile(hostnamePattern)
	if err != nil {
		log.Fatalf("Error while compiling hostname pattern: '%s'", hostnamePattern)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	ingressControllerAddr, err := net.LookupHost(ingressControllerServiceURL)
	if err != nil {
		log.Debugf("Unable to resolve host '%s'. Root cause: %v", ingressControllerServiceURL, err)
		minikubeIp := tryToGetMinikubeIp()
		ingressControllerAddr = []string{minikubeIp}
	}
	log.Infof("Ingress controller addresses: '%s'", ingressControllerAddr)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {

			log.Debugf("'Resolving '%s'...", addr)
			hostname := hostnameRegexp.FindString(addr)
			if hostname != "" {
				addr = strings.Replace(addr, hostname, ingressControllerAddr[0], 1)
			}
			log.Debugf("'Resolved: '%s'", addr)
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	return client
}

func tryToGetMinikubeIp() string {
	mipCmd := exec.Command("minikube", "ip")
	if mipOut, err := mipCmd.Output(); err != nil {
		log.Warnf("Error while getting minikube IP. Root cause: %s", err)
		return "127.0.0.1"
	} else {
		return strings.Trim(string(mipOut), "\n")
	}
}

func defaultConfigOrExit() *rest.Config {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		log.Debugf("unable to load local kube config. Root cause: %v", err)
		if config, err2 := rest.InClusterConfig(); err2 != nil {
			log.Fatalf("unable to load kube config. Root cause: %v", err2)
		} else {
			kubeConfig = config
		}
	}
	return kubeConfig
}

func k8sInterfaceOrExit(kubeConfig *rest.Config) kubernetes.Interface {

	k8sInterface, k8sErr := kubernetes.NewForConfig(kubeConfig)
	if k8sErr != nil {
		log.Fatalf("can create k8s clientset. Root cause: %v", k8sErr)
	}
	return k8sInterface
}
