package app

// **Ein Nachladefehler ist kein gescheitertes Annehmen** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ein Auslöser im echten Bestand** trägt das eben angelegte Deck in ein
// anderes Profil — so, wie es ein zweiter Prozess zwischen Schreiben und
// Nachlesen löschen könnte. Angelegt ist dann; das Nachlesen findet es
// nicht. Bis #1970 kam das als `fehler` zurück, **ohne Deck** — und wer es
// noch einmal annahm, legte ein zweites an.

import (
	"testing"
)

func TestEinNachladefehlerIstKeinGescheitertesAnnehmen(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := anwendung.imProfil(t.Context())

	leer := anwendung.DeckAnlegen(ctx, "Leer", "magic", "")
	if leer.Fehler != nil {
		t.Fatalf("Deck: %+v", leer.Fehler)
	}

	ziel := einUmschlagsziel(t, "leer")
	if geschnuert := anwendung.UmschlagSchnueren(ctx, "deck", leer.Deck.ID, ziel); geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	if _, fehler := anwendung.db.ExecContext(ctx, `CREATE TRIGGER verschobenes_deck
AFTER INSERT ON decks
BEGIN UPDATE decks SET profil_id = 999 WHERE id = NEW.id; END`); fehler != nil {
		t.Fatalf("der Auslöser: %v", fehler)
	}

	antwort := anwendung.UmschlagDeck(ctx, ziel)

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("als gescheitert gemeldet: %+v", antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("der Nachladefehler fehlt: %+v", antwort)
	case antwort.Deck == nil || antwort.Deck.ID == "":
		t.Fatalf("das angelegte Deck fehlt: %+v", antwort)
	}

	// **Und angelegt ist es** — gesehen an der Zeile.
	var name string
	if fehler := anwendung.db.QueryRowContext(ctx,
		`SELECT name FROM decks WHERE id = ?`, antwort.Deck.ID).Scan(&name); fehler != nil ||
		name != antwort.Deck.Name {
		t.Fatalf("das Deck %s: %q (%v)", antwort.Deck.ID, name, fehler)
	}

	// **Ein gescheiterter Schritt bleibt einer.**
	if gescheitert := anwendung.UmschlagDeck(ctx, einUmschlagsziel(t, "gibtsnicht")); gescheitert.Fehler == nil ||
		gescheitert.Nachladefehler != nil {
		t.Fatalf("der fehlende Umschlag: %+v", gescheitert)
	}
}
