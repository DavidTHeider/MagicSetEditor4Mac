package sicherung_test

// **Eine fremde Sicherung bestimmt nicht, wie lange gerechnet wird** — QA-06c
// (#1976),
// Fläche 23.
//
// Der Kopf einer age-Sicherung nennt den Arbeitsfaktor der Schlüsselableitung
// selbst, und `age.NewScryptIdentity` nimmt ohne Deckel bis 2^22 an — „15s on
// a modern machine" (`scrypt.go`, Zeile 129, age v1.3.2) und mit r = 8 vier
// GiB Speicher, **bevor** das Kennwort geprüft ist. Diese Anwendung schreibt
// 2^18, ebenso `age -p`. Was darüber liegt, stammt von keinem der beiden.

import (
	"bytes"
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// scryptzeile ist die Zeile im age-Kopf, die Salz und Arbeitsfaktor nennt.
var scryptzeile = regexp.MustCompile(`(?m)^-> scrypt (\S+) (\d+)$`)

func TestEinZuHoherArbeitsfaktorWirdNichtGerechnet(t *testing.T) {
	ziel, _ := verschluesselteSicherung(t)

	roh, fehler := os.ReadFile(ziel) //nolint:gosec // die Datei dieses Laufs
	if fehler != nil {
		t.Fatal(fehler)
	}

	if !scryptzeile.Match(roh) {
		t.Fatalf("kein scrypt im Kopf:\n%q", roh[:min(len(roh), 200)])
	}

	// **2^19 und nicht 2^22**: ohne Deckel rechnet der rote Lauf wirklich,
	// und eine Sekunde ist ein Lauf, vier GiB wären eine Maschine.
	verlangt := scryptzeile.ReplaceAll(roh, []byte("-> scrypt $1 19"))
	if bytes.Equal(verlangt, roh) {
		t.Fatal("der Arbeitsfaktor ließ sich nicht ändern")
	}

	fehler = os.WriteFile(ziel, verlangt, 0o600) //nolint:gosec // die Datei dieses Laufs
	if fehler != nil {
		t.Fatal(fehler)
	}

	_, fehler = sicherung.EntpackenVerschluesselt(ziel, t.TempDir(), probekennwort)

	var feld *domain.FieldError
	if errors.As(fehler, &feld) && feld.Field == "kennwort" {
		t.Fatalf("der Faktor wurde gerechnet und dann das Kennwort verworfen: %v", fehler)
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet eine Abweisung, bekam %v", fehler)
	}
}
