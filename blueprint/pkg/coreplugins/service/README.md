<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# service

```go
import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
```

## Index

- [type Method](<#Method>)
- [type ServiceInterface](<#ServiceInterface>)
- [type ServiceNode](<#ServiceNode>)
- [type Variable](<#Variable>)


<a name="Method"></a>
## type Method



```go
type Method interface {
    GetName() string
    GetArguments() []Variable
    GetReturns() []Variable
}
```

<a name="ServiceInterface"></a>
## type ServiceInterface



```go
type ServiceInterface interface {
    GetName() string
    GetMethods() []Method
}
```

<a name="ServiceNode"></a>
## type ServiceNode

Any IR node that represents a callable service should implement this interface.

```go
type ServiceNode interface {

    // Returns the interface of this service
    GetInterface(ctx ir.BuildContext) (ServiceInterface, error)
}
```

<a name="Variable"></a>
## type Variable



```go
type Variable interface {
    GetName() string
    GetType() string // a "well-known" type
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)