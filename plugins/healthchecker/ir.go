package healthchecker

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type HealthCheckerServerWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

func (node *HealthCheckerServerWrapper) ImplementsGolangNode() {}

func (node *HealthCheckerServerWrapper) Name() string {
	return node.InstanceName
}

func (node *HealthCheckerServerWrapper) String() string {
	return node.Name() + " = HealthCheckerServerWrapper(" + node.Wrapped.Name() + ")"
}

func newHealthCheckerServerWrapper(name string, server blueprint.IRNode) (*HealthCheckerServerWrapper, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, fmt.Errorf("healthchecker server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &HealthCheckerServerWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	return node, nil
}
