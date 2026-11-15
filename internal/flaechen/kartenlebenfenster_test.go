package flaechen_test

// **Der Kartenlebenslauf am Fenster** — C04
// (#2229),
// Zweitblick über Phase C, Befund 3: der Lauf fuhr Kommandozeile, MCP und
// Dienst, und das Fenster — die Fläche, an der Menschen Karten anlegen —
// fehlte.
//
// **Über die Bindungen, wie Wails sie ruft** (`fensterAus`, mit Toren und
// Rufkontext), und die Antwort **serialisiert, wie Wails sie serialisiert**:
// gelesen wird sie von denselben Lesern wie die der übrigen Flächen.

import (
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func kartenlaufFenster(t *testing.T, umgebung []string) kartenstand {
	t.Helper()

	b := fensterAus(t, umgebung)
	ctx := t.Context()

	geoeffnet := b.SetAusDatei(ctx, probeset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("SetAusDatei: %+v", geoeffnet.Fehler)
	}

	setID := geoeffnet.SetID

	liste := func() string {
		t.Helper()

		antwort := b.Karten(ctx, setID)
		if antwort.Fehler != nil {
			t.Fatalf("Karten: %+v", antwort.Fehler)
		}

		roh, fehler := json.Marshal(antwort)
		if fehler != nil {
			t.Fatalf("Karten serialisieren: %v", fehler)
		}

		return string(roh)
	}

	zaehle := func() []string { return kartenkennungen(t, liste()) }

	gelang := func(was string, fehler *app.Fehler) {
		t.Helper()

		if fehler != nil {
			t.Fatalf("%s: %+v", was, fehler)
		}
	}

	vorher := zaehle()
	vorbild := kartenname(t, liste(), vorher[0])

	gelang("KarteAnlegen", b.KarteAnlegen(ctx, setID).Fehler)
	nachAnlegen := zaehle()

	gelang("KarteKopieren", b.KarteKopieren(ctx, setID, vorher[0]).Fehler)
	nachKopieren := zaehle()
	namen := namenGezaehlt(t, liste())

	gelang("KartenTauschen", b.KartenTauschen(ctx, setID, nachKopieren[0], nachKopieren[1]).Fehler)
	nachTausch := zaehle()

	gelang("KarteLoeschen", b.KarteLoeschen(ctx, setID, nachTausch[0], false).Fehler)
	nachLoeschen := zaehle()

	return kartenstand{
		vorher: len(vorher), nachAnlegen: len(nachAnlegen),
		nachKopieren: len(nachKopieren), nachTausch: len(nachTausch),
		nachLoeschen: len(nachLoeschen),
		getauscht:    nachTausch[0] == nachKopieren[1] && nachTausch[1] == nachKopieren[0],

		vorbild: vorbild, namenNachKopieren: namen,
	}
}
