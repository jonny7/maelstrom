package k8s

import (
	"context"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockgen -source=k8s.go -destination mocks/mock_k8s.go -package mocks

type K8s interface {
	// @todo make this nicer
	Scale(replicas int) (int, error)
}

type k8s struct {
	client    *kubernetes.Clientset
	namespace string
}

func New(namespace string) (K8s, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return nil, fmt.Errorf("error building k8 config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating k8 clientset: %w", clientSet)
	}
	return &k8s{
		client:    clientSet,
		namespace: namespace,
	}, nil
}

func (k k8s) Scale(replicas int) (int, error) {
	if replicas < 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("replicas less than zero is invalid")
	}

	cur, err := k.client.AppsV1().
		StatefulSets(k.namespace).
		GetScale(context.Background(), "maelstrom", metav1.GetOptions{})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	sc := *cur
	sc.Spec.Replicas = int32(replicas)

	if _, err = k.client.AppsV1().
		StatefulSets(k.namespace).
		UpdateScale(context.Background(), "maelstrom", &sc, metav1.UpdateOptions{}); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
