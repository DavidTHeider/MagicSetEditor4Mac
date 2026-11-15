package fenster

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

// Speicherstand ist, was der Go-Heap nach einer Bereinigung über sich sagt —
// für den Dauerbetrieb (E06,
// #2278).
//
// **Zwei Fragen, zwei Zahlen**: wächst `HeapAlloc` über einen Lauf, hält der
// Prozess etwas fest; wächst nur `HeapIdle − HeapReleased`, hat er Speicher,
// den er dem System noch nicht zurückgegeben hat. Der RSS allein, den der
// Dauerlauf misst, unterscheidet beides nicht.
type Speicherstand struct {
	HeapAlloc     uint64 `json:"heapAlloc"`
	HeapInuse     uint64 `json:"heapInuse"`
	HeapIdle      uint64 `json:"heapIdle"`
	HeapReleased  uint64 `json:"heapReleased"`
	HeapSys       uint64 `json:"heapSys"`
	Sys           uint64 `json:"sys"`
	Objekte       uint64 `json:"objekte"`
	Bereinigungen uint32 `json:"bereinigungen"`
}

// speicherabzug räumt den Heap auf, liest seinen Stand und schreibt — wenn
// ein Pfad dasteht — ein Heap-Profil dorthin, für `go tool pprof -base`.
//
// **Erst bereinigen, dann zählen**: ohne `runtime.GC` zählte `HeapAlloc`
// Müll mit, der nur noch nicht eingesammelt ist, und zwei Abzüge eines
// Laufs sagten über ihn nichts.
func speicherabzug(pfad string) (Speicherstand, error) {
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stand := Speicherstand{
		HeapAlloc: m.HeapAlloc, HeapInuse: m.HeapInuse, HeapIdle: m.HeapIdle,
		HeapReleased: m.HeapReleased, HeapSys: m.HeapSys, Sys: m.Sys,
		Objekte: m.HeapObjects, Bereinigungen: m.NumGC,
	}

	if pfad == "" {
		return stand, nil
	}

	datei, fehler := os.OpenFile(pfad, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- der Ort kommt von der Probe, nur im Abnahmebau
	if fehler != nil {
		return Speicherstand{}, fmt.Errorf("das Heap-Profil anlegen: %w", fehler)
	}

	// **Die Zahlen gehen mit, auch wenn das Profil scheitert**: die Leitung
	// meldet dann nur den Fehler (`antwortAus`), und der sagt, dass die Datei
	// fehlt.
	return stand, profilAblegen(datei)
}

// profilAblegen schreibt das Heap-Profil und schließt das Ziel — der erste
// Fehler zählt, auch einer beim Schließen: erst dann steht die Datei.
func profilAblegen(ziel io.WriteCloser) error {
	schreibfehler := pprof.Lookup("heap").WriteTo(ziel, 0)
	if schliessfehler := ziel.Close(); schreibfehler == nil {
		schreibfehler = schliessfehler
	}

	if schreibfehler != nil {
		return fmt.Errorf("das Heap-Profil schreiben: %w", schreibfehler)
	}

	return nil
}
