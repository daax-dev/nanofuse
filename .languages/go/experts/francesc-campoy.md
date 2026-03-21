# Francesc Campoy - Go Education Pioneer & Practical Learning Expert

## Expertise Focus
**Go Education • Developer Tooling • Code Generation • Web Development • Community Building**

- **Current Role**: Software Engineer at Apple (formerly Google Go Team Developer Advocate)
- **Key Contribution**: JustForFunc educational series, Go tooling mastery, practical Go education at scale
- **Learning Focus**: Go tooling ecosystem, code generation patterns, web development, educational methodologies

## Direct Learning Resources

### Essential Educational Series

#### **[JustForFunc YouTube Series](https://www.youtube.com/channel/UC_BzFbxG2za3bp5NRRRXJSw)**
- **GitHub**: [github.com/campoy/justforfunc](https://github.com/campoy/justforfunc)
- **Episodes**: 100+ practical Go programming episodes
- **Learn**: Real-world Go development, live coding, debugging, problem-solving
- **Go Concepts**: Context usage, gRPC, dependency injection, Kubernetes client
- **Apply**: Learning Go through practical examples and live coding sessions

#### **Key JustForFunc Episodes**:

##### **["Understanding the Context Package"](https://www.youtube.com/watch?v=LSzR0VEraWw)**
```go
// Campoy's context patterns from JustForFunc episodes
func fetchUserData(ctx context.Context, userID string) (*User, error) {
    // Context with timeout for external API calls
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Channel to receive result or timeout
    resultCh := make(chan *User, 1)
    errorCh := make(chan error, 1)
    
    go func() {
        user, err := api.GetUser(userID)
        if err != nil {
            errorCh <- err
            return
        }
        resultCh <- user
    }()
    
    select {
    case user := <-resultCh:
        return user, nil
    case err := <-errorCh:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

##### **["gRPC and Protocol Buffers"](https://www.youtube.com/watch?v=uolTUtioIrc)**
- **Learn**: gRPC service implementation, protocol buffers, streaming
- **Pattern**: Microservices communication, type-safe APIs
- **Apply**: Building distributed systems with Go and gRPC

### Key GitHub Projects

#### **[go-tooling-workshop](https://github.com/campoy/go-tooling-workshop)**
- **Stars**: 2.8k+ | **Description**: Comprehensive Go tooling workshop
- **Learn**: go build, test, fmt, vet, lint, mod, generate
- **Pattern**: Complete development workflow with Go tools
- **Apply**: Mastering Go developer productivity and code quality

```bash
# Campoy's Go tooling workflow from workshop
# Format code
go fmt ./...

# Check for common errors
go vet ./...

# Run tests with coverage
go test -cover ./...

# Check for race conditions
go test -race ./...

# Generate documentation
godoc -http=:8080

# Profile performance
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

#### **[embedmd](https://github.com/campoy/embedmd)**
- **Stars**: 1k+ | **Description**: Embed code blocks in markdown files
- **Learn**: Documentation automation, code synchronization
- **Pattern**: Keeping code and documentation in sync
- **Apply**: Automated documentation generation and maintenance

```go
// embedmd usage pattern from Campoy
//go:generate embedmd -w README.md

// Embed code directly in markdown:
// [embedmd]:# (example.go go /func main/ /^}/)

func main() {
    fmt.Println("This code is embedded in README.md")
    fmt.Println("And stays synchronized with embedmd")
}
```

#### **[jsonenums](https://github.com/campoy/jsonenums)**
- **Stars**: 400+ | **Description**: Type-safe JSON marshaling for Go enums
- **Learn**: Code generation, type safety, JSON handling
- **Pattern**: Compile-time safety for enum marshaling
- **Apply**: Building robust APIs with type-safe enums

```go
// Campoy's enum pattern with jsonenums
//go:generate jsonenums -type=Status
type Status int

const (
    StatusPending Status = iota
    StatusActive
    StatusInactive
)

// Generated methods provide type-safe JSON marshaling
func ExampleAPI() {
    status := StatusActive
    data, _ := json.Marshal(status) // "active"
    
    var parsed Status
    json.Unmarshal(data, &parsed) // Type-safe unmarshaling
}
```

### Conference Talks & Workshops

#### **["Go Tooling in Action"](https://www.youtube.com/watch?v=uBjoTxosSys)**
- **Duration**: 45 minutes | **Event**: GopherCon 2016
- **Learn**: Complete Go tooling ecosystem, workflow optimization
- **Go Concepts**: go generate, build tags, testing, profiling
- **Apply**: Setting up efficient Go development environments

#### **["Machine Learning with Go"](https://www.youtube.com/watch?v=Xt5M6rFfh7s)**
- **Duration**: 30 minutes | **Event**: GopherCon 2017
- **Learn**: Go for data science, TensorFlow integration, ML pipelines
- **Go Concepts**: CGO, numerical computing, concurrency for ML
- **Apply**: Building machine learning systems in Go

#### **["Understanding Go Interfaces"](https://www.youtube.com/watch?v=F4wUrj6pmSI)**
- **Duration**: 25 minutes | **Event**: dotGo 2017
- **Learn**: Interface design patterns, composition, Duck typing
- **Apply**: Building flexible, testable Go architectures

### Educational Workshop Materials

#### **[go-web-workshop](https://github.com/campoy/go-web-workshop)**
- **GitHub**: Complete web development workshop with Go
- **Learn**: HTTP servers, templates, App Engine deployment
- **Pattern**: Full-stack Go web development
- **Apply**: Building production web applications with Go

```go
// Campoy's web development patterns from workshop
package main

import (
    "html/template"
    "net/http"
    "log"
)

type Server struct {
    templates *template.Template
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
    // Campoy emphasizes proper error handling in web apps
    if err := s.templates.ExecuteTemplate(w, "home.html", nil); err != nil {
        log.Printf("Template error: %v", err)
        http.Error(w, "Internal Server Error", 500)
        return
    }
}

func main() {
    server := &Server{
        templates: template.Must(template.ParseGlob("templates/*.html")),
    }
    
    http.HandleFunc("/", server.handleHome)
    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Code Generation Expertise

### Generate-First Development
```go
// Campoy's code generation philosophy
// Use go:generate for repetitive code

//go:generate stringer -type=Color
type Color int

const (
    Red Color = iota
    Green
    Blue
)

//go:generate mockgen -source=user.go -destination=mocks/user.go
type UserService interface {
    GetUser(id string) (*User, error)
    UpdateUser(*User) error
}

//go:generate embedmd -w README.md
// Embed code examples in documentation

// Run all generators with: go generate ./...
```

### Template-Based Code Generation
```go
// Campoy's template approach for custom generators
package main

import (
    "go/ast"
    "go/parser"
    "go/token"
    "text/template"
)

const crudTemplate = `
// Generated CRUD operations for {{.Type}}
func (s *Service) Create{{.Type}}(item *{{.Type}}) error {
    return s.db.Create(item)
}

func (s *Service) Get{{.Type}}(id string) (*{{.Type}}, error) {
    var item {{.Type}}
    err := s.db.First(&item, "id = ?", id)
    return &item, err
}
`

func generateCRUD(typeName string) {
    tmpl := template.Must(template.New("crud").Parse(crudTemplate))
    data := struct{ Type string }{Type: typeName}
    // Generate code using template
    tmpl.Execute(os.Stdout, data)
}
```

## Educational Methodology

### Live Coding Approach
- **Start with problems**: Real-world scenarios drive learning
- **Make mistakes intentionally**: Show debugging process
- **Incremental development**: Build complexity gradually
- **Community interaction**: Respond to live questions

### Error-Driven Development
```go
// Campoy's teaching pattern: embrace errors as learning opportunities
func demonstrateErrorHandling() {
    // Step 1: Show the wrong way
    // data := readFile("config.json") // Ignoring errors
    
    // Step 2: Show the right way with explanation
    data, err := readFile("config.json")
    if err != nil {
        // Explain different error handling strategies
        log.Printf("Failed to read config: %v", err)
        return
    }
    
    // Step 3: Show error wrapping for better context
    if err != nil {
        return fmt.Errorf("loading configuration: %w", err)
    }
}
```

## Developer Tooling Philosophy

### Tool Integration Patterns
```go
// Campoy's approach to Go tooling integration
// Makefile for consistent development workflow

# From go-tooling-workshop
.PHONY: build test lint fmt vet

build:
	go build -v ./...

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

vet:
	go vet ./...

# Generate all code
generate:
	go generate ./...

# Full check before commit
check: fmt vet lint test
```

## Learning Resources

### Study Campoy's Work
1. **JustForFunc Series**: [YouTube Channel](https://www.youtube.com/channel/UC_BzFbxG2za3bp5NRRRXJSw)
2. **Go Tooling Workshop**: [github.com/campoy/go-tooling-workshop](https://github.com/campoy/go-tooling-workshop)
3. **Code Generation Tools**: Study embedmd and jsonenums source
4. **Web Workshop**: [github.com/campoy/go-web-workshop](https://github.com/campoy/go-web-workshop)

### Essential Campoy Principles
- **"Learning by doing"** - Practical examples over theory
- **"Make mistakes publicly"** - Debugging is part of development
- **"Tools should be simple"** - Automate repetitive tasks
- **"Community learns together"** - Share knowledge openly

## For AI Agents
- **Reference Campoy's educational patterns** for explaining complex Go concepts
- **Apply his code generation techniques** for automating repetitive code
- **Use his tooling workflows** for development environment setup
- **Follow his error-driven approach** when demonstrating debugging

## For Human Engineers
- **Watch JustForFunc series** for practical Go development techniques
- **Complete go-tooling-workshop** to master Go development workflow
- **Use embedmd for documentation** to keep code and docs synchronized
- **Apply code generation patterns** from jsonenums and custom tools
- **Follow his teaching methodology** when mentoring other developers

## Key Insights
Campoy's work demonstrates that great technical education comes from combining practical examples with community engagement. His tools solve real developer pain points, and his teaching methodology makes complex topics accessible through live coding and error-driven development.

**Core Lesson**: The best way to learn Go is by building real projects, making mistakes publicly, and using tools to automate repetitive tasks. Focus on practical problems and let the community learn together through shared experiences.