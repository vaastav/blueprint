<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# news

```go
import "gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/news"
```

package news implements the ts\-news\-service from the TrainTicket application

## Index

- [type News](<#News>)
- [type NewsService](<#NewsService>)
- [type NewsServiceImpl](<#NewsServiceImpl>)
  - [func NewNewsServiceImpl\(ctx context.Context\) \(\*NewsServiceImpl, error\)](<#NewNewsServiceImpl>)
  - [func \(n \*NewsServiceImpl\) Hello\(ctx context.Context, val string\) \(string, error\)](<#NewsServiceImpl.Hello>)


<a name="News"></a>
## type [News](<https://gitlab.mpi-sws.org/cld/blueprint2/blueprint/blob/main/examples/train_ticket/workflow/news/data.go#L3-L6>)



```go
type News struct {
    Title   string `bson:"Title"`
    Content string `bson:"Content"`
}
```

<a name="NewsService"></a>
## type [NewsService](<https://gitlab.mpi-sws.org/cld/blueprint2/blueprint/blob/main/examples/train_ticket/workflow/news/newsService.go#L7-L9>)

News Service provides the latest news about the application

```go
type NewsService interface {
    Hello(ctx context.Context, val string) (string, error)
}
```

<a name="NewsServiceImpl"></a>
## type [NewsServiceImpl](<https://gitlab.mpi-sws.org/cld/blueprint2/blueprint/blob/main/examples/train_ticket/workflow/news/newsService.go#L12>)

News Service Implementation

```go
type NewsServiceImpl struct{}
```

<a name="NewNewsServiceImpl"></a>
### func [NewNewsServiceImpl](<https://gitlab.mpi-sws.org/cld/blueprint2/blueprint/blob/main/examples/train_ticket/workflow/news/newsService.go#L14>)

```go
func NewNewsServiceImpl(ctx context.Context) (*NewsServiceImpl, error)
```



<a name="NewsServiceImpl.Hello"></a>
### func \(\*NewsServiceImpl\) [Hello](<https://gitlab.mpi-sws.org/cld/blueprint2/blueprint/blob/main/examples/train_ticket/workflow/news/newsService.go#L18>)

```go
func (n *NewsServiceImpl) Hello(ctx context.Context, val string) (string, error)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)