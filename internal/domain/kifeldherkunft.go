package domain

// **Woher ein übernommener Feldwert kommt** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Sie heißt **Feld**herkunft, und das ist kein Beiwerk
//
// **[KIHerkunft] gibt es schon** (A6): das ist die Kette, die sagt, woher ein
// **Modell** stammt — Gewichte, Lizenz, Prüfsummen. Hier geht es um etwas
// anderes: woher ein **Feldwert** stammt. Zwei Fragen, zwei Namen.
//
// # Warum das mehr ist als die Marke »ki«
//
// **Seit A7 trägt ein übernommenes Feld die Herkunft `ki`.** Das sagt, dass
// es nicht getippt wurde, und sonst nichts — nicht **wer** geantwortet hat,
// nicht **worauf**, nicht **wann**.
//
// Das Issue verlangt fünf Dinge, und zwar **bei der ersten Übernahme** und
// nicht später: „Provider/Modell/Aufgabe/Zeit/Inhaltsbezug".
//
// # Und eines steht ausdrücklich **nicht** darin
//
// **Der Rohprompt.** Das Issue: „Rohprompt standardmäßig aus." Die
// **Aufgabe** ist, was ein Mensch getippt hat; der **Rohprompt** ist, was
// daraus wurde — Anweisung, Schema, Feldliste, mitgeschickte Bilder. Das eine
// gehört hierher, das andere nicht: es ist lang, es enthält den ganzen
// Kartenstand, und niemand hat darum gebeten, dass er aufbewahrt wird.

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// KIFeldherkunft sagt, woher der Wert eines Feldes stammt.
type KIFeldherkunft struct {
	// Feld ist, an welchem Feld der Wert steht.
	Feld string

	// Anbieter ist, wen man gefragt hat.
	Anbieter string

	// Modell ist, was **geantwortet** hat.
	//
	// **Nicht der Alias, unter dem gefragt wurde**: ein Anbieter darf
	// tauschen, was dahinterliegt, und dann wäre die Auskunft eine über
	// etwas anderes (A9, [KIModellprobe]).
	Modell string

	// Aufgabe ist, was ein Mensch getippt hat — **nicht** der Rohprompt.
	Aufgabe string

	// Auftrag ist die Kennung des Auftrags, aus dem der Wert stammt.
	//
	// **Damit sich zwei Felder einer Übernahme zusammenfinden**: wer im
	// Verlauf sieht, dass vier Felder zugleich anders wurden, soll sehen,
	// dass es **eine** Entscheidung war.
	Auftrag string

	// Zeitpunkt ist, wann übernommen wurde.
	Zeitpunkt time.Time

	// Bezug ist der Abdruck des übernommenen Werts ([KIInhaltsbezug]).
	Bezug string
}

// KIFeldherkunftsfelder nennt die Stücke, aus denen sie besteht.
//
// **Für Wächter, die nachzählen** — und für den einen, der prüft, dass hier
// **kein** Feld für den Rohprompt steht.
func KIFeldherkunftsfelder() []string {
	return []string{
		"feld", "anbieter", "modell", "aufgabe", "auftrag",
		"zeitpunkt", "bezug",
	}
}

// KIInhaltsbezug macht aus den übernommenen Werten einen Abdruck.
//
// # Wofür er da ist
//
// **Damit sich später sagen lässt, ob das Feld noch das trägt, was das
// Modell vorgeschlagen hat.** Wer den Wert danach von Hand ändert, darf das —
// das Issue sagt es ausdrücklich —, und dann soll die Herkunft nicht
// behaupten, der Inhalt stamme von der KI.
//
// **Ein Abdruck und keine Kopie**: der Wert selbst steht schon im Feld; ihn
// hier zweitzuschreiben hieße, ihn an zwei Stellen zu pflegen.
//
// **Die Grenze zwischen den Teilen zählt mit.** Ein Trennzeichen, das in
// keinem Wert vorkommen kann, macht aus `["ab","c"]` und `["a","bc"]` zwei
// verschiedene Abdrücke — mit einem Komma wären es dieselben.
func KIInhaltsbezug(werte []string) string {
	summe := sha256.New()

	for _, wert := range werte {
		summe.Write([]byte(wert)) //nolint:errcheck // ein Hash schreibt immer
		summe.Write([]byte{0})    //nolint:errcheck // siehe oben
	}

	return "sha256:" + hex.EncodeToString(summe.Sum(nil))
}

// Validate prüft, dass keines der fünf Dinge fehlt.
func (h KIFeldherkunft) Validate() []*FieldError {
	var befunde []*FieldError

	for _, paar := range []struct{ feld, wert string }{
		{"feld", h.Feld},
		{"anbieter", h.Anbieter},
		{"modell", h.Modell},
		{"aufgabe", h.Aufgabe},
		{"auftrag", h.Auftrag},
		{"bezug", h.Bezug},
	} {
		if strings.TrimSpace(paar.wert) == "" {
			befunde = append(befunde, NewFieldError(paar.feld,
				SchluesselPflichtfeld,
				"Eine Herkunft nennt ihren %s.", paar.feld))
		}
	}

	// **Die Zeit ist eines der fünf Dinge.** Ohne sie sagt die Herkunft
	// nicht, ob sie von heute oder von einem Modellstand vor einem Jahr
	// stammt.
	if h.Zeitpunkt.IsZero() {
		befunde = append(befunde, NewFieldError("zeitpunkt",
			SchluesselPflichtfeld,
			"Eine Herkunft nennt, wann übernommen wurde."))
	}

	return befunde
}
