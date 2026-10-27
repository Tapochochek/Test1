// Package emailcode отправляет одноразовые коды подтверждения на почту (SMTP).
// Если SMTP не настроен, код печатается в лог сервера (удобно для демо).
package emailcode

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Host, Port, User, Password, From string
}

type entry struct {
	code     string
	expires  time.Time
	sentAt   time.Time
	attempts int
}

type Service struct {
	cfg   Config
	mu    sync.Mutex
	codes map[string]entry
}

func New(cfg Config) *Service {
	return &Service{cfg: cfg, codes: map[string]entry{}}
}

func (s *Service) configured() bool { return strings.TrimSpace(s.cfg.Host) != "" }

// Send генерирует 6-значный код и отправляет его на email.
// Возвращает delivered=false, если SMTP не настроен и код записан только в лог.
func (s *Service) Send(email string) (delivered bool, err error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return false, errors.New("Укажите корректный email")
	}
	s.mu.Lock()
	if old, ok := s.codes[email]; ok && time.Since(old.sentAt) < 60*time.Second {
		s.mu.Unlock()
		return false, errors.New("Код уже отправлен, повторить можно через минуту")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		s.mu.Unlock()
		return false, err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	s.codes[email] = entry{code: code, expires: time.Now().Add(10 * time.Minute), sentAt: time.Now()}
	s.mu.Unlock()

	if !s.configured() {
		log.Printf("[emailcode] SMTP не настроен. Код для %s: %s", email, code)
		return false, nil
	}
	if err := s.send(email, code); err != nil {
		log.Printf("[emailcode] ошибка отправки на %s: %v", email, err)
		return false, errors.New("Не удалось отправить письмо: проверьте SMTP-настройки и логи api")
	}
	return true, nil
}

// Verify проверяет код; успешный код одноразовый.
func (s *Service) Verify(email, code string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.codes[email]
	if !ok || time.Now().After(e.expires) || e.attempts >= 5 {
		return false
	}
	if strings.TrimSpace(code) != e.code {
		e.attempts++
		s.codes[email] = e
		return false
	}
	delete(s.codes, email)
	return true
}

func (s *Service) send(to, code string) error {
	from := s.cfg.From
	if from == "" {
		from = s.cfg.User
	}
	port := s.cfg.Port
	if port == "" {
		port = "587"
	}
	addr := net.JoinHostPort(s.cfg.Host, port)
	msg := "From: " + from + "\r\nTo: " + to + "\r\nSubject: =?UTF-8?B?" + b64("Код подтверждения ФСП 56") + "?=\r\n" +
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" +
		"Ваш код подтверждения: " + code + "\r\nКод действует 10 минут.\r\n"
	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	}
	if port != "465" { // STARTTLS (587/25) делает smtp.SendMail сам
		return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	}
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.cfg.Host})
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}
	defer c.Close()
	if auth != nil {
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
}

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
