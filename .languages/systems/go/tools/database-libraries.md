# Go Database Libraries & Patterns

## SQL Database Drivers

### PostgreSQL Drivers

#### pgx - Advanced PostgreSQL Driver
- **Repository**: github.com/jackc/pgx
- **Stars**: 10k+
- **Features**:
  - Native PostgreSQL protocol
  - Connection pooling
  - Advanced data types support
  - High performance
  - Context support
```go
import (
    "context"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    // Single connection
    conn, err := pgx.Connect(context.Background(), "postgres://user:pass@localhost/db")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(context.Background())

    // Connection pool
    pool, err := pgxpool.New(context.Background(), "postgres://user:pass@localhost/db")
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Query with context
    var name string
    var age int
    err = pool.QueryRow(context.Background(), 
        "SELECT name, age FROM users WHERE id = $1", 
        userID).Scan(&name, &age)
}
```

#### pq - PostgreSQL Driver (database/sql)
- **Repository**: github.com/lib/pq
- **Stars**: 9k+
- **Features**:
  - Pure Go implementation
  - database/sql compatibility
  - SSL support
  - Array support
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", 
        "postgres://user:password@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    rows, err := db.Query("SELECT id, name FROM users WHERE age > $1", 18)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            log.Fatal(err)
        }
        fmt.Printf("ID: %d, Name: %s\n", id, name)
    }
}
```

### MySQL Drivers

#### go-sql-driver/mysql - Official MySQL Driver
- **Repository**: github.com/go-sql-driver/mysql
- **Stars**: 14k+
- **Features**:
  - Pure Go implementation
  - Fast performance
  - Connection pooling
  - Prepared statements
```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Prepared statement
    stmt, err := db.Prepare("SELECT name FROM users WHERE id = ?")
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

    var name string
    err = stmt.QueryRow(1).Scan(&name)
    if err != nil {
        log.Fatal(err)
    }
}
```

### SQLite Drivers

#### go-sqlite3 - SQLite Driver
- **Repository**: github.com/mattn/go-sqlite3
- **Stars**: 8k+
- **Features**:
  - CGO-based (requires gcc)
  - Full SQLite support
  - User-defined functions
  - Extensions support
```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./database.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create table
    sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        age INTEGER
    );`
    _, err = db.Exec(sqlStmt)
    if err != nil {
        log.Fatal(err)
    }
}
```

## ORM Libraries

### GORM - Full-Featured ORM
- **Repository**: github.com/go-gorm/gorm
- **Stars**: 36k+
- **Features**:
  - Auto migrations
  - Associations
  - Hooks
  - Transactions
  - Context support

```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

type User struct {
    ID       uint   `gorm:"primaryKey"`
    Name     string `gorm:"size:100;not null"`
    Email    string `gorm:"uniqueIndex"`
    Age      int
    Profile  Profile
}

type Profile struct {
    ID     uint
    UserID uint
    Bio    string
}

func main() {
    dsn := "host=localhost user=user password=pass dbname=test port=5432"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Auto-migrate
    db.AutoMigrate(&User{}, &Profile{})

    // Create
    user := User{Name: "John", Email: "john@example.com", Age: 30}
    result := db.Create(&user)
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    // Query
    var users []User
    db.Where("age > ?", 18).Find(&users)

    // Update
    db.Model(&user).Update("age", 31)

    // Delete
    db.Delete(&user, user.ID)

    // Preload associations
    var usersWithProfiles []User
    db.Preload("Profile").Find(&usersWithProfiles)
}

// Hooks
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
    if u.Name == "" {
        return errors.New("name cannot be empty")
    }
    return
}
```

### Ent - Entity Framework
- **Repository**: github.com/ent/ent
- **Stars**: 15k+
- **Features**:
  - Code generation
  - Type safety
  - Schema as code
  - Hooks and interceptors
```go
// Schema definition
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

type User struct {
    ent.Schema
}

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
        field.String("email").Unique(),
        field.Int("age").Positive(),
    }
}

func (User) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("posts", Post.Type),
    }
}

// Usage
func main() {
    client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Run migration
    if err := client.Schema.Create(context.Background()); err != nil {
        log.Fatal(err)
    }

    // Create user
    user, err := client.User.
        Create().
        SetName("John").
        SetEmail("john@example.com").
        SetAge(30).
        Save(context.Background())

    // Query
    users, err := client.User.
        Query().
        Where(user.AgeGT(18)).
        All(context.Background())
}
```

## Query Builders

### Squirrel - SQL Query Builder
- **Repository**: github.com/Masterminds/squirrel
- **Stars**: 7k+
- **Purpose**: Fluent SQL generation
```go
import "github.com/Masterminds/squirrel"

func main() {
    psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
    
    // SELECT query
    query, args, err := psql.
        Select("name", "age").
        From("users").
        Where(squirrel.Eq{"active": true}).
        Where(squirrel.Gt{"age": 18}).
        OrderBy("name").
        Limit(10).
        ToSql()
    
    if err != nil {
        log.Fatal(err)
    }
    
    rows, err := db.Query(query, args...)
    
    // INSERT query
    query, args, err = psql.
        Insert("users").
        Columns("name", "email", "age").
        Values("John", "john@example.com", 30).
        Values("Jane", "jane@example.com", 25).
        ToSql()
    
    // UPDATE query
    query, args, err = psql.
        Update("users").
        Set("age", squirrel.Expr("age + 1")).
        Where(squirrel.Eq{"name": "John"}).
        ToSql()
}
```

### sqlx - Extensions to database/sql
- **Repository**: github.com/jmoiron/sqlx
- **Stars**: 16k+
- **Purpose**: Enhanced standard library
```go
import "github.com/jmoiron/sqlx"

type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
    Age   int    `db:"age"`
}

func main() {
    db, err := sqlx.Connect("postgres", "postgres://user:pass@localhost/db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Scan into struct
    var user User
    err = db.Get(&user, "SELECT * FROM users WHERE id = $1", 1)
    if err != nil {
        log.Fatal(err)
    }

    // Scan into slice of structs
    var users []User
    err = db.Select(&users, "SELECT * FROM users WHERE age > $1", 18)
    if err != nil {
        log.Fatal(err)
    }

    // Named queries
    user = User{Name: "John", Email: "john@example.com", Age: 30}
    _, err = db.NamedExec(
        "INSERT INTO users (name, email, age) VALUES (:name, :email, :age)",
        &user)
}
```

## NoSQL Databases

### MongoDB Driver
- **Repository**: github.com/mongodb/mongo-go-driver
- **Stars**: 8k+
- **Features**:
  - Official MongoDB driver
  - Context support
  - BSON marshaling
  - Aggregation pipeline
```go
import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
)

type User struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Email string             `bson:"email"`
    Age   int                `bson:"age"`
}

func main() {
    client, err := mongo.Connect(context.TODO(), 
        options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(context.TODO())

    collection := client.Database("testdb").Collection("users")

    // Insert
    user := User{Name: "John", Email: "john@example.com", Age: 30}
    result, err := collection.InsertOne(context.TODO(), user)
    if err != nil {
        log.Fatal(err)
    }

    // Find one
    var foundUser User
    err = collection.FindOne(context.TODO(), 
        bson.M{"name": "John"}).Decode(&foundUser)

    // Find many
    cursor, err := collection.Find(context.TODO(), 
        bson.M{"age": bson.M{"$gt": 18}})
    if err != nil {
        log.Fatal(err)
    }
    defer cursor.Close(context.TODO())

    var users []User
    if err = cursor.All(context.TODO(), &users); err != nil {
        log.Fatal(err)
    }

    // Update
    filter := bson.M{"name": "John"}
    update := bson.M{"$set": bson.M{"age": 31}}
    _, err = collection.UpdateOne(context.TODO(), filter, update)

    // Aggregation
    pipeline := []bson.M{
        {"$match": bson.M{"age": bson.M{"$gte": 18}}},
        {"$group": bson.M{
            "_id": "$age",
            "count": bson.M{"$sum": 1},
        }},
    }
    cursor, err = collection.Aggregate(context.TODO(), pipeline)
}
```

### Redis Client
- **Repository**: github.com/redis/go-redis
- **Stars**: 20k+
- **Features**:
  - Connection pooling
  - Pub/Sub support
  - Pipeline operations
  - Cluster support
```go
import "github.com/redis/go-redis/v9"

func main() {
    ctx := context.Background()
    
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    defer rdb.Close()

    // Basic operations
    err := rdb.Set(ctx, "key", "value", 0).Err()
    if err != nil {
        log.Fatal(err)
    }

    val, err := rdb.Get(ctx, "key").Result()
    if err != nil {
        log.Fatal(err)
    }

    // Hash operations
    err = rdb.HSet(ctx, "user:1", "name", "John", "age", 30).Err()
    vals := rdb.HGetAll(ctx, "user:1").Val()

    // List operations
    err = rdb.LPush(ctx, "queue", "task1", "task2").Err()
    result, err := rdb.BRPop(ctx, 0, "queue").Result()

    // Pipeline
    pipe := rdb.Pipeline()
    incr := pipe.Incr(ctx, "pipeline_counter")
    pipe.Expire(ctx, "pipeline_counter", time.Hour)
    _, err = pipe.Exec(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(incr.Val())

    // Pub/Sub
    pubsub := rdb.Subscribe(ctx, "mychannel")
    defer pubsub.Close()

    go func() {
        for msg := range pubsub.Channel() {
            fmt.Printf("Received: %s\n", msg.Payload)
        }
    }()

    err = rdb.Publish(ctx, "mychannel", "hello").Err()
}
```

## Migration Tools

### golang-migrate - Database Migrations
- **Repository**: github.com/golang-migrate/migrate
- **Stars**: 15k+
- **Purpose**: Database schema migrations
```bash
# Install CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq create_users_table

# Run migrations
migrate -database "postgres://user:pass@localhost/db?sslmode=disable" -path migrations up

# Rollback
migrate -database "postgres://user:pass@localhost/db?sslmode=disable" -path migrations down 1
```

Migration files:
```sql
-- 001_create_users_table.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 001_create_users_table.down.sql
DROP TABLE users;
```

Programmatic usage:
```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations() {
    m, err := migrate.New(
        "file://migrations",
        "postgres://user:pass@localhost/db?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }
}
```

## Database Testing Patterns

### Test Containers
- **Repository**: github.com/testcontainers/testcontainers-go
- **Purpose**: Integration testing with real databases
```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image:        "postgres:13",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "password",
            "POSTGRES_USER":     "user",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }

    postgres, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    require.NoError(t, err)

    host, err := postgres.Host(ctx)
    require.NoError(t, err)

    port, err := postgres.MappedPort(ctx, "5432")
    require.NoError(t, err)

    dsn := fmt.Sprintf("host=%s port=%d user=user password=password dbname=testdb sslmode=disable",
        host, port.Int())

    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)

    return db, func() {
        db.Close()
        postgres.Terminate(ctx)
    }
}

func TestUserRepository(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewUserRepository(db)
    
    // Run your tests
    user := &User{Name: "John", Email: "john@example.com"}
    err := repo.Create(user)
    assert.NoError(t, err)
}
```

### In-Memory Databases for Testing
```go
func TestWithSQLite(t *testing.T) {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    defer db.Close()

    // Create schema
    _, err = db.Exec(`
        CREATE TABLE users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL
        )
    `)
    require.NoError(t, err)

    // Test your repository
    repo := NewUserRepository(db)
    // ... tests
}
```

These database libraries and patterns provide comprehensive solutions for all data persistence needs in Go applications.