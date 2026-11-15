// **Der Austausch der Einstellungen — an jeder Fläche** — Werkstück O3 von
// Phase 5.2 (#1427).
//
// # Der Ablauf, und er ist an allen dreien derselbe
//
// Eine Einstellung setzen → ausführen → nachsehen, dass **kein Geheimnis, kein
// Pfad und keine lokale Kennung** darin steht → die Vorschau holen → mit ihrer
// Revision übernehmen → nachsehen, dass es angekommen ist.
//
// Und dieselben drei Negativfälle: eine **neuere Fassung**, ein **Schlüssel
// der falschen Ebene** und eine **überholte Revision**. Der letzte ist der,
// für den es die Marke gibt — er lässt den Bestand unverändert.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// austauschAusZukunft ist eine Datei, die dieses Programm nicht lesen kann.
const austauschAusZukunft = `{"fassung":99,"bereiche":{}}`

// austauschMitFalscherEbene legt einen Pfad dieser Platte in einen Abschnitt,
// in den er nicht gehört.
const austauschMitFalscherEbene = `{"fassung":1,"bereiche":` +
	`{"profil":{"vorgabe.hochladepfad":"/Users/jemand/Bilder"}}}`

// austauschMitSprache ist eine gültige Datei mit einer Zeile.
const austauschMitSprache = `{"fassung":1,"bereiche":{"profil":{"sprache":"en"}}}`

// pruefeAusfuhr hält, was an jeder Fläche gilt.
//
// **Was nicht darin stehen darf, steht hier als Liste** — der Wächter über die
// Klasse steht in `internal/werkzeuge/austauschsperre_test.go`; dieser Lauf
// ist die Probe am **gebauten** Programm.
func pruefeAusfuhr(t *testing.T, inhalt string) {
	t.Helper()

	if !strings.Contains(inhalt, `"fassung"`) {
		t.Errorf("die Ausfuhr nennt keine Fassung: %s", inhalt)
	}

	for _, verboten := range []string{
		"vorgabe.hochladepfad", "anbieter.", "aktualisierung", "ersteschritte",
	} {
		if strings.Contains(inhalt, verboten) {
			t.Errorf("die Ausfuhr trägt %q: %s", verboten, inhalt)
		}
	}
}

// **Der ganze Weg an der Kommandozeile.**
func TestDerGanzeAustauschwegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "schriftgroesse", "gross")

	inhalt := mussKommandozeile(t, umgebung, "einstellungen", "ausfuhr")
	pruefeAusfuhr(t, inhalt)

	if !strings.Contains(inhalt, "schriftgroesse") {
		t.Errorf("die gesetzte Einstellung fehlt: %s", inhalt)
	}

	datei := alsDatei(t, austauschMitSprache)

	vorschau := mussKommandozeile(t, umgebung, "einstellungen", "vorschau", datei)
	if !strings.Contains(vorschau, "sprache") {
		t.Errorf("die Vorschau nennt die Zeile nicht: %s", vorschau)
	}

	// **Und wann sie wirkt** — die sechste Spalte, OPS-01.AK-02 (#1718).
	if !strings.Contains(vorschau, "\tsofort\n") {
		t.Errorf("die Vorschau nennt die Wirkung nicht: %q", vorschau)
	}

	// **Eine überholte Revision lässt alles stehen.**
	if _, _, code := kommandozeile(t, umgebung, "einstellungen", "einfuhr",
		datei, "vongestern"); code != 1 {
		t.Errorf("die überholte Revision kam durch (%d)", code)
	}

	bericht := mussKommandozeile(t, umgebung, "einstellungen", "einfuhr",
		datei, revisionAus(t, vorschau))
	if !strings.Contains(bericht, "uebernommen\t1") {
		t.Errorf("es wurde nichts übernommen: %s", bericht)
	}

	if danach := mussKommandozeile(t, umgebung, "einstellungen"); !strings.Contains(
		danach, "sprache\ten") {
		t.Errorf("die Sprache blieb nicht: %s", danach)
	}

	// **Ein Verb, das es nicht gibt, ist ein Bedienfehler** — und nicht die
	// stille Übernahme einer Datei, die jemand nur ansehen wollte.
	if _, _, code := kommandozeile(t, umgebung, "einstellungen", "quatsch",
		datei); code != 2 {
		t.Errorf("das erfundene Verb gab %d", code)
	}

	// **Und die zwei Dateien, die nicht durchkommen.**
	if _, _, code := kommandozeile(t, umgebung, "einstellungen", "vorschau",
		alsDatei(t, austauschAusZukunft)); code != 1 {
		t.Errorf("die Fassung aus der Zukunft kam durch (%d)", code)
	}

	falsch := mussKommandozeile(t, umgebung, "einstellungen", "vorschau",
		alsDatei(t, austauschMitFalscherEbene))
	if !strings.Contains(falsch, "befund") {
		t.Errorf("die falsche Ebene wurde nicht gemeldet: %s", falsch)
	}
}

// revisionAus liest die Marke aus einer Vorschau an der Kommandozeile.
func revisionAus(t *testing.T, vorschau string) string {
	t.Helper()

	for zeile := range strings.SplitSeq(vorschau, "\n") {
		if marke, gefunden := strings.CutPrefix(zeile, "revision\t"); gefunden {
			return strings.TrimSpace(marke)
		}
	}

	t.Fatalf("die Vorschau nennt keine Revision: %s", vorschau)

	return ""
}

// alsDatei legt einen Text auf die Platte und gibt seinen Pfad.
func alsDatei(t *testing.T, inhalt string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "austausch.json")
	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("die Datei schreiben: %v", err)
	}

	return pfad
}

// **Derselbe Weg über MCP.**
func TestDerGanzeAustauschwegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	var ausfuhr struct {
		Inhalt string `json:"inhalt"`
	}

	auspacken(t, mussMCP(t, umgebung, "lesen.einstellungen.ausfuhr",
		map[string]any{}), &ausfuhr)
	pruefeAusfuhr(t, ausfuhr.Inhalt)

	var sicht struct {
		Revision string `json:"revision"`
		Zeilen   []struct {
			Schluessel string `json:"schluessel"`
			Herkunft   string `json:"herkunft"`
		} `json:"zeilen"`
	}

	auspacken(t, mussMCP(t, umgebung, "lesen.einstellungen.vorschau",
		map[string]any{"inhalt": austauschMitSprache}), &sicht)

	if len(sicht.Zeilen) != 1 || sicht.Zeilen[0].Herkunft == "" {
		t.Fatalf("die Vorschau: %+v", sicht)
	}

	// **Eine überholte Revision lässt alles stehen.**
	if roh, istFehler := mcpRuf(t, umgebung, "schreiben.einstellungen.einfuhr",
		map[string]any{
			"inhalt": austauschMitSprache, "revision": "vongestern",
		}); !istFehler {
		t.Errorf("die überholte Revision kam durch: %s", roh)
	}

	var bericht struct {
		Uebernommen int `json:"uebernommen"`
	}

	auspacken(t, mussMCP(t, umgebung, "schreiben.einstellungen.einfuhr",
		map[string]any{
			"inhalt": austauschMitSprache, "revision": sicht.Revision,
		}), &bericht)

	if bericht.Uebernommen != 1 {
		t.Errorf("es wurde nichts übernommen: %+v", bericht)
	}

	// **Und die zwei Dateien, die nicht durchkommen.**
	if roh, istFehler := mcpRuf(t, umgebung, "lesen.einstellungen.vorschau",
		map[string]any{"inhalt": austauschAusZukunft}); !istFehler {
		t.Errorf("die Fassung aus der Zukunft kam durch: %s", roh)
	}

	if gemeldet := mussMCP(t, umgebung, "lesen.einstellungen.vorschau",
		map[string]any{"inhalt": austauschMitFalscherEbene}); !strings.Contains(
		gemeldet, "vorgabe.hochladepfad") {
		t.Errorf("die falsche Ebene wurde nicht gemeldet: %s", gemeldet)
	}
}

// auspacken liest eine Antwort in eine Gestalt.
func auspacken(t *testing.T, roh string, ziel any) {
	t.Helper()

	if err := json.Unmarshal([]byte(roh), ziel); err != nil {
		t.Fatalf("die Antwort lesen: %v — %s", err, roh)
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeAustauschwegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	var ausfuhr struct {
		Inhalt string `json:"inhalt"`
	}

	auspacken(t, mussREST(t, an, http.MethodGet, "/einstellungen/ausfuhr", ""),
		&ausfuhr)
	pruefeAusfuhr(t, ausfuhr.Inhalt)

	var sicht struct {
		Revision string `json:"revision"`
		Zeilen   []struct {
			Herkunft string `json:"herkunft"`
		} `json:"zeilen"`
	}

	auspacken(t, mussREST(t, an, http.MethodPost, "/einstellungen/vorschau",
		`{"inhalt":`+alsJSONtext(t, austauschMitSprache)+`}`), &sicht)

	if len(sicht.Zeilen) != 1 || sicht.Zeilen[0].Herkunft == "" {
		t.Fatalf("die Vorschau: %+v", sicht)
	}

	// **Eine überholte Revision lässt alles stehen.**
	if status, gemeldet := an.ruf(t, http.MethodPost, "/einstellungen/einfuhr",
		`{"inhalt":`+alsJSONtext(t, austauschMitSprache)+
			`,"revision":"vongestern"}`); status < 400 {
		t.Errorf("die überholte Revision kam durch: %d (%s)", status, gemeldet)
	}

	var bericht struct {
		Uebernommen int `json:"uebernommen"`
	}

	auspacken(t, mussREST(t, an, http.MethodPost, "/einstellungen/einfuhr",
		`{"inhalt":`+alsJSONtext(t, austauschMitSprache)+
			`,"revision":"`+sicht.Revision+`"}`), &bericht)

	if bericht.Uebernommen != 1 {
		t.Errorf("es wurde nichts übernommen: %+v", bericht)
	}

	// **Und die zwei Dateien, die nicht durchkommen.**
	if status, gemeldet := an.ruf(t, http.MethodPost, "/einstellungen/vorschau",
		`{"inhalt":`+alsJSONtext(t, austauschAusZukunft)+`}`); status < 400 {
		t.Errorf("die Fassung aus der Zukunft kam durch: %d (%s)", status, gemeldet)
	}

	if gemeldet := mussREST(t, an, http.MethodPost, "/einstellungen/vorschau",
		`{"inhalt":`+alsJSONtext(t, austauschMitFalscherEbene)+`}`); !strings.Contains(
		gemeldet, "vorgabe.hochladepfad") {
		t.Errorf("die falsche Ebene wurde nicht gemeldet: %s", gemeldet)
	}
}

// alsJSONtext packt einen Text als JSON-Zeichenkette.
func alsJSONtext(t *testing.T, inhalt string) string {
	t.Helper()

	roh, fehler := json.Marshal(inhalt)
	if fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	return string(roh)
}
