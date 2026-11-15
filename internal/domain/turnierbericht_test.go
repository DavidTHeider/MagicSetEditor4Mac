package domain_test

// **Der Abschlussbericht eines Turniers** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// berichtworte sind Wörter, die sich im Bericht wiederfinden lassen.
func berichtworte() domain.Berichttexte {
	return domain.Berichttexte{
		Spiel: "SPIEL", Format: "FORMAT", Stand: "STAND", Abgeschlossen: "ZU AM", Laufend: "LÄUFT",
		Teilnehmende: "LEUTE", Ausgestiegen: "RAUS", Runden: "RUNDEN",
		Rangliste: "RANG", Platz: "PLATZ", Name: "NAME", Punkte: "PUNKTE", Bilanz: "BILANZ",
		Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
		Runde: "RUNDE", Tisch: "TISCH", SeiteA: "EINS", SeiteB: "ZWEI", Ergebnis: "STAND DER DINGE",
		Bye: "FREILOS", Offen: "OFFEN",
	}
}

func TestDerBerichtNenntKopfRanglisteUndRunden(t *testing.T) {
	t.Parallel()

	turnier := gespieltesTurnier()
	turnier.Einzelspiel = false
	turnier.Teilnehmer[2].Name = "Cleo | die Dritte"

	bericht := domain.Turnierbericht(turnier, "Pauper", berichtworte())

	for _, zeile := range []string{
		"# Freitag\n",
		"- SPIEL: magic\n",
		"- FORMAT: Pauper\n",
		"- STAND: ZU AM 2026-09-20\n",
		"- LEUTE: 3 (RAUS: 1)\n",
		"- RUNDEN: 2\n",
		"## RANG\n",
		"| PLATZ | NAME | PUNKTE | BILANZ | OMW | GW | OGW |\n",
		// **Wie im Fenster**: Siege samt Byes, Quoten auf vier Stellen.
		"| 1 | Cleo \\| die Dritte | 6 | 2–0–0 | ",
		"| 3 | Bert (RAUS) | 0 | 0–1–0 | ",
		"## RUNDE 1\n",
		"| TISCH | EINS | ZWEI | STAND DER DINGE |\n",
		"| 1 | Anna | Bert (RAUS) | 1–0 |\n",
		"| 2 | Cleo \\| die Dritte | FREILOS | — |\n",
		"| 1 | Anna | Cleo \\| die Dritte | 0–1 |\n",
	} {
		if !strings.Contains(bericht, zeile) {
			t.Errorf("es fehlt %q in:\n%s", zeile, bericht)
		}
	}
}

// **Ein laufendes Einzelspielturnier ohne Format** — ohne die zwei Spalten,
// die dann Kopien der Matchzahlen wären (MTR §3.1, Fassung 2026-02-27; Vermerk 0062/5), mit
// offenem Tisch und Unentschieden.
func TestDerBerichtEinesLaufendenEinzelspielturniers(t *testing.T) {
	t.Parallel()

	turnier := domain.Turnier{
		Name: "Küche", Spiel: "magic", Einzelspiel: true,
		Teilnehmer: []domain.Teilnehmer{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}, {ID: 3, Name: "C"}, {ID: 4, Name: "D"}},
		Runden: []domain.Runde{{Nummer: 1, Paarungen: []domain.Paarung{
			{A: 1, B: 2, Ergebnis: &domain.Ergebnis{Unentschieden: 1}},
			{A: 3, B: 4},
		}}},
	}

	bericht := domain.Turnierbericht(turnier, "", berichtworte())

	for _, zeile := range []string{
		"- STAND: LÄUFT\n", "- LEUTE: 4\n",
		"| PLATZ | NAME | PUNKTE | BILANZ | OMW |\n",
		"| 1 | A | 1 | 0–0–1 | 0.3333 |\n",
		"| 1 | A | B | 0–0–1 |\n",
		"| 2 | C | D | OFFEN |\n",
	} {
		if !strings.Contains(bericht, zeile) {
			t.Errorf("es fehlt %q in:\n%s", zeile, bericht)
		}
	}

	if strings.Contains(bericht, "FORMAT") || strings.Contains(bericht, "| GW") {
		t.Errorf("Format oder Spielquote stehen da:\n%s", bericht)
	}
}
