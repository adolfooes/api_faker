package handler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

// webhookAllowedHosts lista hosts (hostname ou IP, sem porta) que podem ser alvo
// de webhook mesmo resolvendo para loopback/rede privada. O api_faker existe
// para disparar webhooks contra serviços locais (infrapay em localhost, outros
// containers), então bloquear toda rede privada quebraria o uso; a lista vem de
// WEBHOOK_ALLOWED_HOSTS (separada por vírgula). Link-local — onde mora o
// metadata service de nuvem, 169.254.169.254 — é bloqueado mesmo se listado.
var webhookAllowedHosts = parseAllowedHosts(os.Getenv("WEBHOOK_ALLOWED_HOSTS"))

const defaultWebhookAllowedHosts = "localhost,127.0.0.1,host.docker.internal"

func parseAllowedHosts(raw string) map[string]bool {
	if strings.TrimSpace(raw) == "" {
		raw = defaultWebhookAllowedHosts
	}
	hosts := map[string]bool{}
	for _, h := range strings.Split(raw, ",") {
		if h = strings.ToLower(strings.TrimSpace(h)); h != "" {
			hosts[h] = true
		}
	}
	return hosts
}

func isAllowedHost(host string) bool {
	return webhookAllowedHosts[strings.ToLower(host)]
}

// isAlwaysBlockedIP: nunca alcançável, nem por host da lista de permitidos.
func isAlwaysBlockedIP(ip net.IP) bool {
	return ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified()
}

// isBlockedIP reports whether ip must not be reached by a webhook target that
// is not in the allowlist: loopback, RFC1918/ULA private ranges plus everything
// in isAlwaysBlockedIP.
func isBlockedIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}
	return ip.IsLoopback() || ip.IsPrivate() || isAlwaysBlockedIP(ip)
}

var errBlockedTarget = errors.New("target_url resolves to a disallowed network address")

// validateTargetURL rejects webhook targets early, with a clear message: non-http(s)
// schemes, empty hosts and hosts that resolve to a blocked address. It is a
// convenience check only — the authoritative check is safeDialContext, which runs
// on the IP actually being connected to (a second DNS resolution could otherwise
// return a different address: DNS rebinding).
func validateTargetURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("target_url is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("target_url scheme must be http or https")
	}
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("target_url must include a host")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("target_url host could not be resolved")
	}
	allowed := isAllowedHost(host)
	for _, ip := range ips {
		if isAlwaysBlockedIP(ip) || (!allowed && isBlockedIP(ip)) {
			return errBlockedTarget
		}
	}
	return nil
}

// safeDialContext valida o IP no momento da conexão (net.Dialer.Control recebe o
// endereço já resolvido), fechando a janela entre a validação e a conexão.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	allowed := isAllowedHost(host)
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			ipStr, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(ipStr)
			if ip == nil || isAlwaysBlockedIP(ip) || (!allowed && isBlockedIP(ip)) {
				return errBlockedTarget
			}
			return nil
		},
	}
	return dialer.DialContext(ctx, network, addr)
}

// sensitiveHeaderPattern matches header names that commonly carry credentials
// or webhook signing secrets (e.g. X-Iugu-Signature, Authorization, X-Api-Key).
var sensitiveHeaderPattern = regexp.MustCompile(`(?i)(authorization|signature|token|api[-_]?key|secret)`)

const maskedHeaderValue = "***REDACTED***"

// maskSensitiveHeaders returns a copy of headers with sensitive values redacted,
// so third-party credentials/signatures are never persisted in webhook_dispatch.
func maskSensitiveHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	masked := make(map[string]string, len(headers))
	for k, v := range headers {
		if sensitiveHeaderPattern.MatchString(k) {
			masked[k] = maskedHeaderValue
		} else {
			masked[k] = v
		}
	}
	return masked
}

// maskSensitiveHTTPHeaders is maskSensitiveHeaders for http.Header (target responses).
func maskSensitiveHTTPHeaders(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	masked := make(http.Header, len(h))
	for k, v := range h {
		if sensitiveHeaderPattern.MatchString(k) {
			masked[k] = []string{maskedHeaderValue}
		} else {
			masked[k] = v
		}
	}
	return masked
}

// maskScenarioWebhooks mascara headers sensíveis dos webhooks de um cenário antes
// de devolvê-lo pela API (segredos de assinatura de integrações de terceiros).
func maskScenarioWebhooks(whs []ScenarioWebhookInput) []ScenarioWebhookInput {
	for i := range whs {
		whs[i].Headers = maskSensitiveHeaders(whs[i].Headers)
	}
	return whs
}

// dispatchLimiter: token bucket por conta para POST /webhook/dispatch — sem isso
// uma conta autenticada usa o serviço como proxy de requisições em volume.
type dispatchBucket struct {
	tokens float64
	last   time.Time
}

const (
	dispatchRatePerSecond = 5.0
	dispatchBurst         = 20.0
)

var (
	dispatchMu      sync.Mutex
	dispatchBuckets = map[int64]*dispatchBucket{}
)

func allowDispatch(ownerID int64, now time.Time) bool {
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	b, ok := dispatchBuckets[ownerID]
	if !ok {
		b = &dispatchBucket{tokens: dispatchBurst, last: now}
		dispatchBuckets[ownerID] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * dispatchRatePerSecond
	if b.tokens > dispatchBurst {
		b.tokens = dispatchBurst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
