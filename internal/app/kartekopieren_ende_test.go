package app_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Kopie steht in der geschriebenen Datei — hinter dem Original**
// (#625).
//
// Der Weg durch alle Schichten: kopieren, speichern, nachsehen. Ein Test gegen
// die Datenbank sagte nur, dass die Zeilen dastehen; ob sie **herauskommen**,
// entscheidet der Schreiber, und der ordnet über `cards.reihenfolge` zu.
func TestEineKopieStehtInDerGeschriebenenDatei(t *testing.T) {
	anwendung, setID := geoeffnetesProbeset(t)

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if kopie := anwendung.KarteKopieren(t.Context(), setID, karten.Karten[0].ID); kopie.Fehler != nil {
		t.Fatalf("kopieren: %+v", kopie.Fehler)
	}

	geschrieben := string(gespeichertesSet(t, anwendung, setID))

	if zahl := strings.Count(geschrieben, "card:"); zahl != len(karten.Karten)+1 {
		t.Errorf("%d Kartenblöcke, erwartet %d:\n%s", zahl, len(karten.Karten)+1, geschrieben)
	}

	// **Die kopierte Karte heißt „Feuerball", und jetzt zweimal.**
	if zahl := strings.Count(geschrieben, "Feuerball"); zahl != 2 {
		t.Errorf("Feuerball steht %dmal in der Datei, erwartet zweimal:\n%s", zahl, geschrieben)
	}

	// **Und die Kopie steht bei den Karten**, nicht hinter dem, was nicht zu
	// ihnen gehört.
	if strings.LastIndex(geschrieben, "Feuerball") >
		strings.Index(geschrieben, "unverstandener block:") {
		t.Errorf("die Kopie steht hinter dem unverstandenen Block:\n%s", geschrieben)
	}
}

// **Was hinzukam und zurückgenommen wurde, hinterlässt keine Spur**
// (#48).
//
// Zwei Wege, dieselbe Probe. Das **Anlegen** stand bis
// #626 gar
// nicht im Verlauf: ⌘Z antwortete mit „nichts zurückzunehmen", und die Karte
// blieb.
func TestEineZurueckgenommeneKarteHinterlaesstKeineSpur(t *testing.T) {
	for _, fall := range []struct {
		name        string
		hinzufuegen func(anwendung *app.App, setID, karteID string) *app.Fehler
	}{
		{
			name: "kopiert",
			hinzufuegen: func(anwendung *app.App, setID, karteID string) *app.Fehler {
				return anwendung.KarteKopieren(t.Context(), setID, karteID).Fehler
			},
		},
		{
			name: "angelegt",
			hinzufuegen: func(anwendung *app.App, setID, _ string) *app.Fehler {
				return anwendung.KarteAnlegen(t.Context(), setID).Fehler
			},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			anwendung, setID := geoeffnetesProbeset(t)

			karten := anwendung.Karten(t.Context(), setID)
			if karten.Fehler != nil {
				t.Fatalf("Karten: %+v", karten.Fehler)
			}

			if befund := fall.hinzufuegen(anwendung, setID, karten.Karten[0].ID); befund != nil {
				t.Fatalf("hinzufügen: %+v", befund)
			}

			// **Die Gegenprobe zuerst**: ohne sie prüfte der Test eine
			// Rücknahme, die nichts zurückzunehmen hatte.
			mehr := anwendung.Karten(t.Context(), setID)
			if len(mehr.Karten) != len(karten.Karten)+1 {
				t.Fatalf("%d Karten, erwartet %d", len(mehr.Karten), len(karten.Karten)+1)
			}

			zurueck := anwendung.Zuruecknehmen(t.Context(), setID, "")
			if zurueck.Fehler != nil {
				t.Fatalf("⌘Z: %+v", zurueck.Fehler)
			}

			if !zurueck.Bewegt {
				t.Fatal("⌘Z hat nichts bewegt — der Schritt steht nicht im Verlauf")
			}

			vorher := lies(t, filepath.Join(probesetpfad(t), "set"))
			nachher := gespeichertesSet(t, anwendung, setID)

			if !bytes.Equal(nachher, vorher) {
				t.Errorf("die Datei ist eine andere:\n%s", unterschied(vorher, nachher))
			}
		})
	}
}
