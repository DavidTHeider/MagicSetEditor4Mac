package app_test

// **Was gespeichert wurde, kommt beim Öffnen wieder** — C05
// (#2230).
//
// Die Rundläufe daneben prüfen ausgedachte Pakete Byte für Byte. Hier
// entstehen die Änderungen aus einem Samen: Karten anlegen, Felder setzen —
// mehrmals dasselbe, neue Felder, Umlaute, Zeichen außerhalb der
// Grundebene, Zeilenumbrüche, Doppelpunkte, führender Leerraum. Dann wird das
// Set gespeichert und in einem **frischen** Bestand geöffnet. Zugesagt ist:
// dieselben Karten in derselben Reihenfolge, je Karte dieselben Felder mit
// denselben Werten.
//
// **Was er nicht hält**: die Bytes der Datei (das tun die Rundläufe
// daneben) und Felder, die das Spiel ableitet — `probe` leitet keine ab.

import (
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

var (
	rundlauffelder = []string{"name", "farbe", "text", "regel text", "neu"}
	rundlaufwerte  = []string{
		"Feuerwand", "grün", "Äußerst «grün»", "漢字 🃏", "zwei\nZeilen",
		"Kosten: 2", "  führend", "1", "drei\nZeilen\nmit: Doppelpunkt", "ende ",
	}
)

// kartenstandAus macht aus einer Kartenantwort Zeilen „Karte n: feld=wert".
func kartenstandAus(t *testing.T, antwort app.KartenAntwort) []string {
	t.Helper()

	if antwort.Fehler != nil {
		t.Fatalf("Karten: %+v", antwort.Fehler)
	}

	var stand []string

	for nummer, karte := range antwort.Karten {
		for _, wert := range karte.Werte {
			stand = append(stand, fmt.Sprintf("Karte %d: %s=%q", nummer, wert.Feld, wert.Wert))
		}
	}

	return stand
}

func TestWasGespeichertWurdeKommtBeimOeffnenWieder(t *testing.T) {
	for samen := range uint64(30) {
		t.Run(fmt.Sprintf("Samen %d", samen), func(t *testing.T) {
			zufall := rand.New(rand.NewPCG(samen, 2230)) //nolint:gosec // eine Folge für einen Test, kein Geheimnis
			anwendung := gestartet(t)
			quelle := kopiereVerzeichnis(t, filepath.Join("..", "..", "testdaten", "probe.mse-set"))

			geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
			if geoeffnet.Fehler != nil {
				t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
			}

			setID := geoeffnet.SetID
			folge := []string{}

			for range 1 + zufall.IntN(12) {
				karten := anwendung.Karten(t.Context(), setID).Karten

				if zufall.IntN(4) == 0 {
					if angelegt := anwendung.KarteAnlegen(probekontext(t), setID); angelegt.Fehler != nil {
						t.Fatalf("Karte anlegen: %+v", angelegt.Fehler)
					}

					folge = append(folge, "+Karte")

					continue
				}

				karte := karten[zufall.IntN(len(karten))]
				feld := rundlauffelder[zufall.IntN(len(rundlauffelder))]
				wert := rundlaufwerte[zufall.IntN(len(rundlaufwerte))]

				if gesetzt := anwendung.FeldSetzen(probekontext(t), setID, karte.ID, feld, wert); gesetzt.Fehler != nil {
					t.Fatalf("%s=%q setzen: %+v", feld, wert, gesetzt.Fehler)
				}

				folge = append(folge, fmt.Sprintf("%s.%s=%q", karte.ID, feld, wert))
			}

			vorher := kartenstandAus(t, anwendung.Karten(t.Context(), setID))
			ziel := filepath.Join(t.TempDir(), "rundlauf.mse-set")

			if gespeichert := anwendung.SetSpeichern(t.Context(), setID, ziel); gespeichert.Fehler != nil {
				t.Fatalf("speichern: %+v", gespeichert.Fehler)
			}

			zweite := gestartet(t)

			wieder := zweite.SetOeffnen(t.Context(), ziel)
			if wieder.Fehler != nil {
				t.Fatalf("wieder öffnen: %+v", wieder.Fehler)
			}

			if nachher := kartenstandAus(t, zweite.Karten(t.Context(), wieder.SetID)); !slices.Equal(nachher, vorher) {
				t.Errorf("nach dem Rundlauf\n%q\nvorher\n%q\n(Folge %q)", nachher, vorher, folge)
			}
		})
	}
}
