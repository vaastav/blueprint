package golang

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

/*
The golang plugin extends Blueprint's IR as follows:

It defines the following IR interfaces:

  - golang.Node is the base interface for any node that lives within a golang process
  - golang.Service is a golang node that has methods that can be directly called by other golang nodes

The golang plugin also defines the following new IR node implementations:

  - golang.Process is a node that represents a runnable Golang process.  It can contain any number of
    other golang.Node IRNodes.  When it's compiled, the golang.Process will generate a go module with
    a runnable main method that instantiates and initializes the contained go nodes.  To achieve this,
    the golang.Process also collects module dependencies from its contained nodes.

To support golang code generation, the following IR interfaces are provided, as defined in ir_codegen.go.
The golang.Process depends on these interfaces for collecting and packaging code, however, usage of these interfaces
is not intended to be private to just the golang.Process plugin.  Other plugins are permitted to
use these interfaces.

  - golang.Instantiable is for golang nodes that can generate instantiation source code snippets
  - golang.RequiresPackage is for golang nodes that generate source files and have module dependencies
  - golang.ProvidesModule is for golang nodes that generate or otherwise provide the full source code of modules
*/
type (
	/*
		golang.Node is the base IRNode interface that should be implemented by any IRNode that
		wishes to exist within a Golang namespace.
	*/
	Node interface {
		blueprint.IRNode
		ImplementsGolangNode() // Idiomatically necessary in Go for typecasting correctly
	}

	/*
		golang.Service is a golang.Node that exposes an interface that can be directly invoked
		by other golang.Nodes.

		For example, services within a workflow spec are represented by golang.Service nodes
		because they have invokable methods.  Similarly plugins such as tracing, which
		wrap service nodes, are themselves also service nodes, because they have invokable methods.

		golang.Service extends the golang.Instantiable interface, which is part of the codegen
		process.  Thus any plugin that provides IRNodes that extend golang.Service must implement
		the code generation methods defined by the golang.Instantiable interface.
	*/
	Service interface {
		Node
		Instantiable              // Services must implement the Instantiable interface in order to create instances
		service.ServiceNode       // The GetServiceInterface() method should return a gocode.ServiceInterface instance
		ImplementsGolangService() // Idiomatically necessary in Go for typecasting correctly
	}
)

/*
Golang code-generation interfaces that IR nodes should implement if they expect to generate code or artifacts
*/
type (
	/*
	   This is an interface for IRNodes that can be used by plugins that want to be able to instantiate things in
	   the generated code.  For example, a service, or a tracing wrapper, will want to instantiate code.

	   All Services are Instantiable, but not all Instantiable are Services.  For example, a Golang GRPC
	   server is instantiable but it does not expose methods that can be directly invoked at the application
	   level.  In constract a Golang GRPC client is a service and is instantiable because it does expose methods
	   that can be directly invoked.

	   The GraphBuilder struct provides functionality for plugins to declare:

	   	(a) how to instantiate instances of things
	   	(b) relevant types and method signatures for use by other plugins
	*/
	Instantiable interface {
		/*
		   AddInstantion will be invoked during compilation to enable an IRNode to add its instantiation code
		   to a generated golang file.
		*/
		AddInstantiation(GraphBuilder) error
	}

	/*
			Some IRNodes generate new interfaces and struct definitions by extending the interfaces declared by other IRNodes.
			They should implement this interface to do so

		   There are three interfaces for contributing source code to generated modules:
		   1. RequiresPackage for adding module dependencies to the generated module's go.mod file
		   2. GeneratesInterfaces for generating struct and interface type declarations
		   3. GeneratesFuncs for generating functions and struct method bodies
	*/
	GeneratesInterfaces interface {
		/*
		   GenerateInterfaces will be invoked during compilation to enable an IRNode to write generated code files to
		   an output module, containing interface and struct definitions
		*/
		GenerateInterfaces(ModuleBuilder) error
	}

	/*
			Some IRNodes generate implementations of service interfaces.  This interface should be used to do so.  This is
			separate from the GeneratesTypes interface because all structs and interfaces need to be declared before
			method bodies can be written, because method bodies might need to use structs and interfaces defined in different
			packages

		   There are three interfaces for contributing source code to generated modules:
		   1. RequiresPackage for adding module dependencies to the generated module's go.mod file
		   2. GeneratesInterfaces for generating struct and interface type declarations
		   3. GeneratesFuncs for generating functions and struct method bodies
	*/
	GeneratesFuncs interface {
		/*
		   GenerateFuncs will be invoked during compilation to enable an IRNode to write generated code files to
		   an output module, containing interface and struct definitions
		*/
		GenerateFuncs(ModuleBuilder) error
	}

	/*
	   This is an interface for IRNodes for plugins that want to include standalone modules in the output workspace.
	   The most straightforward example is the workflow spec, which will be copied into the output workspace
	   using this interface.

	   The IRNode must implement the `AddToWorkspace` method, to interact with the `WorkspaceBuilder` to copy
	   relevant modules.
	*/
	ProvidesModule interface {
		/*
			AddToWorkspace will be invoked during compimlation to enable an IRNode to copy a local Go module
			directly into the output workspace directory
		*/
		AddToWorkspace(WorkspaceBuilder) error
	}
)

/*
APIs used by the above IR nodes when they are generating code.

The main implementation of these interfaces is in the [goprocess](../goprocess) plugin
*/
type (
	WorkspaceInfo struct {
		Path string // fully-qualified path on the filesystem to this workspace
	}

	/*
	   WorkspaceBuilder is used by plugins if they want to collect and combine Golang code and modules.

	   An IRNode must implement the ProvidesModule interface; then during compilation, `AddToWorkspace`
	   will be called, enabling the IRNode to add its code and modules to the output workspace directory
	   using the methods on `WorkspaceBuilder`.
	*/
	WorkspaceBuilder interface {
		irutil.VisitTracker

		/*
			Metadata into about the workspace being built
		*/
		Info() WorkspaceInfo

		/*
			This is equivalent to calling node.AddToModule, if node implements it, for the specified nodes
		*/
		Visit(nodes []blueprint.IRNode) error

		/*
		   This method is used by plugins if they want to copy a locally-defined module into the generated workspace.

		   The specified moduleSrcPath must point to a valid Go module with a go.mod file.
		*/
		AddLocalModule(shortName string, moduleSrcPath string) error

		/*
			This is a variant of `AddLocalMethod` provided for convenience; instead of an absolute filesystem path, the
			specified path is relative to the caller
		*/
		AddLocalModuleRelative(shortName string, relativeModuleSrcPath string) error

		/*
			This method is used by plugins if they want to create a module in the workspace to then generate code into.

			The specified moduleName must be a golang style module name.

			This will create the directory for the module and an empty go.mod file.

			Returns the path to the module in the output directory
		*/
		CreateModule(moduleName string, moduleVersion string) (string, error)

		/*
		   If the specified module exists locally within the workspace, gets the subdirectory within the workspace that it exists in, the module
		   version, and returns true.

		   Returns "", false otherwise
		*/
		GetLocalModule(modulePath string) (string, bool)
	}

	ModuleInfo struct {
		Name    string // Fully-qualified module name being built
		Version string // Version of the module being built
		Path    string // The path on the filesystem to the directory containing the module
	}

	/*
	   ModuleBuilder is used by IRNodes for plugins that want to generate Golang code and collect it into a module.

	   An IRNode must implement the RequiresPackages interface; then during compilation, `AddToModule`
	   will be called, enabling the IRNode to add its dependencies and code to the output module using the
	   methods on `ModuleBuilder`.

	   After creating a module builder, plugins can directly create directories and copy files into
	   the ModuleDir.  Any go dependencies should be added with the Require function.

	   When finished building the module, plugins should call Finish to finish building the go.mod
	   file
	*/
	ModuleBuilder interface {
		irutil.VisitTracker

		/*
			Metadata into about the module being built
		*/
		Info() ModuleInfo

		/*
			This is equivalent to calling node.AddToModule, if node implements it
		*/
		Visit(nodes []blueprint.IRNode) error

		/*
			Gets the WorkspaceBuilder that contains this ModuleBuilder
		*/
		Workspace() WorkspaceBuilder
	}

	GraphInfo struct {
		PackageName string // Fully qualified package name being built
		FilePath    string // Path to the file being built
	}

	/*
	   GraphBuilder is used by IRNodes that implement the Instantiable interface.  The GraphBuilder provides
	   the following methods that can be used by plugins to provide instantiation code:

	     - `Import` declares that a particular package should be imported, as it will be used by the
	       instantiation code

	     - `Declare` provides a buildFunc as a string that will be inserted into the output file; buildFunc
	       is used at runtime to create the instance

	   In the generated golang code, instances are declared and created using a simple dependency injection
	   style.  The runtime dependency injection interface is defined in runtime/plugins/golang/di.go

	   The basic requirement of an instantiable node is that it can provide a buildFunc definition that
	   will be invoked at runtime to create the instance.  A buildFunc has method signature:

	   	func(ctr golang.Container) (any, error)

	   The buildFunc will instantiate and return an instance or an error.  If the node needs to be
	   able to call other instances, it can acquire the instances through the golang.Container's Get
	   method.  For example, the following pseudocode for a tracing wrapper class would get the
	   underlying handler then return the wrapper class:

	   	func(ctr golang.Container) (any, error) {
	   		handler, err := ctr.Get("serviceA.handler")
	   		if err != nil {
	   			return nil, err
	   		}

	   		serviceA, isValid := handler.(ServiceA)
	   		if !isValid {
	   			return nil, blueprint.Errorf("serviceA.handler does not implement ServiceA interface")
	   		}

	   		return newServiceATracingWrapper(serviceA), nil
	   	}

	   The above code makes reference to names like `serviceA.handler`; rarely should these names
	   be hard-coded, instead they would typically be provided by calling or inspecting the IR
	   dependencies of this node.
	*/
	GraphBuilder interface {
		irutil.VisitTracker

		/*
			Metadata into about the graph being built
		*/
		Info() GraphInfo

		/*
			This is equivalent to calling node.AddToModule, if node implements it
		*/
		Visit(nodes []blueprint.IRNode) error

		/*
			Adds an import statement to the generated file; this is necessary for any types
			declared in other packages that are going to be used in a DI declaration.

			This method returns the type alias that should be used in the generated code.
			By default the type alias is just the package name, but if there are multiple
			different imports with the same package name, then aliases will be created
		*/
		Import(packageName string) string

		/*
			If the provided type is a user type or a builtin type, adds an import statement
			similar to the `Import` method.

			Returns the name that should be used in code for the type.  For example, if it's
			a type from an imported package, then would return mypackage.Foo.
		*/
		ImportType(typeName gocode.TypeName) string

		/*
			Provides the source code of a buildFunc that will be invoked at runtime by the
			generated code, to build the named instance
		*/
		Declare(instanceName string, buildFuncSrc string) error

		/*
			This is like Declare, but instead of having to manually construct the source
			code, the GraphBuilder will automatically create the build func src code,
			invoking the specified constructor and passing the provided nodes as args
		*/
		DeclareConstructor(name string, constructor *gocode.Constructor, args []blueprint.IRNode) error

		/*
			Gets the ModuleBuilder that contains this GraphBuilder
		*/
		Module() ModuleBuilder
	}
)
