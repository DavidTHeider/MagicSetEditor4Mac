package openai

// **Die Modellliste, und was sie nicht belegt** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Der Fund
//
// **Nachgelesen am 2026-09-12** unter
// <https://developers.openai.com/api/docs/api-reference/models/list>: ein
// Modellobjekt trägt `id`, `object`, `created`, `owned_by` und ein optionales
// `shutdown_date`. **Mehr nicht.**
//
// Anthropic nennt `capabilities.structured_outputs.supported` (A8), Ollama
// nennt `capabilities` in `/api/show` (A7). Hier nennt niemand etwas — und
// [0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)
// sagt, was dann gilt: „unbekannt bleibt unbekannt; Modelllisten garantieren
// keine Capability."
//
// # Was daraus folgt
//
// **Aus dieser Liste entsteht kein Katalogeintrag** — je Modell ein Befund,
// der den Grund nennt und sagt, was ihn auflöst. Wer aus dem Namen `gpt-`
// schlösse, dass ein Modell Bilder liest, hätte Fachlogik aus einem
// Bezeichner abgeleitet (§15, Satz 5).

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// modellantwort ist die Gestalt von `/v1/models`.
type modellantwort struct {
	Daten []modellDTO `json:"data"`
}

// modellDTO ist ein Modell, wie OpenAI es nennt.
//
// **Vier Felder, und keines davon ist eine Fähigkeit.** `Abschaltung` steht
// hier, weil es die einzige Angabe ist, die etwas über die Zukunft sagt — und
// sie gehört in den Befund: ein Modell, das abgeschaltet wird, lohnt keine
// Prüfung mehr.
type modellDTO struct {
	Kennung     string `json:"id"`
	Angelegt    int64  `json:"created"`
	Eigner      string `json:"owned_by"`
	Abschaltung string `json:"shutdown_date"`
}

// Modelle gibt, was OpenAI führt — als Befunde.
//
// **Und die leere Modell-Liste ist kein Fehler.** Sie ist die Auskunft: es
// steht nichts fest. Der Einrichtungsstand daneben heißt dann
// `ohnefaehigkeit`, und genau dafür gibt es ihn (A7).
func (k *Klient) Modelle(ctx context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	var antwort modellantwort

	if fehler := k.hole(ctx, Modellweg, nil, &antwort); fehler != nil {
		return nil, nil, fehler
	}

	modelle := make([]domain.KIModell, 0)
	befunde := make([]domain.Befund, 0, len(antwort.Daten))

	for stelle, dto := range antwort.Daten {
		befunde = append(befunde, ohneAuskunft(dto, stelle))
	}

	return modelle, befunde, nil
}

// ohneAuskunft baut den Befund über ein Modell ohne Fähigkeitsangabe.
//
// **Er nennt den Weg, nicht bloß das Modell.** Wer liest „gpt-probe: keine
// Fähigkeiten", sucht den Fehler bei sich; wer liest, dass `/v1/models`
// keine nennt, weiß, dass es nichts zu suchen gibt.
func ohneAuskunft(dto modellDTO, stelle int) domain.Befund {
	meldung := fmt.Sprintf(
		"OpenAI nennt an %s keine Fähigkeiten — auch nicht für %q. Eine "+
			"Modellliste ist keine belegte Fähigkeit; was dieses Modell "+
			"kann, steht erst fest, wenn es gemessen wurde.",
		Modellweg, dto.Kennung)

	// **Eine angekündigte Abschaltung gehört dazu.** Sie ist die einzige
	// Angabe der Liste, die etwas über die Zukunft sagt — und sie ändert die
	// Antwort auf „lohnt sich eine Messung?".
	if dto.Abschaltung != "" {
		meldung += fmt.Sprintf(" Es wird am %s abgeschaltet.", dto.Abschaltung)
	}

	return domain.Befund{
		Datei: dto.Kennung, Ort: stelleIn(stelle),
		Art: domain.ArtUmgebung, Meldung: meldung,
	}
}

// stelleIn benennt einen Eintrag der Modellliste über seine Stelle.
//
// # Warum das dasteht
//
// **Weil ein Eintrag ohne `id` sonst nicht zu benennen wäre** — Werkstück A19
// von Phase 5.3
// (#1550),
// gefunden vom Fuzzer bei der Saat `{"data":[{}]}`. Zwei solche Einträge
// gaben zwei **wortgleiche** Befunde, und §8 verlangt für jeden Fehler die
// Stelle. Derselbe Fund stand zugleich in `anthropic`, `xai` und `ollama` —
// es ist eine Klasse und kein Fall (§10).
func stelleIn(stelle int) string {
	return "data[" + strconv.Itoa(stelle) + "]"
}
