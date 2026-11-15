// **Die Marke über eine Rücksetzungsvorschau** — Befund 12 des Zweitblicks
// über Phase 5.2 (#1435).
package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Dieselbe Vorschau gibt dieselbe Marke — eine andere eine andere.**
func TestDieRuecksetzungsmarkeHaengtAmUmfang(t *testing.T) {
	t.Parallel()

	drei := []domain.Betroffenes{
		{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins"}},
		{Was: domain.BetroffenKarten, Anzahl: 107},
	}

	marke := domain.Ruecksetzungsmarke(domain.RuecksetzungProfil, drei)

	if marke == "" || len(marke) != domain.MarkeLaenge {
		t.Fatalf("die Marke ist %q", marke)
	}

	// **Die Reihenfolge zählt nicht** — Go durchläuft nichts zweimal gleich,
	// und eine Marke, die sich dabei änderte, wäre keine.
	gedreht := []domain.Betroffenes{drei[1], drei[0]}
	if domain.Ruecksetzungsmarke(domain.RuecksetzungProfil, gedreht) != marke {
		t.Error("die Reihenfolge änderte die Marke")
	}

	// **Die Reihenfolge der Namen zählt ebenso wenig** — sie kommt aus dem
	// Bestand, und der ordnet nicht zu.
	andersherum := []domain.Betroffenes{
		{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Zwei", "Eins"}},
		{Was: domain.BetroffenKarten, Anzahl: 107},
	}
	mitZwei := []domain.Betroffenes{
		{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins", "Zwei"}},
		{Was: domain.BetroffenKarten, Anzahl: 107},
	}

	if domain.Ruecksetzungsmarke(domain.RuecksetzungProfil, andersherum) !=
		domain.Ruecksetzungsmarke(domain.RuecksetzungProfil, mitZwei) {
		t.Error("die Reihenfolge der Namen änderte die Marke")
	}

	// **Und die Namen nicht auch noch sortieren** — der Aufruf oben darf die
	// Eingabe des Aufrufers nicht umstellen.
	if andersherum[0].Namen[0] != "Zwei" {
		t.Errorf("die Eingabe wurde umgestellt: %v", andersherum[0].Namen)
	}

	for name, anders := range map[string]struct {
		art       domain.Ruecksetzungsart
		betroffen []domain.Betroffenes
	}{
		"eine andere Art": {domain.RuecksetzungNutzer, drei},
		"eine Zahl mehr": {domain.RuecksetzungProfil, []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 4},
			{Was: domain.BetroffenKarten, Anzahl: 107},
		}},
		"eine Gruppe mehr": {domain.RuecksetzungProfil, append(
			[]domain.Betroffenes{{Was: domain.BetroffenDecks, Anzahl: 1}}, drei...)},
		"gar nichts": {domain.RuecksetzungProfil, nil},
		// **Befund 9 des Nachlaufs**: dieselbe Zahl, ein anderes Set. Die
		// Vorschau nannte »Eins«, weggenommen würde »Anderes« — und die Marke
		// sagte Ja, weil sie nur bis zur Zahl las.
		"dieselbe Zahl, andere Namen": {domain.RuecksetzungProfil, []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Anderes"}},
			{Was: domain.BetroffenKarten, Anzahl: 107},
		}},
		"ein Name mehr": {domain.RuecksetzungProfil, []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins", "Zwei"}},
			{Was: domain.BetroffenKarten, Anzahl: 107},
		}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if domain.Ruecksetzungsmarke(anders.art, anders.betroffen) == marke {
				t.Error("dieselbe Marke für etwas anderes")
			}
		})
	}
}
