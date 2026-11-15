package domain_test

// **Die Kartenkennung in der Projektion** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Eine eigene Datei neben `partiesicht_test.go`**, weil jene über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, **was**
// eine Sicht zeigt und verbirgt, hier, **woran** ein Bild hängt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Kartenkennung kommt nur mit, wo hineingesehen werden darf.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Sie ist der Weg zum Bild** (G2,
// #1459) —
// und damit derselbe Weg zur Karte wie ihr Name. Wer sie an einem geheimen
// Objekt mitgäbe, verriete die Karte über einen Umweg, und die Projektion
// daneben sähe weiter richtig aus: kein Name, kein Druck, keine Ordnung — und
// trotzdem die Karte.
func TestDieKartenkennungKommtNurAnEinemOffenenObjektMit(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Sitze[0].Deck.Karten[0].Seiten = []domain.Abzugsseite{
		{Name: "Wald", KarteID: 77},
	}

	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "offen", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "k1",
		},
		{
			Kennung: "verdeckt", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneBibliothek, Karte: "k1",
		},
	}

	sicht := partie.Sicht("sitz-1")

	gefunden := map[string]int64{}
	for kennung, objekt := range sichtobjekte(partie, sicht) {
		gefunden[kennung] = objekt.KarteID
	}

	if gefunden["offen"] != 77 {
		t.Errorf("das offene Objekt gibt %d, erwartet 77", gefunden["offen"])
	}

	if gefunden["verdeckt"] != 0 {
		t.Errorf("das verdeckte Objekt verrät die Karte %d",
			gefunden["verdeckt"])
	}
}

// **Ein Objekt ohne Sitz und eines ohne eigene Karte geben null.**
//
// **Beides gibt es wirklich**: ein Spielstein gehört zu keinem Abzug, und ein
// Platz kann statt einer eigenen Karte einen echten Druck nennen — dann gibt
// es kein Bild dieses Programms.
func TestOhneSitzUndOhneEigeneKarteGibtEsKeineKennung(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Sitze[0].Deck.Karten[0].Seiten = []domain.Abzugsseite{
		{Name: "Fremd", KarteID: 0},
	}

	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "ohneSitz", Art: domain.ObjektSpielstein,
			Beherrscher: "niemand", Eigner: "niemand",
			Zone: domain.ZoneTisch, Karte: "k1",
		},
		{
			Kennung: "echterDruck", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "k1",
		},
		{
			Kennung: "unbekannt", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "gibtesnicht",
		},
	}

	for _, objekt := range partie.Sicht("sitz-1").Objekte {
		if objekt.KarteID != 0 {
			t.Errorf("%q gibt %d, erwartet null", objekt.Kennung, objekt.KarteID)
		}
	}
}
