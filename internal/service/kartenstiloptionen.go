package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die Stiloptionen, die eine Karte selbst mitbringt**
// (#547).
//
//	IndexMap<FieldP,ValueP>& Set::stylingDataFor(const CardP& card) {
//	  if (card && card->has_styling) return card->styling_data;
//	  else                           return stylingDataFor(stylesheetFor(card));
//	}
//
// `src/data/set.cpp`, Zeilen 102–105, Fassung f9d9356 vom 2021-07-06.

// Kartenstilfelder sind die beiden Kartenwerte, in denen die Stiloptionen einer
// Karte stehen.
//
// **Sie sind keine Kartenfelder des Spiels**, sondern Eigenschaften jeder Karte
// — wie `notes`. Deshalb stehen sie in keiner `card field:`-Liste, und der
// Import legt sie als gewöhnliche Werte ab.
const (
	Stilschalterfeld = "has styling"
	Stilwertefeld    = "styling data"
)

// kartenstilwerte gibt die Stiloptionen, die diese Karte selbst nennt.
//
// # Warum es eine Überlagerung wird und kein Ersetzen
//
// MSE gibt die **ganze** Abbildung der Karte zurück — sie ist aber schon aus
// den `styling_fields` des Stils vorbelegt: `styling_data.init(stylesheet->styling_fields)`
// steht **vor** `REFLECT(styling_data)` (`src/data/card.cpp`, Zeilen 76–90,
// Fassung f9d9356 vom 2021-07-06). Was die Karte nicht nennt, trägt damit die
// Vorgabe des Stils, und der Aufrufer legt diese Werte **darüber**.
//
// # `has_styling` wird gefragt, nicht erschlossen
//
// MSE prüft den Schalter, statt zu schauen, ob Daten dastehen. Eine Karte mit
// `has_styling: false` und einem alten `styling_data:` benutzt die Vorgaben des
// Stils — und genau das ist der Sinn des Schalters.
func kartenstilwerte(karte *domain.CardRow) map[string]string {
	werte := map[string]string{}
	if karte == nil {
		return werte
	}

	schalter, _ := karte.Value(Stilschalterfeld)
	if strings.TrimSpace(schalter) != "true" {
		return werte
	}

	roh, _ := karte.Value(Stilwertefeld)

	for _, zeile := range strings.Split(roh, "\n") {
		// **Getrennt am ersten Doppelpunkt.** `other_options` führt Listen,
		// und ein Wert darf selbst einen tragen.
		name, wert, getrennt := strings.Cut(zeile, ":")
		if !getrennt {
			continue
		}

		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		werte[strings.ReplaceAll(name, " ", "_")] = strings.TrimSpace(wert)
	}

	return werte
}

// mitKartenstil legt die Werte der Karte über die Vorgaben des Stils.
//
// **Nur bekannte Namen.** Ein `styling_data:`-Eintrag ohne erklärtes
// `styling field:` gehört zu einem anderen Stilpaket — eine Karte behält ihre
// Stildaten, wenn jemand den Stil wechselt, und MSEs `IndexMap` kennt dort
// ebenfalls nur die Felder des aktuellen Stils.
func mitKartenstil(
	vorgaben []msescript.Eintrag, werte map[string]string,
	farben map[string]farbfeld,
) []msescript.Eintrag {
	if len(werte) == 0 {
		return vorgaben
	}

	for stelle := range vorgaben {
		wert, genannt := werte[vorgaben[stelle].Name]
		if !genannt {
			continue
		}

		// **Ein Farbfeld trägt eine Farbe, auch wenn die Karte sie nennt.**
		// Ohne diesen Zweig käme der rohe Wahlname in die Umgebung, und
		// `font_color(styling.text_color)` bekäme einen Namen statt einer
		// Farbe — dieselbe Falle wie eine Schleife weiter oben.
		if feld, ist := farben[vorgaben[stelle].Name]; ist {
			vorgaben[stelle].Wert = farbfeldwert(wert, feld)

			continue
		}

		vorgaben[stelle].Wert = msescript.Text(wert)
	}

	return vorgaben
}

// Setstilfeld ist der Setwert, unter dem die setweiten Stiloptionen liegen.
//
//	IndexMap<FieldP, ValueP>& Set::stylingDataFor(const StyleSheet& stylesheet) {
//	  return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
//	}
//
// `src/data/set.cpp`, Zeilen 99–101, Fassung f9d9356 vom 2021-07-06 —
// **geschlüsselt nach dem Namen des Stilpakets**, und der ist der präfixierte
// (`magic-m15-saga`), nicht die Kurzform.
const Setstilfeld = "styling"

// setstilwerte gibt die setweiten Stiloptionen für ein Stilpaket.
//
// **Der Schlüssel ist der präfixierte Paketname.** So schreibt es
// `magic-set-generator` (`src/mse.rs`, Zeilen 151–174, §11), und so steht es in
// jedem kanonischen Set:
//
//	styling:
//		magic-m15-saga:
//			text_box_mana_symbols: magic-mana-small.mse-symbol-font
//
// **Abgelegt als eine Zeile je Option**, `<paket>/<name>: <wert>` — derselbe
// Weg wie bei der Karte ([kartenstilwerte]).
func setstilwerte(setangaben []domain.Value, paket string) map[string]string {
	werte := map[string]string{}

	if paket == "" {
		return werte
	}

	for _, angabe := range setangaben {
		if angabe.Field != Setstilfeld {
			continue
		}

		for _, zeile := range strings.Split(angabe.Text, "\n") {
			ort, rest, getrennt := strings.Cut(zeile, "/")
			if !getrennt || strings.TrimSpace(ort) != paket {
				continue
			}

			// **Getrennt am ersten Doppelpunkt**, wie bei der Karte:
			// `other_options` führt Listen, und ein Wert darf selbst einen
			// tragen.
			name, wert, hat := strings.Cut(rest, ":")
			if !hat || strings.TrimSpace(name) == "" {
				continue
			}

			werte[strings.ReplaceAll(strings.TrimSpace(name), " ", "_")] = strings.TrimSpace(wert)
		}
	}

	return werte
}
