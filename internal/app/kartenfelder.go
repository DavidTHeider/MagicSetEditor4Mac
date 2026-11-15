package app

// Die Feldbeschreibungen eines Sets — **was ein Feld ist**, nicht was drinsteht.
//
// # Wofür
//
// Der Karteneditor stellte jedes Feld als einzeiliges Textfeld dar: den
// Regeltext, die Seltenheit, die Manakosten, die Kennung. Was ein Feld ist,
// steht im Spielpaket und kam nie im Frontend an — `WertDTO` trägt nur Name und
// Wert (#560).
//
// # Warum eine eigene Bindung
//
// Die Werte ändern sich bei jedem Tastendruck, die Beschreibungen nicht. Ein
// gemeinsamer Aufruf schickte die Erklärung jedes der 249 Kartenfelder durch
// jede Kartenliste.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// FeldbeschreibungDTO sagt, welche Form ein Kartenfeld hat.
type FeldbeschreibungDTO struct {
	Name string `json:"name"`

	// Aufschrift ist `caption:` in der gemeldeten Sprache — **die
	// Beschriftung, nicht der Bezeichner** ([domain.Field.Beschriftet]).
	//
	// **Hier gewählt und nicht im Frontend.** `localized caption:` trägt
	// Sprachkürzel aus dem Paket — `ch-s`, `jp`, `pt-br`; wer die Auswahl
	// hinausschöbe, schickte jedem Feld zwölf Übersetzungen mit, damit elf
	// davon verworfen werden.
	Aufschrift string `json:"aufschrift"`

	// Art ist `type:` aus dem Paket — `text`, `choice`, `multiple choice`,
	// `image`, `color`, `boolean`, `info`. **Unübersetzt** (§9): der Wert steht
	// so in der Datei, und das Frontend schaltet darauf.
	Art string `json:"art"`

	// Mehrzeilig ist `multi line:`.
	Mehrzeilig bool `json:"mehrzeilig"`

	// Wahlen sind die `choice:`-Namen einer Auswahl. Leer bei allem anderen.
	Wahlen []string `json:"wahlen"`

	// Beschreibung ist `description:` — der Satz, den MSE als Hilfetext zeigt.
	Beschreibung string `json:"beschreibung"`

	// Vorgabe ist der **Anfangswert** — was eine Karte ohne eigenen Wert in
	// diesem Feld trägt ([service.Anfangswert]).
	//
	// **Ohne `default:`.** Das ist ein Skript über die Karte, und diese
	// Bindung spricht über das Set. Ein Feld mit `default:` steht im Editor
	// deshalb leer, obwohl die gezeichnete Karte einen Wert zeigt —
	// #563.
	Vorgabe string `json:"vorgabe"`

	// Editierbar ist `editable:` — ob jemand dieses Feld bearbeiten darf.
	//
	// **Nicht dasselbe wie „wird gespeichert"** ([domain.Field.Editierbar]):
	// MSE lässt ein nicht editierbares Feld nicht einmal auswählen
	// (`src/gui/control/card_editor.cpp`, Zeilen 155–156, Fassung f9d9356 vom
	// 2021-07-06). 24 der 249 Kartenfelder von `magic.mse-game` tragen es
	// (#708).
	Editierbar bool `json:"editierbar"`

	// Gespeichert ist `save value:` — ob der Wert dieses Feldes in die Datei
	// geht oder aus anderen entsteht.
	//
	// **Nicht aus `Teilfelder` zu erraten.** Der erste Anlauf der
	// Massenänderung tat genau das: `teilfelder.length === 0` galt ihr als
	// „gespeichert". Ein berechnetes Feld, dessen Skript keine Glieder
	// hergibt, hat aber ebenfalls keine — `copyright line` in
	// `magic.mse-game` ist so eines. Es stand damit in der Auswahl, und der
	// Ruf dahinter endete zuverlässig im Fehler.
	//
	// Angestrichen vom Nachlauf über Phase 3.5 als Befund 1.
	Gespeichert bool `json:"gespeichert"`

	// Teilfelder sind die Felder, aus denen dieses Feld **entsteht** — leer bei
	// allem, was für sich steht ([service.Teilfelder]).
	//
	// **`text` ist der Fall, für den es das gibt.** Es trägt `save value:
	// false` und setzt sich aus Regeltext und Flavortext zusammen; wer es als
	// Textkasten anbot, nahm eine Eingabe entgegen, die nirgends ankam
	// (#696).
	//
	// **Die Reihenfolge ist die des Aufrufs**, denn sie ist die Reihenfolge
	// auf der Karte: Regeltext oben, Flavortext darunter.
	Teilfelder []string `json:"teilfelder"`

	// Reihe ist das Grundfeld der **Feldreihe**, zu der dieses Feld gehört —
	// leer, wenn es keine Geschwister hat
	// (#1167).
	//
	// **`name`, `name 2`, `name 3` und `name 4` tragen alle `name`.** Damit
	// sieht die Oberfläche, dass sie zusammengehören, ohne die Regel ein
	// zweites Mal zu kennen — sie steht in [domain.Game.Feldreihen], und zwei
	// Fassungen davon liefen an dem Tag auseinander, an dem jemand eine
	// ändert (§2).
	Reihe string `json:"reihe"`

	// Stelle ist die Position in der Reihe — 1 für das Grundfeld, 0 ohne
	// Reihe.
	//
	// **Was eine Stelle bedeutet, steht hier nicht.** Bei
	// `magic-m15-mainframe-dfc.mse-style` ist Stelle 2 die Rückseite, bei
	// `additional credit 2` die zweite Zeile einer Danksagung. Die Deutung
	// gehört dorthin, wo der Stil bekannt ist — hier steht die Gestalt.
	Stelle int `json:"stelle"`
}

// FeldbeschreibungenAntwort trägt die Beschreibungen oder einen Fehler.
type FeldbeschreibungenAntwort struct {
	Felder []FeldbeschreibungDTO `json:"felder"`
	Fehler *Fehler               `json:"fehler"`
}

// Kartenfelder gibt die Feldbeschreibungen des Spiels, zu dem ein Set gehört.
func (a *App) Kartenfelder(ctx context.Context, setID string) FeldbeschreibungenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return FeldbeschreibungenAntwort{Felder: []FeldbeschreibungDTO{}, Fehler: befund}
	}

	gelesen, err := a.stile.Kartenfelder(ctx, kennung)
	if err != nil {
		return FeldbeschreibungenAntwort{
			Felder: []FeldbeschreibungDTO{}, Fehler: Uebersetze(ctx, err),
		}
	}

	return FeldbeschreibungenAntwort{Felder: beschreibungenAus(gelesen, a.sprache)}
}

// beschreibungenAus macht aus den Feldern des Spiels ihre DTOs.
//
// **Über den Index**: ein [domain.Field] trägt inzwischen zwei Scheiben und
// sieben Zeichenketten, und `gocritic` zählt jede Kopie in einer Schleife nach.
func beschreibungenAus(felder []domain.Field, sprache string) []FeldbeschreibungDTO {
	umgesetzt := make([]FeldbeschreibungDTO, 0, len(felder))

	reihenlage := reihenlagen(felder)

	for i := range felder {
		// **Eine leere Liste und kein `nil`.** `json:"wahlen"` ohne
		// `omitempty` gäbe sonst `null`, und das Frontend prüfte auf zwei
		// Dinge statt auf eines.
		wahlen := felder[i].Choices
		if wahlen == nil {
			wahlen = []string{}
		}

		// **Eine leere Liste und kein `nil`** — dieselbe Begründung wie bei
		// den Wahlen.
		teile := service.Teilfelder(felder, felder[i])
		if teile == nil {
			teile = []string{}
		}

		reihe, stelle := reihenlage[felder[i].Name].Grundfeld, reihenlage[felder[i].Name].Stelle

		umgesetzt = append(umgesetzt, FeldbeschreibungDTO{
			Reihe:        reihe,
			Stelle:       stelle,
			Name:         felder[i].Name,
			Art:          string(felder[i].Kind),
			Mehrzeilig:   felder[i].Mehrzeilig,
			Wahlen:       wahlen,
			Beschreibung: felder[i].Beschreibung,
			Aufschrift:   felder[i].Beschriftet(sprache),
			Vorgabe:      service.Anfangswert(felder[i]),
			Editierbar:   felder[i].Editierbar,
			Gespeichert:  felder[i].Gespeichert,
			Teilfelder:   teile,
		})
	}

	return umgesetzt
}

// reihenlage sagt, zu welcher Reihe ein Feld gehört und an welcher Stelle.
type reihenlage struct {
	Grundfeld string
	Stelle    int
}

// reihenlagen ordnet jedem Feld einer Reihe seine Lage zu (#1167).
//
// **Das Grundfeld bekommt die Stelle 1.** Es ist die erste, und wer es ansieht,
// soll sehen, dass es Geschwister hat — ein Feld ohne Reihe bekommt gar nichts.
func reihenlagen(felder []domain.Field) map[string]reihenlage {
	lagen := make(map[string]reihenlage)

	for _, reihe := range (domain.Game{Fields: felder}).Feldreihen() {
		lagen[reihe.Grundfeld] = reihenlage{Grundfeld: reihe.Grundfeld, Stelle: 1}

		for _, stelle := range reihe.Stellen {
			lagen[stelle.Feld] = reihenlage{Grundfeld: reihe.Grundfeld, Stelle: stelle.Nummer}
		}
	}

	return lagen
}
