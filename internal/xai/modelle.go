package xai

// **Die Modellliste und was sie belegt** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Der Fund
//
// **Nachgelesen am 2026-09-12** unter
// <https://docs.x.ai/developers/rest-api-reference/inference/models>: ein
// Modell von `/v1/language-models` trägt `input_modalities`,
// `output_modalities`, `aliases`, `fingerprint`, `version` und Preise.
//
// **Damit stehen zwei Fähigkeiten fest** — Text und Bildlesen —, und eine
// steht ausdrücklich **nicht** dabei: die **feste Form**. Die Schnittstelle
// kann sie (`response_format`), die Liste sagt nichts über das einzelne
// Modell, und was nicht dasteht, ist nicht belegt
// ([0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)).
// Dafür gibt es die Messung aus A9.

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Modalitäten, die dieses Programm deuten kann.
const (
	// modalitaetText heißt: Text hinein oder heraus.
	modalitaetText = "text"

	// modalitaetBild heißt: Bilder hinein.
	modalitaetBild = "image"
)

// modellantwort ist die Gestalt von `/v1/language-models`.
//
// **`models` und nicht `data`** — auch darin weicht xAI ab.
type modellantwort struct {
	Modelle []modellDTO `json:"models"`
}

// modellDTO ist ein Modell, wie xAI es nennt.
type modellDTO struct {
	Kennung  string `json:"id"`
	Angelegt int64  `json:"created"`

	// Fingerabdruck ist das Nächste, was ein Anbieter bisher zu „was läuft
	// hier wirklich" gesagt hat.
	Fingerabdruck string `json:"fingerprint"`

	Fassung string `json:"version"`

	// Zweitnamen sind weitere Namen desselben Modells.
	Zweitnamen []string `json:"aliases"`

	Eingaben []string `json:"input_modalities"`
	Ausgaben []string `json:"output_modalities"`
}

// Modelle gibt, was xAI führt — mit belegten Fähigkeiten.
func (k *Klient) Modelle(ctx context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	var antwort modellantwort

	if fehler := k.hole(ctx, Modellweg, nil, &antwort); fehler != nil {
		return nil, nil, fehler
	}

	modelle := make([]domain.KIModell, 0, len(antwort.Modelle))
	befunde := make([]domain.Befund, 0)

	for stelle, dto := range antwort.Modelle {
		// **Ohne Textausgabe gehört es nicht in diesen Katalog** — und
		// verschwindet trotzdem nicht still.
		if !slices.Contains(dto.Ausgaben, modalitaetText) {
			befunde = append(befunde, ohneTextausgabe(dto, stelle))

			continue
		}

		if tagAus(dto.Angelegt) == "" {
			befunde = append(befunde, ohneTag(dto, stelle))

			continue
		}

		modelle = append(modelle, alsModell(dto))
	}

	return modelle, befunde, nil
}

// alsModell macht aus einem DTO ein Modell mit belegten Fähigkeiten.
func alsModell(dto modellDTO) domain.KIModell {
	return domain.KIModell{
		Anbieter: Anbieter,

		// **Die Kennung und nicht ein Zweitname.** Ein Zweitname zeigt heute
		// hierauf und morgen woandershin; wer ihn als eigenen Eintrag
		// führte, führte dieselbe Sache mehrfach.
		Alias: dto.Kennung,

		// **Cloud, ohne Einstufung nach Adresse.** xAI rechnet nicht auf
		// diesem Rechner, und die Adresse steht fest.
		Ort: domain.KIOrtCloud,

		// **Fingerabdruck und Fassung stehen in der Laufzeit.**
		//
		// Eine **Herkunftskette** wäre der richtige Ort — sie verlangt aber
		// Lizenz und Dateiprüfsummen ([domain.KIGlied]), und beides nennt
		// xAI nicht. Eine halbe Kette ist schlimmer als keine (0097).
		Laufzeit: laufzeitAus(dto),

		Geprueft: belegeAus(dto),
	}
}

// laufzeitAus baut die Laufzeitangabe aus Fassung und Fingerabdruck.
func laufzeitAus(dto modellDTO) string {
	laufzeit := "xAI " + dto.Kennung

	if dto.Fassung != "" {
		laufzeit += " " + dto.Fassung
	}

	if dto.Fingerabdruck != "" {
		laufzeit += " (" + dto.Fingerabdruck + ")"
	}

	return laufzeit
}

// belegeAus macht aus den Modalitäten belegte Fähigkeiten.
func belegeAus(dto modellDTO) []domain.KIGeprueft {
	tag := tagAus(dto.Angelegt)
	woher := woherAus(dto)

	belege := []domain.KIGeprueft{{
		Faehigkeit: domain.KITextgenerierung, Datum: tag,
		Beleg: fmt.Sprintf(
			"xAI nennt für %s output_modalities %v unter %s%s",
			dto.Kennung, dto.Ausgaben, Modellweg, woher),
	}}

	// **Bilder nur, wenn sie unter den Eingaben stehen.** `image` bei den
	// **Ausgaben** hieße etwas anderes — Bilderzeugung, und die steht in A12.
	if slices.Contains(dto.Eingaben, modalitaetBild) {
		belege = append(belege, domain.KIGeprueft{
			Faehigkeit: domain.KIVision, Datum: tag,
			Beleg: fmt.Sprintf(
				"xAI nennt für %s input_modalities %v unter %s%s",
				dto.Kennung, dto.Eingaben, Modellweg, woher),
		})
	}

	return belege
}

// woherAus nennt Fingerabdruck und Zweitnamen — leer, wo es keine gibt.
//
// **Sie gehören in den Beleg**: wer ihn liest, sieht, worüber die Aussage
// spricht, und merkt, wenn hinter demselben Namen etwas anderes liegt.
func woherAus(dto modellDTO) string {
	var teile []string

	if dto.Fingerabdruck != "" {
		teile = append(teile, "Fingerabdruck "+dto.Fingerabdruck)
	}

	if len(dto.Zweitnamen) > 0 {
		teile = append(teile, "auch "+strings.Join(dto.Zweitnamen, ", "))
	}

	if len(teile) == 0 {
		return ""
	}

	return " (" + strings.Join(teile, "; ") + ")"
}

// tagAus macht aus `created` den Tag des Belegs.
//
// **Eine Sekundenzahl und kein RFC 3339** — auch darin weicht xAI ab. Null
// heißt: keine Angabe, und dann entsteht kein Eintrag.
func tagAus(sekunden int64) string {
	if sekunden <= 0 {
		return ""
	}

	return time.Unix(sekunden, 0).UTC().Format(time.DateOnly)
}

// ohneTextausgabe baut den Befund über ein Modell, das keinen Text gibt.
func ohneTextausgabe(dto modellDTO, stelle int) domain.Befund {
	return domain.Befund{
		Datei: dto.Kennung, Ort: stelleIn(stelle), Art: domain.ArtUmgebung,
		Meldung: fmt.Sprintf(
			"xAI führt %q mit output_modalities %v — daraus entsteht kein "+
				"Kartenfeld. Es gibt das Modell; diese Anwendung kann damit "+
				"nichts anfangen, und das ist zweierlei.",
			dto.Kennung, dto.Ausgaben),
	}
}

// ohneTag baut den Befund über ein Modell ohne lesbares Datum.
func ohneTag(dto modellDTO, stelle int) domain.Befund {
	return domain.Befund{
		Datei: dto.Kennung, Ort: stelleIn(stelle), Art: domain.ArtUmgebung,
		Meldung: fmt.Sprintf(
			"xAI nennt für %q kein lesbares Datum (created war %d). Ein "+
				"Beleg ohne Tag sagt nicht, seit wann er gilt — das Modell "+
				"bleibt aus dem Katalog.", dto.Kennung, dto.Angelegt),
	}
}

// stelleIn benennt einen Eintrag der Modellliste über seine Stelle.
//
// # Warum das dasteht
//
// **Weil ein Eintrag ohne `id` sonst nicht zu benennen wäre** — Werkstück A19
// von Phase 5.3
// (#1550),
// gefunden vom Fuzzer bei der Saat `{"models":[{}]}`. Zwei solche Einträge
// gaben zwei **wortgleiche** Befunde, und §8 verlangt für jeden Fehler die
// Stelle. Derselbe Fund stand zugleich in `anthropic` und `ollama` — es ist
// eine Klasse und kein Fall (§10).
func stelleIn(stelle int) string {
	return "models[" + strconv.Itoa(stelle) + "]"
}
