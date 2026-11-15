package domain_test

// **Ein lokaler Kontakt** — ID-04
// (#1703).

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func ada() domain.Kontakt {
	return domain.Kontakt{
		Name:     "Ada Lovelace",
		Ursprung: "0A1B",
		Wege: []domain.Kontaktweg{
			{Art: domain.KontaktwegEmail, Wert: "ada@example.org"},
			{Art: domain.KontaktwegTelefon, Wert: "+49 170 0000000"},
		},
	}
}

func TestEinKontaktWirdGeglaettetUndGeprueft(t *testing.T) {
	t.Parallel()

	kontakt := ada()
	kontakt.Name = "  Ada Lovelace "
	kontakt.Wege[0].Wert = " ada@example.org "
	kontakt.Normalisieren()

	if kontakt.Name != "Ada Lovelace" || kontakt.Wege[0].Wert != "ada@example.org" {
		t.Errorf("%+v", kontakt)
	}

	if befunde := kontakt.Validate(); len(befunde) != 0 {
		t.Errorf("ein guter Kontakt: %v", befunde)
	}

	for name, fall := range map[string]struct {
		aendern func(*domain.Kontakt)
		feld    string
	}{
		"ohne Namen":           {func(k *domain.Kontakt) { k.Name = "" }, "name"},
		"ein zu langer Name":   {func(k *domain.Kontakt) { k.Name = strings.Repeat("a", 201) }, "name"},
		"ein leerer Weg":       {func(k *domain.Kontakt) { k.Wege[0].Wert = "" }, "wege"},
		"eine unbekannte Art":  {func(k *domain.Kontakt) { k.Wege[0].Art = "fax" }, "wege"},
		"ein Bild ohne Art":    {func(k *domain.Kontakt) { k.Bild = []byte{1} }, "bild"},
		"ein Bild fremder Art": {func(k *domain.Kontakt) { k.Bild, k.Bildart = []byte{1}, "image/tiff" }, "bild"},
		"ein zu großes Bild": {func(k *domain.Kontakt) {
			k.Bild, k.Bildart = make([]byte, domain.KontaktbildGrenze+1), "image/jpeg"
		}, "bild"},
	} {
		kaputt := ada()
		kaputt.Wege = append([]domain.Kontaktweg(nil), kaputt.Wege...)
		fall.aendern(&kaputt)

		befunde := kaputt.Validate()
		if len(befunde) == 0 || befunde[0].Field != fall.feld || !errors.Is(befunde[0], domain.ErrValidation) {
			t.Errorf("%s: %v", name, befunde)
		}
	}
}

// **Ein zweiter Import legt keine unkontrollierte Dublette an** — ID-04.AK-02.
func TestDerAbgleichFindetDenVorhandenen(t *testing.T) {
	t.Parallel()

	vorhanden := ada()
	vorhanden.ID = 7

	fremd := domain.Kontakt{ID: 8, Name: "Bo", Wege: []domain.Kontaktweg{{Art: domain.KontaktwegEmail, Wert: "bo@example.org"}}}

	for name, fall := range map[string]struct {
		neu    func() domain.Kontakt
		stand  domain.Kontaktabgleich
		treffe int64
	}{
		"derselbe noch einmal": {ada, domain.KontaktGleich, 7},
		"derselbe mit neuer Nummer": {func() domain.Kontakt {
			k := ada()
			k.Wege = []domain.Kontaktweg{{Art: domain.KontaktwegTelefon, Wert: "+49 171 1111111"}}

			return k
		}, domain.KontaktGeaendert, 7},
		"derselbe mit neuem Bild": {func() domain.Kontakt {
			k := ada()
			k.Bild, k.Bildart = []byte{1}, "image/png"

			return k
		}, domain.KontaktGeaendert, 7},
		"ohne Ursprung, gleicher Name und gleiche Adresse in anderer Schreibung": {func() domain.Kontakt {
			k := ada()
			k.Ursprung = ""
			k.Wege = []domain.Kontaktweg{{Art: domain.KontaktwegEmail, Wert: "ADA@example.org"}}

			return k
		}, domain.KontaktVermutlich, 7},
		"ohne Ursprung, gleicher Name und dieselbe Nummer anders geschrieben": {func() domain.Kontakt {
			k := ada()
			k.Ursprung = ""
			k.Wege = []domain.Kontaktweg{{Art: domain.KontaktwegTelefon, Wert: "+49 (170) 000-0000"}}

			return k
		}, domain.KontaktVermutlich, 7},
		"gleicher Name ohne gemeinsamen Weg ist ein anderer Mensch": {func() domain.Kontakt {
			k := ada()
			k.Ursprung = "anders"
			k.Wege = nil

			return k
		}, domain.KontaktNeu, 0},
		"dieselben Angaben ohne Ursprung sind derselbe": {func() domain.Kontakt {
			k := ada()
			k.Ursprung = ""

			return k
		}, domain.KontaktGleich, 7},
		"ein anderer": {func() domain.Kontakt {
			return domain.Kontakt{Name: "Cy", Ursprung: "cy"}
		}, domain.KontaktNeu, 0},
	} {
		stand, getroffen := domain.KontaktAbgleichen([]domain.Kontakt{fremd, vorhanden}, fall.neu())

		var kennung int64
		if getroffen != nil {
			kennung = getroffen.ID
		}

		if stand != fall.stand || kennung != fall.treffe {
			t.Errorf("%s: %s an %d", name, stand, kennung)
		}
	}
}

// **Ein Kontakt ist kein Gegenüber und nichts an ihm ist geprüft** —
// ID-04.AK-04. Der Typ trägt kein Feld, das es behaupten könnte; dieser
// Lauf hält die Felder fest, damit ein neues auffällt.
func TestEinKontaktBehauptetKeinePruefung(t *testing.T) {
	t.Parallel()

	for _, typ := range []reflect.Type{reflect.TypeFor[domain.Kontakt](), reflect.TypeFor[domain.Kontaktweg]()} {
		for feld := range typ.Fields() {
			for _, verboten := range []string{"Bestaetigt", "Geprueft", "Verifiziert", "Peer", "Kennung", "Fingerabdruck"} {
				if strings.Contains(feld.Name, verboten) {
					t.Errorf("%s trägt %q", typ.Name(), feld.Name)
				}
			}
		}
	}
}
