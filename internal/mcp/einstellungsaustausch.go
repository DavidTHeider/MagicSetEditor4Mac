// **Einstellungen ausführen und einlesen, über MCP** — Werkstück O3 von
// Phase 5.2 (#1427).
//
// # Warum die Vorschau ein eigenes Werkzeug ist und kein Schalter
//
// **Weil ein Schalter am schreibenden Werkzeug eine Höflichkeit wäre.** Ein
// Modell, das „übernehmen: nein" setzen kann, kann es auch weglassen; ein
// **lesendes** Werkzeug kann gar nicht schreiben — und ein Gegenüber, das die
// Liste auf die lesenden kürzt, behält es und verliert die Übernahme (0044).
//
// # Und warum die Übernahme eine Revision verlangt
//
// **Weil ein Modell schnell ist.** Zwischen „zeig mir den Unterschied" und
// „übernimm" liegt hier kein Mensch, der zögert; die Marke ist die einzige
// Stelle, an der auffällt, dass inzwischen jemand anderes geschrieben hat
// (§7, Regel 2).

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Austauschquelle ist die Naht zu `internal/app`.
type Austauschquelle interface {
	EinstellungenAusfuehren(ctx context.Context, abschnitte []string) (string, error)
	EinstellungenVorschau(ctx context.Context, roh []byte) (Einstellungssicht, error)
	EinstellungenUebernehmen(
		ctx context.Context, roh []byte, revision string, nur []string,
	) (Einstellungsbericht, error)
}

// Einstellungsdiffzeile ist eine Zeile des Unterschieds.
type Einstellungsdiffzeile struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Vorher     string `json:"vorher"`
	Nachher    string `json:"nachher"`
	Herkunft   string `json:"herkunft"`
}

// Einstellungsbefund ist, was an einer Zeile nicht stimmt.
type Einstellungsbefund struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Grund      string `json:"grund"`
}

// Einstellungssicht ist, was eine Datei ändern würde.
type Einstellungssicht struct {
	Fassung      int                     `json:"fassung"`
	Revision     string                  `json:"revision"`
	Zeilen       []Einstellungsdiffzeile `json:"zeilen"`
	Befunde      []Einstellungsbefund    `json:"befunde"`
	Unveraendert int                     `json:"unveraendert"`
}

// Einstellungsbericht ist, was eine Übernahme bewirkt hat.
type Einstellungsbericht struct {
	Uebernommen int                     `json:"uebernommen"`
	Revision    string                  `json:"revision"`
	Zeilen      []Einstellungsdiffzeile `json:"zeilen"`
	Befunde     []Einstellungsbefund    `json:"befunde"`
}

// austauschwerkzeuge gibt die zwei lesenden und das schreibende.
func austauschwerkzeuge(quelle Austauschquelle) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{einstellungenAusfuehren(quelle), einfuhrVorschau(quelle)},
		[]Werkzeug{einfuhrUebernehmen(quelle)}
}

// einstellungenAusfuehren gibt die Einstellungen als Austauschtext.
func einstellungenAusfuehren(quelle Austauschquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.einstellungen.ausfuhr", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt die gewählten Einstellungen als versionierten JSON-Text. " +
			"Ohne »abschnitte« kommen »nutzer« und »profil«; »arbeitsplatz« nur auf Wunsch. " +
			"Geheimnisse, Pfade und lokale Kennungen sind nicht darin.",
		Schema: objekt(map[string]any{
			"abschnitte": objektliste(
				"Die gewünschten Abschnitte — leer heißt »nutzer« und »profil«.",
				text("»nutzer«, »profil« oder »arbeitsplatz«.")),
		}),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				Abschnitte []string `json:"abschnitte"`
			}

			if fehler := argumenteLesen(argumente, &wahl); fehler != nil {
				return "", fehler
			}

			inhalt, fehler := quelle.EinstellungenAusfuehren(ctx, wahl.Abschnitte)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"inhalt": inhalt})
		},
	}
}

// einfuhrVorschau sagt, was ein Text ändern würde.
func einfuhrVorschau(quelle Austauschquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.einstellungen.vorschau", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, was ein Austauschtext ändern würde — ohne etwas zu ändern. " +
			"Die Antwort trägt »revision«; genau die verlangt »schreiben.einstellungen.einfuhr«.",
		Schema: objekt(map[string]any{
			"inhalt": text("Der JSON-Text aus »lesen.einstellungen.ausfuhr«."),
		}, "inhalt"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				Inhalt string `json:"inhalt"`
			}

			if fehler := argumenteLesen(argumente, &wahl); fehler != nil {
				return "", fehler
			}

			sicht, fehler := quelle.EinstellungenVorschau(ctx, []byte(wahl.Inhalt))
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(sicht)
		},
	}
}

// einfuhrUebernehmen schreibt die gewählten Zeilen.
func einfuhrUebernehmen(quelle Austauschquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.einstellungen.einfuhr", Art: Schreibend,
		Erklaerung: "ÄNDERT die Einstellungen: übernimmt die Zeilen eines Austauschtextes. " +
			"Verlangt die »revision« aus »lesen.einstellungen.vorschau« — stimmt sie nicht mehr, " +
			"bleibt alles stehen. Mit »nur« lassen sich einzelne Zeilen wählen (»ebene:schluessel«).",
		Schema: objekt(map[string]any{
			"inhalt":   text("Der JSON-Text, der übernommen werden soll."),
			"revision": text("Die Marke aus der Vorschau."),
			"nur": objektliste("Die gewünschten Zeilen — leer heißt alle.",
				text("Eine Zeile als »ebene:schluessel«.")),
		}, "inhalt", "revision"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				Inhalt   string   `json:"inhalt"`
				Revision string   `json:"revision"`
				Nur      []string `json:"nur"`
			}

			if fehler := argumenteLesen(argumente, &wahl); fehler != nil {
				return "", fehler
			}

			bericht, fehler := quelle.EinstellungenUebernehmen(
				ctx, []byte(wahl.Inhalt), wahl.Revision, wahl.Nur)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(bericht)
		},
	}
}

// argumenteLesen entziffert die Argumente eines Werkzeugs.
//
// **Eine Stelle für drei Werkzeuge**: dreimal derselbe Satz über dieselbe
// Klage wäre dreimal die Gelegenheit, ihn anders zu formulieren.
func argumenteLesen(argumente json.RawMessage, ziel any) error {
	if len(argumente) == 0 {
		return nil
	}

	if err := json.Unmarshal(argumente, ziel); err != nil {
		return fmt.Errorf("die Argumente sind kein JSON: %w", err)
	}

	return nil
}
