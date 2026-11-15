package domain

import (
	"strings"
	"time"
)

// Kartenverweis ist der Bezug einer entworfenen Karte auf einen echten Druck
// — Werkstück 2 von Phase 3.6
// (#1167).
//
// # Warum ein Verweis und keine Identität
//
// `0018_sammlung.sql` hat es beim Anlegen der Sammlung vorgezeichnet: „Eine
// Verbindung zu einer entworfenen Karte gibt es später als **Verweis** (Phase
// 3.6), nicht als Identität." Eine entworfene Karte **ist** kein Druck — sie
// darf ihm ähneln, von ihm abweichen oder ihn nur als Vorlage benutzen.
//
// # Warum ein Druck und kein Name
//
// Bis hierher ging jeder Weg zu Scryfall über einen **Namen**: `/cards/named`
// wählt dann selbst einen Druck aus. Wer in der Trefferliste einen bestimmten
// angeklickt hat, meinte aber genau den — und beim nächsten Abgleich kam
// wieder irgendeiner. Angestrichen im Werkstück davor
// (#1214):
// „Was er noch nicht tut: den genauen Druck wählen."
//
// # Warum `new` und wo gesucht wurde
//
// **MSE2 kennt Scryfall nicht.** Gesucht am 2026-08-27 in
// `/Users/david/mse2-quelle` (Fassung f9d9356 vom 2021-07-06) nach `scryfall`,
// `gatherer`, `oracle`, `multiverse`, `download` über `src/` — nichts, was
// eine Karte an einen fremden Datensatz bände. `magic-set-generator` (MIT,
// §11) geht den umgekehrten Weg: es **erzeugt** ein Set aus Scryfall-Daten und
// führt danach keinen Bezug mehr.
//
// **Der Verweis bleibt deshalb neben der Karte** und nicht in ihren
// Feldwerten: was in `card_values` steht, geht in jede `.mse-set`, die jemand
// weitergibt — und MSEs `Reader::unknownKey()` überspringt einen unbekannten
// Schlüssel und verliert ihn beim Speichern still
// ([0053](../../docs/entscheidungen/0053-marken-liegen-neben-dem-set.md)).
// Ein Verweis, der den Rundlauf nicht überlebt, wäre eine Zusage, die keiner
// einlöst.
type Kartenverweis struct {
	// Printing ist die Scryfall-Kennung des Drucks — eine UUID.
	//
	// **Derselbe Bezug wie in der Sammlung** (`sammlung.printing`): wer eine
	// entworfene Karte auf einen Druck bezieht, den er besitzt, soll beides
	// nebeneinanderlegen können, ohne dass jemand zwei Kennungen abgleicht.
	Printing string

	// DruckID ist die lokale Zeile des Drucks (A2, #1441, 0086) — neben dem
	// Alias, wie beim [Exemplar].
	DruckID int64

	// Name, Setcode und Nummer sind der **letzte gute Stand** der Anzeige.
	//
	// **Ohne sie wäre der Verweis ohne Netz eine UUID** — §16, Regel 5: der
	// letzte gute Stand bleibt, und er sagt, wie alt er ist. Das sagt
	// [Kartenverweis.Gesetzt].
	Name    string
	Setcode string
	Nummer  string

	// Adresse ist die Kartenseite bei Scryfall.
	//
	// **Eine Adresse und keine Bezeichnung**, wie bei der Wertherkunft: ein
	// Name allein ließe später niemanden nachsehen.
	Adresse string

	// Gesetzt ist der Zeitpunkt, an dem der Verweis entstand.
	Gesetzt time.Time
}

// Normalisieren glättet, was aus einer Eingabe kommt.
//
// **Getrennt von [Kartenverweis.Validate]**, wie beim Exemplar: eine Prüfung,
// die nebenbei ändert, lässt den Aufrufer im Unklaren, was gespeichert wurde.
//
// **Kennung und Setcode werden kleingeschrieben**, weil Scryfall sie so führt
// — nachgemessen am 2026-08-27: `/cards/516E137A-…` antwortet mit
// `516e137a-…`. Zwei Schreibweisen wären zwei Verweise auf denselben Druck,
// und wer neben dem einen den anderen suchte, fände nichts.
func (v *Kartenverweis) Normalisieren() {
	v.Printing = strings.ToLower(strings.TrimSpace(v.Printing))
	v.Setcode = strings.ToLower(strings.TrimSpace(v.Setcode))
	v.Name = strings.TrimSpace(v.Name)
	v.Nummer = strings.TrimSpace(v.Nummer)
	v.Adresse = strings.TrimSpace(v.Adresse)
}

// Validate gibt einen Befund je Feld, das nicht stimmt.
//
// **Eine Liste und kein einzelner Fehler** (§6), wie beim Exemplar.
func (v Kartenverweis) Validate() []*FieldError {
	var befunde []*FieldError

	if strings.TrimSpace(v.Printing) == "" && v.DruckID == 0 {
		befunde = append(befunde, NewFieldError("printing", SchluesselPflichtfeld,
			"Ein Verweis zeigt auf einen Druck."))
	}

	// **Der Name ist Pflicht und nicht Schmuck.** Ein Verweis ohne ihn zeigt
	// ohne Netz eine UUID, und niemand könnte sagen, worauf er zeigt.
	if strings.TrimSpace(v.Name) == "" {
		befunde = append(befunde, NewFieldError("name", SchluesselPflichtfeld,
			"Ein Verweis nennt die Karte, auf die er zeigt."))
	}

	return befunde
}

// Vorlaeufig sagt, ob der Verweis einen Druck ohne Alias nennt (0086).
func (v Kartenverweis) Vorlaeufig() bool { return druckVorlaeufig(v.DruckID, v.Printing) }
