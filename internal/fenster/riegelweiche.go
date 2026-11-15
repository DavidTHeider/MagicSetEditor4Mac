package fenster

import (
	"net/http"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// riegelweiche legt einen Riegel um die Weiche der Webansicht.
//
// **Hier fängt niemand sonst**: Wails beantwortet jede Anfrage der
// Webansicht in einer eigenen Goroutine ohne `recover`
// (`assetserver_webview.go`, `ServeWebViewRequest`, v3.0.0-beta.27), und
// sein `PanicHandler` sieht sie nie. Eine Panik in der Bildweiche, der
// Schutznaht, dem Rufbuch oder dem Dateiserver dahinter wird hier zu einer
// 500 und einer Zeile mit Vorgangskennung.
func riegelweiche(innen application.Middleware) application.Middleware {
	return func(weiter http.Handler) http.Handler {
		weiche := innen(weiter)

		return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
			// **Nur der Status**: eine Hülle nimmt aus `net/http` nichts als
			// Typen und Zahlen (`internal/schichten/netzhuellen_test.go`).
			if !riegeln("eine Anfrage der Webansicht", func() { weiche.ServeHTTP(schreiben, anfrage) }) {
				schreiben.WriteHeader(http.StatusInternalServerError)
			}
		})
	}
}
