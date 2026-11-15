package domain

import (
	"errors"
	"fmt"
	"log"
)

// ErrZusammengebrochen sagt, dass ein Nebenlauf mit einer Panik endete — BK-05
// (#2111).
//
// **Eine Panik in einer eigenen Goroutine fängt niemand**: nicht Wails, nicht
// `net/http` — der Prozess endet, und mit ihm jeder Tisch, jeder Dienst und
// jede offene Bearbeitung. Wer einen Nebenlauf startet, fängt sie deshalb
// selbst und macht daraus diesen Fehler.
var ErrZusammengebrochen = errors.New("ein Nebenlauf ist zusammengebrochen")

// Auffangen macht aus einer Panik einen Fehler und reicht ihn an `melde`.
//
// **Nur als `defer domain.Auffangen(…)` wirksam**: `recover` greift allein
// in der Funktion, die aufgeschoben wurde — eine Hülle darum fände nichts.
//
// **Ein geworfener Fehler bleibt erkennbar**: `errors.Is` geht durch beide,
// den Zusammenbruch und das, was geworfen wurde.
//
// **Und ein Zusammenbruch beim Melden bleibt hier** (Zweitblick über A und B,
// Befund 1): wer in `melde` aufräumt und dabei selbst zusammenbricht, nähme
// sonst doch noch den Prozess mit. Er steht dann auf dem Standardfehler — dem
// einzigen Weg, der sicher bleibt, wenn die Meldung selbst nicht ging.
func Auffangen(melde func(error)) {
	grund := recover()
	if grund == nil {
		return
	}

	zusammenbruch := fmt.Errorf("%w: %v", ErrZusammengebrochen, grund)
	if geworfen, ist := grund.(error); ist {
		zusammenbruch = fmt.Errorf("%w: %w", ErrZusammengebrochen, geworfen)
	}

	defer func() {
		if beimMelden := recover(); beimMelden != nil {
			log.Printf("%v — und beim Melden: %v", zusammenbruch, beimMelden)
		}
	}()

	melde(zusammenbruch)
}
