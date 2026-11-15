package app

// **Ein Zug erreicht den Tisch — und ein toter Draht hält ihn nicht auf** —
// #1627.
//
// # Wogegen hier geprüft wird
//
// **Gegen die Versuchung, eine Zustellung zur Bedingung zu machen.** Der
// Befehl hat gewirkt: er steht im Bestand, und der Strom hat seine Folge.
// Eine Leitung, die dabei zugegangen ist, ist ein Zustellproblem — wer
// daraus einen Fehler machte, nähme dem Absender die Antwort auf einen Zug,
// den er längst getan hat (§7, Regel 3).

import (
	"errors"
	"strconv"
	"testing"
)

// TestEinToterDrahtNimmtDemZugNichtSeineAntwort hält genau das.
func TestEinToterDrahtNimmtDemZugNichtSeineAntwort(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	anwendung := MitPartienleitung(NeueApp(""), db)
	anwendung.db = db
	anwendung.profil.Store(1)

	gebunden := NeueBindungen(anwendung)
	deckID := einBaerImBestand(t, db)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	// **Erst der Beginn, bevor der Gast sich setzt** (#1851): gefragt ist
	// nach dem Zug danach und nicht danach, wie die Starthand zugestellt wird.
	beginnAbschliessen(t, ueberDieBindungen(t.Context(), gebunden), angelegt.Partie)

	erstes := inDerBibliothek(t, gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1").
		Partie.Objekte, "sitz-1")

	// **Ein Gast sitzt am Tisch, und seine Leitung ist tot.**
	griff := eineAngenommene(t, "partie")
	griff.sendefehler = errors.New("die Leitung ist weg")
	anwendung.partie.runde.Setze(einGastplatz(griff))

	gestellt := gebunden.PartieBefehl(t.Context(), angelegt.Partie.ID, "b1", "sitz-1",
		"objekt.schieben", map[string]any{"objekt": erstes, "zone": "tisch"}, "")
	if gestellt.Fehler != nil {
		t.Fatalf("der tote Draht hat den Zug gekippt: %+v", gestellt.Fehler)
	}

	// **Und der Zug hat wirklich gewirkt** — sonst prüfte dieser Lauf nur,
	// dass ein Fehlschlag keinen Fehler gibt (§10).
	if gestellt.Fassung == "" {
		t.Fatal("die Antwort nennt keine Fassung")
	}

	if fassung, fehler := strconv.ParseInt(
		gestellt.Fassung, 10, 64); fehler != nil || fassung < 1 {
		t.Errorf("die Fassung ist %q", gestellt.Fassung)
	}

	if len(gestellt.Ereignisse) == 0 {
		t.Error("der Zug hat kein Ereignis erzeugt")
	}

	// **Der Platz steht noch** — ein Zustellfehler räumt ihn nicht weg; das
	// tut die Schleife, wenn die Leitung wirklich endet
	// ([service.Partierunde.Geht]).
	if sitze := anwendung.partie.runde.Sitze(); len(sitze) != 1 {
		t.Errorf("am Tisch sitzen %v", sitze)
	}
}
