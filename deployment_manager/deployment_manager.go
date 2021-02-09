package deployment_manager

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/client-go/rest"

	seldon_v1_api "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldon_v1_client "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	seldon_typed_v1 "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned/typed/machinelearning.seldon.io/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"log"
	"time"
)

const (
	operationTimeout = time.Second * 60
)

type SeldonDeploymentManager struct {
	SeldonDeployment seldon_typed_v1.SeldonDeploymentInterface
}

func NewSeldonDeploymentManagerFromFlags(kubeConfigPath string, namespace string) (*SeldonDeploymentManager, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return NewSeldonDeploymentManager(config, namespace)
}

func NewSeldonDeploymentManager(config *rest.Config, namespace string) (*SeldonDeploymentManager, error) {
	clientset, err := seldon_v1_client.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &SeldonDeploymentManager{
		SeldonDeployment: clientset.MachinelearningV1().SeldonDeployments(namespace),
	}, nil
}

// ---------------------------------------------------------------------------------------------------------------------

func (s *SeldonDeploymentManager) CreateDeployment(dpl *seldon_v1_api.SeldonDeployment) error {
	dpl, err := s.SeldonDeployment.Create(context.Background(), dpl, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Printf("ACTION: Creating seldon deployment: %s\n", dpl.Name)
	return nil
}

// Perhaps add retries for stale bug? https://github.com/SeldonIO/seldon-core/issues/2095
func (s *SeldonDeploymentManager) UpdateDeploymentReplicas(name string, replicas int32) error {
	dpl, err := s.SeldonDeployment.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Feels a bit hard-coded but works for now
	dpl.Spec.Predictors[0].Replicas = Int32Ptr(replicas)
	_, err = s.SeldonDeployment.Update(context.Background(), dpl, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	log.Printf("ACTION: Updating seldon deployment replicas to: %s, %v\n", name, replicas)
	return nil
}

func (s *SeldonDeploymentManager) DeleteDeployment(name string) error {
	err := s.SeldonDeployment.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	log.Printf("ACTION: Deleting seldon deployment: %s\n", name)
	return nil
}

func (s *SeldonDeploymentManager) WatchDeploymentForReadyReplicas(name string, replicas int) error {
	watcher, err := s.SeldonDeployment.Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for {
		select {
		case event := <-watcher.ResultChan():
			if dpl, ok := event.Object.(*seldon_v1_api.SeldonDeployment); ok {
				if dpl.Name == name && dpl.Status.State == seldon_v1_api.StatusStateAvailable {
					// Sum the ready replicas
					readyReplicas := 0
					for _, v := range dpl.Status.DeploymentStatus {
						readyReplicas += int(v.AvailableReplicas)
					}
					if readyReplicas == replicas {
						log.Printf("----- Succesfully waited for deployment (%s) to have %v ready replicas -----\n", name, replicas)
						return nil
					}
				}
				log.Printf("Got event for deployment (%s), event type (%v), current deployment status (%v)\n", name, event.Type, dpl.Status.State)
			}
		case <-time.After(operationTimeout):
			return errors.New(fmt.Sprintf("watcher timed out after %v", operationTimeout))
		}
	}
}

func (s *SeldonDeploymentManager) WatchDeploymentForDeleted(name string) error {
	for {
		dpl, err := s.SeldonDeployment.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			statusError, ok := err.(*k8s_errors.StatusError)
			if ok && statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
				log.Printf("----- Deployment (%s) succesfully finished deletion -----\n", name)
				return nil
			} else {
				return err
			}
		}
		log.Printf("Deployment (%s) not yet deleted, waiting for deletion...\n", dpl.Name)
		time.Sleep(3 * time.Second)
	}
}

func Int32Ptr(i int32) *int32 { return &i }
