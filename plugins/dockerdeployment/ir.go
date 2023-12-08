package dockerdeployment

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
)

/* A deployment is a collection of containers */
type Deployment struct {
	/* The implemented build targets for dockercompose.DockerCompose nodes */
	dockerComposeDeployer /* Can be deployed as a docker-compose file; implemented in deploydockercompose.go */

	DeploymentName string
	ArgNodes       []ir.IRNode
	ContainedNodes []ir.IRNode
}

func newContainerDeployment(name string, argNodes, containedNodes []ir.IRNode) *Deployment {
	node := Deployment{
		DeploymentName: name,
		ArgNodes:       argNodes,
		ContainedNodes: containedNodes,
	}
	return &node
}

func (node *Deployment) Name() string {
	return node.DeploymentName
}

func (node *Deployment) String() string {
	return ir.PrettyPrintNamespace(node.DeploymentName, "DockerApp", node.ArgNodes, node.ContainedNodes)
}
