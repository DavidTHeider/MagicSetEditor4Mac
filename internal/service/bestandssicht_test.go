package service_test

// **Für welchen Bestand eine Sicht gerechnet wird** — P9
// (#2377),
// Vermerk 0198.
//
// Der Wirt rechnet jedem Gast seine Sicht. Mit den Zeilen des **eigenen**
// Bestands darin schlüge der Gast sie in **seinem** nach — und fände dort
// eine andere Karte. Deshalb reist im Kontext, für wessen Bestand gerechnet
// wird: am Fenster des Wirts für keinen (seinen eigenen), am Gastplatz für
// den Peer, der dort sitzt.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bestandsnehmer merkt sich, für welchen Bestand ein Tischbild gerechnet wurde.
type bestandsnehmer struct {
	nehmerdoppel
	bestaende []string
}

func (n *bestandsnehmer) Tischbild(
	ctx context.Context, kennung, fuer string,
) (domain.Tischbild, error) {
	n.bestaende = append(n.bestaende, fuer+"@"+service.BestandssichtAus(ctx))

	return n.nehmerdoppel.Tischbild(ctx, kennung, fuer)
}

func TestDieLeitungRechnetFuerDenBestandImKontext(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)
	partie := einePartie()
	// **Je Sitz eine eigene Kartenliste** — `einePartie` teilt eine.
	karte := func(zeile int64) []domain.Abzugskarte {
		return []domain.Abzugskarte{{
			Kennung: "k1", Name: "Wald", Menge: 40,
			Seiten: []domain.Abzugsseite{{KarteID: zeile}},
		}}
	}
	partie.Sitze[0].Deck.Karten = karte(3)
	partie.Sitze[1].Deck.Karten = karte(7)
	partie.Sitze[1].Deck.Quelle = "peer-bo"
	partie.Objekte = []domain.Spielobjekt{
		{Kennung: "o1", Art: domain.ObjektKarte, Eigner: "sitz-1", Beherrscher: "sitz-1", Zone: domain.ZoneTisch, Karte: "k1"},
		{Kennung: "o2", Art: domain.ObjektKarte, Eigner: "sitz-2", Beherrscher: "sitz-2", Zone: domain.ZoneTisch, Karte: "k1"},
	}

	ctx := context.Background()
	if _, fehler := leitung.Anlegen(ctx, partie); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	kennungen := func(sicht domain.Partiesicht) (int64, int64) {
		return sicht.Objekte[0].KarteID, sicht.Objekte[1].KarteID
	}

	amWirt, fehler := leitung.Partie(ctx, "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Partie: %v", fehler)
	}

	if eins, zwei := kennungen(amWirt); eins != 3 || zwei != 0 {
		t.Errorf("am Wirt: %d, %d", eins, zwei)
	}

	amGast, fehler := leitung.Partie(service.MitBestandssicht(ctx, "peer-bo"), "p1", "sitz-2")
	if fehler != nil {
		t.Fatalf("Partie: %v", fehler)
	}

	if eins, zwei := kennungen(amGast); eins != 0 || zwei != 7 {
		t.Errorf("am Gast: %d, %d", eins, zwei)
	}
}

// **Der Funk rechnet jedem Gast mit seinem Bestand, dem Fenster mit keinem.**
func TestDerFunkRechnetJedemGastSeinenBestand(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}
	anna, _ := einPlatz("anna")
	bert, _ := einPlatz("bert")
	runde.Setze(anna.MitPeer("peer-anna"))
	runde.Setze(bert.MitPeer("peer-bert"))

	nehmer := &bestandsnehmer{}
	fenster := &mitschrift{}
	funk := service.NeuerTischfunk("p1", nehmer, runde).MitFenster("wirt", fenster)

	if fehler := funk.Verkuende(context.Background(), "",
		service.Befehlsausgang{Fassung: 2}); fehler != nil {
		t.Fatalf("Verkuende: %v", fehler)
	}

	gefragt := map[string]bool{}
	for _, eintrag := range nehmer.bestaende {
		gefragt[eintrag] = true
	}

	for _, erwartet := range []string{"anna@peer-anna", "bert@peer-bert", "wirt@"} {
		if !gefragt[erwartet] {
			t.Errorf("%q fehlt in %v", erwartet, nehmer.bestaende)
		}
	}

	// **Ein Sitz ohne Gast hat keinen Bestand außer dem eigenen.**
	if peer := runde.Peer("carla"); peer != "" {
		t.Errorf("ein leerer Sitz nennt %q", peer)
	}
}

// **Der Wirt eines Gastes rechnet dessen Tischbild mit dessen Bestand.**
func TestDerWirtRechnetDasTischbildMitDemBestandDesGastes(t *testing.T) {
	t.Parallel()

	nehmer := &bestandsnehmer{}
	wirt, _ := einWirt(nehmer)
	wirt = wirt.MitGegenueber("peer-gast")

	befehl := domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "gast", Art: "objekt.tappen",
	}
	if fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t, befehl)); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if len(nehmer.bestaende) == 0 || nehmer.bestaende[0] != "gast@peer-gast" {
		t.Errorf("gerechnet für %v", nehmer.bestaende)
	}
}
