package ollama

// **Die Vermessung eines Einbettungsmodells** — Werkstück A7 von Phase 5.3
// (#1446),
// Ergänzung vom 2026-09-10.
//
// # Vier Teile, vier Quellen — und keine geraten
//
// A6 verlangt von einem Einbettungsmodell vier Angaben
// ([domain.KIEinbettungsvertrag]). Sie kommen von hier:
//
//	| Teil             | Woher                                                |
//	| ---------------- | ---------------------------------------------------- |
//	| Zusammenfassung  | `model_info`, Schlüssel auf `.pooling_type`           |
//	| Zerteiler        | `model_info["tokenizer.ggml.model"]`                  |
//	| Dimension        | **gemessen** an einem echten Lauf                     |
//	| Normalisiert     | **gemessen** am Betrag des Vektors                    |
//
// **Das Issue verlangt genau das**: „Capabilities und Modellfingerprint statt
// Namensheuristik", und als Abnahme „echter lokaler Embeddinglauf". Ein
// Vertrag aus einem Namen wäre eine Behauptung; dieser ist eine Messung mit
// zwei nachgelesenen Metadaten daneben.
//
// # Was fehlt, wird gesagt und nicht ergänzt
//
// **Ohne Zusammenfassung gibt es keinen Vertrag.** Sie ist der eine Teil, der
// sich nicht messen lässt, und wer sie riete, bekäme Vektoren, die
// systematisch danebenliegen, **ohne dass etwas scheitert**.

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// derZerteilerschluessel ist der feste Schlüssel des Tokenizers.
//
// **Nachgelesen am 2026-09-11** in `gguf-py/gguf/constants.py` von llama.cpp:
// `Keys.Tokenizer.MODEL = "tokenizer.ggml.model"` — ohne Architekturpräfix.
const derZerteilerschluessel = "tokenizer.ggml.model"

// derZusammenfassungssuffix ist das Ende des Poolingschlüssels.
//
// **Ein Suffix und kein ganzer Name**: derselbe Quelltext schreibt
// `Keys.LLM.POOLING_TYPE = "{arch}.pooling_type"` — das Präfix ist die
// Architektur und heißt bei jedem Modell anders.
const derZusammenfassungssuffix = ".pooling_type"

// diePoolings ordnet GGUFs Zahlen den Namen aus A6 zu.
//
// **Nachgelesen am 2026-09-11**, ebenda: `PoolingType` zählt `NONE = 0`,
// `MEAN = 1`, `CLS = 2`, `LAST = 3`, `RANK = 4`.
//
// **Null und vier stehen nicht darin, und das ist die Aussage**: bei `NONE`
// kommt ein Vektor je Zeichen heraus und keiner je Text, bei `RANK` eine
// Bewertung statt einer Einbettung. Für beides kennt A6 keinen Vertrag — und
// „Reranking bleibt eigene Capability und wird nicht aus vorhandenem
// `/api/embed` abgeleitet" steht so im Issue.
func diePoolings() map[int]domain.KIPooling {
	return map[int]domain.KIPooling{
		1: domain.KIPoolingMittel,
		2: domain.KIPoolingErster,
		3: domain.KIPoolingLetzter,
	}
}

// dieProbe ist der Text, an dem gemessen wird.
const dieProbe = "Blitzschlag"

// Vermessen baut den Vertrag eines Einbettungsmodells.
//
// **Die Befunde sind das Ergebnis und nicht der Fehler**: ein Modell, dessen
// Vertrag nicht zu füllen ist, ist kein Ausfall der Laufzeit. Der Fehler
// bleibt dem vorbehalten, was gar nicht geantwortet hat (§8).
func (k *Klient) Vermessen(ctx context.Context, alias string) (
	domain.KIEinbettungsvertrag, []domain.Befund, error,
) {
	var schau struct {
		Faehigkeiten []string                   `json:"capabilities"`
		Angaben      map[string]json.RawMessage `json:"model_info"`
	}

	if fehler := k.hole(ctx, Schauweg,
		map[string]any{"model": alias}, &schau); fehler != nil {
		return domain.KIEinbettungsvertrag{}, nil, fehler
	}

	if !slices.Contains(schau.Faehigkeiten, derEinbetter) {
		return domain.KIEinbettungsvertrag{}, []domain.Befund{{
			Datei: alias, Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf("%q meldet %q nicht — vermessen wird, was die "+
				"Laufzeit als Einbetter führt", alias, derEinbetter),
		}}, nil
	}

	vertrag, befunde := ausAngaben(alias, schau.Angaben)
	if len(befunde) > 0 {
		return domain.KIEinbettungsvertrag{}, befunde, nil
	}

	return k.messen(ctx, alias, vertrag)
}

// ausAngaben liest Zusammenfassung und Zerteiler aus den Metadaten.
func ausAngaben(alias string, angaben map[string]json.RawMessage) (
	domain.KIEinbettungsvertrag, []domain.Befund,
) {
	var vertrag domain.KIEinbettungsvertrag

	zerteiler, nennt := angaben[derZerteilerschluessel]
	if !nennt {
		return vertrag, []domain.Befund{fehlt(alias, derZerteilerschluessel,
			"ohne ihn ist nicht zu sagen, was gleich behandelt wird")}
	}

	if fehler := json.Unmarshal(zerteiler, &vertrag.Zerteiler); fehler != nil {
		return vertrag, []domain.Befund{fehlt(alias, derZerteilerschluessel,
			"der Wert dort ist kein Name")}
	}

	roh, schluessel := poolingAus(angaben)
	if schluessel == "" {
		return vertrag, []domain.Befund{fehlt(alias, derZusammenfassungssuffix,
			"sie lässt sich nicht messen, und geraten läge jeder Vektor "+
				"systematisch daneben")}
	}

	pooling, bekannt := diePoolings()[roh]
	if !bekannt {
		return vertrag, []domain.Befund{{
			Datei: alias, Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf("%s steht auf %d — dafür kennt dieses "+
				"Programm keinen Einbettungsvertrag", schluessel, roh),
		}}
	}

	vertrag.Pooling = pooling

	return vertrag, nil
}

// poolingAus sucht den Schlüssel auf `.pooling_type` und seinen Wert.
func poolingAus(angaben map[string]json.RawMessage) (wert int, schluessel string) {
	for name, roh := range angaben {
		if !strings.HasSuffix(name, derZusammenfassungssuffix) {
			continue
		}

		var gelesen int

		if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
			continue
		}

		return gelesen, name
	}

	return 0, ""
}

// messen fährt den echten Lauf und füllt Dimension und Normalisierung.
func (k *Klient) messen(
	ctx context.Context, alias string, vertrag domain.KIEinbettungsvertrag,
) (domain.KIEinbettungsvertrag, []domain.Befund, error) {
	var antwort einbettungsantwort

	if fehler := k.hole(ctx, Einbettungsweg, einbettungsanfrage{
		Modell: alias, Eingaben: []string{dieProbe}, Kuerzen: false,
	}, &antwort); fehler != nil {
		return domain.KIEinbettungsvertrag{}, nil, fehler
	}

	if len(antwort.Vektoren) != 1 {
		return domain.KIEinbettungsvertrag{}, []domain.Befund{{
			Datei: alias, Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf("auf einen Text kamen %d Vektoren zurück",
				len(antwort.Vektoren)),
		}}, nil
	}

	vertrag.Dimension = len(antwort.Vektoren[0])
	vertrag.Normalisiert = istNormalisiert(antwort.Vektoren[0])

	// **Ein unvollständiger Vertrag ist ein Befund und kein Fehler.** Die
	// Laufzeit hat geantwortet; was sie gab, trägt nur keinen Vertrag. Er
	// wird deshalb umgewandelt und nicht durchgereicht — so steht nirgends
	// ein `nil` neben einer Ursache, und niemand muss raten, ob das Absicht
	// war (dieselbe Gestalt wie in `service.formfehlerAlsBefund`).
	if befunde := formfehlerAlsBefund(alias, vertrag); befunde != nil {
		return domain.KIEinbettungsvertrag{}, befunde, nil
	}

	return vertrag, nil, nil
}

// formfehlerAlsBefund macht aus einem unvollständigen Vertrag einen Befund.
func formfehlerAlsBefund(
	alias string, vertrag domain.KIEinbettungsvertrag,
) []domain.Befund {
	fehler := vertrag.Pruefe()
	if fehler == nil {
		return nil
	}

	return []domain.Befund{{
		Datei: alias, Art: domain.ArtUmgebung, Meldung: fehler.Error(),
	}}
}

// istNormalisiert sagt, ob der Vektor die Länge eins hat.
//
// **Gemessen und nicht behauptet**: die Toleranz ist dieselbe, gegen die der
// Vertragslauf aus A6 prüft — sonst vermäße dieses Paket etwas, das jener
// gleich darauf beanstandete.
func istNormalisiert(vektor []float32) bool {
	quadrat := float64(0)

	for _, wert := range vektor {
		quadrat += float64(wert) * float64(wert)
	}

	return math.Abs(math.Sqrt(quadrat)-1) <= 1e-4
}

// fehlt baut den Befund über einen fehlenden Schlüssel.
func fehlt(alias, schluessel, grund string) domain.Befund {
	return domain.Befund{
		Datei: alias, Art: domain.ArtUmgebung,
		Meldung: fmt.Sprintf("%q nennt %s nicht — %s", alias, schluessel, grund),
	}
}
