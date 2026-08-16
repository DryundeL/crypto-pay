package write

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/webhook/internal/application/ports"
)

const (
	defaultTimeout   = 10 * time.Second
	maxResponseBytes = 64 << 10
)

type HTTPDeliverer struct {
	client   *http.Client
	now      func() time.Time
	lookupIP func(ctx context.Context, host string) ([]net.IP, error)
}

func NewHTTPDeliverer() *HTTPDeliverer {
	d := &HTTPDeliverer{
		now: time.Now,
		lookupIP: func(ctx context.Context, host string) ([]net.IP, error) {
			return net.DefaultResolver.LookupIP(ctx, "ip", host)
		},
	}
	d.client = &http.Client{
		Timeout: defaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("redirects are not allowed")
		},
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           d.dialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return d
}

func (d *HTTPDeliverer) Deliver(
	ctx context.Context,
	deliveryID, eventName, targetURL string,
	payload []byte,
	signingSecret string,
) (ports.DeliverResult, error) {
	if signingSecret == "" {
		return ports.DeliverResult{
			Retryable: false,
			Error:     "signing secret is required",
		}, nil
	}
	if err := d.validateURL(ctx, targetURL); err != nil {
		return ports.DeliverResult{
			Retryable: false,
			Error:     err.Error(),
		}, nil
	}

	ts := strconv.FormatInt(d.now().UTC().Unix(), 10)
	sig := signPayload([]byte(signingSecret), ts, payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return ports.DeliverResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "crypto-pay-webhook/1.0")
	req.Header.Set("X-Webhook-Event", eventName)
	req.Header.Set("X-Webhook-Delivery-Id", deliveryID)
	req.Header.Set("X-Webhook-Timestamp", ts)
	req.Header.Set("X-Webhook-Signature", "sha256="+sig)

	resp, err := d.client.Do(req)
	if err != nil {
		return ports.DeliverResult{
			Retryable: true,
			Error:     err.Error(),
		}, err
	}
	defer resp.Body.Close()
	_, _ = io.CopyN(io.Discard, resp.Body, maxResponseBytes)

	result := ports.DeliverResult{StatusCode: resp.StatusCode}
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return result, nil
	case resp.StatusCode == http.StatusRequestTimeout,
		resp.StatusCode == http.StatusTooManyRequests,
		resp.StatusCode >= 500:
		result.Retryable = true
		result.Error = fmt.Sprintf("http status %d", resp.StatusCode)
		return result, nil
	default:
		result.Retryable = false
		result.Error = fmt.Sprintf("http status %d", resp.StatusCode)
		return result, nil
	}
}

func (d *HTTPDeliverer) validateURL(ctx context.Context, raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url host is required")
	}
	if isBlockedHost(host) {
		return fmt.Errorf("url host is not allowed")
	}

	ips, err := d.lookupIP(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve host: %w", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("resolve host: no addresses")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("url resolves to a blocked address")
		}
	}
	return nil
}

func (d *HTTPDeliverer) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := d.lookupIP(ctx, host)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, ip := range ips {
		if isBlockedIP(ip) {
			lastErr = fmt.Errorf("blocked address")
			continue
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second}
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no dialable addresses")
	}
	return nil, lastErr
}

func signPayload(secret []byte, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func isBlockedHost(host string) bool {
	h := strings.ToLower(host)
	switch h {
	case "localhost", "metadata.google.internal":
		return true
	}
	return strings.HasSuffix(h, ".localhost") || strings.HasSuffix(h, ".local")
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	// AWS/GCP metadata
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}
	return false
}
