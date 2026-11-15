package domain_test

// **Was an einem Schalter hängt und läuft, endet mit ihm** — OPS-10b
// (#1978).

import (
	"regexp"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Je Schalter die Läufe, deren Beenden er sperrt** — und nur die.
func TestEinLaufEndetMitDemSchalterSeinesBeendens(t *testing.T) {
	t.Parallel()

	erwartet := map[string][]domain.Lauf{
		"netz": {
			domain.LaufPartiebeitritt, domain.LaufPartiefreigabe,
			domain.LaufSichtbarkeit, domain.LaufHorchen, domain.LaufFernhorchen,
		},
		"spielen":        {domain.LaufPartiebeitritt, domain.LaufPartiefreigabe},
		"ki":             {domain.LaufIndexaufbau, domain.LaufBildauftraege},
		"dienst":         {domain.LaufREST},
		"werkzeugserver": {domain.LaufMCP},
	}

	tafel := domain.EingebauteWegetafel()

	for _, schalter := range domain.Schalterverzeichnis() {
		stand := domain.SchalterstandAus(map[string]string{
			domain.SchalterEinstellung(schalter.Kennung): domain.SchalterAus,
		})

		if laeufe := tafel.GesperrteLaeufe(stand); !slices.Equal(laeufe, erwartet[schalter.Kennung]) {
			t.Errorf("%s aus: %v, erwartet %v", schalter.Kennung, laeufe, erwartet[schalter.Kennung])
		}
	}

	if laeufe := tafel.GesperrteLaeufe(domain.SchalterstandAus(nil)); len(laeufe) != 0 {
		t.Errorf("alles an: %v", laeufe)
	}
}

// **Gesperrt ist ein Beenden erst, wenn jede Fähigkeit aus ist** — dieselbe
// Regel wie für jeden anderen Weg, und nicht eine zweite daneben.
func TestEinBeendenMitZweiFaehigkeitenBrauchtBeideAus(t *testing.T) {
	t.Parallel()

	tafel := domain.Wegetafel{domain.FlaecheFenster: {
		domain.LaufHorchen.Beendenweg(): {{"netz"}, {"spielen"}},
	}}

	aus := func(kennungen ...string) domain.Schalterstand {
		werte := map[string]string{}
		for _, kennung := range kennungen {
			werte[domain.SchalterEinstellung(kennung)] = domain.SchalterAus
		}

		return domain.SchalterstandAus(werte)
	}

	if laeufe := tafel.GesperrteLaeufe(aus("netz")); len(laeufe) != 0 {
		t.Errorf("eine von zwei: %v", laeufe)
	}

	if laeufe := tafel.GesperrteLaeufe(aus("netz", "spielen")); !slices.Equal(laeufe, []domain.Lauf{domain.LaufHorchen}) {
		t.Errorf("beide: %v", laeufe)
	}
}

// **Jedes gesperrte Beenden gehört einem Lauf — oder steht hier mit Grund.**
//
// Der Wächter über die Klasse aus #1978: ein neuer Lauf, dessen Beenden an
// einem Schalter hängt, wäre nach dem Abschalten nicht mehr anzuhalten.
// **Was er nicht hält**: einen Lauf, dessen Beenden anders heißt. Die Namen
// der Bindungen enden heute alle auf *Beenden* oder *Abbrechen*.
func TestJedesGesperrteBeendenGehoertEinemLauf(t *testing.T) {
	t.Parallel()

	ausgenommen := map[string]string{
		// Eine Partie ist ein Datensatz und kein Lauf: wer `spielen` wieder
		// einschaltet, findet sie, wie er sie verließ (Vermerk 0158).
		"PartieBeenden": "Datensatz",
		// Der gemeinsame Weg hinter `RESTBeenden` und `MCPBeenden`.
		"dienstBeenden": "unexportiert",
	}

	beenden := map[string]domain.Lauf{}
	for _, lauf := range domain.Laeufe() {
		beenden[lauf.Beendenweg()] = lauf
	}

	endet := regexp.MustCompile(`(Beenden|Abbrechen)$`)

	for weg := range domain.EingebauteWegetafel()[domain.FlaecheFenster] {
		_, gefuehrt := beenden[weg]
		if _, frei := ausgenommen[weg]; endet.MatchString(weg) && !gefuehrt && !frei {
			t.Errorf("%s hängt an einem Schalter und beendet keinen Lauf", weg)
		}
	}

	// **Und umgekehrt**: ein Beenden, das die Tafel nicht führt, sperrte
	// nie — der Lauf liefe dann still weiter.
	for weg, lauf := range beenden {
		if len(domain.EingebauteWegetafel()[domain.FlaecheFenster][weg]) == 0 {
			t.Errorf("%s: %s steht nicht in der Tafel", lauf, weg)
		}
	}
}

// **Ein unbekannter Lauf hat kein Beenden** — und sperrt damit nichts.
func TestEinUnbekannterLaufHatKeinBeenden(t *testing.T) {
	t.Parallel()

	if weg := domain.Lauf("unbekannt").Beendenweg(); weg != "" {
		t.Errorf("%q", weg)
	}
}
