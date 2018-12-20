package helpers

import (
	"crypto/x509"
	"errors"
	"flag"
	"fmt"

	"github.com/elastic/stack-operators/stack-operator/pkg/apis/deployments/v1alpha1"
	"github.com/elastic/stack-operators/stack-operator/pkg/controller/elasticsearch/client"
	"github.com/elastic/stack-operators/stack-operator/pkg/dev/portforward"
	"github.com/elastic/stack-operators/stack-operator/pkg/utils/net"
)

// if `--auto-port-forward` is passed to `go test`, then use a custom
// dialer that sets up port-forwarding to services running within k8s
// (useful when running tests on a dev env instead of as a batch job)
var autoPortForward = flag.Bool(
	"auto-port-forward", false,
	"enables automatic port-forwarding (for dev use only as it exposes "+
		"k8s resources on ephemeral ports to localhost)")

// NewElasticsearchClient returns an ES client for the given stack's ES cluster
func NewElasticsearchClient(stack v1alpha1.Stack, k *K8sHelper) (*client.Client, error) {
	password, err := k.GetElasticPassword(stack.Name)
	if err != nil {
		return nil, err
	}
	esUser := client.User{
		Name:     "elastic",
		Password: password,
	}
	caCert, err := k.GetCACert(stack.Name)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, errors.New("Cannot append CA cert to cert pool")
	}
	inClusterURL := fmt.Sprintf("https://%s-es-public.%s.svc.cluster.local:9200", stack.Name, stack.Namespace)
	var dialer net.Dialer
	if *autoPortForward {
		dialer = portforward.NewForwardingDialer()
	}
	client := client.NewElasticsearchClient(dialer, inClusterURL, esUser, certPool)
	return client, nil
}