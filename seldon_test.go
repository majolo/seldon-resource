package main

import (
	"k8s.io/client-go/rest"
	"testing"

	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/stretchr/testify/require"
	k8s_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"majolo.uk/seldon/deployment_manager"
)

func TestParsingDeployment(t *testing.T) {
	dpl, err := openSeldonDeploymentJson("seldon-deployment.json")
	require.NoError(t, err)
	require.Equal(t, getTestSeldonDeployment(), dpl)

	_, err = openSeldonDeploymentJson("seldon-invalid.json")
	require.Error(t, err)
}

func TestNewDeploymentManager(t *testing.T) {
	config := rest.Config{}
	_, err := deployment_manager.NewSeldonDeploymentManager(&config, "default")
	require.NoError(t, err)
}

// TODO: further tests using a mocked SeldonDeploymentInterface

func getTestSeldonDeployment() *v1.SeldonDeployment {
	predictiveType := v1.PredictiveUnitType("MODEL")
	dpl := &v1.SeldonDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SeldonDeployment",
			APIVersion: "machinelearning.seldon.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "seldon-model",
		},
		Spec: v1.SeldonDeploymentSpec{
			Name: "test-deployment",
			Predictors: []v1.PredictorSpec{{
				Name: "example",
				Graph: v1.PredictiveUnit{
					Name:     "classifier",
					Children: []v1.PredictiveUnit{},
					Type:     &predictiveType,
					Endpoint: &v1.Endpoint{
						Type: "REST",
					},
				},
				ComponentSpecs: []*v1.SeldonPodSpec{{
					Spec: k8s_v1.PodSpec{
						Containers: []k8s_v1.Container{{
							Name:  "classifier",
							Image: "seldonio/mock_classifier:1.5.0",
						}},
					},
				}},
				Replicas: deployment_manager.Int32Ptr(1),
			}},
		},
	}
	return dpl
}
