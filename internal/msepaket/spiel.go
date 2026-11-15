package msepaket

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// feldarten ordnet die Werte von `type:` den Arten der Domäne zu.
//
// Die Namen stammen aus den Paketen, nicht von uns (`docs/FORMATE.md`). Was
// hier fehlt, wird `KindUnknown` und als Text behandelt — eine Art zu erfinden,
// die kein Paket kennt, hilft niemandem, und eine unbekannte abzuweisen hieße,
// ein Spiel nicht öffnen zu können, weil es ein Feld mehr hat als wir kennen.
var feldarten = map[string]domain.FieldKind{
	"text":    domain.KindText,
	"choice":  domain.KindChoice,
	"image":   domain.KindImage,
	"color":   domain.KindColor,
	"boolean": domain.KindBoolean,
	"info":    domain.KindInfo,

	// **`multiple choice` ist eine eigene Art**, keine Spielart von `choice`
	// (#556).
	// Bis dahin fiel sie hier durch und wurde als Text gereicht — samt der
	// Reihenfolge, in der sie in der Datei stand.
	"multiple choice": domain.KindMultipleChoice,

	// **`package choice` ist die Wahl unter installierten Paketen** (#632).
	// Bis dahin fiel sie hier durch und wurde als Text gereicht — 627
	// Vorkommen in 349 von 376 Stilpaketen des Packs (#630).
	"package choice": domain.KindPackageChoice,

	// **`symbol` ist eine Zeichnung im Paket** (#1339). Bis zum 2026-08-31
	// fiel sie hier durch — auch an den sechs Spielpaketen des Packs, deren
	// `set field: symbol` damit als Text dastand.
	"symbol": domain.KindSymbol,
}

// Spielmodell liest ein Spielpaket in die Domäne.
//
// Die Befunde des Pakets reisen mit: ein Feld ohne Namen ist ein Vorbehalt,
// kein Abbruch (AGENTS.md §8).
func (p *Paket) Spielmodell() (domain.Game, []Befund) {
	// **Das init script kommt mit.** Die Feldskripte des Spiels rufen
	// Funktionen, die es anlegt — ohne sie rechnet `full name` nicht (#1238).
	spiel := domain.Game{Name: p.KurzerName, Skript: p.Initskript()}

	var befunde []Befund

	for _, knoten := range p.Kopf.Alle("card field") {
		feld, befund := feldAus(knoten, "Kartenfeld")
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		spiel.Fields = append(spiel.Fields, feld)
	}

	// **Und die Setfelder**: `magic.mse-game/set_fields` führt 89 davon, und
	// die Stile lesen sie 1.400-mal — `set.symbol` allein 420-mal. Ohne die
	// Erklärung kennte die Umgebung nur, was ein Set **ausgefüllt** hat; ein
	// leer gelassenes Feld gäbe nichts statt einer leeren Zeichenkette
	// (#321).
	for _, knoten := range p.Kopf.Alle("set field") {
		feld, befund := feldAus(knoten, "Setfeld")
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		spiel.SetFields = append(spiel.SetFields, feld)
	}

	return spiel, befunde
}

// aufschriften liest einen `localized caption:`-Block — je Sprachkürzel eine
// Zeile.
//
//	localized caption:
//		ch-s: 标题
//		ch-t: 標題
//
// **`nil` bei einem fehlenden Block**, damit ein Feld ohne Übersetzungen von
// einem mit leerem Block unterscheidbar bleibt: MSEs `LocalizedString` fragt
// `translations.empty()`, bevor es überhaupt nach der Sprache sieht
// (`src/data/locale.cpp`, Zeilen 138–141, Fassung f9d9356 vom 2021-07-06).
func aufschriften(knoten *Knoten) map[string]string {
	if knoten == nil {
		return nil
	}

	gelesen := map[string]string{}

	for _, kind := range knoten.Kinder {
		if kind.Zeilenart != ArtEintrag || kind.Schluessel == "" {
			continue
		}

		gelesen[kind.Schluessel] = kind.Wert
	}

	if len(gelesen) == 0 {
		return nil
	}

	return gelesen
}

// feldAus macht aus einem `card field`- oder `set field`-Block eine
// Felddefinition. `art` benennt die Sorte für den Befund.
//
// **Dieselbe Gestalt, derselbe Leser**: beide Blöcke tragen `type:`, `name:`,
// `default:` und `choice:`. Zwei Leser danebenzustellen hieße, eine spätere
// Erkenntnis zweimal einzupflegen.
func feldAus(knoten *Knoten, art string) (domain.Field, *Befund) {
	name := knoten.Ersten("name")
	if name == nil || name.Wert == "" {
		return domain.Field{}, &Befund{
			Stelle:  knoten.Stelle,
			Meldung: "ein " + art + " ohne `name` — es lässt sich nichts zuordnen",
		}
	}

	feld := domain.Field{
		Name:    name.Wert,
		Kind:    artAus(knoten),
		Default: wertOderLeer(knoten.Ersten("default")),
		Initial: wertOderLeer(knoten.Ersten("initial")),

		// **Auch ein Kartenfeld hat einen Leernamen** (#2006) — `empty
		// choice:` steht in `magic.mse-game` an sieben davon.
		Leerwahl:      wertOderLeer(knoten.Ersten("empty choice")),
		Skript:        skripttextOderLeer(knoten.Ersten("script")),
		Sortierskript: skripttextOderLeer(knoten.Ersten("sort script")),

		// **Die Form des Feldes** — wofür der Karteneditor das Steuerelement
		// wählt (#560).
		Mehrzeilig: wertOderLeer(knoten.Ersten("multi line")) == "true",

		// **Ohne Angabe wahr** — MSE setzt `editable` im Konstruktor auf
		// `true` (`src/data/field.cpp`, `Field::Field`, Fassung f9d9356 vom
		// 2021-07-06). 24 der 249 Kartenfelder von `magic.mse-game` setzen es
		// auf `false` (#708).
		Editierbar: wertOderLeer(knoten.Ersten("editable")) != "false",

		// **Ohne Angabe wahr** — MSE setzt `save_value` im Konstruktor auf
		// `true` (`src/data/field.cpp`, `Field::Field`, Fassung f9d9356 vom
		// 2021-07-06). Nur wer ausdrücklich `false` schreibt, meint es (#696).
		Gespeichert:  wertOderLeer(knoten.Ersten("save value")) != "false",
		Beschreibung: wertOderLeer(knoten.Ersten("description")),
		// **Die Aufschrift, nicht der Bezeichner** — MSE beschriftet damit das
		// Eingabefeld (`src/gui/control/native_look_editor.cpp`, Zeile 42,
		// Fassung f9d9356 vom 2021-07-06). Abgeleitet wird in der Domäne
		// ([domain.Field.Beschriftet]); hier steht, was dasteht.
		Beschriftung:   wertOderLeer(knoten.Ersten("caption")),
		Uebersetzungen: aufschriften(knoten.Ersten("localized caption")),

		// **Und der Listenname** — über `magic.mse-game` das einzige Deutsch,
		// das ein Kartenfeld mitbringt: 243-mal `card list name:`, 25-mal
		// `localized card list name:` mit `de:`, und **kein einziges**
		// `caption:` (#699).
		Listenname:           wertOderLeer(knoten.Ersten("card list name")),
		Listenuebersetzungen: aufschriften(knoten.Ersten("localized card list name")),

		// **Was das Spiel über die Spalte sagt** (#1168).
		//
		// **Gemessen über `magic.mse-game`** und über diesen Leser —
		// `listenspaltenfremd_test.go` hält die Zahlen: 243 Felder mit einer
		// Stelle und einer Breite, elf sichtbar, sechs verboten, elf rechts
		// ausgerichtet. Ein `grep` über die **Dateien** des ganzen Packs zählt
		// anderes (284 Zeilen `card list column:` über alle Spiele); die
		// Reichweite einer Messung ist die Reichweite ihrer Aussage (§10).
		Listenangaben: domain.Listenangaben{
			Listenstelle: ganzzahlOderNull(knoten.Ersten("card list column")),
			Listenbreite: ganzzahlOderNull(knoten.Ersten("card list width")),

			// **Ohne Angabe unsichtbar** — MSE setzt `card_list_visible` im
			// Konstruktor auf `false` (`src/data/field.cpp`, `Field::Field`,
			// Fassung f9d9356 vom 2021-07-06). Die meisten Spalten sind
			// vorhanden und **nicht** von Anfang an zu sehen.
			Listensichtbar: wertOderLeer(knoten.Ersten("card list visible")) == "true",

			// **Ohne Angabe erlaubt** — dort steht `card_list_allow(true)`.
			Listenerlaubt: wertOderLeer(knoten.Ersten("card list allow")) != "false",

			Listenausrichtung: domain.ListenausrichtungAus(
				wertOderLeer(knoten.Ersten("card list alignment"))),
		},
	}

	// **Ein `boolean` ist ein Auswahlfeld mit zwei festen Wahlen.**
	//
	//	BooleanField::BooleanField() {
	//	  choices->choices.push_back(make_intrusive<Choice>(_("yes")));
	//	  choices->choices.push_back(make_intrusive<Choice>(_("no")));
	//	  choices->initIds();
	//	}
	//
	// `src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom
	// 2021-07-06. Sie stehen im Paket nicht als `choice:` — sie gehören zum
	// **Typ**, und deshalb stehen sie hier und nicht in `internal/service`.
	//
	// **Ohne `initial:` gilt damit die erste — `yes`.** Über
	// `Full-Magic-Pack` haben 15 der 27 `boolean`-Felder kein `initial:`; sie
	// standen bei uns alle auf „nein", wo MSE „ja" liest. Unter ihnen
	// `shorten types for rarity`, an dem die Breite des Seltenheitssymbols
	// hängt — es war auf jeder Karte null
	// (#456).
	if feld.Kind == domain.KindBoolean {
		feld.Choices = []string{"yes", "no"}
	}

	for _, moeglich := range knoten.Alle("choice") {
		feld.Choices = append(feld.Choices, wahlname(moeglich))

		// **Eine Wahl eines Farbfeldes hat einen Namen und eine Farbe** — in
		// zwei Formen:
		//
		//	choice:
		//		name: white
		//		color: rgb(255,255,255)
		//	choice: white
		//
		// MSE liest die einzeilige Form als Namen **und** Farbe, weil es
		// `parse_color` auf den Namen anwendet (`ColorField::Choice`,
		// `src/data/field/color.cpp`, Zeilen 41–54, Fassung f9d9356 vom
		// 2021-07-06). Die erste Wahl ist die Vorgabe des Feldes
		// (#270).
		//
		// **Gedeutet wird hier nichts** — was `white` bedeutet, weiß
		// `internal/service`; dieser Leser gibt heraus, was dasteht (§2).
		// **Nur an einem Farbfeld.** An einem Auswahlfeld stünde in dieser
		// Liste jeder Wahlname, und ein späterer Leser hielte ihn für eine
		// Farbangabe.
		if feld.Kind == domain.KindColor {
			feld.Wahlfarben = append(feld.Wahlfarben, wahlfarbe(moeglich))
		}
	}

	feld.Statistik, feld.Statistikfarben = statistikAus(knoten, feld.Kind)

	return feld, nil
}

// artAus liest `type:` und bildet es auf eine Feldart ab.
func artAus(knoten *Knoten) domain.FieldKind {
	art := knoten.Ersten("type")
	if art == nil {
		return domain.KindUnknown
	}

	return feldarten[art.Wert]
}

// skripttextOderLeer gibt den Text eines Skriptblocks.
//
// **Nicht [wertOderLeer].** Ein `script:` steht als eingerückter Block, und
// dessen Text liegt nicht im Wert des Knotens, sondern in seinen Zeilen — es ist
// dieselbe Form wie bei `init script:`
// (#345).
func skripttextOderLeer(knoten *Knoten) string {
	if knoten == nil {
		return ""
	}

	// **Der Wert entscheidet, welche Form dasteht.**
	//
	// `magic.mse-game/card_fields` schreibt Wertskripte in zwei Formen:
	// einzeilig (`script: card_color_filter(value)`) und als eingerückten
	// Block darunter. Bei der ersten steht das Skript im **Wert** des Knotens,
	// bei der zweiten in seinen Zeilen.
	//
	// **Und die Zeilen sind bei der ersten Form nicht leer.** Auf ein
	// einzeiliges `script:` folgen im Pack regelmäßig Kommentarzeilen
	// (`##### Name line`), und [Knoten.Skripttext] gibt sie mit. Wer den Block
	// bevorzugte, zerteilte einen reinen Kommentar und bekäme „hier fehlt ein
	// Ausdruck" — 3.990-mal über die 190 Stile
	// (#345).
	if knoten.Wert != "" {
		return knoten.Wert
	}

	return knoten.Skripttext()
}

func wertOderLeer(knoten *Knoten) string {
	if knoten == nil {
		return ""
	}

	return knoten.Wert
}

// Initskript gibt den Text des `init script:`-Blocks im Kopf dieses Pakets.
//
// # Wozu — und warum am Paket und nicht am Spielmodell
//
// Ein **Spielpaket** legt die Funktionen an, die jeder Stil dieses Spiels
// benutzt: `magic.mse-game` führt `init script: include file: script`, und aus
// dieser Datei kommen unter anderem `additional_credit_brush_image` (458
// Vorkommen) und `pt_font_size` (75). Ohne sie sind die zugehörigen Feldstile
// unausrechenbar
// (#195).
//
// **`domain.Game` bekommt das nicht.** Es beschreibt, welche Felder eine Karte
// hat — eine Aussage über Daten. Ein Skripttext ist eine Aussage über
// Darstellung, und `internal/domain` importiert nichts, was ihn auswerten
// könnte (§2).
//
// **Es steht am Paket und nicht am Spielmodell**, weil auch ein Stilpaket einen
// solchen Block trägt; dieselbe Frage, dieselbe Antwort.
//
// Leer heißt: es gibt keinen. Vor [Paket.ModuleEinbinden] fehlen darin die
// `include file:`-Module.
func (p *Paket) Initskript() string {
	knoten := p.Kopf.Ersten("init script")
	if knoten == nil {
		return ""
	}

	return knoten.Skripttext()
}

// wahlfarbe gibt die Farbangabe einer `choice:` — aus dem Block oder dem Wert.
//
// Siehe [feldAus] für die Herkunft: MSE liest `choice: white` als Namen und
// Farbe zugleich.
func wahlfarbe(moeglich *Knoten) string {
	if farbe := moeglich.Ersten("color"); farbe != nil {
		return farbe.Wert
	}

	return moeglich.Wert
}
