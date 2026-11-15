package msescript

import (
	"fmt"
	"math"
	"strconv"
)

// Zugriff auf einen Teil eines Werts — über einen Namen oder über eine Stelle.
//
// **Getrennt von `auswerter_formen.go`**, weil es eine eigene Frage ist: dort
// steht, was eine Form *bedeutet*, hier, wie man in einen Wert hineingreift.
// Beide Wege sind bewusst nachsichtig — ein fehlender Schlüssel ist in einer
// fremden Vorlage der Normalfall und darf sie nicht anhalten (§8).

// index liest `abbildung[schlüssel]`.
//
// Ein fehlender Schlüssel ergibt `nil` und keinen Fehler. `template_prefix[type]`
// läuft über jeden Kartentyp; einer, für den kein Rahmen hinterlegt ist, ist in
// einer fremden Vorlage der Normalfall und darf sie nicht anhalten (§8).
func (a *auswerter) index(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	ziel, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	// **Auch das ist ein Fehlerwert und kein Abbruch** — siehe [Fehlerwert].
	if ziel.Art != ArtAbbildung && ziel.Art != ArtTextwert {
		return Fehlerwert(fmt.Sprintf(
			"hier wird %s indiziert, das geht nur bei einer Abbildung oder einer Zeichenkette",
			ziel.Art)), nil
	}

	schluessel, err := a.werte(knoten.Kinder[1], umgebung)
	if err != nil {
		return Nichts, err
	}

	// **Eine Zeichenkette gibt das Zeichen an der Stelle.**
	//
	// Herkunft: `ScriptString::getMember` in `src/script/value.cpp:376` von
	// MSE2. `color_combination` schreibt
	// `base := base_hybrid[shape][color_count]`, und die erste Stufe davon
	// liefert eine Zeichenkette — ohne diesen Zweig fiel `card color → image`
	// aus, das Feld, das den Rahmen zeichnet
	// (#240).
	//
	// **Über das Ende hinaus ist ein Fehlerwert**, den `or else` auffängt
	// (`ScriptString::getMember`, `src/script/value.cpp`, Zeilen 377–384,
	// Fassung f9d9356 vom 2021-07-06).
	if ziel.Art == ArtTextwert {
		return zeichenAn(ziel.Text, schluessel), nil
	}

	wert, da := ziel.Eintrag(schluessel.AlsText())
	if !da {
		return Fehlerwert("die Abbildung führt kein Glied " +
			strconv.Quote(schluessel.AlsText())), nil
	}

	return wert, nil
}

// zeichenAn gibt das Zeichen einer Zeichenkette an einer Stelle.
func zeichenAn(text string, schluessel Wert) Wert {
	stelle, err := schluessel.AlsZahl()
	if err != nil {
		return Fehlerwert("eine Zeichenkette wird über eine Zahl indiziert, hier steht " +
			schluessel.AlsText())
	}

	// **In Zeichen, nicht in Bytes.** Ein Kartenname mit Umlaut gäbe sonst ein
	// halbes Zeichen zurück.
	zeichen := []rune(text)

	if stelle != math.Trunc(stelle) || stelle < 0 || int(stelle) >= len(zeichen) {
		return Fehlerwert("die Zeichenkette hat kein Zeichen an der Stelle " +
			schluessel.AlsText())
	}

	return Text(string(zeichen[int(stelle)]))
}

// zugriff liest `ziel.name` — dieselbe Nachsicht wie beim Index.
func (a *auswerter) zugriff(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	ziel, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	// **Eine Zeichenkette gibt ihre Zeichen her.**
	//
	//	ScriptValueP getMember(const String& name) const override {
	//	  // get member returns characters
	//	  long index;
	//	  if (name.ToLong(&index) && index >= 0 && (size_t)index < value.size()) {
	//	    return to_script(String(1,value[index]));
	//	  }
	//	  …
	//	}
	//
	// `ScriptString::getMember`, `src/script/value.cpp`, Zeilen 377–384,
	// Fassung f9d9356 vom 2021-07-06. `split.0` ist damit das erste Wort einer
	// Zerlegung und `text.0` das erste Zeichen — dieselbe Schreibweise wie der
	// Index, denn MSE kennt beide nicht getrennt.
	//
	// **480 Befunde** über die 190 Stile von `M15-Magic-Pack` lauteten darauf
	// (#345).
	if ziel.Art == ArtTextwert {
		return zeichenAn(ziel.AlsText(), Text(knoten.Text)), nil
	}

	// **Ein Zugriff ins Leere ist ein Wert, kein Abbruch** — MSEs
	// `delay_error` (siehe [Fehlerwert]). `or else` fängt ihn; wer ihn benutzt,
	// bekommt die Meldung.
	if ziel.Art != ArtAbbildung {
		return Fehlerwert(fmt.Sprintf("%q gibt es nur an einer Abbildung, hier steht %s",
			knoten.Text, ziel.Art)), nil
	}

	// `card.rarity` wird als **ein** Name gemerkt. Ein Feld hängt an dem, was es
	// wirklich liest; wer nur `card` protokollierte, rechnete bei jeder
	// Änderung an irgendeinem Feld alles neu.
	if unter := knoten.Kinder[0]; unter.Form == FormName {
		a.merke(unter.Text + "." + knoten.Text)
	}

	wert, da := ziel.Eintrag(knoten.Text)
	if !da {
		return Fehlerwert(fmt.Sprintf("die Abbildung führt kein Glied %q", knoten.Text)), nil
	}

	return wert, nil
}
