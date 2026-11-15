// **Der Probelauf über MCP** — UXR-F7
// (#2015).
//
// # Warum eigene Werkzeuge und kein Feld an den schreibenden
//
// `schreiben.set.bogen` und `schreiben.vorlage.ausfuehren` tragen die Art
// *schreibend*, und ein Gegenüber, das Werkzeuge nach ihrer Art zulässt, würde
// ein Feld `probe` an ihnen als Schreiben lesen. Die zwei neuen sind **lesend
// und idempotent**: sie rechnen, was der schreibende Aufruf täte, und lassen
// die Platte, wie sie ist.
//
// Einordnung nach §16: **erweiternd** — zwei Werkzeuge kommen hinzu, kein
// vorhandenes ändert seine Bedeutung.

package mcp

import (
	"context"
	"encoding/json"
)

// Bogenprobelauf sagt, wie ein Druckbogen würde.
type Bogenprobelauf struct {
	Karten          int `json:"karten"`
	Seiten          int `json:"seiten"`
	Spalten         int `json:"spalten"`
	Zeilen          int `json:"zeilen"`
	JeSeite         int `json:"jeSeite"`
	AufLetzterSeite int `json:"aufLetzterSeite"`
	Bildbreite      int `json:"bildbreite"`

	// Maße in Millimetern.
	Seitenbreite float64 `json:"seitenbreite"`
	Seitenhoehe  float64 `json:"seitenhoehe"`
	Kartenbreite float64 `json:"kartenbreite"`
	Kartenhoehe  float64 `json:"kartenhoehe"`
	Anschnitt    float64 `json:"anschnitt"`
	Rand         float64 `json:"rand"`

	DPI           int  `json:"dpi"`
	Schnittmarken bool `json:"schnittmarken"`
}

// Vorlagenprobelauf sagt, was eine Exportvorlage ergäbe.
type Vorlagenprobelauf struct {
	Karten        int      `json:"karten"`
	Dateiname     string   `json:"dateiname"`
	Bytes         int      `json:"bytes"`
	Beigaben      int      `json:"beigaben"`
	BeigabenBytes int      `json:"beigabenBytes"`
	Verzeichnis   string   `json:"verzeichnis"`
	Geschaetzt    bool     `json:"geschaetzt"`
	Befunde       []string `json:"befunde"`
}

// Probelauf ist, was die zwei Werkzeuge dieser Datei brauchen.
//
// **Wie die Werkzeuge, die schreiben, nur ohne das Schreiben**: dieselben
// Angaben, derselbe Weg im Dienst, dieselben Fehler.
type Probelauf interface {
	// BogenProbelauf rechnet den Druckbogen durch, den `schreiben.set.bogen`
	// legte. `ziel` ist freiwillig: ist es genannt, antwortet der Probelauf
	// auch, wenn dort schon etwas liegt, wie der schreibende Aufruf.
	BogenProbelauf(ctx context.Context, setID int64, ziel string) (Bogenprobelauf, error)

	// VorlageProbelauf wertet die Exportvorlage aus, die
	// `schreiben.vorlage.ausfuehren` ausführte.
	VorlageProbelauf(ctx context.Context, setID int64, paket, ziel string,
		wahl map[string]string) (Vorlagenprobelauf, error)
}

// zielwunsch beschreibt das freiwillige Ziel beider Werkzeuge.
func zielwunsch() map[string]any {
	return text("Freiwillig: der Pfad, an den der schreibende Aufruf schriebe. " +
		"Er wird NICHT angelegt; liegt dort schon etwas, antwortet der Probelauf " +
		"mit demselben Nein wie der schreibende Aufruf.")
}

// probelaufwerkzeuge gibt die zwei lesenden Werkzeuge.
//
// **`nil` ist erlaubt** — dann gibt es sie nicht.
func probelaufwerkzeuge(probelauf Probelauf) []Werkzeug {
	if probelauf == nil {
		return nil
	}

	return []Werkzeug{
		{
			Name: "lesen.set.bogenprobelauf", Art: Lesend, Idempotent: true,
			Erklaerung: "Schreibt NICHTS: rechnet durch, wie der Druckbogen von " +
				"»schreiben.set.bogen« würde — wie viele Karten, wie viele Seiten, " +
				"wie viele Plätze je Seite, Papier und Auflösung — und gibt dieselben " +
				"Fehler (ein Set, das es nicht gibt, ein Set ohne Karte).",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"ziel":  zielwunsch(),
			}, "setID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64  `json:"setID"`
					Ziel  string `json:"ziel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				bogen, err := probelauf.BogenProbelauf(ctx, argumente.SetID, argumente.Ziel)
				if err != nil {
					return "", err
				}

				return alsJSON(struct {
					Probelauf bool `json:"probelauf"`
					Bogenprobelauf
				}{true, bogen})
			},
		},
		{
			Name: "lesen.vorlage.probelauf", Art: Lesend, Idempotent: true,
			Erklaerung: "Schreibt NICHTS: wertet eine Exportvorlage über dem Set aus, " +
				"wie »schreiben.vorlage.ausfuehren« es täte, und sagt, für wie viele " +
				"Karten, wie groß die Hauptdatei würde, wie viele Dateien daneben " +
				"entstünden — mit denselben Fehlern und Befunden. »wahl« setzt " +
				"einstellbare Felder für DIESEN Lauf.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"paket": text("Der Paketname der Vorlage aus »lesen.vorlagen«."),
				"ziel":  zielwunsch(),
				"wahl": map[string]any{
					"type":        "object",
					"description": "Optionswerte für DIESEN Lauf: {feld: wert}.",
				},
			}, "setID", "paket"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64             `json:"setID"`
					Paket string            `json:"paket"`
					Ziel  string            `json:"ziel"`
					Wahl  map[string]string `json:"wahl"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				vorlage, err := probelauf.VorlageProbelauf(
					ctx, argumente.SetID, argumente.Paket, argumente.Ziel, argumente.Wahl)
				if err != nil {
					return "", err
				}

				vorlage.Befunde = liste(vorlage.Befunde)

				return alsJSON(struct {
					Probelauf bool `json:"probelauf"`
					Vorlagenprobelauf
				}{true, vorlage})
			},
		},
	}
}
