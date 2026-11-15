// **Der Zugang zur Kartensuche** — Werkstück 1 von Phase 3.6
// (#1167).
//
// **Eine eigene Datei**, weil `scryfall.go` mit ihr über 400 Zeilen ginge
// (§2) — und weil der Schnitt einer Frage folgt: dort steht der Abgleich, hier
// die Suche.

package app

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Suchnaht baut die Kartensuche samt Klient und Zwischenspeicher (#1167).
//
// **Derselbe Klient wie Auskunft, Abgleich und Bildweg** — und aus demselben
// Grund: er bringt beide Drosseln mit, und zwei Drosseln unter derselben
// Grenze halten sie nicht ein.
//
// **Kein Bulkrückfall.** Die Sammeldatei `oracle_cards` führt je Karte einen
// Eintrag, nicht je Druck, und sie kennt Scryfalls Suchsyntax nicht. Eine
// Suche ohne Netz wäre eine andere Suche — und §16 Regel 5 verlangt, dass die
// Anwendung ohne Netz **weiterarbeitet**, nicht dass jede Auskunft offline
// dieselbe ist.
//
// # Seit Phase 4.8 verteilt sie
//
// **Alle vier Dienste können suchen** (Vermerk 0063), und welcher es tut,
// steht in den Einstellungen. Der Verteiler sitzt **hier** und nicht an den
// Flächen: Fenster, Kommandozeile, MCP und Dienst rufen diese Naht, und damit
// erben sie die Fähigkeit, ohne dass eine von ihnen angefasst wird (§16).
//
// **Ohne Zuhause gibt es keinen Scryfall-Klienten** — und das hält den ganzen
// Weg nicht mehr auf: die anderen drei brauchen keines, und wer einen von
// ihnen gewählt hat, soll suchen können.
func Suchnaht(
	wahl Dienstwahl, bund Geheimnisleser, frage Fragewahl,
) Suchquelle {
	// **Hier wird nichts mehr entschieden.** Der erste Anlauf fragte an dieser
	// Stelle, ob überhaupt ein Dienst bereitsteht, und gab sonst einen Fehler
	// zurück — eine Entscheidung zur **Verdrahtungszeit**, und das Fenster
	// verdrahtet einmal, beim Start. Wer danach seinen einzigen Dienst
	// einschaltete, hatte bis zum Neustart gar keine Suche: `a.suche` blieb
	// leer, weil diese Naht gescheitert war.
	//
	// **Dieselbe Gestalt wie Befund 1 des Zweitblicks**, eine Ebene höher.
	// [service.Dienstsuche] sagt jetzt je Anfrage, warum der gewählte Dienst
	// nicht antwortet — mit Namen und Grund.
	return service.NeueDienstsuche(Dienstquellen{wahl: wahl, bund: bund}, frage)
}
