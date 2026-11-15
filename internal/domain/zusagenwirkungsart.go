package domain

// **Was ein Zustandswechsel an den festgehaltenen Exemplaren bedeutet** —
// Befund 3 des Nachlaufs zum Zweitblick über Phase 5.6
// (#1477).
//
// # Warum das ein Wert ist und keine Handlung
//
// **Weil beides in **einer** Transaktion stehen muss.** Der Vermerk über eine
// Zustellung, die neue Revision und die Wirkung an den Zusagen gehören
// zusammen: geschah die Wirkung vorher für sich, blieb nach einem
// gescheiterten Schreiben eine Zusage als „übergeben" stehen, während der
// Vorgang noch auf `vorbereitet` stand — und `Freigeben` gab sie nie wieder
// frei.
//
// **Der Dienst entscheidet, der Bestand führt aus.** Deshalb reist die
// Entscheidung als Wert bis dorthin, wo die Transaktion ist.

// Zusagenwirkungsart ist, was mit den Zusagen eines Vorgangs geschieht.
type Zusagenwirkungsart string

const (
	// ZusagenBleiben heißt: nichts geschieht — der Normalfall.
	ZusagenBleiben Zusagenwirkungsart = ""

	// ZusagenUebergeben heißt: sie sind über den Tisch gegangen, und erst
	// dadurch findet die Verbuchung ihre Abgangszeilen.
	ZusagenUebergeben Zusagenwirkungsart = "uebergeben"

	// ZusagenFreigeben heißt: der Handel ist zu Ende, ohne dass etwas
	// übergeben wurde.
	ZusagenFreigeben Zusagenwirkungsart = "freigeben"
)

// ZusagenwirkungFuer sagt, was dieser Wechsel an den Zusagen bedeutet.
//
//   - Nach `uebergabeOffen` sind die Exemplare über den Tisch gegangen.
//   - Ein Ende ohne Übergabe gibt frei, was niemand mehr braucht.
//   - Alles andere lässt sie in Ruhe, und das ist die Mehrzahl.
//
// **Am Wechsel und nicht am Schritt**: derselbe Wechsel entsteht auf zwei
// Wegen — als Schritt von Hand und als Nachricht von drüben —, und er
// bedeutet beide Male dasselbe.
func ZusagenwirkungFuer(vorher, nachher Handelszustand) Zusagenwirkungsart {
	if nachher == vorher {
		return ZusagenBleiben
	}

	switch nachher {
	case HandelUebergabeOffen:
		return ZusagenUebergeben

	case HandelAbgelehnt, HandelAbgebrochen, HandelAbgelaufen:
		return ZusagenFreigeben
	}

	return ZusagenBleiben
}
