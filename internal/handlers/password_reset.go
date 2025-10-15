package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// PasswordResetHandler gerencia operações de recuperação de senha
type PasswordResetHandler struct {
	db           *sql.DB
	emailService *services.EmailService
}

// NewPasswordResetHandler cria uma nova instância do handler
func NewPasswordResetHandler(db *sql.DB, emailService *services.EmailService) *PasswordResetHandler {
	return &PasswordResetHandler{
		db:           db,
		emailService: emailService,
	}
}

// PasswordResetCodeRequest representa a requisição de esqueci minha senha com código
type PasswordResetCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyResetCodeRequest representa a requisição de verificação de código
type VerifyResetCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// CompletePasswordResetRequest representa a requisição de redefinição de senha com código
type CompletePasswordResetRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// generateResetCode gera um código aleatório de 6 dígitos
func generateResetCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)

	randomBytes := make([]byte, 6)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i := 0; i < 6; i++ {
		code[i] = digits[int(randomBytes[i])%len(digits)]
	}

	return string(code), nil
}

// ForgotPassword solicita recuperação de senha e envia código por email
// @Summary Solicitar recuperação de senha
// @Description Envia um código de 6 dígitos para o email do usuário para recuperação de senha
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body PasswordResetCodeRequest true "Email do usuário"
// @Success 200 {object} map[string]interface{} "Código enviado com sucesso"
// @Failure 400 {object} map[string]interface{} "Dados inválidos"
// @Failure 429 {object} map[string]interface{} "Muitas tentativas, aguarde"
// @Failure 500 {object} map[string]interface{} "Erro interno"
// @Router /api/v1/auth/forgot-password [post]
func (h *PasswordResetHandler) ForgotPassword(c *gin.Context) {
	var req PasswordResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email inválido"})
		return
	}

	// Buscar usuário
	var userID uuid.UUID
	var userName string
	var deletedAt sql.NullTime

	err := h.db.QueryRow(`
		SELECT id, name, deleted_at 
		FROM users 
		WHERE email = $1
	`, req.Email).Scan(&userID, &userName, &deletedAt)

	if err == sql.ErrNoRows || deletedAt.Valid {
		// Por segurança, não revelamos se o email existe ou não
		c.JSON(http.StatusOK, gin.H{
			"message": "Se o email existir em nossa base, um código de recuperação será enviado",
		})

		logger.Warn("Tentativa de recuperação para email não encontrado",
			zap.String("email", req.Email),
			zap.String("ip", c.ClientIP()))
		return
	}

	if err != nil {
		logger.Error("Erro ao buscar usuário",
			zap.Error(err),
			zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar solicitação"})
		return
	}

	// Verificar rate limiting (máximo 3 tentativas em 15 minutos)
	var recentAttempts int
	err = h.db.QueryRow(`
		SELECT COUNT(*) 
		FROM password_reset_tokens 
		WHERE user_id = $1 
		AND created_at > NOW() - INTERVAL '15 minutes'
	`, userID).Scan(&recentAttempts)

	if err != nil {
		logger.Error("Erro ao verificar tentativas", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar solicitação"})
		return
	}

	if recentAttempts >= 3 {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Muitas tentativas. Aguarde 15 minutos e tente novamente",
		})
		return
	}

	// Gerar código de 6 dígitos
	code, err := generateResetCode()
	if err != nil {
		logger.Error("Erro ao gerar código", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar código"})
		return
	}

	// Salvar token no banco
	expiresAt := time.Now().Add(15 * time.Minute)
	_, err = h.db.Exec(`
		INSERT INTO password_reset_tokens 
		(user_id, token_code, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, code, expiresAt, c.ClientIP(), c.Request.UserAgent())

	if err != nil {
		logger.Error("Erro ao salvar token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar solicitação"})
		return
	}

	// Enviar email com código
	err = h.emailService.SendPasswordResetCode(req.Email, code, userName)
	if err != nil {
		logger.Error("Erro ao enviar email",
			zap.Error(err),
			zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar email"})
		return
	}

	logger.Info("Código de recuperação enviado",
		zap.String("user_id", userID.String()),
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Código de recuperação enviado para seu email",
		"expires_in": "15 minutos",
	})
}

// VerifyResetCode verifica se o código é válido
// @Summary Verificar código de recuperação
// @Description Verifica se o código enviado por email é válido
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyResetCodeRequest true "Email e código"
// @Success 200 {object} map[string]interface{} "Código válido"
// @Failure 400 {object} map[string]interface{} "Código inválido ou expirado"
// @Failure 500 {object} map[string]interface{} "Erro interno"
// @Router /api/v1/auth/verify-reset-code [post]
func (h *PasswordResetHandler) VerifyResetCode(c *gin.Context) {
	var req VerifyResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Buscar usuário
	var userID uuid.UUID
	err := h.db.QueryRow(`
		SELECT id 
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`, req.Email).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email não encontrado"})
		return
	}

	// Verificar token
	var tokenID uuid.UUID
	var usedAt sql.NullTime
	var expiresAt time.Time

	err = h.db.QueryRow(`
		SELECT token_id, used_at, expires_at 
		FROM password_reset_tokens 
		WHERE user_id = $1 
		AND token_code = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, req.Code).Scan(&tokenID, &usedAt, &expiresAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código inválido"})
		return
	}

	if err != nil {
		logger.Error("Erro ao verificar token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar código"})
		return
	}

	// Verificar se já foi usado
	if usedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código já foi utilizado"})
		return
	}

	// Verificar se expirou
	if time.Now().After(expiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código expirado. Solicite um novo código"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Código válido",
		"valid":   true,
	})
}

// ResetPassword redefine a senha usando o código válido
// @Summary Redefinir senha
// @Description Redefine a senha do usuário usando o código de recuperação
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body CompletePasswordResetRequest true "Email, código e nova senha"
// @Success 200 {object} map[string]interface{} "Senha alterada com sucesso"
// @Failure 400 {object} map[string]interface{} "Código inválido ou senha fraca"
// @Failure 500 {object} map[string]interface{} "Erro interno"
// @Router /api/v1/auth/reset-password [post]
func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req CompletePasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Validar força da senha
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha deve ter no mínimo 8 caracteres"})
		return
	}

	// Buscar usuário
	var userID uuid.UUID
	var userName string
	err := h.db.QueryRow(`
		SELECT id, name 
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`, req.Email).Scan(&userID, &userName)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email não encontrado"})
		return
	}

	// Iniciar transação
	tx, err := h.db.Begin()
	if err != nil {
		logger.Error("Erro ao iniciar transação", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar solicitação"})
		return
	}
	defer tx.Rollback()

	// Verificar e marcar token como usado
	var tokenID uuid.UUID
	var usedAt sql.NullTime
	var expiresAt time.Time

	err = tx.QueryRow(`
		SELECT token_id, used_at, expires_at 
		FROM password_reset_tokens 
		WHERE user_id = $1 
		AND token_code = $2
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`, userID, req.Code).Scan(&tokenID, &usedAt, &expiresAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código inválido"})
		return
	}

	if err != nil {
		logger.Error("Erro ao verificar token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar código"})
		return
	}

	if usedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código já foi utilizado"})
		return
	}

	if time.Now().After(expiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código expirado. Solicite um novo código"})
		return
	}

	// Hash da nova senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Erro ao gerar hash da senha", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar senha"})
		return
	}

	// Atualizar senha
	_, err = tx.Exec(`
		UPDATE users 
		SET password = $1, updated_at = NOW()
		WHERE id = $2
	`, string(hashedPassword), userID)

	if err != nil {
		logger.Error("Erro ao atualizar senha", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar senha"})
		return
	}

	// Marcar token como usado
	_, err = tx.Exec(`
		UPDATE password_reset_tokens 
		SET used_at = NOW()
		WHERE token_id = $1
	`, tokenID)

	if err != nil {
		logger.Error("Erro ao marcar token como usado", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar token"})
		return
	}

	// Invalidar todas as sessões ativas do usuário (por segurança)
	_, err = tx.Exec(`
		UPDATE user_sessions 
		SET active = false
		WHERE user_id = $1 AND active = true
	`, userID)

	if err != nil {
		logger.Error("Erro ao invalidar sessões", zap.Error(err))
		// Não retornamos erro aqui pois a senha já foi alterada
	}

	// Criar audit log para mudança de senha via recovery
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	metadata := map[string]interface{}{
		"change_method": "password_recovery",
		"token_id":      tokenID.String(),
		"ip_address":    clientIP,
		"user_agent":    userAgent,
		"changed_at":    time.Now().Format(time.RFC3339),
	}

	metadataJSON, _ := json.Marshal(metadata)
	resourceIDStr := userID.String()

	_, err = tx.Exec(`
		INSERT INTO audit_logs (
			id, user_id, action, resource, resource_id,
			ip_address, user_agent, metadata, success, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, uuid.New(), userID, "password_change", "user", resourceIDStr,
		clientIP, userAgent, metadataJSON, true, time.Now())

	if err != nil {
		logger.Error("Failed to create audit log for password reset",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		// Don't fail the request if audit log fails
	}

	// Commit da transação
	if err := tx.Commit(); err != nil {
		logger.Error("Erro ao finalizar transação", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao finalizar operação"})
		return
	}

	// Enviar email de confirmação (async)
	go func() {
		err := h.emailService.SendPasswordResetConfirmation(req.Email, userName)
		if err != nil {
			logger.Error("Erro ao enviar email de confirmação",
				zap.Error(err),
				zap.String("email", req.Email))
		}
	}()

	logger.Info("Senha redefinida com sucesso",
		zap.String("user_id", userID.String()),
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

	c.JSON(http.StatusOK, gin.H{
		"message": "Senha alterada com sucesso. Faça login com sua nova senha",
	})
}
