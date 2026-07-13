package repositories

import (
	"time"

	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type TokenRepository struct{}

func NewTokenRepository() *TokenRepository {
	return &TokenRepository{}
}

func (r *TokenRepository) CreateRefreshToken(t *models.RefreshToken) error {
	return database.DB.Create(t).Error
}

func (r *TokenRepository) FindRefreshTokenByHash(hash string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	if err := database.DB.Where("token_hash = ?", hash).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) RevokeRefreshToken(hash string) error {
	return database.DB.Model(&models.RefreshToken{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

// RevokeRefreshTokenIfActive atomically flips revoked=false -> true in a
// single conditional UPDATE and reports whether this call was the one that
// did it. Used for single-use rotation: without this, a separate
// read-then-write (find, check Revoked, then revoke) lets two concurrent
// requests presenting the same still-valid refresh token both pass the
// check before either revokes it, letting one token mint two token pairs.
func (r *TokenRepository) RevokeRefreshTokenIfActive(hash string) (bool, error) {
	result := database.DB.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked = ?", hash, false).
		Update("revoked", true)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *TokenRepository) RevokeAllUserRefreshTokens(userID uuid.UUID) error {
	return database.DB.Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

func (r *TokenRepository) CreatePasswordResetToken(t *models.PasswordResetToken) error {
	return database.DB.Create(t).Error
}

func (r *TokenRepository) FindPasswordResetTokenByHash(hash string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	if err := database.DB.Where("token_hash = ? AND used = false AND expires_at > ?", hash, time.Now()).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) MarkPasswordResetTokenUsed(id uuid.UUID) error {
	return database.DB.Model(&models.PasswordResetToken{}).Where("id = ?", id).Update("used", true).Error
}

func (r *TokenRepository) CreateSession(s *models.Session) error {
	return database.DB.Create(s).Error
}

func (r *TokenRepository) DeactivateSessionByUser(userID uuid.UUID) error {
	return database.DB.Model(&models.Session{}).Where("user_id = ?", userID).Update("is_active", false).Error
}
