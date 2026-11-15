package scryfall

import (
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setangabe ist, was Scryfall über einen Satz sagt.
//
// **Nur die Felder, die dieses Programm benutzt.** Scryfall schickt gut
// zwanzig; sie alle abzubilden hieße, die Gestalt eines fremden Anbieters in
// der eigenen Fachlichkeit zu führen — §16: „Wer den fremden Verbund direkt in
// `service` reicht, hat die Struktur eines fremden Anbieters in seiner
// Fachlogik."
//
// Gelesen am 2026-08-24 unter `https://scryfall.com/docs/api/sets`.
type Setangabe struct {
	// Code ist das drei- bis sechsstellige Kürzel — `lea`, `m15`, `neo`.
	//
	// **Es ist der Schlüssel, unter dem MSE Sets kennt**, und deshalb Pflicht.
	Code string

	Name string

	// Art ist `set_type` — `core`, `expansion`, `masters`, `token` und weitere.
	Art string

	// Erschienen ist `released_at`, als Zeichenkette.
	//
	// **Ohne Umrechnung**, und mit demselben Grund wie bei [Regelhinweis.Datum]:
	// die Dokumentation nennt „GMT-8 Pacific time", und wer daraus einen
	// Kalendertag in der Zone des Rechners machte, verschöbe jedes Datum.
	Erschienen string

	// Karten ist `card_count`.
	Karten int

	// Digital sagt, ob der Satz nur in einem Videospiel erschienen ist.
	Digital bool

	// Adresse ist die Setseite bei Scryfall — für die Herkunftsangabe.
	Adresse string
}

// bekannteSetarten sind die Werte, die dieses Programm kennt.
//
// **Gelesen, nicht geraten** — die Liste steht unter „Set Types" auf
// `https://scryfall.com/docs/api/sets`, Stand 2026-08-24.
var bekannteSetarten = []string{
	"core", "expansion", "masters", "alchemy", "masterpiece", "arsenal",
	"from_the_vault", "spellbook", "premium_deck", "duel_deck",
	"draft_innovation", "treasure_chest", "commander", "planechase",
	"archenemy", "vanguard", "funny", "starter", "box", "promo", "token",
	"memorabilia", "minigame",
}

// rohsatz ist die Gestalt, die Scryfall schickt.
type rohsatz struct {
	Objekt     string `json:"object"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Art        string `json:"set_type"`
	Erschienen string `json:"released_at"`
	Karten     int    `json:"card_count"`
	Digital    bool   `json:"digital"`
	Adresse    string `json:"scryfall_uri"`
}

// LiesSetangabe liest eine Antwort von `/sets/:code`.
//
// **`code` und `name` sind Pflicht.** Ein Satz ohne Kürzel ließe sich keinem
// MSE-Set zuordnen, und einer ohne Namen wäre in einer Liste nicht
// wiederzufinden. Alles Übrige darf fehlen.
func LiesSetangabe(roh []byte) (*Setangabe, []domain.Befund, error) {
	var gelesen rohsatz

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		return nil, nil, fmt.Errorf("%w: die Setangabe ließ sich nicht lesen: %s",
			domain.ErrValidation, fehler.Error())
	}

	if gelesen.Code == "" || gelesen.Name == "" {
		return nil, nil, fmt.Errorf(
			"%w: die Setangabe nennt weder Kürzel noch Namen vollständig "+
				"(code=%q, name=%q)", domain.ErrValidation, gelesen.Code, gelesen.Name)
	}

	return &Setangabe{
		Code: gelesen.Code, Name: gelesen.Name, Art: gelesen.Art,
		Erschienen: gelesen.Erschienen, Karten: gelesen.Karten,
		Digital: gelesen.Digital, Adresse: gelesen.Adresse,
	}, setartbefunde(gelesen), nil
}

// setartbefunde meldet eine Satzart, die dieses Programm nicht kennt.
func setartbefunde(gelesen rohsatz) []domain.Befund {
	if gelesen.Art == "" || istBekannt(bekannteSetarten, gelesen.Art) {
		return nil
	}

	return []domain.Befund{{
		Datei: gelesen.Code,
		Ort:   "set_type",
		Art:   domain.ArtUmgebung,
		Meldung: fmt.Sprintf(
			"Scryfall meldet für den Satz %q die unbekannte Art %q. Die Angaben "+
				"werden übernommen; die Art bleibt so stehen, wie sie kam.",
			gelesen.Code, gelesen.Art),
	}}
}

// istBekannt sagt, ob ein Wert in der Liste steht.
func istBekannt(liste []string, wert string) bool {
	for _, einer := range liste {
		if einer == wert {
			return true
		}
	}

	return false
}
