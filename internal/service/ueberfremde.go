// **Wer über einen anderen Menschen schreibt** — ID-03
// (#1702).
//
// # Der Befund
//
// **Die Matrix sagt es seit 0078 für jede Fähigkeit am Träger Nutzer**:
// „wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren
// zweiten Faktor" (0074). Kein Dienst hielt es. `FaktorEntfernen` nahm jedem
// Nutzer seinen Faktor, gleich wer arbeitete — danach war der Wechsel zu ihm
// frei, und der Faktor schützte vor niemandem, der den Knopf kannte.
//
// # Die Regel
//
// **Über einen anderen Menschen schreibt, wer ihn werden kann, ohne dass
// dessen Faktor fragt.** Wer das nicht kann, wechselt zuerst zu ihm, und dort
// fragt der Faktor — so verlangt die Anwendung ihn, ohne dass jede Bindung
// einen Nachweis mitführen müsste.
//
// **Für einen Menschen ohne Faktor ändert sich nichts**: ihn schützt nichts
// (0074, „Was das nicht löst"), und wer ihn verwalten will, könnte ihn
// ebenso gut werden. Eine Sperre, die man mit einem Umweg nimmt, wäre eine
// Behauptung.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// UeberNutzer sagt, ob der arbeitende Mensch über `ziel` schreiben darf.
func (v *Nutzerverwaltung) UeberNutzer(ctx context.Context, ziel int64) error {
	arbeitet, fehler := v.WerArbeitet(ctx)
	if fehler != nil {
		return fehler
	}

	if arbeitet == ziel || v.faktor == nil {
		return nil
	}

	verlangt, auskunftsfehler := v.faktor.VerlangtBeimWechsel(ctx, ziel)
	if auskunftsfehler != nil {
		return auskunftsfehler
	}

	if !verlangt {
		return nil
	}

	return fmt.Errorf("%w — wechsle zuerst zu diesem Nutzer, dabei fragt die "+
		"Anwendung nach dem zweiten Faktor", domain.ErrFremderNutzer)
}

// UeberProfil sagt, ob der arbeitende Mensch über dieses Profil schreiben
// darf — seine Angaben ändern oder es löschen.
//
// # Zwei Fragen
//
// **Die nach dem Menschen, dem es gehört** ([Nutzerverwaltung.UeberNutzer]):
// ein Profil ist ein Stück seines Eigners.
//
// **Und im gezeigten Profil die nach der Rolle** — der native Befund F01.
// Wer ohne Rolle in ein Profil kommt — ein Eigner ohne Faktor lässt jeden
// herein —, sieht „nur lesen", und ohne diese Frage ließen sich die Angaben
// des Profils trotzdem ändern. **Nur im gezeigten**: von außen, aus
// dem eigenen Profil heraus, gilt allein die erste Frage (0074).
func (v *Nutzerverwaltung) UeberProfil(ctx context.Context, profilID int64) error {
	profil, fehler := v.profilSuchen(ctx, profilID)
	if fehler != nil {
		return fehler
	}

	if fremd := v.UeberNutzer(ctx, profil.NutzerID); fremd != nil {
		return fremd
	}

	gezeigt, lesefehler := v.Gezeigt(ctx)
	if lesefehler != nil {
		return lesefehler
	}

	if gezeigt.ID != profilID {
		return nil
	}

	rolle, rollenfehler := v.RolleImGezeigten(ctx)
	if rollenfehler != nil {
		return rollenfehler
	}

	if (domain.Akteur{Rolle: rolle}).Darf(domain.WirkungSchreiben) != nil {
		return fmt.Errorf("%w: in diesem Profil wird nur gelesen — auch seine Angaben",
			domain.ErrNurLesend)
	}

	return nil
}
