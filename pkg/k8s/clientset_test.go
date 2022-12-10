package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
)

var (
	emptyConfig = `
apiVersion: v1
kind: Config
preferences: {}
clusters: []
users: []
contexts: []
current-context: ""
`
	fullConfig = `
apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:443
  name: kind-default
contexts:
- context:
    cluster: kind-default
    user: kind-default
  name: kind-default
current-context: kind-default
kind: Config
preferences: {}
users:
- name: kind-default
`
)

func TestNewClienteset(t *testing.T) {
	emptyF, err := os.CreateTemp("", random.UniqueId())
	if err != nil {
		t.Errorf("error creating temp kubeconfig: %s", err.Error())
	}
	defer os.Remove(emptyF.Name())
	emptyF.Write([]byte(emptyConfig))

	fullF, err := os.CreateTemp("", random.UniqueId())
	if err != nil {
		t.Errorf("error creating temp kubeconfig: %s", err.Error())
	}
	defer os.Remove(fullF.Name())
	fullF.Write([]byte(fullConfig))

	cmd := exec.Command(fmt.Sprintf("cat %s", fullF.Name()))
	_ = cmd.Run()

	tests := []struct {
		path string
		err  bool
	}{
		{"this_path_should_fail", true},
		{emptyF.Name(), true},
		{fullF.Name(), false},
	}

	for _, test := range tests {
		_, err := NewClientset(test.path)
		if err != nil && !test.err {
			t.Errorf("got err: %s, on path: %s", err.Error(), test.path)
		}
	}
}
