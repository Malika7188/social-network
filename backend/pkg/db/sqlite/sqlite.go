package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// DB represents the database connection
type DB struct {
	*sql.DB
	config Config
}

// Config holds the database configuration
type Config struct {
	DBPath         string
	MigrationsPath string
}

// New creates a new database connection
func New(config Config) (*DB, error) {
	// Ensure the directory exists
	dbDir := filepath.Dir(config.DBPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	dsn := fmt.Sprintf(
		"%s?_foreign_keys=on&_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000",
		config.DBPath,
	)

	// Open the database connection
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
        db.Close()
        return nil, fmt.Errorf("pragma journal_mode: %w", err)
    }

	return &DB{db, config}, nil
}

// runMigrations runs the database migrations
func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Check if database is dirty
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	// If database is dirty, force the version
	if dirty {
		logger.Warn("Database is in dirty state at version %d, forcing version", version)
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("failed to force version: %w", err)
		}
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations applied successfully")
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// TableInfo holds information about a database table
type TableInfo struct {
	Name                string
	Columns             []ColumnInfo
	Indexes             []IndexInfo
	CompositePrimaryKey []string
}

// ColumnInfo holds information about a table column
type ColumnInfo struct {
	Name       string
	Type       string
	PrimaryKey bool
	NotNull    bool
	Unique     bool
	Default    string
	References string
}

// IndexInfo holds information about a table index
type IndexInfo struct {
	Name    string
	Columns []string
	Unique  bool
}

// DiscoverModelStructs returns all exported struct types from the models package
func DiscoverModelStructs() []interface{} {
	// Define the models we know about
	// This is still manual but only needs to be updated when adding a new model type(Database Table)
	return []interface{}{
		models.User{},
		models.Session{},
		models.Post{},
		models.PostViewer{},
		models.Comment{},
		models.FollowRequest{},
		models.Follower{},
		models.UserStat{},
		models.PostLike{},
		models.UserStatus{},
		models.Group{},
		models.GroupMember{},
		models.GroupChatMessage{},
		models.GroupPost{},
		models.GroupEvent{},
		models.EventResponse{},
		models.PrivateMessage{},
		models.ChatContact{},
		models.Notification{},
		models.UserProfile{},
		// Add new models here
	}
}

// CreateMigrationFromStruct generates migration files from a struct
func (db *DB) CreateMigrationFromStruct(modelStruct interface{}, migrationName string) error {
	// Check if migration already exists for this table
	tableName := camelToSnake(reflect.TypeOf(modelStruct).Name())
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}
	existingMigration, err := checkMigrationExists(db.config.MigrationsPath, tableName)
	if err != nil {
		return fmt.Errorf("failed to check existing migrations: %w", err)
	}

	if existingMigration {
		logger.Info("Migration for table %s already exists, skipping", tableName)
		return nil
	}

	// Get table info from struct
	tableInfo, err := extractTableInfoFromStruct(modelStruct)
	if err != nil {
		return fmt.Errorf("failed to extract table info: %w", err)
	}

	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll(db.config.MigrationsPath, 0o755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate migration sequence number
	seq, err := getNextMigrationSequence(db.config.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get next migration sequence: %w", err)
	}

	// Generate migration file names
	upFileName := fmt.Sprintf("%06d_%s.up.sql", seq, migrationName)
	downFileName := fmt.Sprintf("%06d_%s.down.sql", seq, migrationName)

	// Generate SQL for up migration
	upSQL := generateCreateTableSQL(tableInfo)

	// Generate SQL for down migration
	downSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableInfo.Name)

	// Write migration files
	upFilePath := filepath.Join(db.config.MigrationsPath, upFileName)
	if err := os.WriteFile(upFilePath, []byte(upSQL), 0o644); err != nil {
		return fmt.Errorf("failed to write up migration file: %w", err)
	}

	downFilePath := filepath.Join(db.config.MigrationsPath, downFileName)
	if err := os.WriteFile(downFilePath, []byte(downSQL), 0o644); err != nil {
		return fmt.Errorf("failed to write down migration file: %w", err)
	}

	logger.Info("Created migration files: %s, %s", upFileName, downFileName)
	return nil
}

// checkMigrationExists checks if a migration for the given table already exists
func checkMigrationExists(migrationsPath string, tableName string) (bool, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if filename contains create_tablename_table
		if strings.Contains(file.Name(), fmt.Sprintf("create_%s_table", tableName)) {
			return true, nil
		}
	}

	return false, nil
}

func CreateMigrations(db *DB) error {
	// Discover all model structs
	logger.Info("Discovering model structs...")
	modelStructs := DiscoverModelStructs()

	// Generate migrations for all models
	logger.Info("Generating database migrations...")
	for _, model := range modelStructs {
		// Get the struct name
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		structName := t.Name()

		// Generate migration name
		tableName := camelToSnake(structName)
		// Only add 's' if the name doesn't already end with 's'
		if !strings.HasSuffix(tableName, "s") {
			tableName += "s"
		}
		migrationName := fmt.Sprintf("create_%s_table", tableName)

		// Create migration
		if err := db.CreateMigrationFromStruct(model, migrationName); err != nil {
			logger.Fatal("Failed to create migration for %s: %v", structName, err)
			return err
		}
	}

	// Check for schema updates
	if err := db.UpdateMigrations(); err != nil {
		logger.Fatal("Failed to update migrations: %v", err)
		return err
	}

	// Run migrations
	if err := runMigrations(db.DB, db.config.MigrationsPath); err != nil {
		db.DB.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// getNextMigrationSequence determines the next sequence number for migrations
func getNextMigrationSequence(migrationsPath string) (int, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxSeq := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if len(name) < 6 {
			continue
		}

		var seq int
		_, err := fmt.Sscanf(name, "%d_", &seq)
		if err == nil && seq > maxSeq {
			maxSeq = seq
		}
	}

	return maxSeq + 1, nil
}

// extractTableInfoFromStruct extracts table information from a struct
func extractTableInfoFromStruct(modelStruct interface{}) (TableInfo, error) {
	tableInfo := TableInfo{}

	t := reflect.TypeOf(modelStruct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if this struct should be ignored for database operations
	if shouldIgnoreStruct(t) {
		return tableInfo, fmt.Errorf("skipping non-DB struct: %s", t.Name())
	}

	// Get table name from struct name (convert CamelCase to snake_case and pluralize)
	tableName := camelToSnake(t.Name())
	// Only add 's' if the name doesn't already end with 's'
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}
	tableInfo.Name = tableName

	// Check for composite primary keys
	var compositePKColumns []string

	// First pass to find composite primary key
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check for composite primary key tag
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			parts := strings.Split(dbTag, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "pk(") && strings.HasSuffix(part, ")") {
					// Extract column names from pk(col1,col2)
					pkCols := strings.TrimSuffix(strings.TrimPrefix(part, "pk("), ")")
					compositePKColumns = strings.Split(pkCols, ",")
				}
			}
		}
	}

	// Process struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get column info from field tags
		column := extractColumnInfoFromField(field)
		if column.Name != "" {
			// Check if this column is part of a composite primary key
			for _, pkCol := range compositePKColumns {
				if pkCol == column.Name {
					// Don't mark individual columns as PK when they're part of a composite key
					column.PrimaryKey = false
				}
			}
			tableInfo.Columns = append(tableInfo.Columns, column)
		}

		// Check for index tag
		indexTag := field.Tag.Get("index")
		if indexTag != "" {
			parts := strings.Split(indexTag, ",")
			indexName := fmt.Sprintf("idx_%s_%s", tableName, column.Name)
			unique := false

			for _, part := range parts {
				if part == "unique" {
					unique = true
				} else if strings.HasPrefix(part, "name=") {
					indexName = strings.TrimPrefix(part, "name=")
				}
			}

			tableInfo.Indexes = append(tableInfo.Indexes, IndexInfo{
				Name:    indexName,
				Columns: []string{column.Name},
				Unique:  unique,
			})
		}
	}

	// Add composite primary key constraint if needed
	if len(compositePKColumns) > 0 {
		tableInfo.CompositePrimaryKey = compositePKColumns
	}

	return tableInfo, nil
}

// shouldIgnoreStruct determines if a struct should be ignored for database operations
func shouldIgnoreStruct(t reflect.Type) bool {
	// Check if the struct has a db tag at the type level
	if dbTag, ok := t.FieldByName("db"); ok {
		if dbTag.Tag.Get("db") == "-" {
			return true
		}
	}

	// Check if all fields have db:"-" tags
	allFieldsIgnored := true
	hasFields := false

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		hasFields = true
		dbTag := field.Tag.Get("db")
		if dbTag != "-" {
			allFieldsIgnored = false
			break
		}
	}

	// If the struct has fields and all of them are ignored, skip this struct
	if hasFields && allFieldsIgnored {
		return true
	}

	// Skip structs that are commonly used for embedding but not as DB tables
	switch t.Name() {
	case "PostUserData", "UserData", "EmbeddedData":
		return true
	}

	return false
}

// extractColumnInfoFromField extracts column information from a struct field
func extractColumnInfoFromField(field reflect.StructField) ColumnInfo {
	column := ColumnInfo{}

	// Get column name from db tag or field name
	dbTag := field.Tag.Get("db")
	if dbTag == "-" {
		return column // Skip this field
	}

	if dbTag != "" {
		parts := strings.Split(dbTag, ",")
		column.Name = parts[0]

		// Process options
		for _, part := range parts[1:] {
			switch part {
			case "pk":
				column.PrimaryKey = true
			case "notnull":
				column.NotNull = true
			case "unique":
				column.Unique = true
			default:
				if strings.HasPrefix(part, "default=") {
					column.Default = strings.TrimPrefix(part, "default=")
				} else if strings.HasPrefix(part, "references=") {
					column.References = strings.TrimPrefix(part, "references=")
				}
			}
		}
	} else {
		column.Name = camelToSnake(field.Name)
	}

	// Map Go types to SQLite types
	column.Type = mapGoTypeToSQLite(field.Type)

	return column
}

// generateCreateTableSQL generates SQL for creating a table
func generateCreateTableSQL(table TableInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", table.Name))

	// Add columns
	for i, col := range table.Columns {
		if i > 0 {
			sb.WriteString(",\n")
		}

		sb.WriteString(fmt.Sprintf("    %s %s", col.Name, col.Type))

		if col.PrimaryKey {
			if strings.ToUpper(col.Type) == "INTEGER" {
				sb.WriteString(" PRIMARY KEY AUTOINCREMENT")
			} else {
				sb.WriteString(" PRIMARY KEY")
			}
		}

		if col.NotNull {
			sb.WriteString(" NOT NULL")
		}

		if col.Unique {
			sb.WriteString(" UNIQUE")
		}

		if col.Default != "" {
			sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}

		if col.References != "" {
			sb.WriteString(fmt.Sprintf(" REFERENCES %s", col.References))
		}
	}

	// Add composite primary key if needed
	if len(table.CompositePrimaryKey) > 0 {
		sb.WriteString(",\n    PRIMARY KEY (")
		for i, col := range table.CompositePrimaryKey {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(col)
		}
		sb.WriteString(")")
	}

	sb.WriteString("\n);\n\n")

	// Add indexes
	for _, idx := range table.Indexes {
		uniqueStr := ""
		if idx.Unique {
			uniqueStr = "UNIQUE "
		}

		columns := strings.Join(idx.Columns, ", ")
		sb.WriteString(fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS %s ON %s(%s);\n",
			uniqueStr, idx.Name, table.Name, columns))
	}

	return sb.String()
}

// mapGoTypeToSQLite maps Go types to SQLite types
func mapGoTypeToSQLite(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.String:
		return "TEXT"
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "TIMESTAMP"
		}
		return "TEXT" // JSON serialization for other structs
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "BLOB" // []byte
		}
		return "TEXT" // JSON serialization for other slices
	default:
		return "TEXT"
	}
}

// camelToSnake converts a CamelCase string to snake_case
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// CreateTable creates a table from a struct if it doesn't exist
func (db *DB) CreateTable(modelStruct interface{}) error {
	tableInfo, err := extractTableInfoFromStruct(modelStruct)
	if err != nil {
		return fmt.Errorf("failed to extract table info: %w", err)
	}

	// Generate SQL
	sql := generateCreateTableSQL(tableInfo)

	// Execute SQL
	_, err = db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	logger.Info("Created table: %s", tableInfo.Name)
	return nil
}

// AutoMigrate automatically creates tables from structs
func (db *DB) AutoMigrate(modelStructs ...interface{}) error {
	for _, model := range modelStructs {
		if err := db.CreateTable(model); err != nil {
			return err
		}
	}
	return nil
}

// UpdateMigrations checks for schema changes and creates migration files if needed
func (db *DB) UpdateMigrations() error {
	logger.Info("Checking for schema changes...")

	// Discover all model structs
	modelStructs := DiscoverModelStructs()

	for _, model := range modelStructs {
		if err := db.checkAndUpdateSchema(model); err != nil {
			return fmt.Errorf("failed to update schema for %T: %w", model, err)
		}
	}

	return nil
}

// checkAndUpdateSchema compares model struct with database table and creates migration if needed
func (db *DB) checkAndUpdateSchema(modelStruct interface{}) error {
	t := reflect.TypeOf(modelStruct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	tableName := camelToSnake(t.Name())
	// Only add 's' if the name doesn't already end with 's'
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}

	// Extract current table info from struct
	newTableInfo, err := extractTableInfoFromStruct(modelStruct)
	if err != nil {
		return fmt.Errorf("failed to extract table info from struct: %w", err)
	}

	// Check if table exists in database
	var count int
	err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	// If table doesn't exist yet, no need to create update migration
	if count == 0 {
		return nil
	}

	// Get current table schema from database
	currentTableInfo, err := db.getTableInfoFromDB(tableName)
	if err != nil {
		return fmt.Errorf("failed to get table info from database: %w", err)
	}

	// Compare schemas
	if !schemasEqual(currentTableInfo, newTableInfo) {
		// Generate migration for schema update
		return db.createSchemaMigration(tableName, currentTableInfo, newTableInfo)
	}

	return nil
}

// getTableInfoFromDB extracts table information from the database
func (db *DB) getTableInfoFromDB(tableName string) (TableInfo, error) {
	tableInfo := TableInfo{
		Name: tableName,
	}

	// Get column information
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return tableInfo, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typeName string
		var notNull, pk int
		var dfltValue interface{}

		if err := rows.Scan(&cid, &name, &typeName, &notNull, &dfltValue, &pk); err != nil {
			return tableInfo, err
		}

		column := ColumnInfo{
			Name:       name,
			Type:       typeName,
			NotNull:    notNull == 1,
			PrimaryKey: pk > 0,
		}

		if dfltValue != nil {
			column.Default = fmt.Sprintf("%v", dfltValue)
		}

		tableInfo.Columns = append(tableInfo.Columns, column)
	}

	// Get index information
	query = fmt.Sprintf("PRAGMA index_list(%s)", tableName)
	rows, err = db.Query(query)
	if err != nil {
		return tableInfo, err
	}
	defer rows.Close()

	for rows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int

		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return tableInfo, err
		}

		// Skip sqlite_autoindex entries
		if strings.HasPrefix(name, "sqlite_autoindex") {
			continue
		}

		// Get columns for this index
		var columns []string
		indexInfoQuery := fmt.Sprintf("PRAGMA index_info(%s)", name)
		indexRows, err := db.Query(indexInfoQuery)
		if err != nil {
			return tableInfo, err
		}

		for indexRows.Next() {
			var seqno, cid int
			var colName string

			if err := indexRows.Scan(&seqno, &cid, &colName); err != nil {
				indexRows.Close()
				return tableInfo, err
			}

			columns = append(columns, colName)
		}
		indexRows.Close()

		tableInfo.Indexes = append(tableInfo.Indexes, IndexInfo{
			Name:    name,
			Columns: columns,
			Unique:  unique == 1,
		})
	}

	return tableInfo, nil
}

// schemasEqual compares two table schemas to check if they're equivalent
func schemasEqual(current, new TableInfo) bool {
	// Compare column count
	if len(current.Columns) != len(new.Columns) {
		return false
	}

	// Create maps for easier comparison
	currentCols := make(map[string]ColumnInfo)
	for _, col := range current.Columns {
		currentCols[col.Name] = col
	}

	// Compare columns
	for _, newCol := range new.Columns {
		currentCol, exists := currentCols[newCol.Name]
		if !exists {
			return false // Column doesn't exist in current schema
		}

		// Compare column properties
		if currentCol.Type != newCol.Type ||
			currentCol.NotNull != newCol.NotNull ||
			currentCol.PrimaryKey != newCol.PrimaryKey {
			return false
		}
	}

	// Compare indexes (simplified)
	if len(current.Indexes) != len(new.Indexes) {
		return false
	}

	return true
}

// createSchemaMigration creates migration files for schema updates
func (db *DB) createSchemaMigration(tableName string, currentSchema, newSchema TableInfo) error {
	// Generate migration sequence number
	seq, err := getNextMigrationSequence(db.config.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get next migration sequence: %w", err)
	}

	migrationName := fmt.Sprintf("update_%s_schema", tableName)

	// Generate migration file names
	upFileName := fmt.Sprintf("%06d_%s.up.sql", seq, migrationName)
	downFileName := fmt.Sprintf("%06d_%s.down.sql", seq, migrationName)

	// Generate SQL for up migration (create temp table, copy data, drop old, rename new)
	var upSQL strings.Builder
	upSQL.WriteString(fmt.Sprintf("-- Migration to update %s table schema\n\n", tableName))
	upSQL.WriteString("PRAGMA foreign_keys=off;\n\n")

	// Create new table with _new suffix
	tempTableName := tableName + "_new"
	upSQL.WriteString(fmt.Sprintf("-- Create new table with updated schema\n"))
	upSQL.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tempTableName))

	// Add columns
	for i, col := range newSchema.Columns {
		if i > 0 {
			upSQL.WriteString(",\n")
		}

		upSQL.WriteString(fmt.Sprintf("    %s %s", col.Name, col.Type))

		if col.PrimaryKey {
			if strings.Contains(col.Type, "INTEGER") {
				upSQL.WriteString(" PRIMARY KEY AUTOINCREMENT")
			} else {
				upSQL.WriteString(" PRIMARY KEY")
			}
		}

		if col.NotNull {
			upSQL.WriteString(" NOT NULL")
		}

		if col.Unique {
			upSQL.WriteString(" UNIQUE")
		}

		if col.Default != "" {
			upSQL.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}

		if col.References != "" {
			upSQL.WriteString(fmt.Sprintf(" REFERENCES %s", col.References))
		}
	}

	// Add composite primary key if needed
	if len(newSchema.CompositePrimaryKey) > 0 {
		upSQL.WriteString(",\n    PRIMARY KEY (")
		for i, col := range newSchema.CompositePrimaryKey {
			if i > 0 {
				upSQL.WriteString(", ")
			}
			upSQL.WriteString(col)
		}
		upSQL.WriteString(")")
	}

	upSQL.WriteString("\n);\n\n")

	// Copy data from old table to new table
	upSQL.WriteString("-- Copy data from old table to new table\n")
	upSQL.WriteString(fmt.Sprintf("INSERT INTO %s (", tempTableName))

	// Find common columns between old and new schemas
	var commonColumns []string
	for _, newCol := range newSchema.Columns {
		for _, currentCol := range currentSchema.Columns {
			if newCol.Name == currentCol.Name {
				commonColumns = append(commonColumns, newCol.Name)
				break
			}
		}
	}

	// Add column names
	for i, col := range commonColumns {
		if i > 0 {
			upSQL.WriteString(", ")
		}
		upSQL.WriteString(col)
	}

	upSQL.WriteString(")\nSELECT ")

	// Add column names again for SELECT
	for i, col := range commonColumns {
		if i > 0 {
			upSQL.WriteString(", ")
		}
		upSQL.WriteString(col)
	}

	upSQL.WriteString(fmt.Sprintf(" FROM %s;\n\n", tableName))

	// Drop old table and rename new table
	upSQL.WriteString(fmt.Sprintf("-- Drop old table and rename new table\n"))
	upSQL.WriteString(fmt.Sprintf("DROP TABLE %s;\n", tableName))
	upSQL.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;\n\n", tempTableName, tableName))

	// Add indexes
	for _, idx := range newSchema.Indexes {
		uniqueStr := ""
		if idx.Unique {
			uniqueStr = "UNIQUE "
		}

		columns := strings.Join(idx.Columns, ", ")
		upSQL.WriteString(fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS %s ON %s(%s);\n",
			uniqueStr, idx.Name, tableName, columns))
	}

	upSQL.WriteString("\nPRAGMA foreign_keys=on;\n")

	// Generate SQL for down migration (restore original schema)
	var downSQL strings.Builder
	downSQL.WriteString(fmt.Sprintf("-- Revert migration for %s table\n\n", tableName))
	downSQL.WriteString("PRAGMA foreign_keys=off;\n\n")

	// Create temp table with original schema
	downSQL.WriteString(fmt.Sprintf("-- Create table with original schema\n"))
	downSQL.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tempTableName))

	// Add columns
	for i, col := range currentSchema.Columns {
		if i > 0 {
			downSQL.WriteString(",\n")
		}

		downSQL.WriteString(fmt.Sprintf("    %s %s", col.Name, col.Type))

		if col.PrimaryKey {
			if strings.Contains(col.Type, "INTEGER") {
				downSQL.WriteString(" PRIMARY KEY AUTOINCREMENT")
			} else {
				downSQL.WriteString(" PRIMARY KEY")
			}
		}

		if col.NotNull {
			downSQL.WriteString(" NOT NULL")
		}

		if col.Unique {
			downSQL.WriteString(" UNIQUE")
		}

		if col.Default != "" {
			downSQL.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
	}

	// Add composite primary key if needed
	if len(currentSchema.CompositePrimaryKey) > 0 {
		downSQL.WriteString(",\n    PRIMARY KEY (")
		for i, col := range currentSchema.CompositePrimaryKey {
			if i > 0 {
				downSQL.WriteString(", ")
			}
			downSQL.WriteString(col)
		}
		downSQL.WriteString(")")
	}

	downSQL.WriteString("\n);\n\n")

	// Copy data back (best effort)
	downSQL.WriteString("-- Copy data back (best effort)\n")
	downSQL.WriteString(fmt.Sprintf("INSERT INTO %s (", tempTableName))

	// Add column names
	for i, col := range commonColumns {
		if i > 0 {
			downSQL.WriteString(", ")
		}
		downSQL.WriteString(col)
	}

	downSQL.WriteString(")\nSELECT ")

	// Add column names again for SELECT
	for i, col := range commonColumns {
		if i > 0 {
			downSQL.WriteString(", ")
		}
		downSQL.WriteString(col)
	}

	downSQL.WriteString(fmt.Sprintf(" FROM %s;\n\n", tableName))

	// Drop new table and rename temp table
	downSQL.WriteString(fmt.Sprintf("-- Drop new table and rename temp table\n"))
	downSQL.WriteString(fmt.Sprintf("DROP TABLE %s;\n", tableName))
	downSQL.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;\n\n", tempTableName, tableName))

	// Add original indexes
	for _, idx := range currentSchema.Indexes {
		uniqueStr := ""
		if idx.Unique {
			uniqueStr = "UNIQUE "
		}

		columns := strings.Join(idx.Columns, ", ")
		downSQL.WriteString(fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS %s ON %s(%s);\n",
			uniqueStr, idx.Name, tableName, columns))
	}

	downSQL.WriteString("\nPRAGMA foreign_keys=on;\n")

	// Write migration files
	upFilePath := filepath.Join(db.config.MigrationsPath, upFileName)
	if err := os.WriteFile(upFilePath, []byte(upSQL.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write up migration file: %w", err)
	}

	downFilePath := filepath.Join(db.config.MigrationsPath, downFileName)
	if err := os.WriteFile(downFilePath, []byte(downSQL.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write down migration file: %w", err)
	}

	logger.Info("Created schema update migration files: %s, %s", upFileName, downFileName)
	return nil
}
