package service_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Datei, die sich nicht lesen lässt, ist eine Ablehnung mit Satz** —
// und nicht „Set öffnen: …" vor einem Satz, der schon sagt, was scheiterte
// (FMT-06.AK-04, #1695). Der ganze Weg bis ins Fenster steht in
// `internal/app/fehlerdateien_test.go`; hier die Naht, an der er entschieden
// wird.
func TestEineUnlesbareDateiIstEineAblehnungMitSatz(t *testing.T) {
	speicher := &speicherDouble{}

	_, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "fehlerdateien", "kein-utf8.mse-set"))

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war eine Validierung", err)
	}

	if satz := domain.SatzVon(err); !strings.HasPrefix(satz, "kein-utf8.mse-set lässt sich nicht lesen") {
		t.Errorf("der Satz beginnt nicht mit der Datei: %q", satz)
	}

	if speicher.set != nil {
		t.Error("aus einer unlesbaren Datei wurde trotzdem ein Set abgelegt")
	}
}

// **Eine Datei, die es nicht gibt, gibt es nicht** — und das ist kein
// unbekannter Fehler (FMT-03.AK-01, #1692). Für ein Setpaket sagte es
// `msepaket` schon; eine MTG-JSON-Datei ging als „MTG JSON öffnen: …"
// hinaus, und der Dienst machte daraus `500 unbekannt` — ein Aufrufer konnte
// einen Tippfehler im Pfad nicht von einem kaputten Programm unterscheiden.
// **Und eine, die sich nicht lesen lässt, sagt es** — an allen drei Wegen,
// ohne den Pfad, der ins Protokoll gehört (§8).
func TestEineFehlendeOderGesperrteDateiSagtEs(t *testing.T) {
	gesperrt := filepath.Join(t.TempDir(), "gesperrt.mse-set")
	lege(t, filepath.Join(gesperrt, "set"), "mse version: 2.0.0\ngame: probe\n")
	unlesbar(t, gesperrt)

	for _, fall := range []struct {
		name   string
		pfad   string
		klasse error
		satz   string
	}{
		{"fehlt", filepath.Join(t.TempDir(), "gibtsnicht.mse-set"), domain.ErrNotFound, ""},
		{"gesperrt", gesperrt, domain.ErrValidation, "gesperrt.mse-set lässt sich nicht lesen — keine Berechtigung"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			for _, weg := range []struct {
				name string
				ruf  func() error
			}{
				{"einfahren", func() error {
					_, err := service.NeuerImporteur(&speicherDouble{}).Set(t.Context(), fall.pfad)

					return err
				}},
				{"vorschau", func() error {
					_, err := service.NeuerImporteur(&speicherDouble{}).Vorschau(t.Context(), fall.pfad)

					return err
				}},
				{"mtgjson", func() error {
					_, err := service.NeuerImporteur(&speicherDouble{}).MTGJSON(t.Context(), fall.pfad, "magic")

					return err
				}},
			} {
				err := weg.ruf()
				if !errors.Is(err, fall.klasse) || !strings.HasPrefix(domain.SatzVon(err), fall.satz) {
					t.Errorf("%s: %v — Satz %q", weg.name, err, domain.SatzVon(err))
				}
			}
		})
	}
}

// **Kaputtes MTG JSON ist eine Ablehnung mit Stelle** — derselbe Grundsatz
// wie beim Setpaket: es ist eine Eigenschaft der Eingabe, und wer sie gewählt
// hat, kann etwas tun.
func TestKaputtesMTGJSONIstEineAblehnung(t *testing.T) {
	kaputt := filepath.Join(t.TempDir(), "kaputt.json")
	lege(t, kaputt, `{"data": {"name": "Probe", `)

	_, err := service.NeuerImporteur(&speicherDouble{}).MTGJSON(t.Context(), kaputt, "magic")
	if !errors.Is(err, domain.ErrValidation) ||
		!strings.HasPrefix(err.Error(), "kaputt.json ist kein lesbares MTG JSON") {
		t.Errorf("%v — Satz %q", err, domain.SatzVon(err))
	}
}

// **Was weder fehlt noch gesperrt ist noch am Inhalt liegt, bleibt ein
// Fehler ohne Einordnung** — ein Pfad, der durch eine Datei hindurchführt,
// meldet das Betriebssystem mit „not a directory", und dieser Satz gehört ins
// Protokoll und nicht an eine Fläche (§8).
func TestEinPfadDurchEineDateiBleibtUneingeordnet(t *testing.T) {
	datei := filepath.Join(t.TempDir(), "datei")
	lege(t, datei, "x")
	pfad := filepath.Join(datei, "kind.mse-set")

	for name, ruf := range map[string]func() error{
		"einfahren": func() error {
			_, err := service.NeuerImporteur(&speicherDouble{}).Set(t.Context(), pfad)

			return err
		},
		"mtgjson": func() error {
			_, err := service.NeuerImporteur(&speicherDouble{}).MTGJSON(t.Context(), pfad, "magic")

			return err
		},
	} {
		if err := ruf(); err == nil || errors.Is(err, domain.ErrValidation) || errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// **Und ein Wert der falschen Art ebenso** — `"name": 5` ist JSON, aber kein
// MTG JSON.
func TestMTGJSONMitFalscherArtIstEineAblehnung(t *testing.T) {
	falsch := filepath.Join(t.TempDir(), "falsch.json")
	lege(t, falsch, `{"data": {"name": 5, "code": "PRB", "cards": []}}`)

	_, err := service.NeuerImporteur(&speicherDouble{}).MTGJSON(t.Context(), falsch, "magic")
	if !errors.Is(err, domain.ErrValidation) ||
		!strings.HasPrefix(err.Error(), "falsch.json ist kein lesbares MTG JSON") {
		t.Errorf("%v — Satz %q", err, domain.SatzVon(err))
	}
}
