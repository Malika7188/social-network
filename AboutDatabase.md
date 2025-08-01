# SQLite Database Migration Guide for Social Network Project

This guide explains how to add new tables or update existing ones in the social network project's SQLite database system.

## Table of Contents

- [Overview](#overview)
- [Database Architecture](#database-architecture)
- [Adding a New Table](#adding-a-new-table)
- [Updating an Existing Table](#updating-an-existing-table)
- [Running Migrations](#running-migrations)
- [Advanced Features](#advanced-features)
- [Troubleshooting](#troubleshooting)

## Overview

The project uses a custom migration system built on top of the golang-migrate/migrate package. This system:

- Automatically generates SQL migration files from Go struct definitions
- Handles schema versioning and updates
- Provides both forward (up) and rollback (down) migrations
- Supports automatic detection of schema changes

## Database Architecture

The database layer consists of:

- **Model definitions**: Go structs in `pkg/models/dbTables/` with struct tags
- **Migration system**: Code in `pkg/db/sqlite/` that generates and runs migrations
- **Migration files**: SQL files in `pkg/db/migrations/sqlite/`

## Adding a New Table

### Step 1: Create a Model Struct

Create a new Go struct in the `pkg/models/dbTables/` directory:

```go
package models

import "time"

// Example represents a new table in the system
type Example struct {
    ID          int64     `db:"id,pk,autoincrement"`
    Name        string    `db:"name,notnull,unique" index:"unique"`
    Description string    `db:"description"`
    UserID      string    `db:"user_id,notnull" index:"" references:"users(id) ON DELETE CASCADE"`
    CreatedAt   time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
    UpdatedAt   time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`
}
```

### Step 2: Register the Model

Add your new model to the DiscoverModelStructs() function in pkg/db/sqlite/sqlite.go:

```go
// DiscoverModelStructs returns all exported struct types from the models package
func DiscoverModelStructs() []interface{} {
    return []interface{}{
        models.User{},
        models.Session{},
        models.Post{},
        models.PostViewer{},
        models.Comment{},
        models.FollowRequest{},
        models.Follower{},
        models.UserStats{},
        models.Example{}, // Add your new model here
        // Add new models here
    }
}
```

### Step 3: Generate and Run Migrations

The system will automatically generate migration files when you run the application.

## Updating an Existing Table

### Step 1: Modify the Model Struct

Update the existing struct in pkg/models/dbTables/ with your changes

```go
// User represents a user in the system
type User struct {
    ID          string    `db:"id,pk"`
    Email       string    `db:"email,notnull,unique" index:"unique"`
    Password    string    `db:"password,notnull"`
    FirstName   string    `db:"first_name,notnull"`
    LastName    string    `db:"last_name,notnull"`
    DateOfBirth string    `db:"date_of_birth,notnull"`
    Avatar      string    `db:"avatar"`
    Nickname    string    `db:"nickname"`
    AboutMe     string    `db:"about_me"`
    IsPublic    bool      `db:"is_public,default=TRUE"`
    // Add a new field
    PhoneNumber string    `db:"phone_number"`
    CreatedAt   time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
    UpdatedAt   time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`
}
```

### Step 2: Generate Update Migrations

The system will automatically detect schema changes and generate migration files when you run:

```go
if err := db.UpdateMigrations(); err != nil {
    log.Fatalf("Failed to update migrations: %v", err)
}
```

This is also called automatically by CreateMigrations().

## Running Migrations

Migrations are automatically run when the application starts. The system:

- Checks for "dirty" migrations (failed previous attempts)
- Applies any pending migrations
- Handles versioning automatically

## Advanced Features

### Struct Tags

The system supports several struct tags to define column properties:

```sh
db:"name,option1,option2,...": Main tag for column definition
pk: Primary key
autoincrement: Auto-incrementing (for INTEGER primary keys)
notnull: NOT NULL constraint
unique: UNIQUE constraint
default=value: Default value
references=table(column): Foreign key reference
pk(col1,col2): Composite primary key
index:"option1,option2,...": Creates an index on the column
unique: Creates a unique index
name=indexname: Custom index name
```

### Composite Primary Keys

For tables with composite primary keys:

```go
type PostViewer struct {
    PostID int64  `db:"post_id,notnull" index:""`
    UserID string `db:"user_id,notnull" index:""`
    // This tag defines a composite primary key on both columns
    _      struct{} `db:"pk(post_id,user_id)"`
}
```

### Custom SQL Types

The system automatically maps Go types to SQLite types:

```sh
bool → BOOLEAN
int/int64 → INTEGER
float64 → REAL
string → TEXT
time.Time → TIMESTAMP
[]byte → BLOB
```

## Troubleshooting

### Dirty Migrations

If a migration fails halfway through, the database will be marked as "dirty". The system will automatically attempt to fix this by forcing the version to the last known good state.

### Migration Conflicts

If you get errors about migration conflicts:

- Check if the migration already exists
- Ensure your model changes are compatible with existing data
- For major schema changes, consider creating a new table and migrating data manually

### Foreign Key Constraints

The system automatically disables foreign key constraints during migrations with `PRAGMA foreign_keys=off` and re-enables them afterward to prevent constraint violations during schema changes.

## For more information contact Ray
