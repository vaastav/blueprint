package grpccodegen

import (
	"os"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
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
		return blueprint.Errorf("unable to create grpc output dir %v due to %v", outputDir, err.Error())
	}

	server.Imports = gogen.NewImports(server.PackageName)
	for _, pkg := range serverImportedPackages {
		server.Imports.AddPackage(pkg)
	}

	outputFilename := service.Name + "_GRPCServer.go"
	return server.GenerateCode(filepath.Join(outputDir, outputFilename))
}

var serverImportedPackages = []string{
	"context", "net",
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

var serverHeader = `// Blueprint: Auto-generated by GRPC Plugin
package {{.PackageShortName}}

{{.Imports}}
`

var serverBody = `
type {{.Name}} struct {
	Unimplemented{{.Service.Name}}Server
	Service {{.Imports.NameOf .Service.UserType}}
	Address string
}

func New_{{.Name}}(service {{.Imports.NameOf .Service.UserType}}, serverAddress string) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	handler.Address = serverAddress
	return handler, nil
}

// Blueprint: Run is called automatically in a separate goroutine by runtime/plugins/golang/di.go
func (handler *{{.Name}}) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", handler.Address)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	Register{{.Service.Name}}Server(s, handler)

	go func() {
		select {
		case <-ctx.Done():
			s.GracefulStop()
		}
	}()

	return s.Serve(lis)
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{$imports := .Imports -}}
{{ range $_, $f := .Service.Methods }}
func (handler *{{$receiver}}) {{$f.Name -}}
		(ctx context.Context, req *{{$service}}_{{$f.Name}}_Request) (*{{$service}}_{{$f.Name}}_Response, error) {
	{{range $i, $arg := $f.Arguments}}{{if $i}}, {{end}}{{$arg.Name}}{{end}} := req.unmarshall()
	{{range $i, $ret := $f.Returns }}ret{{$i}}, {{end}}err := handler.Service.{{$f.Name -}}
		(ctx {{- range $i, $arg := $f.Arguments}}, {{$arg.Name}}{{end}})
	if err != nil {
		return nil, err
	}

	rsp := &{{$service}}_{{$f.Name}}_Response{}
	rsp.marshall(
		{{- range $i, $arg := $f.Returns}}{{if $i}}, {{end}}ret{{$i}}{{end -}}
	)
	return rsp, nil
}
{{end}}
`

/*
Generates the file within its module
*/
func (server *serverTemplateArgs) GenerateCode(outputFilePath string) error {
	// Execute the body before the header, to automatically add all import statements for the header
	body, err := gogen.ExecuteTemplate("GRPCServerBody", serverBody, server)
	if err != nil {
		return err
	}

	header, err := gogen.ExecuteTemplate("GRPCServerHeader", serverHeader, server)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(header + body)
	return err
}
