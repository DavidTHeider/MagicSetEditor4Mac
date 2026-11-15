package app

// **Eine eigene Karte des Absenders kommt nicht mit** — FMT-04.AK-04
// (#1693).
//
// Ein Deckplatz auf eine **eigene** Karte trägt die Kennung aus dem Bestand
// des Absenders. Beim Empfänger zeigt dieselbe Zahl auf eine andere Karte —
// oder auf keine. Bis FMT-04 wurde sie trotzdem gelegt, und das Deck trug
// still eine fremde Karte.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineEigeneKarteDesAbsendersZeigtNichtAufEineHiesige(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	set := anwendung.SetAnlegen(ctx, "Hiesiges Set", "probe", "probe-schlicht")
	if set.Fehler != nil {
		t.Fatalf("Set: %+v", set.Fehler)
	}

	hiesige := anwendung.KarteAnlegen(ctx, set.ID)
	if hiesige.Fehler != nil {
		t.Fatalf("Karte: %+v", hiesige.Fehler)
	}

	// **Dieselbe Zahl, eine andere Karte** — der Absender meinte seine.
	antwort := anwendung.deckAnnehmen(ctx, DeckDTO{
		Name: "Fremddeck", Spiel: "probe",
		Plaetze: []DeckplatzDTO{{
			KarteID: hiesige.KarteID, Bezeichnung: "Glutwind", Menge: 2, Bereich: "hauptdeck",
		}},
	})

	if antwort.Deck == nil {
		t.Fatalf("es kam kein Deck: %+v", antwort.Fehler)
	}

	for _, platz := range antwort.Deck.Plaetze {
		if platz.KarteID != "" {
			t.Errorf("eine fremde eigene Karte liegt im Deck: %+v", platz)
		}
	}

	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "Glutwind") {
		t.Errorf("die Meldung nennt die Karte nicht: %+v", antwort.Fehler)
	}
}

// **Was privat ist, geht nicht mit** — FMT-05.AK-04
// (#1694):
// die Notiz eines Decks und seine Herkunft (die Profilnamen und Quellen
// nennt) stehen nicht im Umschlag. Geprüft am **Inhalt**, nicht am
// angenommenen Deck — das trüge die Herkunft ohnehin nicht.
func TestEinUmschlagTraegtWederNotizNochHerkunft(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	set := anwendung.SetAnlegen(ctx, "Geheime Quelle", "probe", "probe-schlicht")
	anwendung.KarteAnlegen(ctx, set.ID)
	anwendung.MitSetdeck(service.NeuesSetdeck(storage.NeuerSetLeser(anwendung.db), storage.NeueDeckablage(anwendung.db)))

	deck := anwendung.DeckAusSet(ctx, set.ID, "Mit Notiz", 1)
	if deck.Fehler != nil {
		t.Fatalf("Deck: %+v", deck.Fehler)
	}

	if umbenannt := anwendung.DeckUmbenennen(ctx, deck.Deck.ID, deck.Deck.Fassung, "Mit Notiz", "nur für mich"); umbenannt.Fehler != nil {
		t.Fatalf("Notiz: %+v", umbenannt.Fehler)
	}

	ziel := einUmschlagsziel(t, "privat")
	if geschnuert := anwendung.UmschlagSchnueren(ctx, "deck", deck.Deck.ID, ziel); geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	sendung, fehler := anwendung.umschlagLesen(ziel)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	roh, fehler := lesenAusSendung(sendung, austausch.ArtDeck, Deckdatei)
	if fehler != nil {
		t.Fatalf("deck.json: %v", fehler)
	}

	for _, privat := range []string{"nur für mich", "Geheime Quelle"} {
		if strings.Contains(string(roh), privat) {
			t.Errorf("im Umschlag steht %q: %s", privat, roh)
		}
	}
}
