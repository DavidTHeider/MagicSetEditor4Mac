package app

// **Ein Nachladefehler ist keine gescheiterte Offenlegung** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ein Auslöser im echten Bestand** macht nach dem Schreiben den Zustand
// der Partie zu etwas, das kein JSON ist. Das Schreiben liest die Partie
// **vorher** und glückt; das Nachlesen danach liest sie wieder und
// scheitert. Bis #1970 kam das als `fehler` zurück — offengelegt oder
// zurückgenommen war trotzdem.

import (
	"database/sql"
	"testing"
)

// einePartieZumOffenlegen gibt eine Anwendung mit einer laufenden Partie.
func einePartieZumOffenlegen(t *testing.T) (anwendung *App, partieID string) {
	t.Helper()

	db := bestandFuerDenRiegel(t)
	anwendung = MitPartienleitung(NeueApp(""), db)
	anwendung.profil.Store(1)

	deckID := einDeckImBestand(t, db)

	partie := PartieflaecheUeber(db).PartieAusWuenschen(riegelkontext(t), "magic", "anna",
		[]Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: deckID},
			{Kennung: "bert", Name: "Bert", DeckID: deckID},
		})
	if partie.Fehler != nil {
		t.Fatalf("PartieAusWuenschen: %+v", partie.Fehler)
	}

	return anwendung, partie.Partie.ID
}

// zustandVerbiegen setzt einen Auslöser, der nach dem genannten Ereignis
// den Zustand der Partie verbiegt — und gibt zurück, wie man ihn heilt.
//
// **Verbogen wird durch ein Präfix und nicht durch Ersetzen**: der Schritt
// schreibt selbst einen Zustand, und wer ihn überschriebe, heilte danach
// zu dem von vorher — die Partie stünde dann in der Spalte `stand` anders
// als in ihrem Zustand.
func zustandVerbiegen(t *testing.T, db *sql.DB, partieID, ereignis string) func() {
	t.Helper()

	ctx := riegelkontext(t)

	if _, fehler := db.ExecContext(ctx, `CREATE TRIGGER krummer_zustand `+ereignis+`
BEGIN UPDATE partien SET zustand = 'krumm' || zustand
       WHERE zustand NOT LIKE 'krumm%'; END`); fehler != nil {
		t.Fatalf("der Auslöser: %v", fehler)
	}

	return func() {
		if _, fehler := db.ExecContext(ctx, `DROP TRIGGER krummer_zustand`); fehler != nil {
			t.Fatalf("den Auslöser entfernen: %v", fehler)
		}

		if _, fehler := db.ExecContext(ctx,
			`UPDATE partien SET zustand = substr(zustand, 6)
			  WHERE id = ? AND zustand LIKE 'krumm%'`, partieID); fehler != nil {
			t.Fatalf("den Zustand heilen: %v", fehler)
		}
	}
}

// offenlegungNachgeladen prüft eine Antwort, deren Schritt galt und deren
// Liste sich danach nicht lesen ließ.
func offenlegungNachgeladen(t *testing.T, schritt string, antwort PartieoffenlegungAntwort) {
	t.Helper()

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("%s: als gescheitert gemeldet: %+v", schritt, antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("%s: der Nachladefehler fehlt: %+v", schritt, antwort)
	case antwort.Offenlegungen == nil:
		t.Fatalf("%s: die Liste ist nil", schritt)
	}
}

// offengelegt zählt die Freigaben der Partie an der Tabelle.
func offengelegt(t *testing.T, db *sql.DB, partieID string) int {
	t.Helper()

	var anzahl int
	if fehler := db.QueryRowContext(riegelkontext(t),
		`SELECT count(*) FROM partiefreigaben WHERE partie_id = ?`, partieID,
	).Scan(&anzahl); fehler != nil {
		t.Fatalf("zählen: %v", fehler)
	}

	return anzahl
}

func TestEinNachladefehlerIstKeineGescheiterteOffenlegung(t *testing.T) {
	t.Parallel()

	anwendung, partieID := einePartieZumOffenlegen(t)
	ctx := riegelkontext(t)

	heilen := zustandVerbiegen(t, anwendung.db, partieID, `AFTER INSERT ON partiefreigaben`)
	offenlegungNachgeladen(t, "offenlegen",
		anwendung.PartieOffenlegen(ctx, partieID, "anna", []string{"bert"}))
	heilen()

	if anzahl := offengelegt(t, anwendung.db, partieID); anzahl != 1 {
		t.Fatalf("nach dem Offenlegen stehen %d da", anzahl)
	}

	heilen = zustandVerbiegen(t, anwendung.db, partieID, `AFTER DELETE ON partiefreigaben`)
	offenlegungNachgeladen(t, "zurücknehmen",
		anwendung.PartieOffenlegungZuruecknehmen(ctx, partieID, "anna"))
	heilen()

	if anzahl := offengelegt(t, anwendung.db, partieID); anzahl != 0 {
		t.Fatalf("nach dem Zurücknehmen stehen %d da", anzahl)
	}

	// **Ein gescheiterter Schritt bleibt einer** — ein Zuschauer hat nichts
	// offenzulegen, und das Nein steht in `fehler`.
	if antwort := anwendung.PartieOffenlegen(
		ctx, partieID, "zuschauer", []string{"bert"}); antwort.Fehler == nil ||
		antwort.Nachladefehler != nil {
		t.Fatalf("der Zuschauer: %+v", antwort)
	}
}
