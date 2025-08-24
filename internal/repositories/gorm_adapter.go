package repositories

import (
	"errors"
	"fmt"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormAdapter adapter for GORM DB
type GormAdapter struct {
	db *gorm.DB
}

// NewGormAdapter creates a new adapter for GORM with config
func NewGormAdapter(dbConfig config.DBConfig) (IDatabase, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &GormAdapter{db: db}, nil
}

// NewGormAdapterFromDB creates a new adapter from existing GORM DB (for testing)
func NewGormAdapterFromDB(db *gorm.DB) IDatabase {
	return &GormAdapter{db: db}
}

func (g *GormAdapter) Create(value interface{}) IDatabase {
	if g.db == nil {
		return &GormAdapter{db: nil}
	}
	return &GormAdapter{db: g.db.Create(value)}
}

func (g *GormAdapter) Where(query interface{}, args ...interface{}) IDatabase {
	if g.db == nil {
		return &GormAdapter{db: nil}
	}
	return &GormAdapter{db: g.db.Where(query, args...)}
}

func (g *GormAdapter) First(dest interface{}, conds ...interface{}) IDatabase {
	if g.db == nil {
		return &GormAdapter{db: nil}
	}
	return &GormAdapter{db: g.db.First(dest, conds...)}
}

func (g *GormAdapter) Model(value interface{}) IDatabase {
	if g.db == nil {
		return &GormAdapter{db: nil}
	}
	return &GormAdapter{db: g.db.Model(value)}
}

func (g *GormAdapter) Count(value *int64) IDatabase {
	if g.db == nil {
		return &GormAdapter{db: nil}
	}
	return &GormAdapter{db: g.db.Count(value)}
}

func (g *GormAdapter) GetError() error {
	if g.db == nil {
		return errors.New("database is nil")
	}
	return g.db.Error
}
