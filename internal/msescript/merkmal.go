// Package msescript zerteilt und wertet die in MSE-Paketen eingebettete
// Sprache aus.
//
// Der englische Satzanfang ist keine Ausnahme von §9, sondern seine Regel: die
// Form `Package <name>` ist Teil der Go-Schnittstelle nach außen — `go doc`
// liest sie, und `staticcheck` besteht darauf. Der Rest des Absatzes ist
// deutsch wie überall.
//
// **Es kennt nichts aus diesem Projekt, auch nicht `domain`** (AGENTS.md §2).
// Eingabe ist Quelltext und eine Umgebung aus Werten, Ausgabe ist ein Wert.
// Eine Kartenentität hierin machte den Auswerter unprüfbar: man könnte ihn
// nicht mehr gegen eine Tabelle aus Ausdruck und Ergebnis fahren, sondern
// bräuchte für jeden Fall eine Karte.
//
// Was die Sprache ist, steht in `docs/FORMATE.md` — erschlossen an fünf
// Paketdateien aus drei Spielen, nicht aus dem MSE2-Quelltext (§11).
package msescript

import "fmt"

// Art benennt, was ein Merkmal ist.
//
// Als Zeichenkette und nicht als Zahl, weil eine fehlgeschlagene Tabelle sonst
// „erwartet 3, bekam 7" meldet. Der Preis ist ein Vergleich über Zeichenketten
// an einer Stelle, die nicht heiß ist.
type Art string

const (
	ArtBezeichner     Art = "Bezeichner"
	ArtZahl           Art = "Zahl"
	ArtZeichenkette   Art = "Zeichenkette"
	ArtSchluesselwort Art = "Schlüsselwort"
	ArtOperator       Art = "Operator"
	ArtZeilenende     Art = "Zeilenende"

	// ArtKommentar steht **nur** im Strom von [Einfaerben]
	// (#1164).
	//
	// Der Zerteiler bekommt sie nicht zu sehen: `#` bis zum Zeilenende ist für
	// ihn Zwischenraum. Ein Editor braucht sie — sie sind das Erste, was
	// jemand eingefärbt sehen will.
	ArtKommentar Art = "Kommentar"
)

// Schluesselwoerter sind die Namen, die die Sprache selbst vergibt.
//
// Erschlossen, nicht erfunden: jedes steht in `docs/FORMATE.md` mit seiner
// Quelle. Was hier fehlt, ist ein gewöhnlicher Bezeichner — auch dann, wenn es
// in einer anderen Sprache ein Schlüsselwort wäre.
var Schluesselwoerter = map[string]bool{
	"if":   true,
	"then": true,
	"else": true,
	"and":  true,
	"mod":  true,
	"or":   true,
	"xor":  true,
	"not":  true,
	"nil":  true,

	// **Die Schleife der Exportvorlagen** (#51): `for each card in cards do …`.
	// Alle 19 Vorlagen von `M15-Magic-Pack` benutzen sie.
	"for":  true,
	"each": true,
	"in":   true,
	"do":   true,
	"from": true,
	"to":   true,

	// **Die Fallunterscheidung** (#214): `case input of "token": true, else:
	// false;`. Achtmal in vier Dateien von `M15-Magic-Pack`. `of` bleibt
	// daneben ein Name — es steht 48-mal als Argumentname und 11-mal als Wert.
	"case": true,
	"of":   true,
}

// Merkmal ist ein einzelnes Stück Quelltext mit seiner Stelle.
type Merkmal struct {
	Art    Art
	Text   string
	Zeile  int
	Spalte int

	// Breite ist, wie viele Zeichen im **Quelltext** standen.
	//
	// **Nicht `len(Text)`.** Bei einer Zeichenkette fehlen im Text die
	// Anführungszeichen: `"x"` ist drei Zeichen breit und trägt einen Text von
	// einem. Der Zerteiler rechnet daraus die Stelle **hinter** dem letzten
	// Merkmal, und ohne dieses Feld meldete `("x"` die fehlende Klammer bei
	// Spalte 3 statt bei 5 — der Finger zeigte mitten in die Zeichenkette.
	//
	// Gefunden vom Zweitblick über Phase 3.
	Breite int
}

func (m Merkmal) String() string {
	return fmt.Sprintf("%s(%q) at %d:%d", m.Art, m.Text, m.Zeile, m.Spalte)
}

// Fehler nennt eine Stelle, nicht nur eine Ursache.
//
// **Die Stelle ist der ganze Zweck** (AGENTS.md §8): ein Paket kann von 2007
// stammen und halb geschrieben sein. „unerwartetes Zeichen" hilft niemandem,
// „Zeile 214, Spalte 9" führt hin.
type Fehler struct {
	Zeile   int
	Spalte  int
	Meldung string

	// Schluessel benennt die Meldung —
	// #1178.
	//
	// **Damit das Fenster sie übersetzen kann.** [Fehler.Meldung] ist der
	// deutsche Satz, und der bleibt: §9 sagt, Fehlermeldungen sind deutsch,
	// und Kommandozeile, MCP und Dienst bekommen ihn so. Das Fenster bekommt
	// zusätzlich Schlüssel und Werte und setzt seinen eigenen Text ein.
	Schluessel Meldungsschluessel

	// Werte füllt die Platzhalter des Satzes.
	//
	// **Als Zeichenketten und nicht als `any`**: sie gehen über die
	// Wails-Naht ans Fenster, und dort ist alles Text.
	Werte []string
}

// NeuerFehler baut einen Fehler aus Stelle, Schlüssel und Werten.
//
// **Der Satz entsteht hier und nicht beim Aufrufer.** Wer ihn selbst
// zusammensetzte, hätte zwei Orte für eine Meldung — und der Vertragstest
// gegen die Kataloge läse dann den halben Bestand.
func NeuerFehler(
	zeile, spalte int, schluessel Meldungsschluessel, werte ...string,
) *Fehler {
	return &Fehler{
		Zeile:      zeile,
		Spalte:     spalte,
		Meldung:    satzZu(schluessel, werte),
		Schluessel: schluessel,
		Werte:      werte,
	}
}

func (f *Fehler) Error() string {
	return fmt.Sprintf("Zeile %d, Spalte %d: %s", f.Zeile, f.Spalte, f.Meldung)
}
