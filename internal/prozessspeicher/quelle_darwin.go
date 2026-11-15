//go:build darwin

package prozessspeicher

// **Die vier Rufe ans System** — über `purego` und ohne cgo, wie
// `internal/biometrie`. Drei sind öffentlich (`libproc.h`): `proc_listallpids`,
// `proc_pidpath`, `proc_pid_rusage`; einer nicht:
// `responsibility_get_pid_responsible_for_pid`.
//
// # Ein Riegel um jeden Ruf
//
// **Fehlt ein Symbol, bleibt seine Funktion leer**, und ihr Ruf geriete in
// eine Panik — [riegel] macht daraus „nicht gemessen". So gibt es keinen
// Zweig, den nur ein anderes System beträte.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Die Konstanten aus dem SDK — `sys/resource.h` und `sys/proc_info.h`.
const (
	rusageInfoV4       = 4
	fussabdruckVersatz = 72 // ri_uuid[16] und sieben uint64 davor
	pfadGrenze         = 4096
	prozessgrenze      = 1 << 14
)

var (
	listallpids        func(puffer unsafe.Pointer, groesse int32) int32
	pidpath            func(pid int32, puffer unsafe.Pointer, groesse uint32) int32
	pidrusage          func(pid int32, art int32, puffer unsafe.Pointer) int32
	verantwortlichFuer func(pid int32) int32
)

func init() {
	_ = binde(&listallpids, "proc_listallpids")                                  //nolint:errcheck // ein fehlendes Symbol fängt der Riegel
	_ = binde(&pidpath, "proc_pidpath")                                          //nolint:errcheck // dito
	_ = binde(&pidrusage, "proc_pid_rusage")                                     //nolint:errcheck // dito
	_ = binde(&verantwortlichFuer, "responsibility_get_pid_responsible_for_pid") //nolint:errcheck // dito
}

// binde setzt `ziel` auf das Symbol `name` aus den Bibliotheken des Prozesses.
func binde(ziel any, name string) error {
	stelle, fehler := purego.Dlsym(purego.RTLD_DEFAULT, name)
	if fehler != nil {
		return fmt.Errorf("%s fehlt: %w", name, fehler)
	}

	purego.RegisterFunc(ziel, stelle)

	return nil
}

// geriegelt ruft `ruf` und macht aus einer Panik darin „nicht gemessen" —
// **an einer Stelle**, denn `recover` greift nur im eigenen `defer`.
func geriegelt[T any](ruf func() T) (wert T, fehler error) {
	defer func() {
		if gefangen := recover(); gefangen != nil {
			fehler = fmt.Errorf("%w: %v", errNichtGemessen, gefangen)
		}
	}()

	return ruf(), nil
}

func echteQuelle() quelle {
	return quelle{
		pids: alleProzesse, pfad: programmpfad, verantwortlich: verantwortlicherProzess, fussabdruck: fussabdruckVon,
		lebt: prozessLebt,
	}
}

func alleProzesse() ([]int, error) {
	puffer := make([]int32, prozessgrenze)

	// **Eine Panik gibt null Prozesse und ihren Fehler** — die Liste bleibt
	// dann leer.
	anzahl, fehler := geriegelt(func() int32 {
		// #nosec G103 G115 -- der Puffer lebt über den Ruf, und 4 · 2^14 passt in int32
		return listallpids(unsafe.Pointer(&puffer[0]), int32(len(puffer)*4))
	})

	pids := make([]int, 0, len(puffer))
	for _, pid := range puffer[:max(0, min(int(anzahl), len(puffer)))] {
		pids = append(pids, int(pid))
	}

	return pids, fehler
}

func programmpfad(pid int) (string, error) {
	puffer := make([]byte, pfadGrenze)

	laenge, fehler := geriegelt(func() int32 {
		// #nosec G103 G115 -- der Puffer lebt über den Ruf, und eine Prozessnummer ist ein pid_t
		return pidpath(int32(pid), unsafe.Pointer(&puffer[0]), pfadGrenze)
	})
	// **Eine Panik gibt die Länge null** und damit diesen Zweig.
	if laenge <= 0 {
		return "", errors.Join(fehler, fmt.Errorf("kein Programm zu Prozess %d", pid))
	}

	return string(puffer[:laenge]), nil
}

func verantwortlicherProzess(pid int) (int, error) {
	wer, fehler := geriegelt(func() int32 {
		return verantwortlichFuer(int32(pid)) // #nosec G115 -- eine Prozessnummer ist ein pid_t
	})

	return int(wer), fehler
}

func fussabdruckVon(pid int) (uint64, error) {
	var info [512]byte

	ergebnis, fehler := geriegelt(func() int32 {
		// #nosec G103 G115 -- der Puffer lebt über den Ruf, und eine Prozessnummer ist ein pid_t
		return pidrusage(int32(pid), rusageInfoV4, unsafe.Pointer(&info[0]))
	})
	if ergebnis != 0 || fehler != nil {
		return 0, errors.Join(fehler, fmt.Errorf("kein Speicher zu Prozess %d", pid))
	}

	return binary.LittleEndian.Uint64(info[fussabdruckVersatz:]), nil
}

// prozessLebt fragt mit dem Signal null — **auch ein fremder lebt**: `EPERM`
// heißt, es gibt ihn, nur darf dieser Prozess ihm nichts schicken.
func prozessLebt(pid int) bool {
	fehler := syscall.Kill(pid, 0)

	return fehler == nil || errors.Is(fehler, syscall.EPERM)
}
