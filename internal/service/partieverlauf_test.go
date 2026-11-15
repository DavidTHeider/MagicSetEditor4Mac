package service_test

// **Der Verlauf, wie ein Sitz ihn sehen darf** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Die Zusage
//
// „Replay/Chat/Logs eines Teilnehmers enthalten nur seine freigegebenen
// Informationen" und „kein Replay exportiert unbemerkt fremde private
// Informationen."
//
// **Die Filterung rechnet `domain.RueckschauFuer`.** Was hier geprüft wird,
// ist der Weg dorthin: dass er genommen wird, und dass niemand am Speicher
// vorbei an den ungefilterten Strom kommt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerVerlaufZeigtNurDieEigenenErgaenzungen(t *testing.T) {
	t.Parallel()

	leitung, speicher := leitungMitVerlauf(t)

	verlauf, fehler := leitung.Verlauf(context.Background(), "p1", "a")
	if fehler != nil {
		t.Fatalf("Verlauf: %v", fehler)
	}

	switch {
	case len(verlauf) != 2:
		t.Fatalf("%d Ereignisse, erwartet 2", len(verlauf))
	case len(verlauf[0].Privat) != 1:
		t.Errorf("%d eigene Ergänzungen, erwartet 1", len(verlauf[0].Privat))
	case len(verlauf[1].Privat) != 0:
		t.Errorf("%d fremde Ergänzungen, erwartet 0", len(verlauf[1].Privat))
	}

	// **Der Speicher hat den vollen Strom hergegeben** — gefiltert wurde
	// danach, und zwar hier.
	if len(speicher.gelieferte[1].Privat) != 1 {
		t.Error("der Speicher lieferte schon gefiltert — dann prüft dieser " +
			"Lauf nichts")
	}
}

// **Die Gegenprobe** (§10): der andere Sitz sieht das Gegenteil.
func TestDerVerlaufZeigtDemAnderenDasGegenteil(t *testing.T) {
	t.Parallel()

	leitung, _ := leitungMitVerlauf(t)

	verlauf, fehler := leitung.Verlauf(context.Background(), "p1", "b")
	if fehler != nil {
		t.Fatalf("Verlauf: %v", fehler)
	}

	switch {
	case len(verlauf[0].Privat) != 0:
		t.Errorf("%d fremde Ergänzungen, erwartet 0", len(verlauf[0].Privat))
	case len(verlauf[1].Privat) != 1:
		t.Errorf("%d eigene Ergänzungen, erwartet 1", len(verlauf[1].Privat))
	}
}

// **Wer nicht am Tisch sitzt, sieht nichts Privates.**
func TestEinFremderSiehtNichtsPrivates(t *testing.T) {
	t.Parallel()

	leitung, _ := leitungMitVerlauf(t)

	verlauf, fehler := leitung.Verlauf(context.Background(), "p1", "zuschauer-1")
	if fehler != nil {
		t.Fatalf("Verlauf: %v", fehler)
	}

	for _, ereignis := range verlauf {
		if len(ereignis.Privat) != 0 {
			t.Errorf("Folge %d trägt Ergänzungen", ereignis.Folge)
		}
	}
}

// **Was der Speicher klagt, wird gemeldet** (§8).
func TestEinKlagenderSpeicherWirdGemeldet(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")
	speicher := &verlaufsspeicher{klage: kaputt}
	speicher.partien = eineBekanntePartie()
	leitung := service.NeuePartieleitung(speicher, nil)

	if _, fehler := leitung.Verlauf(
		context.Background(), "p1", "a",
	); !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Partie ohne Ereignisse gibt eine leere Liste** — und nicht `nil`,
// damit eine Fläche sie ohne Fallunterscheidung ausgeben kann.
func TestEinLeererVerlaufIstEineLeereListe(t *testing.T) {
	t.Parallel()

	speicher := &verlaufsspeicher{}
	speicher.partien = eineBekanntePartie()
	leitung := service.NeuePartieleitung(speicher, nil)

	verlauf, fehler := leitung.Verlauf(context.Background(), "p1", "a")
	switch {
	case fehler != nil:
		t.Fatalf("Verlauf: %v", fehler)
	case verlauf == nil:
		t.Error("nil statt einer leeren Liste")
	case len(verlauf) != 0:
		t.Errorf("%d Ereignisse, erwartet 0", len(verlauf))
	}
}

// **Eine unbekannte Partie ist kein leerer Verlauf.**
//
// **Der Unterschied zwischen „noch nichts geschehen" und „gibt es nicht"**:
// der Speicher gibt für eine erfundene Kennung eine leere Liste — er fragt
// eine Tabelle nach Zeilen, und keine ist keine. Wer das durchreichte,
// antwortete auf einen Tippfehler mit „in dieser Partie ist noch nichts
// geschehen", und der Fragende suchte danach am falschen Ende (§8).
func TestEineUnbekanntePartieHatKeinenVerlaufSondernKeineExistenz(t *testing.T) {
	t.Parallel()

	leitung, _ := leitungMitVerlauf(t)

	_, fehler := leitung.Verlauf(context.Background(), "gibt-es-nicht", "a")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v, erwartet ErrNotFound", fehler)
	}
}

// verlaufsspeicher gibt einen festen Strom her — **ungefiltert**.
//
// **Er baut auf `partieprobe` auf**, damit dieser Lauf nur die eine Frage
// stellt: wird gefiltert, und wo? Eine zweite vollständige Probe daneben
// liefe beim ersten neuen Speicherweg auseinander.
type verlaufsspeicher struct {
	partieprobe

	gelieferte []domain.Ereignis
	klage      error
}

func (v *verlaufsspeicher) Ereignisse(
	_ context.Context, _ string, _, _ int64,
) ([]domain.Ereignis, error) {
	if v.klage != nil {
		return nil, v.klage
	}

	return v.gelieferte, nil
}

func leitungMitVerlauf(t *testing.T) (*service.Partieleitung, *verlaufsspeicher) {
	t.Helper()

	speicher := &verlaufsspeicher{gelieferte: []domain.Ereignis{
		{
			PartieID: "p1", Folge: 1, Art: "objekt.ziehen",
			Oeffentlich: map[string]any{"sitz": "a"},
			Privat: []domain.Ergaenzung{
				{Empfaenger: "a", Inhalt: map[string]any{"karte": "Wald"}},
			},
		},
		{
			PartieID: "p1", Folge: 2, Art: "objekt.ziehen",
			Oeffentlich: map[string]any{"sitz": "b"},
			Privat: []domain.Ergaenzung{
				{Empfaenger: "b", Inhalt: map[string]any{"karte": "Berg"}},
			},
		},
	}}

	speicher.partien = eineBekanntePartie()

	return service.NeuePartieleitung(speicher, nil), speicher
}

// eineBekanntePartie ist der Eintrag, ohne den jede Rückschau ins Leere
// fragte — die Partie muss es geben, bevor sie einen Verlauf hat.
func eineBekanntePartie() map[string]domain.Partie {
	return map[string]domain.Partie{"p1": {ID: "p1", Fassung: 1}}
}
