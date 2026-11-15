package app

import (
	"context"
	"sync"
	"testing"
	"time"
)

// **Die drei Wege der abbrechbaren Sperre**: frei, nach Warten erhalten, und
// aufgegeben — dann gibt sie die Sperre zurück, sobald sie sie bekommt.
func TestDieSperreMitKontext(t *testing.T) {
	t.Parallel()

	var sperre sync.Mutex

	if !sperreMitKontext(t.Context(), sperre.TryLock, sperre.Lock, sperre.Unlock) {
		t.Fatal("eine freie Sperre wurde nicht genommen")
	}

	// **Gehalten, dann frei, bevor der Ruf aufgibt**: er bekommt sie.
	go func() {
		time.Sleep(50 * time.Millisecond)
		sperre.Unlock()
	}()

	if !sperreMitKontext(t.Context(), sperre.TryLock, sperre.Lock, sperre.Unlock) {
		t.Fatal("eine Sperre, die frei wurde, wurde nicht genommen")
	}

	// **Gehalten, und der Ruf gibt auf**: kein Erhalt — und danach ist sie
	// wieder frei, obwohl die wartende Goroutine sie zwischendurch bekam.
	abbruch, abbrechen := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer abbrechen()

	if sperreMitKontext(abbruch, sperre.TryLock, sperre.Lock, sperre.Unlock) {
		t.Fatal("eine gehaltene Sperre galt als genommen")
	}

	sperre.Unlock()

	frist := time.Now().Add(time.Second)
	for !sperre.TryLock() {
		if time.Now().After(frist) {
			t.Fatal("die aufgegebene Sperre blieb gehalten")
		}

		time.Sleep(time.Millisecond)
	}
}
