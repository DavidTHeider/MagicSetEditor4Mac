package mcp

import (
	"bytes"
	"errors"
	"testing"
)

// **Wer vorzeitig aufhört, lässt den Leser zu Ende lesen** — `leeren` nimmt
// ab, was noch kommt, bis der Kanal zu ist.
func TestLeerenNimmtAbWasNochKommt(t *testing.T) {
	zeilen := make(chan []byte, 2)
	zeilen <- []byte("a")
	zeilen <- []byte("b")
	close(zeilen)

	leeren(zeilen)

	if len(zeilen) != 0 {
		t.Errorf("%d Zeilen liegen noch", len(zeilen))
	}
}

// **Nach dem Ende nimmt der Ausgang nichts mehr an** — BK-07c (#2117): ein
// Ruf, der nach dem Abbruch noch Fortschritt meldet, schriebe sonst in ein
// Protokoll, das schon beendet ist.
func TestDerAusgangNimmtNachDemEndeNichtsAn(t *testing.T) {
	t.Parallel()

	var geschrieben bytes.Buffer

	ausgang := &schliessbarerAusgang{aus: &geschrieben}
	if _, fehler := ausgang.Write([]byte("vorher\n")); fehler != nil {
		t.Fatal(fehler)
	}

	ausgang.schliesse()

	if _, fehler := ausgang.Write([]byte("danach\n")); !errors.Is(fehler, errBedienungBeendet) {
		t.Errorf("nach dem Ende: %v", fehler)
	}

	if geschrieben.String() != "vorher\n" {
		t.Errorf("geschrieben: %q", geschrieben.String())
	}
}
