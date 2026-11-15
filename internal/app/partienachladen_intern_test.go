package app

// **Ein Nachladefehler ist kein gescheiterter Partieschritt** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ein Auslöser im echten Bestand** macht nach dem Anlegen oder nach einem
// Wechsel des Stands den Zustand der Partie zu etwas, das kein JSON ist
// ([zustandVerbiegen]). Geschrieben ist dann; die Sicht danach scheitert. Bis
// #1970 kam das als `fehler` zurück — **und beim Anlegen ohne Kennung**: wer
// es daraufhin noch einmal versuchte, legte eine zweite Partie an.

import (
	"testing"
)

// partieNachgeladen prüft eine Antwort, deren Schritt galt und deren Sicht
// sich danach nicht lesen ließ.
func partieNachgeladen(t *testing.T, schritt string, antwort PartieAntwort) {
	t.Helper()

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("%s: als gescheitert gemeldet: %+v", schritt, antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("%s: der Nachladefehler fehlt: %+v", schritt, antwort)
	case antwort.Partie.Sitze == nil || antwort.Erlaubt == nil:
		t.Fatalf("%s: die Listen sind nil: %+v", schritt, antwort)
	}
}

func TestEinNachladefehlerNachDemAnlegenNenntDiePartie(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)
	deckID := einDeckImBestand(t, db)

	if _, fehler := db.ExecContext(ctx, `CREATE TRIGGER krumme_partie
AFTER INSERT ON partien
BEGIN UPDATE partien SET zustand = 'krumm' WHERE id = NEW.id; END`); fehler != nil {
		t.Fatalf("der Auslöser: %v", fehler)
	}

	antwort := PartieflaecheUeber(db).PartieAusWuenschen(ctx, "magic", "anna",
		[]Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: deckID},
			{Kennung: "bert", Name: "Bert", DeckID: deckID},
		})
	partieNachgeladen(t, "anlegen", antwort)

	// **Kennung und Sicht gehen trotzdem mit** — ohne sie fände niemand die
	// Partie wieder, die eben entstanden ist, und das Fenster wüsste nicht,
	// für wen es nachliest.
	if antwort.Partie.Fuer != "anna" {
		t.Errorf("die Sicht ist für %q", antwort.Partie.Fuer)
	}

	var stand string
	if fehler := db.QueryRowContext(ctx, `SELECT stand FROM partien WHERE id = ?`,
		antwort.Partie.ID).Scan(&stand); fehler != nil || stand != "laeuft" {
		t.Fatalf("die Partie %q: %q (%v)", antwort.Partie.ID, stand, fehler)
	}
}

func TestEinNachladefehlerIstKeinGescheiterterStandwechsel(t *testing.T) {
	t.Parallel()

	anwendung, partieID := einePartieZumOffenlegen(t)
	ctx := riegelkontext(t)

	for _, schritt := range []struct {
		name, stand string
		weg         func() PartieAntwort
	}{
		{"pausieren", "pausiert", func() PartieAntwort {
			return anwendung.PartiePausieren(ctx, partieID, "anna")
		}},
		{"fortsetzen", "laeuft", func() PartieAntwort {
			return anwendung.PartieFortsetzen(ctx, partieID, "anna")
		}},
		{"beenden", "beendet", func() PartieAntwort {
			return anwendung.PartieBeenden(ctx, partieID, "anna")
		}},
	} {
		heilen := zustandVerbiegen(t, anwendung.db, partieID, `AFTER UPDATE OF stand ON partien`)
		antwort := schritt.weg()
		heilen()

		partieNachgeladen(t, schritt.name, antwort)

		if antwort.Partie.ID != partieID || antwort.Partie.Fuer != "anna" {
			t.Errorf("%s: die Partie heißt %q, gesehen von %q",
				schritt.name, antwort.Partie.ID, antwort.Partie.Fuer)
		}

		var stand string
		if fehler := anwendung.db.QueryRowContext(ctx,
			`SELECT stand FROM partien WHERE id = ?`, partieID).Scan(&stand); fehler != nil ||
			stand != schritt.stand {
			t.Fatalf("%s: der Stand ist %q (%v)", schritt.name, stand, fehler)
		}
	}

	// **Ein gescheiterter Schritt bleibt einer** — eine beendete Partie geht
	// nicht weiter, und das Nein steht in `fehler`.
	if antwort := anwendung.PartieFortsetzen(ctx, partieID, "anna"); antwort.Fehler == nil ||
		antwort.Nachladefehler != nil {
		t.Fatalf("eine beendete fortgesetzt: %+v", antwort)
	}
}
