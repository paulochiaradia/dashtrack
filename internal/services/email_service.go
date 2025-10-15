package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"strings"

	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/logger"
	"go.uber.org/zap"
)

// EmailService gerencia o envio de emails
type EmailService struct {
	config *config.Config
}

// NewEmailService cria uma nova instância do serviço de email
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

// EmailData representa os dados de um email
type EmailData struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

// SendEmail envia um email usando SMTP
func (s *EmailService) SendEmail(data EmailData) error {
	// Validação básica
	if data.To == "" {
		return fmt.Errorf("email destinatário não pode estar vazio")
	}
	if data.Subject == "" {
		return fmt.Errorf("assunto do email não pode estar vazio")
	}

	// Configuração SMTP do Umbler
	from := s.config.SMTP.From
	password := s.config.SMTP.Password
	smtpHost := s.config.SMTP.Host
	smtpPort := s.config.SMTP.Port

	// Autenticação
	auth := smtp.PlainAuth("", s.config.SMTP.Username, password, smtpHost)

	// Construir mensagem
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.SMTP.FromName, from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", data.To))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))

	if data.IsHTML {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(data.Body)

	// Endereço do servidor SMTP
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	// Conectar com TLS (opcional mas recomendado)
	if s.config.SMTP.UseTLS {
		return s.sendWithTLS(addr, auth, from, []string{data.To}, msg.Bytes())
	}

	// Enviar sem TLS
	return smtp.SendMail(addr, auth, from, []string{data.To}, msg.Bytes())
}

// sendWithTLS envia email com criptografia STARTTLS (porta 587)
func (s *EmailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Separar host da porta
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("endereço SMTP inválido: %s", addr)
	}
	host := parts[0]

	// Conectar ao servidor SMTP (sem TLS inicialmente)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logger.Error("Erro ao conectar com servidor SMTP",
			zap.Error(err),
			zap.String("host", addr))
		return fmt.Errorf("erro ao conectar: %w", err)
	}
	defer conn.Close()

	// Criar cliente SMTP
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente SMTP: %w", err)
	}
	defer client.Close()

	// Iniciar STARTTLS
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		logger.Error("Erro ao iniciar STARTTLS",
			zap.Error(err),
			zap.String("host", addr))
		return fmt.Errorf("erro ao iniciar STARTTLS: %w", err)
	}

	// Autenticar
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			logger.Error("Erro de autenticação SMTP", zap.Error(err))
			return fmt.Errorf("erro de autenticação: %w", err)
		}
	}

	// Definir remetente
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("erro ao definir remetente: %w", err)
	}

	// Definir destinatários
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("erro ao definir destinatário %s: %w", addr, err)
		}
	}

	// Enviar corpo da mensagem
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("erro ao preparar envio: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("erro ao escrever mensagem: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("erro ao finalizar envio: %w", err)
	}

	logger.Info("Email enviado com sucesso",
		zap.Strings("to", to))

	return client.Quit()
}

// SendPasswordResetCode envia código de recuperação de senha
func (s *EmailService) SendPasswordResetCode(email, code, userName string) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 5px; margin-top: 20px; }
        .code { background-color: #fff; border: 2px dashed #4CAF50; padding: 15px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
        .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 10px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Recuperação de Senha - DashTrack</h1>
        </div>
        <div class="content">
            <p>Olá <strong>{{.UserName}}</strong>,</p>
            <p>Você solicitou a recuperação de senha da sua conta DashTrack.</p>
            <p>Use o código abaixo para redefinir sua senha:</p>
            
            <div class="code">{{.Code}}</div>
            
            <div class="warning">
                <strong>⚠️ Atenção:</strong>
                <ul style="margin: 5px 0;">
                    <li>Este código expira em <strong>15 minutos</strong></li>
                    <li>Pode ser usado apenas <strong>uma vez</strong></li>
                    <li>Se você não solicitou esta recuperação, ignore este email</li>
                </ul>
            </div>
            
            <p style="margin-top: 20px;">
                Para sua segurança, nunca compartilhe este código com ninguém.
            </p>
        </div>
        <div class="footer">
            <p>DashTrack - Sistema de Gestão de Entregas</p>
            <p>Este é um email automático, não responda.</p>
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("reset").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("erro ao criar template: %w", err)
	}

	var body bytes.Buffer
	err = t.Execute(&body, map[string]string{
		"UserName": userName,
		"Code":     code,
	})
	if err != nil {
		return fmt.Errorf("erro ao executar template: %w", err)
	}

	return s.SendEmail(EmailData{
		To:      email,
		Subject: "Recuperação de Senha - DashTrack",
		Body:    body.String(),
		IsHTML:  true,
	})
}

// SendPasswordResetConfirmation envia confirmação de senha alterada
func (s *EmailService) SendPasswordResetConfirmation(email, userName string) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 5px; margin-top: 20px; }
        .success { background-color: #d4edda; border-left: 4px solid #28a745; padding: 15px; margin: 15px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Senha Alterada com Sucesso</h1>
        </div>
        <div class="content">
            <p>Olá <strong>{{.UserName}}</strong>,</p>
            
            <div class="success">
                <strong>Sua senha foi alterada com sucesso!</strong>
            </div>
            
            <p>Se você não realizou esta alteração, entre em contato com o suporte imediatamente.</p>
            
            <p style="margin-top: 20px;">
                Você já pode fazer login com sua nova senha.
            </p>
        </div>
        <div class="footer">
            <p>DashTrack - Sistema de Gestão de Entregas</p>
            <p>Este é um email automático, não responda.</p>
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("confirmation").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("erro ao criar template: %w", err)
	}

	var body bytes.Buffer
	err = t.Execute(&body, map[string]string{
		"UserName": userName,
	})
	if err != nil {
		return fmt.Errorf("erro ao executar template: %w", err)
	}

	return s.SendEmail(EmailData{
		To:      email,
		Subject: "Senha Alterada - DashTrack",
		Body:    body.String(),
		IsHTML:  true,
	})
}
