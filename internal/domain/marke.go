package domain

import (
	"strings"
	"unicode/utf8"
)

// **Marken an Karten** — Werkstück 3 von Phase 3.5
// (#1208).
//
// # Was eine Marke ist
//
// Ein **Wort, das jemand an eine Karte hängt** — „fertig", „prüfen", „noch
// kein Bild". Sie steht nicht in der Karte, sondern daneben: eine Aussage über
// die **Arbeit** an ihr und keine über das Spiel.
//
// # Warum sie neben dem Set liegt
//
// Was in `card_values` steht, geht in jede `.mse-set`, die jemand weitergibt —
// und käme dort nie wieder heraus: MSEs `Reader::unknownKey()` überspringt
// einen unbekannten Schlüssel und verliert ihn beim Speichern **still**
// (`src/util/io/reader.cpp`, Zeilen 231–246, Fassung `c9c708d` vom
// 2020-05-27). Eine Marke in einem Kartenfeld wäre nach einem Rundlauf durch
// MSE2 fort, und niemand sähe, wann
// ([0053](../../docs/entscheidungen/0053-marken-liegen-neben-dem-set.md)).
//
// # Warum `new` und wo gesucht wurde
//
// **MSE2 kennt nichts dergleichen.** Gesucht am 2026-08-28 in
// `/Users/david/mse2-quelle` (Fassung f9d9356 vom 2021-07-06) über `src/` nach
// `tag`, `label`, `category`, `flag`, `marker`: `src/data/card.hpp` führt
// `data`, `notes`, zwei Zeitstempel, `stylesheet`, `styling_data`,
// `extra_data` und `keyword_usage` — keine freie Beschriftung. Was dem am
// nächsten kommt, ist `notes`, und das ist **ein** Text je Karte, kein
// Schlüssel, nach dem sich filtern ließe.
//
// Damit gibt es keine Vorlage, und jede Frage ist hier entschieden — jede
// Entscheidung steht bei ihrem Gegenstand.

// Markengrenze ist die größte Länge einer Marke, in **Zeichen**.
//
// **Eine Marke ist ein Schlüssel und kein Satz.** Wer einen Absatz an eine
// Karte hängen will, hat dafür die Notizen ([Notizenfeld]); eine Marke, die
// nicht in eine Kachel passt, ist keine mehr.
//
// **In Zeichen und nicht in Bytes**: „prüfen" ist auf Deutsch sieben Bytes und
// sechs Zeichen, und eine Grenze, die je nach Sprache anders greift, ist keine.
const Markengrenze = 40

// Markenfeld ist der Feldname, unter dem ein Markentreffer in einer Suche
// steht.
//
// **Ein eigener Name und kein Kartenfeld**: die Trefferliste zeigt, **wo**
// etwas gefunden wurde, und „in einer Marke" ist etwas anderes als „im
// Regeltext". Ein Kartenfeld dieses Namens gibt es in keiner Vorlage des
// Packs — nachgezählt am 2026-08-28 über `full-magic-pack/data`.
const Markenfeld = "marke"

// Markennormalform bringt eine Marke in die Gestalt, in der sie gespeichert
// wird.
//
// # Warum kleingeschrieben
//
// **Damit ein Filter nicht still auseinanderfällt.** Wer an einer Karte
// „Fertig" und an der nächsten „fertig" schreibt, hätte sonst zwei Marken, und
// jede Auswahl zeigte die Hälfte. Eine Marke ist ein Schlüssel; zwei
// Schreibweisen desselben Wortes sind kein Merkmal, sondern ein Tippfehler.
//
// **Der Preis steht dazu**: deutsche Substantive stehen danach klein. Das ist
// die Seite, die man sieht — die andere sähe niemand.
//
// # Warum der Innenraum gefaltet wird
//
// Zwei Leerzeichen sind in einer Kachel nicht zu sehen. Eine Marke, die sich
// von einer anderen nur darin unterscheidet, wäre ein Fehler ohne Anzeige.
func Markennormalform(roh string) string {
	return strings.ToLower(strings.Join(strings.Fields(roh), " "))
}

// PruefeMarke gibt die Normalform — oder sagt, warum es keine gibt.
func PruefeMarke(roh string) (string, error) {
	marke := Markennormalform(roh)

	if marke == "" {
		return "", NewFieldError("marke", SchluesselPflichtfeld,
			"Eine Marke ohne Text gibt es nicht.")
	}

	if utf8.RuneCountInString(marke) > Markengrenze {
		return "", NewFieldError("marke", SchluesselValidierung,
			"Eine Marke ist höchstens %d Zeichen lang.", Markengrenze)
	}

	return marke, nil
}
