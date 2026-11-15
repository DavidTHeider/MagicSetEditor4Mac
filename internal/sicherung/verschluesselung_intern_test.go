package sicherung

import (
	"errors"
	"io"
	"testing"

	"filippo.io/age"
)

type kaputterSchreiber struct{}

func (kaputterSchreiber) Write([]byte) (int, error) { return 0, errors.New("die Platte ist voll") }

// **Scheitert schon der Kopf, wird nichts geschrieben** — und gemeldet.
func TestDieChiffrehuelleMeldetEinenKaputtenKopf(t *testing.T) {
	empfaenger, fehler := age.NewScryptRecipient("kennwort")
	if fehler != nil {
		t.Fatal(fehler)
	}

	empfaenger.SetWorkFactor(1)

	geschrieben := false

	fehler = chiffrehuelle(empfaenger)(kaputterSchreiber{}, func(io.Writer) error {
		geschrieben = true

		return nil
	})
	if fehler == nil || geschrieben {
		t.Errorf("Fehler %v, geschrieben %v", fehler, geschrieben)
	}
}
