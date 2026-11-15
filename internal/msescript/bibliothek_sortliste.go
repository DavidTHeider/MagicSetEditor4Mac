package msescript

import (
	"context"
	"fmt"
	"sort"
)

// `sort_list` ordnet eine Liste nach einem Schlüsselausdruck.
//
// # Erschlossen, nicht erfunden — der Vergleich seit #2006 ported (siehe [kleiner])
//
// Vier Exportvorlagen von `M15-Magic-Pack` rufen sie, und alle in derselben
// Form:
//
//	for each card in sort_list(cards, order_by: {card.card_number}) do …
//
// `magic-forum-spoiler` schreibt `{input.card_number}` statt `{card.…}`.
// **Dasselbe Glied unter zwei Namen** — der Schlüsselausdruck bekommt es
// deshalb als beide (#202).
//
// **Unsere Referenzfassung tut das nicht**, nachgelesen und gefragt am
// 2026-10-10 (F05, #2247):
// `sort_script` setzt `card`, wenn die Liste ein Set ist, sonst `input`
// (`src/script/functions/basic.cpp`, Zeile 507, Fassung f9d9356 vom
// 2021-07-06), und das Orakel endet bei `{card.name}` über `set.cards` in
// „Variable not set: card". **Die Packs tragen `mse version: 2.5.8`** und
// benutzen beide Namen über beiden Arten von Listen — unter f9d9356
// scheiterten `magic-toexp`, `magic-skeleton-extractor`, `planechase-forum`
// und `saga-forum`. Portierungsmarke **adapted**: dieser Auswerter folgt den
// Packs; dass eine neuere Fassung beide Namen setzt, ist erschlossen
// (`docs/FORMATE.md`).
//
// # Was sie nicht kann
//
// **Absteigend.** Keine der gesichteten Vorlagen ordnet rückwärts, und ein
// `descending:`-Argument zu bauen, das niemand benutzt, hieße MSE etwas
// anzudichten (`docs/FORMATE.md`).
func init() {
	eintragen("sort_list", ordneListe)
}

// ordneListe ist `sort_list(liste, order_by: {…})`.
func ordneListe(ctx context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("sort_list"); err != nil {
		return Nichts, err
	}

	liste := ruf.Stellung[0]
	if liste.Art != ArtAbbildung {
		return Nichts, NeuerFehler(ruf.Zeile, ruf.Spalte,
			SchluesselKeineListe, string(liste.Art))
	}

	schluessel, err := schluesselVon(ctx, ruf, liste)
	if err != nil {
		return Nichts, err
	}

	stellen := make([]int, len(liste.Eintraege))
	for i := range stellen {
		stellen[i] = i
	}

	// **Stabil**, und das ist keine Feinheit: eine Liste, deren Ordnung sich
	// bei gleichen Schlüsseln von Lauf zu Lauf änderte, ergäbe eine
	// Ausgabedatei, die sich ohne Änderung unterscheidet.
	sort.SliceStable(stellen, func(a, b int) bool {
		return kleiner(schluessel[stellen[a]], schluessel[stellen[b]])
	})

	// **Die Namen werden neu vergeben.** Eine Liste ist in MSE-Script eine
	// Abbildung mit Zahlen als Namen; `colors.0` greift auf ihr erstes Glied
	// zu. Wer die alten Namen mitnähme, ordnete die Glieder und ließe die
	// Zugriffe darauf zeigen, wo sie vorher waren.
	geordnet := Wert{Art: ArtAbbildung, Eintraege: make([]Eintrag, 0, len(stellen))}
	for neu, alt := range stellen {
		geordnet.Eintraege = append(geordnet.Eintraege, Eintrag{
			Name: fmt.Sprintf("%d", neu),
			Wert: liste.Eintraege[alt].Wert,
		})
	}

	return geordnet, nil
}

// schluesselVon rechnet je Glied seinen Ordnungsschlüssel aus.
func schluesselVon(ctx context.Context, ruf Ruf, liste Wert) ([]Wert, error) {
	schluessel := make([]Wert, len(liste.Eintraege))

	nachWas, benannt := ruf.Benannter("order_by")

	// **Wer die eingebaute Funktion von Hand ruft, gibt keine Umgebung mit.**
	// Dann bleibt die Bibliothek — ein Rumpf ohne eigene Namen ordnet damit
	// genauso.
	aussen := ruf.Umgebung
	if aussen == nil {
		aussen = Bibliothek()
	}

	for i := range liste.Eintraege {
		glied := liste.Eintraege[i].Wert

		// **Ohne `order_by` ordnet das Glied sich selbst.** Keine der
		// gesichteten Vorlagen tut das; abzuweisen wäre trotzdem falsch, denn
		// „ordne diese Liste" ist die naheliegendste Bedeutung des Aufrufs
		// ohne Beiwerk.
		if !benannt {
			schluessel[i] = glied

			continue
		}

		// **Über der Umgebung des Aufrufers und nicht über der Bibliothek**
		// (#651):
		//
		//	while (ScriptValueP v = it->next()) {
		//	  ctx.setVariable(set ? _("card") : _("input"), v);
		//	  values.push_back(make_pair(order_by.eval(ctx)->toString(), v));
		//	}
		//
		// `src/script/functions/basic.cpp`, Zeilen 489–520, Fassung f9d9356
		// vom 2021-07-06. **Derselbe `ctx`** — der Rumpf sieht, was das Skript
		// selbst erklärt hat. `magic-forum-booster` ordnet nach einem eigenen
		// `booster_sort`, und über einer frischen Umgebung ist der Name fort.
		//
		// **Darüber und nicht darin**, anders als MSE: dort setzt
		// `ctx.setVariable` in die laufende Ebene, und `input` bleibt nach dem
		// Ordnen stehen. Eine eigene Ebene je Glied ist dasselbe für den
		// Rumpf und lässt nichts zurück.
		umgebung := aussen.Darueber()
		umgebung.Setze("input", glied)
		umgebung.Setze("card", glied)

		wert, err := gliedschluessel(ctx, nachWas, umgebung, ruf)
		if err != nil {
			return nil, err
		}

		schluessel[i] = wert
	}

	return schluessel, nil
}

// gliedschluessel wertet den Schlüsselausdruck eines Glieds aus.
func gliedschluessel(
	ctx context.Context, nachWas Wert, umgebung *Umgebung, ruf Ruf,
) (Wert, error) {
	// **Ein Block und kein Aufruf.** `order_by: {card.card_number}` übergibt
	// einen Funktionswert; ausgewertet wird sein **Rumpf** in einer Umgebung,
	// die das Glied kennt. Wer ihn als Funktion riefe, müsste sich auf einen
	// Argumentnamen festlegen — und die Vorlagen benutzen zwei.
	if nachWas.Art != ArtFunktion || nachWas.Funktion == nil ||
		nachWas.Funktion.Rumpf == nil {
		return Nichts, ruf.fehlerAn(fmt.Sprintf(
			"`order_by` erwartet einen Ausdruck in geschweiften Klammern, hier steht %s",
			nachWas.Art))
	}

	auswerten := &auswerter{ctx: ctx}

	return auswerten.werte(nachWas.Funktion.Rumpf, umgebung)
}

// kleiner vergleicht zwei Ordnungsschlüssel — **als Text, und klug**
// (#2006):
//
//	values.push_back(make_pair(order_by.eval(ctx)->toString(), v));
//	…
//	sort(values.begin(), values.end(), smart_less_first);
//
// `src/script/functions/basic.cpp`, Zeilen 489–520, Fassung f9d9356 vom
// 2021-07-06. Hier stand ein Vergleich Byte für Byte, mit Zahlen als Zahlen;
// die Kartennummer von `magic.mse-game` ist aber **Text** wie `2/99`, und
// `10/99` stand vor `2/99`. [KlugVergleiche] zählt die Zahl im Text als Zahl.
//
// **adapted: stabil.** MSE nimmt `std::sort`, und gleiche Schlüssel stehen
// danach, wie es der Algorithmus will; hier bleiben sie in der Reihenfolge der
// Eingabe — eine Ausgabe, die sich ohne Änderung unterscheidet, wäre schlimmer.
func kleiner(links, rechts Wert) bool {
	return KlugVergleiche(links.AlsText(), rechts.AlsText()) < 0
}
