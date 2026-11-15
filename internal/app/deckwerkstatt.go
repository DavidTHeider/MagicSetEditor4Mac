package app

// **Die Deckwerkstatt für MCP und REST** — Werkstück 7 von Phase 4.6
// (#1293).
//
// # Warum sie hier steht und nicht in beiden Kommandos
//
// **Weil sie zweimal dieselbe wäre.** `mcp.Decks` und `rest.Deckwerkstatt`
// sind gleichlautend — dieselbe Fähigkeit, dieselbe Naht (§16) —, und zwei
// Abschriften desselben Durchreichers laufen bei der ersten Änderung
// auseinander. Gos strukturelle Typisierung besorgt die Zuordnung an beiden
// Enden.
//
// **Sie gibt `(any, error)`**, weil beide Flächen das so brauchen: MCP trennt
// Ergebnis und Fehlschlag, REST setzt daraus einen Statuscode.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

// Deckwerkstatt reicht die Antworten der Anwendung an MCP und REST durch.
type Deckwerkstatt struct{ db *sql.DB }

// NeueDeckwerkstatt baut sie über einem Bestand.
func NeueDeckwerkstatt(db *sql.DB) *Deckwerkstatt { return &Deckwerkstatt{db: db} }

// flaeche baut die Anwendung — je Ruf einmal.
//
// **Ohne Fehlerergebnis**, und das ist keine Nachlässigkeit:
// [DeckflaecheUeber] gibt **immer** eine Fläche, auch wenn Scryfall fehlt.
// Ein Zweig „sie ließ sich nicht bauen" wäre einer, den kein Lauf betreten
// kann (§3).
//
// **Der fehlende Scryfall-Zugang trifft nur die Prüfung**, und die meldet es
// selbst. Wer deswegen die Deckliste verweigerte, machte aus einem fehlenden
// Zuhause einen Totalausfall.
func (d *Deckwerkstatt) flaeche() *Deckflaeche {
	flaeche, _ := DeckflaecheUeber(d.db) //nolint:errcheck // siehe oben

	return flaeche
}

func (d *Deckwerkstatt) Decks(ctx context.Context, spiel string) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.Decks(ctx, spiel)

	return ausMCP(antwort.Decks, antwort.Fehler)
}

func (d *Deckwerkstatt) Deck(ctx context.Context, deckID int64) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.Deck(ctx, zahl64(deckID))

	return ausMCP(antwort.Deck, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckAnlegen(ctx context.Context, name, spiel, notiz string) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.DeckAnlegen(ctx, name, spiel, notiz)

	return ausMCP(antwort.Deck, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckUmbenennen(
	ctx context.Context, deckID, fassung int64, name, notiz string,
) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.DeckUmbenennen(
		ctx, zahl64(deckID), zahl64(fassung), name, notiz)

	return ausMCP(antwort.Deck, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckFehlkartenWuenschen(ctx context.Context, deckID int64) (any, error) {
	antwort := d.flaeche().DeckFehlkartenWuenschen(ctx, zahl64(deckID))

	return ausMCP(map[string]any{
		"neu": antwort.Neu, "vorhanden": antwort.Vorhanden, "unbekannt": antwort.Unbekannt,
	}, antwort.Fehler)
}

// DeckAusSet legt aus den Karten eines Sets ein Deck an (FMT-04, #1693).
func (d *Deckwerkstatt) DeckAusSet(ctx context.Context, setID int64, name string, menge int) (any, error) {
	antwort := d.flaeche().DeckAusSet(ctx, zahl64(setID), name, menge)

	return ausMCP(antwort.Deck, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckKopieren(ctx context.Context, deckID int64, name string) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.DeckKopieren(ctx, zahl64(deckID), name)

	return ausMCP(antwort.Deck, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckLoeschen(ctx context.Context, deckID int64) error {
	flaeche := d.flaeche()

	return alsFehler(flaeche.DeckLoeschen(ctx, zahl64(deckID)).Fehler)
}

func (d *Deckwerkstatt) DeckplatzLegen(
	ctx context.Context, deckID, karteID int64, printing string, menge int, bereich string,
) (any, error) {
	flaeche := d.flaeche()

	// **Eine Null ist keine Kartenkennung** — sie heißt „kein Bezug über eine
	// eigene Karte", und die Naht erwartet dafür die leere Zeichenkette.
	karte := ""
	if karteID != 0 {
		karte = zahl64(karteID)
	}

	antwort := flaeche.DeckplatzLegen(
		ctx, zahl64(deckID), karte, printing, menge, bereich)

	return ausMCP(antwort.Platz, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckplatzNehmen(ctx context.Context, deckID, platzID int64) error {
	flaeche := d.flaeche()

	return alsFehler(flaeche.DeckplatzNehmen(ctx, zahl64(deckID), zahl64(platzID)).Fehler)
}

// DeckplatzVerschieben legt einen Platz in den anderen Bereich — in einem Zug.
func (d *Deckwerkstatt) DeckplatzVerschieben(
	ctx context.Context, deckID, platzID int64, bereich string,
) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.DeckplatzVerschieben(
		ctx, zahl64(deckID), zahl64(platzID), bereich)

	return ausMCP(antwort.Platz, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckFormatWaehlen(ctx context.Context, deckID, formatID int64) (any, error) {
	flaeche := d.flaeche()

	// **Eine Null nimmt das Format herunter.**
	format := ""
	if formatID != 0 {
		format = zahl64(formatID)
	}

	antwort := flaeche.DeckFormatWaehlen(ctx, zahl64(deckID), format)

	return ausMCP(antwort.Deck, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckPruefen(ctx context.Context, deckID int64) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.DeckPruefen(ctx, zahl64(deckID))
	if antwort.Fehler != nil {
		return nil, alsFehler(antwort.Fehler)
	}

	// **Die Befunde reisen mit** — sie sagen, worüber die Prüfung nichts
	// erfahren konnte.
	return map[string]any{"urteil": antwort.Urteil, "befunde": antwort.Befunde}, nil
}

func (d *Deckwerkstatt) DeckAbgleichen(ctx context.Context, deckID int64) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.DeckAbgleichen(ctx, zahl64(deckID))

	return ausMCP(map[string]any{
		"zeilen": antwort.Zeilen, "fehlend": antwort.Fehlend,
		"unbekannt": antwort.Unbekannt,
	}, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckZuordnungen(ctx context.Context, deckID int64) (any, error) {
	antwort := d.flaeche().DeckZuordnungen(ctx, zahl64(deckID))

	return ausMCP(map[string]any{
		"zeilen": antwort.Zeilen, "fehlend": antwort.Fehlend,
		"unbekannt": antwort.Unbekannt,
	}, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckZuordnen(
	ctx context.Context, deckID, platzID, exemplarID int64, menge int,
) (any, error) {
	antwort := d.flaeche().DeckZuordnen(ctx, zahl64(deckID), zahl64(platzID), zahl64(exemplarID), menge)

	return ausMCP(antwort.Zuordnung, antwort.Fehler)
}

func (d *Deckwerkstatt) DeckZuordnungLoesen(ctx context.Context, deckID, zuordnungID int64) error {
	antwort := d.flaeche().DeckZuordnungLoesen(ctx, zahl64(deckID), zahl64(zuordnungID))
	if antwort.Fehler != nil {
		return alsFehler(antwort.Fehler)
	}

	return nil
}

func (d *Deckwerkstatt) DeckFehlkarten(ctx context.Context, deckID int64) (any, error) {
	antwort := d.flaeche().DeckFehlkarten(ctx, zahl64(deckID))

	return ausMCP(map[string]any{"liste": antwort.Liste}, antwort.Fehler)
}

func (d *Deckwerkstatt) ExemplarDecks(ctx context.Context, exemplarID int64) (any, error) {
	antwort := d.flaeche().ExemplarDecks(ctx, zahl64(exemplarID))

	return ausMCP(antwort.Decks, antwort.Fehler)
}

// zahl64 gibt eine Kennung als Text — die Naht rechnet in Zeichenketten.
func zahl64(kennung int64) string { return strconv.FormatInt(kennung, 10) }

// ausMCP macht aus einer Antwort der Naht ein Ergebnis für MCP.
//
// **Ein Fehler ist dort ein `error` und kein Feld.** Das Protokoll trennt
// Ergebnis und Fehlschlag; ein Ergebnis mit einem Fehler darin läse sich wie
// ein Erfolg.
func ausMCP(ergebnis any, fehler *Fehler) (any, error) {
	if fehler != nil {
		return nil, alsFehler(fehler)
	}

	return ergebnis, nil
}

// alsFehler macht aus dem Befund der Naht einen Go-Fehler.
//
// **Der Schlüssel reist im Satz mit**, damit ein Modell die Art unterscheiden
// kann, ohne den deutschen Text zu deuten.
//
// # Und die Fehlerart reist als Fehlerart mit
//
// **Sonst kommt sie an keiner Fläche an.** Hier stand ein `fmt.Errorf` ohne
// `%w`, und daneben lag in `nahtfehler.go` längst [Fehler.AlsFehler] mit
// seiner Tabelle: die Sentinelwerte gingen verloren, und `internal/rest`
// prüft mit `errors.Is` — jede Abweisung eines Deckwegs kam damit als
// „unbekannt: der Aufruf ist gescheitert" heraus, statt als Validierung, als
// Konflikt oder als fehlende Erlaubnis.
func alsFehler(fehler *Fehler) error {
	if fehler == nil {
		return nil
	}

	satz := fmt.Sprintf("%s (%s)", fehler.Meldung, fehler.Schluessel)
	if fehler.Feld != "" {
		satz = fmt.Sprintf("%s (%s, Feld %s)",
			fehler.Meldung, fehler.Schluessel, fehler.Feld)
	}

	// **Dieselbe Tabelle wie [Fehler.AlsFehler]** (`nahtfehler.go`) — sie kennt
	// acht Codes, und eine zweite daneben liefe auseinander.
	if klasse, bekannt := arten[fehler.Code]; bekannt {
		return nahtfehler{satz: satz, klasse: klasse}
	}

	return errors.New(satz)
}

func (d *Deckwerkstatt) Bauarten(_ context.Context) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.Bauarten()

	return ausMCP(antwort.Bauarten, antwort.Fehler)
}

func (d *Deckwerkstatt) Formate(ctx context.Context, spiel string) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.Formate(ctx, spiel)

	return ausMCP(antwort.Formate, antwort.Fehler)
}

func (d *Deckwerkstatt) FormatAusBauart(
	ctx context.Context, bauart, name, spiel, scryfallname string,
) (any, error) {
	flaeche := d.flaeche()

	antwort := flaeche.FormatAusBauart(ctx, bauart, name, spiel, scryfallname)

	return ausMCP(antwort.Format, antwort.Fehler)
}

// FormatSetzen nimmt die Angaben als Abbildung entgegen.
//
// # Warum als Abbildung und nicht als Verbund
//
// **Weil eine fehlende Zahl etwas anderes heißt als eine Null.** Ein Modell,
// das nur den Namen ändern will, schickt `{"id":5,"fassung":1,"name":"…"}`
// — und eine Null bei `mindestgroesse` bedeutete „keine Grenze", nicht
// „unverändert". Die Abbildung sagt, **was gesagt wurde**; was fehlt, kommt
// aus dem vorhandenen Format.
func (d *Deckwerkstatt) FormatSetzen(ctx context.Context, angaben map[string]any) (any, error) {
	flaeche := d.flaeche()

	dto := FormatDTO{
		Name:              zeichenkette(angaben, "name"),
		Spiel:             zeichenkette(angaben, "spiel"),
		Scryfallname:      zeichenkette(angaben, "scryfallname"),
		Mindestgroesse:    ganzzahl(angaben, "mindestgroesse"),
		Hoechstgroesse:    ganzzahl(angaben, "hoechstgroesse"),
		Sideboardgroesse:  ganzzahl(angaben, "sideboardgroesse"),
		SideboardOffen:    wahrheit(angaben, "sideboardOffen"),
		HoechstzahlGleich: ganzzahl(angaben, "hoechstzahlGleich"),
	}

	if kennung := ganzzahl(angaben, "id"); kennung != 0 {
		dto.ID = zahl64(int64(kennung))

		// **Der vorhandene Stand füllt, was nicht gesagt wurde.** Ohne diesen
		// Ruf setzte ein Umbenennen jede Zahl auf null — und null heißt bei
		// drei von vieren „keine Grenze".
		vorher := flaeche.Formate(ctx, "")
		for _, format := range vorher.Formate {
			if format.ID == dto.ID {
				dto = ergaenzt(dto, format, angaben)
			}
		}
	}

	if fassung := ganzzahl(angaben, "fassung"); fassung != 0 {
		dto.Fassung = zahl64(int64(fassung))
	}

	antwort := flaeche.FormatSetzen(ctx, dto)

	return ausMCP(antwort.Format, antwort.Fehler)
}

func (d *Deckwerkstatt) FormatLoeschen(ctx context.Context, formatID int64) error {
	flaeche := d.flaeche()

	return alsFehler(flaeche.FormatLoeschen(ctx, zahl64(formatID)).Fehler)
}

// ergaenzt füllt aus dem vorhandenen Format, was der Ruf nicht genannt hat.
func ergaenzt(dto, vorher FormatDTO, angaben map[string]any) FormatDTO {
	for name, setze := range map[string]func(){
		"name":              func() { dto.Name = vorher.Name },
		"spiel":             func() { dto.Spiel = vorher.Spiel },
		"scryfallname":      func() { dto.Scryfallname = vorher.Scryfallname },
		"mindestgroesse":    func() { dto.Mindestgroesse = vorher.Mindestgroesse },
		"hoechstgroesse":    func() { dto.Hoechstgroesse = vorher.Hoechstgroesse },
		"sideboardgroesse":  func() { dto.Sideboardgroesse = vorher.Sideboardgroesse },
		"sideboardOffen":    func() { dto.SideboardOffen = vorher.SideboardOffen },
		"hoechstzahlGleich": func() { dto.HoechstzahlGleich = vorher.HoechstzahlGleich },
	} {
		if _, gesagt := angaben[name]; !gesagt {
			setze()
		}
	}

	return dto
}

// zeichenkette liest ein Textfeld aus der Abbildung — leer, wenn es fehlt.
func zeichenkette(angaben map[string]any, name string) string {
	wert, da := angaben[name].(string)
	if !da {
		return ""
	}

	return wert
}

// ganzzahl liest ein Zahlenfeld — **JSON kennt nur `float64`**.
func ganzzahl(angaben map[string]any, name string) int {
	wert, da := angaben[name].(float64)
	if !da {
		return 0
	}

	return int(wert)
}

// wahrheit liest ein Ja-Nein-Feld — falsch, wenn es fehlt.
//
// **Der Nullwert genügt hier**, anders als bei [ganzzahl]: der Unterschied
// zwischen „nicht genannt" und „falsch" entscheidet [ergaenzt] über
// `angaben`, nicht über den Wert.
func wahrheit(angaben map[string]any, name string) bool {
	wert, da := angaben[name].(bool)

	return da && wert
}
