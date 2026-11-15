// **Ein Bezug räumt kein Paket weg, in das jemand schreibt** —
// #1190.
//
// # Der Befund
//
// `msepaket.DateiSchreiben` nimmt seit
// #1189 eine
// Sperre neben dem **Paket** (`<paket>.schreibsperre`) und hält sie über
// Lesen, Prüfen und Tauschen. Alle vier Flächen der Paketwerkstatt gehen da
// durch.
//
// **Ein zweiter Schreibweg nahm sie nicht.** `Paketbezieher.Beziehen` tauscht
// einen **Baum** aus (`pakete/M15-Magic-Pack/`); die Schreibsperre liegt an
// einer **Paketdatei** darin (`…/data/magic-m15.mse-style.schreibsperre`).
// Zwei Sperren, die einander nie sehen.
//
// Seit dem 2026-08-28 wird der bisherige Stand beiseitegelegt statt gelöscht
// (#1170,
// Befund 3) — die Datei ist also nicht mehr **weg**, sondern woanders. Für
// jemanden, der gerade schreibt, ist das derselbe Verlust.
//
// # Was jetzt gilt
//
// Der Bezug nimmt die **ausschließliche** Sperre am Bezugsziel, der
// Schreibweg zusätzlich eine **geteilte** am umgebenden Bezugsbaum. Damit
// sehen sich die beiden Ebenen.
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinBezugWartetAufDenSchreiberImBaum(t *testing.T) {
	dienst := paketserver(t)
	wurzel := t.TempDir()
	ort := filepath.Join(wurzel, "Probe")

	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Jemand schreibt im Baum** — er hält die geteilte Sperre am Bezugsziel.
	griff, err := dateisperre.Oeffnen(ort + service.Bezugssperrendung)
	if err != nil {
		t.Fatalf("die Sperre nehmen: %v", err)
	}

	if fehler := griff.Teilen(); fehler != nil {
		t.Fatalf("teilen: %v", fehler)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Test

	// **Kurze Frist**, damit der Lauf nicht zehn Sekunden steht.
	zurueck := service.MitBezugsfrist(50 * 1000 * 1000)
	t.Cleanup(zurueck)

	_, bezugsfehler := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), dienst.URL, "3.1.0", "data", "Probe")
	if bezugsfehler == nil {
		t.Fatal("der Bezug lief durch, während jemand im Baum schreibt")
	}

	if !strings.Contains(bezugsfehler.Error(), "schreibt gerade jemand") {
		t.Errorf("die Meldung %q sagt nicht, woran es lag", bezugsfehler.Error())
	}
}

// **Ohne Schreiber läuft der Bezug wie immer.**
func TestOhneSchreiberLaeuftDerBezugDurch(t *testing.T) {
	dienst := paketserver(t)
	wurzel := t.TempDir()

	ergebnis, err := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), dienst.URL, "3.1.0", "data", "Probe")
	if err != nil {
		t.Fatalf("Beziehen: %v", err)
	}

	if _, statfehler := os.Stat(filepath.Join(ergebnis.Ort, "probe.mse-game")); statfehler != nil {
		t.Errorf("das Paket steht nicht am Ort: %v", statfehler)
	}

	// **Und die Sperrdatei bleibt liegen** — das ist Absicht.
	//
	// Hier stand das Gegenteil, und die Sperre zerfiel daran: `flock` hängt an
	// der **Inode**, nicht am Pfad. Ein Wartender hält die alte Inode schon
	// offen; wird die Datei entfernt, legt der nächste Öffner am selben Pfad
	// eine neue an und sperrt die — dann halten zwei Prozesse gleichzeitig
	// „die" Sperre. Angestrichen vom Nachlauf über die Nacharbeit
	// (#1282,
	// Befund 5).
	if _, statfehler := os.Stat(ergebnis.Ort + service.Bezugssperrendung); statfehler != nil {
		t.Errorf("die Sperrdatei ist weg: %v — ein Wartender und ein Dritter "+
			"bekommen dann zwei unabhängige Sperren am selben Pfad", statfehler)
	}
}

// **Und ein Schreiber im Baum hält einen Bezug auf.**
//
// Das ist die andere Hälfte: die Werkstatt nimmt die **geteilte** Sperre am
// umgebenden Bezugsbaum, damit ein Bezug nicht dazwischenkommt.
func TestEinSchreiberImBaumHaeltDenBezugAuf(t *testing.T) {
	wurzel := t.TempDir()
	baum := filepath.Join(wurzel, "M15-Magic-Pack")
	paket := filepath.Join(baum, "data", "probe.mse-style")

	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(paket, "style"),
		[]byte("mse version: 2.0.0\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Ein Bezug läuft gerade** — er hält die ausschließliche Sperre.
	griff, err := dateisperre.Oeffnen(baum + service.Bezugssperrendung)
	if err != nil {
		t.Fatalf("die Sperre nehmen: %v", err)
	}

	if fehler := griff.Allein(); fehler != nil {
		t.Fatalf("allein: %v", fehler)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Test

	zurueck := service.MitBezugsfrist(50 * 1000 * 1000)
	t.Cleanup(zurueck)

	werkstatt := service.NeuePaketwerkstatt(wurzel)

	_, schreibfehler := werkstatt.Schreiben(t.Context(),
		filepath.Join("M15-Magic-Pack", "data", "probe.mse-style"), "style",
		"mse version: 2.0.0\nneu: ja\n", "")
	if schreibfehler == nil {
		t.Fatal("es wurde geschrieben, während ein Bezug läuft")
	}

	if !strings.Contains(schreibfehler.Error(), "schreibt gerade jemand") {
		t.Errorf("die Meldung %q sagt nicht, woran es lag", schreibfehler.Error())
	}
}
