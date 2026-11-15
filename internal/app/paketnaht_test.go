// **Die Paketnaht für die Flächen ohne Fenster**
// (#1107).
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Naht frischt auf und listet — das ist ihr Zweck.**
//
// Der Katalog steht im Bestand und ist auf einem frischen Rechner leer; ohne
// das Auffrischen gäbe jede Fläche ohne Fenster eine andere Antwort als das
// Fenster (§16).
func TestDiePaketnahtFrischtAufUndListet(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	weg, nahtfehler := app.PaketnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	pakete, befunde, lesefehler := weg.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if len(pakete) == 0 {
		t.Fatalf("kein Paket — die Saat liegt im Zuhause (Befunde: %+v)", befunde)
	}

	gefunden := false

	for _, zeile := range pakete {
		// **Von Hand ins verwaltete Verzeichnis gelegt ist eigen**
		// (DOM-01.AK-01, #1696) — so legt `saatInsZuhause` die Saat hin. Im
		// Betrieb liegt sie im Bündel und ist mitgeliefert.
		if zeile.Paket.Name == "probe" {
			gefunden = zeile.Herkunftsart == string(domain.PaketEigen)
		}
	}

	if !gefunden {
		t.Errorf("»probe« fehlt: %+v", pakete)
	}

	// **Und die beiden übrigen Methoden antworten auch.**
	if _, err := weg.Auffrischen(probekontext(t)); err != nil {
		t.Errorf("Auffrischen: %v", err)
	}

	if _, err := weg.Abhaengigkeiten(probekontext(t)); err != nil {
		t.Errorf("Abhaengigkeiten: %v", err)
	}
}

// **Ein Bezug ins Leere scheitert und frischt nicht auf.**
//
// **Die Adresse zeigt auf einen geschlossenen Port und nicht ins Netz** (§3):
// ein Test, der wirklich lädt, prüft die Leitung und nicht den Code.
func TestDerBezugScheitertOhneNetz(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	weg, nahtfehler := app.PaketnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	if _, err := weg.Beziehen(probekontext(t),
		"http://127.0.0.1:1/gibtsnicht", "abc1234", "data", "probe"); err == nil {
		t.Error("ein Bezug ins Leere kam durch")
	}
}

// **Ein Lesefehler beim Auffrischen kommt durch.**
func TestDiePaketnahtReichtIhreFehlerDurch(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)

	weg, nahtfehler := app.PaketnahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE packages`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, _, err := weg.Liste(probekontext(t)); err == nil {
		t.Error("das Auffrischen ohne Tabelle kam durch")
	}

	if _, err := weg.Abhaengigkeiten(probekontext(t)); err == nil {
		t.Error("die Abhängigkeiten ohne Tabelle kamen durch")
	}
}

// **Ein Lesefehler nach geglücktem Auffrischen kommt durch.**
//
// **Nur über ein Double zu erreichen**: im Betrieb hängen Auffrischen und
// Lesen an derselben Tabelle, und wer sie entfernt, scheitert schon am ersten
// Zug. `paketfuehrer.listenfehler` scheitert **nur** beim Lesen — genau dieser
// Weg (§10, „ein Wächter, der die halbe Zusage hält").
func TestEinLesefehlerNachDemAuffrischenKommtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, nahtfehler := app.PaketnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	weg.MitVerwalter(&paketfuehrer{listenfehler: errors.New("die Liste ist zu")})

	if _, _, err := weg.Liste(probekontext(t)); err == nil {
		t.Error("der Lesefehler kam nicht durch")
	}
}

// **Ein geglückter Bezug frischt danach auf.**
//
// **Über ein Double und nicht über das Netz** (§3): ein Test, der wirklich
// lädt, prüft die Leitung — und lädt bei jedem Lauf.
func TestEinGeglueckterBezugFrischtAuf(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, nahtfehler := app.PaketnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	weg.MitBezieher(&bezugsprobe{ergebnis: &service.Bezugsergebnis{
		Ort: "/pakete/M15", Stand: "3.1.0", Pruefsumme: "sha256:…",
	}})

	ergebnis, fehler := weg.Beziehen(probekontext(t),
		"https://github.com/x/y", "3.1.0", "data", "M15")
	if fehler != nil {
		t.Fatalf("Beziehen: %v", fehler)
	}

	if ergebnis.Ort != "/pakete/M15" {
		t.Errorf("das Ergebnis kam verstümmelt an: %+v", ergebnis)
	}
}

// **Und ein Auffrischfehler nach geglücktem Bezug wird gemeldet — das
// Ergebnis bleibt trotzdem stehen.**
//
// Geschrieben ist geschrieben (§7, Regel 3): das Paket liegt auf der Platte,
// gleich ob der Katalog danach lesbar war.
func TestEinAuffrischfehlerNachDemBezugWirdGemeldet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, nahtfehler := app.PaketnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	weg.MitBezieher(&bezugsprobe{ergebnis: &service.Bezugsergebnis{Ort: "/pakete/M15"}}).
		MitVerwalter(&paketfuehrer{fehler: errors.New("der Katalog ist zu")})

	ergebnis, fehler := weg.Beziehen(probekontext(t),
		"https://github.com/x/y", "3.1.0", "data", "M15")
	if fehler == nil {
		t.Error("der Auffrischfehler kam nicht durch")
	}

	if ergebnis == nil || ergebnis.Ort != "/pakete/M15" {
		t.Errorf("das Ergebnis ging verloren: %+v", ergebnis)
	}
}
