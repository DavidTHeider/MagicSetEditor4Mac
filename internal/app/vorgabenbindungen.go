// **Die Bindungen der Profilvorgaben** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.

package app

import "context"

// Profilvorgaben gibt die Vorgaben eines Profils samt ihrem Angebot.
//
// **Beides in einer Antwort**, weil die Seite beides zugleich zeigt: was
// gewählt ist und was sich wählen ließe (§7, Regel 2).
func (b *Bindungen) Profilvorgaben(ctx context.Context, profilID string) VorgabenAntwort {
	return b.app(ctx).Profilvorgaben(b.ruf(ctx), profilID)
}

// ProfilvorgabeSetzen ersetzt die Werte **einer** Art und gibt den ganzen
// Stand zurück.
func (b *Bindungen) ProfilvorgabeSetzen(ctx context.Context, angaben VorgabeangabenDTO) VorgabenAntwort {
	return b.app(ctx).ProfilvorgabeSetzen(b.ruf(ctx), angaben)
}
