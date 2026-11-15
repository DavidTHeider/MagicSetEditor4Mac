package app_test

// **Ein Nachladefehler ist keine gescheiterte Vorgabe** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ohne Auslöser, an einem echten Bestand**: das Setzen fragt nur das
// Angebot **seiner** Art, das Nachlesen danach das aller fünf. Ein Format,
// dessen offenes Sideboard weder ja noch nein ist, lässt sich nicht lesen —
// eine Deckvorgabe glückt, und der Stand danach scheitert an den Formaten.

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestEinNachladefehlerIstKeineGescheiterteVorgabe(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)
	ctx := probekontext(t)

	ergebnis, fehler := db.ExecContext(ctx,
		`INSERT INTO decks (profil_id, name, spiel) VALUES (1, 'Mono Rot', 'magic')`)
	if fehler != nil {
		t.Fatalf("das Deck: %v", fehler)
	}

	deckID, _ := ergebnis.LastInsertId() //nolint:errcheck // SQLite gibt sie immer
	deck := strconv.FormatInt(deckID, 10)

	if _, fehler := db.ExecContext(ctx, `INSERT INTO formate
		(profil_id, name, spiel, sideboard_offen) VALUES (1, 'Krumm', 'magic', 2)`); fehler != nil {
		t.Fatalf("das krumme Format: %v", fehler)
	}

	flaeche := app.NutzerflaecheUeber(db)

	antwort := flaeche.ProfilvorgabeSetzen(ctx,
		app.VorgabeangabenDTO{ProfilID: "1", Art: "deck", Werte: []string{deck}})

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("als gescheitert gemeldet: %+v", antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("der Nachladefehler fehlt: %+v", antwort)
	case antwort.Vorgaben == nil:
		t.Fatal("die Liste ist nil")
	}

	var gesetzt string
	if fehler := db.QueryRowContext(ctx,
		`SELECT wert FROM profilvorgaben WHERE profil_id = 1 AND art = 'deck'`,
	).Scan(&gesetzt); fehler != nil || gesetzt != deck {
		t.Fatalf("gesetzt ist %q (%v)", gesetzt, fehler)
	}

	// **Ein gescheiterter Schritt bleibt einer** — ein Deck, das es nicht
	// gibt, wird abgewiesen, und das Nein steht in `fehler`.
	if abgewiesen := flaeche.ProfilvorgabeSetzen(ctx, app.VorgabeangabenDTO{
		ProfilID: "1", Art: "deck", Werte: []string{"999"},
	}); abgewiesen.Fehler == nil || abgewiesen.Nachladefehler != nil {
		t.Fatalf("das fremde Deck: %+v", abgewiesen)
	}
}
