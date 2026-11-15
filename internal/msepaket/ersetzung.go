package msepaket

// AutoErsetzung ist eine Regel aus einem `auto replace:`-Block.
//
// # Feste Zeichenketten, keine Muster
//
// In `magic.mse-game/auto_replace` stehen rund 500 Regeln, darunter
// `match: (C)` → `©` und `match: **` → `AUTOREM>`. Als reguläre Ausdrücke wäre
// das erste eine Gruppe um ein `C` und das zweite gar nicht übersetzbar. In 1541
// Zeilen kommt **kein** Rückstrich vor und kein `[^`, `.*` oder `\d`.
//
// Das ist die Antwort auf die Frage, ob eine fremde Vorlage die Anwendung mit
// einem bösartigen Ausdruck anhalten kann: **sie kann es nicht, weil es keinen
// Ausdruck gibt.** Eine Zeitschranke um eine Mustererkennung, die nicht
// stattfindet, wäre eine Vorkehrung gegen eine erfundene Gefahr.
type AutoErsetzung struct {
	Muster     string
	Ersatz     string
	GanzesWort bool
}

// **Die Maschine darüber steht in `internal/service`** —
// #897.
//
// **Wann eine Regel greift, ist keine Frage des Formats** (§2), sondern eine
// fachliche. Hier bleibt, was eine Regel **ist** und wie sie aus einem Paket
// kommt (`autoersetzungen.go`); `service.Autoersetzer` sagt, wann sie greift.
