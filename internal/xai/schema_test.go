package xai_test

// **Was xAI von einem Schema annimmt** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Warum das hier geprüft wird und nicht beim Anbieter
//
// **Das Issue verlangt es**: „Unterstützte Schemasubsets lokal
// nachvalidieren." Der Grund steht eine Zeile weiter unten in derselben
// Leitlinie wie bei [xai.Frage.pruefe]: ein Schema, das der Dienst abweist,
// **kostet den Ruf trotzdem** — und was zurückkommt, ist ein 400, den diese
// Seite selbst kennen kann.
//
// **Vier Fälle, und sie sind nachgelesen und nicht ausgedacht**
// (2026-09-12, <https://docs.x.ai/developers/model-capabilities/text/structured-outputs>):
// xAI nennt ausdrücklich, was es mit 400 beantwortet.
//
// # Und was hier **nicht** abgewiesen wird
//
// **Alles, was xAI »best effort« nennt** — `not`, `if`/`then`/`else`,
// mehrteiliges `allOf`, Längen über den garantierten Grenzen. Es wird
// **angenommen** und nur nicht erzwungen. Wer es hier abwiese, hätte sich
// eine Regel ausgedacht und dem Aufrufer etwas verboten, das geht (§10).

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// **Was xAI mit 400 beantwortet, geht hier nicht hinaus.**
func TestEinAbgewiesenesSchemaGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		schema map[string]any
		kern   string
	}{
		{
			"enum ohne einen einzigen Wert",
			map[string]any{"type": "object", "properties": map[string]any{
				"farbe": map[string]any{"enum": []any{}},
			}},
			"enum",
		},
		{
			"anyOf ohne eine einzige Möglichkeit",
			map[string]any{"type": "object", "properties": map[string]any{
				"kosten": map[string]any{"anyOf": []any{}},
			}},
			"anyOf",
		},
		{
			// **`true` ist in JSON-Schema ein gültiges Schema** — „alles ist
			// erlaubt". xAI nimmt es nicht.
			"eine Eigenschaft, deren Schema true ist",
			map[string]any{"type": "object", "properties": map[string]any{
				"beliebig": true,
			}},
			"beliebig",
		},
		{
			"eine Eigenschaft, deren Schema false ist",
			map[string]any{"type": "object", "properties": map[string]any{
				"nie": false,
			}},
			"nie",
		},
		{
			"maxContains",
			map[string]any{"type": "object", "properties": map[string]any{
				"fähigkeiten": map[string]any{
					"type": "array", "maxContains": 2,
				},
			}},
			"maxContains",
		},
		{
			"minContains",
			map[string]any{"type": "object", "properties": map[string]any{
				"fähigkeiten": map[string]any{
					"type": "array", "minContains": 1,
				},
			}},
			"minContains",
		},
		{
			// **Die Tupelform von JSON-Schema 2019** — xAI will
			// `prefixItems`, und die Meldung sagt es.
			"items als Liste statt als Schema",
			map[string]any{"type": "object", "properties": map[string]any{
				"paar": map[string]any{"type": "array", "items": []any{
					map[string]any{"type": "string"},
				}},
			}},
			"prefixItems",
		},
		{
			// **Ganz oben**, ohne Stelle darunter — die Meldung hat dann
			// keinen Ort zu nennen und nennt trotzdem den Grund.
			"maxContains im Schema selbst",
			map[string]any{"type": "object", "maxContains": 1},
			"maxContains",
		},
		{
			// **In einer Wahl**, also unter einer Liste von Schemata.
			"ein verbotenes Wort in einem anyOf-Zweig",
			map[string]any{"type": "object", "properties": map[string]any{
				"farbe": map[string]any{"anyOf": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "array", "minContains": 1},
				}},
			}},
			"minContains",
		},
		{
			// **Und eine Eigenschaft, die gar kein Schema ist.** Sie kommt
			// aus einer Vorlage, und dort steht, was dort steht.
			"items als Liste unter prefixItems",
			map[string]any{"type": "object", "prefixItems": []any{
				map[string]any{"type": "array", "items": []any{}},
			}},
			"prefixItems",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fuhrwerk := &mitschrift{}

			klient := xai.NeuerKlient("",
				netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
				xai.Zugang{Schluessel: "geheim"})

			_, fehler := klient.Rede(t.Context(), xai.Frage{
				Alias: "m", Auftrag: "x", Schema: fall.schema,
			})

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.kern) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.kern, fehler)
			}

			// **Und der Ruf ist nicht hinausgegangen** — das ist der ganze
			// Zweck: ein 400 vom Anbieter wäre bezahlt.
			if fuhrwerk.adresse != "" {
				t.Errorf("gerufen wurde trotzdem %q", fuhrwerk.adresse)
			}
		})
	}
}

// **Und was xAI »best effort« nennt, geht hinaus.**
//
// **Die Gegenprobe zur Liste oben.** Ein Prüfer, der zu viel abweist, sieht
// aus wie einer, der sorgfältig ist — und verbietet dem Aufrufer etwas, das
// der Dienst annimmt.
func TestEinBestEffortSchemaGehtHinaus(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"model":"m","choices":[{"message":` +
		`{"content":"x"},"finish_reason":"stop"}]}`}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", Auftrag: "x",
		Schema: map[string]any{
			"type": "object",

			// **`additionalProperties: false` ist der Normalfall** und
			// **kein** Schema an einer Eigenschaft — wer beide Wahrheitswerte
			// über einen Kamm schöre, wiese jedes übliche Schema ab.
			"additionalProperties": false,

			"properties": map[string]any{
				"name": map[string]any{
					"type": "string", "maxLength": 4096,
				},
				"text": map[string]any{
					"not": map[string]any{"type": "number"},
				},
				"paar": map[string]any{
					"type":        "array",
					"prefixItems": []any{map[string]any{"type": "string"}},
				},
			},
		},
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if fuhrwerk.adresse == "" {
		t.Error("der Ruf ist nicht hinausgegangen")
	}
}

// **Tief liegende Fälle werden auch gefunden.**
//
// **Weil ein Kartenschema verschachtelt ist**: die Felder einer Karte stehen
// unter `properties`, und was darunter liegt, ist wieder ein Schema. Ein
// Prüfer, der nur die oberste Ebene ansähe, wäre grün und hielte nichts.
func TestEinTiefLiegenderFallWirdGefunden(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", Auftrag: "x",
		Schema: map[string]any{"type": "object", "$defs": map[string]any{
			"karte": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"seiten": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"tief": map[string]any{"enum": []any{}},
							},
						},
					},
				},
			},
		}},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if fuhrwerk.adresse != "" {
		t.Errorf("gerufen wurde trotzdem %q", fuhrwerk.adresse)
	}
}

// **Eine »properties«-Angabe, die kein Objekt ist, hält hier nichts auf.**
//
// **Weil sie kein dokumentierter Ablehnungsgrund ist.** Sie ist Unsinn, und
// xAI sagt über Unsinn nichts — wer sie hier abwiese, hätte sich die Regel
// ausgedacht (§10). Was der Dienst damit macht, sagt der Dienst.
func TestEineUnsinnigeEigenschaftsangabeHaeltNichtsAuf(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"model":"m","choices":[{"message":` +
		`{"content":"x"},"finish_reason":"stop"}]}`}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", Auftrag: "x",
		Schema: map[string]any{"type": "object", "properties": "unsinn"},
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if fuhrwerk.adresse == "" {
		t.Error("der Ruf ist nicht hinausgegangen")
	}
}
