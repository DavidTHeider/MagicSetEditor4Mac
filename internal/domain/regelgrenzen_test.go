package domain_test

// **Jede Zusage nennt ihre härteste Grenze** — der Wächter dazu.
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was ein Profil überhaupt ist**; hier steht, **woran seine
// Zusagen scheitern** — und die Liste wächst mit jedem Werkstück der
// Regelmaschine. `regelprofil_test.go` stand mit ihr bei 401 Zeilen (§2).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jede Zusage nennt die Grenze, an der sie heute scheitert.**
//
// # Warum diese Nummern und nicht irgendwelche
//
// **Weil die Abdeckungssicht sie gezählt hat** (`docs/REGELABDECKUNG.md`).
// Ein Bereich auf `automatisch` ohne Grenze liest sich wie „nach den
// Regeln", und eine Zusage, die zur Hälfte hält, ist schlimmer als keine
// (§10).
//
// # Drei Erwartungen sind mit M18 gefallen
//
// **`zonen`, `marken` und `anhaengsel` nannten alle drei CR 400.7** — den
// Zonenwechsel, der kein neues Objekt macht. Seit `M18`
// (#1619)
// macht er eines, und die Grenze ist keine mehr. **Bei `zonen` bleibt die
// Nummer stehen**, weil dort jetzt das **Gegenteil** als Grenze steht: die
// Kennung überlebt, und die zwölf Ausnahmen aus CR 400.7a–m fehlen. Bei den
// anderen beiden wäre eine Nummer ohne Aussage ein Wächter, der auf eine
// Zeichenkette prüft statt auf eine Zusage.
func TestJedeZusageNenntIhreHarteGrenze(t *testing.T) {
	t.Parallel()

	matrix := domain.Grundprofil("magic").Matrix

	erwartet := map[domain.Regelbereich][]string{
		domain.BereichZonen: {"400.7"},

		// **Die zwei aus G9** (#1592): jede nennt, was sie **nicht** tut.
		//
		// **CR 500.5 stand hier und steht nicht mehr da**: `M3`
		// (#1605)
		// leert den Manavorrat wirklich, und eine Grenze, die nicht mehr
		// gilt, ist eine falsche Zusage (§10). Was von CR 500.5 **offen**
		// bleibt — Effekte, die mit dem Abschnitt enden —, steht als
		// „im Aufräumschritt endet kein Effekt" daneben und gehört zu
		// CR 514.2.
		// **CR 514.2 stand hier und steht es nicht mehr**: `M10`
		// (#1612)
		// baut die zweite Hälfte — im Aufräumschritt endet jetzt auch, was
		// „bis zum Ende des Zuges" gilt. Was von der Zugstruktur offen
		// bleibt, sind die fünf zugbasierten Handlungen aus CR 703.4.
		domain.BereichZugstruktur: {"502.1", "703.4"},
		domain.BereichPrioritaet:  {"608", "514.3a", "732"},

		// **Und der neue aus M3**: er sagt, dass niemand bezahlt.
		domain.BereichMana: {"601.2f", "305.6", "107.4e"},

		// **Und der aus M4**: er sagt, dass niemand ausführt, was auf einem
		// Zauber steht — die größte Grenze, die dieses Programm hat.
		domain.BereichStapel: {"608.2c", "601.2b", "115.1", "116.2a"},

		// **Und der aus M6**: er sagt, welche Schlüsselwörter er liest und
		// welche nicht.
		//
		// **CR 702.7 stand hier und steht es nicht mehr**: `M14`
		// (#1616)
		// zählt nicht mehr die drei gelesenen auf, sondern nennt **die
		// Zahl** und die drei Gestalten, die auch mit einer längeren Liste
		// nicht gehen — Schnellzauber, Farblosigkeit und Schutz.
		domain.BereichCombat: {
			"508.1c", "702", "702.8", "702.16", "509.2", "506.4",
		},

		// **Und der aus M9** (#1611): er sagt, dass es genau **vier**
		// ausgelöste Fähigkeiten gibt und dass jede weitere ein
		// Kartentext ist — und er nennt die drei Gestalten, die fehlen.
		domain.BereichWechselwirkungen: {
			"608.2c", "603.1", "603.8", "603.7", "603.12", "603.3b", "707.2",
		},
	}

	for bereich, nummern := range erwartet {
		eintrag, da := matrix.Eintrag(bereich)
		if !da {
			t.Fatalf("%q steht nicht in der Matrix", bereich)
		}

		alle := strings.Join(eintrag.Grenzen, " | ")

		for _, nummer := range nummern {
			if !strings.Contains(alle, nummer) {
				t.Errorf("%q nennt CR %s in keiner Grenze — dort steht nur: %s",
					bereich, nummer, alle)
			}
		}
	}
}
