package domain_test

// **Was eine Berichtigung verschöbe** — DOM-04.AK-02
// (#1699):
// *„Ergebniskorrekturen zeigen Auswirkungen auf folgende Runden und
// Ranglisten."*
//
// **Gerechnet über dieselbe Rangliste** wie jede Anzeige (Vermerk 0062,
// Entscheidung 1): die Folgen sind der Unterschied zweier Ranglisten, nicht
// eine zweite Rechnung.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiRunden ist ein Turnier aus vier Namen und zwei gespielten Runden.
//
// Runde 1: Anna schlägt Bert, Cleo schlägt Dora. Runde 2 (nach dem Stand
// der ersten gepaart): Anna schlägt Cleo, Bert schlägt Dora.
func zweiRunden() domain.Turnier {
	sieg := &domain.Ergebnis{SpieleA: 2}

	return domain.Turnier{
		ID: 1, Name: "Freitag", Spiel: "magic",
		Teilnehmer: []domain.Teilnehmer{
			{ID: 1, Name: "Anna", Stelle: 1},
			{ID: 2, Name: "Bert", Stelle: 2},
			{ID: 3, Name: "Cleo", Stelle: 3},
			{ID: 4, Name: "Dora", Stelle: 4},
		},
		Runden: []domain.Runde{
			{ID: 10, Nummer: 1, Paarungen: []domain.Paarung{
				{ID: 100, A: 1, B: 2, Ergebnis: sieg}, {ID: 101, A: 3, B: 4, Ergebnis: sieg},
			}},
			{ID: 11, Nummer: 2, Paarungen: []domain.Paarung{
				{ID: 110, A: 1, B: 3, Ergebnis: sieg}, {ID: 111, A: 2, B: 4, Ergebnis: sieg},
			}},
		},
	}
}

func TestEineBerichtigungNenntIhreFolgen(t *testing.T) {
	t.Parallel()

	// **Bert gewinnt Runde 1 doch** — er zieht an Anna vorbei, und Runde 2
	// wurde nach dem alten Stand gepaart.
	turnier := zweiRunden()

	folgen, err := domain.FolgenDerBerichtigung(turnier, 100, &domain.Ergebnis{SpieleB: 2})
	if err != nil {
		t.Fatalf("Folgen: %v", err)
	}

	// **Gerechnet wird an einer Abschrift** — das Turnier des Aufrufers
	// trägt danach noch sein Ergebnis.
	if turnier.Runden[0].Paarungen[0].Ergebnis.SpieleA != 2 {
		t.Error("die Frage hat das Ergebnis im Turnier des Aufrufers geändert")
	}

	if folgen.Runde != 1 || len(folgen.SpaetereRunden) != 1 || folgen.SpaetereRunden[0] != 2 {
		t.Errorf("Runde %d, später %v — erwartet 1 und [2]", folgen.Runde, folgen.SpaetereRunden)
	}

	verschoben := map[string]domain.Platzwechsel{}
	for _, wechsel := range folgen.Verschoben {
		verschoben[wechsel.Name] = wechsel
	}

	anna, bert := verschoben["Anna"], verschoben["Bert"]
	if anna.PunkteVorher != 6 || anna.PunkteNachher != 3 || bert.PunkteVorher != 3 || bert.PunkteNachher != 6 {
		t.Errorf("Punkte: Anna %+v, Bert %+v", anna, bert)
	}

	if bert.PlatzNachher >= bert.PlatzVorher {
		t.Errorf("Bert steigt nicht: %+v", bert)
	}

	// **Nur wer sich ändert, steht da.**
	for _, wechsel := range folgen.Verschoben {
		if wechsel.PlatzVorher == wechsel.PlatzNachher && wechsel.PunkteVorher == wechsel.PunkteNachher {
			t.Errorf("%s steht in den Folgen, ohne dass sich etwas ändert", wechsel.Name)
		}
	}
}

// **Dasselbe Ergebnis noch einmal verschiebt nichts** — und die letzte Runde
// hat keine spätere.
func TestEinUnveraendertesErgebnisHatKeineFolgen(t *testing.T) {
	t.Parallel()

	folgen, err := domain.FolgenDerBerichtigung(zweiRunden(), 111, &domain.Ergebnis{SpieleA: 2})
	if err != nil {
		t.Fatalf("Folgen: %v", err)
	}

	if folgen.Runde != 2 || len(folgen.SpaetereRunden) != 0 || len(folgen.Verschoben) != 0 {
		t.Errorf("Folgen: %+v, erwartet keine", folgen)
	}
}

// **Auch das Zurücknehmen hat Folgen** — ein Ergebnis weniger ist ein
// anderer Stand.
func TestAuchDasZuruecknehmenHatFolgen(t *testing.T) {
	t.Parallel()

	folgen, err := domain.FolgenDerBerichtigung(zweiRunden(), 110, nil)
	if err != nil {
		t.Fatalf("Folgen: %v", err)
	}

	if len(folgen.Verschoben) == 0 {
		t.Error("das Zurücknehmen von Annas zweitem Sieg verschiebt nichts")
	}
}

// **Eine Paarung, die es nicht gibt, ist nicht gefunden.**
func TestFolgenEinerFremdenPaarung(t *testing.T) {
	t.Parallel()

	if _, err := domain.FolgenDerBerichtigung(zweiRunden(), 999, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}
