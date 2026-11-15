package app

import (
	"path/filepath"
	"testing"
)

// **Und ein Widerruf gibt danach den Stand.**
func TestEinWiderrufGibtDanachDenStand(t *testing.T) {
	t.Parallel()

	anwendung := eineLeihanwendung(t)
	ctx := riegelkontext(t)

	deckID := einDeckImBestand(t, anwendung.db)

	angeboten := anwendung.DeckleiheAnbieten(ctx, LeihwunschDTO{
		Entleiher: "12D3KooWBert", DeckID: deckID,
		Bindungsart: "partie", Woran: "p1",
		Rechte: []string{"spielen"}, Stunden: 24, Modus: "zuende",
	}, filepath.Join(t.TempDir(), "l.msepackage"))

	if angeboten.Fehler != nil {
		t.Fatalf("anbieten: %+v", angeboten.Fehler)
	}

	widerrufen := anwendung.DeckleiheWiderrufen(
		ctx, angeboten.Leihen[0].Kennung)

	switch {
	case widerrufen.Fehler != nil:
		t.Fatalf("widerrufen: %+v", widerrufen.Fehler)
	case len(widerrufen.Leihen) != 1:
		t.Fatalf("der Stand ist %+v", widerrufen.Leihen)
	case !widerrufen.Leihen[0].Widerrufen:
		t.Error("die Leihe steht nicht als widerrufen")
	}

	// **Und eine tote Datenbank meldet sich beim Lesen.**
	if fehler := anwendung.db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if antwort := anwendung.Deckleihen(ctx); antwort.Fehler == nil {
		t.Fatal("die geschlossene Datenbank kam nicht an")
	}
}
