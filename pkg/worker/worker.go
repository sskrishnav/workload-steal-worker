package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	nats "github.com/nats-io/nats.go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Config struct {
	NATSURL     string
	NATSSubject string
}

type consume struct {
	nconfig Config
	cli     *kubernetes.Clientset
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, err
}

func New(config Config) (*consume, error) {
	clientset, err := getClientSet()
	if err != nil {
		return nil, err
	}
	return &consume{
		cli:     clientset,
		nconfig: config,
	}, nil
}

func (n *consume) Start(stopChan chan<- bool) error {
	defer func() { stopChan <- true }()

	// Configure structured logging with slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to NATS server
	natsConnect, err := nats.Connect(n.nconfig.NATSURL)
	if err != nil {
		slog.Error("Failed to connect to NATS server: ", "error", err)
		return err
	}
	defer natsConnect.Close()
	slog.Info("Connected to NATS server")

	var pod corev1.Pod
	slog.Info("Subscribe to Pod creation messages...")
	// Subscribe to the subject
	natsConnect.Subscribe(n.nconfig.NATSSubject, func(msg *nats.Msg) {
		slog.Info("Received message", "subject", n.nconfig.NATSSubject, "data", string(msg.Data))
		// Deserialize the entire Pod metadata to JSON
		if err := json.Unmarshal(msg.Data, &pod); err != nil {
			slog.Error("Failed to decode Pod from message", "error", err)
			return
		}
	})

	// Create the Pod in Kubernetes
	createdPod, err := n.cli.CoreV1().Pods(pod.Namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create Pod", "error", err)
		return err
	}
	slog.Info("Succefully stole the wrokload", "Pod", createdPod)
	return nil
}
