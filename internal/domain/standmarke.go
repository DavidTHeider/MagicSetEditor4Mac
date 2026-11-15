// **Die Marke über einen Stand** — Werkstück O3 von Phase 5.2
// (#1427),
// nach [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md).
//
// # Wozu
//
// **Damit eine Übernahme gegen den Stand geht, den jemand gesehen hat**
// (§7, Regel 2). Zwischen der Vorschau und dem Klick auf „übernehmen" liegen
// Sekunden, und in ihnen kann eine andere Fläche geschrieben haben — der
// MCP-Server schreibt auf dieselbe Datei
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). Wer
// dann übernähme, überschriebe eine Änderung, die er nie gesehen hat.
//
// # Warum eine Marke und kein Zähler
//
// **Weil ein Zähler in der Tafel stehen müsste.** Vermerk 0078 hat für die
// Rechte genau das getan — und dort steigt er in **einem** Auslöser. Hier wären
// es **zwei** Tafeln, `settings` und `nutzereinstellungen`, und jeder Schreibweg
// müsste ihn treffen; wer einen vergäße, hätte einen Stand, der sich ändert,
// ohne dass es jemand sieht. Eine Marke über die Werte selbst kann das nicht:
// sie **ist** der Stand, nicht eine Aussage darüber.
//
// **Und sie fasst mit, wer wo arbeitet.** Ein Profilwechsel zwischen Vorschau
// und Übernahme wäre sonst nur dann zu sehen, wenn die Werte drüben andere
// sind — bei zwei frischen Profilen also gar nicht.
//
// # Was sie nicht ist
//
// **Kein Geheimnis und kein Nachweis.** Sie geht über die Naht und kommt
// zurück; wer sie fälscht, bekommt keinen Zutritt, sondern einen Konflikt an
// der falschen Stelle. Was jemand **darf**, entscheidet der Prüfer (0078).

package domain

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
)

// MarkeLaenge ist die Zahl der Hexzeichen, die eine Marke trägt.
//
// **Sechzehn Zeichen sind 64 Bit.** Verglichen werden zwei Marken desselben
// Bestands im Abstand von Sekunden; eine zufällige Gleichheit zweier
// verschiedener Stände wäre ein Ereignis, auf das man länger wartet als dieses
// Programm existiert. Der ganze Hash wäre 64 Zeichen und stünde in jeder
// Antwort und jedem Befehlsaufruf.
const MarkeLaenge = 16

// Standmarke gibt eine kurze Marke über einen Stand der Einstellungen.
func Standmarke(profilID, nutzerID int64, profil, nutzer map[string]string) string {
	bau := &strings.Builder{}

	fmt.Fprintf(bau, "profil=%d\nnutzer=%d\n", profilID, nutzerID)

	// **Je Ebene ein benannter Block.** Ohne den Namen ergäbe ein Wert im
	// Profil dieselbe Marke wie derselbe Wert beim Menschen — zwei Lagen, die
	// sich verschieden verhalten.
	for _, ebene := range []struct {
		name  string
		werte map[string]string
	}{{"profil", profil}, {"nutzer", nutzer}} {
		fmt.Fprintf(bau, "[%s]\n", ebene.name)

		for _, schluessel := range sortiert(ebene.werte) {
			fmt.Fprintf(bau, "%s=%s\n", schluessel, ebene.werte[schluessel])
		}
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(bau.String())))[:MarkeLaenge]
}

// sortiert gibt die Schlüssel einer Abbildung in fester Reihenfolge.
func sortiert(werte map[string]string) []string {
	schluessel := make([]string, 0, len(werte))
	for einer := range werte {
		schluessel = append(schluessel, einer)
	}

	slices.Sort(schluessel)

	return schluessel
}
