package app

// **Ein Sitz, der mit einer Leihe an den Tisch geht** — Werkstück G6 von
// Phase 5.5 (#1463).
//
// **Eine eigene Datei**, damit `partie_intern_test.go` unter der Zeilengrenze
// aus §2 bleibt — und weil es eine eigene Frage ist: nicht „wie entsteht eine
// Partie", sondern „womit tritt jemand an".

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Sitz mit Leihe braucht keine Deckkennung** — Werkstück G6 von Phase
// 5.5 (#1463).
//
// **Und wer trotzdem eine krumme schickt, bekommt eine Meldung** — an einem
// Sitz **ohne** Leihe. Beides in einem Lauf, denn es ist dieselbe Weiche.
func TestEinSitzMitLeiheBrauchtKeineDeckkennung(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPeerbund(&peerprobebund{})

	flaeche := PartieflaecheUeber(anwendung.db)

	// **Ohne angenommene Leihe endet es an der Leihe und nicht an der
	// Deckkennung** — genau das belegt, dass die Weiche greift.
	antwort := flaeche.PartieAusWuenschen(riegelkontext(t), "magic", "anna",
		[]Sitzwunsch{{Kennung: "anna", Name: "Anna", LeiheID: "l1"}})

	if antwort.Fehler == nil {
		t.Fatal("die Leihe, die es nicht gibt, kam durch")
	}

	if antwort.Fehler.Meldung == "" {
		t.Error("die Meldung ist leer")
	}
}

// **Ein Wunsch, den der Wirt nicht bekommt, wird gemeldet** — Werkstück G5
// von Phase 5.5
// (#1462).
//
// **Beide Hälften**: erst scheitert das Einreihen, dann das Leeren. Ein Gast,
// dessen Wunsch still verschwände, säße vor einem Tisch, der nicht reagiert —
// und das ist schlimmer als eine Leitung, die ehrlich klagt (§8).
func TestEinWunschDerNichtHinausgehtWirdGemeldet(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*angenommenendoppel){
		"das Einreihen scheitert": func(d *angenommenendoppel) {
			d.sendefehler = errNichtHinaus
		},
		"das Leeren scheitert": func(d *angenommenendoppel) {
			d.leerfehler = errNichtHinaus
		},
	}

	for name, richte := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := freigabeanwendung(t)
			anwendung.melden = nil

			sitzung := eineAngenommene(t, domain.ZweckPartie)
			richte(sitzung)

			anwendung.beitrittMerken(sitzung, "p1")

			antwort := anwendung.PartieWuenschen(t.Context(),
				domain.Spielbefehl{
					Kennung: "b1", Absender: "bert", Art: "objekt.tappen",
					Nutzlast: map[string]any{"objekt": "o1"},
				})

			if antwort.Fehler == nil {
				t.Fatal("der verlorene Wunsch kam nicht an")
			}
		})
	}
}

// errNichtHinaus ist der Grund, den die Leitung in diesen Läufen gibt.
var errNichtHinaus = errors.New("die Leitung ist zu")
