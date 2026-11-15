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

// **Ein ganzer Satz von Scryfall** —
// #1033.
//
// Bis hierher kamen ganze Sets **nur** über MTG JSON. Damit fehlte der
// Einstieg, der für die meisten der naheliegende ist: „gib mir diesen Satz als
// Set, dann bearbeite ich ihn."

// satzprobe gibt einen festen Satz.
type satzprobe struct {
	satz    scryfall.Satz
	angabe  *scryfall.Setangabe
	befunde []domain.Befund
	fehler  error
}

func (s *satzprobe) Setkarten(context.Context, string) (
	scryfall.Satz, []domain.Befund, error,
) {
	if s.fehler != nil {
		return scryfall.Satz{}, nil, s.fehler
	}

	return s.satz, s.befunde, nil
}

func (s *satzprobe) Setangaben(context.Context, string) (
	*scryfall.Setangabe, time.Duration, []domain.Befund, error,
) {
	return s.angabe, 0, nil, nil
}

// speicherprobe merkt sich das angelegte Set.
type speicherprobe struct {
	set    *domain.Set
	fehler error
}

func (s *speicherprobe) SetAnlegen(_ context.Context, set *domain.Set, _ []domain.RawBlock) (
	int64, error,
) {
	if s.fehler != nil {
		return 0, s.fehler
	}

	s.set = set

	return 42, nil
}

// koepfeprobe sagt, was schon im Bestand liegt.
type koepfeprobe struct{ koepfe []domain.SetHead }

func (k koepfeprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return k.koepfe, nil
}

func einSatz() scryfall.Satz {
	return scryfall.Satz{
		Gesamt: 2,
		Karten: []*scryfall.Karte{
			{Name: "Erste", Manakosten: "{R}", Typzeile: "Instant", Seltenheit: "common"},
			{Name: "Zweite", Typzeile: "Creature — Bird", Staerke: "0", Widerstand: "1"},
		},
	}
}

func TestEinSatzWirdZuEinemSet(t *testing.T) {
	speicher := &speicherprobe{}

	einlesen := service.NeuerSatzimport(&satzprobe{
		satz:   einSatz(),
		angabe: &scryfall.Setangabe{Code: "lea", Name: "Limited Edition Alpha"},
	}, speicher, koepfeprobe{})

	ergebnis, fehler := einlesen.Einlesen(t.Context(), "lea", "magic")
	if fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if ergebnis.SetID != 42 || ergebnis.Karten != 2 {
		t.Errorf("Ergebnis = %+v", ergebnis)
	}

	if speicher.set.Name != "Limited Edition Alpha" {
		t.Errorf("Name = %q — der Name kommt aus den Setangaben", speicher.set.Name)
	}

	if speicher.set.Game != "magic" {
		t.Errorf("Spiel = %q", speicher.set.Game)
	}

	// **Die Herkunft ist die Regel „eine Quelle, ein Set"** (#886).
	if speicher.set.Origin != "scryfall:lea" {
		t.Errorf("Herkunft = %q", speicher.set.Origin)
	}

	if len(speicher.set.Cards) != 2 {
		t.Fatalf("%d Karten", len(speicher.set.Cards))
	}

	// **Dieselbe Zuordnung wie der Einzelabgleich** — eine zweite liefe beim
	// ersten Nachziehen auseinander.
	erste, hat := speicher.set.Cards[0].Value("casting cost")
	if !hat || erste != "R" {
		t.Errorf("Manakosten = %q — erwartet die übersetzte Form", erste)
	}
}

// **Eine Quelle, ein Set** (#886): ein zweiter Aufruf legt nichts Neues an.
func TestDerselbeSatzWirdNichtZweimalEingelesen(t *testing.T) {
	speicher := &speicherprobe{}

	einlesen := service.NeuerSatzimport(&satzprobe{
		satz:   einSatz(),
		angabe: &scryfall.Setangabe{Code: "lea", Name: "Alpha"},
	}, speicher, koepfeprobe{koepfe: []domain.SetHead{
		{ID: 7, Origin: "scryfall:lea", Cards: 2},
	}})

	ergebnis, fehler := einlesen.Einlesen(t.Context(), "lea", "magic")
	if fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if !ergebnis.SchonOffen || ergebnis.SetID != 7 {
		t.Errorf("Ergebnis = %+v, erwartet das vorhandene Set", ergebnis)
	}

	if speicher.set != nil {
		t.Error("es wurde ein zweites Set angelegt")
	}
}

// **Ein Satz, den es nicht gibt, legt nichts an.**
func TestEinUnbekannterSatzLegtNichtsAn(t *testing.T) {
	speicher := &speicherprobe{}

	einlesen := service.NeuerSatzimport(
		&satzprobe{fehler: domain.ErrNotFound}, speicher, koepfeprobe{})

	if _, fehler := einlesen.Einlesen(
		t.Context(), "gibbesnicht", "magic"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
	}

	if speicher.set != nil {
		t.Error("es wurde ein Set angelegt, obwohl es den Satz nicht gibt")
	}
}

// **Ein Spiel ohne Zuordnung wird abgewiesen und nicht geraten.**
func TestEinSpielOhneZuordnungWirdAbgewiesen(t *testing.T) {
	einlesen := service.NeuerSatzimport(&satzprobe{
		satz:   einSatz(),
		angabe: &scryfall.Setangabe{Code: "lea", Name: "Alpha"},
	}, &speicherprobe{}, koepfeprobe{})

	if _, fehler := einlesen.Einlesen(t.Context(), "lea", "schach"); fehler == nil {
		t.Fatal("kein Fehler bei einem Spiel ohne Zuordnung")
	}
}

// **Die Befunde der Antwort reisen mit** (§8).
func TestDieBefundeDesSatzesReisenMit(t *testing.T) {
	einlesen := service.NeuerSatzimport(&satzprobe{
		satz:    einSatz(),
		angabe:  &scryfall.Setangabe{Code: "lea", Name: "Alpha"},
		befunde: []domain.Befund{{Ort: "cards/search", Meldung: "gekürzt"}},
	}, &speicherprobe{}, koepfeprobe{})

	ergebnis, fehler := einlesen.Einlesen(t.Context(), "lea", "magic")
	if fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if len(ergebnis.Befunde) != 1 {
		t.Errorf("%d Befunde, erwartet einer: %v", len(ergebnis.Befunde), ergebnis.Befunde)
	}
}

// **Ein Bestand, der das Set nicht ablegen kann, meldet das.**
func TestEinGescheitertesAblegenWirdGemeldet(t *testing.T) {
	einlesen := service.NeuerSatzimport(&satzprobe{
		satz:   einSatz(),
		angabe: &scryfall.Setangabe{Code: "lea", Name: "Alpha"},
	}, &speicherprobe{fehler: errors.New("der Bestand ist voll")}, koepfeprobe{})

	if _, fehler := einlesen.Einlesen(t.Context(), "lea", "magic"); fehler == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ohne Kopfliste gibt es keine Regel „eine Quelle, ein Set"** — und das ist
// kein Fehler, sondern ihre Abwesenheit.
func TestOhneKopflisteWirdNichtNachgesehen(t *testing.T) {
	speicher := &speicherprobe{}

	einlesen := service.NeuerSatzimport(&satzprobe{
		satz:   einSatz(),
		angabe: &scryfall.Setangabe{Code: "lea", Name: "Alpha"},
	}, speicher, nil)

	if _, fehler := einlesen.Einlesen(t.Context(), "lea", "magic"); fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if speicher.set == nil {
		t.Error("es wurde kein Set angelegt")
	}
}

// **Ein Bestand, der nicht antwortet, hält das Einlesen auf.**
//
// Anders als beim Namen: ohne die Kopfliste ist „eine Quelle, ein Set" nicht
// zu halten, und ein zweites Set wäre schlimmer als eine Meldung.
func TestEinUnlesbarerBestandHaeltDasEinlesenAuf(t *testing.T) {
	einlesen := service.NeuerSatzimport(&satzprobe{satz: einSatz()},
		&speicherprobe{}, stolperndeKoepfe{})

	if _, fehler := einlesen.Einlesen(t.Context(), "lea", "magic"); fehler == nil {
		t.Fatal("kein Fehler")
	}
}

// stolperndeKoepfe antwortet nicht.
type stolperndeKoepfe struct{}

func (stolperndeKoepfe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, errors.New("der Bestand ist nicht erreichbar")
}

// **Ohne Setangaben heißt das Set wie sein Kürzel** (§8).
//
// Der Name ist Beiwerk: ein Satz, dessen Angaben nicht zu holen sind, ist
// immer noch ein brauchbares Set.
func TestOhneSetangabenHeisstDasSetWieSeinKuerzel(t *testing.T) {
	speicher := &speicherprobe{}

	einlesen := service.NeuerSatzimport(
		&satzprobe{satz: einSatz()}, speicher, koepfeprobe{})

	if _, fehler := einlesen.Einlesen(t.Context(), "lea", "magic"); fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if speicher.set.Name != "lea" {
		t.Errorf("Name = %q, erwartet das Kürzel", speicher.set.Name)
	}
}

// **Ein Satz ohne Namen wird abgewiesen** — `ValidateBasics` sagt es.
//
// Er entsteht, wenn Scryfall weder Angaben noch einen brauchbaren Code
// liefert; ein Set ohne Namen stünde namenlos in der Liste.
func TestEinSatzOhneNamenWirdAbgewiesen(t *testing.T) {
	speicher := &speicherprobe{}

	einlesen := service.NeuerSatzimport(
		&satzprobe{satz: einSatz()}, speicher, koepfeprobe{})

	if _, fehler := einlesen.Einlesen(t.Context(), "", "magic"); fehler == nil {
		t.Fatal("kein Fehler bei einem Satz ohne Namen")
	}

	if speicher.set != nil {
		t.Error("es wurde ein namenloses Set angelegt")
	}
}
