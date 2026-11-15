package xai

// **Die Unterhaltung** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Warum `/v1/chat/completions` und nicht `/v1/responses`
//
// **Beide gibt es.** Die Entscheidung fiel am 2026-09-12 aus drei Gründen
// ([0101](../../docs/entscheidungen/0101-was-xai-belegt.md)):
//
//   - **`/v1/responses` hält Antworten 30 Tage.** „New responses will be
//     stored for 30 days and then permanently deleted." Eine Kartenanfrage,
//     die dort liegen bleibt, ist das Gegenteil dessen, was die Hilfe zusagt.
//   - **Die feste Form ist nur an `/v1/chat/completions` dokumentiert** —
//     `response_format.json_schema` mit `name`, `schema` und `strict`.
//   - **Die Ausgänge sind dort dokumentiert**: `stop`, `length`, `end_turn`
//     oder `null`.
//
// # Der Ausgang kennt ein Wort, das es sonst nicht gibt
//
// **`end_turn`.** Anthropic kennt es, OpenAI nicht, und xAI führt es neben
// `stop`. Wer die Liste von einem anderen Anbieter übernähme, hielte eine
// fertige Antwort für einen unbekannten Ausgang.

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// hoechstensZeichen ist die Obergrenze der Ausgabe.
const hoechstensZeichen = 4096

// Bild ist ein Bild, das mitgeschickt wird.
type Bild struct {
	// Medientyp ist `image/jpeg` oder `image/png`.
	//
	// **Zwei und nicht vier.** Nachgelesen am 2026-09-12: „jpg/jpeg or png".
	// Webp und GIF nimmt Anthropic, nicht xAI.
	Medientyp string

	// Daten sind die Bytes — base64 kodiert.
	Daten string
}

// Frage ist, was ein Modell beantworten soll.
type Frage struct {
	Alias   string
	System  string
	Auftrag string

	// Bilder reisen **vor** dem Text mit.
	Bilder []Bild

	// Schema ist das JSON-Schema, in das die Antwort gezwungen wird.
	Schema map[string]any
}

// Ausgang ist, wie eine Antwort geendet hat.
type Ausgang string

// Die vier Ausgänge, und zwei davon sind kein Ergebnis.
const (
	AusgangFertig         Ausgang = "fertig"
	AusgangUnvollstaendig Ausgang = "unvollstaendig"
	AusgangAbgelehnt      Ausgang = "abgelehnt"
	AusgangUnbekannt      Ausgang = "unbekannt"
)

// Antwort ist, was zurückkam.
type Antwort struct {
	Inhalt   string
	Ausgang  Ausgang
	Grund    string
	Naeheres string
	Modell   string

	Verbrauch Verbrauch
}

// Verbrauch ist, was ein Ruf gekostet hat.
//
// # Er reist in der Antwort und liegt nicht am Klienten
//
// **Weil ein Klient mehrere Rufe trägt.** Ein Feld „was der letzte gekostet
// hat" wäre veränderlicher Zustand an einem Ding, das sich teilen lässt — und
// wer ihn läse, läse ihn nach dem falschen Ruf. Der Preis gehört zu **der**
// Antwort, aus der er stammt (§7, Regel 2).
//
// **Und wer ihn heute ausgibt, ist niemand.** Er kommt bis zur Naht und
// endet dort: [domain.KIAuftrag] führt ein Budget in Zeichen, und ein zweites
// in Geld daneben wäre eine zweite Buchführung. Was daraus wird, entscheidet
// A13 (#1452).
//
// # Und hier steht es **in Geld**
//
// **Das kann sonst keiner.** Anthropic und OpenAI nennen Zeichen; wer daraus
// Geld machen will, braucht einen Tarif und rechnet ihn nach. xAI nennt den
// Preis dieses einen Rufs — `usage.cost_in_usd_ticks`, und ein Tick ist
// definiert: „TICKS_IN_USD_CENT: i64 = 100_000_000 which means there is
// 10'000'000'000 ticks in one dollar."
type Verbrauch struct {
	Eingabe  int `json:"prompt_tokens"`
	Ausgabe  int `json:"completion_tokens"`
	Zusammen int `json:"total_tokens"`

	// Kosten sind der Preis dieses Rufs in **Ticks**.
	//
	// **Als Tick und nicht als Betrag**: ein Zehnmilliardstel Dollar ist
	// kleiner als jede Währungseinheit, und wer hier rundete, verlöre genau
	// das, wofür die Zahl da ist.
	Kosten int64 `json:"cost_in_usd_ticks"`
}

// TicksJeDollar ist, was die Dokumentation dazu sagt.
const TicksJeDollar = 10_000_000_000

// antwortrumpf ist die Gestalt einer Antwort.
type antwortrumpf struct {
	Modell string `json:"model"`

	Wahlen []struct {
		Nachricht struct {
			Inhalt    string `json:"content"`
			Ablehnung string `json:"refusal"`
		} `json:"message"`

		Grund string `json:"finish_reason"`
	} `json:"choices"`

	Verbrauch Verbrauch `json:"usage"`
}

// dieFertigenGruende sind die Ausgänge, nach denen eine Antwort gilt.
var dieFertigenGruende = []string{"stop", "end_turn"}

// dieAbbruchgruende sind die, nach denen sie abgeschnitten ist.
var dieAbbruchgruende = []string{"length"}

// Rede fragt ein Modell.
func (k *Klient) Rede(ctx context.Context, frage Frage) (Antwort, error) {
	if fehler := frage.pruefe(); fehler != nil {
		return Antwort{}, fehler
	}

	var roh antwortrumpf

	if fehler := k.hole(ctx, Unterhaltungsweg, frage.alsAnfrage(), &roh); fehler != nil {
		return Antwort{}, fehler
	}

	// **Keine Wahl ist kein leeres Ergebnis, sondern ein Fehler.** Eine
	// Antwort ohne `choices` hat nichts zu sagen, und ein leerer Inhalt sähe
	// aus wie ein Modell, das nichts gefunden hat.
	if len(roh.Wahlen) == 0 {
		return Antwort{}, fmt.Errorf(
			"%w: xAI hat keine Antwort geschickt (%s)", domain.ErrConflict,
			roh.Modell)
	}

	wahl := roh.Wahlen[0]

	return Antwort{
		Inhalt: wahl.Nachricht.Inhalt, Grund: wahl.Grund,
		Ausgang:  ausgangFuer(wahl.Grund, wahl.Nachricht.Ablehnung),
		Naeheres: wahl.Nachricht.Ablehnung,
		Modell:   roh.Modell, Verbrauch: roh.Verbrauch,
	}, nil
}

// ausgangFuer ordnet Grund und Ablehnung einem der vier Ausgänge zu.
//
// **Die Ablehnung zuerst**, denn sie steht neben `finish_reason: stop`: wer
// den Grund zuerst läse, hielte sie für ein Ergebnis.
func ausgangFuer(grund, ablehnung string) Ausgang {
	if strings.TrimSpace(ablehnung) != "" {
		return AusgangAbgelehnt
	}

	switch {
	case slices.Contains(dieFertigenGruende, grund):
		return AusgangFertig
	case slices.Contains(dieAbbruchgruende, grund):
		return AusgangUnvollstaendig
	}

	return AusgangUnbekannt
}
