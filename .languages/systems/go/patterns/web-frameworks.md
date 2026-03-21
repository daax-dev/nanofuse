# Go Web Framework Patterns

## Production Web Frameworks

### Gin - High Performance HTTP Framework
- **Repository**: github.com/gin-gonic/gin
- **Stars**: 77k+
- **Use Case**: High-performance REST APIs and web services
- **Key Features**:
  - Fast routing with radix tree
  - Middleware support
  - JSON validation and rendering
  - Error management
  - Built-in rendering (JSON, XML, HTML)

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // Middleware
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    
    // Routes
    r.GET("/users/:id", getUserHandler)
    r.POST("/users", createUserHandler)
    
    r.Run(":8080")
}

func getUserHandler(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{"user_id": id})
}
```

### Echo - Minimalist Web Framework
- **Repository**: github.com/labstack/echo
- **Stars**: 29k+
- **Use Case**: Lightweight APIs with minimal overhead
- **Key Features**:
  - Optimized router
  - Extensible middleware
  - Data binding and validation
  - Templating support
  - WebSocket support

```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    e := echo.New()
    
    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    // Routes
    e.GET("/users/:id", getUser)
    e.POST("/users", createUser)
    
    e.Logger.Fatal(e.Start(":8080"))
}

func getUser(c echo.Context) error {
    id := c.Param("id")
    return c.JSON(http.StatusOK, map[string]string{
        "user_id": id,
    })
}
```

### Fiber - Express-inspired Framework
- **Repository**: github.com/gofiber/fiber
- **Stars**: 33k+
- **Use Case**: Node.js developers transitioning to Go
- **Key Features**:
  - Express-like API
  - Fast HTTP engine (fasthttp)
  - Built-in rate limiting
  - WebSocket support
  - Template engines

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
    app := fiber.New()
    
    // Middleware
    app.Use(logger.New())
    
    // Routes
    app.Get("/users/:id", getUser)
    app.Post("/users", createUser)
    
    log.Fatal(app.Listen(":8080"))
}

func getUser(c *fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(fiber.Map{
        "user_id": id,
    })
}
```

### Chi - Lightweight Router
- **Repository**: github.com/go-chi/chi
- **Stars**: 18k+
- **Use Case**: Composable HTTP services with stdlib
- **Key Features**:
  - net/http compatible
  - Lightweight and fast
  - Middleware stack composition
  - Route groups and sub-routers
  - Context-based request handling

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    // Routes
    r.Route("/users", func(r chi.Router) {
        r.Get("/{id}", getUserHandler)
        r.Post("/", createUserHandler)
    })
    
    http.ListenAndServe(":8080", r)
}
```

### Gorilla Mux - Powerful URL Router
- **Repository**: github.com/gorilla/mux
- **Stars**: 20k+
- **Use Case**: Complex routing with standard library
- **Key Features**:
  - Flexible request matching
  - URL variables and patterns
  - Middleware support
  - Subrouter support
  - HTTP method matching

```go
package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

func main() {
    r := mux.NewRouter()
    
    // Routes with patterns
    r.HandleFunc("/users/{id:[0-9]+}", getUserHandler).Methods("GET")
    r.HandleFunc("/users", createUserHandler).Methods("POST")
    
    // Subrouters
    api := r.PathPrefix("/api/v1").Subrouter()
    api.HandleFunc("/health", healthHandler)
    
    http.ListenAndServe(":8080", r)
}
```

## Framework Selection Criteria

### Performance Benchmarks
```
Framework    | Requests/sec | Memory Usage | Features
-------------|--------------|--------------|----------
Gin          | ~40k        | Low          | High
Fiber        | ~50k        | Medium       | High  
Echo         | ~35k        | Low          | Medium
Chi          | ~38k        | Very Low     | Medium
Gorilla/Mux  | ~30k        | Low          | High
```

### Use Case Matrix
```
Framework | API Development | Web Apps | Microservices | Learning Curve
----------|----------------|----------|---------------|---------------
Gin       | Excellent      | Good     | Excellent     | Easy
Fiber     | Excellent      | Good     | Good          | Easy (JS devs)
Echo      | Excellent      | Good     | Good          | Easy
Chi       | Good           | Excellent| Good          | Medium
Gorilla   | Good           | Excellent| Fair          | Medium
```

## Common Web Patterns

### Middleware Pattern
```go
// Authentication middleware
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !isValidToken(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Handler Patterns
```go
// Handler with dependency injection
type UserHandler struct {
    userService UserService
    logger      *log.Logger
}

func NewUserHandler(service UserService, logger *log.Logger) *UserHandler {
    return &UserHandler{
        userService: service,
        logger:      logger,
    }
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    
    user, err := h.userService.GetUser(id)
    if err != nil {
        h.logger.Printf("Failed to get user %s: %v", id, err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// Handler factory pattern
func makeHandler(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := fn(w, r); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}
```

### Server Configuration Pattern
```go
type Server struct {
    router      *mux.Router
    httpServer  *http.Server
    logger      *log.Logger
    config      Config
}

type Config struct {
    Port            string        `env:"PORT" envDefault:"8080"`
    ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"30s"`
    WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
    ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
}

func NewServer(cfg Config, logger *log.Logger) *Server {
    s := &Server{
        router: mux.NewRouter(),
        logger: logger,
        config: cfg,
    }
    
    s.httpServer = &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      s.router,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
    }
    
    s.setupRoutes()
    return s
}

func (s *Server) Start() error {
    s.logger.Printf("Starting server on port %s", s.config.Port)
    return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.httpServer.Shutdown(ctx)
}
```

### Graceful Shutdown Pattern
```go
func main() {
    server := NewServer(config, logger)
    
    // Start server in goroutine
    go func() {
        if err := server.Start(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
}
```

## Testing Web Applications

### HTTP Testing Patterns
```go
func TestUserHandler(t *testing.T) {
    // Setup
    mockService := &MockUserService{}
    handler := NewUserHandler(mockService, log.New(os.Stderr, "", log.LstdFlags))
    
    // Test cases
    tests := []struct {
        name           string
        userID         string
        mockUser       *User
        mockError      error
        expectedStatus int
    }{
        {
            name:           "valid user",
            userID:         "123",
            mockUser:       &User{ID: "123", Name: "John"},
            expectedStatus: http.StatusOK,
        },
        {
            name:           "user not found",
            userID:         "456",
            mockError:      errors.New("user not found"),
            expectedStatus: http.StatusInternalServerError,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock
            mockService.On("GetUser", tt.userID).Return(tt.mockUser, tt.mockError)
            
            // Create request
            req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
            w := httptest.NewRecorder()
            
            // Execute
            handler.GetUser(w, req)
            
            // Assert
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            // Cleanup mock
            mockService.AssertExpectations(t)
        })
    }
}
```

### Integration Testing
```go
func TestAPIIntegration(t *testing.T) {
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    client := &http.Client{Timeout: 5 * time.Second}
    
    t.Run("create and get user", func(t *testing.T) {
        // Create user
        user := User{Name: "John Doe", Email: "john@example.com"}
        userJSON, _ := json.Marshal(user)
        
        resp, err := client.Post(
            server.URL+"/users",
            "application/json",
            bytes.NewBuffer(userJSON),
        )
        assert.NoError(t, err)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
        
        var createdUser User
        json.NewDecoder(resp.Body).Decode(&createdUser)
        resp.Body.Close()
        
        // Get user
        resp, err = client.Get(server.URL + "/users/" + createdUser.ID)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var retrievedUser User
        json.NewDecoder(resp.Body).Decode(&retrievedUser)
        resp.Body.Close()
        
        assert.Equal(t, user.Name, retrievedUser.Name)
        assert.Equal(t, user.Email, retrievedUser.Email)
    })
}

func setupTestServer(t *testing.T) *httptest.Server {
    handler := setupTestRoutes(t)
    return httptest.NewServer(handler)
}
```

These patterns provide a solid foundation for building production-ready web applications in Go.