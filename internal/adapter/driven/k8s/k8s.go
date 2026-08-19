// Package k8s scales the Maelstrom cluster through the kubernetes API,
// implementing the application.Scaler port.
package k8s

import (
	"context"
	"fmt"

	"github.com/caarlos0/env/v10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Namespace   string `env:"NAMESPACE" envDefault:"default"`
	StatefulSet string `env:"STATEFULSET" envDefault:"maelstrom"`
}

type Cluster struct {
	client *kubernetes.Clientset
	cfg    Config
}

func New() (*Cluster, error) {
	var cfg Config
	if err := env.ParseWithOptions(&cfg, env.Options{Prefix: "K8S_"}); err != nil {
		return nil, fmt.Errorf("k8s config failed to load: %w", err)
	}

	restCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return nil, fmt.Errorf("error building k8 config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating k8 clientset: %w", err)
	}
	return &Cluster{
		client: clientSet,
		cfg:    cfg,
	}, nil
}

// Scale resizes the Maelstrom StatefulSet to the given replica count.
func (s *Cluster) Scale(replicas int) error {
	sets := s.client.AppsV1().StatefulSets(s.cfg.Namespace)

	cur, err := sets.GetScale(context.Background(), s.cfg.StatefulSet, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("reading current scale: %w", err)
	}

	sc := *cur
	sc.Spec.Replicas = int32(replicas)

	if _, err = sets.UpdateScale(context.Background(), s.cfg.StatefulSet, &sc, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("updating scale: %w", err)
	}
	return nil
}

// DeleteNode deletes the named pod; the StatefulSet controller replaces it.
func (s *Cluster) DeleteNode(node string) error {
	if err := s.client.CoreV1().Pods(s.cfg.Namespace).Delete(context.Background(), node, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("deleting node %s: %w", node, err)
	}
	return nil
}
