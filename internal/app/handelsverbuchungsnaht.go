package app

// **Die Naht der Handelsverbuchung** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// **Weil nur hier `internal/storage` stehen darf** (§2) — derselbe Grund wie
// bei [tauschnaht] und [handelsausgang]. Nebenan steht, **was** eine
// Verbuchung ist; hier steht, **woraus** sie gebaut wird.
//
// # Warum der Bestand aus drei Ablagen zusammengesetzt wird
//
// **Weil die Verbuchung drei Dinge braucht und keine davon die anderen
// kennt**: eine Besitzzeile, die Zusagen eines Vorgangs und den Abschluss.
// Sie liegen in derselben Datenbank und in derselben Transaktion — aber sie
// zu **einer** Ablage zu machen, hieße, drei Fragen zu einer zu erklären.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// verbuchungsbestand hält die drei Ablagen zusammen.
//
// **Eingebettet und nicht durchgereicht**: Go befördert die Verfahren, und
// jede Weiterleitung von Hand wäre eine Stelle, an der eine fehlen kann.
type verbuchungsbestand struct {
	*storage.Sammlungsspeicher
	*storage.Reservierungsablage
	*storage.Abschlussablage
}

// handelsablage baut die Ablage der Handelsstände (T10, #1475).
//
// **Hier und nicht in `handelsfreigabe.go`**: nur die Nähte dürfen
// `internal/storage` kennen (§2), und eine zweite Datei dieser Art wäre ein
// zweiter Composition Root für dieselbe Fähigkeit.
func (a *App) handelsablage() *storage.Handelsablage {
	return storage.NeueHandelsablage(a.laufendeVerbindung())
}

// abschlussablage baut die Ablage für Journal und Beleg.
func (a *App) abschlussablage() *storage.Abschlussablage {
	return storage.NeueAbschlussablage(a.laufendeVerbindung())
}

// verbuchungsstelle baut die Verbuchung für diesen Ruf.
func (a *App) verbuchungsstelle() *service.Verbuchungsstelle {
	db := a.laufendeVerbindung()

	// **Mit dem Riegel vor den verbindlichen Schritten** (T10,
	// #1475):
	// dieselbe [service.Autorisierung], die auch die Kosten hütet — und
	// damit dieselbe Reichweite für Token und dieselben Nachweise.
	return service.NeueVerbuchungsstelle(
		storage.NeueHandelsablage(db),
		verbuchungsbestand{
			Sammlungsspeicher:   storage.NeuerSammlungsspeicher(db),
			Reservierungsablage: storage.NeueReservierungsablage(db),
			Abschlussablage:     storage.NeueAbschlussablage(db),
		},
		a.peerwerk(),
		storage.NeueDrucke(db),
	).MitWache(AutorisierungAus(db))
}
