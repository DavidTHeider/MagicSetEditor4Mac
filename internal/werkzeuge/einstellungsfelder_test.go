// **Was das Fenster über eine Einstellung erfährt, erfahren auch die
// anderen drei Flächen** — UX-04.AK-02
// (#1676),
// §16.
//
// # Der Anlass
//
// **Der Zeitpunkt der Wirksamkeit kam am 2026-09-21 dazu**, und er kam an
// vier Stellen dazu: `app.EinstellungDTO`, `rest.Einstellung`,
// `mcp.Einstellungszeile` und die zwei Umsetzer dazwischen. Wer eine
// vergisst, merkt es nicht: jede Fläche sieht für sich vollständig aus, und
// die Lücke liegt **zwischen** ihnen (§10, Dimension „Nähte").
//
// Dieselbe Gestalt wie `sitzfelder_test.go` — und derselbe Grund: dort
// blieben fünf Felder in der Domäne liegen, weil niemand die Naht gezählt
// hat.
//
// # Was er prüft
//
// Jedes exportierte Feld von `app.EinstellungDTO` steht in
// `rest.Einstellung` **und** in `mcp.Einstellungszeile`.
//
// # Was er nicht prüft
//
// **Dass das Feld gefüllt wird.** Ein Feld, das dasteht und leer bleibt,
// geht hier durch — dafür stehen die Läufe an den Umsetzern. Und er prüft
// keine Bedeutung: ob `Wirkung` überall dasselbe heißt, sagt kein
// Namensvergleich.
//
// **Und die Kommandozeile prüft er gar nicht.** Sie reicht `EinstellungDTO`
// unverändert als JSON durch (`cmd/mse/bestandsbefehle.go`) — es gibt dort
// keine zweite Gestalt, die auseinanderlaufen könnte. Ihre **Tabellenform**
// zeigt ohnehin nur drei Spalten; das ist eine Entscheidung über die
// Lesbarkeit und keine über die Fläche.
package werkzeuge_test

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestJedesEinstellungsfeldErreichtJedeFlaeche(t *testing.T) {
	t.Parallel()

	fenster := felderVon(t,
		filepath.Join("..", "app", "einstellungen.go"), "EinstellungDTO")

	flaechen := map[string][]string{
		"rest.Einstellung": felderVon(t,
			filepath.Join("..", "rest", "setnaehte.go"), "Einstellung"),
		"mcp.Einstellungszeile": felderVon(t,
			filepath.Join("..", "mcp", "bestandswerkzeuge.go"),
			"Einstellungszeile"),
	}

	for name, felder := range flaechen {
		for _, feld := range fenster {
			if !slices.Contains(felder, feld) {
				t.Errorf("%s fehlt das Feld %q aus app.EinstellungDTO",
					name, feld)
			}
		}
	}
}
