package mcp

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135.
//
// # Vier lesend, drei schreibend — und drei gar nicht
//
// **Umbenennen, Zusammenführen und überall Entfernen gibt es hier nicht.**
// Sie ändern jedes Objekt, das die Marke trägt, auf einmal; ein Werkzeug
// schreibt nur, was es nicht überschreibt (0044). Eine Marke an ein Objekt
// zu hängen oder von ihm zu nehmen lässt sich zurücknehmen — das geht.

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Objektmarken sind die Marken jeder Art, wie das Modell sie ruft.
type Objektmarken interface {
	ObjektmarkeSetzen(ctx context.Context, art, kennung, marke string) error
	ObjektmarkeEntfernen(ctx context.Context, art, kennung, marke string) error
	ObjektmarkenAn(ctx context.Context, art, kennung string) ([]string, error)
	MarkenJeObjekt(ctx context.Context, art string) (map[string][]string, error)
	MarkeVergeben(ctx context.Context, art string, kennungen []string, marke string) error
	MarkenFiltern(ctx context.Context, art string, marken []string, verknuepfung string) ([]string, error)
	Markenkatalog(ctx context.Context) ([]domain.Markenposten, error)
}

// artenerklaerung sagt, welche Arten eine Marke tragen.
const artenerklaerung = "Die Art des Objekts: karte, set, exemplar (ein Posten der Sammlung), " +
	"deck, paket (als art/name, etwa game/magic), turnier, partie oder handel."

// objektmarkenArgumente sind die Argumente aller sieben Werkzeuge.
type objektmarkenArgumente struct {
	Art          string   `json:"art"`
	Kennung      string   `json:"kennung"`
	Kennungen    []string `json:"kennungen"`
	Marke        string   `json:"marke"`
	Marken       []string `json:"marken"`
	Verknuepfung string   `json:"verknuepfung"`
}

// objektmarkenwerkzeuge gibt die sieben — oder keines ohne Naht.
func objektmarkenwerkzeuge(marken Objektmarken) (lesend, schreibend []Werkzeug) {
	if marken == nil {
		return nil, nil
	}

	stelle := objekt(map[string]any{
		"art": text(artenerklaerung), "kennung": text("Die Kennung des Objekts."),
		"marke": text("Das Wort."),
	}, "art", "kennung", "marke")

	return []Werkzeug{
		{
			Name: "lesen.objektmarken", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die Marken jedes Objekts einer Art — nach Kennung. ÄNDERT NICHTS. " +
				"Eine Marke gehört dem Profil; ein anderes Profil sieht sie nicht.",
			Schema: objekt(map[string]any{"art": text(artenerklaerung)}, "art"),
			Ruf: mitArgumenten(func(ctx context.Context, a objektmarkenArgumente) (any, error) {
				je, fehler := marken.MarkenJeObjekt(ctx, a.Art)

				return map[string]any{"marken": je}, fehler
			}),
		},
		{
			Name: "lesen.objektmarke", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die Marken eines Objekts. ÄNDERT NICHTS.",
			Schema: objekt(map[string]any{
				"art": text(artenerklaerung), "kennung": text("Die Kennung des Objekts."),
			}, "art", "kennung"),
			Ruf: mitArgumenten(func(ctx context.Context, a objektmarkenArgumente) (any, error) {
				an, fehler := marken.ObjektmarkenAn(ctx, a.Art, a.Kennung)

				return map[string]any{"marken": listeOderLeer(an)}, fehler
			}),
		},
		{
			Name: "lesen.markenfilter", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die Objekte einer Art, die die Marken tragen. ÄNDERT NICHTS. " +
				"»und« heißt jede genannte Marke, »oder« wenigstens eine; ohne Angabe gilt »und«.",
			Schema: objekt(map[string]any{
				"art":          text(artenerklaerung),
				"marken":       textliste("Die Marken, nach denen gefiltert wird.", "Eine Marke."),
				"verknuepfung": text("»und« oder »oder«."),
			}, "art", "marken"),
			Ruf: mitArgumenten(func(ctx context.Context, a objektmarkenArgumente) (any, error) {
				kennungen, fehler := marken.MarkenFiltern(ctx, a.Art, a.Marken, a.Verknuepfung)

				return map[string]any{"kennungen": listeOderLeer(kennungen)}, fehler
			}),
		},
		{
			Name: "lesen.markenkatalog", Art: Lesend, Idempotent: true,
			Erklaerung: "Zählt die Marken des Profils über alle Arten. ÄNDERT NICHTS.",
			Schema:     objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				katalog, fehler := marken.Markenkatalog(ctx)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"marken": katalog})
			},
		},
	}, []Werkzeug{
		{
			Name: "schreiben.objektmarke", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: hängt eine Marke an ein Objekt. Sie wird getrimmt " +
				"und kleingeschrieben; zweimal dieselbe ist kein Fehler. Das Objekt bleibt " +
				"unberührt — auch ein Paket, an dem sie nur eine Anmerkung des Profils ist.",
			Schema: stelle,
			Ruf: mitArgumenten(func(ctx context.Context, a objektmarkenArgumente) (any, error) {
				return map[string]any{"marke": domain.Markennormalform(a.Marke), "gesetzt": true},
					marken.ObjektmarkeSetzen(ctx, a.Art, a.Kennung, a.Marke)
			}),
		},
		{
			Name: "schreiben.objektmarkeentfernen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: nimmt eine Marke von einem Objekt.",
			Schema:     stelle,
			Ruf: mitArgumenten(func(ctx context.Context, a objektmarkenArgumente) (any, error) {
				return map[string]any{"marke": domain.Markennormalform(a.Marke), "entfernt": true},
					marken.ObjektmarkeEntfernen(ctx, a.Art, a.Kennung, a.Marke)
			}),
		},
		{
			Name: "schreiben.markevergeben", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: hängt eine Marke an mehrere Objekte einer Art — " +
				"ganz oder gar nicht: gibt es eines nicht, bekommt keines sie.",
			Schema: objekt(map[string]any{
				"art":       text(artenerklaerung),
				"kennungen": textliste("Die Objekte.", "Eine Kennung."),
				"marke":     text("Das Wort."),
			}, "art", "kennungen", "marke"),
			Ruf: mitArgumenten(func(ctx context.Context, a objektmarkenArgumente) (any, error) {
				return map[string]any{"marke": domain.Markennormalform(a.Marke), "vergeben": len(a.Kennungen)},
					marken.MarkeVergeben(ctx, a.Art, a.Kennungen, a.Marke)
			}),
		},
	}
}

// mitArgumenten liest die Argumente, ruft und schreibt die Antwort — oder
// den Fehler.
func mitArgumenten(tun func(context.Context, objektmarkenArgumente) (any, error)) func(context.Context, json.RawMessage) (string, error) {
	return func(ctx context.Context, roh json.RawMessage) (string, error) {
		var argumente objektmarkenArgumente
		if fehler := lies(roh, &argumente); fehler != nil {
			return "", fehler
		}

		antwort, fehler := tun(ctx, argumente)
		if fehler != nil {
			return "", fehler
		}

		return alsJSON(antwort)
	}
}
