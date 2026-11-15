//go:build darwin

// **Was der Platz übersieht** — Werkstück 6 von Phase 4.13
// (#1352).
//
// # Der Befund, gemessen
//
// **Am 2026-09-07 auf dem Rechner des Auftraggebers**: der Anwendungsordner
// wog 13 GB. Die vier Posten aus Werkstück 4 nannten davon `bestand.db`
// (127 MB), `beiwerk/` (leer), `pakete/` (1,0 GB) und die `bestand-vor-*.db`.
// **Nicht gezählt** waren unter anderem
//
//	mse.log                                     1,3 MB
//	bestand.db.beiseite-vor-0035-neuschnitt     114 MB
//	beiwerk-vorher/, beiwerk-vorher-weg/
//	kanonische-sets/, mse2-ausgabe/             9,3 GB
//
// **Eine Aufzählung ist eine Zusage** (§10), und diese hielt sie nicht: die
// Abnahme lautet „ein Ort, an dem steht, wie viel Platz was belegt", und wer
// dort 1,5 GB liest, während der Finder 13 GB zeigt, glaubt keiner der beiden
// Zahlen mehr.
//
// **Die Abhilfe ist kein längeres Verzeichnis von Namen.** Ein Posten
// *Sonstiges* nimmt, was keiner der anderen beansprucht — damit stimmt die
// Summe **von selbst** und nicht, solange jemand die Liste pflegt.
package app_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Summe der Posten ist der Ordner.**
//
// Das ist die Zusage, und sie hält ohne Liste: was kein Posten beansprucht,
// fällt an *Sonstiges*.
func TestDieSummeDerPostenIstDerGanzeOrdner(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	ort := filepath.Dir(bestand)

	// **Genau die Dateien, die am 2026-09-07 danebenlagen und niemandem
	// gehörten.**
	for _, name := range []string{
		app.Protokollname,
		app.Protokollname + ".1",
		"bestand.db.beiseite-vor-0035-neuschnitt",
		"bestand-vor-0.3.0-b19191e4cb99.db.beiseite-2026-09-04",
	} {
		if fehler := os.WriteFile(
			filepath.Join(ort, name), make([]byte, 512), 0o600,
		); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}
	}

	// **Und ein Verzeichnis, das keiner kennt.**
	if fehler := os.MkdirAll(filepath.Join(ort, "beiwerk-vorher"), 0o700); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	if fehler := os.WriteFile(
		filepath.Join(ort, "beiwerk-vorher", "bild.png"), make([]byte, 1024), 0o600,
	); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	antwort := app.PlatzflaecheUeber(bestand).Platz(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	var gezaehlt int64
	for _, posten := range antwort.Posten {
		gezaehlt += posten.Bytes
	}

	if wirklich := ordnergewicht(t, ort); gezaehlt != wirklich {
		t.Errorf("gezählt sind %d Byte, im Ordner liegen %d — "+
			"die Differenz sieht niemand:\n%+v", gezaehlt, wirklich, antwort.Posten)
	}
}

// **Das Protokoll ist ein eigener Posten — und die laufende Datei bleibt.**
//
// # Warum nicht beide weg
//
// **Weil in die eine gerade geschrieben wird.** Sie wegzuräumen nähme der
// Anwendung den Griff, unter dem sie schreibt; die Vorfassung ist das, was
// wirklich Platz belegt und niemand mehr braucht.
func TestDasProtokollLaesstSichWegraeumenBisAufDasLaufende(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	ort := filepath.Dir(bestand)

	// **Auch die weiteren Vorfassungen** — seit OPS-04 bewahrt die Drehung
	// bis zu zehn auf, und `mse.log.2` fiel bis zum Zweitblick über Phase 6.4
	// in den Rest, den niemand räumen kann. `mse.log.alt` ist keine.
	vorfassungen := []string{app.Vorprotokollname, app.Protokollname + ".2", app.Protokollname + ".10"}

	for _, name := range append([]string{app.Protokollname, app.Protokollname + ".alt"}, vorfassungen...) {
		if fehler := os.WriteFile(
			filepath.Join(ort, name), make([]byte, 2048), 0o600,
		); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}
	}

	antwort := app.PlatzflaecheUeber(bestand).PlatzFreigeben(
		t.Context(), string(domain.PostenProtokoll))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ort, app.Protokollname)); fehler != nil {
		t.Errorf("das laufende Protokoll ist fort: %v", fehler)
	}

	for _, name := range vorfassungen {
		if _, fehler := os.Stat(filepath.Join(ort, name)); fehler == nil {
			t.Errorf("die Vorfassung %s blieb liegen", name)
		}
	}

	if _, fehler := os.Stat(filepath.Join(ort, app.Protokollname+".alt")); fehler != nil {
		t.Errorf("eine fremde Datei ist fort: %v", fehler)
	}
}

// ordnergewicht summiert, was wirklich unter einem Ort liegt.
func ordnergewicht(t *testing.T, ort string) int64 {
	t.Helper()

	var summe int64

	fehler := filepath.WalkDir(ort, func(
		_ string, eintrag fs.DirEntry, unterwegs error,
	) error {
		if unterwegs != nil {
			return unterwegs
		}

		if eintrag.IsDir() {
			return nil
		}

		angabe, angabefehler := eintrag.Info()
		if angabefehler != nil {
			return angabefehler
		}

		summe += angabe.Size()

		return nil
	})
	if fehler != nil {
		t.Fatalf("wiegen: %v", fehler)
	}

	return summe
}

// **Was keine Sicherung ist, wird nicht weggeräumt.**
//
// # Der Befund
//
// **Befund 1 des Zweitblicks über Phase 4.13** (2026-09-07): `vorsicherungen`
// ordnete jede Datei mit dem bloßen Namensanfang `bestand-vor-` den **räumbaren**
// Vorsicherungen zu. Der Erzeuger schreibt aber eine engere Form —
// `bestand-vor-<fassung>.db`, bei einem zweiten Stand mit Marke.
//
// Eine Datei wie `bestand-vor-notizen.txt`, die jemand dort abgelegt hat,
// verschwand damit beim Wegräumen. **Eine löschende Funktion vertraute einem
// Namensanfang statt einem Vertrag** — und die Domäne sagt das Gegenteil: was
// die Anwendung nicht kennt, ist *Sonstiges* und nicht räumbar.
//
// # Warum die Zählung das nicht gezeigt hat
//
// Der Lauf über die Summe prüft, dass **jede** Datei irgendwo gezählt wird.
// Ob sie im **richtigen** Posten steht, sieht er nicht — die Summe stimmt in
// beiden Fällen.
func TestWasKeineSicherungIstWirdNichtGeraeumt(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	ort := filepath.Dir(bestand)

	// **Zwei echte Sicherungen**, damit überhaupt etwas zu räumen ist.
	for _, name := range []string{"bestand-vor-0.1.0.db", "bestand-vor-0.2.0.db"} {
		schreibeProbe(t, filepath.Join(ort, name))
	}

	// **Und vier, die den Anfang tragen und keine sind.**
	fremde := []string{
		"bestand-vor-notizen.txt",
		"bestand-vor-0.1.0.db.bak",
		"bestand-vorgestern",
		"bestand-vor-plan.md",

		// **Und fremde Namen mit der richtigen Endung** — der Nachlauf zum
		// Zweitblick hat gezeigt, dass die Liste darüber nur fremde
		// **Endungen** abdeckte. `bestand-vor-notizen.db` ging durch, weil
		// das erste Muster jeden Inhalt vor dem `.db` nahm.
		"bestand-vor-notizen.db",
		"bestand-vor-dem-urlaub.db",
		"bestand-vor-0.1.db",
	}

	for _, name := range fremde {
		schreibeProbe(t, filepath.Join(ort, name))
	}

	antwort := app.PlatzflaecheUeber(bestand).PlatzFreigeben(
		t.Context(), string(domain.PostenVorsicherungen))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, name := range fremde {
		if _, fehler := os.Stat(filepath.Join(ort, name)); fehler != nil {
			t.Errorf("%s wurde weggeräumt — es ist keine Sicherung: %v", name, fehler)
		}
	}
}

// schreibeProbe legt eine Datei mit etwas Inhalt an.
func schreibeProbe(t *testing.T, pfad string) {
	t.Helper()

	if fehler := os.WriteFile(pfad, make([]byte, 128), 0o600); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}
}
