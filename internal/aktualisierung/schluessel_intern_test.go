// **Was geschieht, wenn der eingebettete Schlüssel nicht taugt** —
// #892.
//
// # Warum innen und nicht daneben
//
// Weil der Fall von außen nicht herzustellen ist: `freigabeschluessel` kommt
// aus `go:embed`, und im Betrieb steht dort eine gültige Datei. Der Zweig, der
// ihr Fehlen meldet, wäre sonst eine **Behauptung** darüber, was im Ernstfall
// geschieht (§3).
//
// **Und der Ernstfall ist real**: wer die Datei versehentlich leert, bekäme
// sonst eine Anwendung, die jede Signatur mit einem unverständlichen Fehler
// abweist — oder schlimmer, gar keinen.
package aktualisierung

import (
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

// mitSchluesseltext tauscht den eingebetteten Schlüssel für die Dauer des Tests.
func mitSchluesseltext(t *testing.T, text string) {
	t.Helper()

	vorher := freigabeschluessel
	freigabeschluessel = text

	t.Cleanup(func() { freigabeschluessel = vorher })
}

// **Ein unlesbarer Schlüssel meldet sich, statt zu schweigen.**
func TestEinUnlesbarerSchluesselMeldetSich(t *testing.T) {
	mitSchluesseltext(t, "das ist kein Schlüssel")

	if _, err := Schluesselfingerabdruck(); err == nil {
		t.Error("ein unlesbarer Schlüssel kam durch")
	}

	if err := SignaturPruefen([]byte("egal"), []byte("egal")); err == nil {
		t.Error("die Prüfung ohne Schlüssel kam durch")
	}
}

// **Und ein Bund mit zwei Schlüsseln, die keine benannten Anker sind, ebenso.**
//
// **Weil dann nicht feststeht, wer freigibt.** Ein Bund, der wächst, weil
// jemand einen zweiten Schlüssel dazulegte, ist genau die Änderung, die
// niemand bemerkt — und sie erlaubte einer zweiten Hand, Freigaben zu
// unterschreiben. **Seit SEC03 (#2356)** dürfen zwei im Bund stehen, aber
// nur als benannte Anker ([freigabeanker]); die Matrix dazu steht in
// `anker_intern_test.go`.
func TestEinBundMitZweiSchluesselnWirdAbgewiesen(t *testing.T) {
	// **Zwei Schlüssel in **einem** Panzerblock.**
	//
	// Zwei Blöcke hintereinander tun es nicht: [openpgp.ReadArmoredKeyRing]
	// liest den **ersten** und übergeht den Rest — der erste Anlauf hängte
	// deshalb den Block an sich selbst und war grün, ohne etwas zu prüfen.
	mitSchluesseltext(t, zweiInEinemBlock(t))

	if _, err := Schluesselfingerabdruck(); err == nil {
		t.Error("zwei Schlüssel kamen durch")
	}
}

// zweiInEinemBlock gibt zwei frische Schlüssel in einem Panzerblock.
func zweiInEinemBlock(t *testing.T) string {
	t.Helper()

	var gepanzert strings.Builder

	schreiber, err := armor.Encode(&gepanzert, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("panzern: %v", err)
	}

	for _, name := range []string{"Eine", "Andere"} {
		wer, fehler := openpgp.NewEntity(name, "für Tests", name+"@beispiel.test", nil)
		if fehler != nil {
			t.Fatalf("Schlüssel erzeugen: %v", fehler)
		}

		if fehler := wer.Serialize(schreiber); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return gepanzert.String()
}

// **Der Nachweis, dass die Probe etwas unterscheidet.**
//
// Ohne ihn stünde nur fest, dass die beiden Fälle rot sind — nicht, dass der
// echte Schlüssel grün ist (§10).
func TestDerEchteSchluesselLiestSich(t *testing.T) {
	abdruck, err := Schluesselfingerabdruck()
	if err != nil {
		t.Fatalf("den echten Schlüssel lesen: %v", err)
	}

	if !strings.HasSuffix(abdruck, Signaturschluessel) {
		t.Errorf("der Fingerabdruck %s endet nicht auf %s", abdruck, Signaturschluessel)
	}
}
