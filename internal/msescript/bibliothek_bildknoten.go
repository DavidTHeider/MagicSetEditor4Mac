// **Wie ein Bildknoten entsteht** — abgetrennt am 2026-09-02, weil
// `bibliothek_bild.go` mit den drei Funktionen aus Phase 4.9 über die
// Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Daneben in `bibliothek_bild.go` steht, **was ein Bild hier ist**: die Art,
// der Baum, die Umwandlung eines Werts in einen. Hier steht, **wie** ein
// Knoten daraus entsteht — welche Argumente Kinder werden und welche Werte.

package msescript

import "context"

// verknuepfeBild baut eine Verknüpfung, die ein Bild aus einem Bild macht.
//
// # Das Bild kommt stellungs- **oder** namensgebunden
//
// `recolor_image(input, red: …)` und `recolor_image(image: input, red: …)`
// stehen beide im Pack. MSE nimmt das erste Stellungsargument als Bild, wenn
// kein `image:` dasteht.
//
// **Ohne Bild ist es ein Fehler, kein leeres Ergebnis** (§8): eine Verknüpfung
// ohne Vorlage hätte nichts zu verknüpfen, und ein stilles `nichts` machte aus
// einer kaputten Vorlage eine leere Karte.
func verknuepfeBild(name string, bildnamen ...string) Eingebaut {
	return func(_ context.Context, ruf Ruf) (Wert, error) {
		eintraege := make([]Eintrag, 0, len(ruf.Benannt)+1)
		bilder := 0

		for _, gesucht := range bildnamen {
			wert, da := ruf.Benannter(gesucht)

			// **Ein unbenanntes Argument bindet `input` — und sonst
			// nichts** (#555).
			//
			//	} else {
			//	  // implicit "input" argument
			//	  arguments.push_back(SCRIPT_VAR_input);
			//	  parseOper(input, script, PREC_SEQ);
			//	}
			//
			// `src/script/parser.cpp`, Zeilen 889–893, Fassung f9d9356 vom
			// 2021-07-06. In MSE gibt es **keine** Stellungsargumente: jedes
			// unbenannte wird zu `input:`, auch das dritte. Und was eine
			// Bildfunktion liest, liest sie beim Namen —
			// `SCRIPT_PARAM(GeneratedImageP, image1)` ist
			// `ctx.getVariable("image1")` (`functions/util.hpp`, Zeilen
			// 98–101).
			//
			// **Hier stand `stelle == 0`**, also „das erste unbenannte
			// Argument ist das erste Bild". Zusammen mit dem Nachschub des
			// umgebenden `input` (`auswerter_aufruf.go`) hieß das: ein
			// `linear_blend(x1:0, y1:0, x2:1, y2:0)`, dessen Bilder aus dem
			// umgebenden Bereich kommen, nahm als `image1` das ambiente
			// `input`. In `magic-m15-split-fusable` war das auf `Alive` die
			// Zeichenkette `"true"`, und die linke Brückenhälfte fehlte.
			//
			// **Gezählt im Pack** über `style`, `script` und die
			// Blendmodule: `set_mask` 27-mal benannt und **null**-mal
			// unbenannt, `combine_blend` 63/0, `linear_blend` 535 benannt.
			// Einzig `recolor_image` steht durchweg unbenannt — 64-mal —,
			// und sein Parameter heißt bei MSE `input`.
			if !da && gesucht == "input" && len(ruf.Stellung) > 0 {
				wert, da = ruf.Stellung[0], true
			}

			// **Was der Aufruf nicht nennt, steht in der Umgebung**
			// (#473).
			//
			//	#define SCRIPT_PARAM(Type, name) SCRIPT_PARAM_N(Type, _(#name), name)
			//	#define SCRIPT_PARAM_N(Type, str, name) \
			//	    Type name = from_script<Type>(ctx.getVariable(str), str)
			//
			// `src/script/functions/util.hpp`, Zeilen 98–101, Fassung f9d9356
			// vom 2021-07-06. `SCRIPT_PARAM` liest aus dem **Kontext**, nicht
			// aus der Argumentliste — ein benanntes Argument bindet dort eine
			// Variable, aber eine anderswo gebundene tut es genauso.
			//
			// **Das Pack lebt davon.** `magic-blends.mse-include/new-blends`
			// schreibt `linear_blend := { do_linear_blend(x1: 0, x2: 0, y1:
			// x1, y2: x2) }` — eine Teilanwendung, deren `image1` und
			// `image2` aus dem umgebenden Bereich kommen. Dasselbe steht in
			// `magic-m15-split-fusable` und drei weiteren Stilen.
			//
			// **Hier wurde ein fehlendes Kind übersprungen**, und geprüft war
			// nur, ob **irgendein** Bild dastand. Ein `masked_blend` mit zwei
			// von drei Kindern kam damit durch, ohne dass jemand in der
			// Umgebung nachgesehen hätte.
			if !da && ruf.Umgebung != nil {
				wert, da = ruf.Umgebung.Suche(gesucht)
			}

			if !da {
				return Nichts, ruf.fehlerAn(name + " braucht " + gesucht + ":")
			}

			// **Die leere Zeichenkette und `nil` sind ein leeres Bild.**
			//
			//	GeneratedImageP toImage() const override {
			//	  if (value.empty()) {
			//	    return make_intrusive<BlankImage>();
			//	  } else {
			//	    return make_intrusive<PackagedImage>(value);
			//	  }
			//	}
			//
			// `ScriptString::toImage`, `src/script/value.cpp`, Zeilen
			// 369–375, und für `nil` dasselbe in Zeile 450. Fassung f9d9356
			// vom 2021-07-06.
			//
			// **Hier und nicht in [Wert.AlsBild]**: MSE wandelt an genau
			// dieser Stelle um — `SCRIPT_PARAM(GeneratedImageP, image1)`
			// ruft `toImage()`. Der `image:`-Block eines Feldstils ist ein
			// anderer Weg: dort heißt „nichts herausgekommen" **„nimm
			// `default:` oder die Wahlbilder"**, und ein leeres Bild
			// verdeckte sie
			// (#509).
			if wert.Art == ArtNichts || (wert.Art == ArtTextwert && wert.Text == "") {
				eintraege = append(eintraege,
					Eintrag{Name: gesucht, Wert: Wert{Art: ArtBild, Text: "blank_image"}})
				bilder++

				continue
			}

			if _, istBild := wert.AlsBild(); !istBild {
				return Nichts, ruf.fehlerAn(
					name + " bekam für " + gesucht + " kein Bild, sondern " + string(wert.Art))
			}

			eintraege = append(eintraege, Eintrag{Name: gesucht, Wert: alsBildwert(wert)})
			bilder++
		}

		for _, benannt := range ruf.Benannt {
			if istBildname(benannt.Name, bildnamen) {
				continue
			}

			eintraege = append(eintraege, benannt)
		}

		return Wert{Art: ArtBild, Text: name, Eintraege: eintraege}, nil
	}
}

// istBildname sagt, ob ein Argument schon als Kind übernommen wurde.
func istBildname(name string, bildnamen []string) bool {
	for _, gesucht := range bildnamen {
		if name == gesucht {
			return true
		}
	}

	return false
}

// Dateibild macht aus einem Pfad einen Bildwert.
//
// **Für Aufrufer außerhalb der Bibliothek.** `internal/service` legt damit den
// Wert eines Bildfeldes in `card` — bei MSE ist das kein Dateiname, sondern
// ein `ImageValueToImage` (`src/data/field/image.cpp`, Zeilen 49–53, Fassung
// f9d9356 vom 2021-07-06)
// (#440).
func Dateibild(pfad string) Wert { return alsBildwert(Text(pfad)) }

// alsBildwert macht aus einer Zeichenkette ein Blatt und lässt ein Bild stehen.
func alsBildwert(wert Wert) Wert {
	if wert.Art == ArtBild {
		return wert
	}

	return Wert{Art: ArtBild, Text: "datei", Eintraege: []Eintrag{{
		Name: "pfad", Wert: Text(wert.Text),
	}}}
}

// `to_image` reicht durch.
//
//	SCRIPT_FUNCTION(to_image) {
//	  SCRIPT_PARAM_C(GeneratedImageP, input);
//	  return input;
//	}
//
// `src/script/functions/image.cpp`, Zeilen 26–29, Fassung f9d9356 vom
// 2021-07-06. **Drei Zeilen, und kein Knoten**: die Umwandlung geschieht in
// `from_script<GeneratedImageP>`, das Ergebnis ist das Bild selbst. Ein
// eigener Baumknoten dafür wäre eine Verknüpfung, die nichts verknüpft.
//
// # Wie er gefunden wurde
//
// Der Saga-Stil schreibt `crop(to_image(…/textmask.png), …)` für seine
// Textmaske. Solange ein Befund neben einem Bild verlorenging
// (#473),
// zeichnete das Programm die **unbeschnittene** Maske und meldete nichts.
// Drei Aufrufe im vollen Pack.
func zuBild(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_image"); err != nil {
		return Nichts, err
	}

	wert := ruf.Stellung[0]
	if _, istBild := wert.AlsBild(); !istBild {
		return Nichts, ruf.fehlerAn(
			"to_image bekam kein Bild, sondern " + string(wert.Art))
	}

	return alsBildwert(wert), nil
}

// eingebautesBild ist `built_in_image(name)`.
//
// **Ein Erzeuger und kein Umformer** — siehe die Registrierung oben. Was ein
// eingebauter Name bedeutet, weiß der Rasterer; hier steht er als benannter
// Wert im Baum, wo auch jedes andere nicht-bildliche Argument steht.
func eingebautesBild(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("built_in_image"); err != nil {
		return Nichts, err
	}

	return Wert{
		Art:  ArtBild,
		Text: "built_in_image",
		Eintraege: []Eintrag{
			{Name: "name", Wert: Text(ruf.Stellung[0].AlsText())},
		},
	}, nil
}
