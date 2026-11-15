//go:build darwin

// **Was Platz belegt, und wie er frei wird** — Werkstück 4 von Phase 4.13
// (#1352).
//
// # Der Anlass, gemessen
//
// **Am 2026-09-07 auf dem Rechner des Auftraggebers gezählt**: neben dem
// Bestand von 127 MB lagen **dreißig** Sicherungen vor Migrationen — zusammen
// über zwei Gigabyte. Sie entstehen bei jedem Start, der migriert, und
// **niemand räumt sie weg**: `sicherungVorMigration` schreibt
// `bestand-vor-<fassung>-<marke>.db`, und die Marke ist bei jedem neuen Stand
// eine andere.
//
// **Das ist kein Fehler in der Sicherung** — sie ist richtig, und die jüngste
// ist der Rückweg. Es fehlt der Ort, an dem steht, was daliegt, und der Weg,
// es loszuwerden.

package app_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Platz wird nach Posten aufgezählt, und der Bestand ist einer.**
func TestDerPlatzNenntSeinePosten(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	antwort := anwendung.Platz(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	nach := map[string]app.PlatzpostenDTO{}
	for _, posten := range antwort.Posten {
		nach[posten.Posten] = posten
	}

	bestand, gibts := nach[string(domain.PostenBestand)]
	if !gibts {
		t.Fatalf("der Bestand fehlt in %+v", antwort.Posten)
	}

	if bestand.Bytes <= 0 {
		t.Errorf("der Bestand wiegt %d Byte", bestand.Bytes)
	}

	// **Er lässt sich nicht wegräumen** — er ist die Anwendung.
	if bestand.Raeumbar {
		t.Error("der Bestand gilt als räumbar")
	}
}

// **Die Sicherungen vor Migrationen werden gezählt — und sind räumbar.**
func TestDieVorsicherungenSindRaeumbar(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	neben := filepath.Dir(bestand)
	for _, name := range []string{
		"bestand-vor-0.1.0.db",
		"bestand-vor-0.2.0-4498ba0291e4.db",
		"bestand-vor-0.3.0-0e3c7fea30ac.db",
	} {
		if err := os.WriteFile(filepath.Join(neben, name),
			make([]byte, 4096), 0o600); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	antwort := anwendung.Platz(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	var vor app.PlatzpostenDTO
	for _, posten := range antwort.Posten {
		if posten.Posten == string(domain.PostenVorsicherungen) {
			vor = posten
		}
	}

	if vor.Dateien != 3 {
		t.Errorf("es wurden %d Vorsicherungen gezählt, nicht 3: %+v", vor.Dateien, vor)
	}

	if !vor.Raeumbar {
		t.Error("die Vorsicherungen gelten als nicht räumbar")
	}
}

// **Und das Wegräumen lässt die jüngste stehen.**
//
// # Warum die jüngste bleibt
//
// **Weil sie der Rückweg ist.** Wer nach einer Migration merkt, dass etwas
// fehlt, spielt sie zurück; wer sie mit wegräumte, nähme genau die Sicherung,
// für die es sie gibt.
func TestDasWegraeumenLaesstDieJuengsteStehen(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	neben := filepath.Dir(bestand)
	namen := []string{
		"bestand-vor-0.1.0.db",
		"bestand-vor-0.2.0-4498ba0291e4.db",
		"bestand-vor-0.3.0-0e3c7fea30ac.db",
	}

	for stelle, name := range namen {
		pfad := filepath.Join(neben, name)
		if err := os.WriteFile(pfad, make([]byte, 4096), 0o600); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}

		// **Der Zeitstempel macht die Reihenfolge** — nicht der Name.
		alt := time.Now().Add(time.Duration(stelle-len(namen)) * time.Hour)
		if err := os.Chtimes(pfad, alt, alt); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	geraeumt := anwendung.PlatzFreigeben(t.Context(),
		string(domain.PostenVorsicherungen))
	if geraeumt.Fehler != nil {
		t.Fatalf("%+v", geraeumt.Fehler)
	}

	uebrig, err := filepath.Glob(filepath.Join(neben, "bestand-vor-*.db"))
	if err != nil {
		t.Fatalf("nachsehen: %v", err)
	}

	if len(uebrig) != 1 {
		t.Fatalf("es blieben %d Dateien: %v", len(uebrig), uebrig)
	}

	if filepath.Base(uebrig[0]) != "bestand-vor-0.3.0-0e3c7fea30ac.db" {
		t.Errorf("es blieb %q — erwartet war die jüngste", filepath.Base(uebrig[0]))
	}
}

// **Ein Posten, den es nicht gibt, wird abgewiesen** (§8).
func TestEinUnbekannterPostenWirdAbgewiesen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if geraeumt := anwendung.PlatzFreigeben(
		t.Context(), "gibtsnicht"); geraeumt.Fehler == nil {
		t.Error("ein unbekannter Posten wurde angenommen")
	}
}

// **Und einer, der nicht räumbar ist, ebenso.**
func TestEinNichtRaeumbarerPostenWirdAbgewiesen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if geraeumt := anwendung.PlatzFreigeben(
		t.Context(), string(domain.PostenBestand)); geraeumt.Fehler == nil {
		t.Error("der Bestand ließ sich wegräumen")
	}
}

// **Die Fläche ohne Fenster gibt dieselbe Zählung** (§16).
//
// **Über den Ort und nicht über eine Verbindung**: gezählt werden Dateien, und
// die liegen auch dann da, wenn der Bestand nicht aufgeht — genau dann will
// jemand wissen, was Platz belegt.
func TestDiePlatzflaecheZaehltDasselbe(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	amFenster := anwendung.Platz(t.Context())
	if amFenster.Fehler != nil {
		t.Fatalf("%+v", amFenster.Fehler)
	}

	ohneFenster := app.PlatzflaecheUeber(bestand).Platz(t.Context())
	if ohneFenster.Fehler != nil {
		t.Fatalf("%+v", ohneFenster.Fehler)
	}

	if len(ohneFenster.Posten) != len(amFenster.Posten) {
		t.Fatalf("%d Posten ohne Fenster, %d am Fenster",
			len(ohneFenster.Posten), len(amFenster.Posten))
	}

	// **Und sie räumt auch weg** — sonst wäre die Fläche eine halbe.
	geraeumt := app.PlatzflaecheUeber(bestand).PlatzFreigeben(
		t.Context(), string(domain.PostenVorsicherungen))
	if geraeumt.Fehler != nil {
		t.Errorf("%+v", geraeumt.Fehler)
	}
}

// **Ein Ort, an dem sich nichts lesen lässt, wird gemeldet** (§8).
//
// **Nicht „null Byte"**: eine Anwendung, die einen Lesefehler als „da liegt
// nichts" ausgibt, sagt jemandem, seine Platte sei leer, während sie voll ist.
func TestEinUnlesbarerOrtWirdGemeldet(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	// **Das Verzeichnis wird unlesbar gemacht** — genau das, was fremde
	// Rechte bewirken.
	neben := filepath.Dir(bestand)

	sperre := filepath.Join(neben, "gesperrt")
	if err := os.Mkdir(sperre, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sperre, "etwas"),
		make([]byte, 16), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	// **Das Bilderverzeichnis zeigt dorthin** — es steht neben dem Bestand.
	bilder := filepath.Join(neben, "beiwerk")
	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.Chmod(bilder, 0o000); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	//nolint:gosec,errcheck // G302: ohne Betretungsrecht räumt `t.TempDir`
	// nicht ab
	t.Cleanup(func() { _ = os.Chmod(bilder, 0o700) })

	if antwort := anwendung.Platz(t.Context()); antwort.Fehler == nil {
		t.Error("ein unlesbares Verzeichnis gab eine Zählung")
	}
}
