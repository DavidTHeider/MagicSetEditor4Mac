// **Die Exportvorlagen am willigen Doppel** — in eigener Datei, weil
// `vorrichtungbestand_test.go` mit U5 (#1416) über die Zeilengrenze ging
// (§2). Der Schnitt folgt der Frage: dort der Bestand, hier die Vorlagen.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Vorlagen(context.Context, int64) ([]rest.Vorlagenzeile, []string, error) {
	return []rest.Vorlagenzeile{{
		Paket: "probe-liste.mse-export-template", Name: "Liste", Dateityp: "*.txt|*.txt",
		Optionen: []rest.Vorlagenoption{{Name: "trenner", Art: "text", Wert: ", "}},
	}}, []string{}, nil
}

func (williger) Ausfuehren(
	context.Context, int64, string, string, map[string]string,
) ([]string, error) {
	return []string{}, nil
}

func (williger) VorlagenoptionSetzen(context.Context, string, string, string) error {
	return nil
}
