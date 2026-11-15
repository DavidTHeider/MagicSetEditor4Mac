package storage_test

// **Die Ablage der Drucke** — Werkstück A2 von Phase 5.3 (#1441), 0086.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func scryfallkennung(wert string) domain.Druckkennung {
	return domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: wert}
}

func TestEinDruckWirdAbgelegtUndUeberSeineKennungGefunden(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDrucke(db)

	druck := &domain.Druck{
		Name: "Lightning Bolt", Setcode: "lea", Nummer: "161", Sprache: "en",
		Kennungen: []domain.Druckkennung{
			scryfallkennung("E3285E6B-3E79-4D7C-BF96-D920F973B122"),
			{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"},
			{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1"},
		},
	}

	if err := ablage.Sichern(probekontext(t), druck); err != nil {
		t.Fatalf("sichern: %v", err)
	}

	if druck.ID == 0 {
		t.Fatal("der Druck hat keine Kennung bekommen")
	}

	gelesen, err := ablage.Fuer(probekontext(t), druck.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Name != "Lightning Bolt" || gelesen.Setcode != "lea" || len(gelesen.Kennungen) != 3 {
		t.Errorf("gelesen: %+v", gelesen)
	}

	// **Normalisiert abgelegt**: die UUID in kleinen Buchstaben.
	if gelesen.Scryfall() != "e3285e6b-3e79-4d7c-bf96-d920f973b122" {
		t.Errorf("der Alias ist %q", gelesen.Scryfall())
	}

	// **Über die Kennung, in beliebiger Schreibung.**
	gefunden, err := ablage.NachKennung(probekontext(t), scryfallkennung("e3285e6b-3e79-4d7c-bf96-d920f973b122"))
	if err != nil || len(gefunden) != 1 || gefunden[0].ID != druck.ID {
		t.Errorf("nach Kennung: %v, %v", gefunden, err)
	}

	// **Und über den Alias.**
	alias, err := ablage.NachScryfall(probekontext(t), "e3285e6b-3e79-4d7c-bf96-d920f973b122")
	if err != nil || alias == nil || alias.ID != druck.ID {
		t.Errorf("nach Scryfall: %v, %v", alias, err)
	}

	if keiner, err := ablage.NachScryfall(probekontext(t), "gibtsnicht"); err != nil || keiner != nil {
		t.Errorf("ein unbekannter Alias gibt %v, %v — erwartet nil ohne Fehler", keiner, err)
	}

	if _, err := ablage.Fuer(probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein unbekannter Druck gibt %v — erwartet ErrNotFound", err)
	}
}

// **Sichern eines Drucks, den es über seine Kennung schon gibt, ist ein
// Wiedersehen, kein zweiter Druck** — und neue Kennungen hängen sich an.
func TestSichernFindetDenDruckWiederUndHaengtKennungenAn(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDrucke(db)

	erster := &domain.Druck{Name: "Bolt", Kennungen: []domain.Druckkennung{scryfallkennung("e3285e6b-3e79-4d7c-bf96-d920f973b122")}}
	if err := ablage.Sichern(probekontext(t), erster); err != nil {
		t.Fatalf("erster: %v", err)
	}

	zweiter := &domain.Druck{Name: "Lightning Bolt", Setcode: "lea", Nummer: "161", Kennungen: []domain.Druckkennung{
		scryfallkennung("e3285e6b-3e79-4d7c-bf96-d920f973b122"),
		{System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck, Wert: "5f8287b1-5bb6-5f4c-ad17-316a40d5bb0c"},
	}}
	if err := ablage.Sichern(probekontext(t), zweiter); err != nil {
		t.Fatalf("zweiter: %v", err)
	}

	if zweiter.ID != erster.ID {
		t.Fatalf("zwei Drucke (%d, %d) für dieselbe Scryfall-Kennung", erster.ID, zweiter.ID)
	}

	gelesen, err := ablage.Fuer(probekontext(t), erster.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Der Name wird nachgetragen, wo er fehlte** — und die zweite Kennung hängt.
	if gelesen.Name != "Lightning Bolt" || gelesen.Setcode != "lea" || len(gelesen.Kennungen) != 2 {
		t.Errorf("gelesen: %+v", gelesen)
	}

	// **Noch einmal dieselbe Kennung ist kein Fehler und keine dritte Zeile.**
	if err := ablage.Sichern(probekontext(t), zweiter); err != nil {
		t.Fatalf("dritter: %v", err)
	}

	if alle, err := ablage.Alle(probekontext(t)); err != nil || len(alle) != 1 {
		t.Errorf("alle: %d Drucke, %v", len(alle), err)
	}
}

// **Eine eindeutige Kennung, die schon einem anderen Druck gehört, ist ein
// Konflikt** — kein stilles Umhängen.
func TestEineFremdeDruckkennungIstEinKonflikt(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDrucke(db)

	alpha := &domain.Druck{Name: "Bolt", Setcode: "lea", Kennungen: []domain.Druckkennung{scryfallkennung("e3285e6b-3e79-4d7c-bf96-d920f973b122")}}
	// **Beta trägt eine MTGJSON-Kennung** — so bleibt der Druck selbst gültig,
	// wenn er Alphas Scryfall-Kennung dazubekommen soll; zwei Scryfall-Kennungen
	// an einem Druck wiese schon `domain` ab.
	mtgjson := domain.Druckkennung{System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck, Wert: "5f8287b1-5bb6-5f4c-ad17-316a40d5bb0c"}
	beta := &domain.Druck{Name: "Bolt", Setcode: "leb", Kennungen: []domain.Druckkennung{mtgjson}}

	for _, druck := range []*domain.Druck{alpha, beta} {
		if err := ablage.Sichern(probekontext(t), druck); err != nil {
			t.Fatalf("sichern: %v", err)
		}
	}

	// Beta soll Alphas Scryfall-Kennung dazubekommen — das geht nicht.
	beta.Kennungen = append(beta.Kennungen, scryfallkennung("e3285e6b-3e79-4d7c-bf96-d920f973b122"))

	err := ablage.Sichern(probekontext(t), beta)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — erwartet ErrConflict", err)
	}

	// **Ein Konzept darf sich wiederholen**, ein Produkt auch.
	konzept := domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"}
	alpha.Kennungen = append(alpha.Kennungen, konzept)
	beta.Kennungen = []domain.Druckkennung{mtgjson, konzept}

	for _, druck := range []*domain.Druck{alpha, beta} {
		if fehler := ablage.Sichern(probekontext(t), druck); fehler != nil {
			t.Fatalf("Konzept an beiden: %v", fehler)
		}
	}

	beide, err := ablage.NachKennung(probekontext(t), konzept)
	if err != nil || len(beide) != 2 {
		t.Errorf("das Konzept trifft %d Drucke, %v — erwartet zwei", len(beide), err)
	}
}

// **Der Rückfall findet über Name, Setcode, Nummer und Sprache** — die
// Kandidaten für [domain.LoeseAuf], wenn keine Kennung trifft.
func TestKandidatenKommenUeberDenRueckfall(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDrucke(db)

	for _, druck := range []*domain.Druck{
		{Name: "Lightning Bolt", Setcode: "lea", Nummer: "161", Sprache: "en"},
		{Name: "Lightning Bolt", Setcode: "lea", Nummer: "161", Sprache: "de"},
		{Name: "Counterspell", Setcode: "lea", Nummer: "54", Sprache: "en"},
	} {
		if err := ablage.Sichern(probekontext(t), druck); err != nil {
			t.Fatalf("sichern: %v", err)
		}
	}

	kandidaten, err := ablage.Kandidaten(probekontext(t), "lightning bolt", "LEA", "161")
	if err != nil || len(kandidaten) != 2 {
		t.Fatalf("Kandidaten: %d, %v — erwartet zwei Sprachen", len(kandidaten), err)
	}

	verweis := domain.Druckverweis{Fassung: 1, Name: "Lightning Bolt", Setcode: "lea", Nummer: "161", Sprache: "de"}
	if ergebnis := domain.LoeseAuf(kandidaten, verweis); ergebnis.Art != domain.AufloesungEindeutig {
		t.Errorf("Ergebnis %+v", ergebnis)
	}
}

func TestEinUngueltigerDruckWirdNichtAbgelegt(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueDrucke(frischeDatenbank(t))

	err := ablage.Sichern(probekontext(t), &domain.Druck{})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — erwartet ErrValidation", err)
	}
}
