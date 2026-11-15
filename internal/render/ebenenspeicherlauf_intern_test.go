package render

import (
	"image"
	"sync"
	"testing"
)

// **Zwei Läufe teilen sich den Ebenenspeicher** —
// #748.
//
// # Warum die Naht und nicht der ganze Weg
//
// Ein Wächter über zwei vollständige Rasterläufe steht daneben
// (`rastererlauf_test.go`) und fängt die **Lauffelder** — `gefragteGroesse` und
// `baumbefunde`. Den Ebenenspeicher betritt er nicht zuverlässig: ob eine Ebene
// überhaupt gemerkt wird, hängt am Bildtyp, an der Größe und daran, ob der
// Eintrag schon dasteht. Ein Wächter, dessen Fall vom Zufall abhängt, ist
// keiner (§10) — und der erste Anlauf hat das dreimal vorgeführt: die
// Gegenprobe ohne Sperre blieb grün, weil `merkeEbenenbild` **null**-mal lief.
//
// Hier wird die Ablage deshalb unmittelbar betreten, mit verschiedenen
// Schlüsseln aus beiden Läufen. Das ist eine Aussage über die Ablage und
// **nicht** darüber, dass der Zeichenweg sie erreicht — das steht daneben.
//
// # Was ohne Sperre geschähe
//
// Zwei schreibende Zugriffe auf eine Go-`map` sind kein falsches Bild, sondern
// ein Abbruch des Programms.
func TestZweiLaeufeSchreibenDenEbenenspeicher(t *testing.T) {
	t.Parallel()

	r := NeuerRasterer(nil)

	var warte sync.WaitGroup

	warte.Add(2)

	for lauf := range 2 {
		go func() {
			defer warte.Done()

			for stelle := range 60 {
				// **Verschiedene Größen je Lauf**, damit beide **schreiben**.
				// Bei gleichen Schlüsseln fände der zweite den Eintrag des
				// ersten und legte nie einen an.
				seite := 4 + lauf*100 + stelle
				r.abgetasteteEbene(
					image.NewNRGBA(image.Rect(0, 0, 2, 2)),
					image.Rect(0, 0, seite, seite), GueteMittel, true,
				)
			}
		}()
	}

	warte.Wait()
}
