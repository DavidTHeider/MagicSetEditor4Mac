package app_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Unter »nur lokal« nennt die Absage die Regel und nicht den Schalter.**
//
// # Warum das zählt
//
// **Weil sie sonst jemanden an den falschen Schalter schickt** (§8). Der
// Schalter des Anbieters steht auf `an`; wer die Meldung „ist nicht
// eingeschaltet" liest, legt ihn um und kommt trotzdem nicht weiter.
//
// **Der Riegel am Transport sagt es richtig** — und er kommt nie dran: die
// Wirksamkeitsschranke liegt davor, und `quelleFuer` liest nur den
// **wirksamen** Wert. Gefunden vom Flächenlauf zur Null-Netz-Probe (A14,
// #1453).
func TestUnterNurLokalNenntDieAbsageDieRegel(t *testing.T) {
	welt := einekiwelt(t)

	einstellungen := storage.NeueEinstellungen(welt.db)
	for _, paar := range [][2]string{
		{domain.DienstEinstellung(domain.DienstAnthropic), domain.DienstAn},
		{domain.RichtlinieNurLokal, domain.NurLokalAn},
	} {
		if fehler := einstellungen.Setzen(probekontext(t),
			paar[0], paar[1]); fehler != nil {
			t.Fatalf("%s setzen: %v", paar[0], fehler)
		}
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"anthropic", "claude-probe", "Schreibe den Regeltext.", "")
	if antwort.Fehler == nil {
		t.Fatal("der Ruf kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "nur lokal") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}
