// Die Ablage des Probelaufs einer Ausfuhr — UXR-F7
// (#2015).
//
// **Sie hält fest, was geschrieben würde**, und schreibt es nicht. Der Name
// ist der Schlüssel: ein Name, den die Vorlage zweimal nennt, ist eine Datei,
// wie der echte Lauf sie überschriebe.
package service_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieProbeablageZaehltDateienUndBekannteBytes(t *testing.T) {
	t.Parallel()

	ablage := service.NeueProbeablage()
	ablage.Lege("liste.txt", 120)
	ablage.Lege("bild.png", 30)

	stand := ablage.Stand()
	if stand.Dateien != 2 || stand.Bytes != 150 || stand.Geschaetzt {
		t.Errorf("Stand: %+v, erwartet 2 Dateien, 150 Bytes, nichts geschätzt", stand)
	}
}

// **Derselbe Name ist dieselbe Datei.** Der echte Lauf überschriebe sie; zweimal
// gezählt wäre die Zahl der Beigaben größer, als je auf der Platte läge.
func TestDieProbeablageZaehltEinenNamenEinmal(t *testing.T) {
	t.Parallel()

	ablage := service.NeueProbeablage()
	ablage.Lege("a.txt", 10)
	ablage.Lege("a.txt", 25)

	stand := ablage.Stand()
	if stand.Dateien != 1 || stand.Bytes != 25 {
		t.Errorf("Stand: %+v, erwartet eine Datei mit den letzten 25 Bytes", stand)
	}
}

// **Eine Datei ohne bekannte Größe zählt, aber sie wiegt nichts — und das
// Ergebnis sagt, dass es geschätzt ist.**
func TestEineDateiOhneGroesseMachtDieSummeGeschaetzt(t *testing.T) {
	t.Parallel()

	ablage := service.NeueProbeablage()
	ablage.Lege("a.txt", 10)
	ablage.LegeOhneGroesse("karte.png")

	stand := ablage.Stand()
	if stand.Dateien != 2 || stand.Bytes != 10 || !stand.Geschaetzt {
		t.Errorf("Stand: %+v, erwartet 2 Dateien, 10 Bytes, geschätzt", stand)
	}

	// **Auch eine spätere bekannte Größe nimmt das nicht zurück**: wer
	// denselben Namen zuerst ohne und dann mit Größe nennt, hat die Datei
	// einmal überschrieben, und die letzte Größe gilt.
	ablage.Lege("karte.png", 7)

	if stand := ablage.Stand(); stand.Geschaetzt || stand.Bytes != 17 {
		t.Errorf("Stand: %+v, erwartet 17 Bytes ohne Schätzung", stand)
	}
}

// **Ein Bildziel wird im Probelauf wie im Lauf geprüft.**
//
// Dieselbe Eingabe an `Bildausfuhr.Schreibe` und an `PruefeBildziel` gibt
// denselben Satz — es ist **eine** Prüfung, die beide rufen.
func TestPruefeBildzielGibtDieFehlerVonSchreibe(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		datei   string
		format  service.Bildformat
		breite  int
		istFehl bool
	}{
		"ein Pfad statt eines Namens": {datei: "../x.png", format: service.FormatPNG, istFehl: true},
		"kein Name":                   {datei: "", format: service.FormatPNG, istFehl: true},
		"kein Bildformat":             {datei: "x.png", format: "gif", istFehl: true},
		"eine Breite über der Grenze": {
			datei: "x.png", format: service.FormatPNG, breite: service.GrenzeZielbreite + 1, istFehl: true,
		},
		"ein brauchbares Ziel": {datei: "x.png", format: service.FormatPNG, breite: 750},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pruefung := service.PruefeBildziel(fall.datei, fall.format, fall.breite)

			_, lauf := bildausfuhr(t, planDouble{}).Schreibe(context.Background(), 1,
				t.TempDir(), fall.datei, fall.format, fall.breite, domain.BildkonfliktUeberschreiben, render.GueteMittel)

			if !fall.istFehl {
				if pruefung != nil {
					t.Errorf("PruefeBildziel: %v", pruefung)
				}

				return
			}

			if pruefung == nil || lauf == nil || pruefung.Error() != lauf.Error() {
				t.Errorf("Prüfung %v, Lauf %v", pruefung, lauf)
			}
		})
	}
}
