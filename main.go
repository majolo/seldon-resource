package main

import (
    "encoding/json"
    "flag"
    "io/ioutil"
    "os"

    "github.com/mitchellh/go-homedir"
    v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
    "majolo.uk/seldon/deployment_manager"
)

func main() {
	resourceFilepath := flag.String("file", "seldon-deployment.json", "supply the filepath to a seldon deployment resource")
	enableEventsLogging := flag.Bool("eventsLogging", false, "whether to log k8s events")
	flag.Parse()

	// Initialise deployment manager and target deployment
	kConfig, err := findKubeConfig()
	if err != nil {
		panic(err)
	}
	manager, err := deployment_manager.NewSeldonDeploymentManagerFromFlags(kConfig, "default")
	if err != nil {
		panic(err)
	}
	dpl, err := openSeldonDeploymentJson(*resourceFilepath)
	if err != nil {
		panic(err)
	}

	// Initialise kubernetes event watcher
	if *enableEventsLogging {
		go manager.WatchKubernetesEvents()
	}

	// Run sequence of create, watch, update, watch, delete, watch
	err = manager.CreateDeployment(dpl)
	if err != nil {
		panic(err)
	}
	err = manager.WatchDeploymentForReadyReplicas(dpl.Name, 1)
	if err != nil {
		panic(err)
	}
	err = manager.UpdateDeploymentReplicas(dpl.Name, 2)
	if err != nil {
		panic(err)
	}
	err = manager.WatchDeploymentForReadyReplicas(dpl.Name, 2)
	if err != nil {
		panic(err)
	}
	err = manager.DeleteDeployment(dpl.Name)
	if err != nil {
		panic(err)
	}
	err = manager.WatchDeploymentForDeleted(dpl.Name)
	if err != nil {
		panic(err)
	}
}

func findKubeConfig() (string, error) {
	env := os.Getenv("KUBECONFIG")
	if env != "" {
		return env, nil
	}
	path, err := homedir.Expand("~/.kube/config")
	if err != nil {
		return "", err
	}
	return path, nil
}

func openSeldonDeploymentJson(filename string) (*v1.SeldonDeployment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var dpl v1.SeldonDeployment
	err = json.Unmarshal(b, &dpl)
	if err != nil {
		return nil, err
	}
	return &dpl, nil
}
