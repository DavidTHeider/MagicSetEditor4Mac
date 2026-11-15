package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// zerteilteFilter zerteilt jeden Filter **einmal** statt einmal je Karte.
//
// # Warum das eine eigene Funktion ist
//
// **Weil die Ersparnis eine Aussage ist, die man prüfen können muss.** Hier
// stand `msescript.AuswertenQuelle` in der inneren Schleife, und die ruft bei
// **jedem** Aufruf `Zerteile`: bei 107 Karten und **zwölf** Packarten mit
// Filter sind das 1.284 Aufrufe statt zwölf. Dreißig Packarten hat das Set,
// aber nur zwölf tragen einen Filter — und nur die stehen in der Schleife
// (Zweitblick über Phase 4.1, Befund 4).
//
// **Und die Ersparnis ist nicht messbar** — das gehört dazu. Gemessen am
// 2026-08-29 über die gebaute Kommandozeile gegen `Vanguard.mse-set` (107
// Karten, Stil `m15-altered`, volles Pack): `mse packarten 7` brauchte vorher
// 4,1 s und danach 4,6 s, also dasselbe im Rauschen. Die Vermutung, das
// Zerteilen sei teuer, war falsch, und sie stand hier schon als Zahl, bevor
// jemand gemessen hatte.
//
// **Wo die Zeit wirklich liegt**, sagt dieselbe Messung mit einer Stoppuhr um
// die beiden Hälften: **4.051 ms** für `umgebungAus` und **1 ms** für alle
// zwölf Filter über alle 107 Karten zusammen. Die Umgebung wird je Karte neu
// gebaut, samt Standardbibliothek, Initskripten und zwei Durchgängen
// Feldstile mit Inhaltsmessung.
//
// **MSE2 macht das anders, und das ist nachgelesen**: `SetScriptContext`
// hält den Kontext **je Stilbogen** und lässt die Initskripte genau einmal
// laufen; je Karte setzt es zwei Variablen, `card` und `styling`
// (`src/script/script_manager.cpp`, Zeilen 30–75, Fassung f9d9356 vom
// 2021-07-06). Das zu portieren ist ein eigenes Werkstück — es betrifft auch
// den Rasterer, und es steht in
// #1304.
//
// **Warum die Zerteilung trotzdem hier heraussteht**: 1.284 Aufrufe gegen
// zwölf ist auch ohne Stoppuhr die richtige Gestalt, und ein Filter mit einem
// Syntaxfehler wird jetzt **einmal** gemeldet statt hundertsiebenmal.
//
// **Ein Filter, der sich nicht zerteilen lässt, wird hier abgewiesen** und
// nicht je Karte erneut. Die Meldung ist dieselbe wie vorher; sie entsteht nur
// einmal statt hundertsiebenmal.
func zerteilteFilter(arten []msepaket.Pack) (
	mitFilter []msepaket.Pack,
	zerteilt map[string]*msescript.Knoten,
	gescheitert map[string]string,
) {
	mitFilter = make([]msepaket.Pack, 0, len(arten))
	zerteilt = make(map[string]*msescript.Knoten, len(arten))
	gescheitert = map[string]string{}

	for _, art := range arten {
		if art.Filter == "" {
			continue
		}

		mitFilter = append(mitFilter, art)

		knoten, fehler := msescript.Zerteile(art.Filter)
		if fehler != nil {
			gescheitert[art.Name] = fehler.Error()

			continue
		}

		zerteilt[art.Name] = knoten
	}

	return mitFilter, zerteilt, gescheitert
}
