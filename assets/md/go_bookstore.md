# Go图书管理API服务

go1.17.7  
服务说明：  
针对前端或其他客户端调用提供图书restful风格的CRUD（创建、检索、更新与删除）服务；

```
项目代码：
main.go
store/store.go
store/factory/factory.go
internal/store/memstore.go
server/middleware/middleware.go
server/server.go
makeRun
```

```
创建目录
$mkdir bookstore
$cd bookstore
$go mod init bookstore
go: creating new go.mod: module bookstore

项目结构
├── cmd/
│   └── bookstore/         // 放置bookstore main包源码
│       └── main.go
├── go.mod                 // module bookstore的go.mod
├── go.sum
├── internal/              // 存放项目内部包的目录
│   └── store/
│       └── memstore.go     
├── server/                // HTTP服务器模块
│   ├── middleware/
│   │   └── middleware.go
│   └── server.go          
└── store/                 // 图书数据存储模块
    ├── factory/
    │   └── factory.go
    └── store.go
```

```
main.go

  package main
 
  import (
      _ "bookstore/internal/store" //internal/store将自身注册到factory中
      "bookstore/server"
      "bookstore/store/factory"
      "context"
      "log"
      "os"
      "os/signal"
      "syscall"
      "time"
 )
 
 func main() {
     s, err := factory.New("mem") // 创建名为"mem"的图书数据存储模块实例
     if err != nil {
         panic(err)
     }
 
     srv := server.NewBookStoreServer(":8080", s) // 创建http服务实例
 
     errChan, err := srv.ListenAndServe() // 运行http服务
     if err != nil {
         log.Println("web server start failed:", err)
         return
     }
     log.Println("web server start ok")
 
     c := make(chan os.Signal, 1)
     signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
 
     select { // 监视来自errChan以及c的事件
     case err = <-errChan:
         log.Println("web server run failed:", err)
         return
     case <-c:
         log.Println("bookstore program is exiting...")
         ctx, cf := context.WithTimeout(context.Background(), time.Second)
         defer cf()
         err = srv.Shutdown(ctx) // 优雅关闭http服务实例
     }
 
     if err != nil {
         log.Println("bookstore program exit error:", err)
         return
     }
     log.Println("bookstore program exit ok")
 }
```

```
// store/store.go
在内存中创建一个 map，以图书 id 作为 key，来保存图书信息
建立一个对应图书条目的抽象数据类型 Book，以及针对 Book 存取的接口类型Store

package store

import "errors"

var (
        ErrNotFound = errors.New("not found")
        ErrExist    = errors.New("exist")
)G

 type Book struct {
     Id      string   `json:"id"`      // 图书ISBN ID
     Name    string   `json:"name"`    // 图书名称
     Authors []string `json:"authors"` // 图书作者
     Press   string   `json:"press"`   // 出版社
 }
 
 type Store interface {
     Create(*Book) error        // 创建一个新图书条目
     Update(*Book) error        // 更新某图书条目
     Get(string) (Book, error)  // 获取某图书信息
     GetAll() ([]Book, error)   // 获取所有图书信息
     Delete(string) error       // 删除某图书条目
 }
```

```
// store/factory/factory.go
使用Go风格的工厂模式来实现store接口实例实现
factory 包采用了一个 map 类型数据，对工厂可以“生产”的、满足 Store 接口的实例类型进行管理。factory 包还提供了 Register 函数，让各个实现 Store 接口的类型可以把自己“注册”到工厂中来。一旦注册成功，factory 包就可以“生产”出这种满足 Store 接口的类型实例。而依赖 Store 接口的使用方，只需要调用 factory 包的 New 函数，再传入期望使用的图书存储实现的名称，就可以得到对应的类型实例了

 igo@igoMacBook bookstore % cat store/factory/factory.go

package factory

import (
    "bookstore/store"
    "fmt"
    "sync"
)

var (
    providersMu sync.RWMutex
    providers   = make(map[string]store.Store)
)

func Register(name string, p store.Store) {
    providersMu.Lock()
    defer providersMu.Unlock()
    if p == nil {
        panic("store: Register provider is nil")
    }

    if _, dup := providers[name]; dup {
        panic("store: Register called twice for provider " + name)
    }
    providers[name] = p
}

func New(providerName string) (store.Store, error) {
    providersMu.RLock()
    p, ok := providers[providerName]
    providersMu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("store: unknown provider %s", providerName)
    }

    return p, nil
}
```

```
// internal/store/memstore.go
基于内存 map 的 Store 接口的实现并自注册到factory包中
memstore 的代码在包的 init 函数中调用 factory 包提供的 Register 函数，把自己的实例以“mem”的名称注册到 factory 中的。这样依赖 Store 接口进行图书数据管理的一方，只要导入 internal/store 这个包，就可以自动完成注册动作

 igo@igoMacBook bookstore % cat internal/store/memstore.go
package store

import (
    mystore "bookstore/store"
    factory "bookstore/store/factory"
    "sync"
)

func init() {
    factory.Register("mem", &MemStore{
        books: make(map[string]*mystore.Book),
    })
}

type MemStore struct {
    sync.RWMutex
    books map[string]*mystore.Book
}

// Create creates a new Book in the store.
func (ms *MemStore) Create(book *mystore.Book) error {
    ms.Lock()
    defer ms.Unlock()

    if _, ok := ms.books[book.Id]; ok {
        return mystore.ErrExist
    }

    nBook := *book
    ms.books[book.Id] = &nBook

    return nil
}

// Update updates the existed Book in the store.
func (ms *MemStore) Update(book *mystore.Book) error {
    ms.Lock()
    defer ms.Unlock()

    oldBook, ok := ms.books[book.Id]
    if !ok {
        return mystore.ErrNotFound
    }

    nBook := *oldBook
    if book.Name != "" {
        nBook.Name = book.Name
    }

    if book.Authors != nil {
        nBook.Authors = book.Authors
    }

    if book.Press != "" {
        nBook.Press = book.Press
    }

    ms.books[book.Id] = &nBook

    return nil
}

// Get retrieves a book from the store, by id. If no such id exists. an
// error is returned.
func (ms *MemStore) Get(id string) (mystore.Book, error) {
    ms.RLock()
    defer ms.RUnlock()

    t, ok := ms.books[id]
    if ok {
        return *t, nil
    }
    return mystore.Book{}, mystore.ErrNotFound
}

// Delete deletes the book with the given id. If no such id exist. an error
// is returned.
func (ms *MemStore) Delete(id string) error {
    ms.Lock()
    defer ms.Unlock()

    if _, ok := ms.books[id]; !ok {
        return mystore.ErrNotFound
    }

    delete(ms.books, id)
    return nil
}

// GetAll returns all the books in the store, in arbitrary order.
func (ms *MemStore) GetAll() ([]mystore.Book, error) {
    ms.RLock()
    defer ms.RUnlock()

    allBooks := make([]mystore.Book, 0, len(ms.books))
    for _, book := range ms.books {
        allBooks = append(allBooks, *book)
    }
    return allBooks, nil
}
```

```
// server/server.go
函数 NewBookStoreServer 接受两个参数，一个是 HTTP 服务监听的服务地址，另外一个是实现了 store.Store 接口的类型实例。这种函数原型的设计是 Go 语言的一种惯用设计方法，也就是接受一个接口类型参数，返回一个具体类型。返回的具体类型组合了传入的接口类型的能力

createBookHandler和getBookHandler大致逻辑:
先获取 http 请求包体数据，然后通过标准库 json 包将这些数据，解码（decode）为我们需要的 store.Book 结构体实例，再通过 Store 接口对图书数据进行存储操作。如果我们是获取图书数据的请求，那么处理函数将通过 response 函数，把取出的图书数据编码到 http 响应的包体中，并返回给客户端

//ListenAndServe函数把 BookStoreServer 内部的 http.Server 的运行，放置到一个单独的轻量级线程 Goroutine 中。这是因为，http.Server.ListenAndServe 会阻塞代码的继续运行，如果不把它放在单独的 Goroutine 中，后面的代码将无法得到执行

//为了检测到 http.Server.ListenAndServe 的运行状态，我们再通过一个 channel（errChan），在新创建的 Goroutine 与主 Goroutine 之间建立的通信渠道。通过这个渠道，这样我们能及时得到 http server 的运行状态

 igo@igoMacBook bookstore % cat server/server.go
package server

import (
    "bookstore/server/middleware"
    "bookstore/store"
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/gorilla/mux"
)

type BookStoreServer struct {
    s   store.Store
    srv *http.Server
}

func NewBookStoreServer(addr string, s store.Store) *BookStoreServer {
    srv := &BookStoreServer{
        s: s,
        srv: &http.Server{
            Addr: addr,
        },
    }

    router := mux.NewRouter()
    router.HandleFunc("/book", srv.createBookHandler).Methods("POST")
    router.HandleFunc("/book/{id}", srv.updateBookHandler).Methods("POST")
    router.HandleFunc("/book/{id}", srv.getBookHandler).Methods("GET")
    router.HandleFunc("/book", srv.getAllBooksHandler).Methods("GET")
    router.HandleFunc("/book/{id}", srv.delBookHandler).Methods("DELETE")

    srv.srv.Handler = middleware.Logging(middleware.Validating(router))
    return srv
}

func (bs *BookStoreServer) ListenAndServe() (<-chan error, error) {
    var err error
    errChan := make(chan error)
    go func() {
        err = bs.srv.ListenAndServe()
        errChan <- err
    }()

    select {
    case err = <-errChan:
        return nil, err
    case <-time.After(time.Second):
        return errChan, nil
    }
}

func (bs *BookStoreServer) Shutdown(ctx context.Context) error {
    return bs.srv.Shutdown(ctx)
}

func (bs *BookStoreServer) createBookHandler(w http.ResponseWriter, req *http.Request) {
    dec := json.NewDecoder(req.Body)
    var book store.Book
    if err := dec.Decode(&book); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := bs.s.Create(&book); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
}

func (bs *BookStoreServer) updateBookHandler(w http.ResponseWriter, req *http.Request) {
    id, ok := mux.Vars(req)["id"]
    if !ok {
        http.Error(w, "no id found in request", http.StatusBadRequest)
        return
    }

    dec := json.NewDecoder(req.Body)
    var book store.Book
    if err := dec.Decode(&book); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    book.Id = id
    if err := bs.s.Update(&book); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
}

func (bs *BookStoreServer) getBookHandler(w http.ResponseWriter, req *http.Request) {
    id, ok := mux.Vars(req)["id"]
    if !ok {
        http.Error(w, "no id found in request", http.StatusBadRequest)
        return
    }

    book, err := bs.s.Get(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    response(w, book)
}

func (bs *BookStoreServer) getAllBooksHandler(w http.ResponseWriter, req *http.Request) {
    books, err := bs.s.GetAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    response(w, books)
}

func (bs *BookStoreServer) delBookHandler(w http.ResponseWriter, req *http.Request) {
    id, ok := mux.Vars(req)["id"]
    if !ok {
        http.Error(w, "no id found in request", http.StatusBadRequest)
        return
    }

    err := bs.s.Delete(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
}

func response(w http.ResponseWriter, v interface{}) {
    data, err := json.Marshal(v)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(data)
}
igo@igoMacBook bookstore %
```

```
// server/middleware/middleware.go
middleware 就是一些通用的 http 处理函数
Logging 函数主要用来输出每个到达的 HTTP 请求的一些概要信息，而 Validating 则会对每个 http 请求的头部进行检查，检查 Content-Type 头字段所表示的媒体类型是否为 application/json。这些通用的 middleware 函数，会被串联到每个真正的处理函数之前，避免我们在每个处理函数中重复实现这些逻辑

  func Logging(next http.Handler) http.Handler {
     return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
         log.Printf("recv a %s request from %s", req.Method, req.RemoteAddr)
         next.ServeHTTP(w, req)
     })
 }
 
 func Validating(next http.Handler) http.Handler {
     return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
         contentType := req.Header.Get("Content-Type")
         mediatype, _, err := mime.ParseMediaType(contentType)
         if err != nil {
             http.Error(w, err.Error(), http.StatusBadRequest)
             return
         }
         if mediatype != "application/json" {
             http.Error(w, "invalid Content-Type", http.StatusUnsupportedMediaType)
             return
         }
         next.ServeHTTP(w, req)
     })
 }
```

```
makeRun

Make:
igo@igoMacBook golang %go mod tidy
go: finding module for package github.com/gorilla/mux
go: found github.com/gorilla/mux in github.com/gorilla/mux v1.8.0

igo@igoMacBook golang %go build bookstore/cmd/bookstore
igo@igoMacBook golang %./bookstore
2021/10/05 16:08:36 web server start ok

Run1:
$curl -X POST -H "Content-Type:application/json" -d '{"id": "978-7-111-55842-2", "name": "The Go Programming Language", "authors":["Alan A.A.Donovan", "Brian W. Kergnighan"],"press": "Pearson Education"}' localhost:8080/book
serv端：
2021/10/05 16:09:10 recv a POST request from [::1]:58021

Run2:
$curl -X GET -H "Content-Type:application/json" localhost:8080/book/978-7-111-55842-2

{"id":"978-7-111-55842-2","name":"The Go Programming Language","authors":["Alan A.A.Donovan","Brian W. Kergnighan"],"press":"Pearson Education"}
```
