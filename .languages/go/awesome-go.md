# Awesome Go - Systems Programming and Cloud-Native Development

## Overview
This curated collection enhances and expands upon the excellent [awesome-go](https://github.com/avelino/awesome-go) repository with production-ready insights, cloud-native patterns, performance optimization techniques, and real-world implementation guidance for modern Go development.

## Table of Contents
- [Core Libraries and Frameworks](#core-libraries-and-frameworks)
- [Web Frameworks and HTTP](#web-frameworks-and-http)
- [Database and Storage](#database-and-storage)
- [Microservices and Distributed Systems](#microservices-and-distributed-systems)
- [Cloud-Native Development](#cloud-native-development)
- [Concurrency and Goroutines](#concurrency-and-goroutines)
- [CLI and Terminal Applications](#cli-and-terminal-applications)
- [Testing and Quality Assurance](#testing-and-quality-assurance)
- [Performance and Profiling](#performance-and-profiling)
- [Security and Cryptography](#security-and-cryptography)
- [DevOps and Infrastructure](#devops-and-infrastructure)
- [Learning Resources](#learning-resources)

## Core Libraries and Frameworks

### Standard Library Excellence
- **Go Standard Library** - Comprehensive, production-ready standard library
  - *HTTP Server*: Production-ready HTTP server and client
  - *Concurrency*: Goroutines and channels for concurrent programming
  - *JSON/XML*: High-performance serialization
  ```go
  package main

  import (
      \"context\"
      \"encoding/json\"
      \"fmt\"
      \"log\"
      \"net/http\"
      \"time\"
  )

  type User struct {
      ID       int       `json:\"id\"`
      Name     string    `json:\"name\"`
      Email    string    `json:\"email\"`
      CreateAt time.Time `json:\"created_at\"`
  }

  type Server struct {
      users []User
      mux   *http.ServeMux
  }

  func NewServer() *Server {
      s := &Server{
          users: []User{},
          mux:   http.NewServeMux(),
      }
      s.routes()
      return s
  }

  func (s *Server) routes() {
      s.mux.HandleFunc(\"/users\", s.handleUsers)
      s.mux.HandleFunc(\"/health\", s.handleHealth)
  }

  func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
      switch r.Method {
      case http.MethodGet:
          s.getUsersHandler(w, r)
      case http.MethodPost:
          s.createUserHandler(w, r)
      default:
          http.Error(w, \"Method not allowed\", http.StatusMethodNotAllowed)
      }
  }

  func (s *Server) getUsersHandler(w http.ResponseWriter, r *http.Request) {
      w.Header().Set(\"Content-Type\", \"application/json\")
      if err := json.NewEncoder(w).Encode(s.users); err != nil {
          http.Error(w, \"Failed to encode response\", http.StatusInternalServerError)
          return
      }
  }

  func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
      var user User
      if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
          http.Error(w, \"Invalid JSON\", http.StatusBadRequest)
          return
      }

      user.ID = len(s.users) + 1
      user.CreateAt = time.Now()
      s.users = append(s.users, user)

      w.Header().Set(\"Content-Type\", \"application/json\")
      w.WriteHeader(http.StatusCreated)
      json.NewEncoder(w).Encode(user)
  }

  func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
      response := map[string]interface{}{
          \"status\":    \"healthy\",
          \"timestamp\": time.Now().Unix(),
          \"uptime\":    time.Since(startTime).Seconds(),
      }
      w.Header().Set(\"Content-Type\", \"application/json\")
      json.NewEncoder(w).Encode(response)
  }

  var startTime = time.Now()

  func main() {
      server := NewServer()

      httpServer := &http.Server{
          Addr:         \":8080\",
          Handler:      server.mux,
          ReadTimeout:  15 * time.Second,
          WriteTimeout: 15 * time.Second,
          IdleTimeout:  60 * time.Second,
      }

      log.Printf(\"Server starting on :8080\")
      log.Fatal(httpServer.ListenAndServe())
  }
  ```

### Essential Third-Party Libraries
- **[logrus](https://github.com/sirupsen/logrus)** - Structured logger for Go
  - *Structured Logging*: JSON and structured text output
  - *Levels*: Configurable log levels
  - *Hooks*: Extensible with custom hooks
  ```go
  package main

  import (
      \"github.com/sirupsen/logrus\"
      \"os\"
  )

  var log = logrus.New()

  func init() {
      // Log as JSON for production
      log.SetFormatter(&logrus.JSONFormatter{})

      // Output to stdout instead of default stderr
      log.SetOutput(os.Stdout)

      // Set log level from environment
      level, err := logrus.ParseLevel(os.Getenv(\"LOG_LEVEL\"))
      if err != nil {
          level = logrus.InfoLevel
      }
      log.SetLevel(level)
  }

  func processUser(userID string) error {
      log.WithFields(logrus.Fields{
          \"user_id\": userID,
          \"action\":  \"process_user\",
      }).Info(\"Starting user processing\")

      // Simulate processing
      if userID == \"\" {
          log.WithFields(logrus.Fields{
              \"user_id\": userID,
              \"error\":   \"empty user ID\",
          }).Error(\"Failed to process user\")
          return fmt.Errorf(\"invalid user ID\")
      }

      log.WithFields(logrus.Fields{
          \"user_id\": userID,
          \"action\":  \"process_user\",
      }).Info(\"User processing completed\")

      return nil
  }
  ```

- **[viper](https://github.com/spf13/viper)** - Configuration management
  - *Multiple Formats*: JSON, YAML, TOML, HCL, envfile, Java properties
  - *Environment Variables*: Automatic environment variable binding
  - *Remote Config*: etcd, Consul support
  ```go
  package config

  import (
      \"fmt\"
      \"github.com/spf13/viper\"
  )

  type Config struct {
      Server   ServerConfig   `mapstructure:\"server\"`
      Database DatabaseConfig `mapstructure:\"database\"`
      Redis    RedisConfig    `mapstructure:\"redis\"`
      Log      LogConfig      `mapstructure:\"log\"`
  }

  type ServerConfig struct {
      Port         int    `mapstructure:\"port\"`
      Host         string `mapstructure:\"host\"`
      ReadTimeout  int    `mapstructure:\"read_timeout\"`
      WriteTimeout int    `mapstructure:\"write_timeout\"`
  }

  type DatabaseConfig struct {
      Host         string `mapstructure:\"host\"`
      Port         int    `mapstructure:\"port\"`
      Username     string `mapstructure:\"username\"`
      Password     string `mapstructure:\"password\"`
      Database     string `mapstructure:\"database\"`
      MaxOpenConns int    `mapstructure:\"max_open_conns\"`
      MaxIdleConns int    `mapstructure:\"max_idle_conns\"`
  }

  type RedisConfig struct {
      Host     string `mapstructure:\"host\"`
      Port     int    `mapstructure:\"port\"`
      Password string `mapstructure:\"password\"`
      DB       int    `mapstructure:\"db\"`
  }

  type LogConfig struct {
      Level  string `mapstructure:\"level\"`
      Format string `mapstructure:\"format\"`
  }

  func LoadConfig(path string) (*Config, error) {
      viper.SetConfigName(\"config\")
      viper.SetConfigType(\"yaml\")
      viper.AddConfigPath(path)
      viper.AddConfigPath(\".\")

      // Environment variable support
      viper.AutomaticEnv()
      viper.SetEnvPrefix(\"APP\")

      // Default values
      viper.SetDefault(\"server.port\", 8080)
      viper.SetDefault(\"server.host\", \"localhost\")
      viper.SetDefault(\"log.level\", \"info\")
      viper.SetDefault(\"log.format\", \"json\")

      if err := viper.ReadInConfig(); err != nil {
          return nil, fmt.Errorf(\"failed to read config file: %w\", err)
      }

      var config Config
      if err := viper.Unmarshal(&config); err != nil {
          return nil, fmt.Errorf(\"failed to unmarshal config: %w\", err)
      }

      return &config, nil
  }
  ```

## Web Frameworks and HTTP

### High-Performance Web Frameworks
- **[Gin](https://github.com/gin-gonic/gin)** - HTTP web framework
  - *Performance*: 40x faster than Martini
  - *Middleware*: Rich middleware ecosystem
  - *JSON Validation*: Built-in JSON binding and validation
  ```go
  package main

  import (
      \"net/http\"
      \"strconv\"
      \"time\"

      \"github.com/gin-gonic/gin\"
      \"github.com/gin-gonic/gin/binding\"
  )

  type User struct {
      ID        uint      `json:\"id\" gorm:\"primaryKey\"`
      Name      string    `json:\"name\" binding:\"required,min=2,max=100\"`
      Email     string    `json:\"email\" binding:\"required,email\"`
      Age       int       `json:\"age\" binding:\"gte=0,lte=150\"`
      CreatedAt time.Time `json:\"created_at\"`
  }

  type UserService struct {
      users []User
  }

  func NewUserService() *UserService {
      return &UserService{
          users: make([]User, 0),
      }
  }

  func (s *UserService) CreateUser(user *User) *User {
      user.ID = uint(len(s.users) + 1)
      user.CreatedAt = time.Now()
      s.users = append(s.users, *user)
      return user
  }

  func (s *UserService) GetUser(id uint) (*User, bool) {
      for _, user := range s.users {
          if user.ID == id {
              return &user, true
          }
      }
      return nil, false
  }

  func (s *UserService) GetUsers() []User {
      return s.users
  }

  // Middleware for request logging
  func RequestLogger() gin.HandlerFunc {
      return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
          return fmt.Sprintf(\"%s - [%s] \\\"%s %s %s %d %s \\\"%s\\\" %s\\\"\\\n\",
              param.ClientIP,
              param.TimeStamp.Format(time.RFC1123),
              param.Method,
              param.Path,
              param.Request.Proto,
              param.StatusCode,
              param.Latency,
              param.Request.UserAgent(),
              param.ErrorMessage,
          )
      })
  }

  // Middleware for CORS
  func CORSMiddleware() gin.HandlerFunc {
      return func(c *gin.Context) {
          c.Writer.Header().Set(\"Access-Control-Allow-Origin\", \"*\")
          c.Writer.Header().Set(\"Access-Control-Allow-Credentials\", \"true\")
          c.Writer.Header().Set(\"Access-Control-Allow-Headers\", \"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With\")
          c.Writer.Header().Set(\"Access-Control-Allow-Methods\", \"POST, OPTIONS, GET, PUT, DELETE\")

          if c.Request.Method == \"OPTIONS\" {
              c.AbortWithStatus(204)
              return
          }

          c.Next()
      }
  }

  func setupRouter(userService *UserService) *gin.Engine {
      r := gin.New()

      // Middleware
      r.Use(RequestLogger())
      r.Use(gin.Recovery())
      r.Use(CORSMiddleware())

      // Health check endpoint
      r.GET(\"/health\", func(c *gin.Context) {
          c.JSON(http.StatusOK, gin.H{
              \"status\":    \"healthy\",
              \"timestamp\": time.Now().Unix(),
          })
      })

      // API routes
      api := r.Group(\"/api/v1\")
      {
          users := api.Group(\"/users\")
          {
              users.GET(\"\", func(c *gin.Context) {
                  users := userService.GetUsers()
                  c.JSON(http.StatusOK, gin.H{\"users\": users})
              })

              users.POST(\"\", func(c *gin.Context) {
                  var user User
                  if err := c.ShouldBindJSON(&user); err != nil {
                      c.JSON(http.StatusBadRequest, gin.H{\"error\": err.Error()})
                      return
                  }

                  createdUser := userService.CreateUser(&user)
                  c.JSON(http.StatusCreated, createdUser)
              })

              users.GET(\"/:id\", func(c *gin.Context) {
                  idStr := c.Param(\"id\")
                  id, err := strconv.ParseUint(idStr, 10, 32)
                  if err != nil {
                      c.JSON(http.StatusBadRequest, gin.H{\"error\": \"Invalid user ID\"})
                      return
                  }

                  user, exists := userService.GetUser(uint(id))
                  if !exists {
                      c.JSON(http.StatusNotFound, gin.H{\"error\": \"User not found\"})
                      return
                  }

                  c.JSON(http.StatusOK, user)
              })
          }
      }

      return r
  }

  func main() {
      userService := NewUserService()
      router := setupRouter(userService)

      server := &http.Server{
          Addr:         \":8080\",
          Handler:      router,
          ReadTimeout:  15 * time.Second,
          WriteTimeout: 15 * time.Second,
      }

      log.Printf(\"Server starting on :8080\")
      log.Fatal(server.ListenAndServe())
  }
  ```

- **[Fiber](https://github.com/gofiber/fiber)** - Express.js inspired web framework
  - *Fast*: Built on Fasthttp, up to 10x faster than net/http
  - *Zero Memory Allocation*: Optimized for performance
  - *Express-like API*: Familiar API for Node.js developers

- **[Echo](https://github.com/labstack/echo)** - High performance, minimalist web framework
  - *HTTP/2*: HTTP/2 support
  - *Middleware*: Extensive middleware collection
  - *Data Binding*: JSON, XML, form data binding

## Database and Storage

### SQL Databases
- **[GORM](https://gorm.io/)** - The fantastic ORM library for Golang
  - *Auto Migration*: Database schema migration
  - *Associations*: Has One, Has Many, Many To Many, Polymorphism
  - *Hooks*: Before/After Create/Save/Update/Delete/Find
  ```go
  package models

  import (
      \"time\"
      \"gorm.io/gorm\"
      \"gorm.io/driver/postgres\"
  )

  type User struct {
      ID        uint           `gorm:\"primaryKey\" json:\"id\"`
      Name      string         `gorm:\"size:100;not null\" json:\"name\"`
      Email     string         `gorm:\"uniqueIndex;size:255;not null\" json:\"email\"`
      Age       int            `gorm:\"check:age >= 0\" json:\"age\"`
      Posts     []Post         `json:\"posts,omitempty\"`
      Profile   Profile        `json:\"profile,omitempty\"`
      CreatedAt time.Time      `json:\"created_at\"`
      UpdatedAt time.Time      `json:\"updated_at\"`
      DeletedAt gorm.DeletedAt `gorm:\"index\" json:\"-\"`
  }

  type Post struct {
      ID        uint           `gorm:\"primaryKey\" json:\"id\"`
      Title     string         `gorm:\"size:255;not null\" json:\"title\"`
      Content   string         `gorm:\"type:text\" json:\"content\"`
      UserID    uint           `gorm:\"not null\" json:\"user_id\"`
      User      User           `json:\"user,omitempty\"`
      Tags      []Tag          `gorm:\"many2many:post_tags;\" json:\"tags,omitempty\"`
      CreatedAt time.Time      `json:\"created_at\"`
      UpdatedAt time.Time      `json:\"updated_at\"`
      DeletedAt gorm.DeletedAt `gorm:\"index\" json:\"-\"`
  }

  type Profile struct {
      ID       uint   `gorm:\"primaryKey\" json:\"id\"`
      Bio      string `gorm:\"type:text\" json:\"bio\"`
      Avatar   string `gorm:\"size:255\" json:\"avatar\"`
      UserID   uint   `gorm:\"uniqueIndex\" json:\"user_id\"`
  }

  type Tag struct {
      ID    uint   `gorm:\"primaryKey\" json:\"id\"`
      Name  string `gorm:\"uniqueIndex;size:50;not null\" json:\"name\"`
      Posts []Post `gorm:\"many2many:post_tags;\" json:\"posts,omitempty\"`
  }

  type Database struct {
      DB *gorm.DB
  }

  func NewDatabase(dsn string) (*Database, error) {
      db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
          // Configure connection pool
          //ConnPool: &sql.DB{},
      })
      if err != nil {
          return nil, err
      }

      // Configure connection pool
      sqlDB, err := db.DB()
      if err != nil {
          return nil, err
      }

      sqlDB.SetMaxIdleConns(10)
      sqlDB.SetMaxOpenConns(100)
      sqlDB.SetConnMaxLifetime(time.Hour)

      // Auto migrate tables
      if err := db.AutoMigrate(&User{}, &Post{}, &Profile{}, &Tag{}); err != nil {
          return nil, err
      }

      return &Database{DB: db}, nil
  }

  // User repository methods
  func (d *Database) CreateUser(user *User) error {
      return d.DB.Create(user).Error
  }

  func (d *Database) GetUserByID(id uint) (*User, error) {
      var user User
      err := d.DB.Preload(\"Posts\").Preload(\"Profile\").First(&user, id).Error
      if err != nil {
          return nil, err
      }
      return &user, nil
  }

  func (d *Database) GetUserByEmail(email string) (*User, error) {
      var user User
      err := d.DB.Where(\"email = ?\", email).First(&user).Error
      if err != nil {
          return nil, err
      }
      return &user, nil
  }

  func (d *Database) UpdateUser(user *User) error {
      return d.DB.Save(user).Error
  }

  func (d *Database) DeleteUser(id uint) error {
      return d.DB.Delete(&User{}, id).Error
  }

  func (d *Database) GetUsersWithPagination(page, pageSize int) ([]User, int64, error) {
      var users []User
      var total int64

      // Count total users
      d.DB.Model(&User{}).Count(&total)

      // Get users with pagination
      offset := (page - 1) * pageSize
      err := d.DB.Offset(offset).Limit(pageSize).Find(&users).Error
      if err != nil {
          return nil, 0, err
      }

      return users, total, nil
  }
  ```

- **[sqlx](https://github.com/jmoiron/sqlx)** - General purpose extensions to database/sql
  - *Structured*: Scan to structs, named parameters
  - *Performance*: Minimal overhead over database/sql
  - *Flexibility*: Use raw SQL when needed

### NoSQL and Key-Value Stores
- **[go-redis](https://github.com/go-redis/redis)** - Redis client for Go
  - *Connection Pooling*: Automatic connection management
  - *Pipelining*: Batch commands for performance
  - *Pub/Sub*: Real-time messaging
  ```go
  package cache

  import (
      \"context\"
      \"encoding/json\"
      \"fmt\"
      \"time\"

      \"github.com/go-redis/redis/v8\"
  )

  type RedisCache struct {
      client *redis.Client
      ctx    context.Context
  }

  func NewRedisCache(addr, password string, db int) *RedisCache {
      rdb := redis.NewClient(&redis.Options{
          Addr:         addr,
          Password:     password,
          DB:           db,
          PoolSize:     10,
          MinIdleConns: 5,
          DialTimeout:  time.Second * 5,
          ReadTimeout:  time.Second * 3,
          WriteTimeout: time.Second * 3,
      })

      return &RedisCache{
          client: rdb,
          ctx:    context.Background(),
      }
  }

  func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
      jsonData, err := json.Marshal(value)
      if err != nil {
          return fmt.Errorf(\"failed to marshal value: %w\", err)
      }

      return r.client.Set(r.ctx, key, jsonData, expiration).Err()
  }

  func (r *RedisCache) Get(key string, dest interface{}) error {
      val, err := r.client.Get(r.ctx, key).Result()
      if err != nil {
          return err
      }

      return json.Unmarshal([]byte(val), dest)
  }

  func (r *RedisCache) Delete(key string) error {
      return r.client.Del(r.ctx, key).Err()
  }

  func (r *RedisCache) Exists(key string) (bool, error) {
      result, err := r.client.Exists(r.ctx, key).Result()
      return result > 0, err
  }

  // Advanced operations
  func (r *RedisCache) IncrementCounter(key string, expiration time.Duration) (int64, error) {
      pipe := r.client.TxPipeline()
      incrCmd := pipe.Incr(r.ctx, key)
      pipe.Expire(r.ctx, key, expiration)

      _, err := pipe.Exec(r.ctx)
      if err != nil {
          return 0, err
      }

      return incrCmd.Val(), nil
  }

  func (r *RedisCache) AddToSet(key string, members ...interface{}) error {
      return r.client.SAdd(r.ctx, key, members...).Err()
  }

  func (r *RedisCache) GetSetMembers(key string) ([]string, error) {
      return r.client.SMembers(r.ctx, key).Result()
  }

  // Pub/Sub functionality
  func (r *RedisCache) Publish(channel string, message interface{}) error {
      jsonData, err := json.Marshal(message)
      if err != nil {
          return fmt.Errorf(\"failed to marshal message: %w\", err)
      }

      return r.client.Publish(r.ctx, channel, jsonData).Err()
  }

  func (r *RedisCache) Subscribe(channel string) *redis.PubSub {
      return r.client.Subscribe(r.ctx, channel)
  }

  func (r *RedisCache) Close() error {
      return r.client.Close()
  }

  func (r *RedisCache) Ping() error {
      return r.client.Ping(r.ctx).Err()
  }
  ```

- **[mongo-driver](https://github.com/mongodb/mongo-go-driver)** - MongoDB driver
  - *Official*: Official MongoDB driver for Go
  - *Performance*: Optimized for high throughput
  - *Features*: Aggregation pipeline, change streams, transactions

## Microservices and Distributed Systems

### Service Communication
- **[gRPC-Go](https://github.com/grpc/grpc-go)** - gRPC implementation in Go
  - *Performance*: High-performance RPC framework
  - *Code Generation*: Protocol buffer code generation
  - *Streaming*: Bidirectional streaming support
  ```go
  // user.proto
  syntax = \"proto3\";

  package user;
  option go_package = \"./proto\";

  service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
    rpc ListUsers(ListUsersRequest) returns (stream User);
  }

  message User {
    int32 id = 1;
    string name = 2;
    string email = 3;
    int64 created_at = 4;
  }

  message GetUserRequest {
    int32 id = 1;
  }

  message GetUserResponse {
    User user = 1;
  }

  // server implementation
  package main

  import (
      \"context\"
      \"fmt\"
      \"log\"
      \"net\"

      \"google.golang.org/grpc\"
      pb \"./proto\"
  )

  type UserServer struct {
      pb.UnimplementedUserServiceServer
      users map[int32]*pb.User
  }

  func NewUserServer() *UserServer {
      return &UserServer{
          users: make(map[int32]*pb.User),
      }
  }

  func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
      user, exists := s.users[req.Id]
      if !exists {
          return nil, fmt.Errorf(\"user not found\")
      }

      return &pb.GetUserResponse{User: user}, nil
  }

  func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
      user := &pb.User{
          Id:        int32(len(s.users) + 1),
          Name:      req.Name,
          Email:     req.Email,
          CreatedAt: time.Now().Unix(),
      }

      s.users[user.Id] = user

      return &pb.CreateUserResponse{User: user}, nil
  }

  func (s *UserServer) ListUsers(req *pb.ListUsersRequest, stream pb.UserService_ListUsersServer) error {
      for _, user := range s.users {
          if err := stream.Send(user); err != nil {
              return err
          }
      }
      return nil
  }

  func main() {
      lis, err := net.Listen(\"tcp\", \":50051\")
      if err != nil {
          log.Fatalf(\"Failed to listen: %v\", err)
      }

      server := grpc.NewServer()
      userServer := NewUserServer()

      pb.RegisterUserServiceServer(server, userServer)

      log.Printf(\"gRPC server listening on :50051\")
      if err := server.Serve(lis); err != nil {
          log.Fatalf(\"Failed to serve: %v\", err)
      }
  }
  ```

### Message Queues
- **[Kafka-go](https://github.com/segmentio/kafka-go)** - Kafka library for Go
  - *Performance*: High-throughput message streaming
  - *Reliability*: At-least-once delivery guarantees
  - *Scalability*: Distributed, partitioned logs
  ```go
  package kafka

  import (
      \"context\"
      \"encoding/json\"
      \"fmt\"
      \"log\"
      \"time\"

      \"github.com/segmentio/kafka-go\"
  )

  type Producer struct {
      writer *kafka.Writer
  }

  type Consumer struct {
      reader *kafka.Reader
  }

  type Message struct {
      ID        string                 `json:\"id\"`
      Type      string                 `json:\"type\"`
      Timestamp time.Time              `json:\"timestamp\"`
      Data      map[string]interface{} `json:\"data\"`
  }

  func NewProducer(brokers []string, topic string) *Producer {
      writer := &kafka.Writer{
          Addr:         kafka.TCP(brokers...),
          Topic:        topic,
          Balancer:     &kafka.LeastBytes{},
          RequiredAcks: kafka.RequireAll,
          Async:        false,
      }

      return &Producer{writer: writer}
  }

  func (p *Producer) PublishMessage(ctx context.Context, key string, message Message) error {
      messageBytes, err := json.Marshal(message)
      if err != nil {
          return fmt.Errorf(\"failed to marshal message: %w\", err)
      }

      kafkaMessage := kafka.Message{
          Key:   []byte(key),
          Value: messageBytes,
          Time:  time.Now(),
      }

      return p.writer.WriteMessages(ctx, kafkaMessage)
  }

  func (p *Producer) Close() error {
      return p.writer.Close()
  }

  func NewConsumer(brokers []string, topic, groupID string) *Consumer {
      reader := kafka.NewReader(kafka.ReaderConfig{
          Brokers:  brokers,
          Topic:    topic,
          GroupID:  groupID,
          MinBytes: 10e3, // 10KB
          MaxBytes: 10e6, // 10MB
      })

      return &Consumer{reader: reader}
  }

  func (c *Consumer) ConsumeMessages(ctx context.Context, handler func(Message) error) error {
      for {
          select {
          case <-ctx.Done():
              return ctx.Err()
          default:
              kafkaMessage, err := c.reader.ReadMessage(ctx)
              if err != nil {
                  log.Printf(\"Error reading message: %v\", err)
                  continue
              }

              var message Message
              if err := json.Unmarshal(kafkaMessage.Value, &message); err != nil {
                  log.Printf(\"Error unmarshaling message: %v\", err)
                  continue
              }

              if err := handler(message); err != nil {
                  log.Printf(\"Error handling message: %v\", err)
                  continue
              }

              log.Printf(\"Processed message: %s\", message.ID)
          }
      }
  }

  func (c *Consumer) Close() error {
      return c.reader.Close()
  }

  // Usage example
  func main() {
      ctx := context.Background()
      brokers := []string{\"localhost:9092\"}

      // Producer
      producer := NewProducer(brokers, \"user-events\")
      defer producer.Close()

      message := Message{
          ID:        \"msg-001\",
          Type:      \"user.created\",
          Timestamp: time.Now(),
          Data: map[string]interface{}{
              \"user_id\": \"12345\",
              \"name\":    \"John Doe\",
              \"email\":   \"john@example.com\",
          },
      }

      if err := producer.PublishMessage(ctx, \"user-12345\", message); err != nil {
          log.Printf(\"Failed to publish message: %v\", err)
      }

      // Consumer
      consumer := NewConsumer(brokers, \"user-events\", \"user-service\")
      defer consumer.Close()

      handler := func(msg Message) error {
          log.Printf(\"Processing message: %s, Type: %s\", msg.ID, msg.Type)
          // Process message here
          return nil
      }

      if err := consumer.ConsumeMessages(ctx, handler); err != nil {
          log.Printf(\"Consumer error: %v\", err)
      }
  }
  ```

## Cloud-Native Development

### Kubernetes Integration
- **[client-go](https://github.com/kubernetes/client-go)** - Kubernetes client library
  - *Official*: Official Kubernetes client for Go
  - *Complete API*: Full Kubernetes API access
  - *Custom Resources*: Custom resource definitions support
  ```go
  package k8s

  import (
      \"context\"
      \"fmt\"
      \"time\"

      appsv1 \"k8s.io/api/apps/v1\"
      corev1 \"k8s.io/api/core/v1\"
      metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"
      \"k8s.io/client-go/kubernetes\"
      \"k8s.io/client-go/rest\"
      \"k8s.io/client-go/tools/clientcmd\"
  )

  type K8sClient struct {
      clientset *kubernetes.Clientset
      config    *rest.Config
  }

  func NewK8sClient(kubeconfig string) (*K8sClient, error) {
      var config *rest.Config
      var err error

      if kubeconfig != \"\" {
          config, err = clientcmd.BuildConfigFromFlags(\"\", kubeconfig)
      } else {
          config, err = rest.InClusterConfig()
      }

      if err != nil {
          return nil, fmt.Errorf(\"failed to create config: %w\", err)
      }

      clientset, err := kubernetes.NewForConfig(config)
      if err != nil {
          return nil, fmt.Errorf(\"failed to create clientset: %w\", err)
      }

      return &K8sClient{
          clientset: clientset,
          config:    config,
      }, nil
  }

  func (k *K8sClient) CreateDeployment(ctx context.Context, namespace, name, image string, replicas int32) error {
      deployment := &appsv1.Deployment{
          ObjectMeta: metav1.ObjectMeta{
              Name:      name,
              Namespace: namespace,
          },
          Spec: appsv1.DeploymentSpec{
              Replicas: &replicas,
              Selector: &metav1.LabelSelector{
                  MatchLabels: map[string]string{
                      \"app\": name,
                  },
              },
              Template: corev1.PodTemplateSpec{
                  ObjectMeta: metav1.ObjectMeta{
                      Labels: map[string]string{
                          \"app\": name,
                      },
                  },
                  Spec: corev1.PodSpec{
                      Containers: []corev1.Container{
                          {
                              Name:  name,
                              Image: image,
                              Ports: []corev1.ContainerPort{
                                  {
                                      ContainerPort: 8080,
                                  },
                              },
                              Resources: corev1.ResourceRequirements{
                                  Requests: corev1.ResourceList{
                                      corev1.ResourceCPU:    resource.MustParse(\"100m\"),
                                      corev1.ResourceMemory: resource.MustParse(\"128Mi\"),
                                  },
                                  Limits: corev1.ResourceList{
                                      corev1.ResourceCPU:    resource.MustParse(\"500m\"),
                                      corev1.ResourceMemory: resource.MustParse(\"512Mi\"),
                                  },
                              },
                              LivenessProbe: &corev1.Probe{
                                  ProbeHandler: corev1.ProbeHandler{
                                      HTTPGet: &corev1.HTTPGetAction{
                                          Path: \"/health\",
                                          Port: intstr.FromInt(8080),
                                      },
                                  },
                                  InitialDelaySeconds: 30,
                                  PeriodSeconds:       10,
                              },
                              ReadinessProbe: &corev1.Probe{
                                  ProbeHandler: corev1.ProbeHandler{
                                      HTTPGet: &corev1.HTTPGetAction{
                                          Path: \"/ready\",
                                          Port: intstr.FromInt(8080),
                                      },
                                  },
                                  InitialDelaySeconds: 5,
                                  PeriodSeconds:       5,
                              },
                          },
                      },
                  },
              },
          },
      }

      _, err := k.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
      return err
  }

  func (k *K8sClient) GetPods(ctx context.Context, namespace, labelSelector string) (*corev1.PodList, error) {
      return k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
          LabelSelector: labelSelector,
      })
  }

  func (k *K8sClient) WaitForDeploymentReady(ctx context.Context, namespace, name string, timeout time.Duration) error {
      ctx, cancel := context.WithTimeout(ctx, timeout)
      defer cancel()

      for {
          select {
          case <-ctx.Done():
              return fmt.Errorf(\"timeout waiting for deployment %s to be ready\", name)
          default:
              deployment, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
              if err != nil {
                  return err
              }

              if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
                  return nil
              }

              time.Sleep(2 * time.Second)
          }
      }
  }
  ```

### Service Mesh and Observability
- **[Prometheus](https://github.com/prometheus/client_golang)** - Metrics collection
  - *Metrics*: Counter, Gauge, Histogram, Summary
  - *Labels*: Multi-dimensional metrics
  - *Integration*: Easy HTTP handler integration

## Concurrency and Goroutines

### Goroutine Patterns
- **Worker Pools** - Efficient concurrent processing
  ```go
  package workers

  import (
      \"context\"
      \"fmt\"
      \"sync\"
      \"time\"
  )

  type Job struct {
      ID   int
      Data interface{}
  }

  type Result struct {
      Job    Job
      Result interface{}
      Error  error
  }

  type WorkerPool struct {
      numWorkers int
      jobsChan   chan Job
      resultChan chan Result
      ctx        context.Context
      cancel     context.CancelFunc
      wg         sync.WaitGroup
  }

  func NewWorkerPool(numWorkers, bufferSize int) *WorkerPool {
      ctx, cancel := context.WithCancel(context.Background())

      return &WorkerPool{
          numWorkers: numWorkers,
          jobsChan:   make(chan Job, bufferSize),
          resultChan: make(chan Result, bufferSize),
          ctx:        ctx,
          cancel:     cancel,
      }
  }

  func (wp *WorkerPool) Start(processor func(Job) (interface{}, error)) {
      // Start workers
      for i := 0; i < wp.numWorkers; i++ {
          wp.wg.Add(1)
          go wp.worker(i, processor)
      }

      // Start result collector
      go wp.resultCollector()
  }

  func (wp *WorkerPool) worker(id int, processor func(Job) (interface{}, error)) {
      defer wp.wg.Done()

      fmt.Printf(\"Worker %d started\\n\", id)

      for {
          select {
          case job := <-wp.jobsChan:
              fmt.Printf(\"Worker %d processing job %d\\n\", id, job.ID)

              result, err := processor(job)

              select {
              case wp.resultChan <- Result{Job: job, Result: result, Error: err}:
              case <-wp.ctx.Done():
                  return
              }

          case <-wp.ctx.Done():
              fmt.Printf(\"Worker %d stopping\\n\", id)
              return
          }
      }
  }

  func (wp *WorkerPool) resultCollector() {
      for {
          select {
          case result := <-wp.resultChan:
              if result.Error != nil {
                  fmt.Printf(\"Job %d failed: %v\\n\", result.Job.ID, result.Error)
              } else {
                  fmt.Printf(\"Job %d completed: %v\\n\", result.Job.ID, result.Result)
              }
          case <-wp.ctx.Done():
              return
          }
      }
  }

  func (wp *WorkerPool) SubmitJob(job Job) {
      select {
      case wp.jobsChan <- job:
      case <-wp.ctx.Done():
      }
  }

  func (wp *WorkerPool) Stop() {
      close(wp.jobsChan)
      wp.wg.Wait()
      wp.cancel()
      close(wp.resultChan)
  }

  // Rate limiter using goroutines
  type RateLimiter struct {
      rate     time.Duration
      tokens   chan struct{}
      ctx      context.Context
      cancel   context.CancelFunc
  }

  func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
      ctx, cancel := context.WithCancel(context.Background())

      rl := &RateLimiter{
          rate:   rate,
          tokens: make(chan struct{}, burst),
          ctx:    ctx,
          cancel: cancel,
      }

      // Fill initial tokens
      for i := 0; i < burst; i++ {
          rl.tokens <- struct{}{}
      }

      // Start token refiller
      go rl.refillTokens()

      return rl
  }

  func (rl *RateLimiter) refillTokens() {
      ticker := time.NewTicker(rl.rate)
      defer ticker.Stop()

      for {
          select {
          case <-ticker.C:
              select {
              case rl.tokens <- struct{}{}:
              default:
                  // Token bucket is full
              }
          case <-rl.ctx.Done():
              return
          }
      }
  }

  func (rl *RateLimiter) Allow() bool {
      select {
      case <-rl.tokens:
          return true
      default:
          return false
      }
  }

  func (rl *RateLimiter) Wait(ctx context.Context) error {
      select {
      case <-rl.tokens:
          return nil
      case <-ctx.Done():
          return ctx.Err()
      case <-rl.ctx.Done():
          return rl.ctx.Err()
      }
  }

  func (rl *RateLimiter) Stop() {
      rl.cancel()
  }

  // Usage example
  func main() {
      // Worker pool example
      pool := NewWorkerPool(3, 10)

      processor := func(job Job) (interface{}, error) {
          // Simulate work
          time.Sleep(time.Second)
          return fmt.Sprintf(\"Processed: %v\", job.Data), nil
      }

      pool.Start(processor)

      // Submit jobs
      for i := 0; i < 10; i++ {
          pool.SubmitJob(Job{
              ID:   i,
              Data: fmt.Sprintf(\"Job %d\", i),
          })
      }

      time.Sleep(15 * time.Second)
      pool.Stop()

      // Rate limiter example
      limiter := NewRateLimiter(100*time.Millisecond, 5)
      defer limiter.Stop()

      for i := 0; i < 10; i++ {
          if limiter.Allow() {
              fmt.Printf(\"Request %d allowed\\n\", i)
          } else {
              fmt.Printf(\"Request %d rate limited\\n\", i)
          }
      }
  }
  ```

## CLI and Terminal Applications

### Command-Line Applications
- **[cobra](https://github.com/spf13/cobra)** - CLI library
  - *Commands*: Nested command structure
  - *Flags*: Persistent and local flags
  - *Completion*: Shell completion generation
  ```go
  package main

  import (
      \"fmt\"
      \"os\"

      \"github.com/spf13/cobra\"
      \"github.com/spf13/viper\"
  )

  var (
      cfgFile     string
      verbose     bool
      outputFormat string
  )

  // rootCmd represents the base command when called without any subcommands
  var rootCmd = &cobra.Command{
      Use:   \"myapp\",
      Short: \"A brief description of your application\",
      Long: `A longer description that spans multiple lines and likely contains
  examples and usage of using your application.`,
  }

  var versionCmd = &cobra.Command{
      Use:   \"version\",
      Short: \"Print the version number\",
      Long:  `All software has versions. This is myapp's`,
      Run: func(cmd *cobra.Command, args []string) {
          fmt.Println(\"myapp v1.0.0\")
      },
  }

  var userCmd = &cobra.Command{
      Use:   \"user\",
      Short: \"User management commands\",
      Long:  `Commands for managing users in the system`,
  }

  var userListCmd = &cobra.Command{
      Use:   \"list\",
      Short: \"List all users\",
      Long:  `List all users in the system with optional filtering`,
      RunE: func(cmd *cobra.Command, args []string) error {
          if verbose {
              fmt.Println(\"Fetching users with verbose output...\")
          }

          users := []struct {
              ID    int    `json:\"id\"`
              Name  string `json:\"name\"`
              Email string `json:\"email\"`
          }{
              {1, \"John Doe\", \"john@example.com\"},
              {2, \"Jane Smith\", \"jane@example.com\"},
          }

          switch outputFormat {
          case \"json\":
              return outputJSON(users)
          case \"table\":
              return outputTable(users)
          default:
              return fmt.Errorf(\"unsupported output format: %s\", outputFormat)
          }
      },
  }

  var userCreateCmd = &cobra.Command{
      Use:   \"create [name] [email]\",
      Short: \"Create a new user\",
      Long:  `Create a new user with the specified name and email`,
      Args:  cobra.ExactArgs(2),
      RunE: func(cmd *cobra.Command, args []string) error {
          name := args[0]
          email := args[1]

          if verbose {
              fmt.Printf(\"Creating user: %s (%s)\\n\", name, email)
          }

          user := struct {
              ID    int    `json:\"id\"`
              Name  string `json:\"name\"`
              Email string `json:\"email\"`
          }{
              ID:    3,
              Name:  name,
              Email: email,
          }

          fmt.Printf(\"User created successfully: %+v\\n\", user)
          return nil
      },
  }

  func init() {
      cobra.OnInitialize(initConfig)

      // Global flags
      rootCmd.PersistentFlags().StringVar(&cfgFile, \"config\", \"\", \"config file (default is $HOME/.myapp.yaml)\")
      rootCmd.PersistentFlags().BoolVarP(&verbose, \"verbose\", \"v\", false, \"verbose output\")

      // User list command flags
      userListCmd.Flags().StringVarP(&outputFormat, \"output\", \"o\", \"table\", \"output format (json|table)\")

      // Add commands
      rootCmd.AddCommand(versionCmd)
      rootCmd.AddCommand(userCmd)
      userCmd.AddCommand(userListCmd)
      userCmd.AddCommand(userCreateCmd)
  }

  func initConfig() {
      if cfgFile != \"\" {
          viper.SetConfigFile(cfgFile)
      } else {
          home, err := os.UserHomeDir()
          cobra.CheckErr(err)

          viper.AddConfigPath(home)
          viper.SetConfigType(\"yaml\")
          viper.SetConfigName(\".myapp\")
      }

      viper.AutomaticEnv()

      if err := viper.ReadInConfig(); err == nil {
          if verbose {
              fmt.Println(\"Using config file:\", viper.ConfigFileUsed())
          }
      }
  }

  func outputJSON(data interface{}) error {
      encoder := json.NewEncoder(os.Stdout)
      encoder.SetIndent(\"\", \"  \")
      return encoder.Encode(data)
  }

  func outputTable(users []struct {
      ID    int    `json:\"id\"`
      Name  string `json:\"name\"`
      Email string `json:\"email\"`
  }) error {
      fmt.Printf(\"%-5s %-20s %-30s\\n\", \"ID\", \"NAME\", \"EMAIL\")
      fmt.Printf(\"%-5s %-20s %-30s\\n\", \"--\", \"----\", \"-----\")
      for _, user := range users {
          fmt.Printf(\"%-5d %-20s %-30s\\n\", user.ID, user.Name, user.Email)
      }
      return nil
  }

  func main() {
      if err := rootCmd.Execute(); err != nil {
          fmt.Fprintln(os.Stderr, err)
          os.Exit(1)
      }
  }
  ```

## Testing and Quality Assurance

### Testing Framework
- **Built-in Testing** - Go's built-in testing framework
  ```go
  package user

  import (
      \"testing\"
      \"time\"
      \"reflect\"
  )

  func TestUserValidation(t *testing.T) {
      tests := []struct {
          name    string
          user    User
          wantErr bool
      }{
          {
              name: \"valid user\",
              user: User{
                  Name:  \"John Doe\",
                  Email: \"john@example.com\",
                  Age:   30,
              },
              wantErr: false,
          },
          {
              name: \"invalid email\",
              user: User{
                  Name:  \"John Doe\",
                  Email: \"invalid-email\",
                  Age:   30,
              },
              wantErr: true,
          },
          {
              name: \"empty name\",
              user: User{
                  Name:  \"\",
                  Email: \"john@example.com\",
                  Age:   30,
              },
              wantErr: true,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              err := tt.user.Validate()
              if (err != nil) != tt.wantErr {
                  t.Errorf(\"User.Validate() error = %v, wantErr %v\", err, tt.wantErr)
              }
          })
      }
  }

  // Benchmark test
  func BenchmarkUserValidation(b *testing.B) {
      user := User{
          Name:  \"John Doe\",
          Email: \"john@example.com\",
          Age:   30,
      }

      b.ResetTimer()
      for i := 0; i < b.N; i++ {
          user.Validate()
      }
  }

  // Example test
  func ExampleUser_String() {
      user := User{
          ID:    1,
          Name:  \"John Doe\",
          Email: \"john@example.com\",
          Age:   30,
      }
      fmt.Println(user.String())
      // Output: User{ID: 1, Name: John Doe, Email: john@example.com, Age: 30}
  }
  ```

### Mocking and Test Doubles
- **[testify](https://github.com/stretchr/testify)** - Testing toolkit
  ```go
  package service

  import (
      \"testing\"
      \"github.com/stretchr/testify/assert\"
      \"github.com/stretchr/testify/mock\"
      \"github.com/stretchr/testify/suite\"
  )

  // Mock implementation
  type MockUserRepository struct {
      mock.Mock
  }

  func (m *MockUserRepository) GetUser(id int) (*User, error) {
      args := m.Called(id)
      return args.Get(0).(*User), args.Error(1)
  }

  func (m *MockUserRepository) CreateUser(user *User) error {
      args := m.Called(user)
      return args.Error(0)
  }

  // Test suite
  type UserServiceTestSuite struct {
      suite.Suite
      mockRepo    *MockUserRepository
      userService *UserService
  }

  func (suite *UserServiceTestSuite) SetupTest() {
      suite.mockRepo = new(MockUserRepository)
      suite.userService = NewUserService(suite.mockRepo)
  }

  func (suite *UserServiceTestSuite) TestGetUser_Success() {
      // Arrange
      userID := 1
      expectedUser := &User{
          ID:    1,
          Name:  \"John Doe\",
          Email: \"john@example.com\",
      }

      suite.mockRepo.On(\"GetUser\", userID).Return(expectedUser, nil)

      // Act
      user, err := suite.userService.GetUser(userID)

      // Assert
      assert.NoError(suite.T(), err)
      assert.Equal(suite.T(), expectedUser, user)
      suite.mockRepo.AssertExpectations(suite.T())
  }

  func (suite *UserServiceTestSuite) TestGetUser_NotFound() {
      // Arrange
      userID := 999
      suite.mockRepo.On(\"GetUser\", userID).Return((*User)(nil), ErrUserNotFound)

      // Act
      user, err := suite.userService.GetUser(userID)

      // Assert
      assert.Error(suite.T(), err)
      assert.Equal(suite.T(), ErrUserNotFound, err)
      assert.Nil(suite.T(), user)
      suite.mockRepo.AssertExpectations(suite.T())
  }

  func TestUserServiceTestSuite(t *testing.T) {
      suite.Run(t, new(UserServiceTestSuite))
  }
  ```

## Performance and Profiling

### Built-in Profiling
- **pprof** - Built-in profiling support
  ```go
  package main

  import (
      \"log\"
      \"net/http\"
      _ \"net/http/pprof\"
      \"runtime\"
      \"time\"
  )

  func init() {
      // Enable profiling endpoint
      go func() {
          log.Println(http.ListenAndServe(\"localhost:6060\", nil))
      }()
  }

  func expensiveOperation() {
      // Simulate CPU intensive work
      for i := 0; i < 1000000; i++ {
          _ = i * i
      }
  }

  func memoryIntensiveOperation() {
      // Simulate memory intensive work
      data := make([][]byte, 1000)
      for i := range data {
          data[i] = make([]byte, 1024*1024) // 1MB each
      }

      // Force GC
      runtime.GC()
  }

  func main() {
      // CPU profiling example
      go func() {
          for {
              expensiveOperation()
              time.Sleep(100 * time.Millisecond)
          }
      }()

      // Memory profiling example
      go func() {
          for {
              memoryIntensiveOperation()
              time.Sleep(1 * time.Second)
          }
      }()

      // Keep the main goroutine alive
      select {}
  }

  // Custom profiling
  func profileFunction() {
      defer func(start time.Time) {
          duration := time.Since(start)
          log.Printf(\"Function took: %v\", duration)
      }(time.Now())

      // Function implementation
      expensiveOperation()
  }
  ```

## Learning Resources

### Official Documentation
- **[Go Documentation](https://golang.org/doc/)** - Official Go documentation
- **[Go by Example](https://gobyexample.com/)** - Learn Go by examples
- **[Effective Go](https://golang.org/doc/effective_go.html)** - Go best practices
- **[Go Tour](https://tour.golang.org/)** - Interactive Go tutorial

### Advanced Topics
- **[Go Memory Model](https://golang.org/ref/mem)** - Memory consistency guarantees
- **[Go Race Detector](https://golang.org/doc/articles/race_detector.html)** - Detecting race conditions
- **[Go Modules](https://golang.org/ref/mod)** - Dependency management
- **[Go Assembly](https://golang.org/doc/asm)** - Go assembler guide

### Books and Resources
- **\"The Go Programming Language\" by Donovan & Kernighan** - Comprehensive Go guide
- **\"Go in Action\" by Kennedy, Ketelsen & St. Martin** - Practical Go development
- **\"Concurrency in Go\" by Katherine Cox-Buday** - Concurrency patterns
- **\"Go Web Programming\" by Sau Sheong Chang** - Web development with Go

### Community
- **[r/golang](https://reddit.com/r/golang)** - Active Go community
- **[Go Forum](https://forum.golangbridge.org/)** - Community discussion forum
- **[GopherCon](https://gophercon.com/)** - Premier Go conference
- **[Go Weekly](https://golangweekly.com/)** - Weekly Go newsletter

## Production Deployment Patterns

### Configuration and Environment
```go
package config

import (
    \"fmt\"
    \"os\"
    \"strconv\"
    \"time\"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Log      LogConfig
}

type ServerConfig struct {
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

func LoadConfig() (*Config, error) {
    config := &Config{}

    // Server configuration
    port, err := strconv.Atoi(getEnv(\"PORT\", \"8080\"))
    if err != nil {
        return nil, fmt.Errorf(\"invalid PORT: %w\", err)
    }

    config.Server = ServerConfig{
        Port:         port,
        ReadTimeout:  parseDuration(\"READ_TIMEOUT\", \"15s\"),
        WriteTimeout: parseDuration(\"WRITE_TIMEOUT\", \"15s\"),
        IdleTimeout:  parseDuration(\"IDLE_TIMEOUT\", \"60s\"),
    }

    return config, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != \"\" {
        return value
    }
    return defaultValue
}

func parseDuration(key, defaultValue string) time.Duration {
    value := getEnv(key, defaultValue)
    duration, err := time.ParseDuration(value)
    if err != nil {
        duration, _ = time.ParseDuration(defaultValue)
    }
    return duration
}
```

### Health Checks and Graceful Shutdown
```go
package main

import (
    \"context\"
    \"fmt\"
    \"net/http\"
    \"os\"
    \"os/signal\"
    \"syscall\"
    \"time\"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set(\"Content-Type\", \"application/json\")
    fmt.Fprintf(w, `{\"status\":\"healthy\",\"timestamp\":%d}`, time.Now().Unix())
}

func gracefulShutdown(server *http.Server) {
    // Create channel to receive OS signals
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Block until signal received
    <-quit
    log.Println(\"Shutting down server...\")

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Shutdown server
    if err := server.Shutdown(ctx); err != nil {
        log.Printf(\"Server forced to shutdown: %v\", err)
    }

    log.Println(\"Server exited\")
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc(\"/health\", healthCheck)

    server := &http.Server{
        Addr:    \":8080\",
        Handler: mux,
    }

    // Start graceful shutdown handler
    go gracefulShutdown(server)

    log.Printf(\"Server starting on :8080\")
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf(\"Server failed to start: %v\", err)
    }
}
```

---

## Conclusion

This awesome Go resource represents the current state of systems programming and cloud-native development with Go. The ecosystem emphasizes simplicity, performance, and concurrent programming, making it ideal for building scalable systems, microservices, and cloud-native applications.

**Key Takeaway**: Go's ecosystem prioritizes simplicity, performance, and built-in concurrency. Choose libraries that align with these principles and leverage Go's strengths in systems programming and cloud-native development.

---

*This document is maintained to reflect the latest in Go ecosystem evolution and cloud-native development patterns.*