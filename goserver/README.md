# Database Code Generation (GORM Gen)

This Go project uses [GORM Gen](https://github.com/go-gorm/gen) to generate type-safe, reflection-free query builders for database models.

---

## 🛠️ How to Run Codegen

To regenerate the query builders after model updates:

1. Navigate to the `goserver/` directory:
   ```bash
   cd goserver
   ```
2. Run the generator script:
   ```bash
   go run cmd/gen/main.go
   ```

This updates all generated query builders located in [internal/db/query/](file:///home/lachl/documents/dealscanner/goserver/internal/db/query/).

---

## ➕ Adding or Updating Database Models

Follow these steps when you need to introduce new tables/models or change existing columns:

### 1. Define the Model Struct
Add or edit the model struct in [internal/models/models.go](file:///home/lachl/documents/dealscanner/goserver/internal/models/models.go). Make sure to include GORM tags specifying column mappings and database constraints:

```go
type MyNewModel struct {
    ID        string    `gorm:"primaryKey;column:id"`
    Name      string    `gorm:"column:name"`
    CreatedAt time.Time `gorm:"column:createdAt"`
}

func (MyNewModel) TableName() string {
    return "MyNewModel"
}
```

### 2. Register with GORM Auto-Migration
To ensure GORM automatically creates/modifies the table in SQLite at startup, add the model to the `AutoMigrate` function in [internal/models/db.go](file:///home/lachl/documents/dealscanner/goserver/internal/models/db.go):

```go
// internal/models/db.go
err = db.AutoMigrate(
    ...
    &MyNewModel{},
)
```

### 3. Register with GORM Gen Builder
Add the model to the generator list in [cmd/gen/main.go](file:///home/lachl/documents/dealscanner/goserver/cmd/gen/main.go):

```go
// cmd/gen/main.go
g.ApplyBasic(
    ...
    &models.MyNewModel{},
)
```

### 4. Run Code Generation
Execute the generator to create the helper files:
```bash
go run cmd/gen/main.go
```

---

## 💻 Usage Example

Using the generated query builder ensures compile-time check verification without manual SQL strings or empty interface assertions:

```go
import (
    "context"
    qry "dealscanner/internal/db/query"
)

func FindActiveListings(ctx context.Context) ([]*models.Listing, error) {
    // Access the database using the generated default query manager
    l := qry.Listing
    
    return l.WithContext(ctx).
        Where(l.Price.Lte(100.0)).
        Order(l.CreatedAt.Desc()).
        Find()
}
```
