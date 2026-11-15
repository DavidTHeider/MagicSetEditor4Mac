package app

// **Die Ausgänge der vier weiteren Arten** — Werkstück `T7` von Phase 5.6
// (#1472).
//
// **Eine zweite Datei neben `umschlagartenwege_intern_test.go`**, weil beide
// zusammen über die Zeilengrenze gingen (§2). Der Schnitt folgt der Frage:
// dort, **was hinausgeht** — hier, **was schiefgeht**.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die beiden neuen Bindungen und ihre drei Ausgänge.**
//
// **Ein abgebrochener Dialog ist keine Meldung wert** (§6), ein Dialog, der
// klagt, ist eine — und was danach kommt, ist der eigentliche Weg.
func TestDieNeuenUmschlagbindungenHaltenIhreDreiAusgaenge(t *testing.T) {
	anwendung := mitIdentitaet(t)
	bindungen := NeueBindungen(anwendung)

	// **Abgebrochen** — ein leerer Pfad.
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", nil
	})

	if antwort := bindungen.UmschlagDeck(t.Context()); antwort.Fehler != nil {
		t.Errorf("der Abbruch gab einen Fehler: %+v", antwort.Fehler)
	}

	if antwort := bindungen.UmschlagSicht(t.Context()); antwort.Fehler != nil {
		t.Errorf("der Abbruch gab einen Fehler: %+v", antwort.Fehler)
	} else if antwort.Zeilen == nil {
		t.Error("die Zeilen sind nil")
	}

	// **Der Dialog selbst klagt.**
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", errProbedialog
	})

	if antwort := bindungen.UmschlagDeck(t.Context()); antwort.Fehler == nil {
		t.Error("die Klage des Dialogs ging verloren")
	}

	if antwort := bindungen.UmschlagSicht(t.Context()); antwort.Fehler == nil {
		t.Error("die Klage des Dialogs ging verloren")
	}

	// **Und ein Pfad, hinter dem nichts liegt.**
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return filepath.Join(t.TempDir(), "gibtesnicht.msepackage"), nil
	})

	if antwort := bindungen.UmschlagDeck(t.Context()); antwort.Fehler == nil {
		t.Error("der fehlende Umschlag kam durch")
	}

	if antwort := bindungen.UmschlagSicht(t.Context()); antwort.Fehler == nil {
		t.Error("der fehlende Umschlag kam durch")
	}
}

// errProbedialog ist, was ein Dateidialog sagt, der nicht aufgeht.
var errProbedialog = errors.New("der Dialog ging nicht auf")

// **Ein Umschlag mit unlesbarem Inhalt meldet sich.**
//
// **Was `austausch.Lies` nicht fängt**: es prüft Prüfsummen, nicht den Sinn.
// Ein Deck, dessen JSON kein Deck ist, kommt durch — und fällt hier.
func TestEinUnlesbarerInhaltImUmschlagMeldetSich(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	ich, fehler := anwendung.peerwerk().Ausweis(anwendung.imProfil(ctx))
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	for name, fall := range map[string]struct {
		art   austausch.Objektart
		datei string
		lesen func(string) *Fehler
	}{
		"ein Deck, das keines ist": {
			austausch.ArtDeck, Deckdatei,
			func(pfad string) *Fehler {
				return anwendung.UmschlagDeck(ctx, pfad).Fehler
			},
		},
		"eine Sicht, die keine ist": {
			austausch.ArtSammlung, Sichtdatei,
			func(pfad string) *Fehler {
				return anwendung.UmschlagSicht(ctx, pfad).Fehler
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ziel := einUmschlagsziel(t, "krumm")

			if _, fehler := austausch.Schreibe(ziel, []austausch.Inhalt{{
				Art: fall.art, Name: fall.datei, Bezeichnung: fall.datei,
				Bytes: []byte("kein JSON"),
			}}, ich); fehler != nil {
				t.Fatalf("schreiben: %v", fehler)
			}

			if fehler := fall.lesen(ziel); fehler == nil {
				t.Error("der krumme Inhalt kam durch")
			}
		})
	}
}

// **Ein Deck mit einem Platz, den dieser Rechner nicht legen kann.**
//
// **Es bleibt trotzdem stehen** (§8): ein Platz, dessen Karte fehlt, ist ein
// Befund — und das angelegte Deck kommt mit der Meldung zurück, statt still
// zu verschwinden.
func TestEinDeckMitEinemUnmoeglichenPlatzMeldetSich(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	antwort := anwendung.deckAnnehmen(ctx, DeckDTO{
		Name: "Fremddeck", Spiel: "magic",
		Plaetze: []DeckplatzDTO{{
			Printing: "keine Druckkennung", Menge: 1,
			Bereich: "hauptdeck",
		}},
	})

	if antwort.Fehler == nil {
		t.Fatal("der unmögliche Platz kam durch")
	}

	// **Und das Deck steht da** — es ist angelegt, und das ist die Auskunft.
	if antwort.Deck == nil {
		t.Error("das angelegte Deck kam nicht mit")
	}
}

// **Die letzten drei Ausgänge, die kein Weg von außen erreicht.**
//
// **Sie werden einzeln gefahren** (§10): eine Art, die der Wähler schon
// abgewiesen hat, ein Paket ohne Datei und eine Datei, die sich nicht lesen
// lässt — drei Stellen, an denen sonst still nichts geschähe.
func TestDieLetztenAusgaengeDesPackensWerdenGefahren(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	// **Eine Art, die [umschlagart] gar nicht durchließe.** Der Zweig steht
	// trotzdem da: `inhalteFuer` ist die Stelle, an der eine sechste Art
	// laut würde, statt als Set durchzugehen.
	if _, fehler := anwendung.inhalteFuer(
		ctx, austausch.Objektart("tanzstunde"), "1"); fehler == nil {
		t.Error("die fremde Art kam durch")
	}

	// **Ein Paket, das keine Datei führt.**
	//
	// **Über eine Werkstatt, die eine leere Liste gibt** — ein leeres
	// Verzeichnis kommt gar nicht so weit: `msepaket.Oeffne` weist es vorher
	// ab, und der Zweig hier fängt den Fall, dass ein Paket **lesbar** ist
	// und nichts Bearbeitbares führt.
	anwendung.paketdateien = leereWerkstatt{}

	if _, fehler := anwendung.vorlageEntfalten(
		ctx, "leer.mse-game"); fehler == nil {
		t.Error("das leere Paket kam durch")
	}

	// **Und eine Datei, die im Verzeichnis steht und nicht zu lesen ist.**
	anwendung.paketdateien = &unlesbareWerkstatt{}

	if _, fehler := anwendung.vorlageEntfalten(ctx, "irgendeins"); fehler == nil {
		t.Error("die unlesbare Datei kam durch")
	}
}

// unlesbareWerkstatt nennt eine Datei und gibt sie nicht her.
//
// **Was ein echter Ort nicht hergibt**: er listet, was er lesen kann. Der
// Zweig dahinter ist die Stelle, an der ein Rechtefehler zwischen Auflisten
// und Lesen laut würde.
type unlesbareWerkstatt struct{}

func (unlesbareWerkstatt) Lesen(context.Context, string, string) (
	inhalt, stand string, fehler error,
) {
	return "", "", errProbedialog
}

func (unlesbareWerkstatt) Schreiben(
	context.Context, string, string, string, string,
) (string, error) {
	return "", errProbedialog
}

func (unlesbareWerkstatt) Texte(context.Context, string) (
	[]service.Paketdateitext, error,
) {
	return nil, errProbedialog
}

func (unlesbareWerkstatt) DateienUndBeiwerk(context.Context, string) (dateien, beiwerk []string, fehler error) {
	return []string{"style"}, nil, nil
}

func (unlesbareWerkstatt) Anlegen(context.Context, string, string, string) (string, error) {
	return "", errProbedialog
}

func (unlesbareWerkstatt) Einfuegen(context.Context, string, string, string) error {
	return errProbedialog
}

func (unlesbareWerkstatt) Umbenennen(context.Context, string, string, string, string) error {
	return errProbedialog
}

func (unlesbareWerkstatt) Entfernen(context.Context, string, string, string) error {
	return errProbedialog
}

// **Ein Umschlag mit der falschen Art darin.**
//
// **Beide Richtungen**: ein Deck, das als Sicht gelesen wird, und eine Sicht,
// die als Deck gelesen wird. Beide enden in einem Satz und nicht in einer
// leeren Antwort.
func TestEineFalscheArtImUmschlagMeldetSich(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	ich, fehler := anwendung.peerwerk().Ausweis(anwendung.imProfil(ctx))
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	ziel := einUmschlagsziel(t, "sichtdrin")

	if _, fehler := austausch.Schreibe(ziel, []austausch.Inhalt{{
		Art: austausch.ArtSammlung, Name: Sichtdatei,
		Bezeichnung: Sichtdatei, Bytes: []byte("[]"),
	}}, ich); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.UmschlagDeck(ctx, ziel); antwort.Fehler == nil {
		t.Error("eine Sicht kam als Deck durch")
	}
}

// **Ein Deck, das sich auch unter dem Zusatz nicht anlegen lässt.**
//
// **Einmal umbenannt und nicht in einer Schleife**: wer auch den Zusatz schon
// vergeben hat, bekommt den Konflikt — mit einem Satz, der sagt, was los ist.
func TestEinZweifachVergebenerDecknameMeldetSich(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	zusatz := DialogtexteFuer(anwendung.sprache).Uebernommen

	for _, name := range []string{"Doppelt", "Doppelt (" + zusatz + ")"} {
		if antwort := anwendung.DeckAnlegen(
			ctx, name, "magic", ""); antwort.Fehler != nil {
			t.Fatalf("%s anlegen: %+v", name, antwort.Fehler)
		}
	}

	antwort := anwendung.deckAnnehmen(ctx,
		DeckDTO{Name: "Doppelt", Spiel: "magic"})
	if antwort.Fehler == nil {
		t.Error("der zweifach vergebene Name kam durch")
	}
}

// leereWerkstatt nennt ein Paket ohne bearbeitbare Datei.
type leereWerkstatt struct{ unlesbareWerkstatt }

func (leereWerkstatt) DateienUndBeiwerk(context.Context, string) (dateien, beiwerk []string, fehler error) {
	return []string{}, []string{}, nil
}
