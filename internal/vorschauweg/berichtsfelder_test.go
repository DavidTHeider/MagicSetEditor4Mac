package vorschauweg

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// **Jedes Feld eines Berichts schreibt die Seite** — angestrichen vom
// Zweitblick über Phase H (#2086): `source-file` kam mit Abfrage und Anker
// ins Protokoll, und ein Bericht muss nicht von WebKit stammen — jedes
// Skript der Seite kann an [Berichtsweg] schicken, was es will. Ins
// Protokoll kommt deshalb aus **jedem** Feld nur, was als Diagnose taugt:
// ein Anweisungsname, ein Ursprung, eine Datei des eigenen Bündels.
func TestKeinBerichtsfeldTraegtGeheimesInsProtokoll(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, bericht string
		erwartet      []string
	}{
		{
			"eine fremde Quelle nur mit Ursprung",
			`{"csp-report":{"effective-directive":"script-src-elem","blocked-uri":"inline","source-file":"https://beispiel.invalid/Privatname.js?token=GEHEIM#Anker","line-number":3}}`,
			[]string{"stelle=https://beispiel.invalid:3"},
		},
		{
			"eine eigene Quelle mit Datei, ohne Abfrage und Anker",
			`{"csp-report":{"effective-directive":"img-src","blocked-uri":"data","source-file":"wails://wails.localhost/assets/index-ABC.js?v=GEHEIM#Anker","line-number":7}}`,
			[]string{"stelle=index-ABC.js:7"},
		},
		{
			"eine Anweisung, die keine ist",
			`{"csp-report":{"effective-directive":"token=GEHEIM","blocked-uri":"inline"}}`,
			[]string{"anweisung=«unbekannt»"},
		},
		{
			"die verletzte Anweisung samt Quellen nur mit Namen",
			`{"csp-report":{"violated-directive":"img-src 'self' GEHEIM","blocked-uri":"data"}}`,
			[]string{"anweisung=img-src"},
		},
		// **Der Wirt der Seite allein macht keine Datei des Bündels** — Nachlauf
		// über Phase H: dieselbe Wirtsangabe kann jedes Skript schicken.
		{
			"ein Name auf dem eigenen Wirt, der nicht im Bündel liegt",
			`{"csp-report":{"effective-directive":"script-src-elem","blocked-uri":"inline","source-file":"wails://wails.localhost/Privatname-GEHEIM.js?token=GEHEIM#Anker","line-number":9}}`,
			[]string{"stelle=«Datei»:9"},
		},
		{
			"eine Quelle, die keine Adresse ist",
			`{"csp-report":{"effective-directive":"img-src","blocked-uri":"data","source-file":"%zzGEHEIM","line-number":1}}`,
			[]string{"stelle=«Adresse»:1"},
		},
		{
			"das Dokument selbst heißt wie seine Datei",
			`{"csp-report":{"effective-directive":"script-src-elem","blocked-uri":"inline","source-file":"wails://wails.localhost/","line-number":776}}`,
			[]string{"stelle=index.html:776"},
		},
		{
			"ein Wort, das wie eine Anweisung aussieht und keine ist",
			`{"csp-report":{"effective-directive":"privatname-geheim","blocked-uri":"inline"}}`,
			[]string{"anweisung=«unbekannt»"},
		},
		{
			"ein abgewiesener Wert, der kein Schlüsselwort ist",
			`{"csp-report":{"effective-directive":"img-src","blocked-uri":"geheimtoken"}}`,
			[]string{"abgewiesen=«Wert»"},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var protokoll bytes.Buffer

			weiche := Schutzweiche(weiter, probebuendel, "default-src 'self'", false, slog.New(slog.NewTextHandler(&protokoll, nil)))
			anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, Berichtsweg, strings.NewReader(fall.bericht))
			anfrage.Host = "wails.localhost"
			weiche.ServeHTTP(httptest.NewRecorder(), anfrage)

			geschrieben := protokoll.String()
			for _, verboten := range []string{"GEHEIM", "Privatname", "privatname", "Anker", "geheimtoken"} {
				if strings.Contains(geschrieben, verboten) {
					t.Errorf("das Protokoll trägt %q: %s", verboten, geschrieben)
				}
			}

			for _, teil := range fall.erwartet {
				if !strings.Contains(geschrieben, teil) {
					t.Errorf("das Protokoll nennt %q nicht: %s", teil, geschrieben)
				}
			}
		})
	}
}

// probebuendel ist eine ausgelieferte Oberfläche mit zwei Dateien.
var probebuendel = fstest.MapFS{
	"index.html":          &fstest.MapFile{},
	"assets/index-ABC.js": &fstest.MapFile{},
}
