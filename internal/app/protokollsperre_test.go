//go:build darwin

// **Ein zweiter Start dreht kein fremdes Protokoll** — Befund 5 des
// Zweitblicks über Phase 4.13
// (#1352).
//
// # Der Befund
//
// Jeder Prozess öffnet denselben Protokollpfad. Beim Start wird ein volles
// `mse.log` **ohne Absprache** nach `mse.log.1` umbenannt.
//
// Startet Prozess B, während A in ein volles `mse.log` schreibt, wird **As
// offene Datei** zu `.1`. Räumt danach jemand den Platz frei, verliert sie
// ihren letzten Namen — POSIX sagt zu, dass der Inhalt bis zum Schließen
// bestehen bleibt, aber niemand findet ihn mehr. A schreibt weiter ins Leere,
// und seine Zeilen fehlen in der Diagnose.
//
// **Die Klasse**: „laufend" wurde aus einem Pfadnamen abgeleitet statt aus der
// Frage, ob jemand anderes schreibt.
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// **Solange ein anderer schreibt, bleibt seine Datei, wo sie ist.**
func TestEinZweiterStartDrehtDasOffeneProtokollNicht(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, app.Protokollname)

	// **Erst klein**, damit der erste Start nichts zu drehen hat — sonst
	// stammte die Vorfassung von ihm, und der Lauf sagte nichts über den
	// zweiten.
	if fehler := os.WriteFile(pfad, []byte("eine Zeile\n"), 0o600); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	// **Der erste Prozess hält es offen.**
	_, erstesSchliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("das erste Protokoll: %v", fehler)
	}

	t.Cleanup(func() { _ = erstesSchliessen() }) //nolint:errcheck // Aufräumweg

	// **Und jetzt wächst es über die Grenze** — so, wie es im Betrieb wächst:
	// derselbe Inode, in den der erste schreibt.
	if fehler := os.WriteFile(pfad,
		[]byte(strings.Repeat("a", int(app.Protokollgrenze)+1)), 0o600); fehler != nil {
		t.Fatalf("aufblasen: %v", fehler)
	}

	// **Und jetzt startet der zweite.**
	_, zweitesSchliessen, zweiterfehler := app.Protokoll(pfad)
	if zweiterfehler != nil {
		t.Fatalf("das zweite Protokoll: %v", zweiterfehler)
	}

	t.Cleanup(func() { _ = zweitesSchliessen() }) //nolint:errcheck // Aufräumweg

	// **Es darf keine Vorfassung geben** — die wäre die Datei, in die der
	// erste noch schreibt.
	if _, fehler := os.Stat(filepath.Join(ort, app.Vorprotokollname)); fehler == nil {
		t.Error("der zweite Start hat das offene Protokoll des ersten gedreht")
	}
}

// **Und wenn niemand sonst schreibt, wird gedreht.**
//
// **Die Gegenprobe zur Zusage** (§10): ein Drehen, das nie stattfindet, wäre
// dasselbe wie keine Grenze — die Datei wüchse weiter.
func TestOhneZweitenProzessWirdGedreht(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, app.Protokollname)

	if fehler := os.WriteFile(pfad,
		[]byte(strings.Repeat("a", int(app.Protokollgrenze)+1)), 0o600); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	_, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	if fehler := schliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ort, app.Vorprotokollname)); fehler != nil {
		t.Errorf("es wurde nicht gedreht: %v", fehler)
	}
}

// **Ohne Sperrdatei wird nicht gedreht — und das Protokoll läuft trotzdem.**
//
// # Warum dieser Zweig zählt
//
// Die Sperre ist eine **Verbesserung** und keine Voraussetzung: lässt sie sich
// nicht anlegen — ein Verzeichnis, in dem man nichts Neues erzeugen darf —,
// muss das Protokoll weiterschreiben. Das ist dieselbe Haltung wie bei
// `bestand.db.sperre`: „ohne Nachbardatei geht es weiter."
//
// **Gedreht wird dann nicht**, und auch das ist richtig: wer nicht fragen
// kann, ob er allein ist, benennt keine fremde Datei um.
func TestOhneSperrdateiWirdNichtGedreht(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, app.Protokollname)

	if fehler := os.WriteFile(pfad,
		[]byte(strings.Repeat("a", int(app.Protokollgrenze)+1)), 0o600); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	// **Nichts Neues darf hier entstehen** — die Sperrdatei also auch nicht.
	//
	//nolint:gosec // die Rechte **sind** der Fall
	if fehler := os.Chmod(ort, 0o500); fehler != nil {
		t.Fatalf("Rechte setzen: %v", fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) }) //nolint:errcheck,gosec // Aufräumweg

	_, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("das Protokoll ließ sich nicht anlegen: %v", fehler)
	}

	if fehler := schliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ort, app.Vorprotokollname)); fehler == nil {
		t.Error("es wurde gedreht, obwohl niemand fragen konnte")
	}
}

// **Ein gleichzeitiger Start bekommt seine geteilte Sperre — er wartet.**
//
// # Der Befund
//
// **Der Nachlauf zum Zweitblick** (2026-09-07): hier stand `griff.Teilen()`
// mit der Begründung, `LOCK_SH` könne auf einem gültigen Deskriptor nicht
// scheitern. **Das ist falsch.** `flock` steht unter `LOCK_NB`; hält ein
// gleichzeitig startender Prozess gerade den **Alleinanspruch**, scheitert die
// geteilte Sperre mit `EWOULDBLOCK`.
//
// Danach schriebe dieser Prozess **ohne Sperre** — und der nächste Start
// dürfte seine offene Datei drehen. Genau der Fall, den Befund 5 abstellen
// sollte, eine Ebene tiefer.
//
// # Warum der alte Lauf ihn nicht traf
//
// Er begann mit einem bereits **geteilten** Griff, und gegen den kann ein
// zweiter geteilter nicht scheitern. Hier hält der Nachbar **ausschließlich**
// — genau die Lage, die im Betrieb einen Augenblick lang besteht.
func TestEinGleichzeitigerStartWartetAufSeineSperre(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, app.Protokollname)

	if fehler := os.WriteFile(pfad, []byte("eine Zeile\n"), 0o600); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	// **Ein Nachbar hält den Alleinanspruch** — so, wie ein Prozess ihn hält,
	// während er dreht.
	nachbar, fehler := dateisperre.Oeffnen(filepath.Join(ort, app.Protokollsperre))
	if fehler != nil {
		t.Fatalf("die Sperre öffnen: %v", fehler)
	}

	if fehler := nachbar.Allein(); fehler != nil {
		t.Fatalf("den Alleinanspruch nehmen: %v", fehler)
	}

	// **Und er gibt ihn gleich wieder her** — wie im Betrieb, nach einem
	// `Stat` und einem `Rename`.
	losgelassen := make(chan struct{})

	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = nachbar.Schliessen() //nolint:errcheck // Aufräumweg
		close(losgelassen)
	}()

	_, schliessen, protokollfehler := app.Protokoll(pfad)
	if protokollfehler != nil {
		t.Fatalf("das Protokoll: %v", protokollfehler)
	}

	t.Cleanup(func() { _ = schliessen() }) //nolint:errcheck // Aufräumweg

	// **Erst warten, bis der Nachbar wirklich fort ist.**
	//
	// **Sonst prüfte der Lauf ihn statt das Protokoll**: ein Alleinanspruch
	// von außen scheitert auch am Nachbarn, und der Lauf wäre grün, ohne
	// etwas über die Sperre des Protokolls zu sagen. Die Gegenprobe hat es
	// gezeigt — sie blieb grün.
	<-losgelassen

	// **Und jetzt hält es die geteilte Sperre wirklich** — daran muss ein
	// Alleinanspruch von außen scheitern.
	pruefer, oeffnungsfehler := dateisperre.Oeffnen(filepath.Join(ort, app.Protokollsperre))
	if oeffnungsfehler != nil {
		t.Fatalf("die Sperre prüfen: %v", oeffnungsfehler)
	}

	t.Cleanup(func() { _ = pruefer.Schliessen() }) //nolint:errcheck // Aufräumweg

	if pruefer.Allein() == nil {
		t.Error("das Protokoll hält keine Sperre — ein nächster Start dürfte " +
			"seine offene Datei drehen")
	}
}
