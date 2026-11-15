package msescript

import "sync"

// **Derselbe Quelltext wird einmal zerteilt**
// (#672).
//
// # Warum es das gibt
//
// `magic.mse-game` bringt ein Initskript von über 10.000 Zeilen mit, und der
// Planbauer wertet es für **jede** Karte aus. Gemessen über 120 Vorschauen
// einer M15-Karte war [Zerteile] allein **8,8 %** der gesamten Rechenzeit —
// dieselbe Arbeit, immer wieder, für dasselbe Ergebnis.
//
// # Warum ein geteilter Baum sicher ist
//
// **Der Auswerter liest ihn nur.** Nachgesehen über das ganze Paket: außerhalb
// des Zerteilers steht keine Zuweisung an ein Feld von [Knoten]. Ein Baum, der
// zweimal ausgewertet wird, kommt beide Male gleich heraus — und zwei
// Auswertungen zugleich stören sich nicht, weil keine schreibt.
//
// **Was den Baum doch änderte, bräche das hier.** Wer eine Zuweisung an einen
// Knoten einbaut, hebt diesen Speicher auf oder gibt eine Kopie heraus.
//
// # Wann er ungültig wird
//
// **Gar nicht.** Der Schlüssel ist der Quelltext selbst; ein geänderter Text
// ist ein anderer Schlüssel. Anders als bei [msepaket.Bildquelle] gibt es
// keinen Fall, in dem derselbe Schlüssel etwas anderes bedeutet.
//
// **Ein Fehlschlag wird nicht gemerkt** — dieselbe Regel wie dort, und hier
// zusätzlich mangels Baum, den man aufheben könnte.
var (
	baumsperre     sync.RWMutex
	gemerkteBaeume = map[string]*Knoten{}

	// baumbelegung ist, was die gemerkten Bäume belegen (`baumbelegung.go`).
	baumbelegung int64
)

// Arbeitsmenge ist die **gemessene** Zahl verschiedener Quelltexte eines Sets.
//
// **561**, erhoben am 2026-08-29: `mse packarten 7` gegen `Vanguard.mse-set`
// (107 Karten, Spiel `magic`, Stil `m15-altered`, volles Pack) mit einem
// Zähler in [merkeBaum]. Ein Set, ein Spiel, ein Stil.
const Arbeitsmenge = 561

// GrenzeBaumspeicher ist die Zahl der Bäume, die vorgehalten werden.
//
// # Warum sie 4.096 ist und nicht 256
//
// **Hier stand 256, und die Begründung war geraten**: „`magic.mse-game`
// erklärt 249 Kartenfelder; ein Stil zeichnet davon ein gutes Dutzend, und die
// Grenze ist damit weit genug, dass ein Set sie nicht erreicht."
//
// Gemessen am 2026-08-29 an **einem** `mse packarten 7`:
//
//	Grenze 256:   316 Räumungen, 81.136 frisch zerteilt
//	Grenze 4096:    0 Räumungen,     561 frisch zerteilt
//
// Der Speicher lief voll, räumte **ganz** und baute alles wieder auf — 316
// Mal in einem einzigen Aufruf. **Ein Zwischenspeicher, der öfter räumt als
// er trifft, ist keiner**; er kostete die Zeit des Zerteilens **und** die des
// Nachschlagens. Der Aufruf brauchte damit 3,90 s Rechenzeit statt 3,46 s
// (#1288).
//
// **Vier Kilo, weil die Arbeitsmenge eines Sets 561 ist** ([Arbeitsmenge]) und
// mehrere Sets zugleich offen sein können. Sieben Sets mit je eigenem Spiel
// und Stil passen hinein.
//
// **Sie räumt weiter ganz, nicht klug.** Dieselbe Wahl wie bei
// [msepaket.Bildquelle]: eine Verdrängung nach Alter verlangt Buchführung über
// jeden Zugriff. Was sich geändert hat, ist nicht die Strategie, sondern die
// Einsicht, wie groß ein Arbeitssatz wirklich ist.
const GrenzeBaumspeicher = 4096

// gemerkterBaum gibt den Baum zu einem Quelltext, wenn er schon dasteht.
func gemerkterBaum(quelle string) (*Knoten, bool) {
	baumsperre.RLock()
	defer baumsperre.RUnlock()

	baum, da := gemerkteBaeume[quelle]

	return baum, da
}

// merkeBaum legt einen Baum ab und räumt an einer Grenze — der Zahl oder den
// Bytes. **Ein Baum über der ganzen Bytegrenze wird nicht gemerkt**, aus
// demselben Grund wie bei [msepaket.Bildquelle].
func merkeBaum(quelle string, baum *Knoten) {
	groesse := baumBytes(quelle, baum)

	baumsperre.Lock()
	defer baumsperre.Unlock()

	if groesse > baumgrenzeBytes {
		return
	}

	// **Zweimal derselbe Quelltext zählt einmal** — zwei Läufe, die ihn
	// zugleich nicht fanden, zerteilen und merken ihn beide.
	if vorher, da := gemerkteBaeume[quelle]; da {
		baumbelegung -= baumBytes(quelle, vorher)
	}

	if len(gemerkteBaeume) >= GrenzeBaumspeicher || baumbelegung+groesse > baumgrenzeBytes {
		gemerkteBaeume = make(map[string]*Knoten, GrenzeBaumspeicher)
		baumbelegung = 0
	}

	gemerkteBaeume[quelle] = baum
	baumbelegung += groesse
}
