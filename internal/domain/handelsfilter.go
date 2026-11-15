package domain

// **Die Liste der Handelsvorgänge: finden, ordnen, eingrenzen** —
// TRADE-01.AK-02 (#1715).
//
// **Dieselbe Bauform wie die [Sammlungssicht]**: eine Sicht, die geprüft,
// angewandt und geordnet wird, ohne die gelesene Liste anzufassen — und die
// Ordnung nach Namen folgt der Sprache des Fensters, denn Namen sind
// Menschentext (`docs/LOKALISIERUNG.md`).

import (
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/collate"
)

// Handelsordnung sagt, wonach die Liste geordnet ist.
type Handelsordnung string

const (
	// OrdnungGeaendert ist die Vorgabe: zuletzt geändert zuerst.
	OrdnungGeaendert Handelsordnung = "geaendert"

	// OrdnungAngelegt ordnet nach dem Anfang des Vorgangs.
	OrdnungAngelegt Handelsordnung = "angelegt"

	// OrdnungLage stellt voran, was zu tun ist ([Handelslagen]).
	OrdnungLage Handelsordnung = "lage"

	// OrdnungPartner ordnet nach dem Namen des Gegenübers.
	OrdnungPartner Handelsordnung = "partner"
)

// Handelssuchgrenze ist die längste Suche — länger tippt niemand, und
// länger ist eine Eingabe, die niemand geprüft hat.
const Handelssuchgrenze = 200

// Handelspartner ist ein Gegenüber eines Vorgangs, so wie ich es kenne.
type Handelspartner struct {
	Kennung string

	// Name ist der, unter dem ich es kenne — leer, wenn ich es nicht kenne.
	Name string
}

// Handelsvorgangzeile ist ein Vorgang in der Liste — die letzte Revision
// und was sich aus ihr für mich ergibt.
type Handelsvorgangzeile struct {
	Angebot   Handelsangebot
	Lage      Handelslage
	Naechster Handelsschritt

	// Gegenueber sind die Parteien außer mir.
	Gegenueber []Handelspartner

	// Karten sind die Namen der Drucke, soweit dieser Bestand sie kennt.
	Karten []string

	Marken              []string
	Angelegt, Geaendert time.Time
	Archiviert          bool
}

// Handelsfilter ist, was jemand von seinen Vorgängen sehen will.
//
// **Leer heißt „alles außer dem Archiv"** — wer die Liste öffnet, will
// sehen, was läuft, und nicht, was vor einem Jahr endete.
type Handelsfilter struct {
	// Partner ist die Kennung eines Gegenübers.
	Partner string

	Lagen []Handelslage

	// Suche geht über Kennung, Gegenüber, Karten und Marken.
	Suche string

	// Marke wird in ihrer Normalform verglichen ([Markennormalform]).
	Marke string

	// Seit und Bis grenzen die letzte Änderung ein — beide eingeschlossen.
	Seit, Bis time.Time

	Ordnung Handelsordnung

	// Archiv zeigt auch archivierte Vorgänge.
	Archiv bool

	// Sprache ordnet die Namen; leer heißt Deutsch (wie [Sammlungssicht]).
	Sprache string
}

// Pruefe weist einen Filter ab, den es so nicht gibt.
func (f Handelsfilter) Pruefe() error {
	for _, lage := range f.Lagen {
		if !lage.Gueltig() {
			return NewFieldError("lagen", SchluesselValidierung,
				"%q ist keine Lage — es gibt »dran«, »wartet«, »abgeschlossen« und »abgebrochen«.",
				string(lage))
		}
	}

	switch f.Ordnung {
	case "", OrdnungGeaendert, OrdnungAngelegt, OrdnungLage, OrdnungPartner:
	default:
		return NewFieldError("ordnung", SchluesselValidierung,
			"%q ist keine Ordnung — es gibt »geaendert«, »angelegt«, »lage« und »partner«.",
			string(f.Ordnung))
	}

	if utf8.RuneCountInString(f.Suche) > Handelssuchgrenze {
		return NewFieldError("suche", SchluesselValidierung,
			"Eine Suche ist höchstens %d Zeichen lang.", Handelssuchgrenze)
	}

	if f.Marke != "" {
		if _, fehler := PruefeMarke(f.Marke); fehler != nil {
			return fehler
		}
	}

	if !f.Seit.IsZero() && !f.Bis.IsZero() && f.Bis.Before(f.Seit) {
		return NewFieldError("bis", SchluesselValidierung,
			"Das Ende liegt vor dem Anfang.")
	}

	return nil
}

// Gesehen gibt die Zeilen, die zu diesem Filter gehören — geordnet, in
// einer Kopie.
func (f Handelsfilter) Gesehen(zeilen []Handelsvorgangzeile) []Handelsvorgangzeile {
	gesucht := strings.ToLower(strings.TrimSpace(f.Suche))
	marke := Markennormalform(f.Marke)

	gesehen := make([]Handelsvorgangzeile, 0, len(zeilen))

	for _, zeile := range zeilen {
		if f.trifft(zeile, gesucht, marke) {
			gesehen = append(gesehen, zeile)
		}
	}

	f.ordne(gesehen)

	return gesehen
}

// trifft hält eine Zeile gegen jede Angabe des Filters.
func (f Handelsfilter) trifft(zeile Handelsvorgangzeile, gesucht, marke string) bool {
	switch {
	case zeile.Archiviert && !f.Archiv,
		f.Partner != "" && !slices.ContainsFunc(zeile.Gegenueber,
			func(p Handelspartner) bool { return p.Kennung == f.Partner }),
		len(f.Lagen) > 0 && !slices.Contains(f.Lagen, zeile.Lage),
		marke != "" && !slices.Contains(zeile.Marken, marke),
		!f.Seit.IsZero() && zeile.Geaendert.Before(f.Seit),
		!f.Bis.IsZero() && zeile.Geaendert.After(f.Bis):
		return false
	}

	return gesucht == "" || slices.ContainsFunc(suchfelder(zeile), func(feld string) bool {
		return strings.Contains(strings.ToLower(feld), gesucht)
	})
}

// suchfelder sind die Texte, in denen eine Suche einen Vorgang findet.
func suchfelder(zeile Handelsvorgangzeile) []string {
	felder := []string{zeile.Angebot.Vorgang}
	for _, partner := range zeile.Gegenueber {
		felder = append(felder, partner.Kennung, partner.Name)
	}

	return append(append(felder, zeile.Karten...), zeile.Marken...)
}

// ordne bringt die Zeilen in die gewählte Ordnung.
//
// **Die letzte Änderung ist immer der zweite Schlüssel und die Kennung der
// dritte** — sonst spränge eine Liste mit gleichen Lagen bei jedem Laden.
func (f Handelsfilter) ordne(zeilen []Handelsvorgangzeile) {
	// **Ein Ordner je Lauf** — `collate.Collator` ist nicht nebenläufig
	// benutzbar (der Grund steht bei [Sammlungssicht.ordne]).
	ordner := collate.New(ordnungssprache(f.Sprache))
	rang := func(lage Handelslage) int { return slices.Index(Handelslagen(), lage) }

	sort.SliceStable(zeilen, func(links, rechts int) bool {
		einer, anderer := zeilen[links], zeilen[rechts]

		switch f.Ordnung {
		case OrdnungAngelegt:
			if !einer.Angelegt.Equal(anderer.Angelegt) {
				return einer.Angelegt.After(anderer.Angelegt)
			}

		case OrdnungLage:
			if rang(einer.Lage) != rang(anderer.Lage) {
				return rang(einer.Lage) < rang(anderer.Lage)
			}

		case OrdnungPartner:
			if vergleich := ordner.CompareString(partnername(einer), partnername(anderer)); vergleich != 0 {
				return vergleich < 0
			}
		}

		if !einer.Geaendert.Equal(anderer.Geaendert) {
			return einer.Geaendert.After(anderer.Geaendert)
		}

		return einer.Angebot.Vorgang < anderer.Angebot.Vorgang
	})
}

// partnername ist, wonach ein Vorgang unter den Gegenübern steht — der
// Name des ersten, ohne Namen seine Kennung.
func partnername(zeile Handelsvorgangzeile) string {
	if len(zeile.Gegenueber) == 0 {
		return ""
	}

	if erster := zeile.Gegenueber[0]; erster.Name != "" {
		return erster.Name
	}

	return zeile.Gegenueber[0].Kennung
}
