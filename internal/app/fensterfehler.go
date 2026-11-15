package app

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"unicode/utf8"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fensterfehler ist ein Fehler, den das Fenster beim Zeichnen gefangen hat —
// F01 (#2340),
// Befund A-01.
//
// # Warum er hierher gemeldet wird
//
// **Die Fehlerfläche im Fenster nennt eine Kennung**, und sie ist nur etwas
// wert, wenn sie sich wiederfinden lässt: im Protokoll, das das
// Diagnosepaket einsammelt (`protokoll.txt`). Ohne diese Zeile stünde im
// Fenster eine Zahl, zu der es nichts gibt.
type Fensterfehler struct {
	// Kennung ist die Zahl, die das Fenster zeigt — `7f3a-21c9`.
	Kennung string `json:"kennung"`

	// Ebene ist `bereich` (eine Seite) oder `wurzel` (das ganze Fenster).
	Ebene string `json:"ebene"`

	// Bereich ist der Schlüssel der Seite, `fenster` an der Wurzel.
	Bereich string `json:"bereich"`

	// Meldung ist der Satz des Fehlers, wie die Oberfläche ihn bekam.
	Meldung string `json:"meldung"`
}

// FensterfehlerHoechstens ist die längste Meldung, die ins Protokoll geht.
//
// **Wer das Protokoll füllen kann, verdrängt die Zeilen**, um derentwillen es
// da ist: ein Fehler, der in einer Schleife wirft, schriebe sonst jedes Mal
// seinen ganzen Stapel. Zweitausend Zeichen tragen einen Satz samt Ort.
const FensterfehlerHoechstens = 2000

// fensterebenen sind die Stellen, an denen das Fenster einen Fehler fängt —
// ein Bereich, die Wurzel, **und seit F03 (#2342)** die Panik einer Bindung
// (`ruf`, der Bereich ist die Methode) und ein Fehler ohne Schlüssel, den
// eine Meldung zeigt (`meldung`).
var fensterebenen = map[string]bool{"bereich": true, "wurzel": true, "ruf": true, "meldung": true}

var (
	fensterkennung = regexp.MustCompile(`^[0-9a-f]{4}-[0-9a-f]{4}$`)
	fensterbereich = regexp.MustCompile(`^[a-z0-9]{1,40}$`)
)

// FensterfehlerMelden schreibt einen Fensterfehler mit Kennung ins Protokoll.
//
// **Geprüft, bevor er geschrieben wird**: eine Kennung in fremder Gestalt —
// etwa mit einem Zeilenumbruch — könnte sonst eine eigene Protokollzeile
// vortäuschen, und eine unbekannte Ebene wäre eine Behauptung, die niemand
// einordnen kann. Abgewiesen wird mit [domain.ErrValidation]; ins Protokoll
// geht dann nichts.
func (a *App) FensterfehlerMelden(ctx context.Context, fehler Fensterfehler) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if !fensterkennung.MatchString(fehler.Kennung) || !fensterbereich.MatchString(fehler.Bereich) ||
		!fensterebenen[fehler.Ebene] {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: ein Fensterfehler braucht Kennung, Ebene und Bereich in ihrer Gestalt",
			domain.ErrValidation))}
	}

	meldung := fehler.Meldung
	if utf8.RuneCountInString(meldung) > FensterfehlerHoechstens {
		meldung = string([]rune(meldung)[:FensterfehlerHoechstens])
	}

	// **Angemeldete Merkmale** (`protokollmerkmale.go`): ein anderes verlöre
	// im Schwärzer seinen Wert, und die Kennung stünde nur im Fenster. Die
	// Ebene ist eine `art`, die Meldung ein `grund` aus dem Unterbau.
	slog.ErrorContext(ctx, "das Fenster fing einen Fehler",
		"fehlerkennung", fehler.Kennung, "art", fehler.Ebene, "bereich", fehler.Bereich,
		"grund", meldung)

	return SchreibAntwort{}
}
