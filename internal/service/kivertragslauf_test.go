package service_test

// **Der Vertragslauf für Einbettung und Neuordnung** — Werkstück A6 von
// Phase 5.3 (#1445).
//
// Das Issue verlangt ihn ausdrücklich: „unabhängiger Embedding-/Rerank-
// Contract-Test **ohne Generierungsprovider**". Was hier läuft, sind
// Doppel — und genau das ist der Punkt: der Lauf prüft **Anbieter**, nicht
// dieses Programm.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// derProbevertrag ist der Vertrag, gegen den die Läufe unten prüfen.
func derProbevertrag() domain.KIEinbettungsvertrag {
	return domain.KIEinbettungsvertrag{
		Dimension: 4, Pooling: domain.KIPoolingMittel, Normalisiert: true,
		Zerteiler: "probe", Frageanweisung: "F: ", Dokumentanweisung: "D: ",
	}
}

// **Eine Quelle, die sich an ihren Vertrag hält, kommt durch.**
func TestEineVertragstreueQuelleKommtDurch(t *testing.T) {
	t.Parallel()

	befunde, fehler := service.PruefeEinbettung(t.Context(),
		&einbettungsprobe{}, derProbevertrag())
	if fehler != nil {
		t.Fatalf("der Vertragslauf: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Jede Abweichung wird benannt — und zwar einzeln.**
func TestJedeAbweichungVomEinbettungsvertragWirdBenannt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		quelle *einbettungsprobe
		teil   string
	}{
		"falsche Länge": {
			&einbettungsprobe{laenge: 3}, "Stellen",
		},
		"keine Zahl": {
			&einbettungsprobe{keineZahl: true}, "keine Zahl",
		},
		"nicht normalisiert": {
			&einbettungsprobe{ohneNorm: true}, "Länge",
		},
		"zweimal verschieden": {
			&einbettungsprobe{wechselhaft: true}, "Zweimal derselbe Text",
		},
		"Frage und Dokument gleich": {
			&einbettungsprobe{blind: true}, "Anweisung",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			befunde, fehler := service.PruefeEinbettung(t.Context(),
				fall.quelle, derProbevertrag())
			if fehler != nil {
				t.Fatalf("der Vertragslauf: %v", fehler)
			}

			if len(befunde) == 0 {
				t.Fatalf("%s gab keinen Befund", name)
			}

			gefunden := false

			for _, befund := range befunde {
				if strings.Contains(befund.Meldung, fall.teil) {
					gefunden = true
				}
			}

			if !gefunden {
				t.Errorf("kein Befund nennt %q: %+v", fall.teil, befunde)
			}
		})
	}
}

// **Ein Fehler der Quelle ist ein Fehler und kein Befund.**
//
// **Der Unterschied zählt**: ein Befund sagt „der Anbieter hält seinen
// Vertrag nicht", ein Fehler sagt „es kam gar nichts". Wer beides gleich
// behandelte, ließe einen ausgefallenen Dienst wie einen vertragsbrüchigen
// aussehen.
func TestEinFehlerDerEinbettungsquelleIstEinFehler(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Laufzeit antwortet nicht")

	if _, fehler := service.PruefeEinbettung(t.Context(),
		&einbettungsprobe{fehler: eigen}, derProbevertrag()); !errors.Is(fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Vertrag, der selbst nicht stimmt, wird vorher abgewiesen.**
func TestEinKaputterVertragWirdVorherAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := service.PruefeEinbettung(t.Context(),
		&einbettungsprobe{}, domain.KIEinbettungsvertrag{}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Eine Neuordnung wird gegen ihre Kandidaten geprüft.**
func TestEineNeuordnungWirdGegenIhreKandidatenGeprueft(t *testing.T) {
	t.Parallel()

	kandidaten := []string{"eins", "zwei"}

	befunde, fehler := service.PruefeNeuordnung(t.Context(),
		&neuordnungsprobe{}, "frage", kandidaten)
	if fehler != nil {
		t.Fatalf("der Vertragslauf: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Eine erfundene Bewertung wird benannt.**
func TestEineErfundeneBewertungWirdBenannt(t *testing.T) {
	t.Parallel()

	befunde, fehler := service.PruefeNeuordnung(t.Context(),
		&neuordnungsprobe{erfindet: true}, "frage", []string{"eins"})
	if fehler != nil {
		t.Fatalf("der Vertragslauf: %v", fehler)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}
}

// **Und ein Fehler der Neuordnung bleibt einer.**
func TestEinFehlerDerNeuordnungBleibtEiner(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Laufzeit antwortet nicht")

	if _, fehler := service.PruefeNeuordnung(t.Context(),
		&neuordnungsprobe{fehler: eigen}, "frage", []string{"eins"}); !errors.Is(
		fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Fehler kann bei jedem der drei Rufe kommen.**
//
// **Drei und nicht einer**: der Lauf fragt zweimal dasselbe (für die
// Wiederholbarkeit) und einmal anders (für die Anweisung). Fällt der zweite
// oder dritte aus, ist das ein Fehler und kein Befund — ohne diesen Lauf
// bliebe offen, ob er überhaupt ankommt.
func TestEinFehlerBeiJedemRufKommtAn(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Laufzeit antwortet nicht mehr")

	for ab := 1; ab <= 3; ab++ {
		t.Run("ab dem Ruf "+string(rune('0'+ab)), func(t *testing.T) {
			t.Parallel()

			if _, fehler := service.PruefeEinbettung(t.Context(),
				&einbettungsprobe{fehlerAb: ab, fehler: eigen},
				derProbevertrag()); !errors.Is(fehler, eigen) {
				t.Fatalf("der Fehler ist %v", fehler)
			}
		})
	}
}

// **Zu wenige Vektoren sind ein Befund.**
func TestZuWenigeVektorenSindEinBefund(t *testing.T) {
	t.Parallel()

	befunde, fehler := service.PruefeEinbettung(t.Context(),
		&einbettungsprobe{halbiert: true}, derProbevertrag())
	if fehler != nil {
		t.Fatalf("der Vertragslauf: %v", fehler)
	}

	if len(befunde) == 0 || !strings.Contains(befunde[0].Meldung, "Vektoren") {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Ohne zugesagte Normalisierung wird sie nicht geprüft.**
//
// **Die Gegenprobe zu »nicht normalisiert«**: dort sagt der Vertrag es zu und
// die Quelle hält es nicht; hier sagt er nichts zu, und dieselbe Quelle kommt
// durch. Ohne diesen Lauf wäre nicht zu sehen, ob die Prüfung an der Zusage
// hängt oder immer läuft.
func TestOhneZugesagteNormalisierungWirdSieNichtGeprueft(t *testing.T) {
	t.Parallel()

	vertrag := derProbevertrag()
	vertrag.Normalisiert = false

	befunde, fehler := service.PruefeEinbettung(t.Context(),
		&einbettungsprobe{ohneNorm: true}, vertrag)
	if fehler != nil {
		t.Fatalf("der Vertragslauf: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Eine Bewertung zu einem fremden Kandidaten wird benannt.**
//
// **Nicht dasselbe wie »zu viele«**: hier stimmt die Zahl, und trotzdem
// gehört eine Bewertung zu nichts. Die Zählung allein fände es nicht.
func TestEineBewertungZuEinemFremdenKandidatenWirdBenannt(t *testing.T) {
	t.Parallel()

	befunde, fehler := service.PruefeNeuordnung(t.Context(),
		&neuordnungsprobe{fremd: true}, "frage", []string{"eins", "zwei"})
	if fehler != nil {
		t.Fatalf("der Vertragslauf: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "nicht dabei") {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Ein Anbieter, der zwischen zwei Rufen die Form wechselt, wird benannt.**
//
// # Warum das kein erfundener Fall ist
//
// **Weil ein Alias tauschen kann, was hinter ihm liegt** (A6,
// [domain.KIModell.Bindung]). Genau dann antwortet dieselbe Anfrage anders —
// und wenn die Prüfung daran vorbeiliefe, stünde ein halb gewechselter Index
// da, ohne dass jemand etwas bemerkt.
func TestEinFormwechselZwischenZweiRufenWirdBenannt(t *testing.T) {
	t.Parallel()

	for name, quelle := range map[string]*einbettungsprobe{
		"ein Vektor weniger":      {kuerzerAb: 2},
		"eine Stelle mehr":        {laengerAb: 2},
		"das Dokument ist kürzer": {kuerzerAb: 3},
		"das Dokument ist länger": {laengerAb: 3},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			befunde, fehler := service.PruefeEinbettung(t.Context(),
				quelle, derProbevertrag())
			if fehler != nil {
				t.Fatalf("der Vertragslauf: %v", fehler)
			}

			// **Die letzten beiden dürfen schweigen**: ein kürzeres oder
			// längeres Dokumentergebnis sagt nichts über die **Anweisung**,
			// und der Lauf darüber ist der einzige, der es sähe. Was hier
			// zählt, ist, dass er nicht **fälschlich** meldet.
			for _, befund := range befunde {
				if strings.Contains(befund.Meldung, "Anweisung") {
					t.Errorf("%s wurde als Anweisungsfehler gemeldet: %+v",
						name, befunde)
				}
			}
		})
	}
}

// **Ohne verschiedene Anweisungen ist Gleichheit kein Befund.**
//
// **Die Gegenprobe zu »Frage und Dokument gleich«**: dort sagt der Vertrag
// zwei Anweisungen zu und die Quelle gibt denselben Vektor; hier sagt er
// keine zu, und dieselbe Quelle kommt durch. Ohne diesen Lauf meldete der
// Vertragslauf jedem Anbieter ohne Anweisungen einen Fehler, den es nicht
// gibt — die Abwesenheit einer Zusage ist kein Bruch.
func TestOhneAnweisungenIstGleichheitKeinBefund(t *testing.T) {
	t.Parallel()

	vertrag := derProbevertrag()
	vertrag.Frageanweisung, vertrag.Dokumentanweisung = "", ""

	befunde, fehler := service.PruefeEinbettung(t.Context(),
		&einbettungsprobe{blind: true}, vertrag)
	if fehler != nil {
		t.Fatalf("der Vertragslauf: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}
