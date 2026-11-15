package domain

// **Eine nicht umgesetzte Fähigkeit ist ein Befund** — Werkstück M14 von
// Phase 5.5
// (#1616).
//
// # Die dritte Abnahme des Issues, wörtlich
//
// **„Eine nicht umgesetzte Fähigkeit auf einer Karte ist ein Befund am Tisch
// und kein stiller Ausfall."** Ohne diese Stelle läge eine Kreatur mit
// Flugfähigkeit **und** Schutzschild auf dem Tisch, das eine gälte und das
// andere nicht — und niemand sähe den Unterschied. Das ist §8 in seiner
// gewöhnlichen Gestalt: fremde Daten sind der Normalfall, und was nicht
// getragen wird, wird **benannt**.
//
// # Warum eine Liste und nicht „alles, was kein Satz ist"
//
// **Weil ein Regeltext überwiegend aus Sätzen besteht.** „Wann immer diese
// Kreatur angreift, ziehe eine Karte" ist keine Schlüsselwortfähigkeit und
// darf kein Befund werden; sonst stünde an jeder zweiten Karte einer, und
// eine Meldung, die überall steht, sagt nichts.
//
// **Die Liste ist dieselbe wie die der Zählung** — 57 Namen aus CR 702, über
// die `internal/werkzeuge/schluesselwoerterzaehlen_test.go` gezählt hat.
// Zwölf davon setzt dieses Programm durch; die übrigen **fünfundvierzig**
// stehen hier und werden gemeldet.
//
// **Was das nicht fängt**, steht in `docs/SCHLUESSELWOERTER.md`: die 137
// weiteren Schlüsselwörter aus CR 702, die in den kanonischen Sets nicht
// vorkamen, und jedes mit einem Parameter („Schutzschild {2}").

import (
	"fmt"
	"slices"
	"strings"
)

// offeneSchluesselwoerter sind die gezählten aus CR 702, die dieses Programm
// **nicht** durchsetzt.
//
// **Nach Häufigkeit**, damit die Reihenfolge einer Meldung nicht von einer
// Kartenreihenfolge abhängt.
var offeneSchluesselwoerter = []string{
	"flash", "devoid", "hexproof", "convoke", "changeling", "infect",
	"prowess", "swampwalk", "protection", "shadow", "cascade", "exalted",
	"shroud", "islandwalk", "forestwalk", "delve", "storm", "horsemanship",
	"soulbond", "wither", "persist", "intimidate", "improvise", "undying",
	"mountainwalk", "split second", "evolve", "extort", "mentor", "myriad",
	"riot", "shield counter", "skulk", "unleash", "afflict", "battle cry",
	"ingest", "melee", "plainswalk", "toxic", "backup", "bargain", "fear",
	"banding", "landwalk",
}

// OffeneSchluesselwoerter gibt die Fähigkeiten eines Regeltextes, die dieses
// Programm **kennt und nicht durchsetzt**.
//
// **Exportiert, damit die Zählung dieselbe Liste sieht** — derselbe Grund wie
// bei [SchluesselwoerterAus].
func OffeneSchluesselwoerter(text string) []string {
	offen := []string{}

	for _, wort := range offeneSchluesselwoerter {
		if enthaeltOffenes(text, wort) {
			offen = append(offen, wort)
		}
	}

	return offen
}

// enthaeltOffenes sagt, ob eine Zeile dieses offene Schlüsselwort trägt.
//
// **Großzügiger als [schluesselzeile]**, und das ist Absicht: dort entscheidet
// eine Aufzählung über eine **Wirkung** — ein Fehlgriff gäbe eine Fähigkeit,
// die niemand hat. Hier entscheidet sie über eine **Meldung**, und ein
// Fehlgriff kostet einen Satz zu viel. Die Richtung des Irrtums folgt der
// Richtung des Schadens (§10).
func enthaeltOffenes(text, wort string) bool {
	for _, zeile := range strings.Split(text, "\n") {
		ohne := ohneEingeklammertes(strings.ToLower(zeile), '(', ')')
		ohne = ohneEingeklammertes(ohne, '<', '>')

		for _, teil := range strings.Split(ohne, ",") {
			teil = strings.TrimSpace(strings.Trim(strings.TrimSpace(teil), "."))

			if teil == wort || istParametrisch(teil, wort) {
				return true
			}
		}
	}

	return false
}

// istParametrisch sagt, ob ein Teil dieses Wort mit einem Wert trägt.
//
// **„Schutzschild {2}" ist Schutzschild** — CR 702 führt Schlüsselwörter mit
// Parameter, und der Name steht vorn.
func istParametrisch(teil, wort string) bool {
	return len(teil) > len(wort)+1 && teil[:len(wort)+1] == wort+" "
}

// tischbefunde nennt, was auf dem Tisch liegt und nicht durchgesetzt wird.
//
// **Nur auf dem Schlachtfeld**: dort gelten statische Fähigkeiten (CR 604.3),
// und dort kostet ein stiller Ausfall etwas. Eine Karte auf der Hand tut
// nichts, was jemand vermissen könnte.
//
// **Und ohne Doppelung**: zwanzig Kreaturen mit Flugfähigkeit, von denen eine
// Schutzschild trägt, geben **einen** Satz und nicht zwanzig.
func (p Partie) tischbefunde() []string {
	var gefunden []string

	for _, objekt := range p.Objekte {
		if objekt.Zone != ZoneTisch {
			continue
		}

		for _, wort := range OffeneSchluesselwoerter(p.regeltext(objekt.Kennung)) {
			if !slices.Contains(gefunden, wort) {
				gefunden = append(gefunden, wort)
			}
		}
	}

	befunde := make([]string, 0, len(gefunden))
	for _, wort := range gefunden {
		befunde = append(befunde, fmt.Sprintf(
			"am Tisch liegt eine Karte mit %q — dieses Programm setzt diese "+
				"Schlüsselwortfähigkeit nicht durch (CR 702)", wort))
	}

	return befunde
}
