// Package handler is the Vercel serverless function behind POST /api/apply.
//
// It takes a careers application (multipart/form-data: name, phone, email,
// cv, plus the _honey honeypot), validates it, and emails it with the CV
// attached through Microsoft Graph sendMail, authenticating app-only with
// client credentials. Standard library only.
//
// Env: MS_TENANT_ID, MS_CLIENT_ID, MS_CLIENT_SECRET, MAIL_FROM (default
// info@plumcut.com), MAIL_TO (default info@plumcut.com).
//
// Logs carry outcomes only. Never log the CV, a name, an email or a phone,
// and never a Graph response body, which can echo addresses back.
package handler

import (
	"bytes"
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
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	// Graph's single-request sendMail caps the whole request at about 4MB and
	// base64 grows the file by a third, so 3,000,000 bytes (not 3MiB, which
	// encodes to exactly 4MiB) leaves room for the rest of the message.
	maxCV      = 3_000_000
	maxBody    = maxCV + 256<<10  // CV plus room for the text fields and multipart framing
	rateLimit  = 5                // submissions per IP per window
	rateWindow = 10 * time.Minute //
	tokenSkew  = 5 * time.Minute  // refresh the token this long before it expires
	subject    = "New application: Sales Representative"
	fallback   = "Something went wrong sending your application. Please try again"
	tooBigMsg  = "Your CV is over 3MB. Please upload a PDF of 3MB or less."
)

var (
	phoneRe = regexp.MustCompile(`^\+?[0-9\s().-]+$`)
	digitRe = regexp.MustCompile(`[0-9]`)

	// Endpoints and clock are variables so tests can point them at fakes.
	tokenURL = func(tenant string) string {
		return "https://login.microsoftonline.com/" + url.PathEscape(tenant) + "/oauth2/v2.0/token"
	}
	graphBase  = "https://graph.microsoft.com/v1.0"
	httpClient = &http.Client{Timeout: 15 * time.Second}
	now        = time.Now
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
// applicant types can break a subject line or the plain-text body layout.
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
		return nil, &fieldError{"cv", tooBigMsg}
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
		return nil, &fieldError{"cv", tooBigMsg}
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

/* ----------------------------------------------------------------- graph */

type config struct{ Tenant, ClientID, Secret, From, To string }

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func loadConfig() config {
	return config{
		Tenant:   env("MS_TENANT_ID", ""),
		ClientID: env("MS_CLIENT_ID", ""),
		Secret:   os.Getenv("MS_CLIENT_SECRET"),
		From:     env("MAIL_FROM", "info@plumcut.com"),
		To:       env("MAIL_TO", "info@plumcut.com"),
	}
}

// stepError names the step that failed and the HTTP status, nothing else,
// so it is always safe to log.
type stepError struct {
	Step   string
	Status int
}

func (e *stepError) Error() string {
	if e.Status == 0 {
		return e.Step
	}
	return fmt.Sprintf("%s status %d", e.Step, e.Status)
}

// The app-only token is cached per instance until shortly before it expires.
var token = struct {
	sync.Mutex
	key    string // tenant + client, so a config change never reuses a token
	value  string
	expiry time.Time
}{}

func dropToken() {
	token.Lock()
	token.value = ""
	token.Unlock()
}

func getToken(c config) (string, error) {
	token.Lock()
	defer token.Unlock()
	key := c.Tenant + "|" + c.ClientID
	if token.value != "" && token.key == key && now().Before(token.expiry) {
		return token.value, nil
	}
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.ClientID},
		"client_secret": {c.Secret},
		"scope":         {"https://graph.microsoft.com/.default"},
	}
	res, err := httpClient.PostForm(tokenURL(c.Tenant), form)
	if err != nil {
		return "", &stepError{Step: "token request"}
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
		return "", &stepError{Step: "token", Status: res.StatusCode}
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&out); err != nil || out.AccessToken == "" {
		return "", &stepError{Step: "token decode", Status: res.StatusCode}
	}
	token.key, token.value = key, out.AccessToken
	token.expiry = now().Add(time.Duration(out.ExpiresIn)*time.Second - tokenSkew)
	return token.value, nil
}

type emailAddress struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
}

type recipient struct {
	EmailAddress emailAddress `json:"emailAddress"`
}

type attachment struct {
	ODataType    string `json:"@odata.type"`
	Name         string `json:"name"`
	ContentType  string `json:"contentType"`
	ContentBytes string `json:"contentBytes"`
}

type message struct {
	Subject      string       `json:"subject"`
	Body         messageBody  `json:"body"`
	ToRecipients []recipient  `json:"toRecipients"`
	ReplyTo      []recipient  `json:"replyTo"`
	Attachments  []attachment `json:"attachments"`
}

type messageBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type sendMailRequest struct {
	Message         message `json:"message"`
	SaveToSentItems bool    `json:"saveToSentItems"`
}

func buildRequest(c config, a *application) sendMailRequest {
	body := "New application for Sales Representative (Commission-Based, Remote).\n\n" +
		"Name:  " + a.Name + "\n" +
		"Phone: " + a.Phone + "\n" +
		"Email: " + a.Email + "\n" +
		"Date:  " + now().UTC().Format("2 January 2006, 15:04 MST") + "\n\n" +
		"CV attached. Reply to this email to answer the applicant.\n"
	return sendMailRequest{
		Message: message{
			Subject:      subject + " - " + a.Name,
			Body:         messageBody{ContentType: "Text", Content: body},
			ToRecipients: []recipient{{emailAddress{Address: c.To}}},
			ReplyTo:      []recipient{{emailAddress{Name: a.Name, Address: a.Email}}},
			Attachments: []attachment{{
				ODataType:    "#microsoft.graph.fileAttachment",
				Name:         a.CVName,
				ContentType:  "application/pdf",
				ContentBytes: base64.StdEncoding.EncodeToString(a.CV),
			}},
		},
		SaveToSentItems: false,
	}
}

func send(c config, a *application) error {
	if c.Tenant == "" || c.ClientID == "" || c.Secret == "" {
		return &stepError{Step: "config (MS_TENANT_ID, MS_CLIENT_ID or MS_CLIENT_SECRET not set)"}
	}
	tok, err := getToken(c)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(buildRequest(c, a))
	if err != nil {
		return &stepError{Step: "encode"}
	}
	req, err := http.NewRequest(http.MethodPost, graphBase+"/users/"+url.PathEscape(c.From)+"/sendMail", bytes.NewReader(payload))
	if err != nil {
		return &stepError{Step: "sendMail request"}
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return &stepError{Step: "sendMail request"}
	}
	defer res.Body.Close()
	io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
	if res.StatusCode == http.StatusUnauthorized {
		dropToken() // the next request fetches a fresh token
	}
	if res.StatusCode != http.StatusAccepted {
		return &stepError{Step: "sendMail", Status: res.StatusCode}
	}
	return nil
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
			reply(w, http.StatusRequestEntityTooLarge, "cv", tooBigMsg)
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
		// Only stepErrors reach here and they carry no PII by construction.
		log.Printf("apply: graph error at %v", err)
		reply(w, http.StatusBadGateway, "", fallback)
		return
	}
	log.Print("apply: sent")
	reply(w, http.StatusOK, "", "")
}
