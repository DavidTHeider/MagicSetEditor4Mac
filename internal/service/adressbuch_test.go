package service_test

// **Kontakte aus einer vCard übernehmen** — ID-04
// (#1703).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// adressbuchdoppel hält Kontakte im Arbeitsspeicher — **alles oder nichts**
// je Übernahme, wie das Geschäft im Bestand.
type adressbuchdoppel struct {
	kontakte  []domain.Kontakt
	naechste  int64
	fehler    error
	geloescht []int64

	// schreibfehler trifft nur das Übernehmen — nach einem geglückten Lesen.
	schreibfehler error
}

func (d *adressbuchdoppel) Kontakte(context.Context) ([]domain.Kontakt, error) {
	return append([]domain.Kontakt(nil), d.kontakte...), d.fehler
}

func (d *adressbuchdoppel) KontakteUebernehmen(_ context.Context, kontakte []*domain.Kontakt) error {
	if d.fehler != nil || d.schreibfehler != nil {
		return errors.Join(d.fehler, d.schreibfehler)
	}

	for _, kontakt := range kontakte {
		kopie := *kontakt
		if kopie.ID == 0 {
			d.naechste++
			kopie.ID, kopie.Fassung = d.naechste, 1
			d.kontakte = append(d.kontakte, kopie)

			continue
		}

		for stelle := range d.kontakte {
			if d.kontakte[stelle].ID == kopie.ID {
				kopie.Fassung++
				d.kontakte[stelle] = kopie
			}
		}
	}

	return nil
}

func (d *adressbuchdoppel) KontaktLoeschen(_ context.Context, id, _ int64) error {
	d.geloescht = append(d.geloescht, id)

	return d.fehler
}

const zweiVCards = "BEGIN:VCARD\r\nFN:Ada Lovelace\r\nUID:ada\r\nEMAIL:ada@example.org\r\nTEL:+49 170 0\r\nEND:VCARD\r\n" +
	"BEGIN:VCARD\r\nFN:Bo\r\nUID:bo\r\nEND:VCARD\r\n"

func TestDieVorschauNenntJedeKarteMitIhremAbgleich(t *testing.T) {
	t.Parallel()

	doppel := &adressbuchdoppel{kontakte: []domain.Kontakt{{ID: 3, Fassung: 2, Name: "Bo", Ursprung: "bo", Wege: []domain.Kontaktweg{}}}}
	buch := service.NeuesAdressbuch(doppel)

	bericht, fehler := buch.Vorschau(t.Context(), []byte(zweiVCards+"BEGIN:VCARD\r\nEMAIL:x@y\r\nEND:VCARD\r\n"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	if len(bericht.Karten) != 2 || len(bericht.Befunde) != 1 || !strings.Contains(bericht.Befunde[0], "keinen Namen") {
		t.Fatalf("%+v", bericht)
	}

	ada, bo := bericht.Karten[0], bericht.Karten[1]
	if ada.Abgleich != domain.KontaktNeu || ada.Vorhanden != nil || len(ada.Kontakt.Wege) != 2 ||
		ada.Kontakt.Wege[0].Art != domain.KontaktwegEmail || ada.Kontakt.Wege[1].Art != domain.KontaktwegTelefon {
		t.Errorf("Ada %+v", ada)
	}

	if bo.Abgleich != domain.KontaktGleich || bo.Vorhanden == nil || bo.Vorhanden.ID != 3 {
		t.Errorf("Bo %+v", bo)
	}
}

// **Ein Bild, das nicht taugt, bleibt weg — der Kontakt nicht.**
func TestEinUntauglichesBildBleibtWeg(t *testing.T) {
	t.Parallel()

	gross := strings.Repeat("A", 4*(domain.KontaktbildGrenze/3+2))
	karte := "BEGIN:VCARD\r\nFN:Ada\r\nPHOTO;ENCODING=b;TYPE=PNG:" + gross + "\r\nEND:VCARD\r\n"

	bericht, fehler := service.NeuesAdressbuch(&adressbuchdoppel{}).Vorschau(t.Context(), []byte(karte))
	if fehler != nil || len(bericht.Karten) != 1 {
		t.Fatalf("%+v %v", bericht, fehler)
	}

	if len(bericht.Karten[0].Kontakt.Bild) != 0 || len(bericht.Befunde) != 1 || !strings.Contains(bericht.Befunde[0], "Bild") {
		t.Errorf("%+v", bericht)
	}
}

func TestEinNameUeberDerGrenzeWirdGenannt(t *testing.T) {
	t.Parallel()

	karte := "BEGIN:VCARD\r\nFN:" + strings.Repeat("a", 201) + "\r\nEND:VCARD\r\n"

	bericht, fehler := service.NeuesAdressbuch(&adressbuchdoppel{}).Vorschau(t.Context(), []byte(karte))
	if fehler != nil || len(bericht.Karten) != 0 || len(bericht.Befunde) != 1 {
		t.Errorf("%+v %v", bericht, fehler)
	}
}

func TestUebernehmenLegtAnUndErsetzt(t *testing.T) {
	t.Parallel()

	doppel := &adressbuchdoppel{kontakte: []domain.Kontakt{{ID: 3, Fassung: 2, Name: "Bo", Ursprung: "bo"}}, naechste: 3}
	buch := service.NeuesAdressbuch(doppel)

	angelegt, ersetzt, fehler := buch.Uebernehmen(t.Context(), []domain.Kontakt{
		{Name: " Ada ", Ursprung: "ada", Wege: []domain.Kontaktweg{{Art: domain.KontaktwegEmail, Wert: "ada@example.org"}}},
		{ID: 3, Fassung: 2, Name: "Bo Neu", Ursprung: "bo"},
	})
	if fehler != nil || angelegt != 1 || ersetzt != 1 {
		t.Fatalf("%d/%d %v", angelegt, ersetzt, fehler)
	}

	if doppel.kontakte[0].Name != "Bo Neu" || doppel.kontakte[1].Name != "Ada" {
		t.Errorf("%+v", doppel.kontakte)
	}
}

// **Ein zweiter Import legt keine unkontrollierte Dublette an** (AK-02): wer
// eine Karte, die es schon gibt, als neu übernehmen will, wird abgewiesen —
// und nichts geht hinein.
func TestEineVorhandeneKarteWirdNichtNochEinmalAngelegt(t *testing.T) {
	t.Parallel()

	doppel := &adressbuchdoppel{kontakte: []domain.Kontakt{{ID: 3, Fassung: 2, Name: "Bo", Ursprung: "bo"}}}

	_, _, fehler := service.NeuesAdressbuch(doppel).Uebernehmen(t.Context(), []domain.Kontakt{
		{Name: "Ada"}, {Name: "Bo", Ursprung: "bo"},
	})
	if !errors.Is(fehler, domain.ErrDoppelt) || len(doppel.kontakte) != 1 {
		t.Errorf("%v, %d Kontakte", fehler, len(doppel.kontakte))
	}
}

func TestUebernehmenPrueftVorDemSchreiben(t *testing.T) {
	t.Parallel()

	doppel := &adressbuchdoppel{}

	_, _, fehler := service.NeuesAdressbuch(doppel).Uebernehmen(t.Context(), []domain.Kontakt{{Name: "Ada"}, {Name: ""}})
	if !errors.Is(fehler, domain.ErrValidation) || len(doppel.kontakte) != 0 {
		t.Errorf("%v, %d Kontakte", fehler, len(doppel.kontakte))
	}
}

func TestDieFehlerDerAblageReichenDurch(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("die Platte klemmt")
	buch := service.NeuesAdressbuch(&adressbuchdoppel{fehler: klemmt})

	if _, fehler := buch.Vorschau(t.Context(), []byte(zweiVCards)); !errors.Is(fehler, klemmt) {
		t.Errorf("Vorschau: %v", fehler)
	}

	if _, _, fehler := buch.Uebernehmen(t.Context(), []domain.Kontakt{{Name: "Ada"}}); !errors.Is(fehler, klemmt) {
		t.Errorf("Übernehmen: %v", fehler)
	}

	if _, fehler := buch.Liste(t.Context()); !errors.Is(fehler, klemmt) {
		t.Errorf("Liste: %v", fehler)
	}

	if fehler := buch.Loeschen(t.Context(), 3, 1); !errors.Is(fehler, klemmt) {
		t.Errorf("Löschen: %v", fehler)
	}

	schreiben := service.NeuesAdressbuch(&adressbuchdoppel{schreibfehler: klemmt})
	if angelegt, _, fehler := schreiben.Uebernehmen(t.Context(), []domain.Kontakt{{Name: "Ada"}}); !errors.Is(fehler, klemmt) || angelegt != 0 {
		t.Errorf("Schreiben: %d %v", angelegt, fehler)
	}
}

// **Auch innerhalb einer Datei und einer Auswahl keine Dublette** — Befund 1
// des Zweitblicks über ID-04: zwei gleiche Karten ohne UID in einer Datei
// kamen beide als neu durch.
func TestEineDateiVerdoppeltSichNichtSelbst(t *testing.T) {
	t.Parallel()

	karte := "BEGIN:VCARD\r\nFN:Ada\r\nEMAIL:ada@example.org\r\nEND:VCARD\r\n"

	bericht, fehler := service.NeuesAdressbuch(&adressbuchdoppel{}).Vorschau(t.Context(), []byte(karte+karte))
	if fehler != nil || len(bericht.Karten) != 1 || len(bericht.Befunde) != 1 || !strings.Contains(bericht.Befunde[0], "wiederholt") {
		t.Errorf("%+v %v", bericht, fehler)
	}

	doppel := &adressbuchdoppel{}
	ada := domain.Kontakt{Name: "Ada", Wege: []domain.Kontaktweg{{Art: domain.KontaktwegEmail, Wert: "ada@example.org"}}}

	if _, _, fehler := service.NeuesAdressbuch(doppel).Uebernehmen(t.Context(), []domain.Kontakt{ada, ada}); !errors.Is(fehler, domain.ErrDoppelt) || len(doppel.kontakte) != 0 {
		t.Errorf("zweimal in einer Auswahl: %v, %d Kontakte", fehler, len(doppel.kontakte))
	}
}

// **Eine Vermutung wird nicht als neu angelegt** — wer sie übernimmt,
// ersetzt den vermutlich selben Kontakt.
func TestEineVermutungWirdNichtNeuAngelegt(t *testing.T) {
	t.Parallel()

	doppel := &adressbuchdoppel{kontakte: []domain.Kontakt{{ID: 3, Fassung: 1, Name: "Ada", Wege: []domain.Kontaktweg{{Art: domain.KontaktwegEmail, Wert: "ada@example.org"}}}}}

	_, _, fehler := service.NeuesAdressbuch(doppel).Uebernehmen(t.Context(), []domain.Kontakt{
		{Name: "Ada", Wege: []domain.Kontaktweg{{Art: domain.KontaktwegEmail, Wert: "ADA@example.org"}}},
	})
	if !errors.Is(fehler, domain.ErrDoppelt) || len(doppel.kontakte) != 1 {
		t.Errorf("%v, %d Kontakte", fehler, len(doppel.kontakte))
	}
}
