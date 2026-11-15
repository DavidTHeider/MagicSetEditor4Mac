package app

// **Die Turnierwerkstatt für MCP und REST** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Wie die Deckwerkstatt und aus demselben Grund**: `mcp.Turniere` und der
// REST-Durchreicher sind gleichlautend, und zwei Abschriften desselben
// Durchreichers laufen bei der ersten Änderung auseinander. Gos strukturelle
// Typisierung besorgt die Zuordnung an beiden Enden.

import (
	"context"
	"database/sql"
)

// Turnierwerkstatt reicht die Antworten der Anwendung an MCP und REST durch.
type Turnierwerkstatt struct{ db *sql.DB }

// NeueTurnierwerkstatt baut sie über einem Bestand.
func NeueTurnierwerkstatt(db *sql.DB) *Turnierwerkstatt {
	return &Turnierwerkstatt{db: db}
}

// flaeche baut die Anwendung — je Ruf einmal.
//
// **Ohne Fehlerergebnis**, und anders als bei der Deckwerkstatt braucht es
// hier keine Erklärung: Turniere hängen an keiner Gegenstelle.
func (t *Turnierwerkstatt) flaeche() *Turnierflaeche {
	return TurnierflaecheUeber(t.db)
}

func (t *Turnierwerkstatt) Turniere(ctx context.Context, spiel string) (any, error) {
	antwort := t.flaeche().Turniere(ctx, spiel)

	return ausMCP(antwort.Turniere, antwort.Fehler)
}

func (t *Turnierwerkstatt) Turnier(ctx context.Context, turnierID int64) (any, error) {
	antwort := t.flaeche().Turnier(ctx, zahl64(turnierID))

	return ausMCP(antwort.Turnier, antwort.Fehler)
}

func (t *Turnierwerkstatt) TurnierRangliste(
	ctx context.Context, turnierID int64,
) (any, error) {
	antwort := t.flaeche().TurnierRangliste(ctx, zahl64(turnierID))

	return ausMCP(antwort.Rangliste, antwort.Fehler)
}

func (t *Turnierwerkstatt) TurnierPaarungsvorschlag(
	ctx context.Context, turnierID int64,
) (any, error) {
	antwort := t.flaeche().TurnierPaarungsvorschlag(ctx, zahl64(turnierID))

	return ausMCP(antwort.Paarungen, antwort.Fehler)
}

// TurnierAnlegen legt ein Turnier an — **eine Formatkennung von null heißt:
// keines**, wie beim Deck eines Teilnehmers.
func (t *Turnierwerkstatt) TurnierAnlegen(
	ctx context.Context, name, spiel, notiz string, einzelspiel bool, formatID int64,
) (any, error) {
	format := ""
	if formatID != 0 {
		format = zahl64(formatID)
	}

	antwort := t.flaeche().TurnierAnlegen(ctx, name, spiel, notiz, einzelspiel, format)

	return ausMCP(antwort.Turnier, antwort.Fehler)
}

// TurnierKopieren legt ein Turnier nach einer Vorlage an (DOM-04, #1699).
func (t *Turnierwerkstatt) TurnierKopieren(
	ctx context.Context, turnierID int64, name string, mitTeilnehmenden bool,
) (any, error) {
	antwort := t.flaeche().TurnierKopieren(ctx, zahl64(turnierID), name, mitTeilnehmenden)

	return ausMCP(antwort.Turnier, antwort.Fehler)
}

func (t *Turnierwerkstatt) TurnierLoeschen(ctx context.Context, turnierID int64) error {
	return alsFehler(t.flaeche().TurnierLoeschen(ctx, zahl64(turnierID)).Fehler)
}

// TurnierAnmelden trägt einen Teilnehmer ein.
//
// **Eine Deckkennung von null heißt: kein Deck.** MCP schickt für ein
// weggelassenes Zahlfeld die Null, und eine leere Zeichenkette ist das, was
// die Naht darunter als „keines" liest.
func (t *Turnierwerkstatt) TurnierAnmelden(
	ctx context.Context, turnierID int64, name string, deckID int64,
) (any, error) {
	deck := ""
	if deckID != 0 {
		deck = zahl64(deckID)
	}

	antwort := t.flaeche().TurnierAnmelden(ctx, zahl64(turnierID), name, deck)

	return ausMCP(antwort.Teilnehmer, antwort.Fehler)
}

func (t *Turnierwerkstatt) TurnierAussteigen(
	ctx context.Context, turnierID, teilnehmerID int64, ausgestiegen bool,
) error {
	return alsFehler(t.flaeche().TurnierAussteigen(
		ctx, zahl64(turnierID), zahl64(teilnehmerID), ausgestiegen).Fehler)
}

func (t *Turnierwerkstatt) TurnierRundeLegen(
	ctx context.Context, turnierID int64,
) (any, error) {
	antwort := t.flaeche().TurnierRundeLegen(ctx, zahl64(turnierID))

	return ausMCP(antwort.Runde, antwort.Fehler)
}

// TurnierErgebnisSetzen trägt ein Ergebnis ein — **`nil` nimmt es zurück**.
//
// **Die Fassung wird hier gelesen und mitgegeben.** MCP und Kommandozeile
// sind einschrittig: sie lesen und schreiben in einem Atemzug, und ein Stand
// von vorhin gäbe es gar nicht. Wer dazwischen schreibt, bekommt trotzdem
// seinen Konflikt — nur ist das Fenster dazwischen kleiner.
func (t *Turnierwerkstatt) TurnierErgebnisSetzen(
	ctx context.Context, turnierID, paarungID int64, stand map[string]any,
) error {
	var ergebnis *ErgebnisDTO

	if stand != nil {
		ergebnis = &ErgebnisDTO{
			SpieleA:       ganzzahlAusAny(stand["spieleA"]),
			SpieleB:       ganzzahlAusAny(stand["spieleB"]),
			Unentschieden: ganzzahlAusAny(stand["unentschieden"]),
		}
	}

	gelesen := t.flaeche().Turnier(ctx, zahl64(turnierID))
	if gelesen.Fehler != nil {
		return alsFehler(gelesen.Fehler)
	}

	return alsFehler(t.flaeche().TurnierErgebnisSetzen(
		ctx, zahl64(turnierID), gelesen.Turnier.Fassung, zahl64(paarungID),
		ergebnis).Fehler)
}

// ganzzahlAusAny liest eine Zahl aus einem Wert unbekannter Gestalt.
//
// **Zwei Fälle, weil zwei Wege hierher führen**: MCP hat sie schon zu `int`
// gemacht, ein JSON-Rumpf bringt sie als `float64`. Was keines von beidem
// ist, zählt null — der Aufrufer hat sie dann nicht genannt.
func ganzzahlAusAny(wert any) int {
	switch zahl := wert.(type) {
	case int:
		return zahl
	case float64:
		return int(zahl)
	}

	return 0
}

// TurnierKorrekturfolgen sagt, was eine Berichtigung verschöbe (DOM-04.AK-02)
// — die Zahlen kommen wie beim Eintragen als `any`.
func (t *Turnierwerkstatt) TurnierKorrekturfolgen(
	ctx context.Context, turnierID, paarungID int64, stand map[string]any,
) (any, error) {
	var ergebnis *ErgebnisDTO

	if stand != nil {
		ergebnis = &ErgebnisDTO{
			SpieleA:       ganzzahlAusAny(stand["spieleA"]),
			SpieleB:       ganzzahlAusAny(stand["spieleB"]),
			Unentschieden: ganzzahlAusAny(stand["unentschieden"]),
		}
	}

	antwort := t.flaeche().TurnierKorrekturfolgen(ctx, zahl64(turnierID), zahl64(paarungID), ergebnis)

	return ausMCP(antwort.Folgen, antwort.Fehler)
}

// TurnierAbschliessen schließt ein Turnier ab (DOM-04.AK-04) — **die Fassung
// liest die Werkstatt selbst**, wie beim Eintragen: MCP und REST sind
// einschrittig.
func (t *Turnierwerkstatt) TurnierAbschliessen(ctx context.Context, turnierID int64) error {
	return t.abschlussSetzen(ctx, turnierID, t.flaeche().TurnierAbschliessen)
}

// TurnierWiedereroeffnen nimmt den Abschluss zurück.
func (t *Turnierwerkstatt) TurnierWiedereroeffnen(ctx context.Context, turnierID int64) error {
	return t.abschlussSetzen(ctx, turnierID, t.flaeche().TurnierWiedereroeffnen)
}

// abschlussSetzen liest die Fassung und ruft den Weg.
func (t *Turnierwerkstatt) abschlussSetzen(
	ctx context.Context, turnierID int64,
	weg func(context.Context, string, string) SchreibAntwort,
) error {
	gelesen := t.flaeche().Turnier(ctx, zahl64(turnierID))
	if gelesen.Fehler != nil {
		return alsFehler(gelesen.Fehler)
	}

	return alsFehler(weg(ctx, zahl64(turnierID), gelesen.Turnier.Fassung).Fehler)
}

// TurnierProtokoll gibt, was an einem Turnier geschah.
func (t *Turnierwerkstatt) TurnierProtokoll(ctx context.Context, turnierID int64) (any, error) {
	antwort := t.flaeche().TurnierProtokoll(ctx, zahl64(turnierID))

	return ausMCP(antwort.Eintraege, antwort.Fehler)
}

// Turniertext ist eine Turnierdatei oder ein Bericht als Text — für MCP und
// REST, die an keiner Platte stehen (DOM-04.AK-03, #1699).
type Turniertext struct {
	Kennung string `json:"kennung"`
	Inhalt  string `json:"inhalt"`
}

// Turniereingang ist ein eingelesenes Turnier samt Befunden (§8).
type Turniereingang struct {
	Kennung string   `json:"kennung"`
	Befunde []string `json:"befunde"`
}

// TurnierAusgabe gibt ein Turnier als Datei im Text.
func (t *Turnierwerkstatt) TurnierAusgabe(ctx context.Context, turnierID int64) (any, error) {
	antwort := t.flaeche().TurnierAusgabeText(ctx, zahl64(turnierID))

	return ausMCP(Turniertext{Kennung: antwort.Kennung, Inhalt: antwort.Inhalt}, antwort.Fehler)
}

// TurnierBericht gibt den Abschlussbericht — auf Deutsch, wie die Fläche.
func (t *Turnierwerkstatt) TurnierBericht(ctx context.Context, turnierID int64) (any, error) {
	antwort := t.flaeche().TurnierBerichtText(ctx, zahl64(turnierID))

	return ausMCP(Turniertext{Kennung: antwort.Kennung, Inhalt: antwort.Inhalt}, antwort.Fehler)
}

// TurnierEinlesen legt ein Turnier aus dem Text einer Ausgabe an.
func (t *Turnierwerkstatt) TurnierEinlesen(ctx context.Context, inhalt string) (any, error) {
	antwort := t.flaeche().TurnierEinlesenText(ctx, inhalt)

	return ausMCP(Turniereingang{Kennung: antwort.Kennung, Befunde: antwort.Befunde}, antwort.Fehler)
}
