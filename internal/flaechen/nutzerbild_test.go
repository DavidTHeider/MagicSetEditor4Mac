// **Das Profilbild — an jeder Fläche** — Werkstück 8 von Phase 4.12
// (#1351).
//
// # Der Ablauf
//
// Bild setzen → nachsehen, dass der Nutzer es trägt und die Datei liegt →
// wegnehmen → nachsehen, dass beides weg ist. **Und zweimal dieselbe
// Ablehnung**: eine veraltete Fassung schreibt nicht, und eine Datei, die kein
// Bild ist, kommt nicht durch.
//
// # Warum jeder Lauf sein eigenes Zuhause bekommt
//
// **Weil das Bild im Benutzerverzeichnis landet und nicht am Bestand.**
// `app.Bilderverzeichnis()` hängt an `os.UserConfigDir()`, und das liest
// `HOME`. Ein Lauf ohne diese Zeile schriebe in das **echte**
// `~/Library/Application Support/MagicSetEditor4Mac/beiwerk/nutzer/` — §3
// nennt genau diesen Fall, und er wäre grün, weil er wirklich funktioniert
// hat.
package flaechen_test

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestDerGanzeBildwegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenesZuhause(t)
	pfad := probegesicht(t)

	aus := mussKommandozeile(t, umgebung, "nutzer-bild", "1", pfad, "1")
	if !strings.Contains(aus, "nutzer-1-") {
		t.Errorf("der Dateiname fehlt: %q", aus)
	}

	// **Die Datei liegt wirklich da.** Ein Bestand, der einen Namen führt,
	// hinter dem nichts liegt, wäre der schlechteste Ausgang (§8).
	if _, err := os.Stat(profilbildort(t, umgebung)); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}

	// **Und eine veraltete Fassung schreibt nicht** — die erste hat sie
	// hochgezählt.
	if _, _, code := kommandozeile(t, umgebung, "nutzer-bild", "1", pfad, "1"); code != 1 {
		t.Errorf("die veraltete Fassung kam durch (%d)", code)
	}

	mussKommandozeile(t, umgebung, "nutzer-bild", "1", "--keines", "2")

	if _, err := os.Stat(profilbildort(t, umgebung)); !os.IsNotExist(err) {
		t.Errorf("die Datei liegt noch: %v", err)
	}
}

// **Derselbe Weg über MCP.**
func TestDerGanzeBildwegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenesZuhause(t)

	mussMCP(t, umgebung, "schreiben.nutzer.bild", map[string]any{
		"nutzerID": "1", "pfad": probegesicht(t), "fassung": "1",
	})

	if _, err := os.Stat(profilbildort(t, umgebung)); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}

	// **Eine Datei, die kein Bild ist, kommt nicht durch.**
	roh, istFehler := mcpRuf(t, umgebung, "schreiben.nutzer.bild", map[string]any{
		"nutzerID": "1", "pfad": keinGesicht(t), "fassung": "2",
	})
	if !istFehler {
		t.Errorf("die kaputte Datei kam durch: %s", roh)
	}

	mussMCP(t, umgebung, "schreiben.nutzer.bild.loeschen",
		map[string]any{"nutzerID": "1", "fassung": "2"})

	if _, err := os.Stat(profilbildort(t, umgebung)); !os.IsNotExist(err) {
		t.Errorf("die Datei liegt noch: %v", err)
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeBildwegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenesZuhause(t)
	an := starteDienst(t, umgebung)

	antwort := mussREST(t, an, http.MethodPut, "/nutzer/1/bild",
		`{"pfad":"`+probegesicht(t)+`","fassung":"1"}`)

	if !strings.Contains(antwort, "nutzer-1-") {
		t.Errorf("der Dateiname fehlt: %s", antwort)
	}

	if _, err := os.Stat(profilbildort(t, umgebung)); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}

	// **Und eine veraltete Fassung ist ein Konflikt.**
	status, gemeldet := an.ruf(t, http.MethodPut, "/nutzer/1/bild",
		`{"pfad":"`+probegesicht(t)+`","fassung":"1"}`)
	if status < 400 {
		t.Errorf("die veraltete Fassung kam durch: %d (%s)", status, gemeldet)
	}

	mussREST(t, an, http.MethodDelete, "/nutzer/1/bild?fassung=2", "")

	if _, err := os.Stat(profilbildort(t, umgebung)); !os.IsNotExist(err) {
		t.Errorf("die Datei liegt noch: %v", err)
	}
}

// **Und jede Fläche gibt einen Dateinamen derselben Gestalt aus.**
//
// **Das ist die Zusage aus §16**, und sie ist mehr als „die Fähigkeit ist
// vorhanden": drei Flächen, ein Ergebnis. **Derselbe Name ist es seit dem
// Nachlauf über Phase 6.1 nicht mehr** (Befund 1): jedes Ablegen würfelt
// seinen Namen, damit zwei gleichzeitige Wechsel sich keine Datei teilen.
// Was jede Fläche gleich hält, ist die Gestalt — `nutzer-<Kennung>-<zwölf
// Hexziffern>.png` —, und ein Name, der je Fläche anders **gebaut** wäre,
// machte die Bilder untereinander unauffindbar.
func TestJedeFlaecheGibtEinenBildnamenDerselbenGestalt(t *testing.T) {
	t.Parallel()

	ausKommandozeile := strings.TrimSpace(strings.SplitN(
		mussKommandozeile(t, eigenesZuhause(t), "nutzer-bild", "1",
			probegesicht(t), "1"), "\t", 2)[1])

	ausMCP := bildnameAus(t, mussMCP(t, eigenesZuhause(t), "schreiben.nutzer.bild",
		map[string]any{"nutzerID": "1", "pfad": probegesicht(t), "fassung": "1"}))

	umgebung := eigenesZuhause(t)
	ausDienst := bildnameAus(t, mussREST(t, starteDienst(t, umgebung),
		http.MethodPut, "/nutzer/1/bild",
		`{"pfad":"`+probegesicht(t)+`","fassung":"1"}`))

	gestalt := regexp.MustCompile(`^nutzer-1-[0-9a-f]{12}\.png$`)

	for flaeche, name := range map[string]string{
		"Kommandozeile": ausKommandozeile, "MCP": ausMCP, "Dienst": ausDienst,
	} {
		if !gestalt.MatchString(name) {
			t.Errorf("%s gibt %q — erwartet nutzer-1-<zwölf Hexziffern>.png", flaeche, name)
		}
	}
}

// bildnameAus liest den Bildnamen aus einer Nutzerantwort.
func bildnameAus(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Nutzer struct {
			Bild string `json:"bild"`
		} `json:"nutzer"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Nutzerantwort lesen: %v — %s", err, roh)
	}

	return gelesen.Nutzer.Bild
}

// profilbildort baut den Ort des Profilbildes in dieser Umgebung.
//
// **Derselbe Weg wie `app.Bilderverzeichnis()`**, von Hand nachgebaut: die
// Flächen laufen als eigene Prozesse, und dieser Lauf kann ihnen den Ort nicht
// abfragen — dieselbe Lage wie bei `probepaket`.
func profilbildort(t *testing.T, umgebung []string) string {
	t.Helper()

	return bildort(t, umgebung, "nutzer-1")
}

// bildort sucht das Bild mit diesem Vorsatz — oder gibt einen Ort, an dem
// keines liegt.
//
// **Gesucht und nicht gebaut**: seit dem Zweitblick über Phase 6.1 trägt der
// Name den Inhalt (`nutzer-1-….png`), und ein abgewiesenes Schreiben lässt
// das alte stehen. Liegen zwei, ist das ein Befund und kein Treffer.
func bildort(t *testing.T, umgebung []string, vorsatz string) string {
	t.Helper()

	zuhause := ""

	for _, eintrag := range umgebung {
		if wert, gefunden := strings.CutPrefix(eintrag, "HOME="); gefunden {
			zuhause = wert
		}
	}

	if zuhause == "" {
		t.Fatal("die Umgebung nennt kein Zuhause")
	}

	ort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "beiwerk", "nutzer")

	treffer, err := filepath.Glob(filepath.Join(ort, vorsatz+"-*.png"))
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	switch len(treffer) {
	case 0:
		return filepath.Join(ort, vorsatz+"-keines.png")
	case 1:
		return treffer[0]
	default:
		t.Fatalf("mehr als ein Bild mit %q: %v", vorsatz, treffer)

		return ""
	}
}

// probegesicht schreibt ein winziges PNG und gibt seinen Pfad.
func probegesicht(t *testing.T) string {
	t.Helper()

	bild := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 255})

	pfad := filepath.Join(t.TempDir(), "gesicht.png")

	datei, err := os.Create(pfad) // #nosec G304 -- ein Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := png.Encode(datei, bild); fehler != nil {
		t.Fatalf("kodieren: %v", fehler)
	}

	if fehler := datei.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return pfad
}

// keinGesicht schreibt eine Datei, die keine Bilddatei ist.
func keinGesicht(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "gesicht.png")
	if err := os.WriteFile(pfad, []byte("kein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// **Das Abzeichen eines Profils — an jeder Fläche** — Werkstück 9 von
// Phase 4.12
// (#1351),
// Punkt 13.
//
// **Ein Lauf für drei Flächen und nicht drei Läufe**: der Weg ist an allen
// dreien derselbe, und was ihn unterscheidet, ist der Aufruf. Was geprüft
// wird, ist die Zusage: das Abzeichen liegt **neben** dem Profilbild und
// trägt die **Profil**kennung.
func TestDerGanzeAbzeichenwegAnJederFlaeche(t *testing.T) {
	t.Parallel()

	kommandozeilenlauf := func(t *testing.T, umgebung []string, pfad string) {
		t.Helper()

		aus := mussKommandozeile(t, umgebung, "profil-badge", "1", pfad, "1")
		if !strings.Contains(aus, "profil-1-") {
			t.Errorf("der Dateiname fehlt: %q", aus)
		}
	}

	mcplauf := func(t *testing.T, umgebung []string, pfad string) {
		t.Helper()

		mussMCP(t, umgebung, "schreiben.profil.badge", map[string]any{
			"profilID": "1", "pfad": pfad, "fassung": "1",
		})
	}

	dienstlauf := func(t *testing.T, umgebung []string, pfad string) {
		t.Helper()

		mussREST(t, starteDienst(t, umgebung), http.MethodPut, "/profile/1/badge",
			`{"pfad":"`+pfad+`","fassung":"1"}`)
	}

	for name, lauf := range map[string]func(*testing.T, []string, string){
		"Kommandozeile": kommandozeilenlauf,
		"MCP":           mcplauf,
		"Dienst":        dienstlauf,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			umgebung := eigenesZuhause(t)
			lauf(t, umgebung, probegesicht(t))

			// **Es liegt unter seinem eigenen Namen** — neben dem
			// Profilbild und nicht an seiner Stelle.
			ort := bildort(t, umgebung, "profil-1")

			if _, err := os.Stat(ort); err != nil {
				t.Errorf("die Datei fehlt: %v", err)
			}
		})
	}
}
