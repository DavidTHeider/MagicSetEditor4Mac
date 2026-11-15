package service_test

// **Die Doppel der Verbuchung** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// **Eine eigene Datei**, damit der Lauf daneben die Läufe zeigt und nicht die
// Vorrichtung (§2).

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// verbuchungsprobe hält Bestand, Zusagen und Belege im Arbeitsspeicher.
type verbuchungsprobe struct {
	stuecke  map[int64]domain.Exemplar
	zusagen  []domain.Reservierung
	belege   map[string]domain.Belegstand
	gebucht  []domain.Handelsabschluss
	klagt    error
	gezeichn []string

	// belegklage und zusagenklage klagen auf den beiden Lesewegen,
	// zeichenklage beim Gegenzeichnen.
	belegklage   error
	zusagenklage error
	zeichenklage error
	druckklage   error

	// stand ist der Vorgangsbestand, den der echte Weg **in derselben
	// Transaktion** mitschreibt.
	//
	// **Ohne ihn wäre dieses Doppel großzügiger als der Betrieb** (§3): der
	// Vorgang bliebe nach dem Buchen dort stehen, wo er vorher war, und der
	// Lauf über die Wiederholung prüfte einen Weg, den es nicht gibt.
	stand *handelsprobe

	// journalklage klagt beim Lesen des Journals.
	journalklage error

	// drucke sind die Namen, die dieser Rechner kennt.
	drucke map[string]domain.Druck
}

func (v *verbuchungsprobe) Journal(
	_ context.Context, vorgang string,
) ([]domain.Journalzeile, error) {
	if v.journalklage != nil {
		return nil, v.journalklage
	}

	zeilen := []domain.Journalzeile{}

	for _, abschluss := range v.gebucht {
		if abschluss.Beleg.Vorgang != vorgang {
			continue
		}

		for stelle, zeile := range abschluss.Beleg.Zeilen {
			zeile.ID = int64(stelle + 1)
			zeilen = append(zeilen, zeile)
		}
	}

	return zeilen, nil
}

func (v *verbuchungsprobe) NachScryfall(
	_ context.Context, printing string,
) (*domain.Druck, error) {
	if v.druckklage != nil {
		return nil, v.druckklage
	}

	druck, da := v.drucke[printing]
	if !da {
		return nil, nil //nolint:nilnil // „kenne ich nicht" ist kein Fehler
	}

	return &druck, nil
}

func (v *verbuchungsprobe) Einzeln(
	_ context.Context, id int64,
) (*domain.Exemplar, error) {
	stueck, da := v.stuecke[id]
	if !da {
		return nil, domain.MitGrund(domain.ErrNotFound, nil,
			"die Besitzzeile gibt es nicht")
	}

	return &stueck, nil
}

func (v *verbuchungsprobe) FuerVorgang(
	_ context.Context, vorgang string,
) ([]domain.Reservierung, error) {
	if v.zusagenklage != nil {
		return nil, v.zusagenklage
	}

	gefunden := []domain.Reservierung{}

	for _, zusage := range v.zusagen {
		if zusage.Vorgang == vorgang {
			gefunden = append(gefunden, zusage)
		}
	}

	return gefunden, nil
}

func (v *verbuchungsprobe) Abschliessen(
	_ context.Context, abschluss domain.Handelsabschluss,
) error {
	if v.klagt != nil {
		return v.klagt
	}

	v.gebucht = append(v.gebucht, abschluss)
	v.stand.staende[abschluss.Angebot.Vorgang] = abschluss.Angebot
	v.belege[abschluss.Beleg.Vorgang] = domain.Belegstand{
		Vorgang: abschluss.Beleg.Vorgang, Abdruck: abschluss.Beleg.Abdruck,
		Form:         string(abschluss.Beleg.Unterschriftsform()),
		Unterschrift: abschluss.Unterschrift,
		Gebucht:      time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC),
	}

	return nil
}

func (v *verbuchungsprobe) Beleg(
	_ context.Context, vorgang string,
) (*domain.Belegstand, error) {
	if v.belegklage != nil {
		return nil, v.belegklage
	}

	stand, da := v.belege[vorgang]
	if !da {
		return nil, nil //nolint:nilnil // „noch nicht" ist hier kein Fehler
	}

	return &stand, nil
}

func (v *verbuchungsprobe) Gegenzeichnen(
	_ context.Context, fremd domain.Fremdbeleg,
) error {
	if v.zeichenklage != nil {
		return v.zeichenklage
	}

	stand, da := v.belege[fremd.Vorgang]
	if !da {
		return domain.MitGrund(domain.ErrNotFound, nil, "kein Beleg")
	}

	stand.Beidseitig = true
	stand.Gegenzeichner = fremd.Von
	v.belege[fremd.Vorgang] = stand
	v.gezeichn = append(v.gezeichn, fremd.Vorgang)

	return nil
}

// festerAusweis gibt immer dieselbe Kennung.
type festerAusweis struct{ ich peer.Identitaet }

func (a festerAusweis) Ausweis(context.Context) (peer.Identitaet, error) {
	return a.ich, nil
}

// derAusweis baut eine Identität aus einer festen Saat.
//
// **Fest und nicht zufällig**: eine Unterschrift, die bei jedem Lauf anders
// ausfällt, lässt sich nicht gegen eine erwartete halten.
func derAusweis(t *testing.T) peer.Identitaet {
	t.Helper()

	saat := make([]byte, ed25519.SeedSize)
	for stelle := range saat {
		saat[stelle] = byte(stelle)
	}

	ich, fehler := peer.AusGeheimnis(
		hex.EncodeToString(ed25519.NewKeyFromSeed(saat)))
	if fehler != nil {
		t.Fatalf("die Identität ließ sich nicht bauen: %v", fehler)
	}

	return ich
}

// eineVerbuchung stellt die Verbuchung samt ihrer Doppel auf.
func eineVerbuchung(t *testing.T) (
	*service.Verbuchungsstelle, *handelsprobe, *verbuchungsprobe,
) {
	t.Helper()

	stand := &handelsprobe{staende: map[string]domain.Handelsangebot{}}
	bestand := &verbuchungsprobe{
		stuecke: map[int64]domain.Exemplar{},
		belege:  map[string]domain.Belegstand{},
		stand:   stand,
		drucke: map[string]domain.Druck{
			"bolt-m10": {
				Name: "Lightning Bolt", Setcode: "m10", Nummer: "146",
			},
		},
	}

	// **Mit williger Wache** (T10, #1475): diese Läufe messen die Verbuchung
	// und nicht den Riegel davor. Dass er greift, misst
	// `handelsfreigabe_test.go` — und dass er **da** ist, hält der Rückfall
	// „ohne Wache wird nicht gebucht", den ein Lauf dort betritt.
	return service.NeueVerbuchungsstelle(
		stand, bestand, festerAusweis{ich: derAusweis(t)}, bestand,
	).MitWache(willigeWache{}), stand, bestand
}

// derFreibrief ist das Freigabezeichen dieser Läufe.
const derFreibrief = "f-1"

// willigeWache lässt jeden Schritt durch.
type willigeWache struct{}

func (willigeWache) PruefeMitNachweis(
	context.Context, domain.Wirkung, string, service.Nachweisbindung,
) error {
	return nil
}

// hexAus liest eine Unterschrift und sagt, ob sie eine war.
func hexAus(t *testing.T, roh string) ([]byte, bool) {
	t.Helper()

	entschluesselt, fehler := hex.DecodeString(roh)
	if fehler != nil {
		t.Errorf("die Unterschrift ist kein Hex: %q", roh)

		return nil, false
	}

	return entschluesselt, true
}
