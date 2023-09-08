package grpccodegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"
)

/*
This function is used by the GRPC plugin to generate the server-side GRPC service.

It is assumed that outputPackage is the same as the one where the .proto is generated to
*/
func GenerateServerHandler(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error {
	server := &serverTemplateArgs{}
	server.Builder = builder
	splits := strings.Split(outputPackage, "/")
	outputPackageName := splits[len(splits)-1]
	server.PackageName = builder.Info().Name + "/" + outputPackage
	server.PackageShortName = outputPackageName
	server.Service = service
	server.Name = service.Name + "_GRPCServerHandler"

	outputDir := filepath.Join(builder.Info().Path, filepath.Join(splits...))
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create grpc output dir %v due to %v", outputDir, err.Error())
	}

	err = server.initImports()
	if err != nil {
		return err
	}

	outputFilename := service.Name + "_GRPCServer.go"
	return server.GenerateCode(filepath.Join(outputDir, outputFilename))
}

var serverRequiredModules = map[string]string{
	"google.golang.org/grpc": "v1.41.0",
}
var serverImportedPackages = []string{
	"context", "errors", "net",
	"google.golang.org/grpc",
}

/*
Arguments to the template code
*/
type serverTemplateArgs struct {
	Builder          golang.ModuleBuilder
	PackageName      string // fully qualified package name
	PackageShortName string // package shortname
	FilePath         string
	Service          *gocode.ServiceInterface
	Name             string         // Name of the generated wrapper class
	Imports          *gogen.Imports // Manages imports for us
}

func (client *serverTemplateArgs) importType(t gocode.TypeName) error {
	client.Imports.AddType(t)
	return client.Builder.RequireType(t)
}

func (client *serverTemplateArgs) initImports() error {
	// In addition to a few GRPC-related requirements,
	// we also depend on the modules that define the
	// argument types to the RPC methods
	client.Imports = gogen.NewImports(client.PackageName)

	for name, version := range serverRequiredModules {
		err := client.Builder.Require(name, version)
		if err != nil {
			return err
		}
	}

	for _, pkg := range serverImportedPackages {
		client.Imports.AddPackage(pkg)
	}

	for _, f := range client.Service.Methods {
		for _, v := range f.Arguments {
			err := client.importType(v.Type)
			if err != nil {
				return err
			}
		}
		for _, v := range f.Returns {
			err := client.importType(v.Type)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var serverTemplate = `// Blueprint: Auto-generated by GRPC Plugin
package {{.PackageShortName}}

{{.Imports}}

type {{.Name}} struct {
	Service *{{.Imports.NameOf .Service}}
	Address string
}

func New_{{.Name}}(service *{{.Imports.NameOf .Service}}, serverAddress string) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	handler.Address = serverAddress
	return handler, nil
}

func (handler *{{.Name}}) Run() error {
	lis, err := net.Listen("tcp", handler.Address)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	Register{{.Service.Name}}Server(grpcServer, handler)
	return grpcServer.Serve(lis)
}

// TODO: server methods and run
`

/*
Generates the file within its module
*/
func (client *serverTemplateArgs) GenerateCode(outputFilePath string) error {
	t, err := template.New("GRPCServerTemplate").Parse(serverTemplate)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(outputFilePath, os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	return t.Execute(f, client)
}
