package domain_test

// **Wer am Zug ist, weiß am Tisch jeder** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
// **Dass der Zugstand durch die Projektion geht und ungekürzt ankommt.** Er
// ist **öffentlich**: CR 102.1 nennt den aktiven Spieler, CR 117 die
// Priorität, und beides ist am Tisch für jeden sichtbar. Ein Sitz, der nicht
// wüsste, wer am Zug ist, könnte nicht passen.
//
// **Und dass er auch den Zuschauer erreicht**: er sieht keine Hand und keine
// Bibliothek, aber er sieht, wessen Zug läuft — sonst sähe er einem Spiel zu,
// dessen Takt er nicht kennt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerZugstandStehtInJederSicht(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.ID = "p1"
	partie = partie.ZugBeginnen("sitz-1")

	for _, fuer := range []string{"sitz-1", "sitz-2", "", "zuschauer"} {
		sicht := partie.Sicht(fuer)

		switch {
		case sicht.Zug.Nummer != 1:
			t.Errorf("für %q ist es Zug %d", fuer, sicht.Zug.Nummer)
		case sicht.Zug.Aktiv != "sitz-1":
			t.Errorf("für %q ist %q aktiv", fuer, sicht.Zug.Aktiv)
		case sicht.Zug.Abschnitt.Schritt != domain.SchrittVersorgung:
			t.Errorf("für %q steht es bei %+v", fuer, sicht.Zug.Abschnitt)
		case sicht.Zug.Prioritaet != "sitz-1":
			t.Errorf("für %q hat %q Priorität", fuer, sicht.Zug.Prioritaet)
		}
	}
}

// **Eine Partie ohne Zugbegriff gibt eine Null und keine erfundene Eins.**
//
// **Das ist die Zusage an die Partien aus G1 bis G8** (§15, Punkt 4): sie
// sind Sandkästen, in denen jeder jederzeit alles darf, und sie sollen
// weiterlaufen. Eine Eins hier stellte sie in den Enttappen-Schritt.
func TestEineSandkastenpartieHatKeinenZug(t *testing.T) {
	t.Parallel()

	sicht := einePartie().Sicht("sitz-1")

	switch {
	case sicht.Zug.Nummer != 0:
		t.Errorf("es ist Zug %d", sicht.Zug.Nummer)
	case sicht.Zug.Prioritaet != "":
		t.Errorf("Priorität hat %q", sicht.Zug.Prioritaet)
	}
}
