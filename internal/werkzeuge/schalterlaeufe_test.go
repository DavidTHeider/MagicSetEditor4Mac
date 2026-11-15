package werkzeuge

// **Die Läufe in SCHALTER.md sind die der Domäne** — OPS-10b
// (#1978).
//
// # Was er hält
//
// - **Jeder Lauf steht in der Tabelle**, in der Ordnung, in der sie enden,
//   mit seiner Beenden-Bindung.
// - **Die Spalte „endet mit“ ist, was die Tafel ergibt** — ein Schalter, der
//   dort fehlt oder zu viel steht, macht ihn rot.
//
// # Was er nicht hält
//
// **Die Spalte „was er hält“** — sie ist für Menschen.

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieLaeufeInSchalterMdSindDieDerDomaene(t *testing.T) {
	t.Parallel()

	roh, fehler := os.ReadFile(schalterkatalog)
	if fehler != nil {
		t.Fatal(fehler)
	}

	// endetMit sammelt je Lauf die Schalter, die ihn allein beenden.
	endetMit := map[domain.Lauf][]string{}

	for _, schalter := range domain.Schalterverzeichnis() {
		stand := domain.SchalterstandAus(map[string]string{
			domain.SchalterEinstellung(schalter.Kennung): domain.SchalterAus,
		})

		for _, lauf := range domain.EingebauteWegetafel().GesperrteLaeufe(stand) {
			endetMit[lauf] = append(endetMit[lauf], "`"+schalter.Kennung+"`")
		}
	}

	var gelesen []domain.Lauf

	beenden := map[string]domain.Lauf{}
	for _, lauf := range domain.Laeufe() {
		beenden["`"+lauf.Beendenweg()+"`"] = lauf
	}

	for _, zeile := range strings.Split(string(roh), "\n") {
		spalten := strings.Split(zeile, "|")
		if len(spalten) != 6 {
			continue
		}

		lauf, gefunden := beenden[strings.TrimSpace(spalten[2])]
		if !gefunden {
			continue
		}

		gelesen = append(gelesen, lauf)

		schalter := strings.Split(strings.TrimSpace(spalten[3]), ", ")
		slices.Sort(schalter)

		erwartet := slices.Sorted(slices.Values(endetMit[lauf]))
		if !slices.Equal(schalter, erwartet) {
			t.Errorf("%s endet laut SCHALTER.md mit %v, laut Tafel mit %v", lauf, schalter, erwartet)
		}
	}

	if !slices.Equal(gelesen, domain.Laeufe()) {
		t.Errorf("SCHALTER.md führt %v, die Domäne %v", gelesen, domain.Laeufe())
	}
}
