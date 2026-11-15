//go:build darwin

package schluesselbund_test

// **Kein Dialog** — gemeldet am 2026-09-02 vom Nutzer dieses Rechners.
//
// # Der Befund
//
// Beim Lauf erschien ein Fenster von macOS:
//
//	Schlüsselbund nicht gefunden
//	Es wurde kein Schlüsselbund gefunden, um „konto" zu sichern.
//	[Abbrechen] [Auf Standard zurücksetzen]
//
// **„konto" ist der Kontoname aus den Tests dieses Pakets.** Sie schreiben in
// einen Bund, den es absichtlich nicht gibt — und `security` fragt dann den
// Menschen, statt zu antworten.
//
// # Warum das mehr ist als ein Ärgernis im Testlauf
//
// **Dasselbe geschähe im Betrieb.** Wer einen Schlüssel hinterlegt, während
// sein Anmeldebund fehlt oder umbenannt wurde, bekäme einen Dialog von einem
// Programm, das gerade eine Einstellung speichern wollte — mit einer Frage,
// die er nicht gestellt hat, und einer Antwort („Auf Standard zurücksetzen"),
// deren Folgen niemand erklärt hat.
//
// **Ein Programm, das im Hintergrund einen Systemdialog auslöst, hat die
// Kontrolle abgegeben.**
//
// # Und der Dialog ist nicht einmal das schlimmste
//
// **Bei der Gegenprobe kam gar kein Dialog mehr** — `security` schrieb das
// Geheimnis **stillschweigend in den Anmeldebund** und meldete Erfolg.
// Nachgeprüft und wieder entfernt.
//
// Deshalb prüft dieser Lauf **zwei** Dinge: dass kein Unterprozess startet
// (an der Zeit) **und** dass die Antwort ein Nein ist. Ein Lauf, der nur die
// Zeit misst, ginge auch dann durch, wenn `security` schnell und falsch
// antwortet.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// **Ein Bund, den es nicht gibt, wird erkannt, bevor `security` läuft.**
//
// Der Beleg dafür ist die **Zeit**: der Dialog wartet auf einen Menschen, und
// die Frist von zwei Sekunden lief vorher ab. Wer gar nicht erst startet, ist
// sofort zurück.
func TestEinFehlenderBundLoestKeinenDialogAus(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "gibtsnicht.keychain")

	bund := schluesselbund.NeuerBund(ort).MitFrist(5 * time.Second)

	begonnen := time.Now()
	err := bund.Schreibe(t.Context(), "mse", "konto", "geheim")
	gedauert := time.Since(begonnen)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Fehler ist %v, erwartet ErrNotFound", err)
	}

	// **Hundert Millisekunden sind großzügig** für ein `os.Stat`. Ein
	// gestarteter `security` mit Dialog bräuchte die volle Frist.
	if gedauert > 100*time.Millisecond {
		t.Errorf("das Schreiben dauerte %v — wurde security doch gestartet?",
			gedauert)
	}

	// **Und die Meldung sagt, was fehlt** — nicht „Zugriff verweigert".
	if !strings.Contains(err.Error(), ort) {
		t.Errorf("die Meldung %q nennt den Bund nicht", err)
	}
}

// **Und ein vorhandener Bund wird beschrieben** — sonst prüfte der Lauf oben
// nur, dass immer abgelehnt wird.
func TestEinVorhandenerBundWirdBeschrieben(t *testing.T) {
	t.Parallel()

	bund := probebund(t)

	if err := bund.Schreibe(t.Context(), "mse", "konto", "geheim"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// **Lesen und Löschen lösen ebenso keinen Dialog aus.**
//
// Gemessen am 2026-09-02: sie antworten mit „could not be found" statt zu
// fragen. Der Lauf hält das fest — änderte sich das, wäre es dieselbe Falle.
func TestLesenUndLoeschenFragenNichtNach(t *testing.T) {
	t.Parallel()

	bund := schluesselbund.NeuerBund(
		filepath.Join(t.TempDir(), "gibtsnicht.keychain")).
		MitFrist(5 * time.Second)

	begonnen := time.Now()

	if _, err := bund.Lies(t.Context(), "mse", "konto"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Lies gab %v, erwartet ErrNotFound", err)
	}

	if err := bund.Loesche(t.Context(), "mse", "konto"); err != nil {
		t.Errorf("Loesche gab %v, erwartet nichts", err)
	}

	if gedauert := time.Since(begonnen); gedauert > time.Second {
		t.Errorf("Lesen und Löschen dauerten %v — kam ein Dialog?", gedauert)
	}
}

// **Der Anmeldebund wird nicht geprüft — und das ist gewollt.**
//
// Seinen Pfad kennt nur macOS; ihn zu erfragen wäre ein zweiter Unterprozess
// für einen Fall, den es auf einem eingerichteten Rechner nicht gibt.
//
// **Dieser Lauf hält fest, dass die Prüfung ihn durchlässt** — sonst wäre der
// Anmeldebund unbenutzbar, und niemand sähe es, weil kein Test ihn anfasst
// (und keiner darf, §3).
func TestDerAnmeldebundWirdNichtGeprueft(t *testing.T) {
	t.Parallel()

	// **Ein leerer Ort ist der Anmeldebund.** Geprüft wird hier nur, dass die
	// Ortsprüfung ihn durchlässt — geschrieben wird **nicht**.
	if fehlt := schluesselbund.Anmeldebund().OrtFehlt(); fehlt != nil {
		t.Errorf("der Anmeldebund gilt als fehlend: %v", fehlt)
	}

	// **Und ein benannter, den es nicht gibt, nicht.**
	fehlend := schluesselbund.NeuerBund(
		filepath.Join(t.TempDir(), "gibtsnicht.keychain"))

	if fehlend.OrtFehlt() == nil {
		t.Error("ein fehlender Bund gilt als vorhanden")
	}
}
