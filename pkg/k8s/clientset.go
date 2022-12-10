package k8s

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/eks"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

type Clientset struct {
	*kubernetes.Clientset
}

func NewClientset(path string) (*Clientset, error) {
	if path == "" {
		val, present := os.LookupEnv("KUBECONFIG")
		if present {
			path = val
		} else {
			path = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Clientset{
		Clientset: clientset,
	}, nil
}

func NewEKSClientset(name string, region string) (*Clientset, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	eksSvc := eks.New(sess)

	input := &eks.DescribeClusterInput{
		Name: aws.String(name),
	}
	result, err := eksSvc.DescribeCluster(input)
	if err != nil {
		return nil, err
	}
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(result.Cluster.Name),
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(result.Cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(
		&rest.Config{
			Host:        aws.StringValue(result.Cluster.Endpoint),
			BearerToken: tok.Token,
			TLSClientConfig: rest.TLSClientConfig{
				CAData: ca,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return &Clientset{
		Clientset: clientset,
	}, nil
}
