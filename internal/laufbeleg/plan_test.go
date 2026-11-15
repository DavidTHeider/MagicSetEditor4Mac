package laufbeleg_test

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// markierteFaelle sind die Fälle, gegen die ein Plan gelesen wird.
var markierteFaelle = []laufbeleg.Fall{
	{Paket: "m.test/internal/app", Test: "TestKanon"},
	{Paket: "m.test/internal/app", Test: "TestAbstand"},
	{Paket: "m.test/internal/app", Test: "TestSvg"},
}

// einPlan ist ein gültiger Plan.
const einPlan = `{
  "lauf": "fremdpakete",
  "tag": "fremdpakete",
  "beleg": "docs/belege/fremdpakete.json",
  "roh": ".laufbeleg/fremdpakete.jsonl",
  "frist": "720h",
  "zeitlimit": "7200s",
  "misst": ["*.go", "go.mod", "go.sum"],
  "eingaben": [{"name": "MSE_KANON", "muster": "*.mse-set"}],
  "schalter": {"MSE_STICHPROBE": "bestimmt, wie viele Karten gemessen werden"},
  "ohneWirkung": {"MSE_AUSGABE": "ein Ablageort, kein Maß"},
  "trend": {"internal/app.TestAbstand": "absichtlich rot, als Verlauf zu lesen"},
  "optional": {"internal/app.TestSvg": "eine Vorrichtung, die nur mit MSE_SVG_AUSGABE schreibt"}
}`

// **Ein Plan sagt, was ein Lauf misst, womit und wohin der Beleg geht.**
func TestEinPlanWirdGelesen(t *testing.T) {
	t.Parallel()

	plan, fehler := laufbeleg.LiesPlan(strings.NewReader(einPlan), "m.test", markierteFaelle)
	if fehler != nil {
		t.Fatal(fehler)
	}

	abstand := laufbeleg.Fall{Paket: "m.test/internal/app", Test: "TestAbstand"}
	svg := laufbeleg.Fall{Paket: "m.test/internal/app", Test: "TestSvg"}

	if plan.Lauf != "fremdpakete" || plan.Tag != "fremdpakete" || plan.Frist != 720*time.Hour ||
		plan.Zeitlimit != 2*time.Hour ||
		plan.Beleg != "docs/belege/fremdpakete.json" || plan.Roh != ".laufbeleg/fremdpakete.jsonl" ||
		len(plan.Misst) != 3 || len(plan.Eingaben) != 1 || plan.Eingaben[0].Muster != "*.mse-set" ||
		len(plan.Schalter) != 1 || plan.Schalter[0] != "MSE_STICHPROBE" {
		t.Errorf("gelesen %+v", plan)
	}

	if len(plan.Erwartung.Faelle) != 3 || plan.Erwartung.Trend[abstand] == "" ||
		plan.Erwartung.Optional[svg] == "" {
		t.Errorf("Erwartung %+v", plan.Erwartung)
	}
}

// **Was ein Plan nicht sein darf** — und jede Abweisung sagt, woran es liegt.
func TestEinFalscherPlanWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		alt, neu, grund string
	}{
		"kein JSON":                     {`{`, `[`, "kein Plan"},
		"ein unbekanntes Feld":          {`"tag"`, `"tags"`, "kein Plan"},
		"ohne Lauf":                     {`"lauf": "fremdpakete"`, `"lauf": ""`, "nennt keinen Lauf"},
		"ohne Tag":                      {`"tag": "fremdpakete"`, `"tag": ""`, "nennt keinen Tag"},
		"ohne Beleg":                    {`"beleg": "docs/belege/fremdpakete.json"`, `"beleg": ""`, "nennt keinen Ort für den Beleg"},
		"ohne Rohprotokoll":             {`"roh": ".laufbeleg/fremdpakete.jsonl"`, `"roh": ""`, "nennt keinen Ort für das Rohprotokoll"},
		"ohne was er misst":             {`"misst": ["*.go", "go.mod", "go.sum"]`, `"misst": []`, "nennt nicht, was er misst"},
		"eine Frist, die keine ist":     {`"frist": "720h"`, `"frist": "ein Monat"`, "Frist"},
		"ein Zeitlimit, das keines ist": {`"zeitlimit": "7200s"`, `"zeitlimit": "lang"`, "Zeitlimit"},
		// **Null hieße bei `go test`: gar keine Grenze** — und ein hängender
		// Lauf hinge dann für immer.
		"ein Zeitlimit von null":           {`"zeitlimit": "7200s"`, `"zeitlimit": "0s"`, "Zeitlimit"},
		"eine Eingabe ohne Namen":          {`{"name": "MSE_KANON", "muster": "*.mse-set"}`, `{"name": ""}`, "eine Eingabe ohne Namen"},
		"ein Trend ohne Grund":             {`"absichtlich rot, als Verlauf zu lesen"`, `" "`, "ohne Grund"},
		"eine Ausnahme ohne Grund":         {`"eine Vorrichtung, die nur mit MSE_SVG_AUSGABE schreibt"`, `""`, "ohne Grund"},
		"ein Trend, den es nicht gibt":     {`"internal/app.TestAbstand"`, `"internal/app.TestGibtsNicht"`, "internal/app.TestGibtsNicht gibt es nicht"},
		"eine Ausnahme, die es nicht gibt": {`"internal/app.TestSvg"`, `"internal/app.TestSVG"`, "internal/app.TestSVG gibt es nicht"},
		"ein Schlüssel ohne Punkt":         {`"internal/app.TestSvg"`, `"TestSvg"`, "TestSvg gibt es nicht"},
		"ein Fall, der beides ist":         {`"internal/app.TestSvg"`, `"internal/app.TestAbstand"`, "ist Trend und Ausnahme zugleich"},
		// **Ein Schalter ist eine Behauptung über die Messung** — ohne Grund
		// ist er keine, und einer, der zugleich Eingabe oder wirkungslos
		// hieße, widerspräche sich.
		"ein Schalter ohne Grund":             {`"bestimmt, wie viele Karten gemessen werden"`, `""`, "ohne Grund"},
		"etwas Wirkungsloses ohne Grund":      {`"ein Ablageort, kein Maß"`, `" "`, "ohne Grund"},
		"ein Schalter, der Eingabe ist":       {`"MSE_STICHPROBE": "bestimmt`, `"MSE_KANON": "bestimmt`, "MSE_KANON steht zweimal"},
		"ein Schalter, der wirkungslos heißt": {`"MSE_AUSGABE": "ein Ablageort`, `"MSE_STICHPROBE": "ein Ablageort`, "MSE_STICHPROBE steht zweimal"},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			text := strings.Replace(einPlan, einer.alt, einer.neu, 1)
			if text == einPlan {
				t.Fatalf("%q steht nicht im Plan — der Fall prüft nichts", einer.alt)
			}

			_, fehler := laufbeleg.LiesPlan(strings.NewReader(text), "m.test", markierteFaelle)
			if fehler == nil || !strings.Contains(fehler.Error(), einer.grund) {
				t.Errorf("Fehler %v, erwartet %q", fehler, einer.grund)
			}
		})
	}
}

// **Ohne Frist gibt es keine** — und eine fehlende ist keine abgelaufene.
func TestEinPlanOhneFristHatKeine(t *testing.T) {
	t.Parallel()

	plan, fehler := laufbeleg.LiesPlan(strings.NewReader(
		strings.Replace(einPlan, `"frist": "720h",`, "", 1)), "m.test", markierteFaelle)
	// **Und der Rest ist gelesen** — sonst hielte diese Zusicherung auch für
	// einen Leser, der gar nichts liest.
	if fehler != nil || plan.Frist != 0 || plan.Lauf != "fremdpakete" {
		t.Errorf("Frist %v, Lauf %q, Fehler %v", plan.Frist, plan.Lauf, fehler)
	}
}
