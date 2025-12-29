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
	assert.Equal(t, "exchangers", u.TableName())
}

func TestNewExchangerRepository(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewExchangerRepository(db, logger)
	assert.NotNil(t, repo)
}

func TestToDomainMapper(t *testing.T) {
	u := &Exchanger{
		ID:        1,
		Name:      "apiexchange",
		ApiKey:    "dsdsd11232d",
		IsActive:  true,
		Url:       "https://api.exchange.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	d := u.toDomainMapper()
	assert.Equal(t, u.Name, d.Name)
	assert.Equal(t, u.ApiKey, d.ApiKey)
	assert.Equal(t, u.IsActive, d.IsActive)
	assert.Equal(t, u.Url, d.Url)
}

func TestFromDomainMapper(t *testing.T) {
	d := &domainExchanger.Exchanger{
		ID:        1,
		Name:      "apiexchange",
		IsActive:  true,
		ApiKey:    "dsdsd11232d",
		Url:       "https://api.exchange.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	u := fromDomainMapper(d)
	assert.Equal(t, d.Name, u.Name)
	assert.Equal(t, d.ApiKey, u.ApiKey)
	assert.Equal(t, d.IsActive, u.IsActive)
	assert.Equal(t, d.Url, u.Url)
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
	repo := NewExchangerRepository(db, logger)
	rows := sqlmock.NewRows([]string{"id", "name", "api_key", "is_active", "url", "created_at", "updated_at"}).
		AddRow(1, "apiexchange", "dsdsd11232d", true, "https://api.exchange.com", time.Now(), time.Now()).
		AddRow(2, "apiexchange2", "dsdsd11232d", true, "https://api.exchange.com", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "exchangers"`)).WillReturnRows(rows)
	users, err := repo.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, *users, 2)
}

func TestRepository_GetByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewExchangerRepository(db, logger)
	rows := sqlmock.NewRows([]string{"id", "name", "api_key", "is_active", "url", "created_at", "updated_at"}).
		AddRow(1, "apiexchange", "dsdsd11232d", true, "https://api.exchange.com", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "exchangers" WHERE id = $1 ORDER BY "exchangers"."id" LIMIT $2`)).
		WithArgs(1, 1).WillReturnRows(rows)
	user, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "apiexchange", user.Name)
	// Not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "exchangers" WHERE id = $1 ORDER BY "exchangers"."id" LIMIT $2`)).
		WithArgs(2, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "api_key", "is_active", "url", "created_at", "updated_at"}))
	user, err = repo.GetByID(2)
	assert.Error(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 0, user.ID) // Should be zero value
}

func TestRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewExchangerRepository(db, logger)
	domainU := &domainExchanger.Exchanger{
		Name:     "apiexchange",
		IsActive: false,
		ApiKey:   "A",
		Url:      "https://api.exchange.com",
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "exchangers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	user, err := repo.Create(domainU)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "apiexchange", user.Name)
}

func TestRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	logger := setupLogger(t)
	repo := NewExchangerRepository(db, logger)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "exchangers" WHERE "exchangers"."id" = $1`)).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err := repo.Delete(1)
	assert.NoError(t, err)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "exchangers" WHERE "exchangers"."id" = $1`)).
		WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	err = repo.Delete(2)
	assert.Error(t, err)
}

// The following tests need refactoring to use sqlmock or should be moved to integration:
// TestRepository_GetOneByMap
// TestRepository_Update
// TestRepository_ErrorCases
// TestRepository_Update_WithMultipleFields

// If you want me to refactor these as well, let me know and I'll do them one by one.
