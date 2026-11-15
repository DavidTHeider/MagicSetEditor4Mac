package service

// **Die Ausstellseite der Nachweise** —
// #1539.
//
// # Der Befund
//
// **`Nachweise.Ausstellen` stand da, und niemand rief es.** `Autorisierung`
// hielt den Dienst, benutzte ihn aber nur zum **Einlösen** — der Kommentar an
// seinem Feld sagte es wörtlich: „nachweise löst Nachweise ein". Damit
// scheiterte am laufenden Fenster **jeder** kostenpflichtige KI-Auftrag an
// „braucht die ausdrückliche Freigabe eines Menschen", und es gab keinen Weg,
// sie zu geben.
//
// **Kein Lauf sah es**, weil jeder die Freigabe selbst setzte
// ([App.MitKIFreigabe]) — die Vorrichtung war großzügiger als der Betrieb
// (§3), und keiner fragte, ob es im Betrieb überhaupt einen Weg dorthin gibt.
//
// # Warum die Ausstellung über die Autorisierung geht
//
// **Weil sie den Akteur baut.** Wer ausstellt, muss am **Fenster** sitzen und
// mindestens mitarbeiten — das prüft [Nachweise.Ausstellen], und den Akteur
// dafür kennt nur die Autorisierung.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ausstellen gibt einen Nachweis für **einen** Aufruf.
//
// **Er gilt einem Aufruf und einer Frist.** Wer ihn zweimal einlösen will,
// braucht zwei — das hält [Nachweise.Einloesen], und hier steht nur, dass
// überhaupt einer entsteht.
func (a *Autorisierung) Ausstellen(
	ctx context.Context, antrag Nachweisantrag,
) (domain.Nachweis, error) {
	if a.nachweise == nil {
		return domain.Nachweis{}, fmt.Errorf(
			"%w: an dieser Fläche werden keine Nachweise ausgestellt",
			domain.ErrErlaubnis)
	}

	wer, fehler := a.Akteur(ctx)
	if fehler != nil {
		return domain.Nachweis{}, fehler
	}

	return a.nachweise.Ausstellen(ctx, wer, antrag)
}

// Grenze gibt die Kostengrenze eines Nachweises, **ohne ihn zu verbrauchen**.
//
// # Wofür es das gibt
//
// **Damit der Auftrag bucht, was ein Mensch freigegeben hat.** Die Buchung
// braucht einen Höchstbetrag ([domain.Betrag.Pruefe] verlangt eine Währung);
// bis zum 2026-09-12 kam er aus einem Feld, das nur ein Lauf setzte, und am
// Fenster stand dort nichts. Ein Auftrag mit leerem Höchstbetrag scheiterte
// an „Eine Währung hat drei Zeichen" — eine Meldung über eine Währung, wo
// eine Freigabe fehlte.
//
// **Gelesen und nicht mitgereicht**: gäbe der Aufrufer den Betrag an, könnte
// er einen kleineren nennen als den freigegebenen und die Tagesgrenze
// unterlaufen. Was gebucht wird, steht im Nachweis.
func (n *Nachweise) Grenze(
	ctx context.Context, wer domain.Akteur, kennung string,
) (domain.Kostengrenze, error) {
	nachweis, fehler := n.ablage.Nachweis(ctx, kennung)
	if fehler != nil {
		return domain.Kostengrenze{}, fehler
	}

	// **Dieselbe Zuordnung wie beim Einlösen** ([domain.Nachweis.Gilt]): wer
	// einen fremden Nachweis nicht einlösen kann, soll auch nicht lesen, auf
	// welchen Betrag er lautet.
	if nachweis.NutzerID != wer.NutzerID || nachweis.ProfilID != wer.ProfilID {
		return domain.Kostengrenze{}, fmt.Errorf(
			"%w: dieser Nachweis gehört jemand anderem", domain.ErrErlaubnis)
	}

	return nachweis.Grenze, nil
}

// Nachweisgrenze gibt die Grenze eines Nachweises dieses Akteurs.
func (a *Autorisierung) Nachweisgrenze(
	ctx context.Context, kennung string,
) (domain.Kostengrenze, error) {
	if a.nachweise == nil {
		return domain.Kostengrenze{}, fmt.Errorf(
			"%w: an dieser Fläche werden keine Nachweise geführt",
			domain.ErrErlaubnis)
	}

	wer, fehler := a.Akteur(ctx)
	if fehler != nil {
		return domain.Kostengrenze{}, fehler
	}

	return a.nachweise.Grenze(ctx, wer, kennung)
}
