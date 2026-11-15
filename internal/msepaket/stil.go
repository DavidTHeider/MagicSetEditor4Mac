package msepaket

import (
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Stil ist das Darstellungsmodell eines `.mse-style`.
//
// # Was hier roh bleibt
//
// **Jede Zahl in einem Stil darf ein Skript sein** (`docs/FORMATE.md`):
// `left: 0` und `left: { if … }` stehen nebeneinander in derselben Datei.
// Dieser Leser gibt beides so heraus, wie es dasteht — was `{ … }` bedeutet,
// weiß `msescript`, und wer es auswertet, ist `service`. Ein Paketleser, der
// Skripte auswertete, hätte zwei Aufgaben und ließe sich nur mit einer
// Skriptumgebung prüfen.
type Stil struct {
	Spiel string

	// Paket ist der Name des Pakets, aus dem dieser Stil stammt —
	// `magic-m15.mse-style`, nicht sein Pfad auf der Platte.
	//
	// **Ohne ihn ist ein einfacher Dateiname nicht auflösbar.** `mask:` und
	// `image:` tragen in den Stilen von `M15-Magic-Pack` beide Formen
	// nebeneinander: `border_mask.png` meint eine Datei im **eigenen** Paket,
	// `/magic-genevensis-00-main.mse-style/pride_mask.png` eine in einem
	// fremden. Wer die erste ohne diese Angabe weiterreicht, sucht sie in der
	// Paketwurzel.
	Paket string

	// KurzerName und VollerName sind `short name:` und `full name:` des
	// Pakets — die Vorlagen lesen `stylesheet.short_name`
	// (`Packaged`, `src/util/io/package.hpp`, Zeile 266, Fassung f9d9356 vom
	// 2021-07-06).
	KurzerName string
	VollerName string

	// Breite und Hoehe sind Bildpunkte bei DPI.
	Breite int
	Hoehe  int
	DPI    int

	// Regionen sind die `card region:` des Kopfes ([Kartenregion]).
	Regionen []Kartenregion

	// Initskript ist der Text des `init script:`-Blocks.
	//
	// **Er legt an, was die Feldskripte benutzen** — `template`, `mana_sort`,
	// `module_card`. Ohne ihn ließ sich von 356 `image:`-Angaben in
	// `M15-Magic-Pack` keine einzige ausrechnen
	// (#186).
	//
	// **Erst [Paket.ModuleEinbinden] macht ihn vollständig**: er beginnt in
	// echten Stilen mit `include file:`-Zeilen, und was die einsetzen, steht
	// vorher nicht darin.
	Initskript string

	// Felder stehen in **Zeichenreihenfolge** — nach `z index`, nicht nach der
	// Reihenfolge im Text.
	Felder []Feldstil

	// Symbolfassungen sind die `variation:`-Blöcke unter `set info style:` —
	// die benannten Fassungen des Setsymbols.
	Symbolfassungen []Symbolfassung

	// Zusatzfelder sind die `extra card field:`-Blöcke — **eigene Felder des
	// Stils**, die es im Spiel nicht gibt.
	//
	//	REFLECT(extra_card_fields);
	//	REFLECT_IF_READING { extra_card_style.init(extra_card_fields); }
	//	REFLECT(extra_card_style);
	//
	// `src/data/stylesheet.cpp`, Zeilen 116–125, Fassung f9d9356 vom
	// 2021-07-06. Daran hängt bei `magic-m15` der Kasten für Stärke und
	// Widerstandskraft — er ist kein Kartenfeld, sondern ein Feld der Vorlage
	// (#270).
	Zusatzfelder []domain.Field

	// Zusatzstile sind die `extra card style:`-Blöcke — die Maße dazu.
	Zusatzstile []Feldstil

	// Stiloptionen sind die `styling field:`-Blöcke des Pakets.
	//
	// **Sie gehören zum Set, nicht zur Karte** — MSE zeigt sie im Stilmenü, und
	// ein Feldskript liest sie über `styling.…`. Sie kommen aus dem Paket und
	// nicht aus einer Liste bei uns
	// (#41);
	// `magic-m15.mse-style` führt dreizehn davon, mit Beschreibung und
	// Vorgabewert.
	Stiloptionen []domain.StyleField

	// Optionsstile sind die `styling style:`-Blöcke — je Stiloption die
	// Angabe, **wie** sie aussieht
	// (#1225).
	//
	// # Warum eine zweite Liste und nicht ein Feld an [domain.StyleField]
	//
	// Weil es zwei Gegenstände sind, und MSE hält sie getrennt:
	//
	//	vector<FieldP> styling_fields;
	//	IndexMap<FieldP, StyleP> styling_style;
	//
	// `src/data/stylesheet.hpp`, Zeilen 47–50, Fassung f9d9356 vom
	// 2021-07-06. Die Erklärung einer Option — Name, Art, Wahlen — ist eine
	// Angabe der **Fachlichkeit** und steht in [domain]; wie sie gezeichnet
	// wird, ist eine des **Pakets** und steht hier. Genau dieselbe Trennung
	// wie [domain.Field] gegen [Feldstil].
	//
	// **Und nicht jede Option hat einen.** `magic-m15` erklärt dreizehn und
	// gibt vieren ein Bild.
	Optionsstile []Feldstil
}

// Optionsstil gibt die Sichtseite einer Stiloption; der zweite Wert sagt, ob
// es sie gibt.
//
// **Dieselbe Gestalt wie [Stil.Feld]**, und aus demselben Grund: ein leerer
// Feldstil ließe sich von einem erklärten nicht unterscheiden, und dann
// zeichnete eine Option ohne Bilder eine leere Bildreihe.
func (s Stil) Optionsstil(name string) (Feldstil, bool) {
	for stelle := range s.Optionsstile {
		if s.Optionsstile[stelle].Feld == name {
			return s.Optionsstile[stelle], true
		}
	}

	return Feldstil{}, false
}

// Feldstil sagt, wo ein Feld steht und wie es gezeichnet wird.
type Feldstil struct {
	Feld string

	// Stelle und Größe, roh. **`Breite` kann leer sein und `Rechts` gesetzt**:
	// `name:` in `magic-m15.mse-style` führt `left`, `right` und `height`, und
	// die Breite ergibt sich. Ein Leser, der `width` voraussetzt, liest dieses
	// Feld falsch.
	// Sichtbar ist `visible:` — leer heißt **sichtbar**, so wie MSEs
	// `visible(true)` (`src/data/field.cpp`, Zeile 106).
	//
	// Fassung f9d9356 vom 2021-07-06.
	Sichtbar string

	Links  string
	Oben   string
	Breite string
	Hoehe  string
	Rechts string
	Unten  string

	ZIndex int

	// Winkel ist `angle:` — die Drehung des Feldes in Grad, gegen den
	// Uhrzeigersinn. Leer heißt null.
	//
	// **Roh, weil er gerechnet sein darf**: im vollen Pack stehen 33 der 213
	// Angaben als `{if …}`
	// (#486).
	Winkel      string
	Ausrichtung string

	// Ecke und Rand gelten für ein **Farbfeld**: der Radius der runden Ecken
	// und die vier Breiten, die aus der Füllung einen Rahmen machen
	// ([Randangabe]). Roh — ausgerechnet wird in `internal/service`.
	Ecke string
	Rand Randangabe

	// AusrichtungSkript ist die Ausrichtung als **Skriptblock** —
	// `alignment:` mit einem `script:` darunter.
	//
	// **216 der 1.597 Ausrichtungen im Pack stehen so.** MSE führt
	// `alignment` als `Scriptable<Alignment>` (`src/data/field/text.cpp`,
	// Zeile 95, Fassung f9d9356 vom 2021-07-06); der Wert des Knotens ist
	// dann leer, und wer nur ihn läse, bekäme die Vorgabe „links oben"
	// (#345).
	AusrichtungSkript string
	Maske             string
	Darstellung       string

	// Bild ist das Skript, das ein Bild liefert (`image:`).
	Bild string

	// Wahlbilder ist der `choice images:`-Block — je Wahl ein Bild.
	//
	// # Der Mechanismus
	//
	// MSE baut daraus das Skript `[a: {…}, b: {…}][input]() or else nil`
	// (`src/data/field/choice.cpp:171`): eine Abbildung von Wahlname auf
	// Bildskript, indiziert mit dem **Wert des Feldes**.
	//
	// **Sie stehen im Stil, nicht im Spiel.** `magic-m15.mse-style:162` führt
	// sie; über die 190 Stile von `M15-Magic-Pack` hängen 342 Feldstile daran
	// (#286).
	//
	// Fassung f9d9356 vom 2021-07-06.
	Wahlbilder []Wahlbild

	// Standardbild ist `default:` — das Bild, wenn das Feld keines trägt.
	//
	// Es ist **nicht** dasselbe wie `Bild`: `image:` gilt immer, `default:` nur
	// dann, wenn die Karte für dieses Feld nichts hinterlegt hat. In
	// `magic-m15.mse-style` führt das Kunstfeld `default:
	// {default_image(card.card_color)}` — ein Ersatzbild nach Kartenfarbe.
	Standardbild string

	Schrift       Schrift
	Symbolschrift Symbolangabe

	// Polster ist `padding left/right/top/bottom` — der Rand innerhalb des
	// Feldes. **279-mal `padding top:`** in den 190 Stilen von
	// `M15-Magic-Pack`.
	Polster Polsterangabe

	// Zeilenhoehen sind `line height soft/hard/line` — je rund 290-mal im Pack.
	Zeilenhoehen Zeilenhoehenangabe

	// ImmerSymbole ist `always symbol:` — der Wert dieses Feldes besteht
	// **immer** aus Symbolen, nicht nur an `<sym>`-Stellen.
	//
	// **133-mal in den 190 Stilen von `M15-Magic-Pack`**, an den Manakosten
	// und nie am Regeltext. MSE entscheidet daran, ob ein Feld überhaupt in
	// Symbole zerlegt wird (`src/render/text/element.cpp`, Zeile 244, Fassung
	// f9d9356 vom 2021-07-06).
	//
	// **Ohne die Angabe muss man raten**, und das ging schief: bis zum
	// 2026-08-17 fragte `internal/service`, ob sich der ganze Wert in Symbole
	// zerlegen lässt — was zutrifft, sobald die Schrift einen Auffangcode
	// führt, und dann wäre auch der Regeltext eine Reihe von Bildern
	// (#293).
	ImmerSymbole bool

	// Abstand und Richtung sind `spacing:` und `direction:` — wie eine
	// Mehrfachwahl als Reihe ihre Wahlen legt (`MultipleChoiceStyle`,
	// `src/data/field/multiple_choice.cpp`, Zeilen 30–40, Fassung f9d9356 vom
	// 2021-07-06; #2297). Der Abstand darf ein Skript sein.
	Abstand  string
	Richtung string
}

// BreiteMM gibt die Kartenbreite in Millimetern.
//
// **Erst mit `card dpi` ist sie berechenbar** — 375 Bildpunkte bei 150 dpi sind
// 2,5 Zoll und damit 63,5 mm. Ohne Auflösung gibt es keine physische Größe, und
// dann ist das Ergebnis 0: eine erfundene Zahl wäre schlimmer.
func (s Stil) BreiteMM() float64 { return millimeter(s.Breite, s.DPI) }

// HoeheMM gibt die Kartenhöhe in Millimetern.
func (s Stil) HoeheMM() float64 { return millimeter(s.Hoehe, s.DPI) }

// millimeter rechnet Bildpunkte bei einer Auflösung in Millimeter um.
func millimeter(punkte, dpi int) float64 {
	if dpi <= 0 {
		return 0
	}

	return float64(punkte) / float64(dpi) * 25.4
}

// Feld gibt den Stil eines Felds; der zweite Wert sagt, ob es ihn gibt.
func (s Stil) Feld(name string) (Feldstil, bool) {
	for stelle := range s.Felder {
		if s.Felder[stelle].Feld == name {
			return s.Felder[stelle], true
		}
	}

	return Feldstil{}, false
}

// Stilmodell liest ein Stilpaket in sein Darstellungsmodell.
//
// **Die Befunde sind die des Lesens, nicht die des Pakets.** Ein Feldstil mit
// unlesbarem `z index` ist ein Vorbehalt, kein Abbruch (AGENTS.md §8) — und er
// kommt hier heraus.
//
// Was beim **Öffnen** oder beim [Paket.ModuleEinbinden] aufgefallen ist, steht
// dagegen an `Paket.Befunde` und wird von dort gelesen. Wer sie hier
// anhängte, meldete dasselbe zweimal, sobald jemand beide Wege liest.
func (p *Paket) Stilmodell() (Stil, []Befund) {
	stil := Stil{
		Paket:      paketname(p.Pfad),
		KurzerName: kopfwert(p.Kopf, "short name"),
		VollerName: kopfwert(p.Kopf, "full name"),

		Spiel:  kopfwert(p.Kopf, "game"),
		Breite: zahl(kopfwert(p.Kopf, "card width")),
		Hoehe:  zahl(kopfwert(p.Kopf, "card height")),
		DPI:    zahl(kopfwert(p.Kopf, "card dpi")),
	}

	stil.Initskript = p.Initskript()
	stil.Regionen = kartenregionenAus(p.Kopf)

	var befunde []Befund

	for _, knoten := range p.Kopf.Alle("styling field") {
		option, befund := stiloptionAus(knoten)
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		stil.Stiloptionen = append(stil.Stiloptionen, option)
	}

	stil.Zusatzfelder, stil.Zusatzstile, befunde = zusatzteile(p.Kopf, befunde)
	stil.Symbolfassungen = symbolfassungenAus(p.Kopf)

	// **Die Sichtseite der Stiloptionen** — dieselbe Gestalt wie `card
	// style:`, und deshalb derselbe Leser (#1225).
	for _, block := range p.Kopf.Alle("styling style") {
		stil.Optionsstile, befunde = feldstileAus(block, stil.Optionsstile, befunde)
	}

	for _, block := range p.Kopf.Alle("card style") {
		stil.Felder, befunde = feldstileAus(block, stil.Felder, befunde)
	}

	// **Nach `z index`, und stabil.** Zwei Felder mit derselben Zahl behalten
	// die Reihenfolge, in der sie dastehen — sonst hinge die Zeichenreihenfolge
	// an der Sortierung und wäre von Lauf zu Lauf eine andere.
	sort.SliceStable(stil.Felder, func(a, b int) bool {
		return stil.Felder[a].ZIndex < stil.Felder[b].ZIndex
	})

	return stil, befunde
}
