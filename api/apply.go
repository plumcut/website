// Package handler is the Vercel serverless function behind POST /api/apply.
//
// It takes a careers application (multipart/form-data: name, phone, email,
// cv, plus the _honey honeypot), validates it, and emails it with the CV
// attached over SMTP with STARTTLS. Standard library only.
//
// Env: SMTP_HOST (default smtp.office365.com), SMTP_PORT (587), SMTP_USER,
// SMTP_PASS, MAIL_TO (info@plumcut.com).
//
// Logs carry outcomes only. Never log the CV, a name, an email or a phone.
package handler

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	maxCV      = 4 << 20          // 4MB, under Vercel's 4.5MB request body limit
	maxBody    = maxCV + 256<<10  // CV plus room for the text fields and multipart framing
	rateLimit  = 5                // submissions per IP per window
	rateWindow = 10 * time.Minute //
	smtpWait   = 8 * time.Second  // dial timeout; Vercel functions have a short budget
	subject    = "New application: Sales Representative"
	fallback   = "Something went wrong sending your application. Please try again"
)

var (
	phoneRe = regexp.MustCompile(`^\+?[0-9\s().-]+$`)
	digitRe = regexp.MustCompile(`[0-9]`)

	// tlsConfig is a variable so tests can trust a local server's certificate.
	tlsConfig = func(host string) *tls.Config { return &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12} }
	now       = time.Now
)

/* ------------------------------------------------------------ rate limit */

// Best effort only: the counts live in this instance's memory, so a cold
// start or a second instance starts from zero.
var hits = struct {
	sync.Mutex
	m map[string][]time.Time
}{m: map[string][]time.Time{}}

func allow(ip string) bool {
	hits.Lock()
	defer hits.Unlock()
	t := now()
	recent := hits.m[ip][:0]
	for _, h := range hits.m[ip] {
		if t.Sub(h) < rateWindow {
			recent = append(recent, h)
		}
	}
	if len(recent) >= rateLimit {
		hits.m[ip] = recent
		return false
	}
	hits.m[ip] = append(recent, t)
	if len(hits.m) > 10000 { // keep a long-lived instance from growing without bound
		for k, v := range hits.m {
			if len(v) == 0 || t.Sub(v[len(v)-1]) >= rateWindow {
				delete(hits.m, k)
			}
		}
	}
	return true
}

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-Ip")); ip != "" {
		return ip
	}
	if f := r.Header.Get("X-Forwarded-For"); f != "" {
		return strings.TrimSpace(strings.Split(f, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

/* ------------------------------------------------------------ validation */

type application struct {
	Name, Phone, Email string
	CVName             string
	CV                 []byte
}

type fieldError struct{ Field, Msg string }

func (e *fieldError) Error() string { return e.Msg }

// clean trims a text field and drops control characters, so nothing an
// applicant types can break out of an email header.
func clean(s string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s))
}

func validate(name, phone, email string, file multipart.File, hdr *multipart.FileHeader) (*application, error) {
	a := &application{Name: clean(name), Phone: clean(phone), Email: clean(email)}
	switch {
	case a.Name == "":
		return nil, &fieldError{"name", "Please enter your full name."}
	case len(a.Name) < 2 || len(a.Name) > 120:
		return nil, &fieldError{"name", "Please enter your full name."}
	case a.Phone == "":
		return nil, &fieldError{"phone", "Please enter your phone number."}
	case !phoneRe.MatchString(a.Phone) || len(digitRe.FindAllString(a.Phone, -1)) < 7 || len(digitRe.FindAllString(a.Phone, -1)) > 15:
		return nil, &fieldError{"phone", "Please enter a valid phone number, for example +961 70 123 456."}
	case a.Email == "":
		return nil, &fieldError{"email", "Please enter your email address."}
	}
	if addr, err := mail.ParseAddress(a.Email); err != nil || addr.Address != a.Email || !strings.Contains(a.Email[strings.LastIndex(a.Email, "@")+1:], ".") || len(a.Email) > 254 {
		return nil, &fieldError{"email", "Please enter a valid email address, for example name@example.com."}
	}

	if file == nil || hdr == nil {
		return nil, &fieldError{"cv", "Please attach your CV as a PDF."}
	}
	if hdr.Size > maxCV {
		return nil, &fieldError{"cv", "Your CV is over 4MB. Please upload a PDF of 4MB or less."}
	}
	ct, _, _ := mime.ParseMediaType(hdr.Header.Get("Content-Type"))
	if ct != "application/pdf" && ct != "application/x-pdf" {
		return nil, &fieldError{"cv", "Your CV must be a PDF file."}
	}
	data, err := io.ReadAll(io.LimitReader(file, maxCV+1))
	if err != nil {
		return nil, &fieldError{"cv", "We could not read your CV. Please try again."}
	}
	if len(data) > maxCV {
		return nil, &fieldError{"cv", "Your CV is over 4MB. Please upload a PDF of 4MB or less."}
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		return nil, &fieldError{"cv", "Your CV must be a PDF file."}
	}
	a.CV = data
	a.CVName = safeFilename(hdr.Filename)
	return a, nil
}

func safeFilename(s string) string {
	s = s[strings.LastIndexAny(s, `/\`)+1:]
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || strings.ContainsRune(`"\;`, r) {
			return '_'
		}
		return r
	}, s)
	if !strings.HasSuffix(strings.ToLower(s), ".pdf") {
		s += ".pdf"
	}
	if len(s) > 120 || s == ".pdf" {
		s = "cv.pdf"
	}
	return s
}

/* ------------------------------------------------------------------ mail */

type config struct{ Host, Port, User, Pass, To string }

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func loadConfig() config {
	return config{
		Host: env("SMTP_HOST", "smtp.office365.com"),
		Port: env("SMTP_PORT", "587"),
		User: env("SMTP_USER", ""),
		Pass: os.Getenv("SMTP_PASS"),
		To:   env("MAIL_TO", "info@plumcut.com"),
	}
}

func wrap76(s string) string {
	var b strings.Builder
	for len(s) > 76 {
		b.WriteString(s[:76] + "\r\n")
		s = s[76:]
	}
	b.WriteString(s + "\r\n")
	return b.String()
}

func buildMessage(c config, a *application) []byte {
	from := c.User
	if from == "" {
		from = c.To
	}
	boundary := "plumcut-" + fmt.Sprint(now().UnixNano())
	date := now().UTC()
	var b bytes.Buffer
	h := func(k, v string) { fmt.Fprintf(&b, "%s: %s\r\n", k, v) }
	h("From", (&mail.Address{Name: "plumcut careers", Address: from}).String())
	h("To", c.To)
	h("Reply-To", (&mail.Address{Name: a.Name, Address: a.Email}).String())
	h("Subject", mime.QEncoding.Encode("utf-8", subject+" - "+a.Name))
	h("Date", date.Format(time.RFC1123Z))
	h("MIME-Version", "1.0")
	h("Content-Type", `multipart/mixed; boundary="`+boundary+`"`)
	b.WriteString("\r\n")

	body := "New application for Sales Representative (Commission-Based, Remote).\r\n\r\n" +
		"Name:  " + a.Name + "\r\n" +
		"Phone: " + a.Phone + "\r\n" +
		"Email: " + a.Email + "\r\n" +
		"Date:  " + date.Format("2 January 2006, 15:04 MST") + "\r\n\r\n" +
		"CV attached. Reply to this email to answer the applicant.\r\n"
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n", boundary)
	b.WriteString(wrap76(base64.StdEncoding.EncodeToString([]byte(body))))

	fname := mime.QEncoding.Encode("utf-8", a.CVName)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: application/pdf; name=\"%s\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n", boundary, fname, fname)
	b.WriteString(wrap76(base64.StdEncoding.EncodeToString(a.CV)))
	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return b.Bytes()
}

// loginAuth is SMTP AUTH LOGIN, which Office 365 offers and net/smtp lacks.
type loginAuth struct{ user, pass string }

func (l loginAuth) Start(s *smtp.ServerInfo) (string, []byte, error) {
	if !s.TLS {
		return "", nil, errors.New("refusing AUTH LOGIN without TLS")
	}
	return "LOGIN", nil, nil
}

func (l loginAuth) Next(from []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(string(from))) {
	case "username:":
		return []byte(l.user), nil
	case "password:":
		return []byte(l.pass), nil
	}
	return nil, errors.New("unexpected AUTH LOGIN challenge")
}

func send(c config, a *application) error {
	if c.User == "" || c.Pass == "" {
		return errors.New("SMTP_USER or SMTP_PASS not set")
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.Host, c.Port), smtpWait)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	_ = conn.SetDeadline(now().Add(25 * time.Second))
	cl, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("greeting: %w", err)
	}
	defer cl.Close()
	if ok, _ := cl.Extension("STARTTLS"); !ok {
		return errors.New("server does not offer STARTTLS")
	}
	if err := cl.StartTLS(tlsConfig(c.Host)); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}
	var auth smtp.Auth = loginAuth{c.User, c.Pass}
	if _, mechs := cl.Extension("AUTH"); strings.Contains(" "+strings.ToUpper(mechs)+" ", " PLAIN ") {
		auth = smtp.PlainAuth("", c.User, c.Pass, c.Host)
	}
	if err := cl.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	from := c.User
	if err := cl.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := cl.Rcpt(c.To); err != nil {
		return fmt.Errorf("rcpt: %w", err)
	}
	w, err := cl.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(buildMessage(c, a)); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("data close: %w", err)
	}
	return cl.Quit()
}

/* --------------------------------------------------------------- handler */

func reply(w http.ResponseWriter, status int, field, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	out := map[string]any{"ok": status == http.StatusOK}
	if msg != "" {
		out["error"] = msg
	}
	if field != "" {
		out["field"] = field
	}
	_ = json.NewEncoder(w).Encode(out)
}

// sendMail is a variable so tests can stub the network.
var sendMail = send

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		reply(w, http.StatusMethodNotAllowed, "", "Use POST.")
		return
	}
	if !allow(clientIP(r)) {
		log.Print("apply: rejected: rate limited")
		reply(w, http.StatusTooManyRequests, "", "Too many attempts. Please wait a few minutes and try again")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			log.Print("apply: rejected: body too large")
			reply(w, http.StatusRequestEntityTooLarge, "cv", "Your CV is over 4MB. Please upload a PDF of 4MB or less.")
			return
		}
		log.Print("apply: rejected: bad form")
		reply(w, http.StatusBadRequest, "", "We could not read the form. Please try again")
		return
	}
	defer r.MultipartForm.RemoveAll()

	if strings.TrimSpace(r.FormValue("_honey")) != "" {
		log.Print("apply: rejected: honeypot")
		reply(w, http.StatusBadRequest, "", "Your application could not be submitted")
		return
	}

	file, hdr, err := r.FormFile("cv")
	if err != nil {
		file, hdr = nil, nil
	} else {
		defer file.Close()
	}
	a, err := validate(r.FormValue("name"), r.FormValue("phone"), r.FormValue("email"), file, hdr)
	if err != nil {
		var fe *fieldError
		errors.As(err, &fe)
		log.Printf("apply: rejected: invalid %s", fe.Field)
		reply(w, http.StatusBadRequest, fe.Field, fe.Msg)
		return
	}

	if err := sendMail(loadConfig(), a); err != nil {
		// SMTP errors can echo addresses back, so log only the stage.
		stage := err.Error()
		if i := strings.Index(stage, ":"); i > 0 {
			stage = stage[:i]
		}
		log.Printf("apply: smtp error at %s", stage)
		reply(w, http.StatusBadGateway, "", fallback)
		return
	}
	log.Print("apply: sent")
	reply(w, http.StatusOK, "", "")
}
