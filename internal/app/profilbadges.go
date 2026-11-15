// **Das Abzeichen eines Profils** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 13 der Kommentierung vom 2026-09-05.
//
// # Warum es neben dem Profilbild steht und nicht darin
//
// **Weil es einer anderen Sache gehört.** Ein Profilbild zeigt den
// **Menschen**; ein Abzeichen zeigt die **Rolle**, in der er gerade arbeitet,
// und liegt deshalb rechts unten darüber. Wer beides in einer Bindung
// zusammenfasste, müsste bei jedem Aufruf sagen, welches von beiden gemeint
// ist — und das ist genau die Frage, die der Name schon beantwortet.
//
// # Und warum die Reihenfolge dieselbe ist
//
// **Weil die Überlegung dieselbe ist**: der Name trägt die Kennung, ein
// zweites Abzeichen überschreibt das erste, und ein abgewiesenes Schreiben
// nimmt die Datei nur zurück, wenn vorher keine dalag. Der Grund steht bei
// [App.NutzerbildSetzen] und wird hier nicht wiederholt.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ProfilbadgeWaehlen fragt nach einer Bilddatei und legt sie ab.
func (a *App) ProfilbadgeWaehlen(
	ctx context.Context, profilID, fassung string,
) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	// **Vor dem Dialog gefragt** (ID-03): wer erst nach einer Datei fragt
	// und dann absagt, hat jemanden umsonst suchen lassen.
	if befund := a.profilVorab(ctx, profilID); befund != nil {
		return EinProfilAntwort{Fehler: befund}
	}

	pfad, err := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  DialogtexteFuer(a.sprache).BildOeffnen,
		Filter: DialogtexteFuer(a.sprache).BildFilter,
		Muster: "*.png;*.jpg;*.jpeg;*.gif",
	})
	if err != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return EinProfilAntwort{}
	}

	return a.ProfilbadgeSetzen(ctx, profilID, fassung, pfad)
}

// ProfilbadgeSetzen legt eine Bilddatei als Abzeichen ab.
func (a *App) ProfilbadgeSetzen(
	ctx context.Context, profilID, fassung, pfad string,
) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	kennung, geleseneFassung, befund := a.nutzerbildziel(profilID, fassung)
	if befund != nil {
		return EinProfilAntwort{Fehler: befund}
	}

	if fremd := a.ueberProfil(ctx, kennung); fremd != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	bild, lesefehler := bilddateiLesen(pfad)
	if lesefehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, lesefehler)}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	vorher, zeilenfehler := a.profilzeile(ctx, kennung)
	if zeilenfehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, zeilenfehler)}
	}

	alt := vorher.Badge

	name, ablegefehler := ablage.ProfilbadgeAblegen(kennung, bild)
	if ablegefehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, ablegefehler)}
	}

	vorher.Badge = name
	vorher.Fassung = geleseneFassung

	// **Dieselbe Reihenfolge wie beim Profilbild** ([App.NutzerbildSetzen]):
	// das neue unter eigenem Namen, der Bestand wechselt, dann geht das alte
	// — und scheitert der Bestand, das neue.
	antwort := a.ProfilAendern(ctx, profilangabenAus(vorher))
	switch {
	case antwort.Fehler != nil && name != alt:
		bildWegraeumen(ctx, ablage, name)
	case antwort.Fehler == nil && alt != "" && alt != name:
		bildWegraeumen(ctx, ablage, alt)
	}

	return antwort
}

// ProfilbadgeEntfernen nimmt das Abzeichen eines Profils wieder weg.
//
// **Erst der Bestand, dann die Datei** — dieselbe Reihenfolge und derselbe
// Grund wie bei [App.NutzerbildEntfernen].
func (a *App) ProfilbadgeEntfernen(
	ctx context.Context, profilID, fassung string,
) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	kennung, geleseneFassung, befund := a.nutzerbildziel(profilID, fassung)
	if befund != nil {
		return EinProfilAntwort{Fehler: befund}
	}

	if fremd := a.ueberProfil(ctx, kennung); fremd != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	vorher, zeilenfehler := a.profilzeile(ctx, kennung)
	if zeilenfehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, zeilenfehler)}
	}

	alt := vorher.Badge
	vorher.Badge = ""
	vorher.Fassung = geleseneFassung

	antwort := a.ProfilAendern(ctx, profilangabenAus(vorher))
	if antwort.Fehler == nil && alt != "" {
		bildWegraeumen(ctx, ablage, alt)
	}

	return antwort
}

// profilzeile gibt ein Profil, wie der Bestand es führt.
//
// **Weil [storage.Nutzerablage.ProfilAendern] alle Spalten schreibt.** Wer
// nur das Abzeichen setzte und die übrigen Felder leer ließe, löschte
// Subslogan und beide Beschreibungen mit.
func (a *App) profilzeile(ctx context.Context, profilID int64) (domain.Profil, error) {
	zeilen, fehler := a.nutzer.Uebersicht(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	for _, zeile := range zeilen {
		for _, eines := range zeile.Profile {
			if eines.ID == profilID {
				return eines, nil
			}
		}
	}

	return domain.Profil{}, fmt.Errorf("%w: das Profil %d führt der Bestand nicht",
		domain.ErrNotFound, profilID)
}

// profilangabenAus baut die Angaben aus einem Profil des Bestands.
func profilangabenAus(eines domain.Profil) ProfilangabenDTO {
	return ProfilangabenDTO{
		ID:                zahltext(eines.ID),
		Name:              eines.Name,
		Subslogan:         eines.Subslogan,
		Kurzebeschreibung: eines.Kurzebeschreibung,
		Langebeschreibung: eines.Langebeschreibung,
		Badge:             eines.Badge,
		Fassung:           zahltext(eines.Fassung),
	}
}
