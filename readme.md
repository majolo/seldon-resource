# Seldon Custom Resource Deployment

## What is this?

 * Create a standalone program in Go which takes in a Seldon Core Custom Resource and creates it over the Kubernetes API 
 * Watch the created resource to wait for it to become available. 
 * Scale the resource to 2 replicas.
 * When it is available delete the Custom Resource. 
 * In parallel to the last 3 steps list the Kubernetes Events with descriptions emitted by the created custom resource until it is deleted.

## How to run it?

### Dependencies
```
brew install minikube

minikube start

kubectl create namespace seldon-system

helm install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --set usageMetrics.enabled=true \
    --namespace seldon-system
```
### Running
Can optionally provide a `--file` flag to provide the filepath of a `.json` resource definition.

Also the logging of "list the Kubernetes Events with descriptions" is quite verbose, so is disabled by default and can be enabled using `--eventsLogging`.

```go run main.go```

## Further work

* Testing is fairly minimal, mainly around setup. Given more time/a productionised system we would run tests against a mocked SeldonDeploymentInterface and possibly integrations tests also.
* The client just uses context.Background() for now, but we could propogate a context if necessary.