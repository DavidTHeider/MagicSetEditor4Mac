package msescript

// **Ein Bildskript gibt ein Bild zurück, keinen Dateinamen.**
//
// # Der Befund
//
// **Fünf Erzeuger und sechzehn Umformungen** — gezählt in
// `src/gfx/generated_image.hpp`, Fassung f9d9356 vom 2021-07-06: 23 Klassen,
// davon `GeneratedImage` die Oberklasse und `SimpleFilterImage` eine
// Zwischenklasse.
//
// **Erzeuger** (aus nichts ein Bild): `BlankImage`, `PackagedImage`,
// `BuiltInImage`, `SymbolToImage`, `ImageValueToImage`.
//
// **Umformungen** (aus Bildern ein Bild): `LinearBlendImage`,
// `MaskedBlendImage`, `CombineBlendImage` und die dreizehn
// `SimpleFilterImage` — `SetMask`, `SetAlpha`, `SetCombine`, `Saturate`,
// `Invert`, `Recolor`, `Recolor2`, `FlipHorizontal`, `FlipVertical`,
// `Rotate`, `Enlarge`, `Crop`, `DropShadow`.
//
// **965 der 2.729 Befunde** über die 190 Stile von `M15-Magic-Pack` lauten
// `der Name "recolor_image" ist hier nicht bekannt`; es ist der größte
// Einzelposten überhaupt
// (#270).
//
// # Warum `msescript` trotzdem nicht weiß, was ein Bild ist
//
// Es trägt einen **Baum** und deutet ihn nicht: eine Knotenart als Text, die
// Argumente als Einträge, die Kinder als Bildwerte darin. Dieselbe Bauform wie
// [Farbe] — was `rgb` zurückgibt, weiß der Auswerter auch nicht.
//
// §2 verbietet `msescript` die Kenntnis **dieses Projekts**, nicht das Tragen
// einer Struktur. Wer hier eine Rasterung fände, hätte den Verstoß gefunden.
func init() {
	// **Das Ausgangsbild heißt `input`**
	// (#483).
	//
	//	SCRIPT_PARAM_C(GeneratedImageP, input);
	//
	// `src/script/functions/image.cpp`, Fassung f9d9356 vom 2021-07-06;
	// `SCRIPT_PARAM_C` heißt „benannt oder stellungsweise". Im Pack schreiben
	// **643** Aufrufe `input:` und **kein einziger** `image:`.
	eintragen("recolor_image", verknuepfeBild("recolor_image", "input"))
	eintragen("masked_blend", verknuepfeBild("masked_blend", "light", "dark", "mask"))
	eintragen("linear_blend", verknuepfeBild("linear_blend", "image1", "image2"))
	eintragen("set_alpha", verknuepfeBild("set_alpha", "input"))
	eintragen("set_combine", verknuepfeBild("set_combine", "input"))
	eintragen("saturate", verknuepfeBild("saturate", "input"))
	eintragen("invert_image", verknuepfeBild("invert_image", "input"))
	eintragen("crop", verknuepfeBild("crop", "input"))
	eintragen("flip_horizontal", verknuepfeBild("flip_horizontal", "input"))
	eintragen("flip_vertical", verknuepfeBild("flip_vertical", "input"))
	eintragen("to_image", verknuepfeBild("to_image", "input"))
	eintragen("combine_blend", verknuepfeBild("combine_blend", "image1", "image2"))
	eintragen("enlarge", verknuepfeBild("enlarge", "input"))
	eintragen("to_image", zuBild)
	eintragen("drop_shadow", verknuepfeBild("drop_shadow", "input"))

	// **Die eine Ausnahme, und sie ist nachgelesen**: `set_mask` nennt sein
	// Ausgangsbild wirklich `image`
	// (`SCRIPT_PARAM(GeneratedImageP, image); SCRIPT_PARAM(…, mask);`).
	eintragen("set_mask", verknuepfeBild("set_mask", "image", "mask"))

	// **Die Drehung** — Werkstück A von Phase 4.9
	// (#1324).
	//
	//	SCRIPT_FUNCTION(rotate) {
	//	  SCRIPT_PARAM_C(GeneratedImageP, input);
	//	  SCRIPT_PARAM(Degrees, angle);
	//	  return make_intrusive<RotateImage>(input,deg_to_rad(angle));
	//	}
	//
	// `src/script/functions/image.cpp`, Zeilen 126–130, Fassung f9d9356 vom
	// 2021-07-06. **`angle` reist als benannter Eintrag mit** — was ein
	// Winkel in Bogenmaß ist, entscheidet der Rasterer und nicht die Sprache.
	eintragen("rotate", verknuepfeBild("rotate", "input"))

	// **`rotate_image` ist derselbe Zeiger:**
	//
	//	ctx.setVariable(_("rotate_image"), script_rotate);
	//
	// `image.cpp`, Zeile 227. Dieselbe Gestalt wie `replace_text` neben
	// `replace` — und die **Knotenart bleibt `rotate`**, damit der Rasterer
	// nicht zwei Namen für dieselbe Sache kennen muss.
	eintragen("rotate_image", verknuepfeBild("rotate", "input"))

	// **`built_in_image` ist ein Erzeuger**, kein Umformer:
	//
	//	SCRIPT_FUNCTION(built_in_image) {
	//	  SCRIPT_PARAM_C(String, input);
	//	  return make_intrusive<BuiltInImage>(input);
	//	}
	//
	// `image.cpp`, Zeilen 204–207. Es macht aus einem **Namen** ein Bild; der
	// Name reist deshalb als benannter Wert mit und nicht als Kind —
	// dieselbe Stelle, an der bei `rotate` der Winkel steht.
	eintragen("built_in_image", eingebautesBild)
}

// ArtBild ist die Art eines Bildwerts.
const ArtBild WertArt = "Bild"

// Bildbaum ist ein Bild, wie ein Skript es zurückgibt.
//
// **Ein Blatt trägt einen Pfad, ein Knoten trägt Kinder.** Was eine Knotenart
// bedeutet, weiß dieses Paket nicht — sie ist eine Zeichenkette, und wer sie
// ausführt, ist `internal/render`.
type Bildbaum struct {
	// Knotenart ist `datei` für ein Blatt, sonst der Name der Verknüpfung.
	Knotenart string

	// Pfad ist der Paketpfad eines Blattes.
	Pfad string

	// Kinder sind die Bilder, aus denen dieser Knoten entsteht — **benannt**.
	//
	// Bei `recolor_image` ist es eines und heißt `image`. Bei `masked_blend`
	// sind es drei — `light`, `dark`, `mask` —, und welches welches ist, steht
	// im Namen: eine Reihenfolge allein trüge die Auskunft nicht.
	Kinder []Bildkind

	// Werte sind die übrigen, benannten Argumente — Farben, Zahlen, Texte.
	Werte []Eintrag

	// Marke ist die [Wert.Marke] des Bildwerts, aus dem das Blatt stammt —
	// leer bei allem, was ein Skript gerechnet oder geschrieben hat. Wer sie
	// liest, weiß, was sie bedeutet (QA-06c, #1976).
	Marke string
}

// Bildkind ist ein benanntes Kind eines Bildbaums.
type Bildkind struct {
	Name string
	Baum Bildbaum
}

// AlsBild gibt den Bildbaum, wenn der Wert einer ist.
//
// **Eine Zeichenkette ist auch ein Bild.** Ein Pfad *ist* eines; das ist die
// Umwandlung, die MSE `to_image` nennt, und ohne sie müsste jede Vorlage sie
// von Hand schreiben. Der Aufrufer bekommt dann ein Blatt.
func (w Wert) AlsBild() (Bildbaum, bool) {
	switch w.Art {
	case ArtBild:
		return w.bildbaum(), true

	case ArtTextwert:
		if w.Text == "" {
			return Bildbaum{}, false
		}

		return Bildbaum{Knotenart: "datei", Pfad: w.Text}, true

	// **Ein `default` und keine Aufzählung der übrigen Arten.** Die Aufzählung
	// war vollständig und ließ trotzdem einen `return` dahinter stehen, den Go
	// verlangt und kein Aufruf erreicht — ein Zweig zu viel (§3).
	default:
		return Bildbaum{}, false
	}
}

// bildbaum liest den Baum aus den Einträgen eines Bildwerts.
func (w Wert) bildbaum() Bildbaum {
	baum := Bildbaum{Knotenart: w.Text, Pfad: w.Zahltext(), Marke: w.Marke}

	for _, eintrag := range w.Eintraege {
		if eintrag.Wert.Art == ArtBild {
			baum.Kinder = append(baum.Kinder, Bildkind{
				Name: eintrag.Name, Baum: eintrag.Wert.bildbaum(),
			})

			continue
		}

		baum.Werte = append(baum.Werte, eintrag)
	}

	return baum
}

// Zahltext gibt den Pfad eines Blattes — leer bei einer Verknüpfung.
func (w Wert) Zahltext() string {
	if w.Art == ArtBild && w.Text == "datei" {
		return w.Eintraege[0].Wert.Text
	}

	return ""
}
