package rest

// **Ein Vermerk geht nach seiner Frist** (#2002) — sonst wüchse das Buch mit
// jeder Ausgabe eines Dienstes, der Wochen läuft. **Und nicht vorher**: ein
// Vermerk, der vor seiner Datei ginge, sperrte den Eigner aus.

import (
	"testing"
	"time"
)

func TestEinVermerkGehtNachSeinerFrist(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC)
	buch := neuesAusgangsbuch(func() time.Time { return jetzt })

	buch.vermerke("alt.pdf", "marke")

	jetzt = jetzt.Add(vermerkfrist)
	buch.vermerke("frisch.pdf", "marke")

	if !buch.gehoert("alt.pdf", "marke") {
		t.Error("ein Vermerk genau an seiner Frist ging schon")
	}

	jetzt = jetzt.Add(time.Second)
	buch.vermerke("noch-einer.pdf", "marke")

	if buch.gehoert("alt.pdf", "marke") {
		t.Error("ein Vermerk über seiner Frist blieb")
	}

	if !buch.gehoert("frisch.pdf", "marke") || buch.gehoert("frisch.pdf", "anderer") {
		t.Error("der frische Vermerk gehört nicht genau seinem Eigner")
	}
}
