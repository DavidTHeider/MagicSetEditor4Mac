package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Was eine Exportvorlage über Karten und Set zu sehen bekommt.**
//
// Getrennt von `exportvorlage.go`, weil es eine andere Frage ist: dort steht,
// **welche** Funktionen in der Umgebung liegen — hier, **welche Daten**. Die
// eine Seite entscheidet der Aufrufer über seine Fähigkeiten, die andere das
// Spielpaket über seine Felder.

// **Eine Liste ist eine Abbildung mit Zahlen als Namen** (`docs/FORMATE.md`) —
// deshalb greift `cards.0` auf die erste Karte zu, und `for each card in cards`
// geht sie der Reihe nach durch.
func kartenliste(karten []*domain.CardRow, kartenfelder []domain.Field) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(karten))

	for stelle, karte := range karten {
		felder := make([]msescript.Eintrag, 0, len(karte.Values)+len(kartenfelder))
		gefuehrt := map[string]bool{}

		for _, wert := range karte.Values {
			felder = append(felder, msescript.Eintrag{
				Name: skriptname(wert.Field),
				Wert: msescript.Text(wert.Text),
			})
			gefuehrt[skriptname(wert.Field)] = true
		}

		// **Was das Spiel erklärt und die Karte nicht führt, kommt leer
		// dazu** (#609). In MSE legt der Kartenkonstruktor für jedes erklärte
		// Feld einen Wert an (`src/data/card.cpp:20–30`, Fassung f9d9356 vom
		// 2021-07-06); `card.notes` gibt es dort immer. Bei uns fehlte es, und
		// seit ein fehlendes Glied ein Fehler ist statt `nil`, scheiterte
		// daran jede Vorlage, die es liest.
		//
		// **Leer und nicht gerechnet.** Der Vorgabewert eines Feldes ist ein
		// Skript über `card`, und das braucht die Umgebung des Zeichenpfads;
		// ihn hier zu rechnen wäre ein zweiter Weg zu derselben Frage. Was
		// fehlt, steht als eigener Punkt in #609.
		for i := range kartenfelder {
			name := skriptname(kartenfelder[i].Name)
			if gefuehrt[name] {
				continue
			}

			felder = append(felder, msescript.Eintrag{Name: name, Wert: msescript.Text("")})
			gefuehrt[name] = true
		}

		// **`notes` steht an jeder Karte, auch an einer ohne** — die Regel und
		// ihre Fundstelle stehen bei [mitNotizen].
		//
		// **`time_created` und `time_modified` reist MSE ebenso mit** (dieselbe
		// Stelle). Sie fehlen hier, weil [domain.CardRow] sie nicht führt; wer
		// sie braucht, holt sie erst in den Bestand.
		felder = mitNotizen(felder, karte)

		eintraege = append(eintraege, msescript.Eintrag{
			Name: fmt.Sprint(stelle),
			Wert: msescript.Wert{
				Art: msescript.ArtAbbildung, Eintraege: felder,

				// **Und die Kennung macht sie zu einem Gegenstand** (#1366).
				// Sie steht **neben** der Marke und ersetzt sie nicht: die
				// Marke ist eine Zahl für den Rasterer, die Kennung ist die
				// Identität für den Vergleich. Beide aus derselben Quelle,
				// beide mit eigenem Zweck.
				Kennung: Skriptkartenkennung(karte),

				// **Die Marke trägt die Kennung** (#216). Kein Skript kann sie
				// lesen; `write_image_file` braucht sie, um zu wissen, welche
				// Karte zu rastern ist. Der Grund steht am Feld selbst.
				Marke: strconv.FormatInt(karte.ID, 10),
			},
		})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}

// optionswert macht aus der Angabe einer Option ihren Wert.
//
// # Warum eine boolesche Option kein Text ist
//
// `if options.text_costs then …` steht 35-mal in den Exportvorlagen von
// `M15-Magic-Pack`, und `text costs` ist dort als `type: boolean` deklariert.
// #202 fragte,
// was MSE aus einer **Zeichenkette** als Wahrheitswert macht, und nannte es
// „nicht erschlossen" — die Frage stellt sich nicht: die Option nennt ihren Typ
// selbst.
//
// **Zu deuten ist nur die Vorgabe.** Von 36 booleschen Optionen tragen 29 ein
// `initial:`, und das ist ausnahmslos `yes` (11) oder `no` (18).
//
// **Was nicht `yes` ist, ist falsch** — auch eine leere Angabe, und das ist der
// Fall der sieben ohne `initial:`. Die andere Lesart, „alles außer `no` ist
// wahr", machte aus einer fehlenden Angabe ein eingeschaltetes Merkmal; ein
// Schalter, der von selbst angeht, ist der teurere Irrtum.
func optionswert(art, angabe string) msescript.Wert {
	if art != "boolean" {
		return msescript.Text(angabe)
	}

	return msescript.Wahrheit(strings.TrimSpace(angabe) == "yes")
}

// skriptname macht aus einem Feldnamen den Namen, unter dem ein Skript ihn
// liest — `rule text` wird `rule_text`.
//
// **Dieselbe Wandlung wie MSEs `canonical_name_form`**
// (`src/util/string.cpp`, Zeilen 121–126, Fassung f9d9356 vom 2021-07-06):
// jedes Leerzeichen wird ein Unterstrich.
func skriptname(feld string) string {
	return strings.ReplaceAll(feld, " ", "_")
}

// setabbildung baut `set` aus den Angaben und den erklärten Feldern.
//
// **Dieselbe Auffüllung wie bei der Karte** (#609): was das Spiel erklärt und
// niemand eingetragen hat, steht leer da — nicht gar nicht.
func setabbildung(
	werte []domain.Value, setfelder []domain.Field, karten msescript.Wert,
) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(werte)+len(setfelder))
	gefuehrt := map[string]bool{}

	for _, wert := range werte {
		name := skriptname(wert.Field)
		eintraege = append(eintraege, msescript.Eintrag{Name: name, Wert: msescript.Text(wert.Text)})
		gefuehrt[name] = true
	}

	for i := range setfelder {
		name := skriptname(setfelder[i].Name)
		if gefuehrt[name] {
			continue
		}

		eintraege = append(eintraege, msescript.Eintrag{Name: name, Wert: msescript.Text("")})
		gefuehrt[name] = true
	}

	return setMitKarten(msescript.Wert{
		Art: msescript.ArtAbbildung, Eintraege: eintraege,

		// **Das Set ist ein Gegenstand** (#2006): `position(of: card, in:
		// set, order_by: …)` erkennt es daran.
		Kennung: "set",
	}, karten.Eintraege)
}

// setMitKarten macht aus den Feldern eines Sets das Set samt Karten.
//
// **Die Karten als Einträge, die Angaben als Glieder** (#653): `set` trägt in
// MSE beide Gesichter — durchlaufen gibt die Karten, `set.<name>` die Angabe.
//
//	ScriptValueP make_iterator(const Set& set) {
//	  return make_intrusive<ScriptCollectionIterator<vector<CardP>>>(&set.cards);
//	}
//
// `src/data/set.cpp`, Zeilen 251–253, Fassung f9d9356 vom 2021-07-06.
//
// **Und `set.cards` ist dieselbe Liste als Glied** (#2006): das Orakel
// beantwortet `set.cards.2.card_number`, und das Initskript von
// `magic.mse-game` liest `set.cards`. Die Angaben stehen davor — heißt ein Feld
// `cards`, gewinnt es; im Pack führt keins eines.
//
// **Kennung und alles andere des Sets bleiben**: wer `set` aus der Stilumgebung
// nimmt, behält „set" als Kennung. Erwartet wird ein Set **ohne** Karten, dessen
// Einträge seine Angaben sind — so liefern es [setangaben] und
// [setabbildung].
func setMitKarten(set msescript.Wert, karten []msescript.Eintrag) msescript.Wert {
	glieder := make([]msescript.Eintrag, 0, len(set.Eintraege)+1)
	glieder = append(glieder, set.Eintraege...)
	glieder = append(glieder, msescript.Eintrag{
		Name: "cards",
		Wert: msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: karten},
	})

	set.Eintraege = karten
	set.Glieder = glieder

	return set
}
