# Go Language Idioms & Best Practices

## Core Language Idioms

### Interface Design

#### Accept Interfaces, Return Structs
```go
// Good: Accept interface, return concrete type
func NewProcessor(r io.Reader) *Processor {
    return &Processor{reader: r}
}

// Bad: Return interface when not necessary
func NewProcessor(r io.Reader) ProcessorInterface {
    return &Processor{reader: r}
}
```

#### Small Interface Principle
```go
// Excellent: Single method interface
type Reader interface {
    Read([]byte) (int, error)
}

// Good: Related methods
type ReadWriter interface {
    Reader
    Writer
}

// Avoid: Large interfaces
type MassiveInterface interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Close() error
    Seek(int64, int) (int64, error)
    // ... 20+ more methods
}
```

### Error Handling Idioms

#### Guard Clause Pattern
```go
func processData(data []byte) error {
    if len(data) == 0 {
        return errors.New("no data provided")
    }
    if !isValid(data) {
        return errors.New("invalid data format")
    }
    // main logic here
    return nil
}
```

#### Error Wrapping
```go
func readConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
    }
    // process data
    return config, nil
}
```

#### Sentinel Errors
```go
var (
    ErrNotFound = errors.New("item not found")
    ErrInvalid  = errors.New("invalid input")
)

func findItem(id string) (*Item, error) {
    // search logic
    if notFound {
        return nil, ErrNotFound
    }
    return item, nil
}

// Usage
item, err := findItem("123")
if errors.Is(err, ErrNotFound) {
    // handle not found case
}
```

### Zero Value Idioms

#### Design for Zero Value
```go
// Good: Mutex works with zero value
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}

// Buffer works with zero value
var buf bytes.Buffer
buf.WriteString("hello")
```

#### Initialization Methods
```go
// When zero value isn't sufficient
type Database struct {
    conn *sql.DB
    cfg  Config
}

func NewDatabase(dsn string) (*Database, error) {
    conn, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    return &Database{conn: conn}, nil
}
```

### Struct Patterns

#### Embedding for Composition
```go
type Engine struct {
    Power int
}

func (e Engine) Start() {
    fmt.Println("Engine starting...")
}

type Car struct {
    Engine  // embedded
    Model   string
}

// Car automatically gets Start() method
car := Car{Engine{200}, "Tesla"}
car.Start() // calls Engine.Start()
```

#### Options Pattern
```go
type Server struct {
    addr     string
    timeout  time.Duration
    maxConns int
}

type Option func(*Server)

func WithTimeout(t time.Duration) Option {
    return func(s *Server) {
        s.timeout = t
    }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:     addr,
        timeout:  30 * time.Second, // default
        maxConns: 100,              // default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage
server := NewServer(":8080", WithTimeout(60*time.Second))
```

## Concurrency Idioms

### Goroutine Management

#### Always Handle Goroutine Lifecycle
```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go worker(ctx)
    
    // Wait for signal
    <-time.After(10 * time.Second)
    cancel() // Signal goroutines to stop
    
    // Wait for cleanup (production would use WaitGroup)
    time.Sleep(100 * time.Millisecond)
}

func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-time.After(1 * time.Second):
            fmt.Println("working...")
        }
    }
}
```

#### Worker Pool Pattern
```go
func workerPool(jobs <-chan Job, results chan<- Result) {
    const numWorkers = 5
    var wg sync.WaitGroup
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- processJob(job)
            }
        }()
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
}
```

### Channel Patterns

#### Pipeline Pattern
```go
func pipeline() {
    input := make(chan int)
    output := make(chan int)
    
    // Stage 1
    go func() {
        defer close(input)
        for i := 0; i < 10; i++ {
            input <- i
        }
    }()
    
    // Stage 2
    go func() {
        defer close(output)
        for num := range input {
            output <- num * 2
        }
    }()
    
    // Consumer
    for result := range output {
        fmt.Println(result)
    }
}
```

#### Fan-out/Fan-in Pattern
```go
func fanOutFanIn(input <-chan int) <-chan int {
    numWorkers := 3
    workers := make([]<-chan int, numWorkers)
    
    // Fan-out
    for i := 0; i < numWorkers; i++ {
        output := make(chan int)
        workers[i] = output
        go func() {
            defer close(output)
            for n := range input {
                output <- expensiveOperation(n)
            }
        }()
    }
    
    // Fan-in
    return fanIn(workers...)
}

func fanIn(inputs ...<-chan int) <-chan int {
    output := make(chan int)
    var wg sync.WaitGroup
    
    for _, input := range inputs {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for n := range ch {
                output <- n
            }
        }(input)
    }
    
    go func() {
        wg.Wait()
        close(output)
    }()
    
    return output
}
```

## Naming Idioms

### Package Naming
```go
// Good: Clear, short package names
package http    // not httputil
package url     // not urlparser
package json    // not jsonencoder

// Package names inform usage
http.Get()      // clear what package this is from
json.Marshal()  // obvious functionality
```

### Method Naming
```go
type User struct {
    name  string
    email string
}

// Good: No "Get" prefix for getters
func (u *User) Name() string { return u.name }
func (u *User) Email() string { return u.email }

// Good: Setter pattern
func (u *User) SetName(name string) { u.name = name }

// Good: Boolean methods start with "Is", "Has", "Can"
func (u *User) IsActive() bool { return u.active }
func (u *User) HasEmail() bool { return u.email != "" }
```

### Variable Naming
```go
// Short names for short scopes
for i, user := range users {
    if user.IsActive() {
        // i and user are clear in this context
    }
}

// Longer names for larger scopes
func processUserRegistration(registrationRequest *RegistrationRequest) error {
    // longer names for function scope
    validatedUser, err := validateUser(registrationRequest.User)
    if err != nil {
        return fmt.Errorf("user validation failed: %w", err)
    }
    // ...
}
```

## Testing Idioms

### Table-Driven Tests
```go
func TestParseInt(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr bool
    }{
        {
            name:  "valid positive number",
            input: "123",
            want:  123,
        },
        {
            name:    "invalid input",
            input:   "abc",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseInt(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseInt() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ParseInt() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Helpers
```go
func TestHandler(t *testing.T) {
    server := newTestServer(t)
    defer server.Close()
    
    resp := makeRequest(t, server.URL+"/api/users")
    assertStatusCode(t, resp, http.StatusOK)
}

func newTestServer(t *testing.T) *httptest.Server {
    t.Helper() // marks this as helper function
    return httptest.NewServer(handler())
}

func assertStatusCode(t *testing.T, resp *http.Response, want int) {
    t.Helper()
    if resp.StatusCode != want {
        t.Errorf("got status %d, want %d", resp.StatusCode, want)
    }
}
```

### Mocking with Interfaces
```go
type UserService interface {
    GetUser(id string) (*User, error)
}

type MockUserService struct {
    users map[string]*User
}

func (m *MockUserService) GetUser(id string) (*User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, errors.New("user not found")
}

func TestUserHandler(t *testing.T) {
    mockService := &MockUserService{
        users: map[string]*User{
            "123": {ID: "123", Name: "John"},
        },
    }
    
    handler := NewUserHandler(mockService)
    // test handler with mock
}
```

## Performance Idioms

### String Building
```go
// Inefficient: creates multiple strings
func buildString() string {
    var result string
    for i := 0; i < 1000; i++ {
        result += fmt.Sprintf("item %d ", i)
    }
    return result
}

// Efficient: use strings.Builder
func buildStringEfficient() string {
    var builder strings.Builder
    for i := 0; i < 1000; i++ {
        builder.WriteString(fmt.Sprintf("item %d ", i))
    }
    return builder.String()
}
```

### Slice Preallocation
```go
// Inefficient: slice grows dynamically
func processItems(items []Item) []Result {
    var results []Result
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// Efficient: preallocate slice
func processItemsEfficient(items []Item) []Result {
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}
```

### Avoid Unnecessary Allocations
```go
// Inefficient: allocates on each call
func formatMessage(msg string) string {
    return fmt.Sprintf("[INFO] %s", msg)
}

// Better: reuse buffer
var msgBuffer = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 64))
    },
}

func formatMessageEfficient(msg string) string {
    buf := msgBuffer.Get().(*bytes.Buffer)
    defer msgBuffer.Put(buf)
    buf.Reset()
    
    buf.WriteString("[INFO] ")
    buf.WriteString(msg)
    return buf.String()
}
```

## JSON Handling Idioms

### Struct Tags
```go
type User struct {
    ID       int       `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email,omitempty"`
    Password string    `json:"-"` // never serialize
    Created  time.Time `json:"created_at"`
}
```

### Custom JSON Marshaling
```go
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    dur, err := time.ParseDuration(s)
    *d = Duration(dur)
    return err
}
```

## HTTP Client/Server Idioms

### HTTP Client with Timeout
```go
func makeHTTPRequest(url string) (*http.Response, error) {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    return client.Do(req)
}
```

### HTTP Handler Pattern
```go
func userHandler(db *Database) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID := r.URL.Query().Get("id")
        if userID == "" {
            http.Error(w, "missing user ID", http.StatusBadRequest)
            return
        }
        
        user, err := db.GetUser(userID)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(user)
    }
}
```

## Anti-Idioms (What NOT to Do)

### Don't Use Empty Interface Unnecessarily
```go
// Bad: loses type safety
func process(data interface{}) {
    // requires type assertions
}

// Good: use generics (Go 1.18+) or specific types
func process[T any](data T) T {
    return data
}
```

### Don't Ignore Errors
```go
// Bad: ignoring errors
data, _ := ioutil.ReadFile("config.json")

// Good: handle errors appropriately
data, err := ioutil.ReadFile("config.json")
if err != nil {
    log.Fatalf("failed to read config: %v", err)
}
```

### Don't Use Goroutines for Everything
```go
// Bad: unnecessary goroutine
func simpleCalculation(x, y int) int {
    result := make(chan int)
    go func() {
        result <- x + y
    }()
    return <-result
}

// Good: direct calculation
func simpleCalculation(x, y int) int {
    return x + y
}
```

These idioms represent the accumulated wisdom of the Go community and should guide your Go development practices.