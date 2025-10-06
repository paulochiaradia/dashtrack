package services

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// TwoFactorService handles 2FA operations
type TwoFactorService struct {
	db *sqlx.DB
}

// NewTwoFactorService creates a new 2FA service
func NewTwoFactorService(db *sqlx.DB) *TwoFactorService {
	return &TwoFactorService{
		db: db,
	}
}

// TwoFactorSetup represents the setup response for 2FA
type TwoFactorSetup struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// SetupTwoFactor initiates 2FA setup for a user
func (tfs *TwoFactorService) SetupTwoFactor(ctx context.Context, userID uuid.UUID, issuer, accountName string) (*TwoFactorSetup, error) {
	// Check if 2FA is already enabled
	existing, err := tfs.getTwoFactorByUserID(ctx, userID)
	if err == nil && existing.Enabled {
		return nil, fmt.Errorf("2FA is already enabled for this user")
	}

	// Generate secret
	secret, err := tfs.generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	// Generate backup codes
	backupCodes, err := tfs.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Generate QR code URL
	qrURL, err := tfs.generateQRCodeURL(secret, issuer, accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code URL: %w", err)
	}

	// Store in database (disabled until verified)
	twoFA := &models.TwoFactorAuth{
		ID:          uuid.New(),
		UserID:      userID,
		Secret:      secret,
		BackupCodes: backupCodes,
		Enabled:     false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = tfs.storeTwoFactor(ctx, twoFA)
	if err != nil {
		return nil, fmt.Errorf("failed to store 2FA setup: %w", err)
	}

	return &TwoFactorSetup{
		Secret:      secret,
		QRCodeURL:   qrURL,
		BackupCodes: backupCodes,
	}, nil
}

// EnableTwoFactor enables 2FA after verifying the initial code
func (tfs *TwoFactorService) EnableTwoFactor(ctx context.Context, userID uuid.UUID, code string) error {
	// Get 2FA setup
	twoFA, err := tfs.getTwoFactorByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA setup not found: %w", err)
	}

	if twoFA.Enabled {
		return fmt.Errorf("2FA is already enabled")
	}

	// Verify code
	valid, err := tfs.verifyTOTPCode(twoFA.Secret, code)
	if err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid verification code")
	}

	// Enable 2FA
	err = tfs.enableTwoFactor(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	logger.Info("2FA enabled for user", zap.String("user_id", userID.String()))
	return nil
}

// VerifyTwoFactor verifies a 2FA code for login
func (tfs *TwoFactorService) VerifyTwoFactor(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	// Get 2FA configuration
	twoFA, err := tfs.getTwoFactorByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("2FA not configured: %w", err)
	}

	if !twoFA.Enabled {
		return false, fmt.Errorf("2FA is not enabled")
	}

	// Try TOTP code first
	valid, err := tfs.verifyTOTPCode(twoFA.Secret, code)
	if err != nil {
		return false, fmt.Errorf("failed to verify TOTP code: %w", err)
	}

	if valid {
		// Update last used time
		err = tfs.updateLastUsed(ctx, userID)
		if err != nil {
			logger.Error("Failed to update 2FA last used time", zap.Error(err))
		}
		return true, nil
	}

	// Try backup codes
	if tfs.isBackupCode(twoFA.BackupCodes, code) {
		// Remove used backup code
		err = tfs.removeBackupCode(ctx, userID, code)
		if err != nil {
			logger.Error("Failed to remove used backup code", zap.Error(err))
		}
		return true, nil
	}

	return false, nil
}

// DisableTwoFactor disables 2FA for a user
func (tfs *TwoFactorService) DisableTwoFactor(ctx context.Context, userID uuid.UUID, code string) error {
	// Verify current code before disabling
	valid, err := tfs.VerifyTwoFactor(ctx, userID, code)
	if err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid verification code")
	}

	// Delete 2FA configuration
	err = tfs.deleteTwoFactor(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	logger.Info("2FA disabled for user", zap.String("user_id", userID.String()))
	return nil
}

// IsTwoFactorEnabled checks if 2FA is enabled for a user
func (tfs *TwoFactorService) IsTwoFactorEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	twoFA, err := tfs.getTwoFactorByUserID(ctx, userID)
	if err != nil {
		return false, nil // Not found means not enabled
	}
	return twoFA.Enabled, nil
}

// GenerateBackupCodes generates new backup codes for a user
func (tfs *TwoFactorService) GenerateBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Verify 2FA is enabled
	enabled, err := tfs.IsTwoFactorEnabled(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check 2FA status: %w", err)
	}

	if !enabled {
		return nil, fmt.Errorf("2FA is not enabled")
	}

	// Generate new backup codes
	backupCodes, err := tfs.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Update in database
	err = tfs.updateBackupCodes(ctx, userID, backupCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return backupCodes, nil
}

// generateSecret generates a new TOTP secret
func (tfs *TwoFactorService) generateSecret() (string, error) {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

// generateBackupCodes generates backup codes
func (tfs *TwoFactorService) generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := tfs.generateBackupCode()
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// generateBackupCode generates a single backup code
func (tfs *TwoFactorService) generateBackupCode() (string, error) {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Convert to 8-digit number
	code := ""
	for _, b := range bytes {
		code += fmt.Sprintf("%02d", int(b)%100)
	}

	return code, nil
}

// generateQRCodeURL generates a QR code URL for TOTP setup
func (tfs *TwoFactorService) generateQRCodeURL(secret, issuer, accountName string) (string, error) {
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		issuer, accountName, secret, issuer))
	if err != nil {
		return "", err
	}
	return key.URL(), nil
}

// verifyTOTPCode verifies a TOTP code
func (tfs *TwoFactorService) verifyTOTPCode(secret, code string) (bool, error) {
	return totp.Validate(code, secret), nil
}

// isBackupCode checks if a code is a valid backup code
func (tfs *TwoFactorService) isBackupCode(backupCodes []string, code string) bool {
	for _, backupCode := range backupCodes {
		if backupCode == code {
			return true
		}
	}
	return false
}

// Database operations

func (tfs *TwoFactorService) getTwoFactorByUserID(ctx context.Context, userID uuid.UUID) (*models.TwoFactorAuth, error) {
	query := `
		SELECT id, user_id, secret, backup_codes, enabled, last_used, created_at, updated_at
		FROM two_factor_auth
		WHERE user_id = $1
	`

	var twoFA models.TwoFactorAuth
	err := tfs.db.GetContext(ctx, &twoFA, query, userID)
	return &twoFA, err
}

func (tfs *TwoFactorService) storeTwoFactor(ctx context.Context, twoFA *models.TwoFactorAuth) error {
	query := `
		INSERT INTO two_factor_auth (id, user_id, secret, backup_codes, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			secret = EXCLUDED.secret,
			backup_codes = EXCLUDED.backup_codes,
			enabled = EXCLUDED.enabled,
			updated_at = EXCLUDED.updated_at
	`

	_, err := tfs.db.ExecContext(ctx, query,
		twoFA.ID, twoFA.UserID, twoFA.Secret, tfs.stringSliceToJSON(twoFA.BackupCodes),
		twoFA.Enabled, twoFA.CreatedAt, twoFA.UpdatedAt)

	return err
}

func (tfs *TwoFactorService) enableTwoFactor(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE two_factor_auth
		SET enabled = true, updated_at = NOW()
		WHERE user_id = $1
	`

	_, err := tfs.db.ExecContext(ctx, query, userID)
	return err
}

func (tfs *TwoFactorService) updateLastUsed(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE two_factor_auth
		SET last_used = NOW(), updated_at = NOW()
		WHERE user_id = $1
	`

	_, err := tfs.db.ExecContext(ctx, query, userID)
	return err
}

func (tfs *TwoFactorService) removeBackupCode(ctx context.Context, userID uuid.UUID, code string) error {
	// Get current backup codes
	twoFA, err := tfs.getTwoFactorByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Remove the used code
	var newCodes []string
	for _, backupCode := range twoFA.BackupCodes {
		if backupCode != code {
			newCodes = append(newCodes, backupCode)
		}
	}

	// Update in database
	query := `
		UPDATE two_factor_auth
		SET backup_codes = $1, updated_at = NOW()
		WHERE user_id = $2
	`

	_, err = tfs.db.ExecContext(ctx, query, tfs.stringSliceToJSON(newCodes), userID)
	return err
}

func (tfs *TwoFactorService) updateBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error {
	query := `
		UPDATE two_factor_auth
		SET backup_codes = $1, updated_at = NOW()
		WHERE user_id = $2
	`

	_, err := tfs.db.ExecContext(ctx, query, tfs.stringSliceToJSON(codes), userID)
	return err
}

func (tfs *TwoFactorService) deleteTwoFactor(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM two_factor_auth WHERE user_id = $1`
	_, err := tfs.db.ExecContext(ctx, query, userID)
	return err
}

// Helper function to convert string slice to JSON
func (tfs *TwoFactorService) stringSliceToJSON(codes []string) string {
	return `["` + strings.Join(codes, `","`) + `"]`
}
