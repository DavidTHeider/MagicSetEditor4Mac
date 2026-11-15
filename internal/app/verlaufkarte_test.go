package app_test

// **Ein Kartenschritt sagt, ob die Karte kam oder ging** — G01
// (#2248).
//
// Gefunden am nativen Fenster (`g01einsteiger`): die erste Karte stand im
// Verlauf als »Karte angelegt oder gelöscht«, mit »kein Wert« vorher und
// nachher. Der Bestand weiß es — `Kartenstand.GabEs` auf beiden Seiten —,
// die Fassade gab es nur nicht weiter.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinKartenschrittSagtObDieKarteKamOderGing(t *testing.T) {
	probe := &verlaufsprobe{eintraege: []domain.Verlaufseintrag{
		{
			ID: 1, KarteID: 7, Zeitpunkt: "2026-10-10T22:50:49Z", Handlung: domain.HandlungKarte,
			NachherKarte: &domain.Kartenstand{GabEs: true}, VorherKarte: &domain.Kartenstand{}, Urheber: domain.UrheberFenster,
		},
		{
			ID: 2, KarteID: 7, Zeitpunkt: "2026-10-10T22:51:00Z", Handlung: domain.HandlungKarte,
			VorherKarte: &domain.Kartenstand{GabEs: true}, NachherKarte: &domain.Kartenstand{}, Urheber: domain.UrheberFenster,
		},
	}}

	antwort := mitVerlauf(probe).Verlauf(probekontext(t), "3")
	if antwort.Fehler != nil || len(antwort.Schritte) != 2 {
		t.Fatalf("Antwort %+v", antwort)
	}

	geloescht, angelegt := antwort.Schritte[0], antwort.Schritte[1]
	if angelegt.GabVorher || !angelegt.GabNachher {
		t.Errorf("angelegt: vorher %v, nachher %v", angelegt.GabVorher, angelegt.GabNachher)
	}

	if !geloescht.GabVorher || geloescht.GabNachher {
		t.Errorf("gelöscht: vorher %v, nachher %v", geloescht.GabVorher, geloescht.GabNachher)
	}
}
