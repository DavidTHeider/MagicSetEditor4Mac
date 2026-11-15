// **Die eigenen Übersetzungen über MCP** — LOC-02
// (#1688).
//
// Einordnung nach §16: **erweiternd** — sieben Werkzeuge kommen hinzu, drei
// lesend und vier schreibend. **Eine Datei ist ein Text**: ein MCP-Server
// steht an keiner Platte, die das Modell sieht (Vermerk 0140).
//
// # Wofür ein Sprachmodell sie braucht
//
// **Zum Übersetzen.** »lesen.sprachpaket« nennt jeden Satz, der fehlt oder
// veraltet ist, »schreiben.sprachpaket.aendern« nimmt viele auf einmal —
// und prüft jeden gegen den deutschen Satz, ehe einer gilt.

package mcp

import (
	"context"
	"encoding/json"
)

// Sprachpakete ist, was MCP für die eigenen Übersetzungen braucht.
type Sprachpakete interface {
	Sprachpakete(ctx context.Context) (any, error)
	Sprachpaket(ctx context.Context, kennung string) (any, error)
	SprachpaketAnlegen(ctx context.Context, kennung, name, basis string) (any, error)
	SprachpaketAendern(ctx context.Context, kennung string, angaben, texte map[string]string) (any, error)
	SprachpaketLoeschen(ctx context.Context, kennung string) (any, error)
	SprachpaketAusgabe(ctx context.Context, kennung string) (any, error)
	SprachpaketEinlesen(ctx context.Context, inhalt string, ersetzen bool) (any, error)
}

// sprachwerkzeuge gibt die lesenden und die schreibenden.
func sprachwerkzeuge(naht Sprachpakete) (lesend, schreibend []Werkzeug) {
	if naht == nil {
		return nil, nil
	}

	return []Werkzeug{lesenSprachpakete(naht), lesenSprachpaket(naht), lesenSprachpaketdatei(naht)},
		[]Werkzeug{
			schreibenSprachpaketAnlegen(naht), schreibenSprachpaketAendern(naht),
			schreibenSprachpaketLoeschen(naht), schreibenSprachpaketEinlesen(naht),
		}
}

// sprachkennung ist das Schema eines Rufs, der nur ein Paket nennt.
func sprachkennung() map[string]any {
	return objekt(map[string]any{
		"kennung": text("Die Kennung der Übersetzung nach BCP 47, etwa pt-BR."),
	}, "kennung")
}

// mitSprachkennung ruft einen Weg mit der Kennung aus den Argumenten.
func mitSprachkennung(weg func(ctx context.Context, kennung string) (any, error)) func(
	context.Context, json.RawMessage,
) (string, error) {
	return func(ctx context.Context, roh json.RawMessage) (string, error) {
		var wahl struct {
			Kennung string `json:"kennung"`
		}

		if err := lies(roh, &wahl); err != nil {
			return "", err
		}

		ergebnis, err := weg(ctx, wahl.Kennung)
		if err != nil {
			return "", err
		}

		return alsJSON(ergebnis)
	}
}

func lesenSprachpakete(naht Sprachpakete) Werkzeug {
	return Werkzeug{
		Name: "lesen.sprachpakete", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die eigenen Übersetzungen dieses Rechners — je mit Kennung, Name, " +
			"Basissprache, Stand, Lizenz und wie viele Sätze übersetzt, fehlend und veraltet " +
			"sind. Ohne die Sätze selbst. ÄNDERT NICHTS.",
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			ergebnis, err := naht.Sprachpakete(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(ergebnis)
		},
	}
}

func lesenSprachpaket(naht Sprachpakete) Werkzeug {
	return Werkzeug{
		Name: "lesen.sprachpaket", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt eine eigene Übersetzung mit jedem übersetzten Satz (samt dem deutschen, " +
			"gegen den er übersetzt wurde) und den Pfaden, die fehlen, veraltet oder unbekannt " +
			"sind. Den deutschen Satz eines fehlenden Pfads nennt die Übersetzung nicht; " +
			"übersetze gegen die Basissprache. ÄNDERT NICHTS.",
		Schema: sprachkennung(),
		Ruf:    mitSprachkennung(naht.Sprachpaket),
	}
}

func lesenSprachpaketdatei(naht Sprachpakete) Werkzeug {
	return Werkzeug{
		Name: "lesen.sprachpaket.datei", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt eine eigene Übersetzung als Datei zum Weitergeben — mit Herkunft, " +
			"Lizenz und Stand, ohne Einstellungen und ohne Geheimnisse. Ohne Lizenz verweigert. " +
			"ÄNDERT NICHTS.",
		Schema: sprachkennung(),
		Ruf:    mitSprachkennung(naht.SprachpaketAusgabe),
	}
}

func schreibenSprachpaketAnlegen(naht Sprachpakete) Werkzeug {
	return Werkzeug{
		Name: "schreiben.sprachpaket.anlegen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt eine eigene Übersetzung über einer Basissprache an — " +
			"als Entwurf, ohne einen Satz; was fehlt, zeigt die Basis. Eine eingebaute Sprache " +
			"(de, en, es, fr, it, ja, ru, zh) ist keine eigene.",
		Schema: objekt(map[string]any{
			"kennung": text("Die Kennung nach BCP 47, etwa pt-BR, zh-Hant oder ja-x-meine."),
			"name":    text("Der Name in der eigenen Sprache, etwa Português do Brasil."),
			"basis":   text("Die Fenstersprache, die zeigt, was noch fehlt: de, en, es, fr, it, ja, ru oder zh."),
		}, "kennung", "name", "basis"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Kennung string `json:"kennung"`
				Name    string `json:"name"`
				Basis   string `json:"basis"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			ergebnis, err := naht.SprachpaketAnlegen(ctx, wahl.Kennung, wahl.Name, wahl.Basis)
			if err != nil {
				return "", err
			}

			return alsJSON(ergebnis)
		},
	}
}

func schreibenSprachpaketAendern(naht Sprachpakete) Werkzeug {
	return Werkzeug{
		Name: "schreiben.sprachpaket.aendern", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: setzt Angaben und Sätze einer eigenen Übersetzung — ganz oder " +
			"gar nicht. Jeder Satz muss dieselben Platzhalter wie sein deutscher tragen ({name}), " +
			"darf **fett** und `Code` nutzen, aber kein HTML. Ein leerer Satz nimmt die " +
			"Übersetzung zurück. Weggelassene Angaben bleiben.",
		Schema: objekt(map[string]any{
			"kennung":  text("Die Kennung der Übersetzung."),
			"name":     text("Ein neuer Name — weglassen, wenn er bleibt."),
			"stand":    text("entwurf, maschinell, geprueft oder abgenommen — weglassen, wenn er bleibt."),
			"lizenz":   text("Die Lizenz, unter der die Übersetzung steht, etwa CC0-1.0."),
			"herkunft": text("Wer übersetzt hat."),
			"texte": map[string]any{
				"type": "object", "additionalProperties": map[string]any{"type": "string"},
				"description": "Pfad → Satz, die Pfade aus »lesen.sprachpaket«.",
			},
		}, "kennung"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Kennung  string            `json:"kennung"`
				Name     *string           `json:"name"`
				Stand    *string           `json:"stand"`
				Lizenz   *string           `json:"lizenz"`
				Herkunft *string           `json:"herkunft"`
				Texte    map[string]string `json:"texte"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			angaben := map[string]string{}
			for schluessel, wert := range map[string]*string{
				"name": wahl.Name, "stand": wahl.Stand, "lizenz": wahl.Lizenz, "herkunft": wahl.Herkunft,
			} {
				if wert != nil {
					angaben[schluessel] = *wert
				}
			}

			ergebnis, err := naht.SprachpaketAendern(ctx, wahl.Kennung, angaben, wahl.Texte)
			if err != nil {
				return "", err
			}

			return alsJSON(ergebnis)
		},
	}
}

func schreibenSprachpaketLoeschen(naht Sprachpakete) Werkzeug {
	return Werkzeug{
		Name: "schreiben.sprachpaket.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: entfernt eine eigene Übersetzung. Wer sie benutzt hat, " +
			"sieht danach ihre Basissprache.",
		Schema: sprachkennung(),
		Ruf:    mitSprachkennung(naht.SprachpaketLoeschen),
	}
}

func schreibenSprachpaketEinlesen(naht Sprachpakete) Werkzeug {
	return Werkzeug{
		Name: "schreiben.sprachpaket.einlesen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt eine Übersetzung aus dem Text einer Datei auf, wie " +
			"»lesen.sprachpaket.datei« sie gibt. Eine vorhandene nur mit ersetzen.",
		Schema: objekt(map[string]any{
			"inhalt":   text("Der Text der Datei."),
			"ersetzen": jaNein("Ersetzt eine Übersetzung mit derselben Kennung."),
		}, "inhalt"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Inhalt   string `json:"inhalt"`
				Ersetzen bool   `json:"ersetzen"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			ergebnis, err := naht.SprachpaketEinlesen(ctx, wahl.Inhalt, wahl.Ersetzen)
			if err != nil {
				return "", err
			}

			return alsJSON(ergebnis)
		},
	}
}
