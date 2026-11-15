// **Die Signaturmatrix** — SEC03
// (#2356),
// Vermerk 0194: gültig, alt, neu, fremd, widerrufen, abgelaufen.
//
// # Warum innen
//
// Die Matrix prüft den **eingebetteten** Weg — `freigabeschluessel` und
// `freigabeanker`, wie sie im Programm stehen — und nicht die Naht
// [MitFreigabebund], die ihn für andere Tests überspringt. Eine Vorrichtung,
// die den Bund ersetzt, prüfte nicht, welche Schlüssel er gelten lässt.
//
// **Jeder Schlüssel ist ein frisch erzeugter.** Der echte Freigabeschlüssel
// steht in keinem Fall; dass der eingebettete der richtige ist, hält
// `signatur_test.go` eigens.
package aktualisierung

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// probeschluesselNamens erzeugt einen Schlüssel — mit `beginn` als
// Erzeugungszeit und `dauer` als Lebensdauer, wenn sie nicht null sind.
func probeschluesselNamens(t *testing.T, name string, beginn time.Time, dauer time.Duration) *openpgp.Entity {
	t.Helper()

	einstellung := &packet.Config{}
	if !beginn.IsZero() {
		einstellung.Time = func() time.Time { return beginn }
		einstellung.KeyLifetimeSecs = uint32(dauer / time.Second)
	}

	wer, err := openpgp.NewEntity(name, "für Tests", strings.ToLower(name)+"@beispiel.test", einstellung)
	if err != nil {
		t.Fatalf("Schlüssel %s erzeugen: %v", name, err)
	}

	return wer
}

// bundAus panzert die öffentlichen Teile in **einen** Block, wie
// `gpg --armor --export` mehrere Schlüssel schreibt.
func bundAus(t *testing.T, schluessel ...*openpgp.Entity) string {
	t.Helper()

	var gepanzert strings.Builder

	schreiber, err := armor.Encode(&gepanzert, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("panzern: %v", err)
	}

	for _, wer := range schluessel {
		if fehler := wer.Serialize(schreiber); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return gepanzert.String()
}

// signiertMit gibt eine abgetrennte Signatur — zur Zeit `wann`, wenn sie
// nicht null ist.
func signiertMit(t *testing.T, wer *openpgp.Entity, text string, wann time.Time) string {
	t.Helper()

	var einstellung *packet.Config
	if !wann.IsZero() {
		einstellung = &packet.Config{Time: func() time.Time { return wann }}
	}

	var gepanzert strings.Builder
	if err := openpgp.ArmoredDetachSign(&gepanzert, wer, strings.NewReader(text), einstellung); err != nil {
		t.Fatalf("signieren: %v", err)
	}

	return gepanzert.String()
}

// abdruckVon gibt den Fingerabdruck, wie [Schluesselfingerabdruck] ihn schreibt.
func abdruckVon(wer *openpgp.Entity) string {
	return strings.ToUpper(fmt.Sprintf("%X", wer.PrimaryKey.Fingerprint))
}

// mitAnkern setzt Bund und Anker für die Dauer des Tests.
func mitAnkern(t *testing.T, bund string, anker ...*openpgp.Entity) {
	t.Helper()

	mitSchluesseltext(t, bund)

	vorher := freigabeanker
	freigabeanker = nil

	for _, wer := range anker {
		freigabeanker = append(freigabeanker, abdruckVon(wer))
	}

	t.Cleanup(func() { freigabeanker = vorher })
}

//nolint:paralleltest // der Bund ist eine Paketvariable
func TestDieSignaturmatrix(t *testing.T) {
	alt := probeschluesselNamens(t, "Alt", time.Time{}, 0)
	neu := probeschluesselNamens(t, "Neu", time.Time{}, 0)
	fremd := probeschluesselNamens(t, "Fremd", time.Time{}, 0)

	text := "0000  ./MagicSetEditor4Mac-1.0.0.dmg\n"

	// **Signiert vor dem Widerruf** — mit einem widerrufenen Schlüssel
	// signiert go-crypto gar nicht erst („no valid signing keys"); die
	// Signatur, die geprüft wird, ist die, die vorher entstand.
	widerrufen := probeschluesselNamens(t, "Widerrufen", time.Time{}, 0)
	vorDemWiderruf := signiertMit(t, widerrufen, text, time.Time{})

	if err := widerrufen.RevokeKey(packet.KeyCompromised, "Probe", nil); err != nil {
		t.Fatalf("widerrufen: %v", err)
	}

	vorEinemJahr := time.Now().AddDate(-1, 0, 0)
	abgelaufen := probeschluesselNamens(t, "Abgelaufen", vorEinemJahr, time.Hour)

	faelle := []struct {
		name      string
		bund      []*openpgp.Entity
		anker     []*openpgp.Entity
		signierer *openpgp.Entity
		wann      time.Time
		gilt      bool
	}{
		{"gültig: der eine Anker", []*openpgp.Entity{alt}, []*openpgp.Entity{alt}, alt, time.Time{}, true},
		{"im Übergang: der alte Anker", []*openpgp.Entity{alt, neu}, []*openpgp.Entity{alt, neu}, alt, time.Time{}, true},
		{"im Übergang: der neue Anker", []*openpgp.Entity{alt, neu}, []*openpgp.Entity{alt, neu}, neu, time.Time{}, true},
		{"nach dem Übergang: der alte", []*openpgp.Entity{neu}, []*openpgp.Entity{neu}, alt, time.Time{}, false},
		{"falsch: ein fremder Schlüssel", []*openpgp.Entity{alt}, []*openpgp.Entity{alt}, fremd, time.Time{}, false},
		{"widerrufen: er selbst", []*openpgp.Entity{neu, widerrufen}, []*openpgp.Entity{neu}, widerrufen, time.Time{}, false},
		{"widerrufen: der Anker daneben", []*openpgp.Entity{neu, widerrufen}, []*openpgp.Entity{neu}, neu, time.Time{}, true},
		{"abgelaufen", []*openpgp.Entity{abgelaufen}, []*openpgp.Entity{abgelaufen}, abgelaufen, vorEinemJahr.Add(time.Minute), false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			mitAnkern(t, bundAus(t, fall.bund...), fall.anker...)

			signatur := vorDemWiderruf
			if fall.signierer != widerrufen {
				signatur = signiertMit(t, fall.signierer, text, fall.wann)
			}

			err := SignaturPruefen([]byte(text), []byte(signatur))
			if gilt := err == nil; gilt != fall.gilt {
				t.Errorf("angenommen: %v, erwartet %v (%v)", gilt, fall.gilt, err)
			}
		})
	}
}

// **Ein Bund, der nicht genau die Anker führt, gilt gar nicht** — auch nicht
// für die Signatur eines Ankers in ihm.
//
//nolint:paralleltest // der Bund ist eine Paketvariable
func TestEinBundMussGenauDieAnkerFuehren(t *testing.T) {
	anker := probeschluesselNamens(t, "Anker", time.Time{}, 0)
	zweiter := probeschluesselNamens(t, "Zweiter", time.Time{}, 0)

	widerrufen := probeschluesselNamens(t, "Widerrufen", time.Time{}, 0)
	if err := widerrufen.RevokeKey(packet.KeySuperseded, "Probe", nil); err != nil {
		t.Fatalf("widerrufen: %v", err)
	}

	text := "0000  ./MagicSetEditor4Mac-1.0.0.dmg\n"

	faelle := map[string]struct {
		bund  []*openpgp.Entity
		anker []*openpgp.Entity
	}{
		"ein Schlüssel, der kein Anker ist": {[]*openpgp.Entity{anker, zweiter}, []*openpgp.Entity{anker}},
		"ein Anker, der im Bund fehlt":      {[]*openpgp.Entity{anker}, []*openpgp.Entity{anker, zweiter}},
		"ein Anker, der widerrufen ist":     {[]*openpgp.Entity{anker, widerrufen}, []*openpgp.Entity{anker, widerrufen}},
		"gar kein Anker":                    {[]*openpgp.Entity{anker}, nil},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			mitAnkern(t, bundAus(t, fall.bund...), fall.anker...)

			if err := SignaturPruefen([]byte(text), []byte(signiertMit(t, anker, text, time.Time{}))); err == nil {
				t.Error("die Signatur des Ankers galt in einem Bund, der nicht genau die Anker führt")
			}
		})
	}
}

// **Ohne einen Anker, der auf den Signaturschlüssel endet, gibt es keinen
// Fingerabdruck** — im Übergang hieße das: `release.sh` signierte mit einem
// Schlüssel, den der Bund nicht führt.
//
//nolint:paralleltest // der Bund ist eine Paketvariable
func TestOhneDenSignaturschluesselKeinFingerabdruck(t *testing.T) {
	fremd := probeschluesselNamens(t, "Fremd", time.Time{}, 0)
	mitAnkern(t, bundAus(t, fremd), fremd)

	if abdruck, err := Schluesselfingerabdruck(); err == nil {
		t.Errorf("ein Bund ohne den Signaturschlüssel nannte %s", abdruck)
	}
}
