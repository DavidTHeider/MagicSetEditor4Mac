package msepaket_test

// **Auch der Schreibweg liest mit Budget** — Befund 4 des Zweitblicks über
// Phase 6.1 (FMT-06.AK-01).
//
// # Der Befund
//
// Der Leser hielt seit FMT-06 an der Grenze einer Paketdatei an. Der
// Schreibweg daneben las die Datei, die er ersetzen soll, mit `os.ReadFile`
// oder `io.ReadAll` ganz ein — **bevor** er ihren Stand verglich. Eine Datei,
// die seit dem Öffnen ausgetauscht wurde oder im Archiv mehr wird, als ihr
// Kopf sagt, belegte damit beliebig viel Speicher, und zurück kam erst danach
// ein Konflikt. Dasselbe beim Lesen einer Symbolzeichnung.
//
// # Wie geprüft wird
//
// **An der Zuteilung und nicht nur am Fehler**: ein Lauf, der „zu groß"
// meldet, nachdem er alles gelesen hat, wäre grün und hätte nichts
// gehalten. Gezählt wird, was der Ruf zuteilt — höchstens 64 MiB, bei einer
// Datei von mehr als 256.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// zugeteilt misst, was ein Ruf an Speicher zuteilt.
func zugeteilt(ruf func()) uint64 {
	var vorher, nachher runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&vorher)

	ruf()

	runtime.ReadMemStats(&nachher)

	return nachher.TotalAlloc - vorher.TotalAlloc
}

// duenneDatei legt eine Datei an, die größer ist als die Grenze, ohne sie zu
// füllen — `Truncate` gibt eine Datei mit Loch, und die Platte bleibt leer.
func duenneDatei(t *testing.T, pfad string) {
	t.Helper()

	if err := os.WriteFile(pfad, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Truncate(pfad, msepaket.GrenzePaketdatei+1); err != nil {
		t.Fatal(err)
	}
}

func TestDasSchreibenInsVerzeichnisLiestKeineUebergrosseDatei(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "probe.mse-set")
	if err := os.Mkdir(ort, 0o750); err != nil {
		t.Fatal(err)
	}

	duenneDatei(t, filepath.Join(ort, "set"))

	var fehler error

	belegt := zugeteilt(func() {
		fehler = msepaket.DateiSchreiben(t.Context(), ort, "set", []byte("neu"), "egal")
	})

	if !errors.Is(fehler, msepaket.ErrZuGross) {
		t.Errorf("Fehler %v, erwartet ErrZuGross", fehler)
	}

	if belegt > 64<<20 {
		t.Errorf("%d MiB zugeteilt — die Datei wurde ganz gelesen", belegt>>20)
	}
}

func TestDasSchreibenInsArchivLiestKeineAufgeblaseneDatei(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bombe.mse-set")
	archivMitNullen(t, pfad, "set", msepaket.GrenzePaketdatei+1)

	var fehler error

	belegt := zugeteilt(func() {
		fehler = msepaket.DateiSchreiben(t.Context(), pfad, "set", []byte("neu"), "egal")
	})

	if !errors.Is(fehler, msepaket.ErrZuGross) {
		t.Errorf("Fehler %v, erwartet ErrZuGross", fehler)
	}

	if belegt > 64<<20 {
		t.Errorf("%d MiB zugeteilt — der Eintrag wurde ganz ausgepackt", belegt>>20)
	}
}

func TestEineUebergrosseSymbolzeichnungWirdAbgewiesen(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "gross.mse-symbol")
	duenneDatei(t, pfad)

	var fehler error

	belegt := zugeteilt(func() {
		_, _, fehler = msepaket.LiesSymbolzeichnung(pfad)
	})

	if !errors.Is(fehler, msepaket.ErrZuGross) {
		t.Errorf("Fehler %v, erwartet ErrZuGross", fehler)
	}

	if belegt > 64<<20 {
		t.Errorf("%d MiB zugeteilt", belegt>>20)
	}
}
