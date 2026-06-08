package token

import (
	"time"

	domainToken "github.com/gbrayhan/microservices-go/src/domain/token"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Token struct {
	ID        int                   `gorm:"primaryKey"`
	UserID    int                   `gorm:"index"` // Foreign key conceptually
	Token     string                `gorm:"uniqueIndex"`
	Type      domainToken.TokenType `gorm:"column:type"`
	ExpiresAt time.Time             `gorm:"column:expires_at"`
	CreatedAt time.Time             `gorm:"autoCreateTime:mili"`
}

type TokenRepositoryInterface interface {
	Create(token *domainToken.Token) error
	GetByTokenAndType(token string, tokenType domainToken.TokenType) (*domainToken.Token, error)
	DeleteByToken(token string) error
	DeleteAllUserTokensByType(userID int, tokenType domainToken.TokenType) error
}

type TokenRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewTokenRepository(db *gorm.DB, loggerInstance *logger.Logger) TokenRepositoryInterface {
	return &TokenRepository{
		DB:     db,
		Logger: loggerInstance,
	}
}

func fromDomainMapper(domain *domainToken.Token) *Token {
	return &Token{
		ID:        domain.ID,
		UserID:    domain.UserID,
		Token:     domain.Token,
		Type:      domain.Type,
		ExpiresAt: domain.ExpiresAt,
		CreatedAt: domain.CreatedAt,
	}
}

func (t *Token) toDomainMapper() *domainToken.Token {
	return &domainToken.Token{
		ID:        t.ID,
		UserID:    t.UserID,
		Token:     t.Token,
		Type:      t.Type,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
	}
}

func (r *TokenRepository) Create(domainTokenData *domainToken.Token) error {
	r.Logger.Info("Creating new token", zap.Int("userID", domainTokenData.UserID), zap.String("type", string(domainTokenData.Type)))
	repoToken := fromDomainMapper(domainTokenData)
	err := r.DB.Create(repoToken).Error
	if err != nil {
		r.Logger.Error("Error creating token", zap.Error(err))
		return err
	}
	domainTokenData.ID = repoToken.ID
	return nil
}

func (r *TokenRepository) GetByTokenAndType(tokenStr string, tokenType domainToken.TokenType) (*domainToken.Token, error) {
	var repoToken Token
	err := r.DB.Where("token = ? AND type = ?", tokenStr, tokenType).First(&repoToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil if not found
		}
		r.Logger.Error("Error getting token", zap.Error(err))
		return nil, err
	}
	return repoToken.toDomainMapper(), nil
}

func (r *TokenRepository) DeleteByToken(tokenStr string) error {
	err := r.DB.Where("token = ?", tokenStr).Delete(&Token{}).Error
	if err != nil {
		r.Logger.Error("Error deleting token", zap.Error(err))
		return err
	}
	return nil
}

func (r *TokenRepository) DeleteAllUserTokensByType(userID int, tokenType domainToken.TokenType) error {
	err := r.DB.Where("user_id = ? AND type = ?", userID, tokenType).Delete(&Token{}).Error
	if err != nil {
		r.Logger.Error("Error deleting user tokens", zap.Error(err))
		return err
	}
	return nil
}
