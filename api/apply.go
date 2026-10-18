// Package handler is the Vercel serverless function behind /api/apply.
//
//	GET  /api/apply  returns a signed form token (the time trap, see below).
//	POST /api/apply  takes a careers application as multipart/form-data
//	                 (name, phone, email, cv, _honey, _t), validates it and
//	                 emails it with the CV attached through Microsoft Graph
//	                 sendMail, authenticating app-only with client credentials.
//
// Standard library only.
//
// Env: MS_TENANT_ID, MS_CLIENT_ID, MS_CLIENT_SECRET, FORM_SIGNING_KEY (at
// least 32 characters of random data), MAIL_FROM and MAIL_TO (both default
// to info@plumcut.com). Vercel sets VERCEL_URL, VERCEL_BRANCH_URL and
// VERCEL_PROJECT_PRODUCTION_URL, which extend the allowed origins.
//
// Logs carry outcomes only. Never log the CV, a name, an email or a phone,
// and never a Graph response body, which can echo addresses back. Clients
// get plain field messages for things they can fix and one generic message
// for everything else.
package handler

import (
	"bytes"
	"compress/zlib"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// Graph's single-request sendMail caps the whole request at about 4MB and
	// base64 grows the file by a third, so 3,000,000 bytes (not 3MiB, which
	// encodes to exactly 4MiB) leaves room for the rest of the message.
	maxCV   = 3_000_000
	maxBody = maxCV + 64<<10 // CV plus room for the text fields and multipart framing

	ipLimit     = 5                // submissions per IP per window
	globalLimit = 30               // submissions per instance per window
	tokenLimit  = 30               // token requests per IP per window
	rateWindow  = 10 * time.Minute //

	tokenMinAge = 3 * time.Second // the time trap: humans take longer than this
	tokenMaxAge = 2 * time.Hour
	tokenSkew   = 5 * time.Minute // refresh the Graph token this long before it expires
	readTimeout = 20 * time.Second

	maxInflate = 20 << 20 // total bytes the PDF scan may decompress

	subject   = "New application: Sales Representative"
	fallback  = "Something went wrong sending your application. Please try again"
	generic   = "Your application could not be submitted. Please try again"
	tooBigMsg = "Your CV is over 3MB. Please upload a PDF of 3MB or less."
	notPDFMsg = "Your CV must be a PDF file."
	badPDFMsg = "We couldn't accept this PDF. Please export it as a standard PDF and try again."
)

var (
	nameRe  = regexp.MustCompile(`^[\p{L}\p{M}][\p{L}\p{M} '.\-]*$`)
	phoneRe = regexp.MustCompile(`^\+?[0-9][0-9 ()\-]*$`)
	digitRe = regexp.MustCompile(`[0-9]`)
	emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]{1,64}@[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?)*\.[A-Za-z]{2,24}$`)

	textFields = map[string]bool{"name": true, "phone": true, "email": true, "_honey": true, "_t": true}

	// Endpoints and clock are variables so tests can point them at fakes.
	tokenURL = func(tenant string) string {
		return "https://login.microsoftonline.com/" + url.PathEscape(tenant) + "/oauth2/v2.0/token"
	}
	graphBase  = "https://graph.microsoft.com/v1.0"
	httpClient = &http.Client{Timeout: 15 * time.Second}
	now        = time.Now
)

/* ------------------------------------------------------------ rate limit */

// Best effort only: counts live in this instance's memory, so a cold start
// or a second instance starts from zero.
type limiter struct {
	sync.Mutex
	limit int
	m     map[string][]time.Time
}

func newLimiter(limit int) *limiter { return &limiter{limit: limit, m: map[string][]time.Time{}} }

func (l *limiter) allow(key string) bool {
	l.Lock()
	defer l.Unlock()
	t := now()
	recent := l.m[key][:0]
	for _, h := range l.m[key] {
		if t.Sub(h) < rateWindow {
			recent = append(recent, h)
		}
	}
	if len(recent) >= l.limit {
		l.m[key] = recent
		return false
	}
	l.m[key] = append(recent, t)
	if len(l.m) > 10000 { // keep a long-lived instance from growing without bound
		for k, v := range l.m {
			if len(v) == 0 || t.Sub(v[len(v)-1]) >= rateWindow {
				delete(l.m, k)
			}
		}
	}
	return true
}

var (
	perIP     = newLimiter(ipLimit)
	global    = newLimiter(globalLimit)
	tokenByIP = newLimiter(tokenLimit)
)

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

/* ---------------------------------------------------------------- origin */

func allowedOrigin(origin string) bool {
	if origin == "https://plumcut.com" || origin == "https://www.plumcut.com" {
		return true
	}
	for _, k := range []string{"VERCEL_URL", "VERCEL_BRANCH_URL", "VERCEL_PROJECT_PRODUCTION_URL"} {
		if h := strings.TrimSpace(os.Getenv(k)); h != "" && origin == "https://"+h {
			return true
		}
	}
	return false
}

// sameSite rejects anything a browser marks as cross-site. Browsers that
// send no Sec-Fetch-Site still have to pass the Origin check on POST.
func sameSite(r *http.Request) bool {
	s := r.Header.Get("Sec-Fetch-Site")
	return s == "" || s == "same-origin"
}

/* ----------------------------------------------------------------- token */

// A form token is v1.<unix seconds>.<16 random bytes>.<HMAC-SHA256>, keyed
// with FORM_SIGNING_KEY. It proves the form was loaded from our page (basic
// CSRF protection on top of the Origin check) and how long ago (the time
// trap). Each token is single use per warm instance, and only counts as used
// once the email has actually gone out, so a retry after a failure works.

var (
	errTokenInvalid = errors.New("invalid")
	errTokenEarly   = errors.New("early")
	errTokenExpired = errors.New("expired")
	errTokenReplay  = errors.New("replay")
	errNoKey        = errors.New("FORM_SIGNING_KEY missing or shorter than 32 characters")
)

var usedTokens = struct {
	sync.Mutex
	m map[string]time.Time
}{m: map[string]time.Time{}}

func signingKey() ([]byte, error) {
	k := os.Getenv("FORM_SIGNING_KEY")
	if len(k) < 32 {
		return nil, errNoKey
	}
	return []byte(k), nil
}

func sign(key []byte, payload string) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func makeToken(key []byte) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := "v1." + strconv.FormatInt(now().Unix(), 10) + "." + base64.RawURLEncoding.EncodeToString(nonce)
	return payload + "." + sign(key, payload), nil
}

func checkToken(key []byte, tok string) error {
	if len(tok) > 200 {
		return errTokenInvalid
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 4 || parts[0] != "v1" {
		return errTokenInvalid
	}
	payload := strings.Join(parts[:3], ".")
	if !hmac.Equal([]byte(parts[3]), []byte(sign(key, payload))) {
		return errTokenInvalid
	}
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return errTokenInvalid
	}
	age := now().Sub(time.Unix(ts, 0))
	switch {
	case age < -30*time.Second:
		return errTokenInvalid // issued in the future: not one of ours
	case age < tokenMinAge:
		return errTokenEarly
	case age > tokenMaxAge:
		return errTokenExpired
	}
	usedTokens.Lock()
	defer usedTokens.Unlock()
	if _, seen := usedTokens.m[tok]; seen {
		return errTokenReplay
	}
	return nil
}

func markUsed(tok string) {
	usedTokens.Lock()
	defer usedTokens.Unlock()
	t := now()
	for k, v := range usedTokens.m {
		if t.Sub(v) > tokenMaxAge+time.Minute {
			delete(usedTokens.m, k)
		}
	}
	usedTokens.m[tok] = t
}

/* ------------------------------------------------------------ validation */

type application struct {
	Name, Phone, Email string
	CVName             string
	CV                 []byte
}

type fieldError struct{ Field, Msg, Why string }

func (e *fieldError) Error() string { return e.Msg }

// clean drops control characters, CR/LF, Unicode format characters (bidi
// overrides, zero-width joiners) and line/paragraph separators, collapses
// whitespace and trims. Nothing an applicant types can then break a subject
// line, spoof its direction, or reshape the plain-text body.
func clean(s string) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == ' ' || r == ' ' {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

func validate(name, phone, email string) (*application, error) {
	a := &application{Name: clean(name), Phone: clean(phone), Email: clean(email)}
	nameLen := utf8.RuneCountInString(a.Name)
	switch {
	case a.Name == "":
		return nil, &fieldError{"name", "Please enter your full name.", "name"}
	case nameLen < 2 || nameLen > 100 || !nameRe.MatchString(a.Name):
		return nil, &fieldError{"name", "Please enter your full name using letters only.", "name"}
	}
	digits := len(digitRe.FindAllString(a.Phone, -1))
	switch {
	case a.Phone == "":
		return nil, &fieldError{"phone", "Please enter your phone number.", "phone"}
	case len(a.Phone) > 20 || !phoneRe.MatchString(a.Phone) || digits < 7 || digits > 15:
		return nil, &fieldError{"phone", "Please enter a valid phone number, for example +961 70 123 456.", "phone"}
	}
	local, _, _ := strings.Cut(a.Email, "@")
	switch {
	case a.Email == "":
		return nil, &fieldError{"email", "Please enter your email address.", "email"}
	case len(a.Email) > 254 || !emailRe.MatchString(a.Email) ||
		strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") || strings.Contains(local, ".."):
		return nil, &fieldError{"email", "Please enter a valid email address, for example name@example.com.", "email"}
	}
	return a, nil
}

func readCV(file multipart.File, hdr *multipart.FileHeader) ([]byte, error) {
	if file == nil || hdr == nil {
		return nil, &fieldError{"cv", "Please attach your CV as a PDF.", "cv missing"}
	}
	if hdr.Size > maxCV {
		return nil, &fieldError{"cv", tooBigMsg, "cv size"}
	}
	ct, _, _ := mime.ParseMediaType(hdr.Header.Get("Content-Type"))
	if ct != "application/pdf" && ct != "application/x-pdf" {
		return nil, &fieldError{"cv", notPDFMsg, "cv type"}
	}
	data, err := io.ReadAll(io.LimitReader(file, maxCV+1))
	if err != nil {
		return nil, &fieldError{"cv", generic, "cv read"}
	}
	if len(data) > maxCV {
		return nil, &fieldError{"cv", tooBigMsg, "cv size"}
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return nil, &fieldError{"cv", notPDFMsg, "cv magic"}
	}
	if why := scanPDF(data); why != "" {
		return nil, &fieldError{"cv", badPDFMsg, "cv " + why}
	}
	return data, nil
}

/* ------------------------------------------------------------- pdf scan */

// banned are PDF names that run code, launch programs, carry other files or
// mean the file is encrypted (and so cannot be scanned). Blocking /JS and
// /JavaScript also covers an /OpenAction that runs JavaScript.
var banned = map[string]bool{
	"JavaScript": true, "JS": true, "Launch": true,
	"EmbeddedFile": true, "EmbeddedFiles": true, "Encrypt": true,
}

// A PDF name is / followed by regular characters; #xx is a hex escape, so
// /J#61vaScript is /JavaScript.
var nameTok = regexp.MustCompile(`/[^\x00\t\n\f\r ()<>\[\]{}/%]+`)

func decodeName(raw []byte) string {
	var b strings.Builder
	for i := 1; i < len(raw); i++ {
		if raw[i] == '#' && i+2 < len(raw) {
			if v, err := hex.DecodeString(string(raw[i+1 : i+3])); err == nil {
				b.WriteByte(v[0])
				i += 2
				continue
			}
		}
		b.WriteByte(raw[i])
	}
	return b.String()
}

func bannedName(data []byte) string {
	for _, m := range nameTok.FindAll(data, -1) {
		if n := decodeName(m); banned[n] {
			return n
		}
	}
	return ""
}

func hasName(data []byte, want string) bool {
	for _, m := range nameTok.FindAll(data, -1) {
		if decodeName(m) == want {
			return true
		}
	}
	return false
}

// scanPDF returns "" for an acceptable file, or a short reason safe to log.
// It is a byte scan, not a parser: it reads the raw file plus every
// Flate-compressed object stream (/ObjStm), where dictionaries can hide.
// It is conservative on purpose: names inside string literals count too,
// because skipping strings sequentially can be fooled by an unbalanced "(" in
// a crafted file. A CV whose uncompressed text literally contains "/JS" may
// be refused; the page then offers the email fallback.
func scanPDF(data []byte) string {
	tail := data
	if len(tail) > 1024 {
		tail = tail[len(tail)-1024:]
	}
	if !bytes.Contains(tail, []byte("%%EOF")) {
		return "eof"
	}
	if n := bannedName(data); n != "" {
		return "name " + n
	}
	budget := maxInflate
	for pos := 0; ; {
		i := bytes.Index(data[pos:], []byte("stream"))
		if i < 0 {
			return ""
		}
		i += pos
		pos = i + 6
		if i >= 3 && string(data[i-3:i]) == "end" {
			continue // endstream
		}
		start := i + 6
		if start < len(data) && data[start] == '\r' {
			start++
		}
		if start < len(data) && data[start] == '\n' {
			start++
		}
		end := bytes.Index(data[start:], []byte("endstream"))
		if end < 0 {
			return ""
		}
		end += start
		pos = end + 9
		dictStart := i - 2048
		if dictStart < 0 {
			dictStart = 0
		}
		dict := data[dictStart:i]
		if o := bytes.LastIndex(dict, []byte("obj")); o >= 0 {
			dict = dict[o:]
		}
		if !hasName(dict, "ObjStm") || !hasName(dict, "FlateDecode") {
			continue
		}
		zr, err := zlib.NewReader(bytes.NewReader(data[start:end]))
		if err != nil {
			continue
		}
		plain, err := io.ReadAll(io.LimitReader(zr, int64(budget)+1))
		zr.Close()
		if len(plain) > budget {
			return "inflate limit"
		}
		budget -= len(plain)
		if err != nil && len(plain) == 0 {
			continue
		}
		if n := bannedName(plain); n != "" {
			return "objstm name " + n
		}
	}
}

// cvFilename ignores whatever the applicant's file was called.
func cvFilename(name string) string {
	keep := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.Is(unicode.M, r) || r == ' ' || r == '-' || r == '\'' {
			return r
		}
		return -1
	}, name)
	keep = strings.Join(strings.Fields(keep), " ")
	if r := []rune(keep); len(r) > 80 {
		keep = strings.TrimSpace(string(r[:80]))
	}
	if keep == "" {
		return "CV.pdf"
	}
	return "CV - " + keep + ".pdf"
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

// buildRequest puts applicant values only in the plain-text body and the
// subject (both already cleaned) and the validated address in Reply-To.
// Reply-To carries no display name, so nothing free-form reaches a header.
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
			ReplyTo:      []recipient{{emailAddress{Address: a.Email}}},
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

type response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	Field string `json:"field,omitempty"`
	Code  string `json:"code,omitempty"` // "retry": fetch a fresh token and resend
	Token string `json:"token,omitempty"`
}

func reply(w http.ResponseWriter, status int, r response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(r)
}

func reject(w http.ResponseWriter, status int, why string) {
	log.Printf("apply: rejected: %s", why)
	reply(w, status, response{Error: generic})
}

// sendMail is a variable so tests can stub the network.
var sendMail = send

func Handler(w http.ResponseWriter, r *http.Request) {
	// vercel.json sets these too; repeating them keeps the function safe if
	// it is ever served without that config. No CORS headers, ever.
	h := w.Header()
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

	switch r.Method {
	case http.MethodGet:
		issueToken(w, r)
	case http.MethodPost:
		submit(w, r)
	default:
		h.Set("Allow", "GET, POST")
		reply(w, http.StatusMethodNotAllowed, response{Error: generic})
	}
}

func issueToken(w http.ResponseWriter, r *http.Request) {
	if !sameSite(r) {
		reject(w, http.StatusForbidden, "token cross-site")
		return
	}
	if !tokenByIP.allow(clientIP(r)) {
		reject(w, http.StatusTooManyRequests, "token rate limited")
		return
	}
	key, err := signingKey()
	if err != nil {
		log.Printf("apply: config error: %v", err)
		reply(w, http.StatusServiceUnavailable, response{Error: generic})
		return
	}
	tok, err := makeToken(key)
	if err != nil {
		log.Print("apply: token error: rand")
		reply(w, http.StatusServiceUnavailable, response{Error: generic})
		return
	}
	reply(w, http.StatusOK, response{OK: true, Token: tok})
}

func submit(w http.ResponseWriter, r *http.Request) {
	if !allowedOrigin(r.Header.Get("Origin")) || !sameSite(r) {
		reject(w, http.StatusForbidden, "origin")
		return
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "multipart/form-data" {
		reject(w, http.StatusUnsupportedMediaType, "content-type")
		return
	}
	if !perIP.allow(clientIP(r)) {
		log.Print("apply: rejected: rate limited ip")
		reply(w, http.StatusTooManyRequests, response{Error: "Too many attempts. Please wait a few minutes and try again"})
		return
	}
	if !global.allow("") {
		log.Print("apply: rejected: rate limited global")
		reply(w, http.StatusTooManyRequests, response{Error: "Too many attempts. Please wait a few minutes and try again"})
		return
	}
	key, err := signingKey()
	if err != nil {
		log.Printf("apply: config error: %v", err)
		reply(w, http.StatusServiceUnavailable, response{Error: fallback})
		return
	}

	// Vercel buffers the body before the function runs, so this matters most
	// when the handler is served by a plain net/http server; where the
	// runtime does not support deadlines the call is a no-op.
	_ = http.NewResponseController(w).SetReadDeadline(now().Add(readTimeout))
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := r.ParseMultipartForm(maxBody); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			log.Print("apply: rejected: body too large")
			reply(w, http.StatusRequestEntityTooLarge, response{Error: tooBigMsg, Field: "cv"})
			return
		}
		reject(w, http.StatusBadRequest, "bad form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	// Only the fields the page sends, each exactly once, and one file.
	for k, v := range r.MultipartForm.Value {
		if !textFields[k] || len(v) != 1 {
			reject(w, http.StatusBadRequest, "fields")
			return
		}
	}
	for k, v := range r.MultipartForm.File {
		if k != "cv" || len(v) != 1 {
			reject(w, http.StatusBadRequest, "fields")
			return
		}
	}

	if strings.TrimSpace(r.FormValue("_honey")) != "" {
		reject(w, http.StatusBadRequest, "honeypot")
		return
	}

	tok := r.FormValue("_t")
	switch err := checkToken(key, tok); err {
	case nil:
	case errTokenEarly, errTokenExpired:
		log.Printf("apply: rejected: token %v", err)
		reply(w, http.StatusBadRequest, response{Error: generic, Code: "retry"})
		return
	default:
		reject(w, http.StatusBadRequest, "token "+err.Error())
		return
	}

	a, err := validate(r.FormValue("name"), r.FormValue("phone"), r.FormValue("email"))
	if err == nil {
		var file multipart.File
		var hdr *multipart.FileHeader
		if f, fh, ferr := r.FormFile("cv"); ferr == nil {
			defer f.Close()
			file, hdr = f, fh
		}
		a.CV, err = readCV(file, hdr)
	}
	if err != nil {
		var fe *fieldError
		errors.As(err, &fe)
		log.Printf("apply: rejected: invalid %s", fe.Why)
		reply(w, http.StatusBadRequest, response{Error: fe.Msg, Field: fe.Field})
		return
	}
	a.CVName = cvFilename(a.Name)

	if err := sendMail(loadConfig(), a); err != nil {
		// Only stepErrors reach here and they carry no PII by construction.
		log.Printf("apply: graph error at %v", err)
		reply(w, http.StatusBadGateway, response{Error: fallback})
		return
	}
	markUsed(tok)
	log.Print("apply: sent")
	reply(w, http.StatusOK, response{OK: true})
}
