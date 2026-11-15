// **Ein Symbol der Bibliothek bearbeiten, über MCP** — DOM-02.AK-02
// (#1697).
//
// Einordnung nach §16: **erweiternd**. Die Teile haben dieselbe Gestalt wie
// bei den Rechnungen des Symboleditors (`lesen.symbol.*`): wer ein Symbol
// umformen will, liest die Teile, rechnet und sichert das Ergebnis.

package mcp

import (
	"context"
	"encoding/json"
)

// symbolkennung ist das Schema eines Rufs, der nur ein Symbol nennt.
func symbolkennung() map[string]any {
	return objekt(map[string]any{
		"id": zahl("Die Kennung des Symbols aus »lesen.symbolbibliothek«."),
	}, "id")
}

func lesenSymbolteile(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "lesen.symbolbibliothek.teile", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt die Zeichnung eines Symbols der Bibliothek als Teile im Feld »teile« — " +
			"dieselbe Gestalt, die die Rechnungen »lesen.symbol.*« nehmen und geben. " +
			"ÄNDERT NICHTS.",
		Schema: symbolkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			id, _, err := kennungUndName(roh)
			if err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolTeile(ctx, id))
		},
	}
}

func schreibenSymbolSichern(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.sichern", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt eine NEUE Fassung eines Symbols der Bibliothek ab; die " +
			"bisherige wird die vorige und lässt sich mit »schreiben.symbolbibliothek.zuruecknehmen« " +
			"zurückholen. Die Sets, die die bisherige tragen, stehen im Feld »veraltet« — sie " +
			"behalten ihre Kopie, bis »schreiben.symbolbibliothek.nachziehen« sie nachzieht.",
		Schema: objekt(map[string]any{
			"id":    zahl("Die Kennung des Symbols aus »lesen.symbolbibliothek«."),
			"teile": objektliste("Die Teile, wie »lesen.symbolbibliothek.teile« sie gibt.", map[string]any{"type": "object"}),
		}, "id", "teile"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				ID    int64           `json:"id"`
				Teile json.RawMessage `json:"teile"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolSichern(ctx, wahl.ID, wahl.Teile))
		},
	}
}

func schreibenSymbolZuruecknehmen(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.zuruecknehmen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: holt die vorige Fassung eines Symbols der Bibliothek zurück; " +
			"die jetzige wird die vorige. Ohne vorige Fassung wird abgewiesen. Die Sets, die die " +
			"ersetzte tragen, stehen im Feld »veraltet«.",
		Schema: symbolkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			id, _, err := kennungUndName(roh)
			if err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolZuruecknehmen(ctx, id))
		},
	}
}

func schreibenSymbolNachziehen(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.nachziehen", Art: Schreibend,
		Erklaerung: "ÄNDERT Sets: schreibt die jetzige Fassung eines Symbols in die Symboldatei jedes " +
			"Sets, das die vorige trägt — unter demselben Namen; neu gezeichnet werden genau deren " +
			"Karten. Das Feld »sets« nennt sie, auch wenn es unterwegs scheitert.",
		Schema: symbolkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			id, _, err := kennungUndName(roh)
			if err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolNachziehen(ctx, id))
		},
	}
}
