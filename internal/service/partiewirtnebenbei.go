package service

// **Was am Tisch gesagt und gezeigt wird** — Chat und Präsenz neben dem
// Strom (G5, #1462).
//
// **Aus `partiewirt.go` herausgelöst**, als sie mit P9 an die Zeilengrenze
// aus §2 stieß. Der Schnitt folgt der Frage: dort steht, was der Wirt mit
// einem Befehl tut, hier, was er mit dem tut, was **kein** Befehl ist.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nebenstrom bekommt, was kein Ereignis ist — Chat und Präsenz.
//
// **Getrennt vom Ereignisweg**, weil sie nichts am Stand ändern: sie gehen
// ans Fenster und nicht in den Strom. Wer sie durch dieselbe Tür schickte,
// müsste bei jedem Chat entscheiden, ob er die Partie fortschreibt.
//
// **Und er bekommt den [Nebenbefund] und nicht die rohe Nachricht.** Wer
// **geredet** hat, steht in der Nutzlast und nicht im Umschlag: dort steht
// der Absender **dieses Sprungs**, also der Wirt, wenn er Annas Satz an Bert
// weiterreicht. Ein Empfänger, der die Nachricht bekäme, müsste die Nutzlast
// ein zweites Mal auspacken — und bekäme dabei einen Fehlerzweig, den kein
// Lauf betreten kann (§10).
type Nebenstrom interface {
	Nebenbei(ctx context.Context, befund Nebenbefund) error
}

// MitNebenstrom setzt, wohin Chat und Präsenz gehen.
func (w *Partiewirt) MitNebenstrom(nebenbei Nebenstrom) *Partiewirt {
	w.nebenbei = nebenbei

	return w
}

// weiterreichen beglaubigt Chat und Präsenz und bringt sie an ihre zwei Orte.
//
// # Die zwei Orte, und warum es zwei sind
//
// **Das Fenster dieses Rechners und der Rest des Tisches.** Der Wirt spielt
// selbst mit; was ein Gast sagt, gehört ihm **und** jedem anderen Gast. Ohne
// den zweiten Weg hätte eine Partie mit drei Sitzen drei Zwiegespräche mit dem
// Wirt und kein gemeinsames.
//
// # Warum der Sprecher aus der Sitzung kommt
//
// **Weil `Von` in der Nutzlast eine Behauptung des Gastes ist** — dieselbe
// Anmaßung wie ein fremder `Spielbefehl.Absender`, und derselbe Schaden: ein
// Gast redete im Namen eines anderen und bewegte dessen Zeiger.
//
// # Und warum ein überholter Zeiger still verschwindet
//
// **Weil er der Normalfall einer verwerfbaren Zustellung ist.** Ein Fehler
// verlangte eine Reaktion; hier ist die richtige Reaktion, nichts zu tun
// ([Nebenbefund.Gilt]).
//
// **Ohne Zuhörer und ohne Tisch geschieht nichts, und auch das ist kein
// Fehler.** Ein Wirt ohne Fenster — an der Kommandozeile etwa — führt die
// Partie trotzdem.
func (w *Partiewirt) weiterreichen(
	ctx context.Context, nachricht domain.Partienachricht,
) error {
	befund, fehler := w.pruefer.Beglaubige(nachricht, w.gast)
	if fehler != nil {
		return fehler
	}

	if !befund.Gilt {
		return nil
	}

	if w.nebenbei != nil {
		if fehler := w.nebenbei.Nebenbei(ctx, befund); fehler != nil {
			return fehler
		}
	}

	if w.runde == nil {
		return nil
	}

	return w.runde.Reiche(ctx, befund)
}
