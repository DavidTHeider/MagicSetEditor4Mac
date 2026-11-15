// **Wie der Tisch beschriftet ist** — seit dem 2026-09-24, auf die Frage, ob
// die Namen der Zonen und die Köpfe der Sitze („Bo · 20 Leben") abschaltbar
// und in der Größe einstellbar sind.
//
// # Der Anlass
//
// **Sie waren es nicht**, und sie standen schlecht: der Kopf eines Sitzes
// lag auf dem Namen der ersten Zone, der Name einer Zone auf ihrer ersten
// Karte, und beides war bei der Anfangskamera rund sechs Punkte hoch. Seitdem
// haben beide ein eigenes Band (`frontend/src/tisch/masse.ts`); wie groß die
// Schrift darin steht — oder ob sie dasteht —, sagt diese Einstellung.
//
// # Warum „aus" die Anzahlen stehen lässt
//
// **Weil eine Anzahl eine Auskunft ist und kein Etikett.** Eine Bibliothek
// zeigt ihre Karten nicht; wie viele es sind, steht nur an ihr — und
// CR 104.3c entscheidet eine Partie daran (Fassung 2026-08-19). Namen und
// Köpfe dagegen stehen auch anderswo: der Kopf in der Spielerleiste neben
// dem Tisch, der Name einer Zone in jedem Kartenknopf, den eine Vorlesehilfe
// liest.

package domain

// EinstellungTischbeschriftung sagt, ob und wie groß der Tisch Zonen und
// Sitze beschriftet.
const EinstellungTischbeschriftung = "tisch.beschriftung"

// Die vier Werte von [EinstellungTischbeschriftung].
const (
	// BeschriftungAus zeigt weder Zonennamen noch Sitzköpfe — die Anzahlen
	// bleiben.
	BeschriftungAus = "aus"

	// BeschriftungKlein, BeschriftungNormal und BeschriftungGross sind drei
	// Größen derselben Schrift.
	BeschriftungKlein  = "klein"
	BeschriftungNormal = "normal"
	BeschriftungGross  = "gross"
)

// init hängt die Beschriftung an die Einstellungsreihe — wie [init] in
// `autoanordnung.go`, und direkt dahinter: Go ordnet `init` nach
// Dateinamen, und beide gehören zum Tisch.
func init() {
	einstellungen = append(einstellungen, einstellung{
		Schluessel: EinstellungTischbeschriftung,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte: []string{
			BeschriftungAus, BeschriftungKlein, BeschriftungNormal,
			BeschriftungGross,
		},
		Vorgabe: BeschriftungNormal,
	})
}
