package msescript

import (
	"context"
	"sort"
)

// **Die Stelle einer Karte im geordneten Set** —
// #2006.
//
// # ported: `Set::positionOfCard` und `OrderCache`
//
//	int Set::positionOfCard(const CardP& card, const ScriptValueP& order_by,
//	                        const ScriptValueP& filter) {
//	  OrderCacheP& order = order_cache[make_pair(order_by,filter)];
//	  if (!order) {
//	    // 1. make a list of the order value for each card
//	    FOR_EACH_CONST(c, cards) {
//	      Context& ctx = getContext(c);
//	      values.push_back(order_by->eval(ctx)->toString());
//	      if (filter) keep.push_back(filter->eval(ctx)->toBool());
//	    }
//	    order = make_intrusive<OrderCache<CardP>>(cards, values, filter ? &keep : nullptr);
//	  }
//	  return order->find(card);
//	}
//
// `src/data/set.cpp`, Zeilen 277–299, und `src/util/order_cache.hpp`, Zeilen
// 44–79, Fassung f9d9356 vom 2021-07-06. `OrderCache` sortiert die behaltenen
// Karten nach `smart_less` über ihren Schlüsseln und merkt sich ihre Stelle;
// `find` gibt −1 für eine Karte, die der Filter ausgenommen hat.
//
// # adapted: stabil, und der Speicher gehört einer Umgebung
//
// MSE sortiert mit `std::sort`; gleiche Schlüssel stehen danach, wie der
// Algorithmus es will. Hier bleiben sie in der Reihenfolge der Liste — dieselbe
// Abweichung wie bei `sort_list` (siehe [kleiner]).
//
// Der Speicher hängt bei MSE am `Set` und wird in `updateRecursive` geleert,
// bevor eine Runde Skripte läuft (`script_manager.cpp`, Zeile 316). Hier hängt
// er an einer [Umgebung] und wird von **außen** geleert
// ([Umgebung.OrdnungenVergessen]): der Aufrufer weiß, wann sich die Karten
// geändert haben, die Sprache nicht.
//
// # Der Schlüssel des Speichers
//
// MSE schlüsselt nach den **Zeigern** der beiden Funktionswerte. Hier sind es
// die Rümpfe: ein `{ … }` im Quelltext ist **ein** Knoten, gleich oft er
// ausgewertet wird, und `set_filter()` gibt je nach Karte einen von neun
// verschiedenen — neun Ordnungen, jede einmal gebaut.

// ordnungsschluessel nennt, welche Ordnung ein Speichereintrag ist.
type ordnungsschluessel struct{ ordnung, filter *Knoten }

// ordnungsspeicher hält je Ordnung die Stelle jeder behaltenen Karte.
type ordnungsspeicher struct {
	stellen map[ordnungsschluessel]map[string]int

	// zahlen hält je Filter, wie viele Karten er behält — MSEs `filter_cache`.
	zahlen map[*Knoten]int
}

// OrdnungenMerken legt an dieser Umgebung einen Speicher für geordnete
// Positionen an.
//
// **Ohne ihn geht es auch, nur langsamer**: jede Frage ordnete das ganze Set
// neu. Bei 3.788 Karten wären das vierzehn Millionen Schlüssel statt 3.788.
//
// **Wer die Karten ändert, ruft [Umgebung.OrdnungenVergessen]** — der Speicher
// weiß nicht, welche Liste er geordnet hat.
func (u *Umgebung) OrdnungenMerken() {
	u.ordnungen = &ordnungsspeicher{
		stellen: map[ordnungsschluessel]map[string]int{},
		zahlen:  map[*Knoten]int{},
	}
}

// OrdnungenVergessen leert den Speicher, den diese Umgebung oder eine unter
// ihr hält. Ohne Speicher tut es nichts.
func (u *Umgebung) OrdnungenVergessen() {
	if speicher := u.speicherDerOrdnungen(); speicher != nil {
		clear(speicher.stellen)
		clear(speicher.zahlen)
	}
}

// speicherDerOrdnungen sucht den nächsten Speicher in der Kette.
func (u *Umgebung) speicherDerOrdnungen() *ordnungsspeicher {
	for ebene := u; ebene != nil; ebene = ebene.darauf {
		if ebene.ordnungen != nil {
			return ebene.ordnungen
		}
	}

	return nil
}

// stelleImSet ist `position(of: card, in: set, order_by:, filter:)`.
//
//	ScriptObject<Set*>*  s = dynamic_cast<ScriptObject<Set*>* >(in.get());
//	ScriptObject<CardP>* c = dynamic_cast<ScriptObject<CardP>*>(of.get());
//	if (s && c) { … } else { throw ScriptError(…); }
//
// `basic.cpp`, Zeilen 452–460. **Das Set erkennt man an der [Wert.Kennung]**
// „set", **die Karte an der Gestalt einer Abbildung**: die Karte im Zeichenpfad
// trägt keine Kennung (`kartenabbildung`) und steht auch in keiner Liste — für
// sie ist die Antwort −1, wie bei einer Karte, die der Filter ausnimmt, und
// `card_number` bleibt dort, was es war.
func stelleImSet(
	ctx context.Context, ruf Ruf, gesucht, liste Wert, ordnung, filter *Wert,
) (Wert, error) {
	if gesucht.Art != ArtAbbildung || liste.Kennung != "set" {
		return Nichts, ruf.fehlerAn("position: order_by und filter gehen nur für Karten im Set")
	}

	// MSE prüft das mit `assert(order_by)` — im Release stürzte es ab.
	if ordnung == nil {
		return Nichts, ruf.fehlerAn("position: ein filter braucht auch ein order_by")
	}

	aussen := ruf.Umgebung
	if aussen == nil {
		aussen = Bibliothek()
	}

	schluessel := ordnungsschluessel{ordnung: ordnung.Funktion.rumpfOderNil()}
	if filter != nil {
		schluessel.filter = filter.Funktion.rumpfOderNil()
	}

	speicher := aussen.speicherDerOrdnungen()
	if speicher != nil {
		if gemerkt, da := speicher.stellen[schluessel]; da {
			return stelleAus(gemerkt, gesucht), nil
		}
	}

	gebaut, err := ordnungBauen(ctx, ruf, aussen, liste, ordnung, filter)
	if err != nil {
		return Nichts, err
	}

	if speicher != nil {
		speicher.stellen[schluessel] = gebaut
	}

	return stelleAus(gebaut, gesucht), nil
}

// rumpfOderNil gibt den Rumpf, auch für eine Funktion, die keine ist.
func (f *Funktionsteil) rumpfOderNil() *Knoten {
	if f == nil {
		return nil
	}

	return f.Rumpf
}

// stelleAus ist `OrderCache::find`: −1, wenn die Karte nicht dasteht.
func stelleAus(gebaut map[string]int, gesucht Wert) Wert {
	if stelle, da := gebaut[gesucht.Kennung]; da {
		return Zahl(float64(stelle))
	}

	return Zahl(-1)
}

// geordnet ist eine behaltene Karte mit ihrem Schlüssel.
type geordnet struct {
	kennung    string
	schluessel string
}

// ordnungBauen ordnet die Karten des Sets — `OrderCache::OrderCache`.
func ordnungBauen(
	ctx context.Context, ruf Ruf, aussen *Umgebung, liste Wert, ordnung, filter *Wert,
) (map[string]int, error) {
	behalten := make([]geordnet, 0, len(liste.Eintraege))

	for i := range liste.Eintraege {
		karte := liste.Eintraege[i].Wert

		// **Was keine Kennung trägt, ist keine Karte.** Im Zeichenpfad
		// führen die Einträge von `set` seine Felder, und deren Werte haben
		// kein `card.name`.
		if karte.Kennung == "" {
			continue
		}

		// **Eine Ebene je Karte**: `getContext(c)` bindet `card` an sie, und
		// die nächste soll nicht sehen, was diese angelegt hat.
		umgebung := aussen.Darueber()
		umgebung.Setze("card", karte)

		if filter != nil {
			trifft, err := filterTrifft(ctx, ruf, *filter, umgebung)
			if err != nil {
				return nil, err
			}

			if !trifft {
				continue
			}
		}

		wert, err := gliedschluessel(ctx, *ordnung, umgebung, ruf)
		if err != nil {
			return nil, err
		}

		behalten = append(behalten, geordnet{kennung: karte.Kennung, schluessel: wert.AlsText()})
	}

	sort.SliceStable(behalten, func(a, b int) bool {
		return KlugVergleiche(behalten[a].schluessel, behalten[b].schluessel) < 0
	})

	stellen := make(map[string]int, len(behalten))
	for stelle, karte := range behalten {
		stellen[karte.kennung] = stelle
	}

	return stellen, nil
}

// filterTrifft wertet den Filter einer Karte aus — `filter->eval(ctx)->toBool()`.
func filterTrifft(ctx context.Context, ruf Ruf, filter Wert, umgebung *Umgebung) (bool, error) {
	wert := filter

	if filter.Art == ArtFunktion && filter.Funktion != nil && filter.Funktion.Rumpf != nil {
		gerechnet, err := gliedschluessel(ctx, filter, umgebung, ruf)
		if err != nil {
			return false, err
		}

		wert = gerechnet
	}

	trifft, err := wert.AlsWahrheitswert()
	if err != nil {
		return false, ruf.fehlerAn("position: der filter " + err.Error())
	}

	return trifft, nil
}
