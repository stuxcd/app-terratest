package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Pod struct {
	Name      string
	Namespace string
	Ready     bool
	Status    string
}

type Pods struct {
	Pods  []Pod
	Ready bool
}

func NewPod(name string, namespace string, ready bool, status string) (*Pod, error) {
	return &Pod{
		Name:      name,
		Namespace: namespace,
		Ready:     ready,
		Status:    status,
	}, nil
}

func NewPods(pods []Pod) (*Pods, error) {
	allReady := true
	for _, pod := range pods {
		if !pod.Ready {
			allReady = false
			break
		}
	}
	return &Pods{
		Pods:  pods,
		Ready: allReady,
	}, nil
}

func (c *Clientset) GetPods(ctx context.Context, namespace string) (*Pods, error) {
	podList, err := c.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	pods := make([]Pod, len(podList.Items))
	for i, pod := range podList.Items {
		podStatus := pod.Status

		var containersReady int
		var totalContainers int
		for container := range pod.Spec.Containers {
			if podStatus.ContainerStatuses[container].Ready {
				containersReady++
			}
			totalContainers++
		}
		name := pod.GetObjectMeta().GetName()
		namespace := pod.GetObjectMeta().GetNamespace()
		ready := containersReady == totalContainers
		status := fmt.Sprintf("%v", podStatus.Phase)

		p, err := NewPod(name, namespace, ready, status)
		if err != nil {
			return nil, err
		}
		pods[i] = *p
	}
	return NewPods(pods)
}

func (c *Clientset) WaitUntilPodsReady(ctx context.Context, namespace string, seconds time.Duration) (*Pods, error) {
	pods, err := c.GetPods(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if pods.Ready {
		return pods, nil
	}
	deadline := time.Now().Add(seconds * time.Second)
	for {
		p, err := c.GetPods(ctx, namespace)
		if err != nil {
			return nil, err
		}
		if p.Ready || time.Now().After(deadline) {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return pods, nil
}
