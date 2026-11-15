package netz

// **Ein Proxy, der `CONNECT` annimmt, ist kein Befund** — D02
// (#2233).
//
// Der Haken sieht **jede** Antwort auf `CONNECT`, auch die glückliche; nur ein
// `407` wird zum Fehler. Der Lauf über eine echte Gegenstelle prüft das `407`
// (`proxyanmeldung_test.go`); dieser hier die Antworten daneben, für die man
// einen ganzen Tunnel bräuchte.

import (
	"net/http"
	"net/url"
	"testing"
)

func TestEinAngenommenesConnectIstKeinBefund(t *testing.T) {
	t.Parallel()

	proxy := &url.URL{Scheme: "http", Host: "127.0.0.1:3128"}

	for _, status := range []int{http.StatusOK, http.StatusForbidden, http.StatusBadGateway} {
		if fehler := proxyanmeldung(t.Context(), proxy, nil, &http.Response{StatusCode: status}); fehler != nil {
			t.Errorf("%d wurde zum Befund: %v", status, fehler)
		}
	}

	// **Und das `407` daneben doch** — Beleg, dass der Haken unterscheidet.
	if proxyanmeldung(t.Context(), proxy, nil, &http.Response{
		StatusCode: http.StatusProxyAuthRequired, Status: "407 Proxy Authentication Required",
	}) == nil {
		t.Error("ein 407 ging als glücklich durch")
	}
}
