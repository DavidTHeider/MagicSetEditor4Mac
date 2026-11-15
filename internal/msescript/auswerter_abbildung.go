package msescript

// **Eine Abbildung entsteht** — `[a: 1, b: 2]` und `["x", "y"]`.
//
// Eigene Datei, weil `auswerter_formen.go` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort, **welche** Form wie ausgewertet wird,
// hier, was beim Ablegen eines Glieds gilt — und das ist seit
// #664 mehr
// als ein `append`.

func (a *auswerter) abbildung(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	wert := Wert{Art: ArtAbbildung}

	for i, kind := range knoten.Kinder {
		eintrag, err := a.werte(kind, umgebung)
		if err != nil {
			return Nichts, err
		}

		wert.Eintraege = eingetragen(wert.Eintraege, knoten.Namen[i], eintrag)
	}

	return wert, nil
}

// eingetragen legt ein Glied ab — und **überschreibt** einen gleichen Namen.
//
//	void Context::makeObject(size_t n) {
//	  ScriptCustomCollectionP ret(new ScriptCustomCollection());
//	  …
//	    if (key != script_nil) { // valid key
//	      ret->key_value[key->toString()] = val;
//	    } else {
//	      ret->value.push_back(val);
//	    }
//
// `src/script/context.cpp`, Zeilen 565–578, Fassung f9d9356 vom 2021-07-06.
// **Eine Zuweisung in eine `std::map`**: der spätere Eintrag ersetzt den
// früheren, und gezählt wird, was danach dasteht
// (#664).
//
// **Ein unbenanntes Glied trifft nichts.** Bei MSE landet es in `value` und
// nicht in `key_value`; eine Liste zöge sich sonst auf ein Glied zusammen.
//
// **Die Stelle des ersten bleibt.** MSE ordnet `key_value` nach dem Schlüssel
// und hat damit gar keine Reihenfolge der Datei; hier ist sie Teil des
// Rundlaufs (`docs/FORMATE.md`), und ein Glied, das beim Überschreiben ans
// Ende rutschte, änderte sie.
//
// Aufgefallen an `magic-m15-altered.mse-style`: `template_prefix` erklärt
// `draft:` zweimal — erst auf ein Verzeichnis, das es im Pack nicht gibt, dann
// auf eines, das es gibt.
func eingetragen(eintraege []Eintrag, name string, wert Wert) []Eintrag {
	if name != "" {
		for i := range eintraege {
			if eintraege[i].Name == name {
				eintraege[i].Wert = wert

				return eintraege
			}
		}
	}

	return append(eintraege, Eintrag{Name: name, Wert: wert})
}
