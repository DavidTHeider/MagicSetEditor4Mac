package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Karte aus einem Namen** —
// #1032.
//
// Bis hierher arbeitete der Abgleich nur an einer **vorhandenen** Karte: erst
// eine leere anlegen, dann abgleichen — zwei Schritte für einen Vorgang, und
// den ersten muss man kennen.

// anlegeprobe zählt die angelegten Karten.
type anlegeprobe struct {
	angelegt int
	fehler   error
}

func (a *anlegeprobe) KarteAnlegen(context.Context, int64) (int64, error) {
	if a.fehler != nil {
		return 0, a.fehler
	}

	a.angelegt++

	return int64(100 + a.angelegt), nil
}

// fremdprobe gibt eine feste Karte — oder scheitert.
type fremdprobe struct {
	karte  *scryfall.Karte
	fehler error
	rufe   int
}

func (f *fremdprobe) KarteNachName(context.Context, string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	f.rufe++

	if f.fehler != nil {
		return nil, 0, nil, f.fehler
	}

	return f.karte, 0, nil, nil
}

// schreibprobe merkt sich, was geschrieben wurde.
type anlageschreiber struct {
	werte      map[string]string
	herkuenfte map[string]string
	fehler     error
}

// FelderVonAussen schreibt alle Werte oder keinen (#1026).
func (s *anlageschreiber) FelderVonAussen(
	ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	for _, wert := range werte {
		if _, _, fehler := s.FeldSetzenVonAussen(
			ctx, setID, karteID, wert.Field, wert.Text, herkunft); fehler != nil {
			return "", false, fehler
		}
	}

	return "", false, nil
}

func (s *anlageschreiber) FeldSetzenVonAussen(
	_ context.Context, _, _ int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	if s.fehler != nil {
		return "", false, s.fehler
	}

	if s.werte == nil {
		s.werte, s.herkuenfte = map[string]string{}, map[string]string{}
	}

	s.werte[feld] = wert
	s.herkuenfte[feld] = herkunft

	return "", false, nil
}

func eineFremdkarte() *scryfall.Karte {
	return &scryfall.Karte{
		Name:       "Lightning Bolt",
		Manakosten: "{R}",
		Typzeile:   "Instant",
		Regeltext:  "Lightning Bolt deals 3 damage to any target.",
		Seltenheit: "common",
		Kuenstler:  "Christopher Rush",
		Adresse:    "https://scryfall.com/card/lea/161/lightning-bolt",
	}
}

func TestEineKarteEntstehtAusEinemScryfallnamen(t *testing.T) {
	anleger := &anlegeprobe{}
	schreiber := &anlageschreiber{}

	anlage := service.NeueScryfallanlage(
		service.NeuerScryfallabgleich(&fremdprobe{karte: eineFremdkarte()}, nil, neuesVerweisdoppel(), schreiber),
		anleger)

	ergebnis, fehler := anlage.Anlegen(t.Context(), 1, "magic", "Lightning Bolt")
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if ergebnis.KarteID != 101 {
		t.Errorf("Kennung = %d, erwartet die der angelegten Karte", ergebnis.KarteID)
	}

	if schreiber.werte["name"] != "Lightning Bolt" ||
		schreiber.werte["rule text"] == "" ||
		schreiber.werte["casting cost"] != "R" {
		t.Errorf("geschrieben wurde %v — erwartet die abgebildeten Felder", schreiber.werte)
	}

	if len(ergebnis.Uebernommen) == 0 {
		t.Error("keine Änderungen gemeldet — der Aufrufer kann dann nichts zurücknehmen")
	}

	// **§16, Regel 6**: „trägt danach seine Herkunft."
	if schreiber.herkuenfte["name"] != "https://scryfall.com/card/lea/161/lightning-bolt" {
		t.Errorf("Herkunft von »name« = %q", schreiber.herkuenfte["name"])
	}
}

// **Ein unbekannter Name legt nichts an.**
//
// Zuerst holen, dann anlegen — sonst bliebe bei jedem Tippfehler eine leere
// Karte zurück, und niemand räumt sie weg.
func TestEinUnbekannterNameLegtNichtsAn(t *testing.T) {
	anleger := &anlegeprobe{}

	anlage := service.NeueScryfallanlage(
		service.NeuerScryfallabgleich(&fremdprobe{fehler: domain.ErrNotFound}, nil, neuesVerweisdoppel(), &anlageschreiber{}),
		anleger)

	if _, fehler := anlage.Anlegen(
		t.Context(), 1, "magic", "Gibbes Nicht"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
	}

	if anleger.angelegt != 0 {
		t.Errorf("%d Karten angelegt, erwartet keine — der Name war nicht zu finden",
			anleger.angelegt)
	}
}

// **Scheitert das Füllen, wird die angelegte Karte genannt.**
//
// §7 Regel 3: geschrieben ist geschrieben. Eine Karte, die entstanden ist,
// verschwindet nicht — aber wer sie nicht **nennt**, lässt sie im Bestand
// zurück, ohne dass jemand davon weiß (#155).
func TestEineAngelegteKarteWirdAuchImFehlerfallGenannt(t *testing.T) {
	anleger := &anlegeprobe{}

	anlage := service.NeueScryfallanlage(
		service.NeuerScryfallabgleich(&fremdprobe{karte: eineFremdkarte()}, nil, neuesVerweisdoppel(), &anlageschreiber{fehler: errors.New("der Bestand ist nicht erreichbar")}),
		anleger)

	ergebnis, fehler := anlage.Anlegen(t.Context(), 1, "magic", "Lightning Bolt")
	if fehler == nil {
		t.Fatal("kein Fehler, obwohl das Schreiben scheitert")
	}

	if ergebnis.KarteID != 101 {
		t.Errorf("Kennung = %d — die angelegte Karte muss genannt werden, sonst "+
			"bleibt sie im Bestand, ohne dass jemand davon weiß", ergebnis.KarteID)
	}
}

// **Ein Bestand, der keine Karte anlegen kann, meldet das.**
func TestEinGescheitertesAnlegenWirdGemeldet(t *testing.T) {
	anlage := service.NeueScryfallanlage(
		service.NeuerScryfallabgleich(&fremdprobe{karte: eineFremdkarte()}, nil, neuesVerweisdoppel(), &anlageschreiber{}),
		&anlegeprobe{fehler: errors.New("der Bestand ist voll")})

	if _, fehler := anlage.Anlegen(t.Context(), 1, "magic", "Lightning Bolt"); fehler == nil {
		t.Fatal("kein Fehler, obwohl das Anlegen scheitert")
	}
}
