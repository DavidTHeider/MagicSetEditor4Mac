package ollama

// **Das Kontextfenster eines Modells** — AI-01.AK-02
// (#1742).
//
// Ollama nennt zwei Zahlen (`/api/show`, nachgelesen am 2026-10-04 in
// `docs/api.md` über Context7): unter `model_info` die **trainierte** Länge,
// `<general.architecture>.context_length`, und unter `parameters` das, was
// die Laufzeit **liest**, `num_ctx`. Gilt, was gelesen wird; fehlt es, steht
// die trainierte Länge da, und der Beleg sagt, dass die Laufzeit weniger
// lesen kann — welche Vorgabe sie dann nimmt, nennt `/api/show` nicht.

import (
	"encoding/json"
	"strconv"
	"strings"
)

// kontextAus liest das Kontextfenster aus Parametern und `model_info`.
func kontextAus(parameter string, angaben map[string]json.RawMessage) (fenster int, beleg string) {
	trainiert := trainierteLaenge(angaben)

	if gelesen := numCtx(parameter); gelesen > 0 {
		beleg = "Ollama, num_ctx aus den Parametern des Modells"
		if trainiert > 0 {
			beleg += " (trainiert für " + strconv.Itoa(trainiert) + ")"
		}

		return gelesen, beleg
	}

	if trainiert > 0 {
		return trainiert, "Ollama, model_info: trainiert für so viele — ohne num_ctx " +
			"nimmt die Laufzeit ihre Vorgabe, und die kann kleiner sein"
	}

	return 0, ""
}

// trainierteLaenge liest `<architektur>.context_length` — 0, wo es fehlt.
func trainierteLaenge(angaben map[string]json.RawMessage) int {
	var architektur string
	if json.Unmarshal(angaben["general.architecture"], &architektur) != nil || architektur == "" {
		return 0
	}

	var laenge int
	if json.Unmarshal(angaben[architektur+".context_length"], &laenge) != nil {
		return 0
	}

	return laenge
}

// numCtx liest `num_ctx` aus den Parametern, eine Zeile je Wert — 0, wo es
// fehlt oder keine Zahl ist.
func numCtx(parameter string) int {
	for _, zeile := range strings.Split(parameter, "\n") {
		name, wert, gefunden := strings.Cut(strings.TrimSpace(zeile), " ")
		if !gefunden || name != "num_ctx" {
			continue
		}

		if zahl, fehler := strconv.Atoi(strings.TrimSpace(wert)); fehler == nil && zahl > 0 {
			return zahl
		}
	}

	return 0
}
