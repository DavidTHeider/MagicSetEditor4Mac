package service

// **Welche Auszeichnung was bedeutet — und welche MSE gar nicht deutet.**
//
// Getrennt von `auszeichnungen.go`, das den Text zerlegt: dort steht der Weg
// durch die Zeichenkette, hier die Frage, was ein gefundener Name heißt.

import "strings"

// istKursiv sagt, ob eine Auszeichnung kursiv setzt.
//
// **Ein Präfix und kein Name.** MSE prüft `is_tag(text, pos, "<i")`
// (`src/render/text/element.cpp`, Zeile 80) — deshalb wirken `<i>`,
// `<i-flavor>` und `<i-auto>` gleich, und ein künftiges `<i-was-auch-immer>`
// ebenso. Eine Aufzählung der drei bekannten Namen ginge beim vierten schief.
//
// Fassung f9d9356 vom 2021-07-06.
func istKursiv(name string) bool {
	return trifft(name, "i")
}

// istFett sagt, ob eine Auszeichnung fett setzt.
//
// **Ein Präfix und kein Name**, wie bei [istKursiv]: MSE prüft
// `is_tag(text, pos, "<b")` (`src/render/text/element.cpp`, Zeile 78, Fassung
// f9d9356 vom 2021-07-06;
// [0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
//
// **`<bullet>` fällt nicht darunter** — seit
// #429.
// Hinter dem Präfix muss ein Endzeichen stehen ([trifft]), und in `<bullet>`
// steht dort ein `u`. Die Gegenprobe steht im Altcode: er führt einen eigenen
// `</bullet`-Zweig (`element.cpp:171`), und den erreichte niemand, wenn `<b`
// ihn vorher schluckte.
func istFett(name string) bool {
	return trifft(name, "b")
}

// istSymbolisch sagt, ob eine Auszeichnung in Symbolen setzt.
//
// **Auch hier ein Präfix** — `is_tag(text, pos, "<sym")`
// (`src/render/text/element.cpp`, Zeile 82, Fassung f9d9356 vom 2021-07-06).
func istSymbolisch(name string) bool {
	return trifft(name, "sym")
}

// istWeich sagt, ob eine Auszeichnung ihren Inhalt **weich** macht.
//
// **Drei Namen, und die Reihenfolge zählt.** MSE prüft `<sep-soft` vor
// `<soft`, weil `is_tag` auf einem Präfix arbeitet; `<soft-line` steht davor
// und ist etwas anderes — ein weicher **Umbruch**, keine weiche Stelle
// (`src/render/text/element.cpp`, Zeilen 86–92, Fassung f9d9356 vom
// 2021-07-06).
//
//	bool soft : 1; ///< Is this a 'soft' character? soft characters are
//	               ///  ignored for alignment
//
// **`<atom-kwpph>` gehört nicht dazu.** Hier stand bis zum 2026-08-21 das
// Gegenteil, mit Verweis auf Zeile 299 — und die entscheidet die
// **Schriftfarbe**, nicht die Weichheit:
//
//	(soft        > 0 ? FONT_SOFT        : FONT_NORMAL) |
//	(kwpph       > 0 ? FONT_SOFT        : FONT_NORMAL) |
//
// Weich im Sinne der Ausrichtung **und des Zeichnens** wird ein Stück in Zeile
// 234, und dort steht allein `soft`. Ein Platzhalter wird also gezeichnet — in
// Trennfarbe — und zählt bei der Ausrichtung mit
// (#725).
//
// **Folgenlos war der Fehler nur, solange weich bloß die Ausrichtung betraf.**
// Seit weicher Text nicht mehr gezeichnet wird, wäre der Platzhalter
// unsichtbar geworden. In den zehn kanonischen Sets und im ganzen
// `M15-Magic-Pack` steht er null Mal; er entsteht erst beim Aufschlüsseln
// eines Schlüsselworts mit leerem Parameter (`src/data/keyword.cpp`, Zeile
// 638).
// istWeicheZeile sagt, ob eine Auszeichnung **weich umbricht** — `<soft-line>`.
//
// **Vor [istWeich] geprüft, wie bei MSE.** `is_tag` vergleicht ein Präfix, und
// `<soft-line` fiele auf `<soft`, stünde es dahinter; MSE prüft es deshalb
// zuerst (`src/render/text/element.cpp`, Zeilen 86–92, Fassung f9d9356 vom
// 2021-07-06). Die beiden sind nicht dasselbe: `<soft>` betrifft die
// Ausrichtung, `<soft-line>` den Umbruch.
func istWeicheZeile(name string) bool { return trifft(name, "soft-line") }

// istRand sagt, ob eine Auszeichnung Absatzränder setzt — `<margin:l:r:t>`.
func istRand(name string) bool { return trifft(name, "margin") }

// istAufzaehlung sagt, ob eine Auszeichnung eine Aufzählung eröffnet — `<li>`.
//
// **Sie tut bei MSE nichts als warnen.** `<li>` zählt nur mit, damit ein
// `<bullet>` außerhalb gemeldet werden kann (`src/render/text/element.cpp`,
// Zeilen 92–93 und 173–175, Fassung f9d9356 vom 2021-07-06); die Wirkung geht
// vom schließenden `</bullet` aus. Hier wird sie deshalb gedeutet und
// verworfen — ein Befund für „nichts zu tun" verdeckte die echten (§8).
func istAufzaehlung(name string) bool { return trifft(name, "li") }

// istImRand sagt, ob eine Auszeichnung ihre Zeichen in den linken Rand legt —
// `<bullet>`. Siehe [render.Abschnitt.ImRand].
func istImRand(name string) bool { return trifft(name, "bullet") }

// istWeich sagt, ob eine Auszeichnung ihren Inhalt **weich** macht.
func istWeich(name string) bool {
	if trifft(name, "soft-line") {
		return false
	}

	return trifft(name, "soft") || trifft(name, "sep-soft")
}

// istSchlichtesAtom sagt, ob eine Auszeichnung ein `<atom…>` ohne Wirkung auf
// eine **gedruckte** Karte ist.
//
// # Wirkung nur im Editor
//
// MSE macht aus jedem `<atom…>`, das nicht `atom-param` oder `atom-kwpph`
// heißt, ein `AtomTextElement` mit einer Hintergrundfarbe
// (`src/render/text/element.cpp`, Zeilen 150–160). Gemalt wird sie nur, wenn
// der Editor zeichnet:
//
//	if (what & DRAW_ACTIVE) {
//	  dc.SetPen(*wxTRANSPARENT_PEN);
//	  dc.SetBrush(background_color);
//	  dc.DrawRectangle(rect);
//	}
//
// `src/render/text/compound.cpp`, Zeilen 57–62. `DRAW_ACTIVE` ist ausdrücklich
// „draw active editor stuff, such as hidden separators and atom highlights"
// (`src/data/draw_what.hpp`, Zeile 23) und beim Ausgeben eines Kartenbildes
// nicht gesetzt.
//
// Fassung f9d9356 vom 2021-07-06.
//
// # Warum das hier steht
//
// `atom` gehört in [gedeuteteAuszeichnungen] — MSE deutet es. Damit fiel jedes
// `<atom-sep>` in den Zweig „MSE tut etwas, wir noch nicht", und über die
// Stichprobe von #555
// standen **50** solcher Befunde. Sie sind Befunde für richtiges Verhalten:
// der Inhalt wird gezeichnet, die Auszeichnung wirkt nicht.
//
// **`<atom-sep>` ist der Regelfall, nicht die Ausnahme.**
// `magic.mse-game/language_map` legt jedem Sprachraum ein
// `subtype_separator : "<atom-sep>, </atom-sep>"` bei; die Typzeile jeder Karte
// mit zwei Untertypen trägt es.
//
// # Und was Befund bleibt
//
// `atom-param` und `atom-kwpph` klammern ihren Inhalt in ⟨ ⟩ und färben
// beziehungsweise mildern ihn (`element.cpp`, Zeilen 237–241 und 298–306,
// Fassung f9d9356 vom 2021-07-06). Das fehlt hier, und der Befund sagt es.
func istSchlichtesAtom(name string) bool {
	if trifft(name, "atom-param") || trifft(name, "atom-kwpph") {
		return false
	}

	return trifft(name, "atom")
}

// istUebergangen sagt, ob MSE eine Auszeichnung gar nicht deutet.
//
// # Die Liste ist abschließend, und sie steht im Altcode
//
// `TextElements::fromString` (`src/render/text/element.cpp`, Zeilen 78–200,
// Fassung f9d9356 vom 2021-07-06) ist **eine** `else if`-Kette. Was sie nicht
// trifft, fällt durch — kommentarlos, und der Inhalt bleibt Text. Damit ist
// die Kette keine Aufzählung von Beispielen, sondern die vollständige Liste
// dessen, was MSE deutet.
//
// **Ein Befund für richtiges Verhalten verdeckt die echten** (§8) — auch
// `kw-a`, `nospellcheck`, `word-list-*` und `param-*` übergeht MSE
// (#429).
//
// # Was gemeldet bleibt
//
// Eine Auszeichnung, die in der Kette **steht** und die dieses Programm noch
// nicht umsetzt — `soft-line`, `margin:`, `li`, `align:`, `code`, `size:`.
// Dort fehlt Verhalten, und der Befund sagt das.
func istUebergangen(name string) bool {
	for _, gedeutet := range gedeuteteAuszeichnungen {
		if trifft(name, gedeutet) {
			return false
		}
	}

	return true
}

// gedeuteteAuszeichnungen ist MSEs Kette, auf ihre kürzesten Präfixe gebracht.
//
// **Kürzeste Präfixe, weil die Endzeichenregel den Rest erledigt**:
// `atom-kwpph` und `atom-param` fallen unter `atom`, `code-kw` und `code-str`
// unter `code`, `soft-line` unter `soft`. `sep-soft` und `ref-param` stehen
// eigens da — `sep` und `ref` kennt die Kette nicht.
// **`bullet` steht dabei, obwohl nur `</bullet` in der Kette vorkommt.** MSE
// setzt dort den Rand des Aufzählungspunktes (`element.cpp:171–176`); das
// öffnende `<bullet>` fällt bei MSE durch. Da [grundform] den Schrägstrich
// abstreift, sind beide hier derselbe Name — und „MSE tut etwas, wir noch
// nicht" ist die richtige Auskunft für beide.
var gedeuteteAuszeichnungen = []string{
	"align", "atom", "b", "bullet", "code", "color", "error", "font", "i",
	"li", "line", "margin", "ref-param", "sep-soft", "size", "soft", "sym",
}

// trifft bildet MSEs [is_tag] nach.
//
//	if (it == end || !is_tag_end_char(*it)) return false;
//
// `src/util/tagged_string.cpp`, Zeilen 165–171; `is_tag_end_char` sind `>`,
// `-`, `:` und das Leerzeichen (ebd. Zeilen 28–30).
//
// **Ohne diese Bedingung wäre jedes Präfix ein Fangnetz**: `b` träfe `bold`,
// `li` träfe `linear`, und die halbe Liste des Editors gälte als gedeutet. Das
// `>` entspricht hier dem Ende des Namens — die Klammern sind schon ab.
//
// Fassung f9d9356 vom 2021-07-06.
func trifft(name, praefix string) bool {
	if !strings.HasPrefix(name, praefix) {
		return false
	}

	if len(name) == len(praefix) {
		return true
	}

	switch name[len(praefix)] {
	case '-', ':', ' ':
		return true
	}

	return false
}

// istLinie sagt, ob eine Auszeichnung einen **Block** schließt.
//
// **`<line>` und nicht `<soft-line>`.** MSE prüft beide, aber in dieser
// Reihenfolge: `<line` zuerst, `<soft-line` danach — und weil `is_tag` ein
// Präfix vergleicht, trifft `<line` das Wort `<soft-line` **nicht** (es beginnt
// anders). Umgekehrt schon: `<soft-line>` fiele auf `<line`, stünde es davor.
// Die Reihenfolge ist die Aussage (`src/render/text/element.cpp`, Zeilen 84–87,
// Fassung f9d9356 vom 2021-07-06).
//
// Ein weicher Zeilenumbruch schneidet **keinen** Block; er ist hier deshalb
// nicht gedeutet und fällt wie bisher als unbekannte Auszeichnung heraus.
func istLinie(name string) bool {
	return trifft(name, "line")
}
