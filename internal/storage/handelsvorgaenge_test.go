package storage_test

// **Handelsvorgänge im Bestand** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was hier gemessen wird
//
// **Dass die Kette liegen bleibt.** Eine neue Revision **ersetzt** die vorige
// nicht: sie kommt daneben. Wer später fragt, worauf man sich geeinigt hatte,
// bekommt eine Antwort — und nicht den letzten Stand mit der Behauptung, es
// sei immer so gewesen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einVorgangsangebot gibt ein Angebot mit den genannten Abweichungen.
func einVorgangsangebot(
	aendere ...func(*domain.Handelsangebot),
) domain.Handelsangebot {
	angebot := domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         "h1",
		Revision:        1,
		Art:             domain.HandelTausch,
		Initiator:       "anna",
		Parteien:        []string{"anna", "bert"},
		Zustand:         domain.HandelEntwurf,
		Positionen: []domain.Handelsposition{{
			Von: "anna", Printing: "abc-123", DruckID: 7, Menge: 2,
			Guete: "NM", Sprache: "de", Finish: "nonfoil",
			Abbild: "abdruckderzeile",
		}},
	}

	for _, aenderung := range aendere {
		aenderung(&angebot)
	}

	return angebot
}

// **Ein Angebot kommt zurück, wie es hineinging.**
func TestEinAngebotKommtZurueckWieEsHineinging(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueHandelsablage(frischeDatenbank(t))
	ctx := probekontext(t)

	hinein := einVorgangsangebot()

	if fehler := ablage.Schreiben(ctx, hinein); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	zurueck, fehler := ablage.Vorgang(ctx, "h1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Über den Abdruck und nicht Feld für Feld**: er bindet jedes
	// fachliche Feld, und ein Vergleich darüber sagt genau das.
	if zurueck.Abdruck() != hinein.Abdruck() {
		t.Errorf("zurück kam %+v", zurueck)
	}

	if zurueck.Zustand != hinein.Zustand {
		t.Errorf("der Zustand ist %q", zurueck.Zustand)
	}
}

// **Einen Vorgang, den es nicht gibt, gibt es nicht.**
func TestEinUnbekannterVorgangIstNichtDa(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueHandelsablage(frischeDatenbank(t))

	if _, fehler := ablage.Vorgang(
		probekontext(t), "gibtsnicht"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Die letzte Revision ist die, die gilt — und die vorige bleibt liegen.**
func TestDieLetzteRevisionGiltUndDieVorigeBleibt(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueHandelsablage(frischeDatenbank(t))
	ctx := probekontext(t)

	erste := einVorgangsangebot()
	if fehler := ablage.Schreiben(ctx, erste); fehler != nil {
		t.Fatalf("die erste schreiben: %v", fehler)
	}

	anders := erste.Positionen[0]
	anders.Menge = 3
	zweite := erste.Fortgeschrieben([]domain.Handelsposition{anders})

	if fehler := ablage.Schreiben(ctx, zweite); fehler != nil {
		t.Fatalf("die zweite schreiben: %v", fehler)
	}

	gilt, fehler := ablage.Vorgang(ctx, "h1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gilt.Revision != 2 || gilt.Positionen[0].Menge != 3 {
		t.Errorf("es gilt %+v", gilt)
	}

	kette, fehler := ablage.Revisionen(ctx, "h1")
	if fehler != nil {
		t.Fatalf("die Kette lesen: %v", fehler)
	}

	if len(kette) != 2 {
		t.Fatalf("die Kette hat %d Glieder", len(kette))
	}

	if kette[0].Revision != 1 || kette[1].Revision != 2 {
		t.Errorf("die Kette steht in der Reihenfolge %d, %d",
			kette[0].Revision, kette[1].Revision)
	}

	if kette[1].Eltern != kette[0].Abdruck() {
		t.Error("das zweite Glied hängt nicht am ersten")
	}
}

// **Derselbe Stand zweimal geschrieben bleibt einer.**
//
// **Eine Zustimmung ändert den Zustand, nicht die Revision** — sie wird also
// **in** dieselbe Zeile geschrieben. Wer dafür eine zweite anlegte, hätte
// zwei Revisionen mit derselben Nummer und demselben Abdruck.
func TestEineZustimmungSchreibtInDieselbeZeile(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueHandelsablage(frischeDatenbank(t))
	ctx := probekontext(t)

	angebot := einVorgangsangebot(func(a *domain.Handelsangebot) {
		a.Zustand = domain.HandelAngeboten
	})

	if fehler := ablage.Schreiben(ctx, angebot); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	zugestimmt, fehler := angebot.Schreite(domain.SchrittZustimmen, "bert")
	if fehler != nil {
		t.Fatalf("zustimmen: %v", fehler)
	}

	if klage := ablage.Schreiben(ctx, zugestimmt); klage != nil {
		t.Fatalf("noch einmal schreiben: %v", klage)
	}

	kette, fehler := ablage.Revisionen(ctx, "h1")
	if fehler != nil {
		t.Fatalf("die Kette lesen: %v", fehler)
	}

	if len(kette) != 1 {
		t.Fatalf("die Kette hat %d Glieder", len(kette))
	}

	if len(kette[0].Zustimmungen) != 1 {
		t.Errorf("es stehen %d Zustimmungen da", len(kette[0].Zustimmungen))
	}
}
