package anthropic

// **Die Modelle und was sie belegen** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Die Fähigkeit kommt aus der Schnittstelle
//
// **`GET /v1/models` nennt sie selbst** — `structured_outputs.supported` und
// `image_input.supported` stehen je Modell in der Antwort (**nachgelesen am
// 2026-09-11** unter <https://platform.claude.com/docs/en/api/models/list>).
//
// **Das ist der Beleg, den A6 verlangt.** Die Dokumentation führt daneben
// eine Liste der Modelle mit strukturierter Ausgabe; sie hier abzuschreiben
// hieße, eine Behauptung mit Verfallsdatum in den Quelltext zu legen — beim
// nächsten Modell stünde sie falsch da, und niemand sähe es.
//
// # Und „null" heißt nicht „kann nichts"
//
// **`capabilities` darf `null` sein.** Dann steht **nichts** fest, und das
// ist etwas anderes als „kann nichts": ein Modell mit null Fähigkeiten sähe
// im Katalog aus wie eines, das geprüft wurde. Es kommt als Befund und nicht
// als Eintrag.

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// koennen ist eine Fähigkeitsangabe der Schnittstelle.
type koennen struct {
	Unterstuetzt bool `json:"supported"`
}

// faehigkeiten sind die Angaben, die dieses Programm liest.
//
// **Vier von vielen.** Die Antwort nennt auch `batch`, `citations`,
// `code_execution`, `effort`, `thinking` und `context_management` — sie
// haben hier keine Entsprechung, und ihnen eine zu geben hieße, eine
// Fähigkeit zu behaupten, die niemand benutzt.
type faehigkeiten struct {
	Bilder    koennen `json:"image_input"`
	FesteForm koennen `json:"structured_outputs"`
	PDF       koennen `json:"pdf_input"`
	Stapel    koennen `json:"batch"`
}

// modellDTO ist ein Modell, wie die Schnittstelle es nennt.
type modellDTO struct {
	Kennung  string        `json:"id"`
	Name     string        `json:"display_name"`
	Angelegt string        `json:"created_at"`
	Koennen  *faehigkeiten `json:"capabilities"`
	Eingabe  int           `json:"max_input_tokens"`
	Ausgabe  int           `json:"max_tokens"`
}

// modellantwort ist die Seite, die `GET /v1/models` gibt.
type modellantwort struct {
	Daten   []modellDTO `json:"data"`
	Weitere bool        `json:"has_more"`
	Letzte  string      `json:"last_id"`
}

// hoechstensJeSeite ist, wie viele Modelle eine Seite bringt.
//
// **Tausend ist das Höchste, das die Schnittstelle zulässt** (nachgelesen am
// 2026-09-11: „Ranges from `1` to `1000`"). Damit passt jeder heutige Stand
// in **eine** Seite — und `has_more` wird trotzdem gelesen, weil „passt
// heute" keine Zusage für morgen ist.
const hoechstensJeSeite = 1000

// Modelle gibt, was bei Anthropic bereitsteht.
//
// **Was keine Fähigkeitsangabe hat, kommt als Befund und nicht als Modell** —
// dieselbe Regel wie bei einem Ollama-Modell ohne Digest (A7).
func (k *Klient) Modelle(ctx context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	var antwort modellantwort

	frage := url.Values{"limit": {strconv.Itoa(hoechstensJeSeite)}}

	if fehler := k.hole(ctx, Modellweg, frage, nil, &antwort); fehler != nil {
		return nil, nil, fehler
	}

	modelle := make([]domain.KIModell, 0, len(antwort.Daten))

	var befunde []domain.Befund

	for stelle, dto := range antwort.Daten {
		if dto.Koennen == nil {
			befunde = append(befunde, ohneAuskunft(dto, stelle))

			continue
		}

		// **Ein Modell ohne lesbaren Tag ist ein Befund und kein Eintrag.**
		//
		// # Warum das hier steht und nicht beim Katalog
		//
		// [service.NeuerModellkatalog] weist einen unvollständigen Eintrag
		// mit einem **Fehler** ab — und nimmt damit den ganzen Katalog mit,
		// auch die Modelle der anderen Anbieter. Das ist dieselbe Regel wie
		// bei fremden Paketdaten (§8): von 300 lesbaren Karten und zwei
		// unlesbaren werden 300 geöffnet und zwei benannt.
		if tagAus(dto.Angelegt) == "" {
			befunde = append(befunde, ohneTag(dto, stelle))

			continue
		}

		modelle = append(modelle, alsModell(dto))
	}

	// **Eine zweite Seite wird gemeldet und nicht geholt.**
	//
	// **Weil eine halbe Liste, die wie eine ganze aussieht, teurer ist als
	// eine Meldung.** Tausend Modelle gibt es heute nicht; käme der Tag,
	// stünde hier eine Schleife — und bis dahin sagt der Befund, dass sie
	// fehlt (§10).
	if antwort.Weitere {
		befunde = append(befunde, domain.Befund{
			// **Der Ort ist die Liste selbst.** Dieser Befund gilt keinem
			// Eintrag, und ein leerer Ort sähe aus wie einer, der seine
			// Stelle verloren hat (A19, #1550).
			Ort: Modellweg, Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf("Anthropic nennt mehr als %d Modelle; diese "+
				"Liste endet bei %q und ist nicht vollständig.",
				hoechstensJeSeite, antwort.Letzte),
		})
	}

	return modelle, befunde, nil
}

// alsModell macht aus einem DTO ein Modell mit belegten Fähigkeiten.
func alsModell(dto modellDTO) domain.KIModell {
	return domain.KIModell{
		Anbieter: Anbieter,
		Alias:    dto.Kennung,

		// **Die Laufzeit ist die Fassung der Schnittstelle.**
		//
		// Anders als bei Ollama gibt es hier keine Programmfassung, die man
		// lesen könnte — was feststeht, ist die Fassung, unter der gefragt
		// wurde. Sie gehört zum Beleg: derselbe Name antwortet unter zwei
		// Fassungen nicht zwingend gleich.
		Laufzeit: "anthropic-version " + Fassung,

		// **Cloud, ohne Einstufung nach Adresse.** Die Messages-API liegt
		// bei Anthropic; hier gibt es nichts zu entscheiden.
		Ort: domain.KIOrtCloud,

		Geprueft: belegeAus(dto),
	}
}

// belegeAus macht aus den Angaben der Schnittstelle belegte Fähigkeiten.
func belegeAus(dto modellDTO) []domain.KIGeprueft {
	tag := tagAus(dto.Angelegt)

	// **Text kann jedes Modell der Messages-API** — das ist die Schnittstelle
	// selbst und keine Angabe unter `capabilities`. Sie führt dort nichts
	// dazu, weil es nichts zu führen gibt.
	belege := []domain.KIGeprueft{{
		Faehigkeit: domain.KITextgenerierung, Datum: tag,
		Beleg: "Anthropic führt " + dto.Kennung + " unter " + Modellweg +
			"; die Messages-API erzeugt Text",
	}}

	for _, paar := range []struct {
		kann       koennen
		faehigkeit domain.KIFaehigkeit
		feld       string
	}{
		{dto.Koennen.FesteForm, domain.KIStrukturierteGenerierung, "structured_outputs"},
		{dto.Koennen.Bilder, domain.KIVision, "image_input"},
	} {
		if !paar.kann.Unterstuetzt {
			continue
		}

		belege = append(belege, domain.KIGeprueft{
			Faehigkeit: paar.faehigkeit, Datum: tag,
			Beleg: "Anthropic meldet " + paar.feld + " für " + dto.Kennung +
				" unter " + Modellweg,
		})
	}

	return belege
}

// tagAus macht aus `created_at` den Tag des Belegs.
//
// **Der Tag des Modells und nicht der des Laufs.** A6 verlangt am Beleg ein
// Datum; das des Laufs wäre bei jedem Ruf ein anderes und sagte nichts über
// die Sache. Was hier steht, ist der Tag, seit dem dieses Modell so ist.
//
// **Ohne lesbare Angabe bleibt es leer**, statt ein erfundenes Datum zu
// tragen — und [Klient.Modelle] macht daraus einen Befund, bevor der Eintrag
// entsteht.
func tagAus(roh string) string {
	wann, fehler := time.Parse(time.RFC3339, roh)
	if fehler != nil {
		return ""
	}

	return wann.Format(time.DateOnly)
}

// ohneTag baut den Befund über ein Modell ohne lesbares Datum.
//
// **Getrennt von [ohneAuskunft]**, weil es eine andere Auskunft ist: dort
// fehlt die Fähigkeit, hier der Tag, an dem sie festgestellt wurde. Wer
// beides in einen Satz legte, ließe den Lesenden raten, was zu tun ist.
//
// **Die Stelle steht dabei** — siehe [ohneAuskunft].
func ohneTag(dto modellDTO, stelle int) domain.Befund {
	return domain.Befund{
		Datei: dto.Kennung, Ort: stelleIn(stelle), Art: domain.ArtUmgebung,
		Meldung: fmt.Sprintf("Anthropic nennt für %q kein lesbares Datum "+
			"(created_at war %q). Ein Beleg ohne Tag sagt nicht, seit wann "+
			"er gilt — das Modell bleibt aus dem Katalog.",
			dto.Kennung, dto.Angelegt),
	}
}

// ohneAuskunft baut den Befund über ein Modell ohne Fähigkeitsangabe.
//
// # Warum die Stelle dabeisteht
//
// **Weil ein Eintrag ohne `id` sonst nicht zu benennen wäre** — Werkstück A19
// von Phase 5.3
// (#1550),
// gefunden vom Fuzzer bei der Saat `{"data":[{}]}`. Zwei solche Einträge
// gaben zwei **wortgleiche** Befunde, und §8 verlangt für jeden Fehler die
// Stelle. Der Eintrag fiel schon vorher aus dem Katalog; was fehlte, war der
// Weg zurück zu ihm.
func ohneAuskunft(dto modellDTO, stelle int) domain.Befund {
	return domain.Befund{
		Datei: dto.Kennung, Ort: stelleIn(stelle), Art: domain.ArtUmgebung,
		Meldung: fmt.Sprintf("Anthropic nennt für %q keine Fähigkeiten. Das "+
			"heißt nicht »kann nichts«, sondern »es steht nichts fest« — ein "+
			"Modell ohne Beleg gehört nicht in den Katalog.", dto.Kennung),
	}
}

// stelleIn benennt einen Eintrag der Modellliste über seine Stelle.
//
// **In der Ausdrucksweise des Formats** ([domain.Befund.Ort]): bei JSON ist
// das ein Pfad, und `data[2]` findet jemand in der rohen Antwort wieder.
func stelleIn(stelle int) string {
	return "data[" + strconv.Itoa(stelle) + "]"
}
