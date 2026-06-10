package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

const mailjetURL = "https://api.mailjet.com/v3.1/send"

type Mailer struct {
	// SMTP config (prioritaire sur Mailjet)
	smtpHost string
	smtpPort string
	smtpUser string
	smtpPass string

	// Mailjet config (fallback)
	apiKey    string
	secretKey string

	fromEmail string
	fromName  string
	client    *http.Client
}

func New() *Mailer {
	from := os.Getenv("MAIL_FROM")
	smtpUser := os.Getenv("SMTP_USER")
	if from == "" {
		if smtpUser != "" {
			from = smtpUser
		} else {
			from = "noreply@oz-backoffice.com"
		}
	}
	return &Mailer{
		smtpHost:  os.Getenv("SMTP_HOST"),
		smtpPort:  os.Getenv("SMTP_PORT"),
		smtpUser:  smtpUser,
		smtpPass:  os.Getenv("SMTP_PASS"),
		apiKey:    os.Getenv("MAILJET_API_KEY"),
		secretKey: os.Getenv("MAILJET_SECRET_KEY"),
		fromEmail: from,
		fromName:  "O'Z Backoffice",
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *Mailer) smtpEnabled() bool {
	return m.smtpHost != "" && m.smtpUser != "" && m.smtpPass != ""
}

func (m *Mailer) mailjetEnabled() bool {
	return m.apiKey != "" && m.secretKey != ""
}

// Enabled retourne true si au moins un transport est configuré.
func (m *Mailer) Enabled() bool {
	return m.smtpEnabled() || m.mailjetEnabled()
}

func (m *Mailer) send(subject, htmlBody, textBody, toEmail, toName string) error {
	if m.smtpEnabled() {
		return m.sendSMTP(subject, htmlBody, toEmail, toName)
	}
	if m.mailjetEnabled() {
		return m.sendMailjet(mjMessage{
			From:     mjContact{Email: m.fromEmail, Name: m.fromName},
			To:       []mjContact{{Email: toEmail, Name: toName}},
			Subject:  subject,
			HTMLPart: htmlBody,
			TextPart: textBody,
		})
	}
	fmt.Printf("[MAILER] (dry-run) To: %s | Subject: %s\n", toEmail, subject)
	return nil
}

// ── SMTP ────────────────────────────────────────────────────────────────────

func (m *Mailer) sendSMTP(subject, htmlBody, toEmail, toName string) error {
	port := m.smtpPort
	if port == "" {
		port = "587"
	}
	addr := m.smtpHost + ":" + port
	auth := smtp.PlainAuth("", m.smtpUser, m.smtpPass, m.smtpHost)

	header := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s <%s>\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		m.fromName, m.fromEmail, toName, toEmail, subject,
	)
	msg := []byte(header + htmlBody)
	return smtp.SendMail(addr, auth, m.fromEmail, []string{toEmail}, msg)
}

// ── Mailjet ─────────────────────────────────────────────────────────────────

type mjMessage struct {
	From     mjContact   `json:"From"`
	To       []mjContact `json:"To"`
	Subject  string      `json:"Subject"`
	HTMLPart string      `json:"HTMLPart"`
	TextPart string      `json:"TextPart"`
}

type mjContact struct {
	Email string `json:"Email"`
	Name  string `json:"Name,omitempty"`
}

type mjPayload struct {
	Messages []mjMessage `json:"Messages"`
}

func (m *Mailer) sendMailjet(msg mjMessage) error {
	body, _ := json.Marshal(mjPayload{Messages: []mjMessage{msg}})
	req, err := http.NewRequest(http.MethodPost, mailjetURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(m.apiKey, m.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailjet request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mailjet responded with status %d", resp.StatusCode)
	}
	return nil
}

// ── API publique ─────────────────────────────────────────────────────────────

func (m *Mailer) SendPasswordReset(toEmail, toName, resetLink string) error {
	html := passwordResetHTML(toName, resetLink)
	text := fmt.Sprintf("Cliquez sur ce lien pour réinitialiser votre mot de passe (valable 15 min) :\n%s", resetLink)
	return m.send("Réinitialisation de votre mot de passe — O'Z", html, text, toEmail, toName)
}

func (m *Mailer) SendLowStockAlert(toEmail, toName string, items []LowStockItem) error {
	if len(items) == 0 {
		return nil
	}
	html := lowStockHTML(toName, items)
	return m.send(
		fmt.Sprintf("⚠️ %d alerte(s) de stock faible — O'Z", len(items)),
		html, lowStockText(toName, items),
		toEmail, toName,
	)
}

func (m *Mailer) SendPendingOrdersDigest(toEmail, toName string, orders []PendingOrderItem) error {
	if len(orders) == 0 {
		return nil
	}
	html := pendingOrdersHTML(toName, orders)
	return m.send(
		fmt.Sprintf("🕐 %d commande(s) en attente depuis +24h — O'Z", len(orders)),
		html, fmt.Sprintf("%d commande(s) en attente de traitement depuis plus de 24h.", len(orders)),
		toEmail, toName,
	)
}

// LowStockItem représente un produit en stock faible pour les alertes email.
type LowStockItem struct {
	ProductName string
	ShopName    string
	Quantity    int
	MinStock    int
}

// PendingOrderItem représente une commande en attente pour les alertes email.
type PendingOrderItem struct {
	UUID         string
	ShopName     string
	HoursPending int
	TotalAmount  int64
}

// ────────────────────────── Templates HTML ──────────────────────────

func baseHTML(title, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<style>
  body { font-family: -apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif; background:#f4f4f5; margin:0; padding:24px; color:#18181b; }
  .card { background:#fff; border-radius:12px; max-width:560px; margin:0 auto; padding:32px; box-shadow:0 1px 4px rgba(0,0,0,.08); }
  .logo { font-size:22px; font-weight:700; color:#4f46e5; margin-bottom:24px; }
  h1 { font-size:20px; font-weight:600; margin:0 0 12px; }
  p { font-size:14px; line-height:1.6; color:#52525b; margin:8px 0; }
  .btn { display:inline-block; background:#4f46e5; color:#fff!important; text-decoration:none; padding:12px 28px; border-radius:8px; font-weight:600; font-size:14px; margin:20px 0; }
  table { width:100%%; border-collapse:collapse; margin-top:16px; font-size:13px; }
  th { background:#f4f4f5; padding:8px 12px; text-align:left; font-weight:600; color:#71717a; }
  td { padding:8px 12px; border-bottom:1px solid #f4f4f5; }
  .badge-red { background:#fee2e2; color:#dc2626; padding:2px 8px; border-radius:99px; font-size:12px; font-weight:600; }
  .footer { text-align:center; font-size:12px; color:#a1a1aa; margin-top:24px; }
</style>
</head>
<body><div class="card">
<div class="logo">O'Z</div>
%s
<div class="footer">O'Z Backoffice — Ne répondez pas à cet email.</div>
</div></body></html>`, title, content)
}

func passwordResetHTML(name, link string) string {
	displayName := name
	if displayName == "" {
		displayName = "Utilisateur"
	}
	content := fmt.Sprintf(`
<h1>Réinitialisation de mot de passe</h1>
<p>Bonjour <strong>%s</strong>,</p>
<p>Vous avez demandé la réinitialisation de votre mot de passe sur O'Z Backoffice.</p>
<p>Cliquez sur le bouton ci-dessous pour choisir un nouveau mot de passe. Ce lien est valable <strong>15 minutes</strong>.</p>
<a class="btn" href="%s">Réinitialiser mon mot de passe</a>
<p>Si vous n'avez pas fait cette demande, ignorez cet email — votre mot de passe reste inchangé.</p>
<p style="font-size:12px;color:#a1a1aa;word-break:break-all;">Lien direct : %s</p>`, displayName, link, link)
	return baseHTML("Réinitialisation de mot de passe", content)
}

func lowStockHTML(name string, items []LowStockItem) string {
	rows := ""
	for _, it := range items {
		rows += fmt.Sprintf(`<tr>
			<td>%s</td>
			<td>%s</td>
			<td><span class="badge-red">%d / %d</span></td>
		</tr>`, it.ProductName, it.ShopName, it.Quantity, it.MinStock)
	}
	content := fmt.Sprintf(`
<h1>⚠️ Alertes stock faible</h1>
<p>Bonjour <strong>%s</strong>,</p>
<p>%d produit(s) ont atteint ou dépassé leur seuil de stock minimum :</p>
<table>
  <tr><th>Produit</th><th>Boutique</th><th>Qté / Seuil</th></tr>
  %s
</table>
<p style="margin-top:16px">Connectez-vous à votre backoffice pour faire une entrée de stock.</p>`, name, len(items), rows)
	return baseHTML("Alerte stock faible", content)
}

func lowStockText(name string, items []LowStockItem) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Bonjour %s,\n\n%d produit(s) en stock faible :\n\n", name, len(items)))
	for _, it := range items {
		sb.WriteString(fmt.Sprintf("- %s (%s) : %d / %d minimum\n", it.ProductName, it.ShopName, it.Quantity, it.MinStock))
	}
	return sb.String()
}

func pendingOrdersHTML(name string, orders []PendingOrderItem) string {
	rows := ""
	for _, o := range orders {
		rows += fmt.Sprintf(`<tr>
			<td>#%s</td>
			<td>%s</td>
			<td><span class="badge-red">%dh</span></td>
		</tr>`, o.UUID, o.ShopName, o.HoursPending)
	}
	content := fmt.Sprintf(`
<h1>🕐 Commandes en attente</h1>
<p>Bonjour <strong>%s</strong>,</p>
<p>%d commande(s) sont en attente depuis plus de 24h :</p>
<table>
  <tr><th>Référence</th><th>Boutique</th><th>Attente</th></tr>
  %s
</table>
<p style="margin-top:16px">Connectez-vous pour traiter ces commandes.</p>`, name, len(orders), rows)
	return baseHTML("Commandes en attente", content)
}
