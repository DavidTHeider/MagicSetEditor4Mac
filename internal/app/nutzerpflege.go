// **Nutzer und Profile ändern und löschen, am Fenster** — Werkstück 5 von
// Phase 4.12
// (#1351).
//
// # Der Name bestätigt das Löschen
//
// **Dieselbe Form wie beim Set** (#1218): eine Kennung allein ist eine Zahl,
// und wer sich vertippt, löscht den falschen Menschen. Der Name kommt aus der
// Liste, die vor dem Löschen dastand — er ist keine zweite Eingabe, sondern
// eine Bestätigung dessen, was gemeint war.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// NutzerAendern schreibt Name und Bild eines Nutzers fort.
func (a *App) NutzerAendern(
	ctx context.Context, angaben NutzerangabenDTO,
) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", angaben.ID)
	if befund != nil {
		return EinNutzerAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	// **Die Fassung reist als Text, wie jede Kennung** — eine
	// JavaScript-Zahl verlöre oberhalb von 2⁵³ die letzten Stellen. Dass sie
	// das hier nie täte, ist kein Grund für eine zweite Bauform.
	fassung, fassungsbefund := benannteKennungAus("fassung", angaben.Fassung)
	if fassungsbefund != nil {
		return EinNutzerAntwort{Fehler: fassungsbefund}
	}

	geaendert, fehler := a.nutzer.NutzerAendern(ctx, domain.Nutzer{
		ID:      kennung,
		Name:    angaben.Name,
		Bild:    angaben.Bild,
		Fassung: fassung,

		Anrede:       angaben.Anrede,
		Namenszusatz: angaben.Namenszusatz,
		Vorname:      angaben.Vorname,
		Weiterenamen: angaben.Weiterenamen,
		Nachname:     angaben.Nachname,
		Slogan:       angaben.Slogan,
		Kurzebio:     angaben.Kurzebio,
		Langebio:     angaben.Langebio,
	})
	if fehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Mit seinen Profilen**, wie überall sonst: eine leere Liste behauptete,
	// er habe keine (§7, Regel 2).
	return a.nutzerNach(ctx, geaendert.ID)
}

// ProfilAendern schreibt die Angaben eines Profils fort.
//
// **Ein Verbund seit Werkstück 9** (#1351) — vorher waren es Kennung, Name
// und Fassung. **Erweiternd und nicht brechend**: die vier neuen Felder
// entstehen mit dieser Wanderung, und wer sie wegließe, könnte nichts
// verlieren, was es vorher gab. Ab jetzt gilt, was bei [App.NutzerAendern]
// steht: es **setzt** und verschmilzt nicht.
func (a *App) ProfilAendern(
	ctx context.Context, angaben ProfilangabenDTO,
) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("profilID", angaben.ID)
	if befund != nil {
		return EinProfilAntwort{Fehler: befund}
	}

	if fremd := a.ueberProfil(ctx, kennung); fremd != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	geleseneFassung, fassungsbefund := benannteKennungAus("fassung", angaben.Fassung)
	if fassungsbefund != nil {
		return EinProfilAntwort{Fehler: fassungsbefund}
	}

	// **Die Nutzerkennung kommt aus dem Bestand und nicht von der Fläche.**
	// Ein Profil wechselt seinen Nutzer nicht; wer sie mitschickte, könnte es.
	eigner, eignerfehler := a.eignerVon(ctx, kennung)
	if eignerfehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, eignerfehler)}
	}

	geaendert, fehler := a.nutzer.ProfilAendern(ctx, domain.Profil{
		ID: kennung, NutzerID: eigner, Name: angaben.Name,
		Fassung: geleseneFassung,

		Subslogan:         angaben.Subslogan,
		Kurzebeschreibung: angaben.Kurzebeschreibung,
		Langebeschreibung: angaben.Langebeschreibung,
		Badge:             angaben.Badge,
		Technisch:         angaben.Technisch,
		Ablauf:            angaben.Ablauf,
	})
	if fehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	gelesen := profilDTO(geaendert)

	return EinProfilAntwort{Profil: &gelesen}
}

// NutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus.
func (a *App) NutzerLoeschen(ctx context.Context, nutzerID, name string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	if fehler := a.nutzer.NutzerLoeschen(ctx, kennung, name); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// ProfilLoeschen nimmt ein Profil heraus.
func (a *App) ProfilLoeschen(ctx context.Context, profilID, name string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("profilID", profilID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if fremd := a.ueberProfil(ctx, kennung); fremd != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	if fehler := a.nutzer.ProfilLoeschen(ctx, kennung, name); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Der Zeiger kann gewandert sein** — der Fachkern stellt ihn um, wenn
	// das gelöschte Profil das gezeigte war. Das Fenster zieht nach, sonst
	// arbeitete es weiter in einem Profil, das es nicht mehr gibt.
	//
	// **Und die Leserverbindung zieht mit** — Nachlauf zum Zweitblick über
	// die Phasen 4.0 bis 4.13 (2026-09-07), Befund 1. Hier stand
	// `profilLaden`, das **nur** die Kennung nachzog: nach dem Löschen des
	// aktiven Profils gehörte die Verbindung noch zur bisherigen Rolle. Wer
	// als Leser in einem fremden Profil stand und es löschte, landete im
	// Ausweichprofil — mit einem Bestand, der weiter nichts annahm.
	a.profilNachziehen(ctx)

	return SchreibAntwort{}
}

// eignerVon gibt den Nutzer, dem dieses Profil gehört.
func (a *App) eignerVon(ctx context.Context, profilID int64) (int64, error) {
	zeilen, fehler := a.nutzer.Uebersicht(ctx)
	if fehler != nil {
		return 0, fehler
	}

	for _, zeile := range zeilen {
		for _, profil := range zeile.Profile {
			if profil.ID == profilID {
				return profil.NutzerID, nil
			}
		}
	}

	return 0, fmt.Errorf("%w: das Profil %d führt der Bestand nicht",
		domain.ErrNotFound, profilID)
}
