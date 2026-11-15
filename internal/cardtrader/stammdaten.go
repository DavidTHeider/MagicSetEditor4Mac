package cardtrader

// **Spiele und Erweiterungen** — die Stammdaten, über die ein Druck dieses
// Bestands seinen Blueprint findet (A4, #1443).
//
// # Warum sie eine eigene Datei sind
//
// **Weil sie anders altern als der Markt.** Welche Erweiterungen es gibt,
// ändert sich einige Male im Jahr; welche Angebote es gibt, im Minutentakt.
// Wer beides in einen Zwischenspeicher legte, holte entweder zu oft Stammdaten
// oder zu selten Preise (§16).

import "context"

// Spiel ist eines der Spiele, die CardTrader führt.
type Spiel struct {
	ID          int64
	Name        string
	Anzeigename string
}

// spielDTO ist die Gestalt, in der CardTrader ein Spiel schickt.
type spielDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Anzeigename string `json:"display_name"`
}

// Erweiterung ist ein Set, wie CardTrader es führt.
type Erweiterung struct {
	ID      int64
	SpielID int64

	// Code ist die Brücke zum Bestand: ein Druck hier kennt seinen Setcode.
	Code string

	Name string
}

// erweiterungDTO ist die Gestalt, in der CardTrader eine Erweiterung schickt.
type erweiterungDTO struct {
	ID      int64  `json:"id"`
	SpielID int64  `json:"game_id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
}

// Spiele gibt, was CardTrader führt.
func (k *Klient) Spiele(ctx context.Context) ([]Spiel, error) {
	var roh []spielDTO

	if fehler := k.hole(ctx, wege.Spiele, nil, &roh); fehler != nil {
		return nil, fehler
	}

	spiele := make([]Spiel, 0, len(roh))
	for _, eines := range roh {
		// **Eine Umwandlung und kein Literal**: die Felder stehen gleich, und
		// ein Literal daneben liefe beim nächsten Feld auseinander.
		spiele = append(spiele, Spiel(eines))
	}

	return spiele, nil
}

// Erweiterungen gibt alle Sets über alle Spiele.
//
// **CardTrader trennt sie nicht nach Spiel** — die Antwort führt jede
// Erweiterung mit ihrer `game_id`, und wer nur Magic will, filtert hier.
func (k *Klient) Erweiterungen(ctx context.Context) ([]Erweiterung, error) {
	var roh []erweiterungDTO

	if fehler := k.hole(ctx, wege.Erweiterungen, nil, &roh); fehler != nil {
		return nil, fehler
	}

	erweiterungen := make([]Erweiterung, 0, len(roh))
	for _, eine := range roh {
		erweiterungen = append(erweiterungen, Erweiterung(eine))
	}

	return erweiterungen, nil
}

// SpielMagic ist die Kennung, unter der CardTrader Magic führt.
//
// **Nachgelesen am 2026-09-11** (0095): `{"id": 1, "name": "Magic",
// "display_name": "Magic: the Gathering"}`. Als Konstante **und** über
// [Klient.Spiele] prüfbar: eine feste Zahl allein wäre eine Behauptung, die
// still veraltet.
const SpielMagic int64 = 1
