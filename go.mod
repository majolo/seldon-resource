module majolo.uk/seldon

go 1.15

require (
	github.com/mitchellh/go-homedir v1.1.0
	github.com/seldonio/seldon-core/operator v0.0.0-20210205201824-426432a83440
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.4
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
