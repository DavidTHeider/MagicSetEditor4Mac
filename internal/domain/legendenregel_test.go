package domain_test

// **Die Legendenregel** — Werkstück M12 der Regelmaschine
// (#1614).
//
// # Die einzige zustandsbasierte Aktion mit einer Wahl
//
// **Das Issue nennt sie so**, und es nennt auch, was zuerst zu bauen ist:
// „Wer CR 704.5j baut, baut zuerst diesen Weg" — den Weg, jemanden zu
// fragen. Hier steht, dass es ihn gibt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineLegendenpartie gibt eine Partie mit zwei gleichnamigen Legenden unter
// derselben Kontrolle.
func eineLegendenpartie(t *testing.T) domain.Partie {
	t.Helper()

	partie := einePartie()

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{
			{
				Kennung: "held", Name: "Held", Menge: 2,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"name": "Held", "super type": "Legendary Creature",
					"sub type": "Human", "power": "2", "toughness": "2",
				}}},
			},
			{
				Kennung: "baer", Name: "Bär", Menge: 2,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"name": "Bär", "super type": "Creature",
					"power": "2", "toughness": "2",
				}}},
			},
		},
	}
	sitze[1].Deck = sitze[0].Deck
	partie.Sitze = sitze

	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "sitz-1/held/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "held",
		},
		{
			Kennung: "sitz-1/held/1", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "held",
		},
	}

	return partie
}

// **Zwei gleichnamige Legenden verlangen eine Wahl** — CR 704.5j
// (Fassung 2026-08-19).
//
// **Und die Prüfung räumt sie nicht selbst weg**: sie stellt die Frage, und
// bis jemand antwortet, liegen beide da. Genau das unterscheidet CR 704.5j
// von den anderen vierundzwanzig.
func TestZweiLegendenVerlangenEineWahl(t *testing.T) {
	t.Parallel()

	danach := eineLegendenpartie(t).Zustandspruefung()

	if anzahl := len(danach.Legendenwahlen); anzahl != 1 {
		t.Fatalf("es stehen %d Wahlen an: %v",
			anzahl, danach.Legendenwahlen)
	}

	wahl := danach.Legendenwahlen[0]
	if wahl.Sitz != "sitz-1" || wahl.Name != "Held" {
		t.Errorf("gewählt werden soll von %q über %q", wahl.Sitz, wahl.Name)
	}

	if anzahl := len(danach.Objekte); anzahl != 2 {
		t.Errorf("es liegen noch %d Objekte da", anzahl)
	}
}

// **Und die Prüfung stellt sie nur einmal.** Sie läuft, sooft jemand
// Priorität bekäme (CR 704.3); eine zweite gleichlautende Wahl wäre eine
// Frage, die niemand beantworten kann.
func TestDieWahlStehtNurEinmalAn(t *testing.T) {
	t.Parallel()

	danach := eineLegendenpartie(t).Zustandspruefung().Zustandspruefung()

	if anzahl := len(danach.Legendenwahlen); anzahl != 1 {
		t.Errorf("es stehen %d Wahlen an", anzahl)
	}
}

// **Drei Dinge verlangen keine Wahl** — CR 704.5j nennt alle drei
// Bedingungen: legendär, gleichnamig, **derselbe** Beherrscher.
func TestWannDieLegendenregelNichtGreift(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Partie){
		"zwei Beherrscher": func(p *domain.Partie) {
			p.Objekte[1].Beherrscher = "sitz-2"
		},
		"zwei Namen": func(p *domain.Partie) {
			p.Objekte[1].Karte = "baer"
		},
		"nicht auf dem Tisch": func(p *domain.Partie) {
			p.Objekte[1].Zone = domain.ZoneHand
		},
		"verdeckt, und damit ohne Namen": func(p *domain.Partie) {
			p.Objekte[1].Verdeckt = true
		},
	}

	for name, aendern := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := eineLegendenpartie(t)
			aendern(&partie)

			if wahlen := partie.Zustandspruefung().Legendenwahlen; len(
				wahlen) != 0 {
				t.Errorf("es stehen %d Wahlen an: %v", len(wahlen), wahlen)
			}
		})
	}
}

// **Ein nichtlegendäres Paar verlangt keine** — CR 704.5j spricht vom
// Supertyp (CR 205.4a).
func TestZweiBaerenVerlangenKeineWahl(t *testing.T) {
	t.Parallel()

	partie := eineLegendenpartie(t)
	partie.Objekte[0].Karte = "baer"
	partie.Objekte[1].Karte = "baer"

	if wahlen := partie.Zustandspruefung().Legendenwahlen; len(wahlen) != 0 {
		t.Errorf("es stehen %d Wahlen an", len(wahlen))
	}
}

// **Der Beherrscher wählt, und der Rest geht in den Friedhof seines
// Eigners** — CR 704.5j.
func TestDieWahlLegtDenRestInDenFriedhof(t *testing.T) {
	t.Parallel()

	partie := eineLegendenpartie(t).Zustandspruefung()

	danach, fehler := partie.Legende("sitz-1", "sitz-1/held/0")
	if fehler != nil {
		t.Fatalf("legende: %v", fehler)
	}

	if zone := objektMit(t, danach,
		"sitz-1/held/0").Zone; zone != domain.ZoneTisch {
		t.Errorf("die gewählte Legende liegt in %q", zone)
	}

	if zone := objektMit(t, danach,
		"sitz-1/held/1").Zone; zone != domain.ZoneFriedhof {
		t.Errorf("die übrige Legende liegt in %q", zone)
	}

	if anzahl := len(danach.Legendenwahlen); anzahl != 0 {
		t.Errorf("es stehen noch %d Wahlen an", anzahl)
	}
}

// **Und eine Wahl, die sich von selbst erledigt, verschwindet.** Stirbt die
// zweite Legende, bevor jemand antwortet, gibt es nichts mehr zu wählen —
// eine Frage ohne Gegenstand wartete sonst für immer (§10).
func TestEineErledigteWahlVerschwindet(t *testing.T) {
	t.Parallel()

	partie := eineLegendenpartie(t).Zustandspruefung()

	zerstoert, fehler := partie.Zerstoeren("sitz-1/held/1")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if wahlen := zerstoert.Zustandspruefung().Legendenwahlen; len(
		wahlen) != 0 {
		t.Errorf("es stehen noch %d Wahlen an: %v", len(wahlen), wahlen)
	}
}

// **Was keine Legendenwahl ist.**
func TestWasKeineLegendenwahlIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		wer, objekt string
		erwartet    error
	}{
		"ein Objekt, das es nicht gibt": {
			wer: "sitz-1", objekt: "sitz-1/greif/0",
			erwartet: domain.ErrNotFound,
		},
		"wer gar nicht wählen soll": {
			wer: "sitz-2", objekt: "sitz-1/held/0",
			erwartet: domain.ErrValidation,
		},
		"ein Objekt, über das niemand wählt": {
			wer: "sitz-1", objekt: "sitz-1/baer/0",
			erwartet: domain.ErrValidation,
		},
	}

	partie := eineLegendenpartie(t)
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "baer",
	})
	partie = partie.Zustandspruefung()

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := partie.Legende(fall.wer, fall.objekt)
			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("legende: %v", fehler)
			}
		})
	}
}

// **Eine Frage wechselt den Adressaten** — CR 110.2a trifft CR 704.5j.
//
// **Gleich viele Fragen sind nicht dieselben Fragen.** Wechselt die
// Kontrolle über beide Legenden, steht weiter **eine** Wahl an — aber ein
// anderer soll sie treffen. Wer nur zählte, ließe den Falschen antworten.
func TestEineWahlWechseltDenAdressaten(t *testing.T) {
	t.Parallel()

	partie := eineLegendenpartie(t).Zustandspruefung()

	for _, kennung := range []string{"sitz-1/held/0", "sitz-1/held/1"} {
		danach, fehler := partie.Beherrschen(kennung, "sitz-2")
		if fehler != nil {
			t.Fatalf("beherrschen %q: %v", kennung, fehler)
		}

		partie = danach
	}

	partie = partie.Zustandspruefung()

	if anzahl := len(partie.Legendenwahlen); anzahl != 1 {
		t.Fatalf("es stehen %d Wahlen an: %v", anzahl, partie.Legendenwahlen)
	}

	if sitz := partie.Legendenwahlen[0].Sitz; sitz != "sitz-2" {
		t.Errorf("wählen soll %q", sitz)
	}
}

// **Eine einzelne Legende ist keine Frage** — CR 704.5j verlangt zwei.
//
// **Und die Antwort darauf wird abgewiesen, obwohl alles andere stimmt**:
// das Objekt ist legendär, es liegt auf dem Tisch, und der Fragende
// beherrscht es. Was fehlt, ist die Frage.
func TestUeberEineEinzelneLegendeWirdNichtGewaehlt(t *testing.T) {
	t.Parallel()

	partie := eineLegendenpartie(t)
	partie.Sitze[0].Deck.Karten = append(partie.Sitze[0].Deck.Karten,
		domain.Abzugskarte{
			Kennung: "weise", Name: "Weise", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"name": "Weise", "super type": "Legendary Creature",
				"power": "1", "toughness": "1",
			}}},
		})
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/weise/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "weise",
	})

	_, fehler := partie.Zustandspruefung().Legende(
		"sitz-1", "sitz-1/weise/0")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("über die einzelne Legende: %v", fehler)
	}
}
