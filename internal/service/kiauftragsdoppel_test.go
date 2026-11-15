package service_test

// **Die Doppel des Auftragswerks** — abgetrennt am 2026-09-11, weil
// `kiauftragswerk_test.go` mit ihnen über die Zeilengrenze aus §2 ging
// (A6, #1445).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen **die Läufe** — was zugesagt ist und warum. Hier stehen die
// **Ablagen und Anbieter**, die sich auf Wunsch danebenbenehmen: jede Flagge
// ist ein Verhalten, das im Betrieb vorkommt.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einWerk baut ein Auftragswerk mit den Doppeln.
func einWerk(t *testing.T, karten *kiKartenprobe) *service.KIAuftragswerk {
	t.Helper()

	return einWerkMit(t, karten)
}

// einWerkMit nimmt **irgendeine** Ablage.
//
// **Seit A13.1b** (#1452): die Gruppe braucht ein Doppel, das
// [service.KIKarten.FelderGruppeSetzen] anders beantwortet, und ein Doppel,
// das die Einzelprobe **einbettet**, ist kein `*kiKartenprobe` mehr.
func einWerkMit(t *testing.T, karten service.KIKarten) *service.KIAuftragswerk {
	t.Helper()

	return service.NeuesKIAuftragswerk(
		derKatalog(t), karten, &kostenprobe{}, &schemaprobe{},
		service.KIOrtsregel{}, func() time.Time { return derLauf })
}

// derKatalog gibt einen Katalog mit einem lokalen und einem Cloudmodell.
func derKatalog(t *testing.T) *service.Modellkatalog {
	t.Helper()

	lokal := einKatalogmodell("schreibt:8b", domain.KITextgenerierung)
	wolke := einKatalogmodell("wolke:gross", domain.KITextgenerierung)
	wolke.Ort = domain.KIOrtCloud

	katalog, fehler := service.NeuerModellkatalog(lokal, wolke)
	if fehler != nil {
		t.Fatalf("der Katalog: %v", fehler)
	}

	return katalog
}

// kiKartenprobe ist eine Karte mit einer Revision — und einem Zähler.
type kiKartenprobe struct {
	revision         string
	schreibvorgaenge int
	fehler           error
	schreibfehler    error

	// geschrieben und herkunft halten fest, was wirklich ankam (A11, #1450).
	//
	// **Beides, weil beides zugesagt ist**: die Auswahl entscheidet über das
	// erste, und was geschrieben wird, trägt das zweite.
	geschrieben map[string][]string
	herkunft    []domain.KIFeldherkunft

	// erwartet hält fest, welcher Stand als Bedingung mitging (Befund 2 des
	// Zweitblicks über Phase 5.3).
	erwartet []string
}

func (k *kiKartenprobe) Revision(_ context.Context, _ int64) (string, error) {
	if k.fehler != nil {
		return "", k.fehler
	}

	return k.revision, nil
}

func (k *kiKartenprobe) FelderSetzen(
	_ context.Context, _ int64, felder map[string][]string,
	herkunft []domain.KIFeldherkunft, erwartet string,
) error {
	k.erwartet = append(k.erwartet, erwartet)

	if k.schreibfehler != nil {
		return k.schreibfehler
	}

	if k.fehler != nil {
		return k.fehler
	}

	if len(felder) > 0 {
		k.schreibvorgaenge++
	}

	k.geschrieben = felder
	k.herkunft = herkunft

	return nil
}

// FelderGruppeSetzen gibt es an der Einzelprobe **nur der Schnittstelle
// wegen**.
//
// **Sie scheitert laut.** Ein Doppel, das hier still gutginge, ließe einen
// Lauf über Gruppen grün werden, der gar keine Gruppe geschrieben hat — §3:
// eine Vorrichtung, die großzügiger ist als der Betrieb. Wer Gruppen prüft,
// nimmt `gruppenkarten` (kigruppe_test.go).
func (k *kiKartenprobe) FelderGruppeSetzen(
	context.Context, []service.KIGruppenschritt,
) error {
	return errors.New("diese Probe schreibt keine Gruppen — nimm gruppenkarten")
}

// kostenprobe zählt Reservierungen.
type kostenprobe struct {
	reservierungen int
	fehler         error
}

func (k *kostenprobe) Reservieren(_ context.Context, kennung string) (string, error) {
	if k.fehler != nil {
		return "", k.fehler
	}

	k.reservierungen++

	return "kosten-" + kennung, nil
}

// schemaprobe gibt immer dasselbe Schema — bis jemand die Fassung wechselt.
type schemaprobe struct {
	fehler        error
	andereFassung bool
}

func (s *schemaprobe) Schema(_ context.Context, _ int64) (domain.KISchema, error) {
	if s.fehler != nil {
		return domain.KISchema{}, s.fehler
	}

	schema := einSchemaprobe()

	if s.andereFassung {
		schema.Fassung = "2026-10-01"
	}

	return schema, nil
}

// einSchemaprobe baut das Schema, gegen das geprüft wird.
func einSchemaprobe() domain.KISchema {
	return domain.KISchema{
		Paket: "magic-m15", Fassung: "2026-09-11",
		// **Drei Felder seit A11** (#1450): mit einem ließe sich eine
		// **Auswahl** nicht prüfen — sie wäre immer die ganze.
		Felder: []domain.KISchemafeld{
			{Name: "name", Typ: domain.KIText, Schreibbar: true},
			{Name: "text", Typ: domain.KIText, Schreibbar: true},
			{Name: "type", Typ: domain.KIText, Schreibbar: true},
		},
	}
}

// **Jeder Fehler der Ablagen reist weiter — und dann wird nichts reserviert.**
//
// **Die zweite Hälfte ist der Punkt**: ein Auftrag, der an der Karte oder am
// Schema scheitert, darf kein Budget verbrauchen. Wer erst bucht und dann
// prüft, zahlt für Aufträge, die nie hinausgingen.
func TestEinFehlerDerAblagenHaeltDasAnlegenAuf(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Ablage klemmt")

	for name, bauen := range map[string]func(*kostenprobe) *service.KIAuftragswerk{
		"die Karte": func(kosten *kostenprobe) *service.KIAuftragswerk {
			return service.NeuesKIAuftragswerk(derKatalog(t),
				&kiKartenprobe{fehler: eigen}, kosten, &schemaprobe{},
				service.KIOrtsregel{}, func() time.Time { return derLauf })
		},
		"das Schema": func(kosten *kostenprobe) *service.KIAuftragswerk {
			return service.NeuesKIAuftragswerk(derKatalog(t),
				&kiKartenprobe{revision: "stand-abc"}, kosten,
				&schemaprobe{fehler: eigen}, service.KIOrtsregel{},
				func() time.Time { return derLauf })
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kosten := &kostenprobe{}

			if _, fehler := bauen(kosten).Anlegen(
				t.Context(), einWunsch()); !errors.Is(fehler, eigen) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if kosten.reservierungen != 0 {
				t.Errorf("%d Reservierungen trotz Fehler", kosten.reservierungen)
			}
		})
	}
}

// **Ein Fehler des Budgets hält das Anlegen auf.**
func TestEinFehlerDesBudgetsHaeltDasAnlegenAuf(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Grenze trägt das nicht")

	werk := service.NeuesKIAuftragswerk(derKatalog(t),
		&kiKartenprobe{revision: "stand-abc"}, &kostenprobe{fehler: eigen},
		&schemaprobe{}, service.KIOrtsregel{},
		func() time.Time { return derLauf })

	if _, fehler := werk.Anlegen(t.Context(), einWunsch()); !errors.Is(fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Vorschlag zu einem Auftrag, den es nicht gibt, wird abgewiesen.**
func TestEinVorschlagOhneAuftragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := einWerk(t, &kiKartenprobe{revision: "stand-abc"}).
		Vorschlagen(t.Context(), domain.KIVorschlag{
			Auftrag: "gibtesnicht",
		}); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Fehler des Schemas beim Vorschlagen ebenso.**
func TestEinSchemafehlerBeimVorschlagenReistWeiter(t *testing.T) {
	t.Parallel()

	eigen := errors.New("das Paket ist weg")
	schema := &schemaprobe{}

	werk := service.NeuesKIAuftragswerk(derKatalog(t),
		&kiKartenprobe{revision: "stand-abc"}, &kostenprobe{}, schema,
		service.KIOrtsregel{}, func() time.Time { return derLauf })

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	schema.fehler = eigen

	if _, fehler := werk.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung,
	}); !errors.Is(fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Eine Übernahme zu einem Auftrag, den es nicht gibt, wird abgewiesen.**
func TestEineUebernahmeOhneAuftragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := einWerk(t, &kiKartenprobe{revision: "stand-abc"}).
		Uebernehmen(t.Context(), "gibtesnicht", []string{"name"}, 7); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein geändertes Schema hält die Übernahme auf.**
//
// **Nicht dasselbe wie eine geänderte Karte**: hier ist die Karte dieselbe,
// und die **Felder** sind andere geworden — ein Stilwechsel etwa. Der
// Vorschlag wurde gegen andere geprüft, und was damals erlaubt war, muss es
// jetzt nicht sein.
func TestEinGeaendertesSchemaHaeltDieUebernahmeAuf(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	schema := &schemaprobe{}

	werk := service.NeuesKIAuftragswerk(derKatalog(t), karten, &kostenprobe{},
		schema, service.KIOrtsregel{}, func() time.Time { return derLauf })

	auftrag, _ := werk.Anlegen(t.Context(), einWunsch())     //nolint:errcheck // geprüft
	_, _ = werk.Vorschlagen(t.Context(), domain.KIVorschlag{ //nolint:errcheck // geprüft
		Auftrag: auftrag.Kennung,
		Felder:  map[string][]string{"name": {"Blitzschlag"}},
	})

	schema.andereFassung = true

	if _, fehler := werk.Uebernehmen(t.Context(), auftrag.Kennung, []string{"name"}, 7); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if karten.schreibvorgaenge != 0 {
		t.Error("trotz geändertem Schema wurde geschrieben")
	}
}

// **Und ein Fehler beim Schreiben bleibt einer.**
func TestEinSchreibfehlerBeiDerUebernahmeBleibtEiner(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Karte ließ sich nicht schreiben")
	karten := &kiKartenprobe{revision: "stand-abc"}

	werk := einWerk(t, karten)

	auftrag, _ := werk.Anlegen(t.Context(), einWunsch())     //nolint:errcheck // geprüft
	_, _ = werk.Vorschlagen(t.Context(), domain.KIVorschlag{ //nolint:errcheck // geprüft
		Auftrag: auftrag.Kennung,
		Felder:  map[string][]string{"name": {"Blitzschlag"}},
	})

	karten.schreibfehler = eigen

	if _, fehler := werk.Uebernehmen(t.Context(), auftrag.Kennung, []string{"name"}, 7); !errors.Is(
		fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
