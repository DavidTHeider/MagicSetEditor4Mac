// Das Menü einer Wortliste
// (#1074).
//
// # Zwei Sorten Eintrag, und der Unterschied ist der ganze Punkt
//
// Ein **Präfix** ist ein Kästchen: `Legendary ` steht vor dem Typ und lässt
// sich mit `Token ` verbinden. Ein gewöhnlicher Eintrag ist ein Punkt:
// `Creature` **ersetzt** `Instant`.
//
// Im echten Pack sind 100 der 432 Wörter Präfixe — gemessen am 2026-08-24 am
// bezogenen `M15-Magic-Pack`.
//
// # Portiert
//
// `DropDownWordList` aus `src/gui/value/text.cpp`, Zeilen 109–300, Fassung
// f9d9356 vom 2021-07-06 — `setWords`, `addWordsFromScript`, `selection` und
// `select`.
//
// **Was hier nicht steht, ist das Fenster.** MSE hält den Zustand am
// aufgeklappten Menü: `selection()` setzt `active` an den Einträgen, und
// `select()` liest sie dort wieder. Hier bekommt das Menü seinen Text beim
// Bauen und rechnet beides daraus — sonst hinge das Ergebnis daran, ob jemand
// vorher hingesehen hat.

package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// KeineWahl sagt, dass kein Eintrag auf den Text passt.
const KeineWahl = -1

// Wortmenue ist eine Auswahlliste in dem Zustand, den ein Text ihr gibt.
type Wortmenue struct {
	// Kaestchen sagt, dass mindestens ein Eintrag ein Präfix ist — das Menü
	// braucht dann eine Spalte für Kästchen (`has_checkboxes`, Zeile 160).
	Kaestchen bool

	// Gewaehlt ist der Eintrag, auf dem die Auswahl steht, oder [KeineWahl].
	Gewaehlt int

	Eintraege []Wortmenueeintrag
}

// Wortmenueeintrag ist eine Zeile im Menü.
type Wortmenueeintrag struct {
	// Name ist der Text, der eingesetzt wird — samt Leerzeichen am Ende, wenn
	// es ein Präfix ist.
	Name string

	// IstPraefix macht aus der Zeile ein Kästchen statt eines Punktes.
	IstPraefix bool

	// ZeileDarunter zieht einen Strich unter die Zeile.
	ZeileDarunter bool

	// IstGruppe sagt, dass die Zeile ein Untermenü öffnet, statt zu wählen.
	IstGruppe bool

	// Aktiv sagt, dass der aktuelle Text diesen Eintrag enthält.
	Aktiv bool

	// Untermenue ist das Menü einer Gruppe — sonst `nil`.
	//
	// **Gleich mitgebaut und nicht auf Zuruf.** MSE baut es beim Aufklappen
	// (`DropDownWordList::submenu`, Zeilen 231–247); hier hinge daran ein
	// zweiter Weg über die Bindung, und der bräuchte einen Pfad, eine
	// Karte und dieselbe Umgebung noch einmal. Gezählt am `M15-Magic-Pack`:
	// die größte Liste (`planeswalker`) hat 23 Gruppen — das ist billig.
	Untermenue *Wortmenue
}

// NeuesWortmenue baut das Menü einer Wortliste für einen Text.
//
// Portiert aus `DropDownWordList::setWords`, `src/gui/value/text.cpp`, Zeilen
// 151–179, Fassung f9d9356 vom 2021-07-06.
//
// **Ein doppelter Name fällt heraus, eine Gruppe nie.** `if (!already_added ||
// w->isGroup())` — zwei gleichnamige Punkte wären für niemanden zu
// unterscheiden, zwei gleichnamige Untermenüs führen woandershin.
func NeuesWortmenue(
	woerter []msepaket.Wortlisteneintrag, text string, auswerten func(string) string,
) Wortmenue {
	menue := Wortmenue{Gewaehlt: KeineWahl}

	for _, wort := range woerter {
		if wort.IstPraefix {
			menue.Kaestchen = true
		}

		if wort.Skript != "" {
			menue.Eintraege = ausSkript(menue.Eintraege, wort, auswerten)

			continue
		}

		if !wort.IstGruppe() && schonDa(menue.Eintraege, wort.Name) {
			continue
		}

		eintrag := Wortmenueeintrag{
			Name:          wort.Name,
			IstPraefix:    wort.IstPraefix,
			ZeileDarunter: wort.ZeileDarunter,
			IstGruppe:     wort.IstGruppe(),
		}

		if eintrag.IstGruppe {
			unten := NeuesWortmenue(wort.Woerter, text, auswerten)
			eintrag.Untermenue = &unten
		}

		menue.Eintraege = append(menue.Eintraege, eintrag)
	}

	return menue.fuer(text)
}

// WaehlenAmWeg gibt den Text nach einem Klick — auch tief im Untermenü.
//
// Der Weg zählt die Einträge von oben: `[1, 0]` ist der erste Eintrag im
// Untermenü des zweiten Eintrags. Der letzte Schritt ist die Wahl, alle davor
// führen hinein.
//
// **Eine Wahl im Untermenü ersetzt den Bereich ganz.**
// `DropDownWordList::submenu` reicht dasselbe `pos` weiter
// (`src/gui/value/text.cpp`, Zeilen 231–247, Fassung f9d9356 vom 2021-07-06),
// und `select` sieht nur die Einträge **seines** Menüs: was eine Ebene höher
// angekreuzt war, ist danach weg. So tut MSE es.
func (m Wortmenue) WaehlenAmWeg(weg []int) (string, bool) {
	if len(weg) == 0 {
		return "", false
	}

	if len(weg) == 1 {
		if weg[0] < 0 || weg[0] >= len(m.Eintraege) {
			return "", false
		}

		// **Ein Gruppenkopf ist keine Wahl.**
		//
		//	// don't hide if there is a child menu
		//	if (selected_item != NO_SELECTION && submenu(selected_item)) return;
		//
		// `src/gui/drop_down_list.cpp`, Zeilen 373–374, Fassung f9d9356 vom
		// 2021-07-06. MSE beendet die Wahl dort nicht, er öffnet das
		// Untermenü.
		//
		// Hier wies das nur das **Fenster** ab; Kommandozeile, MCP und REST
		// nahmen den Kopf an und schrieben seinen Namen wörtlich ins Feld.
		// Angestrichen vom Zweitblick durch Codex über Phase W — eine
		// Zusage, die eine von vier Flächen hält, ist keine (§16).
		if m.Eintraege[weg[0]].Untermenue != nil {
			return "", false
		}

		return m.Waehlen(weg[0]), true
	}

	if weg[0] < 0 || weg[0] >= len(m.Eintraege) || m.Eintraege[weg[0]].Untermenue == nil {
		return "", false
	}

	return m.Eintraege[weg[0]].Untermenue.WaehlenAmWeg(weg[1:])
}

// fuer rechnet aus, welche Einträge der Text enthält und auf welchem die
// Auswahl steht.
//
// Portiert aus `DropDownWordList::selection`, `src/gui/value/text.cpp`, Zeilen
// 250–274, Fassung f9d9356 vom 2021-07-06.
//
// **Ein Präfix wird abgeschnitten, ein Punkt verglichen.** Passt `Legendary `
// vorn, bleibt `Creature` für den nächsten übrig; deshalb entscheidet die
// Reihenfolge der Liste und nicht die des Textes.
//
// **Die Auswahl steht lieber auf einem Punkt als auf einem Kästchen.** Der
// zweite Zweig — `prefix_selected && !item.is_prefix` — rückt sie von einem
// Präfix auf den ersten gewöhnlichen Eintrag, der auch aktiv ist.
func (m Wortmenue) fuer(text string) Wortmenue {
	rest := text
	praefixGewaehlt := true

	for i := range m.Eintraege {
		eintrag := &m.Eintraege[i]

		if eintrag.IstPraefix {
			eintrag.Aktiv = strings.HasPrefix(rest, eintrag.Name)
			if eintrag.Aktiv {
				rest = rest[len(eintrag.Name):]
			}
		} else {
			eintrag.Aktiv = rest == eintrag.Name
		}

		if !eintrag.Aktiv {
			continue
		}

		if m.Gewaehlt == KeineWahl || (praefixGewaehlt && !eintrag.IstPraefix) {
			m.Gewaehlt = i
			praefixGewaehlt = eintrag.IstPraefix
		}
	}

	return m
}

// Waehlen gibt den Text, der nach einem Klick auf einen Eintrag dasteht.
//
// Portiert aus `DropDownWordList::select`, `src/gui/value/text.cpp`, Zeilen
// 276–297, Fassung f9d9356 vom 2021-07-06.
//
// # Die beiden Regeln
//
//   - **Ein Präfix wird umgeschaltet**: es steht danach da, wenn es vorher
//     nicht dastand — und die anderen Präfixe bleiben, wie sie waren.
//   - **Ein Punkt ersetzt**: der geklickte kommt, alle anderen gehen. Es sei
//     denn, geklickt wurde ein Präfix — dann bleibt der aktive Punkt stehen.
//
// **Die Reihenfolge ist die der Liste**, nicht die des Klickens: `select`
// läuft einmal über die Einträge und hängt an, was danach dastehen soll.
func (m Wortmenue) Waehlen(eintrag int) string {
	praefixUmgeschaltet := m.Eintraege[eintrag].IstPraefix

	var neu strings.Builder

	for i, dieser := range m.Eintraege {
		if dieser.IstPraefix {
			if dieser.Aktiv != (i == eintrag) {
				neu.WriteString(dieser.Name)
			}

			continue
		}

		if i == eintrag || (praefixUmgeschaltet && dieser.Aktiv) {
			neu.WriteString(dieser.Name)
		}
	}

	return neu.String()
}

// schonDa sagt, ob ein Name bereits im Menü steht.
func schonDa(eintraege []Wortmenueeintrag, name string) bool {
	for _, eintrag := range eintraege {
		if eintrag.Name == name {
			return true
		}
	}

	return false
}
