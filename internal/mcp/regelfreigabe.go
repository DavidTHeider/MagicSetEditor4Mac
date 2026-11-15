package mcp

// **Die Regelfreigabe als Werkzeug** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Warum ein Modell das fragen darf
//
// **Weil es rein lesend ist und nichts öffnet.** Es sagt, was dieses Programm
// durchsetzt und was nicht — eine Auskunft über die eigene Reichweite, keine
// über den Bestand und keine über das Netz dieses Menschen.
//
// **Und weil eine Sperre, die nur das Fenster kennt, keine ist** (#1461:
// „weder über UI noch REST/CLI/MCP").

import (
	"context"
	"encoding/json"
	"fmt"
)

// Regelbereichsangaben sind, was die Freigabe über einen Bereich sagt.
type Regelbereichsangaben struct {
	Bereich string   `json:"bereich"`
	Grad    string   `json:"grad"`
	Quelle  string   `json:"quelle"`
	Grenzen []string `json:"grenzen"`
}

// Regelfreigabeangaben sind die Antwort auf „darf es so losgehen?".
type Regelfreigabeangaben struct {
	Erlaubt   bool   `json:"erlaubt"`
	Regelwerk string `json:"regelwerk"`
	Fassung   int    `json:"fassung"`
	Wirksam   string `json:"wirksam"`
	Hoechste  string `json:"hoechste"`

	Matrix      []Regelbereichsangaben `json:"matrix"`
	Luecken     []string               `json:"luecken"`
	Hindernisse []string               `json:"hindernisse"`

	// Modus ist der Bedienmodus, über den gerechnet wurde — PLAY-02.AK-01
	// (#1710).
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
	// wer ihn nicht kennt, übergeht ihn.
	Modus string `json:"modus"`

	// Baum ordnet die Bereiche nach Ablauf und Mechanik (AK-02).
	Baum []Regelgruppenangaben `json:"baum"`

	// Widersprueche sind die Abwahlen, an denen noch etwas hängt — sie
	// **erklären** und verhindern nichts.
	Widersprueche []Regelwiderspruchsangaben `json:"widersprueche"`
}

// Regelgruppenangaben sind ein Ast des Regelbaums.
type Regelgruppenangaben struct {
	Kennung  string   `json:"kennung"`
	Bereiche []string `json:"bereiche"`
}

// Regelwiderspruchsangaben sind eine Abwahl, an der noch etwas hängt.
type Regelwiderspruchsangaben struct {
	Bereich string `json:"bereich"`
	Haengt  string `json:"haengt"`
	Quelle  string `json:"quelle"`
}

// Regelquellenleser ist, was dieses Werkzeug braucht.
type Regelquellenleser interface {
	Regelfreigabe(ctx context.Context, spiel, strenge string, sitze int,
		modus string, abgewaehlt []string) (
		Regelfreigabeangaben, error)
}

// regelfreigabewerkzeug baut das eine.
func regelfreigabewerkzeug(quelle Regelquellenleser) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.regelfreigabe", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, welche Spielregeln dieses Programm automatisch " +
			"durchsetzt und welche nicht, und ob eine Partie in der " +
			"gewünschten Strenge anfangen darf. ÄNDERT NICHTS. Eine Lücke " +
			"ist kein Fehler: sie heißt, dass die Menschen am Tisch es " +
			"selbst machen. »automatisch« gilt NUR in den Grenzen, die unter " +
			"»grenzen« stehen. »streng« steht nur zur Wahl, wenn JEDER " +
			"Regelbereich durchgesetzt wird — welche es sind, sagt »matrix«.",
		Schema: objekt(map[string]any{
			"spiel":   text("Das Spiel — »magic«, »space«."),
			"strenge": text("»sandkasten«, »begleitet« oder »streng«."),
			"sitze":   zahl("Wie viele mitspielen."),

			// **Beide sind freiwillig** (PLAY-02, #1710): wer nichts sagt,
			// bekommt den Anfängermodus und keine Abwahl — so viel Hilfe,
			// wie dieses Programm tragen kann.
			"modus": text("»anfaenger« (Vorgabe) oder »frei«. Nur im " +
				"freien Modus lässt sich ein Bereich abwählen."),
			"abgewaehlt": textliste("Die Regelbereiche, die die Runde "+
				"SELBST übernimmt. Der Baum in der Antwort nennt sie alle.",
				"Ein Regelbereich, etwa »combat«."),
		}, "spiel", "strenge", "sitze"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var gefragt struct {
				Spiel      string   `json:"spiel"`
				Strenge    string   `json:"strenge"`
				Sitze      int      `json:"sitze"`
				Modus      string   `json:"modus"`
				Abgewaehlt []string `json:"abgewaehlt"`
			}

			if err := json.Unmarshal(roh, &gefragt); err != nil {
				return "", fmt.Errorf("die Angaben lesen: %w", err)
			}

			angaben, err := quelle.Regelfreigabe(
				ctx, gefragt.Spiel, gefragt.Strenge, gefragt.Sitze,
				gefragt.Modus, gefragt.Abgewaehlt)
			if err != nil {
				return "", err
			}

			return alsJSON(angaben)
		},
	}}
}
