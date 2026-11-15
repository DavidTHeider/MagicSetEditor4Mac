package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// BulkError nennt die Karten, an denen eine Massenänderung gescheitert ist.
//
// **Der Grund, warum es diesen Typ gibt:** eine Massenänderung läuft in einer
// Transaktion, und wenn sie scheitert, ist *nichts* geschrieben. Dann ist die
// einzige brauchbare Auskunft, *welche* Karten daran schuld sind — „die
// Änderung ging schief" lässt jemanden zwanzig Karten einzeln durchprobieren.
//
// Geprüft wird mit `errors.As`, nie am Text (§8).
type BulkError struct {
	CardIDs []int64
	Err     error
}

func (e *BulkError) Error() string {
	namen := make([]string, 0, len(e.CardIDs))
	for _, kennung := range e.CardIDs {
		namen = append(namen, strconv.FormatInt(kennung, 10))
	}

	return fmt.Sprintf("%v: die Karten %s", e.Err, strings.Join(namen, ", "))
}

// Unwrap gibt den zugrunde liegenden Fehler frei — `errors.Is` findet darüber
// `ErrNotFound`.
func (e *BulkError) Unwrap() error { return e.Err }
