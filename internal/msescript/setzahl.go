package msescript

import "context"

// zahlImSet ist `length(set, filter:)` und `number_of_items(in: set, filter:)`
// — Folge aus #2006.
//
// # ported: `Set::numberOfCards`
//
//	int Set::numberOfCards(const ScriptValueP& filter) {
//	  if (!filter) return (int)cards.size();
//	  … filter_cache …
//	  FOR_EACH_CONST(c, cards) {
//	    if (filter->eval(getContext(c))->toBool()) ++n;
//	  }
//
// `src/data/set.cpp`, Zeilen 301–314, und `script_length_of`,
// `src/script/functions/basic.cpp`, Zeilen 555–563, Fassung f9d9356 vom
// 2021-07-06. **Karten sind die Glieder mit [Wert.Kennung]** — wie in
// [ordnungBauen]; die Felder des Sets zählen nicht. Im Zeichenpfad trägt `set`
// keine Karten, und die Zahl ist null: dort bleibt `card_count()`, was es war.
//
// **Gemerkt je Filter**, wenn die Umgebung einen Speicher hat
// ([Umgebung.OrdnungenMerken]) — wie MSEs `filter_cache`, mit dem Rumpf des
// Filters als Schlüssel statt seines Zeigers.
func zahlImSet(ctx context.Context, ruf Ruf, set Wert) (Wert, error) {
	_, filter := ordnungsangaben(ruf)

	aussen := ruf.Umgebung
	if aussen == nil {
		aussen = Bibliothek()
	}

	var schluessel *Knoten
	if filter != nil {
		schluessel = filter.Funktion.rumpfOderNil()
	}

	speicher := aussen.speicherDerOrdnungen()
	if filter != nil && speicher != nil {
		if gemerkt, da := speicher.zahlen[schluessel]; da {
			return Zahl(float64(gemerkt)), nil
		}
	}

	zahl := 0

	for i := range set.Eintraege {
		karte := set.Eintraege[i].Wert
		if karte.Kennung == "" {
			continue
		}

		if filter != nil {
			umgebung := aussen.Darueber()
			umgebung.Setze("card", karte)

			trifft, err := filterTrifft(ctx, ruf, *filter, umgebung)
			if err != nil {
				return Nichts, err
			}

			if !trifft {
				continue
			}
		}

		zahl++
	}

	if filter != nil && speicher != nil {
		speicher.zahlen[schluessel] = zahl
	}

	return Zahl(float64(zahl)), nil
}
