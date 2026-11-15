package mcp

// **Der lokale Suchindex über MCP** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Zwei Werkzeuge, beide lesend
//
// **Stand und Suche.** Der Aufbau steht nicht hier, und zwar aus demselben
// Grund wie bei der Rücksetzung
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
// er läuft Minuten, bindet einen Kern, tauscht die aktive Generation und ist
// von hier aus nicht abbrechbar — ein Modell, das ihn anstößt, nähme dem
// Menschen am Fenster die Maschine weg. Kommandozeile und Dienst haben ihn.
//
// # Und die Suche ist das, wofür es A15 gibt
//
// **Ein Modell, das eine Regelfrage beantworten soll, braucht Fundstellen.**
// Was hier zurückkommt, trägt seine Herkunft: Quelle, Revision,
// Gültigkeitsdatum und Abrufzeit — damit eine Antwort belegbar wird statt
// plausibel.

import (
	"context"
	"encoding/json"
)

// Indexstand ist die Auskunft über den Suchindex.
type Indexstand struct {
	Aktiv    bool            `json:"aktiv"`
	Modell   string          `json:"modell"`
	Laufzeit string          `json:"laufzeit"`
	Staende  []Korpusstand   `json:"korpusstaende"`
	Suchbar  bool            `json:"suchbar"`
	Grund    string          `json:"grund"`
	Aufbau   Indexaufbaulage `json:"aufbau"`
}

// Korpusstand ist der Stand einer Quelle im Index.
type Korpusstand struct {
	Quelle   string `json:"quelle"`
	Revision string `json:"revision"`
	Datum    string `json:"datum"`
}

// Indexaufbaulage ist der Fortschritt eines Aufbaus.
type Indexaufbaulage struct {
	Laeuft      bool   `json:"laeuft"`
	Geschrieben int    `json:"geschrieben"`
	Fehler      string `json:"fehler"`
}

// Indextreffer ist ein Fund samt Herkunft.
type Indextreffer struct {
	Rang    int    `json:"rang"`
	Kennung string `json:"kennung"`
	Quelle  string `json:"quelle"`
	Adresse string `json:"adresse"`

	Titel      string `json:"titel"`
	Text       string `json:"text"`
	Ausschnitt string `json:"ausschnitt"`

	Sprache     string `json:"sprache"`
	Regelnummer string `json:"regelnummer"`
	Urheber     string `json:"urheber"`

	Quellrevision string `json:"quellrevision"`
	Quelldatum    string `json:"quelldatum"`
	Abgerufen     string `json:"abgerufen"`
}

// Indexsuche ist das Ergebnis einer Suche.
type Indexsuche struct {
	Treffer []Indextreffer `json:"treffer"`
	Modus   string         `json:"modus"`
	Grund   string         `json:"grund"`
}

// Indexfilter ist die geprüfte Filterstruktur.
type Indexfilter struct {
	Quellen  []string `json:"quellen"`
	Sprachen []string `json:"sprachen"`
	Sets     []string `json:"sets"`

	Farben         []string `json:"farben"`
	Farbidentitaet []string `json:"farbidentitaet"`
	Typen          []string `json:"typen"`

	ManawertVon *float64 `json:"manawertVon"`
	ManawertBis *float64 `json:"manawertBis"`

	Format      string `json:"format"`
	Formatlage  string `json:"formatlage"`
	Formatstand string `json:"formatstand"`
}

// Indexauskunft ist die Naht zur Anwendung.
type Indexauskunft interface {
	Indexstand(ctx context.Context) (Indexstand, error)
	Indexsuche(ctx context.Context, frage string, grenze int, filter Indexfilter) (
		Indexsuche, error)
}

func indexwerkzeuge(naht Indexauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{indexstandwerkzeug(naht), indexsuchwerkzeug(naht)}
}

// indexstandwerkzeug nennt, wogegen gesucht wird.
func indexstandwerkzeug(naht Indexauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.indexstand", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, ob ein lokaler Suchindex dasteht und wogegen " +
			"gesucht wird — Modell, Laufzeit und je Quelle ihr Stand. " +
			"»suchbar« sagt, ob **semantisch** gesucht werden kann; ist es " +
			"falsch, nennt »grund« den Grund, und »lesen.indexsuche« " +
			"antwortet trotzdem — dann aber nur lexikalisch. WICHTIG: die " +
			"Korpusstände nennen je Quelle eine eigene Revision; es gibt " +
			"keinen gemeinsamen »aktuellen Regelstand«.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			stand, fehler := naht.Indexstand(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(stand)
		},
	}
}

// indexsuchwerkzeug beantwortet eine Frage gegen den Index.
func indexsuchwerkzeug(naht Indexauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.indexsuche", Art: Lesend, Idempotent: true,
		Erklaerung: "Sucht lokal in Kartentexten, Regelstellen und " +
			"freigegebenen eigenen Inhalten — lexikalisch **und** semantisch, " +
			"zusammengeführt über ihre Ränge. Jeder Treffer trägt seine " +
			"Herkunft: Quelle, Revision, Gültigkeitsdatum und Abrufzeit. " +
			"»modus« sagt, wie gesucht wurde: »hybrid«, »lexikalisch« (kein " +
			"Einbettungsmodell) oder »ohne-index«. FILTER SIND STRUKTUR, " +
			"NICHT TEXT: »farben« ist die Farbe der Karte, »farbidentitaet« " +
			"ihre Identität nach Regel 903.4 — sie sind nicht dasselbe. Ein " +
			"Merkmal, das ein Dokument nicht hat, ist kein Treffer: eine " +
			"Regelstelle hat keine Farbe. »format« antwortet nur zusammen " +
			"mit dem Datenstand, aus dem die Legalität stammt.",
		Schema: objekt(map[string]any{
			"frage":  text("wonach gesucht wird"),
			"grenze": zahl("wie viele Treffer höchstens — Vorgabe 10"),
			"filter": objekt(map[string]any{
				"quellen": textliste("die Quellen, auf die eingeschränkt wird",
					"oracle, ruling, regelwerk, eigenekarte oder eigenertext"),
				"sprachen": textliste("die Sprachen", "ein Kürzel wie »en« oder »de«"),
				"sets":     textliste("die Sets", "ein Setcode wie »neo«"),
				"farben": textliste("die Farbe der Karte selbst",
					"W, U, B, R oder G"),
				"farbidentitaet": textliste(
					"die Farbidentität nach Regel 903.4 — nicht dasselbe wie die Farbe",
					"W, U, B, R oder G"),
				"typen": textliste(
					"Typzeilenwörter; **alle** müssen zutreffen",
					"ein Wort klein geschrieben, etwa »creature«"),
				"manawertVon": zahl("kleinster Manawert, einschließlich"),
				"manawertBis": zahl("größter Manawert, einschließlich"),
				"format":      text("ein Format wie »modern«"),
				"formatlage":  text("die verlangte Lage, etwa »legal«"),
				"formatstand": text(
					"der verlangte Datenstand der Legalität; leer heißt: jeder"),
			}),
		}, "frage"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var eingabe struct {
				Frage  string      `json:"frage"`
				Grenze int         `json:"grenze"`
				Filter Indexfilter `json:"filter"`
			}

			if fehler := json.Unmarshal(roh, &eingabe); fehler != nil {
				return "", fehler
			}

			if eingabe.Grenze <= 0 {
				eingabe.Grenze = 10
			}

			ergebnis, fehler := naht.Indexsuche(
				ctx, eingabe.Frage, eingabe.Grenze, eingabe.Filter)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}
}
