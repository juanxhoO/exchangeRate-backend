package exchanger

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domainExchanger "github.com/gbrayhan/microservices-go/src/domain/exchanger"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)
	cleanup := func() { db.Close() }
	return gormDB, mock, cleanup
}

func setupLogger(t *testing.T) *logger.Logger {
	loggerInstance, err := logger.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return loggerInstance
}

func TestTableName(t *testing.T) {
	u := &Exchanger{}
	assert.Equal(t, "users", u.TableName())
}

func TestNewUserRepository(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewUserRepository(db, logger)
	assert.NotNil(t, repo)
}

func TestToDomainMapper(t *testing.T) {
	u := &Exchanger{
		ID:        1,
		Name:      "testuser",
		ApiKey:    "test@example.com",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	d := u.toDomainMapper()
	assert.Equal(t, u.Name, d.Name)
}

func TestFromDomainMapper(t *testing.T) {
	d := &domainExchanger.Exchanger{
		ID:        1,
		Name:      "testuser",
		IsActive:  true,
		ApiKey:    "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	u := fromDomainMapper(d)
	assert.Equal(t, d.Name, u.Name)
}

func TestArrayToDomainMapper(t *testing.T) {
	arr := &[]Exchanger{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}
	d := arrayToDomainMapper(arr)
	assert.Len(t, *d, 2)
	assert.Equal(t, "A", (*d)[0].Name)
}

func TestRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewUserRepository(db, logger)
	rows := sqlmock.NewRows([]string{"id", "user_name", "email", "first_name", "last_name", "status", "hash_password"}).
		AddRow(1, "user1", "a@a.com", "A", "B", true, "hash1").
		AddRow(2, "user2", "b@b.com", "C", "D", false, "hash2")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).WillReturnRows(rows)
	users, err := repo.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, *users, 2)
}

func TestRepository_GetByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewUserRepository(db, logger)
	rows := sqlmock.NewRows([]string{"id", "user_name", "email", "first_name", "last_name", "status", "hash_password"}).
		AddRow(1, "user1", "a@a.com", "A", "B", true, "hash1")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(1, 1).WillReturnRows(rows)
	user, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user1", user.Name)
	// Not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(2, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "user_name", "email", "first_name", "last_name", "status", "hash_password"}))
	user, err = repo.GetByID(2)
	assert.Error(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 0, user.ID) // Should be zero value
}

func TestRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewUserRepository(db, logger)
	domainU := &domainExchanger.Exchanger{
		Name:     "user1",
		IsActive: false,
		ApiKey:   "A",
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	user, err := repo.Create(domainU)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user1", user.Name)
}

func TestRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewUserRepository(db, logger)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err := repo.Delete(1)
	assert.NoError(t, err)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	err = repo.Delete(2)
	assert.Error(t, err)
}

// The following tests need refactoring to use sqlmock or should be moved to integration:
// TestRepository_GetOneByMap
// TestRepository_Update
// TestRepository_Create_DuplicateEmail
// TestRepository_ErrorCases
// TestRepository_GetOneByMap_WithFilters
// TestRepository_Update_WithMultipleFields
//
// If you want me to refactor these as well, let me know and I'll do them one by one.
