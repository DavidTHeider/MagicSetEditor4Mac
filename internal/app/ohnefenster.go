package app

// **Die Nähte, solange kein Fenster da ist** — ENG-02
// (#1727).
//
// # Warum es sie gibt
//
// Unter Wails 2 setzte `NeueApp` die echten Nähte selbst — Menüleiste,
// Meldung, Dateidialog, Zwischenablage —, und sie standen als `…Echt` in
// dieser Schicht. Seit Wails 3 kennt `internal/app` die Hülle nicht mehr:
// `internal/fenster` setzt die Nähte, bevor das Fenster aufgeht.
//
// **Bis dahin gelten diese** — und in jedem Lauf, der keine eigene setzt.
// Sie melden nichts und setzen keine Leiste; wer einen Dialog oder die
// Zwischenablage braucht, bekommt einen Fehler mit Namen statt eines
// Nullzeigers. Ein Nullzeiger wäre der Absturz, den Wails 2 mit seinem
// `nil`-Kontext hatte (`menuefenster.go`, vor ENG-02).

import (
	"context"
	"errors"
)

// ErrOhneFenster heißt: dafür braucht es das Fenster, und es steht keines.
//
// **Kein Fehler des Aufrufers** — eine Anwendung ohne Hülle ist ein
// Verdrahtungsfehler, und der gehört ins Protokoll (`CodeUnbekannt`), nicht an
// ein Feld.
var ErrOhneFenster = errors.New("dafür braucht es das Fenster, und es steht keines")

func menueOhneFenster(context.Context, *Menue) {}

func meldenOhneFenster(context.Context, string, ...any) {}

func dateiwahlOhneFenster(context.Context, Dateiauswahl) (string, error) {
	return "", ErrOhneFenster
}

func ablageLesenOhneFenster(context.Context) (string, error) { return "", ErrOhneFenster }

func ablageSchreibenOhneFenster(context.Context, string) error { return ErrOhneFenster }
