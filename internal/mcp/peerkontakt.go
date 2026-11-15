package mcp

// **Die bekannten Gegenüber als Werkzeuge** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum das Lesen hier steht und das Freigeben nicht
//
// **Weil eine Freigabe eine Vertrauensentscheidung ist**
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
// Ein Modell, das einen Zweck freigäbe, entschiede, wem dieser Rechner Karten
// überlässt — und der Mensch, der zugesehen hat, hielte es für eine
// Zusammenfassung.
//
// **Das Sperren steht trotzdem hier.** Es ist die Richtung, in die ein
// Irrtum nichts kostet: wer zu Unrecht gesperrt wird, wird von Hand wieder
// entsperrt; wer zu Unrecht freigegeben wird, hat schon gehandelt. Ein
// Werkzeugsatz, der nur lesen und nie schützen darf, ist in der falschen
// Richtung vorsichtig.

import (
	"context"
	"encoding/json"
	"fmt"
)

// Peerkontakt ist ein bekanntes Gegenüber, wie das Werkzeug es nennt.
//
// # Warum die Adresse hier fehlt
//
// **Weil sie etwas über das Netz dieses Menschen sagt** und nichts über den
// Kontakt. `192.168.1.7:41234` nennt ein Gerät im Wohnzimmer; ein Modell
// braucht das für keine Frage, die es beantworten soll, und was nicht
// hinausgeht, kann auch nicht weitergereicht werden.
//
// **Sie steht deshalb an drei Flächen und nicht an vier** (N2, #1455) — der
// Eintrag dazu steht in `docs/FAEHIGKEITEN.md`.
type Peerkontakt struct {
	Kennung              string   `json:"kennung"`
	Fingerabdruck        string   `json:"fingerabdruck"`
	SelbstgewaehlterName string   `json:"selbstgewaehlterName"`
	EigenerName          string   `json:"eigenerName"`
	Stand                string   `json:"stand"`
	Gesehen              string   `json:"gesehen"`
	Rechte               []string `json:"rechte"`
}

// Peerkontakte ist die Antwort beider Werkzeuge.
type Peerkontakte struct {
	Kontakte []Peerkontakt `json:"kontakte"`
	Zwecke   []string      `json:"zwecke"`

	// Nachladefehler sagt: **gesperrt ist**, nur die Liste danach nicht zu
	// lesen — dann ist `kontakte` leer (§7, Regel 3; #1960). Eine leere
	// Liste ohne dieses Feld hielte ein Modell für „niemand mehr bekannt".
	// **Erweiternd**: wo nichts fehlt, steht es nicht da.
	Nachladefehler *Nachladefehler `json:"nachladefehler,omitempty"`
}

// Kontaktauskunft ist die Naht zur Anwendung.
type Kontaktauskunft interface {
	Peerkontakte(ctx context.Context) (Peerkontakte, error)
	PeerkontaktSperren(ctx context.Context, kennung string) (Peerkontakte, error)
}

// kontaktwerkzeuge gibt das lesende und das schreibende getrennt.
//
// **Getrennt, weil der Satz sie getrennt führt**: ein Werkzeugsatz „nur
// lesend" darf das zweite nicht enthalten, und eine Liste, die beides
// zusammen gäbe, machte die Trennung zur Sache des Aufrufers.
func kontaktwerkzeuge(naht Kontaktauskunft) (lesend, schreibend []Werkzeug) {
	if naht == nil {
		return nil, nil
	}

	return []Werkzeug{{
		Name: "lesen.peerkontakte", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Gegenüber, die DIESES Profil kennt, und was " +
			"jedes darf. ÄNDERT NICHTS. »stand« ist »unbestaetigt« (man ist " +
			"sich begegnet, mehr nicht), »bestaetigt« (ein Mensch hat den " +
			"Fingerabdruck verglichen) oder »gesperrt«. »rechte« sind die " +
			"freigegebenen Zwecke aus »zwecke«; eine Bestätigung ALLEIN gibt " +
			"KEINEN Zweck frei. »selbstgewaehlterName« ist KEIN Nachweis — " +
			"verglichen wird »fingerabdruck«.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			kontakte, fehler := naht.Peerkontakte(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(kontakte)
		},
	}}, []Werkzeug{{
		Name: "schreiben.peerkontakt.sperren", Art: Schreibend,
		Erklaerung: "Schließt ein Gegenüber aus und nimmt ihm ALLES, was es " +
			"durfte. Gibt danach die ganze Liste zurück. Das Entsperren und " +
			"das Freigeben eines Zwecks gibt es hier NICHT: beides ist eine " +
			"Vertrauensentscheidung und gehört einem Menschen. Sperren ist " +
			"die Richtung, in der ein Irrtum nichts kostet. Steht " +
			"»nachladefehler« da, IST gesperrt — nur die Liste danach ließ " +
			"sich nicht lesen; NICHT noch einmal sperren.",
		Schema: objekt(map[string]any{
			"kennung": map[string]any{
				"type":        "string",
				"description": "Die Peerkennung des Gegenübers.",
			},
		}, "kennung"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wunsch struct {
				Kennung string `json:"kennung"`
			}

			if fehler := json.Unmarshal(roh, &wunsch); fehler != nil {
				return "", fmt.Errorf("die Argumente lesen: %w", fehler)
			}

			kontakte, fehler := naht.PeerkontaktSperren(ctx, wunsch.Kennung)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(kontakte)
		},
	}}
}
