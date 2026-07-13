package services

import (
	"errors"
	"strconv"
	"time"

	"legaltech-backend/internal/config"
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	cfg          *config.Config
	userRepo     *repositories.UserRepository
	tokenRepo    *repositories.TokenRepository
	clientRepo   *repositories.ClientRepository
	emailService *EmailService
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		cfg:          cfg,
		userRepo:     repositories.NewUserRepository(),
		tokenRepo:    repositories.NewTokenRepository(),
		clientRepo:   repositories.NewClientRepository(),
		emailService: NewEmailService(cfg),
	}
}

func (s *AuthService) Signup(req models.SignupRequest) (*models.User, string, string, error) {
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, "", "", errors.New("email already registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", "", errors.New("failed to hash password")
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Phone:    req.Phone,
		Role:     "advocate",
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, "", "", errors.New("failed to create user")
	}

	accessToken, refreshToken, err := s.issueTokenPair(&user)
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, "", "", errors.New("invalid email or password")
	}

	// Client accounts are temporary: blocked the moment their lawyer has no
	// open case left for them (see CaseService.syncClientAccountStatus).
	if user.Role == models.RoleClient {
		if client, err := s.clientRepo.FindByAccountUserID(user.ID); err == nil &&
			client.AccountStatus == models.ClientAccountInactive {
			return nil, "", "", errors.New("Your case has been closed. Please contact your advocate if you need further access.")
		}
	}

	accessToken, refreshToken, err := s.issueTokenPair(user)
	if err != nil {
		return nil, "", "", err
	}

	s.setOnline(user, true)

	return user, accessToken, refreshToken, nil
}

// setOnline is the single place IsOnline/LastSeenAt ever change — a real
// login-session presence, not tied to whether a WebSocket happens to be
// connected at this instant (see internal/ws/hub.go, which is now only a
// best-effort real-time push, not the source of truth for "online").
func (s *AuthService) setOnline(user *models.User, online bool) {
	now := time.Now()
	user.IsOnline = online
	user.LastSeenAt = &now
	_ = s.userRepo.Update(user)
}

func (s *AuthService) RefreshToken(rawRefreshToken string) (string, string, error) {
	hash := utils.HashToken(rawRefreshToken)

	stored, err := s.tokenRepo.FindRefreshTokenByHash(hash)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return "", "", errors.New("refresh token expired or revoked")
	}

	user, err := s.userRepo.FindByID(stored.UserID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// Rotate: atomically revoke the used refresh token and issue a new pair.
	// The atomic conditional update is what makes rotation single-use —
	// if two requests race on the same token, only one flips revoked=false
	// -> true and gets a new pair; the other is rejected here instead of
	// both succeeding.
	revoked, err := s.tokenRepo.RevokeRefreshTokenIfActive(hash)
	if err != nil {
		return "", "", err
	}
	if !revoked {
		return "", "", errors.New("refresh token expired or revoked")
	}

	accessToken, newRefreshToken, err := s.issueTokenPair(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(rawRefreshToken string) error {
	hash := utils.HashToken(rawRefreshToken)

	// Best-effort: mark the user offline before revoking. A lookup failure
	// here (token already revoked/expired) shouldn't block logout itself.
	if stored, err := s.tokenRepo.FindRefreshTokenByHash(hash); err == nil {
		if user, err := s.userRepo.FindByID(stored.UserID); err == nil {
			s.setOnline(user, false)
		}
	}

	return s.tokenRepo.RevokeRefreshToken(hash)
}

func (s *AuthService) LogoutAllDevices(userID uuid.UUID) error {
	if user, err := s.userRepo.FindByID(userID); err == nil {
		s.setOnline(user, false)
	}
	return s.tokenRepo.RevokeAllUserRefreshTokens(userID)
}

// ForgotPassword always succeeds from the caller's perspective to avoid leaking
// whether an email is registered. It emails a short-lived reset code when the
// account exists.
func (s *AuthService) ForgotPassword(email string) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return
	}

	code, err := utils.GenerateResetCode(8)
	if err != nil {
		return
	}

	resetToken := models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: utils.HashToken(code),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.tokenRepo.CreatePasswordResetToken(&resetToken); err != nil {
		return
	}

	s.emailService.SendPasswordResetEmail(user.Email, user.Name, code)
}

func (s *AuthService) ResetPassword(code, newPassword string) error {
	hash := utils.HashToken(code)

	resetToken, err := s.tokenRepo.FindPasswordResetTokenByHash(hash)
	if err != nil {
		return errors.New("invalid or expired reset code")
	}

	user, err := s.userRepo.FindByID(resetToken.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword
	user.MustChangePassword = false
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	_ = s.tokenRepo.MarkPasswordResetTokenUsed(resetToken.ID)
	_ = s.tokenRepo.RevokeAllUserRefreshTokens(user.ID)

	return nil
}

func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !utils.CheckPasswordHash(currentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword
	user.MustChangePassword = false
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *AuthService) issueTokenPair(user *models.User) (string, string, error) {
	accessToken, err := utils.GenerateJWT(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiryHr)
	if err != nil {
		return "", "", errors.New("failed to generate access token")
	}

	rawRefreshToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", errors.New("failed to generate refresh token")
	}

	refreshDays, err := strconv.Atoi(s.cfg.JWTRefreshExpiryDays)
	if err != nil {
		refreshDays = 30
	}

	refreshRecord := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: utils.HashToken(rawRefreshToken),
		ExpiresAt: time.Now().Add(time.Duration(refreshDays) * 24 * time.Hour),
	}

	if err := s.tokenRepo.CreateRefreshToken(&refreshRecord); err != nil {
		return "", "", errors.New("failed to persist refresh token")
	}

	return accessToken, rawRefreshToken, nil
}
