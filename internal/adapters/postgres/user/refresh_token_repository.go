package user

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func rtToDomain(m RefreshTokenModel) user.RefreshToken {
	return user.RefreshToken{
		ID:              m.ID,
		UserID:          m.UserID,
		FamilyID:        m.FamilyID,
		FamilyCreatedAt: m.FamilyCreatedAt,
		TokenHash:       m.TokenHash,
		ExpiresAt:       m.ExpiresAt,
		Revoked:         m.Revoked,
		CreatedAt:       m.CreatedAt,
		UserAgent:       m.UserAgent,
		IPAddress:       m.IPAddress,
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, rt user.RefreshToken) (user.RefreshToken, error) {
	m := RefreshTokenModel{
		UserID:          rt.UserID,
		FamilyID:        rt.FamilyID,
		FamilyCreatedAt: rt.FamilyCreatedAt,
		TokenHash:       rt.TokenHash,
		ExpiresAt:       rt.ExpiresAt,
		Revoked:         rt.Revoked,
		CreatedAt:       rt.CreatedAt,
		UserAgent:       rt.UserAgent,
		IPAddress:       rt.IPAddress,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return user.RefreshToken{}, err
	}
	return rtToDomain(m), nil
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (user.RefreshToken, error) {
	var m RefreshTokenModel
	err := r.db.WithContext(ctx).First(&m, "token_hash = ?", tokenHash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user.RefreshToken{}, shared.ErrNotFound
	}
	if err != nil {
		return user.RefreshToken{}, err
	}
	return rtToDomain(m), nil
}

// Revoke ignores tokenHash - Postgres already has id, kept only so a
// caching decorator wrapping this repository can invalidate by hash too.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id int64, tokenHash string) error {
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("id = ?", id).Update("revoked", true).Error
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

func (r *RefreshTokenRepository) FindTokenHashesByUserID(ctx context.Context, userID int64) ([]string, error) {
	var hashes []string
	err := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Pluck("token_hash", &hashes).Error
	return hashes, err
}
