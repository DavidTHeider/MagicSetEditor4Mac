// **Eine Kartenliste über MCP einlesen** — Werkstück 2 von Phase 4.5
// (#1292).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Zwei Werkzeuge, weil es zwei Schritte sind
//
// `lesen.liste.deuten` schreibt **nichts** — es sagt, welche Zeile welchen
// Druck meint. `schreiben.liste.einlesen` schreibt, was ihm ausdrücklich
// genannt wird.
//
// **Der Schnitt ist hier strenger als an der Kommandozeile.** Dort nimmt
// `liste-einlesen` das Eindeutige selbst; hier muss jede Zeile mit ihrer
// Druckkennung übergeben werden. Der Grund steht in
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md):
// wer tippt, sieht seinen Aufruf; ein Sprachmodell ruft, während niemand
// hinsieht. Ein Werkzeug, das selbst entscheidet, welche von 72 Ausgaben
// gemeint ist, entscheidet es unbeaufsichtigt.

package mcp

import (
	"context"
	"encoding/json"
)

// Listeneinfuhr ist, was MCP zum Einlesen einer Kartenliste braucht.
type Listeneinfuhr interface {
	Deuten(ctx context.Context, text string) (Listendeutung, error)
	Einlesen(ctx context.Context, zeilen []Listeneinfuhrzeile, zusammenfuehrung string) (
		Einfuhrbericht, error)

	// Vorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
	Vorschauen(ctx context.Context, zeilen []Listeneinfuhrzeile) (Sammelvorschau, error)
}

// Listentreffer ist ein Druck, der zu einer Zeile passt.
type Listentreffer struct {
	Printing   string `json:"printing"`
	Name       string `json:"name"`
	Setcode    string `json:"setcode"`
	Setname    string `json:"setname"`
	Nummer     string `json:"nummer"`
	Seltenheit string `json:"seltenheit"`
}

// Listendeutungszeile ist eine gelesene Zeile samt ihren Drucken.
type Listendeutungszeile struct {
	Roh          string          `json:"roh"`
	Zeilennummer int             `json:"zeilennummer"`
	Menge        int             `json:"menge"`
	Name         string          `json:"name"`
	Setcode      string          `json:"setcode"`
	Nummer       string          `json:"nummer"`
	Foil         bool            `json:"foil"`
	Finish       string          `json:"finish"`
	Treffer      []Listentreffer `json:"treffer"`
	Drucke       int             `json:"drucke"`
	Eindeutig    bool            `json:"eindeutig"`
	Befund       string          `json:"befund"`

	// Weitere sagt, dass es mehr Drucke gibt, als eine Seite nennt — dann
	// ist `drucke` eine Untergrenze (G01).
	Weitere bool `json:"weitere"`

	// Fehler ist der Ausfall dieser Zeile, mit dem Schlüssel, den auch das
	// Fenster übersetzt; nil heißt, die Quelle hat geantwortet, auch mit
	// »kein Druck« (G01).
	Fehler *Zeilenfehler `json:"fehler"`
}

// Zeilenfehler ist, woran die Frage einer Zeile scheiterte.
type Zeilenfehler struct {
	Code       string `json:"code"`
	Schluessel string `json:"schluessel"`
	Meldung    string `json:"meldung"`
}

// Listendeutung ist, was aus einer Liste zu machen wäre.
type Listendeutung struct {
	Zeilen        []Listendeutungszeile `json:"zeilen"`
	Befunde       []string              `json:"befunde"`
	AlterSekunden int64                 `json:"alterSekunden"`
}

// Listeneinfuhrzeile ist eine Zeile mit entschiedenem Druck.
type Listeneinfuhrzeile struct {
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Foil ist der alte Wahrheitswert — abgekündigt seit T2 (0119); beim
	// Schreiben entscheidet Finish, wenn es dasteht.
	Foil bool `json:"foil"`

	// Finish ist die Ausführung: `nonfoil`, `foil` oder `etched`.
	Finish string `json:"finish"`

	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
}

// Einfuhrbericht sagt, was eingelesen wurde.
type Einfuhrbericht struct {
	Angelegt []int64 `json:"angelegt"`
	Erhoeht  int     `json:"erhoeht"`

	// Ausgelassen zählt die Zeilen, die beim Auslassen stehen blieben.
	Ausgelassen int `json:"ausgelassen"`

	Stueck  int      `json:"stueck"`
	Befunde []string `json:"befunde"`
}

// listenzeilenschema beschreibt die Zeilen mit entschiedenem Druck — für
// das Einlesen und seine Vorschau, damit es **eine** Beschreibung ist.
func listenzeilenschema() map[string]any {
	return objektliste("Die Zeilen mit entschiedenem Druck.", objekt(
		map[string]any{
			"printing": text("Die Scryfall-Druckkennung aus »lesen.liste.deuten«."),
			"menge":    zahl("Die Stückzahl — mindestens 1."),
			"zustand": text("Der Erhaltungszustand: mint, nearMint, excellent, " +
				"good, lightPlayed, played oder poor."),
			"sprache": text("Der Sprachcode des Drucks, etwa »en« oder »de«."),
			"foil": jaNein("ABGEKÜNDIGT seit 0.5.0 — benutze »finish«. " +
				"Er entscheidet nur, solange »finish« fehlt."),
			"finish":  text("Die Ausführung: nonfoil, foil oder etched."),
			"name":    text("Der lesbare Kartenname."),
			"setcode": text("Der Setcode."),
			"nummer":  text("Die Sammlernummer."),
		}, "printing", "menge"))
}

// listeneinfuhr gibt die Werkzeuge — lesend zuerst.
func listeneinfuhr(einfuhr Listeneinfuhr) []Werkzeug {
	if einfuhr == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.liste.deuten", Art: Lesend,
		Erklaerung: "Liest eine Kartenliste (»4 Lightning Bolt (M11) 149«, eine " +
			"Zeile je Karte) und sagt zu jeder Zeile, welche Drucke bei Scryfall " +
			"dazu passen. SCHREIBT NICHTS. Eine Zeile ohne Setangabe passt oft " +
			"auf viele Drucke — »eindeutig« sagt, ob sie ohne Rückfrage " +
			"übernommen werden kann.",
		Schema: objekt(map[string]any{
			"text": text("Die Liste, eine Karte je Zeile. Leerzeilen, »Deck«, " +
				"»Sideboard« und Zeilen mit # oder // werden übergangen."),
		}, "text"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Text string `json:"text"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			deutung, err := einfuhr.Deuten(ctx, argumente.Text)
			if err != nil {
				return "", err
			}

			return alsJSON(deutung)
		},
	}, {
		Name: "schreiben.liste.einlesen", Art: Schreibend,
		Erklaerung: "ÄNDERT die Sammlung: nimmt die genannten Zeilen auf. Jede " +
			"Zeile nennt ihre Druckkennung — die aus »lesen.liste.deuten«. Eine " +
			"Zeile, die es schon gibt (gleicher Druck, Zustand, Sprache und " +
			"Ausführung), wird ERHÖHT statt abgewiesen; der Bericht sagt, wie " +
			"viele. Welchen von mehreren Drucken jemand besitzt, entscheidet " +
			"dieses Werkzeug NICHT.",
		Schema: objekt(map[string]any{
			"zeilen":           listenzeilenschema(),
			"zusammenfuehrung": zusammenfuehrungsfeld(),
		}, "zeilen"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Zeilen           []Listeneinfuhrzeile `json:"zeilen"`
				Zusammenfuehrung string               `json:"zusammenfuehrung"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			bericht, err := einfuhr.Einlesen(ctx, argumente.Zeilen, argumente.Zusammenfuehrung)
			// **Ein Fehler nennt, was schon geschrieben ist** (§7, Regel 3).
			//
			// MCP trägt entweder einen Text oder einen Fehler, nicht beides.
			// Deshalb steht der Bericht **in** der Fehlermeldung: wer bei
			// Zeile 217 von 300 abbricht, muss erfahren, dass 216 stehen —
			// sonst schreibt der zweite Versuch sie noch einmal.
			//
			// Befund 3 des Zweitblicks über Phase 4.5
			// (#1292).
			if err != nil {
				return "", mitBericht(err, bericht.Angelegt, bericht.Erhoeht)
			}

			return alsJSON(bericht)
		},
	}, lesenListenvorschau(einfuhr)}
}
