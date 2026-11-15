package render

import (
	"image"
	"sync"
	"testing"
	"time"
)

// **Zwei gleichzeitige Läufe füllen denselben Eintrag einmal** —
// #1133.
//
// # Der Befund
//
// Nachsehen und Eintragen nahmen die Sperre **getrennt**, und dazwischen lag
// die Rechnung. Zwei Läufe mit demselben Schlüssel verfehlten deshalb beide,
// rechneten beide dasselbe Bild und trugen es beide ein — **und jeder erhöhte
// `platz` erneut**.
//
// Kein falsches Bild und kein ungeschützter Zugriff auf die Map: beide
// rechnen dasselbe, und die Sperre um sie steht seit #748. Was es kostete:
//
//   - **doppelte teure Arbeit** — das Abtasten einer Rahmenebene ist genau
//     das, was der Speicher sparen soll,
//   - **eine falsche Kapazitätsrechnung** — dasselbe Bild mehrfach gezählt,
//     die Grenze von 64 MB zu früh erreicht, und dann wird der **ganze**
//     Speicher geräumt.
//
// Angestrichen vom Zweitblick über Phase Ω, der ihn ausdrücklich außerhalb der
// Phase eingeordnet hat: die Klasse stammt aus dem älteren Speicherdesign.
//
// # Woran es zu sehen ist
//
// **Am Zeiger.** Rechnen zwei Läufe getrennt, bekommt jeder sein eigenes Bild;
// teilen sie sich einen Eintrag, ist es für alle dasselbe.
//
// Der Paralleltest daneben (`ebenenspeicherlauf_intern_test.go`) fährt
// ausdrücklich **verschiedene** Schlüssel — er belegt Mapsicherheit, nicht das
// gleichzeitige Befüllen desselben Eintrags.
func TestGleichzeitigeLaeufeFuellenEinmal(t *testing.T) {
	const laeufe = 8

	rasterer := NeuerRasterer(nil)
	quelle := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	rahmen := image.Rect(0, 0, 120, 120)

	bilder := make([]*image.NRGBA, laeufe)

	var start, fertig sync.WaitGroup

	start.Add(1)
	fertig.Add(laeufe)

	for i := range laeufe {
		go func() {
			defer fertig.Done()

			start.Wait()

			bilder[i] = rasterer.abgetasteteEbene(quelle, rahmen, GueteMittel, true)
		}()
	}

	start.Done()
	fertig.Wait()

	for i := 1; i < laeufe; i++ {
		if bilder[i] != bilder[0] {
			t.Errorf("Lauf %d bekam ein eigenes Bild — es wurde doppelt gerechnet", i)
		}
	}

	// **Und der Platz zählt es einmal.** Ohne das erreicht die Grenze von
	// 64 MB zu früh, und dann wird der ganze Speicher geräumt.
	rasterer.ebenen.sperre.Lock()
	platz := rasterer.ebenen.platz
	rasterer.ebenen.sperre.Unlock()

	if gewollt := len(bilder[0].Pix); platz != gewollt {
		t.Errorf("platz = %d, erwartet %d — dasselbe Bild wurde mehrfach gezählt",
			platz, gewollt)
	}
}

// **Ein zweiter Lauf wartet auf den ersten** — und das hier ohne Zufall.
//
// `TestGleichzeitigeLaeufeFuellenEinmal` betritt den Wartezweig nur, wenn
// sich die Läufe **zeitlich überlappen**. Gemessen am 2026-09-25: in einer
// Vorprüfung ohne Detektor taten sie es nicht, und drei Anweisungen standen
// unbetreten da — ein Wächter, dessen Fall vom Zufall abhängt, ist keiner
// (§10). Hier steht der erste Lauf fest „unterwegs", bevor der zweite fragt.
func TestEinZweiterLaufWartetAufDenErsten(t *testing.T) {
	r := NeuerRasterer(nil)
	quelle := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	rahmen := image.Rect(0, 0, 120, 120)

	schluessel, merkbar := ebenenschluesselVon(quelle, rahmen, GueteMittel)
	if !merkbar {
		t.Fatal("die Quelle lässt sich nicht merken — der Fall liefe ins Leere")
	}

	// **Der erste Lauf trägt sich ein und rechnet noch.**
	if bild, lauf := r.eintragOderLauf(schluessel); bild != nil || lauf != nil {
		t.Fatal("der Speicher war nicht leer")
	}

	zweiter := make(chan *image.NRGBA, 1)

	go func() { zweiter <- r.abgetasteteEbene(quelle, rahmen, GueteMittel, true) }()

	select {
	case <-zweiter:
		t.Fatal("der zweite Lauf wartete nicht auf den ersten")
	case <-time.After(50 * time.Millisecond):
	}

	gerechnet := image.NewNRGBA(rahmen)

	r.ebenen.sperre.Lock()
	r.ebenen.unterwegs[schluessel].bild = gerechnet
	r.ebenen.sperre.Unlock()

	r.laufBeenden(schluessel)

	if bekommen := <-zweiter; bekommen != gerechnet {
		t.Error("der zweite Lauf bekam ein eigenes Bild statt des ersten")
	}
}
