package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// anlageprobe ist die Naht zum Anlegen.
type anlageprobe struct {
	setID    int64
	spiel    string
	name     string
	ergebnis service.Anlageergebnis
	fehler   error
}

func (a *anlageprobe) Anlegen(_ context.Context, setID int64, spiel, name string) (
	service.Anlageergebnis, error,
) {
	a.setID, a.spiel, a.name = setID, spiel, name

	return a.ergebnis, a.fehler
}

// setprobe gibt genau ein Set — das Spiel steht daran, nicht am Aufruf.
type setprobe struct{}

func (setprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 7, Game: "magic"}}, nil
}

func (setprobe) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, nil
}

func (setprobe) Rohbloecke(context.Context, int64) ([]domain.RawBlock, error) {
	return nil, nil
}

func einAnlageergebnis() service.Anlageergebnis {
	return service.Anlageergebnis{
		KarteID: 42,
		Uebernommen: []service.Feldvorschlag{
			{Feld: "name", Neu: "Lightning Bolt"},
			{Feld: "rule text", Neu: "…"},
		},
		Befunde:  []domain.Befund{{Ort: "rarity", Meldung: "unbekannte Seltenheit"}},
		Herkunft: "https://scryfall.com/card/lea/161",
		Alter:    2 * time.Hour,
	}
}

// **Eine Karte aus einem Namen, über die Fassade** —
// #1032.
func TestScryfallKarteAnlegen(t *testing.T) {
	naht := &anlageprobe{ergebnis: einAnlageergebnis()}

	anwendung := app.NeueApp("")
	anwendung.MitBestand(setprobe{})
	anwendung.MitScryfallanlage(naht)

	antwort := anwendung.ScryfallKarteAnlegen(
		probekontext(t), "7", "Lightning Bolt", "", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if naht.setID != 7 || naht.spiel != "magic" || naht.name != "Lightning Bolt" {
		t.Errorf("die Naht bekam %d/%q/%q", naht.setID, naht.spiel, naht.name)
	}

	if antwort.KarteID != "42" {
		t.Errorf("KarteID = %q", antwort.KarteID)
	}

	if len(antwort.Felder) != 2 || antwort.Felder[0] != "name" {
		t.Errorf("Felder = %v, erwartet die geschriebenen", antwort.Felder)
	}

	// **Die Befunde reisen mit** (§8): eine Karte mit einer unbekannten
	// Seltenheit ist brauchbar — die Auffälligkeit gehört daneben.
	if len(antwort.Befunde) != 1 {
		t.Errorf("Befunde = %v, erwartet einen", antwort.Befunde)
	}

	if antwort.AlterStunden != 2 {
		t.Errorf("AlterStunden = %v, erwartet 2 — §16, Regel 5", antwort.AlterStunden)
	}
}

// **Die angelegte Karte wird auch im Fehlerfall genannt** (#155).
func TestScryfallKarteAnlegenNenntDieKarteAuchImFehlerfall(t *testing.T) {
	naht := &anlageprobe{
		ergebnis: service.Anlageergebnis{KarteID: 42},
		fehler:   errors.New("der Bestand ist nicht erreichbar"),
	}

	anwendung := app.NeueApp("")
	anwendung.MitBestand(setprobe{})
	anwendung.MitScryfallanlage(naht)

	antwort := anwendung.ScryfallKarteAnlegen(
		probekontext(t), "7", "Lightning Bolt", "", "")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.KarteID != "42" {
		t.Errorf("KarteID = %q — eine angelegte Karte, die niemand nennt, bleibt "+
			"im Bestand, ohne dass jemand davon weiß", antwort.KarteID)
	}
}

// **Eine Bildart ohne Feld wird abgewiesen.**
//
// Welches Feld das Bild aufnimmt, weiß nur die Vorlage; es zu raten hieße, bei
// jeder fremden Vorlage falsch zu raten.
func TestEineBildartOhneFeldWirdAbgewiesen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitBestand(setprobe{})
	anwendung.MitScryfallanlage(&anlageprobe{ergebnis: einAnlageergebnis()})

	antwort := anwendung.ScryfallKarteAnlegen(
		probekontext(t), "7", "Lightning Bolt", "", "ausschnitt")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler, obwohl das Bildfeld fehlt")
	}
}

// **Ohne eingerichteten Zugang gibt es eine Meldung, keinen Absturz.**
func TestScryfallKarteAnlegenOhneZugang(t *testing.T) {
	antwort := app.NeueApp("").ScryfallKarteAnlegen(
		probekontext(t), "7", "Lightning Bolt", "", "")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler, obwohl der Zugang fehlt")
	}
}

// **Eine unbrauchbare Setkennung erreicht die Naht nicht.**
func TestScryfallKarteAnlegenMitKrummerKennung(t *testing.T) {
	naht := &anlageprobe{ergebnis: einAnlageergebnis()}

	anwendung := app.NeueApp("")
	anwendung.MitBestand(setprobe{})
	anwendung.MitScryfallanlage(naht)

	antwort := anwendung.ScryfallKarteAnlegen(
		probekontext(t), "keine Zahl", "Lightning Bolt", "", "")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if naht.name != "" {
		t.Errorf("die Naht wurde gerufen, obwohl die Kennung krumm ist")
	}
}

// **Das Bild kommt in demselben Aufruf mit** —
// #1032.
//
// **Über denselben Weg wie sonst** (`ScryfallBild` und `ScryfallBildEinsetzen`):
// eine zweite Fassung liefe bei der ersten Änderung an der Nennungspflicht
// auseinander.
func TestScryfallKarteAnlegenNimmtDasBildMit(t *testing.T) {
	holer := &bildholer{holung: einBildfund(t, "Christopher Rush")}
	anwendung, schreiber, _ := scryfallbildaufbau(t, holer)
	anwendung.MitBestand(setprobe{})
	anwendung.MitScryfallanlage(&anlageprobe{
		ergebnis: service.Anlageergebnis{KarteID: 1},
	})

	antwort := anwendung.ScryfallKarteAnlegen(
		probekontext(t), "7", "Lightning Bolt", "kunst", "normal")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if schreiber.felder["kunst"] == "" {
		t.Errorf("das Bildfeld ist leer: %v", schreiber.felder)
	}

	if antwort.Nennung == "" {
		t.Error("die Nennung fehlt — sie gehört zum Bild, nicht zur Oberfläche")
	}

	if len(antwort.Felder) == 0 || antwort.Felder[len(antwort.Felder)-1] != "kunst" {
		t.Errorf("Felder = %v, erwartet das Bildfeld am Ende", antwort.Felder)
	}
}

// **Scheitert das Bild, steht die Karte trotzdem da.**
//
// §7 Regel 3: geschrieben ist geschrieben. Die Kennung wird genannt, sonst
// bleibt die Karte im Bestand, ohne dass jemand davon weiß (#155).
func TestEinGescheitertesBildLaesstDieKarteStehen(t *testing.T) {
	holer := &bildholer{fehler: errors.New("Scryfall antwortet nicht")}
	anwendung, _, _ := scryfallbildaufbau(t, holer)
	anwendung.MitBestand(setprobe{})
	anwendung.MitScryfallanlage(&anlageprobe{
		ergebnis: service.Anlageergebnis{KarteID: 1},
	})

	antwort := anwendung.ScryfallKarteAnlegen(
		probekontext(t), "7", "Lightning Bolt", "kunst", "normal")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler, obwohl das Bild scheitert")
	}

	if antwort.KarteID != "1" {
		t.Errorf("KarteID = %q — die angelegte Karte muss genannt werden", antwort.KarteID)
	}
}

// **Und scheitert das Ablegen, ebenso.**
func TestEinGescheitertesAblegenLaesstDieKarteStehen(t *testing.T) {
	holer := &bildholer{holung: einBildfund(t, "Christopher Rush")}
	anwendung, schreiber, _ := scryfallbildaufbau(t, holer)
	schreiber.fehler = errors.New("der Bestand ist nicht erreichbar")
	anwendung.MitBestand(setprobe{})
	anwendung.MitScryfallanlage(&anlageprobe{
		ergebnis: service.Anlageergebnis{KarteID: 1},
	})

	antwort := anwendung.ScryfallKarteAnlegen(
		probekontext(t), "7", "Lightning Bolt", "kunst", "normal")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler, obwohl das Ablegen scheitert")
	}

	if antwort.KarteID != "1" {
		t.Errorf("KarteID = %q", antwort.KarteID)
	}
}

// **Das Spiel steht am Set, und wenn es das Set nicht gibt, sagt es das.**
//
// Drei Wege, und der letzte ist der, den man vergisst: ein Set, das es nicht
// gibt, ergäbe sonst ein leeres Spiel — und die Meldung nennte dann die
// fehlende Zuordnung statt das fehlende Set.
func TestDasSpielKommtVomSet(t *testing.T) {
	for name, fall := range map[string]struct {
		bestand app.Bestandsleser
		setID   string
	}{
		"ohne Bestand":               {nil, "7"},
		"ein Bestand ohne Antwort":   {stolperndeSets{}, "7"},
		"ein Set, das es nicht gibt": {setprobe{}, "8"},
	} {
		t.Run(name, func(t *testing.T) {
			anwendung := app.NeueApp("")
			anwendung.MitScryfallanlage(&anlageprobe{ergebnis: einAnlageergebnis()})

			if fall.bestand != nil {
				anwendung.MitBestand(fall.bestand)
			}

			antwort := anwendung.ScryfallKarteAnlegen(
				probekontext(t), fall.setID, "Lightning Bolt", "", "")
			if antwort.Fehler == nil {
				t.Fatal("kein Fehler")
			}
		})
	}
}

// stolperndeSets antwortet nicht.
type stolperndeSets struct{}

func (stolperndeSets) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, errors.New("der Bestand ist nicht erreichbar")
}

func (stolperndeSets) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, nil
}

func (stolperndeSets) Rohbloecke(context.Context, int64) ([]domain.RawBlock, error) {
	return nil, nil
}
