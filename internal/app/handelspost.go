package app

// **Was zu einem Gegenüber noch aussteht** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Warum hier Zahlen stehen und keine Nachrichten
//
// **Weil eine Diagnose sagen soll, ob etwas klemmt, und nicht, worum es
// geht.** Das Issue verlangt „technische Zustände und Zähler, ohne
// Rohsammlungen/Decklisten oder Geheimnisse"; wer den Inhalt zeigte, hätte
// eine zweite Sicht auf den Handel gebaut — und die gibt es schon.

import (
	"context"
)

// PoststandDTO sagt, was zu einem Gegenüber noch aussteht.
type PoststandDTO struct {
	// Kennung ist die Peerkennung des Gegenübers.
	Kennung string `json:"kennung"`

	// Offen ist die Zahl der Nachrichten, die noch auf ihre Bestätigung
	// warten.
	Offen int `json:"offen"`

	// HoechsteVersuche ist, wie oft die hartnäckigste schon hinausging —
	// **die höchste und nicht die Summe**.
	HoechsteVersuche int `json:"hoechsteVersuche"`
}

// PoststandAntwort trägt den Stand samt seinem Fehler.
type PoststandAntwort struct {
	Stand  PoststandDTO `json:"stand"`
	Fehler *Fehler      `json:"fehler"`
}

// Handelspoststand sagt, was zu einem Gegenüber noch aussteht.
func (a *App) Handelspoststand(
	ctx context.Context, kennung string,
) PoststandAntwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.handelsausgang().Stand(ctx, kennung)
	if fehler != nil {
		return PoststandAntwort{
			Stand: PoststandDTO{Kennung: kennung}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return PoststandAntwort{Stand: PoststandDTO{
		Kennung: kennung, Offen: stand.Offen,
		HoechsteVersuche: stand.HoechsteVersuche,
	}}
}
