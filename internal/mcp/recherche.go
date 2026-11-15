package mcp

// **Die Recherche über MCP** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Zwei Werkzeuge, beide lesend
//
// **Recherche und Ähnlichkeit.** Was zurückkommt, ist mehr als eine
// Trefferliste: jeder Fund trägt seine Herkunft, und **eine leere Antwort
// trägt ihren Grund**. „Nichts gefunden" heißt sechs Dinge, und fünf davon
// sagen dem Modell, was es tun kann.
//
// # Und das Kontextpaket steht nicht hier
//
// **Weil der Aufrufer hier das Modell ist.** Ein Kontextpaket ist ein Budget
// **für ein Zielmodell** — Fenster, Ausgabereserve, Zählweise —, und wer es
// über MCP anforderte, budgetierte für jemand anderen. Was das Paket an
// Belegen vergibt, steht hier ohnehin an jedem Treffer; was es zusätzlich
// könnte, ist kürzen, und das Kürzen gehört dem, der das Fenster kennt.
//
// **Kommandozeile und Dienst haben es**, und dort ergibt es Sinn: ein Skript
// füttert ein Modell und will vorher sehen, womit.

import (
	"context"
	"encoding/json"
)

// Recherchetreffer ist ein Fund samt auflösbarer Herkunft.
type Recherchetreffer struct {
	Rang    int    `json:"rang"`
	Kennung string `json:"kennung"`
	Quelle  string `json:"quelle"`

	// Kennungssystem, Fachkennung und Semantik sind die Identität nach A2 —
	// **kein Name als Identität**.
	Kennungssystem string `json:"kennungssystem"`
	Fachkennung    string `json:"fachkennung"`
	Semantik       string `json:"semantik"`

	// EigeneKarteID ist gesetzt, wo der Treffer aus diesem Bestand stammt.
	EigeneKarteID int64 `json:"eigeneKarteID"`

	Titel      string `json:"titel"`
	Ausschnitt string `json:"ausschnitt"`
	Text       string `json:"text"`

	Sprache     string `json:"sprache"`
	Regelnummer string `json:"regelnummer"`
	Urheber     string `json:"urheber"`

	Bezugsadresse string `json:"bezugsadresse"`

	// Quellrevision, Quelldatum und Abgerufen sind die Herkunft — drei
	// Angaben, weil das Abrufdatum kein Gültigkeitsdatum ist.
	Quellrevision string `json:"quellrevision"`
	Quelldatum    string `json:"quelldatum"`
	Abgerufen     string `json:"abgerufen"`

	// Weitere nennt die gebündelten Abschnitte desselben Konzepts.
	Weitere []string `json:"weitere"`
}

// Rechercheergebnis ist die Antwort auf eine Frage.
type Rechercheergebnis struct {
	Treffer []Recherchetreffer `json:"treffer"`

	Modus string `json:"modus"`

	// Leere sagt, **warum** nichts kam — leer, wo etwas kam.
	Leere string `json:"leere"`
	Grund string `json:"grund"`

	Modell            string `json:"modell"`
	Suchkonfiguration string `json:"suchkonfiguration"`

	Korpusstaende []Korpusstand `json:"korpusstaende"`

	// Einschraenkungen nennt, was **nicht** getan wurde.
	Einschraenkungen []string `json:"einschraenkungen"`

	Fortsetzung string `json:"fortsetzung"`
}

// Rechercheauskunft ist die Naht zur Anwendung.
type Rechercheauskunft interface {
	// **`freigegeben` nennt die privaten Korpora, die gesucht werden
	// dürfen** — A13.4 (#1452). Leer heißt: keine.
	Recherche(ctx context.Context, fall, frage string, grenze int,
		sprache, korpusstand, fortsetzung string,
		freigegeben []string) (Rechercheergebnis, error)
	AehnlicheKarten(ctx context.Context, karteID int64, grenze int,
		sprache, fortsetzung string,
		freigegeben []string) (Rechercheergebnis, error)
}

func recherchewerkzeuge(naht Rechercheauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{recherchewerkzeug(naht), aehnlichkeitswerkzeug(naht)}
}

// recherchewerkzeug beantwortet eine Frage gegen den Wissensbestand.
func recherchewerkzeug(naht Rechercheauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.recherche", Art: Lesend, Idempotent: true,
		Erklaerung: "Fragt den lokalen Wissensbestand — Kartentexte, " +
			"Rulings, Regelwerk und freigegebene eigene Inhalte. »fall« " +
			"wählt, wonach gesucht wird: »karten«, »regeln« oder " +
			"»rulings«; jeder Fall fragt andere Quellen. JEDER TREFFER " +
			"TRÄGT SEINE HERKUNFT — Quelle, Revision, Gültigkeitsdatum und " +
			"Abrufzeit sind vier verschiedene Angaben. WICHTIG: »leere« " +
			"sagt, WARUM nichts kam — »nichts-gefunden«, »ohne-index«, " +
			"»ohne-modell«, »ohne-quelle«, »gesperrt« oder »budget« — " +
			"»ohne-modell« heißt: das lokale EINBETTUNGSmodell fehlt, und " +
			"ohne es gibt es nur die Volltexthälfte. Nur " +
			"das erste heißt, dass es die Antwort nicht gibt; bei den " +
			"anderen fehlt etwas, das sich einrichten lässt. Antworte nicht " +
			"aus dem Gedächtnis, wenn hier nichts kam — sag, was fehlt.",
		Schema: objekt(map[string]any{
			"fall": text(
				"»karten«, »regeln« oder »rulings« — Vorgabe »karten«"),
			"frage":  text("wonach gesucht wird"),
			"grenze": zahl("wie viele Treffer höchstens — Vorgabe 10"),
			"sprache": text(
				"auf eine Sprache einschränken, etwa »de«; leer heißt: jede"),
			"korpusstand": text(
				"einen bestimmten Korpusstand verlangen; leer heißt: der aktive"),
			"fortsetzung": text(
				"der Zeiger aus einer vorigen Antwort für die nächste Seite"),
			"freigegeben": freigabeschema(),
		}, "frage"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var eingabe struct {
				Fall        string   `json:"fall"`
				Frage       string   `json:"frage"`
				Grenze      int      `json:"grenze"`
				Sprache     string   `json:"sprache"`
				Korpusstand string   `json:"korpusstand"`
				Fortsetzung string   `json:"fortsetzung"`
				Freigegeben []string `json:"freigegeben"`
			}

			if fehler := json.Unmarshal(roh, &eingabe); fehler != nil {
				return "", fehler
			}

			ergebnis, fehler := naht.Recherche(ctx, eingabe.Fall,
				eingabe.Frage, grenzeOder(eingabe.Grenze), eingabe.Sprache,
				eingabe.Korpusstand, eingabe.Fortsetzung, eingabe.Freigegeben)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}
}

// aehnlichkeitswerkzeug sucht Karten, die einer eigenen ähneln.
func aehnlichkeitswerkzeug(naht Rechercheauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.aehnlichekarten", Art: Lesend, Idempotent: true,
		Erklaerung: "Findet Karten, die einer Karte AUS DIESEM BESTAND " +
			"ähneln — gesucht wird mit ihrem Text, nicht mit ihrem Namen. " +
			"Die Ausgangskarte selbst steht nicht in der Antwort. " +
			"GESUCHT WIRD IN ÖFFENTLICHEN KARTENTEXTEN; die Karten dieses " +
			"Bestands kommen nur dazu, wenn »freigegeben« sie nennt — sonst " +
			"sagt »einschraenkungen«, dass sie ausgelassen wurden. " +
			"BRAUCHT EINEN SEMANTISCHEN INDEX: ohne Einbettungsmodell steht " +
			"in »leere« »ohne-modell«, und es kommt nichts — ein " +
			"lexikalischer Rückfall auf den Kartentext fände Wörter statt " +
			"Ähnlichkeit.",
		Schema: objekt(map[string]any{
			"karte":  zahl("die Kennung der Karte in diesem Bestand"),
			"grenze": zahl("wie viele Treffer höchstens — Vorgabe 10"),
			"sprache": text(
				"auf eine Sprache einschränken; leer heißt: jede"),
			"fortsetzung": text("der Zeiger für die nächste Seite"),
			"freigegeben": freigabeschema(),
		}, "karte"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var eingabe struct {
				Karte       int64    `json:"karte"`
				Grenze      int      `json:"grenze"`
				Sprache     string   `json:"sprache"`
				Fortsetzung string   `json:"fortsetzung"`
				Freigegeben []string `json:"freigegeben"`
			}

			if fehler := json.Unmarshal(roh, &eingabe); fehler != nil {
				return "", fehler
			}

			ergebnis, fehler := naht.AehnlicheKarten(ctx, eingabe.Karte,
				grenzeOder(eingabe.Grenze), eingabe.Sprache,
				eingabe.Fortsetzung, eingabe.Freigegeben)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}
}

// freigabeschema beschreibt die ausdrückliche Freigabe privater Korpora.
//
// **An einer Stelle für beide Werkzeuge** — zwei Fassungen desselben Satzes
// liefen beim ersten Sonderfall auseinander, und ein Modell läse dann zweimal
// etwas anderes über dieselbe Entscheidung.
func freigabeschema() map[string]any {
	return map[string]any{
		"type":  "array",
		"items": map[string]any{"type": "string"},
		"description": "Welche PRIVATEN Korpora mitgesucht werden dürfen — " +
			"»eigenekarte« und »eigenertext«. LEER HEISST KEINE: der eigene " +
			"Bestand bleibt dann außen, und »einschraenkungen« sagt es. " +
			"FRAG EINEN MENSCHEN, bevor du das setzt — es ist seine " +
			"Entscheidung und nicht deine.",
	}
}

// grenzeOder gibt zehn, wo keine Zahl steht.
func grenzeOder(grenze int) int {
	if grenze <= 0 {
		return 10
	}

	return grenze
}
