// Package k8s scales the Maelstrom cluster through the kubernetes API,
// implementing the application.Scaler port.
package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Scaler struct {
	client    *kubernetes.Clientset
	namespace string
}

func New(namespace string) (*Scaler, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return nil, fmt.Errorf("error building k8 config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating k8 clientset: %w", err)
	}
	return &Scaler{
		client:    clientSet,
		namespace: namespace,
	}, nil
}

// Scale resizes the maelstrom StatefulSet to the given replica count.
func (s *Scaler) Scale(replicas int) error {
	sets := s.client.AppsV1().StatefulSets(s.namespace)

	cur, err := sets.GetScale(context.Background(), "maelstrom", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("reading current scale: %w", err)
	}

	sc := *cur
	sc.Spec.Replicas = int32(replicas)

	if _, err = sets.UpdateScale(context.Background(), "maelstrom", &sc, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("updating scale: %w", err)
	}
	return nil
}
