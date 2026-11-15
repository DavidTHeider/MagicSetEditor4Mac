package service

// **Eine Fähigkeit messen, statt sie zu glauben** — Werkstück A9 von
// Phase 5.3
// (#1448).
//
// # Warum es diesen Weg gibt
//
// **Weil OpenAI seine Modelle nicht beschreibt.** `/v1/models` nennt keine
// Fähigkeiten
// ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)), und
// [0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)
// sagt, was dann gilt: ohne Beleg gehört ein Modell nicht in den Katalog.
//
// Ohne einen dritten Weg wäre der Anbieter ein Schalter, der nichts tut. Den
// dritten Weg hat A7 schon gebaut, nur für etwas anderes: den
// Einbettungsvertrag **misst** es, statt ihn zu raten. Hier ist es dasselbe.
//
// # Was eine Probe prüft — und was nicht
//
// **Sie prüft, dass etwas Schemakonformes zurückkommt.** Das belegt zweierlei:
// das Modell erzeugt Text, und es hält ein Schema ein. Mehr belegt sie nicht —
// **nicht**, dass es Bilder liest, nicht, dass es einbettet, und nicht, dass
// es morgen dasselbe tut.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KIProbequelle ist, was das Prüfwerk von einem Anbieter braucht.
//
// **Dieselben zwei Methoden wie beim Katalog**, und mit Absicht: wer eine
// dritte verlangte, machte aus der Probe etwas, das nur ein Teil der
// Anbieter kann.
type KIProbequelle interface {
	// Fassung nennt, womit geantwortet wird.
	Fassung(ctx context.Context) (string, error)

	// Vorschlagen fragt mit festem Schema.
	Vorschlagen(ctx context.Context, alias, auftrag string, schema map[string]any) (
		KIProbeantwort, error)
}

// KIProbeantwort ist, was dabei zurückkommt.
type KIProbeantwort struct {
	Inhalt        string
	Abgeschnitten bool
	Modell        string
}

// KIProbenablage merkt sich, was gemessen wurde.
type KIProbenablage interface {
	Merken(ctx context.Context, profilID int64, probe domain.KIModellprobe) error
}

// dasProbewort ist, worum die Probe bittet.
//
// **Ein Wort und kein Satz.** Was gemessen wird, ist die **Form**, nicht der
// Inhalt; ein langer Auftrag kostete mehr und sagte nichts mehr.
const dasProbewort = "bereit"

// DasProbeschema ist das Schema, gegen das gemessen wird.
//
// **Bewusst winzig**: ein Feld, eine Zeichenkette, nichts Zusätzliches. Ein
// Modell, das daran scheitert, scheitert auch an einem Kartenschema — und
// eines, das es besteht, hat die Form eingehalten, um die es geht.
func DasProbeschema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"probe": map[string]any{"type": "string"},
		},
		"required":             []string{"probe"},
		"additionalProperties": false,
	}
}

// DerProbeauftrag ist, was hinausgeht.
func DerProbeauftrag() string {
	return fmt.Sprintf(
		"Antworte ausschließlich mit dem JSON-Objekt {\"probe\": %q}.",
		dasProbewort)
}

// KIModellpruefwerk misst die Fähigkeit eines Modells.
type KIModellpruefwerk struct {
	ablage KIProbenablage
	uhr    func() time.Time
}

// NeuesKIModellpruefwerk baut es.
//
// **Die Uhr kommt von außen** (§3): ein Lauf, der den heutigen Tag prüfen
// will, braucht einen, den er kennt.
func NeuesKIModellpruefwerk(
	ablage KIProbenablage, uhr func() time.Time,
) *KIModellpruefwerk {
	return &KIModellpruefwerk{ablage: ablage, uhr: uhr}
}

// Pruefen misst ein Modell und merkt sich das Ergebnis.
//
// **Es gibt die Messungen zurück, die es geschrieben hat** — zwei, wenn es
// geklappt hat: Text und feste Form. Eine gescheiterte Probe schreibt
// **nichts** und sagt, woran es lag.
func (w *KIModellpruefwerk) Pruefen(
	ctx context.Context, profilID int64, quelle KIProbequelle, anbieter, alias string,
) ([]domain.KIModellprobe, error) {
	if strings.TrimSpace(anbieter) == "" || strings.TrimSpace(alias) == "" {
		return nil, fmt.Errorf(
			"%w: eine Messung nennt Anbieter und Modell", domain.ErrValidation)
	}

	fassung, fehler := quelle.Fassung(ctx)
	if fehler != nil {
		return nil, fehler
	}

	antwort, fehler := quelle.Vorschlagen(ctx, alias, DerProbeauftrag(),
		DasProbeschema())
	if fehler != nil {
		return nil, fehler
	}

	if fehler := haeltDieForm(antwort); fehler != nil {
		return nil, fehler
	}

	return w.merke(ctx, profilID, probenAus(
		anbieter, alias, fassung, antwort, w.uhr().Format(time.DateOnly)))
}

// haeltDieForm prüft, ob die Antwort das Schema einhält.
//
// **Eine abgeschnittene Antwort zählt nicht**, auch wenn das JSON darin
// zufällig vollständig wäre: was gemessen wird, ist, dass das Modell **fertig
// wird**, und eine halbe Antwort sieht aus wie eine ganze (A7).
func haeltDieForm(antwort KIProbeantwort) error {
	if antwort.Abgeschnitten {
		return fmt.Errorf(
			"%w: die Antwort war abgeschnitten — damit ist nichts belegt",
			domain.ErrValidation)
	}

	var gelesen struct {
		Probe *string `json:"probe"`
	}

	if fehler := json.Unmarshal([]byte(antwort.Inhalt), &gelesen); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Antwort auf die Probe ist kein JSON")
	}

	if gelesen.Probe == nil {
		return fmt.Errorf(
			"%w: die Antwort ist JSON, hält aber das Schema nicht ein — "+
				"das Feld »probe« fehlt", domain.ErrValidation)
	}

	return nil
}

// probenAus macht aus einer geglückten Messung ihre Belege.
//
// **Zwei und nicht einer.** Eine schemakonforme Antwort belegt beides: dass
// das Modell Text erzeugt und dass es eine Form einhält. Wer nur das zweite
// führte, ließe ein Modell ungeprüft, das gerade eben geantwortet hat.
func probenAus(
	anbieter, alias, fassung string, antwort KIProbeantwort, tag string,
) []domain.KIModellprobe {
	// **Das antwortende Modell steht daneben** — leer heißt: der Anbieter
	// hat keines genannt, und dann ist der Alias das Beste, was feststeht.
	modell := antwort.Modell
	if strings.TrimSpace(modell) == "" {
		modell = alias
	}

	beleg := fmt.Sprintf(
		"gemessen an %s: ein Probeauftrag mit festem Schema kam schemakonform "+
			"zurück (%s)", tag, fassung)

	proben := make([]domain.KIModellprobe, 0, 2)

	for _, faehigkeit := range []domain.KIFaehigkeit{
		domain.KITextgenerierung, domain.KIStrukturierteGenerierung,
	} {
		proben = append(proben, domain.KIModellprobe{
			Anbieter: anbieter, Alias: alias, Faehigkeit: faehigkeit,
			Datum: tag, Beleg: beleg, Laufzeit: fassung, Modell: modell,
		})
	}

	return proben
}

// merke schreibt die Messungen.
func (w *KIModellpruefwerk) merke(
	ctx context.Context, profilID int64, proben []domain.KIModellprobe,
) ([]domain.KIModellprobe, error) {
	for _, probe := range proben {
		if fehler := w.ablage.Merken(ctx, profilID, probe); fehler != nil {
			return nil, fehler
		}
	}

	return proben, nil
}
