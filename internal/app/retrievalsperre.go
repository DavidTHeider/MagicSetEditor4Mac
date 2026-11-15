package app

// **Die Absprache über Prozessgrenzen beim Indexbau** — Werkstück A15 von
// Phase 5.3
// (#1514).
//
// # Warum eine Datei und nicht ein Eintrag im Bestand
//
// **Weil eine Zeile in der Datenbank nach einem Absturz stehen bleibt.** Ein
// Prozess, den jemand abschießt, räumt sie nicht weg, und der nächste Aufbau
// fände einen Bau, den es nicht mehr gibt — dann bräuchte es eine zweite
// Regel darüber, wann ein Eintrag als verwaist gilt, und die ist immer
// entweder zu kurz oder zu lang.
//
// **`flock` fällt mit dem Prozess.** Das Betriebssystem gibt die Sperre frei,
// wenn der Griff geschlossen wird — auch wenn ihn niemand mehr schließt. Das
// ist die Crashstrategie, und sie besteht darin, keine zu brauchen.
//
// # Und warum neben dem Bestand
//
// **Weil die Sperre dem Index gilt und der Index im Bestand liegt.** Zwei
// Bestände sind zwei Indizes; eine Sperre an einem festen Ort im Zuhause
// hielte den zweiten Bau auf, der einen ganz anderen Index meint.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Indexbauendung hängt an den Pfad des Bestands.
const Indexbauendung = ".indexbau"

// indexsperre ist die Absprache über `flock`.
type indexsperre struct{ pfad func() string }

// Nehmen versucht die alleinige Sperre, ohne zu warten.
//
// **Ohne Pfad gibt es keine Absprache** — dann gilt nur der Riegel im
// Prozess, und das ist der Stand vor dieser Datei. Ein Bestand im
// Arbeitsspeicher hat keinen Ort, an dem eine Sperrdatei liegen könnte.
func (i indexsperre) Nehmen() (freigeben func(), genommen bool, fehler error) {
	ort := i.pfad()
	if ort == "" {
		return func() {}, true, nil
	}

	griff, fehler := dateisperre.Oeffnen(ort + Indexbauendung)
	if fehler != nil {
		return nil, false, fehler
	}

	// **Ein belegter Griff ist keine Störung.** Er sagt, dass ein anderer
	// Prozess baut — und der Griff wird geschlossen, sonst hielte dieser
	// Prozess eine offene Datei für nichts.
	//
	// **Der Fehler von `Allein` wird bewusst verworfen**, und der Linter
	// stimmt zu Recht ein: `flock` gibt hier `EWOULDBLOCK`, und das ist die
	// **Antwort** und kein Fehlschlag. Ihn weiterzureichen hieße, „ein
	// anderer baut" als Störung auszugeben.
	//
	// **Und ein Erbe baut nicht** (#2401): ein Kind, das dieser Prozess gerade
	// startet, hält die Sperre bis zu seinem `exec`.
	if belegt := griff.AlleinOhneErbe(context.Background()); belegt != nil {
		_ = griff.Schliessen() //nolint:errcheck // der Griff ist ohnehin unbrauchbar

		return nil, false, nil //nolint:nilerr // siehe oben
	}

	return func() { _ = griff.Schliessen() }, true, nil //nolint:errcheck // geschlossen ist geschlossen
}

// Indexsperre gibt die Absprache für einen Bestand an diesem Ort.
//
// **Der Pfad kommt als Funktion**, weil ein Zurückspielen den Bestand
// austauscht — dieselbe Entscheidung wie bei der Verbindung daneben.
func Indexsperre(pfad func() string) service.Prozesssperre {
	return indexsperre{pfad: pfad}
}
