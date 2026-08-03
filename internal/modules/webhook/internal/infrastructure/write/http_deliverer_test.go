package write

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestHTTPDelivererSignsPayload(t *testing.T) {
	var gotSig, gotTS, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Webhook-Signature")
		gotTS = r.Header.Get("X-Webhook-Timestamp")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := NewHTTPDeliverer("test-signing-secret-at-least-32-chars!!")
	d.now = func() time.Time { return time.Unix(1700000000, 0).UTC() }
	// validateURL uses lookup; actual dial goes through httptest client.
	d.lookupIP = func(ctx context.Context, host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("8.8.8.8")}, nil
	}
	d.client = srv.Client()
	d.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	payload := []byte(`{"event":"invoice.paid"}`)
	res, err := d.Deliver(context.Background(), "del-1", "invoice.paid", srv.URL, payload)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if gotBody != string(payload) {
		t.Fatalf("body = %s", gotBody)
	}
	if gotTS != "1700000000" {
		t.Fatalf("timestamp = %s", gotTS)
	}

	mac := hmac.New(sha256.New, []byte("test-signing-secret-at-least-32-chars!!"))
	_, _ = mac.Write([]byte(gotTS))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if gotSig != want {
		t.Fatalf("signature = %s want %s", gotSig, want)
	}
}

func TestHTTPDelivererRejectsPrivateIP(t *testing.T) {
	d := NewHTTPDeliverer("test-signing-secret-at-least-32-chars!!")
	d.lookupIP = func(ctx context.Context, host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}

	res, err := d.Deliver(
		context.Background(),
		"del-1",
		"invoice.paid",
		"https://evil.example/hook",
		[]byte(`{}`),
	)
	if err != nil {
		t.Fatalf("expected soft failure, got err %v", err)
	}
	if res.Retryable {
		t.Fatal("private IP should not be retryable")
	}
	if res.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestSignPayloadStable(t *testing.T) {
	ts := strconv.FormatInt(1700000000, 10)
	got := signPayload([]byte("secret"), ts, []byte("body"))
	if got == "" || len(got) != 64 {
		t.Fatalf("unexpected signature %q", got)
	}
}
