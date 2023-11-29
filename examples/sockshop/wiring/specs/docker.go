package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Used by main.go
var Docker = wiringcmd.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends.",
	Build:       makeDockerSpec,
}

// Creates a basic sockshop wiring spec.
// Returns the names of the nodes to instantiate or an error
func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := mongodb.PrebuiltContainer(spec, "user_db")
	user_service := workflow.Define(spec, "user_service", "UserService", user_db)
	user_ctr := applyDockerDefaults(spec, user_service, "user_proc", "user_container")

	payment_service := workflow.Define(spec, "payment_service", "PaymentService")
	payment_ctr := applyDockerDefaults(spec, payment_service, "payment_proc", "payment_container")

	cart_db := mongodb.PrebuiltContainer(spec, "cart_db")
	cart_service := workflow.Define(spec, "cart_service", "CartService", cart_db)
	cart_ctr := applyDockerDefaults(spec, cart_service, "cart_proc", "cart_ctr")

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := mongodb.PrebuiltContainer(spec, "shipping_db")
	shipping_service := workflow.Define(spec, "shipping_service", "ShippingService", shipqueue, shipdb)
	shipping_ctr := applyDockerDefaults(spec, shipping_service, "shipping_proc", "shipping_ctr")

	// Deploy queue master to the same process as the shipping proc
	// TODO: after distributed queue is supported, move to separate containers
	queue_master := workflow.Define(spec, "queue_master", "QueueMaster", shipqueue, shipping_service)
	goproc.AddChildToProcess(spec, "shipping_proc", queue_master)

	order_db := mongodb.PrebuiltContainer(spec, "order_db")
	order_service := workflow.Define(spec, "order_service", "OrderService", user_service, cart_service, payment_service, shipping_service, order_db)
	order_ctr := applyDockerDefaults(spec, order_service, "order_proc", "order_ctr")

	tests := gotests.Test(spec, user_service, payment_service, cart_service, shipping_service, order_service)

	return []string{user_ctr, payment_ctr, cart_ctr, shipping_ctr, order_ctr, tests}, nil
}

func applyDockerDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}