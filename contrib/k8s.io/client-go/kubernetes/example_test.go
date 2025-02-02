package kubernetes_test

import (
	"fmt"

	kubernetestrace "github.com/adityayuga/signalfx-go-tracing/contrib/k8s.io/client-go/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
)

func Example() {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// Use this to trace all calls made to the Kubernetes API
	cfg.WrapTransport = kubernetestrace.WrapRoundTripper

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	pods, err := client.CoreV1().Pods("default").List(meta_v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println(pods.Items)
}
