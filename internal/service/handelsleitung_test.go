package service_test

// **Die Handelsleitung** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was hier gemessen wird
//
// **Dass eine Ablehnung nichts schreibt.** Das ist die erste beobachtbare
// Abnahme des Issues: „ohne Bestandswirkung". Ein Doppel, das seine Schreibe
// zählt, macht daraus eine Messung statt einer Beteuerung — dieselbe Gestalt
// wie bei `partieprobe`.
//
// # Und die Wiederholung
//
// **Derselbe Vorschlag zweimal ist einmal.** Über eine Leitung geht eine
// Nachricht verloren und kommt noch einmal; der zweite Ruf bekommt denselben
// Stand und **keinen** Konflikt (§7, Regel 4).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// handelsprobe hält Vorgänge im Arbeitsspeicher und zählt die Schreibe.
type handelsprobe struct {
	staende     map[string]domain.Handelsangebot
	geschrieben int

	// klagt klagt beim **Lesen** — und zwar mit einem Fehler, der nicht
	// [domain.ErrNotFound] ist: „gibt es nicht" ist beim Eröffnen die gute
	// Antwort, und die beiden auseinanderzuhalten ist der Punkt.
	klagt error

	// schreibklage klagt beim **Schreiben**.
	schreibklage error

	// letzte ist, was `Letzte` gibt — **fest gesetzt** und nicht aus den
	// Ständen gelesen: deren Reihenfolge ist die einer Map.
	letzte []domain.Handelsangebot

	// dazwischen schreibt ein zweiter Schreiber zwischen Lesen und
	// Schreiben — der Fall, den `Fortschreiben` abweist (TRADE-03).
	dazwischen func(staende map[string]domain.Handelsangebot)
}

// Protokoll gibt einen festen Wechsel — die Regel steht im Bestand.
func (h *handelsprobe) Protokoll(
	context.Context, string,
) ([]domain.Handelsprotokollzeile, error) {
	return []domain.Handelsprotokollzeile{{
		Revision: 1, Nach: domain.HandelEntwurf,
	}}, h.klagt
}

// Fortschreiben prüft wie der Bestand: **Zustand und Stimmen** der Revision,
// von der aus geschrieben wird, müssen noch dastehen — und eine neue
// Revision entsteht einmal. Großzügiger als der Betrieb prüfte es sich
// selbst (§3).
func (h *handelsprobe) Fortschreiben(
	ctx context.Context, vorher, danach domain.Handelsangebot,
) error {
	if h.dazwischen != nil {
		h.dazwischen(h.staende)
	}

	// **Ohne Vorstand wird nur angelegt**, was es noch nicht gibt (Befund N2
	// des Nachlaufs über Phase 6.8).
	if vorher.Revision == 0 {
		if _, da := h.staende[danach.Vorgang]; da {
			return domain.ErrUeberholt
		}

		return h.Schreiben(ctx, danach)
	}

	heute := h.staende[vorher.Vorgang]

	if heute.Revision != vorher.Revision || heute.Zustand != vorher.Zustand ||
		len(heute.Zustimmungen) != len(vorher.Zustimmungen) {
		return domain.ErrUeberholt
	}

	return h.Schreiben(ctx, danach)
}

func (h *handelsprobe) Letzte(context.Context) ([]domain.Handelsangebot, error) {
	if h.klagt != nil {
		return nil, h.klagt
	}

	return h.letzte, nil
}

func (h *handelsprobe) Vorgang(
	_ context.Context, kennung string,
) (domain.Handelsangebot, error) {
	if h.klagt != nil {
		return domain.Handelsangebot{}, h.klagt
	}

	stand, da := h.staende[kennung]
	if !da {
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrNotFound,
			nil, "den Handelsvorgang gibt es nicht")
	}

	return stand, nil
}

func (h *handelsprobe) Schreiben(
	_ context.Context, angebot domain.Handelsangebot,
) error {
	if h.schreibklage != nil {
		return h.schreibklage
	}

	h.geschrieben++

	if h.staende == nil {
		h.staende = map[string]domain.Handelsangebot{}
	}

	h.staende[angebot.Vorgang] = angebot

	return nil
}

// eineHandelsleitung gibt sie über einer leeren Ablage.
func eineHandelsleitung() (*service.Handelsleitung, *handelsprobe) {
	ablage := &handelsprobe{staende: map[string]domain.Handelsangebot{}}

	return service.NeueHandelsleitung(ablage, &zusagenhaltprobe{}), ablage
}

// einHandelsangebot ist das kleinste, das prüft.
func einHandelsangebot() domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         "h1",
		Revision:        1,
		Art:             domain.HandelTausch,
		Initiator:       "anna",
		Parteien:        []string{"anna", "bert"},
		Zustand:         domain.HandelEntwurf,
		Positionen: []domain.Handelsposition{{
			Von: "anna", Printing: "abc-123", Menge: 2,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
	}
}

// angebotenerHandel eröffnet einen Vorgang und bietet ihn an.
//
// **Ein Vorschlag setzt ein Angebot voraus** — aus einem Entwurf führt kein
// Gegenangebot (QA-06b, #1967).
// Die Läufe über Vorschläge begannen bis dahin beim Entwurf, weil die Leitung
// ihn durchließ.
func angebotenerHandel(t *testing.T, leitung *service.Handelsleitung) domain.Handelsangebot {
	t.Helper()

	if _, fehler := leitung.Eroeffnen(t.Context(), einHandelsangebot()); fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	angeboten, fehler := leitung.Schreiten(t.Context(), "h1", domain.SchrittAnbieten, "anna")
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	return angeboten
}

// **Ein Vorgang wird eröffnet.**
func TestEinHandelsvorgangWirdEroeffnet(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	angelegt, fehler := leitung.Eroeffnen(
		context.Background(), einHandelsangebot())
	if fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	if angelegt.Revision != 1 || angelegt.Zustand != domain.HandelEntwurf {
		t.Errorf("eröffnet wurde %+v", angelegt)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}

// **Auch nicht, wenn ein zweiter ihn zwischen Nachschlag und Anlage
// eröffnet** — Befund N2 des Nachlaufs über Phase 6.8: `Schreiben` ersetzte
// dann die erste Revision, und beide bekamen Erfolg.
func TestEineEroeffnungImWettlaufIstEinKonflikt(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	ablage.dazwischen = func(staende map[string]domain.Handelsangebot) {
		zuerst := einHandelsangebot()
		zuerst.Positionen[0].Printing = "zuerst-da"
		staende[zuerst.Vorgang] = zuerst
	}

	if _, fehler := leitung.Eroeffnen(context.Background(),
		einHandelsangebot()); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.staende[einHandelsangebot().Vorgang].Positionen[0].Printing != "zuerst-da" {
		t.Error("die zweite Eröffnung hat die erste ersetzt")
	}
}

// **Ein Vorgang, der schon läuft, wird nicht ein zweites Mal eröffnet.**
func TestEinLaufenderVorgangWirdNichtNochEinmalEroeffnet(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	ctx := context.Background()

	if _, fehler := leitung.Eroeffnen(ctx, einHandelsangebot()); fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	vorher := ablage.geschrieben

	if _, fehler := leitung.Eroeffnen(
		ctx, einHandelsangebot()); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != vorher {
		t.Errorf("die Ablehnung hat geschrieben: %d statt %d",
			ablage.geschrieben, vorher)
	}
}

// **Ein Angebot, das nicht prüft, wird nicht eröffnet — und schreibt nichts.**
func TestEinUngueltigesAngebotSchreibtNichts(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	kaputt := einHandelsangebot()
	kaputt.Positionen = nil

	if _, fehler := leitung.Eroeffnen(
		context.Background(), kaputt); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}

// **Ein Vorschlag geht durch, der zweite gegen denselben Stand nicht.**
func TestDerZweiteVorschlagBekommtEinenKonfliktUndSchreibtNichts(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	ctx := context.Background()

	erste := angebotenerHandel(t, leitung)

	erster := erste.Fortgeschrieben(mitMenge(erste, 3))
	zweiter := erste.Fortgeschrieben(mitMenge(erste, 4))

	if _, fehler := leitung.Vorschlagen(ctx, erster, "bert"); fehler != nil {
		t.Fatalf("der erste Vorschlag: %v", fehler)
	}

	vorher := ablage.geschrieben

	if _, fehler := leitung.Vorschlagen(
		ctx, zweiter, "bert"); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der zweite Vorschlag ergibt %v", fehler)
	}

	if ablage.geschrieben != vorher {
		t.Errorf("der Konflikt hat geschrieben: %d statt %d",
			ablage.geschrieben, vorher)
	}
}

// **Derselbe Vorschlag zweimal ist einmal.**
func TestDerselbeVorschlagZweimalIstEinmal(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	ctx := context.Background()

	erste := angebotenerHandel(t, leitung)

	vorschlag := erste.Fortgeschrieben(mitMenge(erste, 3))

	einmal, fehler := leitung.Vorschlagen(ctx, vorschlag, "bert")
	if fehler != nil {
		t.Fatalf("der erste Ruf: %v", fehler)
	}

	vorher := ablage.geschrieben

	zweimal, fehler := leitung.Vorschlagen(ctx, vorschlag, "bert")
	if fehler != nil {
		t.Fatalf("der zweite Ruf: %v", fehler)
	}

	if zweimal.Abdruck() != einmal.Abdruck() {
		t.Error("der zweite Ruf gibt einen anderen Stand")
	}

	if ablage.geschrieben != vorher {
		t.Errorf("die Wiederholung hat geschrieben: %d statt %d",
			ablage.geschrieben, vorher)
	}
}

// **Ein Schritt geht über die Leitung und steht danach im Bestand.**
func TestEinSchrittWirdGeschrieben(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	ctx := context.Background()

	if _, fehler := leitung.Eroeffnen(ctx, einHandelsangebot()); fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	danach, fehler := leitung.Schreiten(
		ctx, "h1", domain.SchrittAnbieten, "anna")
	if fehler != nil {
		t.Fatalf("Schreiten: %v", fehler)
	}

	if danach.Zustand != domain.HandelAngeboten {
		t.Errorf("der Zustand ist %q", danach.Zustand)
	}

	if ablage.staende["h1"].Zustand != domain.HandelAngeboten {
		t.Errorf("im Bestand steht %q", ablage.staende["h1"].Zustand)
	}
}

// **Ein Schritt, den jemand nicht darf, schreibt nichts.**
func TestEinUnerlaubterSchrittSchreibtNichts(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	ctx := context.Background()

	if _, fehler := leitung.Eroeffnen(ctx, einHandelsangebot()); fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	vorher := ablage.geschrieben

	if _, fehler := leitung.Schreiten(
		ctx, "h1", domain.SchrittAnbieten, "bert",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != vorher {
		t.Errorf("die Ablehnung hat geschrieben: %d statt %d",
			ablage.geschrieben, vorher)
	}
}

// mitMenge gibt die Positionen eines Angebots mit einer anderen Menge.
func mitMenge(
	angebot domain.Handelsangebot, menge int,
) []domain.Handelsposition {
	anders := angebot.Positionen[0]
	anders.Menge = menge

	return []domain.Handelsposition{anders}
}
