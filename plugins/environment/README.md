<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# environment

```go
import "github.com/blueprint-uservices/blueprint/plugins/environment"
```

Package environment provides a plugin for generating a .env file in the root Blueprint output directory that automatically sets address configuration variables \(hostnames and ports for dial and bind addresses\).

The plugin is intended for convenience so that Blueprint users do not have to manually allocate ports and pass them as environment variables. However, in more complex deployment, Blueprint users may wish to disable this plugin to afford themselves more control.

### Wiring Spec Usage

The environment plugin is automatically enabled if you are using the [cmdbuilder](<https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/cmdbuilder>). See the cmdbuilder documentation for command\-line flags you can pass to control environment generation. Otherwise you can manually use the environment plugin from your wiring spec as follows

```
environment.AssignPorts(spec, 12345)
```

Ports will be automatically assigned to services starting from 12345 and incrementing.

### Generated Artifacts

The plugin will generate several .env files to the root output directory. The .env files take the form:

```
USER_SERVICE_GRPC_DIAL_ADDR=user_service:12345
USER_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12345
```

The plugin generates two different env files:

- .local.env assumes all services will be deployed on a single machine; it uses localhost for dial hostnames and 0.0.0.0 for bind hostnames, e.g. localhost:12345 and 0.0.0.0:12345
- .env uses the service name as dial hostname and 0.0.0.0 for bind hostname, e.g. user\_service:12345 and 0.0.0.0:12345. To use this .env file you will need to ensure that service hostnames are mapped in your /etc/hosts or dns server.

### Running Artifacts

Before running the application or a client, you can source one of the .env files to avoid having to manually set environment variables or command line arguments.

For example, if you are running a docker\-compose deployment, you can run:

```
cd build
set -a
. ./.local.env
cd docker
docker compose up
```

Or:

```
cd build/docker
docker compose --env-file=../.local.env build
```

Similarly, workload generator clients and tests will check environment variables for default values.

If you are using .env then the hostnames for services will need to be mapped in your /etc/hosts file or dns server.

The plugin does not guarantee that the ports \(e.g. 12345\) are actually available for use on any machine. This is up to the user.

## Index

- [func AssignPorts\(initialPort uint16\)](<#AssignPorts>)


<a name="AssignPorts"></a>
## func [AssignPorts](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/environment/wiring.go#L80>)

```go
func AssignPorts(initialPort uint16)
```

[AssignPorts](<#AssignPorts>) can be called from a wiring spec to auto\-generate .env files in the root of the build directory that set environment variables for service bind and dial addresses.

Ports are allocated starting from the specified initialPort.

If you are using the [cmdbuilder](<https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/cmdbuilder>) then [AssignPorts](<#AssignPorts>) is called automatically with a default initialPort of 12345; you can use the \-\-port argument to override the initialPort value, or \-\-env=false to disable.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)