package main

// **Die Liste der Partien sagt in Worten, ob man dran ist** — UXR-F16
// (#2031).
//
// **Gerechnet hat es `service`**; hier steht nur, wie die Zeile es nennt. Dass
// es an Dienst und MCP dasselbe ist, hält `internal/flaechen/partiendran_test.go`.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Zeile nennt den eigenen Sitz und, ob die Partie auf ihn wartet.**
//
// **Und wo das Profil nicht sitzt, bleibt sie, wie sie war**: wer die Zeilen
// ausschneidet, findet an der alten Stelle dasselbe (§16, erweiternd).
func TestDieTextzeileNenntDenEigenenSitzUndObManDranIst(t *testing.T) {
	t.Parallel()

	sitze := []app.ZeilensitzDTO{
		{Kennung: "anna", Name: "Anna"}, {Kennung: "bert", Name: "Bert"},
	}

	var raus, rfehler bytes.Buffer

	an := &schreiber{ziel: &raus, fehler: &rfehler}

	code := partienAusgeben(an, app.PartienAntwort{Partien: []app.PartiezeileDTO{
		{ID: "p1", Stand: "laeuft", Objekte: 4, Sitze: sitze, EigenerSitz: "bert", Dran: true},
		{ID: "p2", Stand: "laeuft", Objekte: 4, Sitze: sitze, EigenerSitz: "bert"},
		{ID: "p3", Stand: "laeuft", Objekte: 4, Sitze: sitze},
	}})
	if code != CodeGut {
		t.Fatalf("Code %d (%s)", code, rfehler.String())
	}

	zeilen := strings.Split(strings.TrimSpace(raus.String()), "\n")
	grund := " — laeuft, 4 Objekte, Sitze: anna (Anna), bert (Bert)"

	for stelle, erwartet := range []string{
		"p1" + grund + ", dein Sitz: bert, dran",
		"p2" + grund + ", dein Sitz: bert",
		"p3" + grund,
	} {
		if len(zeilen) != 3 || zeilen[stelle] != erwartet {
			t.Errorf("Zeile %d: %q, erwartet %q", stelle+1, zeilen, erwartet)
		}
	}
}
