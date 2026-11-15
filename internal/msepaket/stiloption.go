package msepaket

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Das Lesen einer einzelnen Stiloption.
//
// **Eigene Datei, und der Schnitt folgt der Frage** (§2): `stil.go` beantwortet
// „was ist ein Stil", hier steht „wie liest man eine seiner Optionen".

// stiloptionAus macht aus einem `styling field`-Block eine Stiloption.
//
// **Dieselbe Gestalt wie `card field`, und trotzdem eigener Code**: die
// Beschreibung führt eine Felddefinition gar nicht, und die Wahlen bedeuten
// etwas anderes. Ein gemeinsamer Leser hätte für jedes Feld zwei Bedeutungen.
//
// **Eine Stiloption führt `initial:` und `default:`**, und MSE liest beides —
// `initial:` im Konstruktor, `default:` als Skript darüber
// (`src/data/field/color.cpp`, Zeilen 82–101, Fassung f9d9356 vom 2021-07-06;
// #590).
func stiloptionAus(knoten *Knoten) (domain.StyleField, *Befund) {
	name := knoten.Ersten("name")
	if name == nil || name.Wert == "" {
		return domain.StyleField{}, &Befund{
			Stelle:  knoten.Stelle,
			Meldung: "eine Stiloption ohne `name` — sie ließe sich keinem Wert zuordnen",
		}
	}

	option := domain.StyleField{
		Name:        name.Wert,
		Kind:        artAus(knoten),
		Description: wertOderLeer(knoten.Ersten("description")),
		Initial:     wertOderLeer(knoten.Ersten("initial")),
		Default:     wertOderLeer(knoten.Ersten("default")),
		Match:       wertOderLeer(knoten.Ersten("match")),
		Required:    pflichtwahl(knoten),
		EmptyName:   leerername(knoten),
		EigeneFarbe: eigenefarbe(knoten),
		Leerwahl:    wertOderLeer(knoten.Ersten("empty choice")),
	}

	// **Ein `boolean` ist ein Auswahlfeld mit zwei festen Wahlen** —
	// **dieselbe Regel wie am Spielfeld** (`spiel.go`), und sie fehlte hier.
	//
	//	BooleanField::BooleanField() {
	//	  choices->choices.push_back(make_intrusive<Choice>(_("yes")));
	//	  choices->choices.push_back(make_intrusive<Choice>(_("no")));
	//	}
	//
	// `src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom
	// 2021-07-06. Sie stehen im Paket nicht als `choice:` — sie gehören zum
	// **Typ**.
	//
	// **#456
	// hat das für die Felder des Spielpakets gebaut und hier stehen lassen.**
	// Das ist die Klasse, die §10 anstreicht: nicht den Fall behandeln,
	// sondern die Klasse. Über das Pack sind es **1.489** `styling field:` vom
	// Typ `boolean`, davon **378 ohne `initial:`**, in **178 von 365**
	// Stilpaketen — alle standen bei uns auf „nein", wo MSE „ja" liest
	// (#531).
	if option.Kind == domain.KindBoolean {
		option.Choices = []string{"yes", "no"}
	}

	for _, moeglich := range knoten.Alle("choice") {
		option.Choices = append(option.Choices, wahlname(moeglich))

		// **Dieselbe Gestalt wie beim `card field:`** — siehe [wahlfarbe].
		// **Nur an einem Farbfeld.** An einem Auswahlfeld stünde in dieser
		// Liste jeder Wahlname, und ein späterer Leser hielte ihn für eine
		// Farbangabe.
		if option.Kind == domain.KindColor {
			option.Wahlfarben = append(option.Wahlfarben, wahlfarbe(moeglich))
		}
	}

	return option, nil
}

// pflichtwahl liest `required:` einer [domain.KindPackageChoice].
//
// **Ohne Angabe gilt Pflicht.** So steht es im Kopf von
// `PackageChoiceField` (`src/data/field/package_choice.hpp`, Zeile 27, Fassung
// f9d9356 vom 2021-07-06):
//
//	PackageChoiceField() : required(true), empty_name(_("none")) {}
//
// **Und nur `false` hebt sie auf.** Ein `required: vielleicht` ist keine
// Verneinung; wer jeden Wert außer `true` als „nein" läse, machte aus einem
// Tippfehler eine Erlaubnis.
func pflichtwahl(knoten *Knoten) bool {
	return wertOderLeer(knoten.Ersten("required")) != "false"
}

// leerername liest `empty name:` — die Beschriftung des leeren Eintrags.
//
// **Ohne Angabe `none`**, aus derselben Zeile wie [pflichtwahl]. Nicht
// übersetzt (§9): der Name kommt aus dem Paket, und die Vorgabe ist der Wert,
// den MSE selbst setzt.
func leerername(knoten *Knoten) string {
	if name := wertOderLeer(knoten.Ersten("empty name")); name != "" {
		return name
	}

	return "none"
}

// eigenefarbe liest `allow custom:` einer [domain.KindColor].
//
// **Ohne Angabe erlaubt.** So steht es im Kopf von `ColorField`
// (`src/data/field/color.cpp`, Zeilen 15–18, Fassung f9d9356 vom 2021-07-06):
//
//	ColorField::ColorField()
//	  : allow_custom(true)
//	  , default_name(_("Default"))
//	{}
//
// **Und nur `false` schaltet sie ab** — dieselbe Regel wie bei
// [pflichtwahl], aus demselben Grund: ein Tippfehler soll keine Erlaubnis
// werden, und hier auch kein Verbot.
func eigenefarbe(knoten *Knoten) bool {
	return wertOderLeer(knoten.Ersten("allow custom")) != "false"
}
