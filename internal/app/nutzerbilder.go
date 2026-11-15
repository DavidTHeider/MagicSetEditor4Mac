// **Das Profilbild am Fenster und an den drei Flächen** — Werkstück 8 von
// Phase 4.12
// (#1351),
// Punkt 17 der Kommentierung vom 2026-09-05.
//
// # Warum es kein Zuschnitt ist
//
// **Ein Kartenbild wird auf eine Fläche der Vorlage zugeschnitten**
// ([App.BildSetzen] fragt den Plan nach `left:`, `width:` und `height:`) —
// ein Profilbild hat keine Vorlage. Es ist ein Abzeichen neben einem Namen,
// und was es braucht, ist eine **Schranke**, keine Fläche: das
// Seitenverhältnis bleibt, und [service.Nutzerbildkante] sagt, wie groß es
// höchstens wird.
//
// # Und warum das Anzeigen nur am Fenster steht
//
// **Dieselbe Begründung wie bei [App.Kartenbilder]**: ein Daten-URI ist kein
// Weg zu einer Auskunft, sondern die **Darstellung** einer Liste. An der
// Kommandozeile wäre er eine Zeichenkette, die niemand ansehen kann.
// **Setzen und Entfernen** stehen dagegen an allen vier Flächen — sie ändern
// den Bestand.

package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Nutzerbildantwort trägt die Bilder als Daten-URI, je Nutzerkennung.
//
// **Eine Tabelle und keine Liste** — dieselbe Bauform wie
// [Kartenbildantwort]: wer kein Bild hat, steht gar nicht darin, und das ist
// der Normalfall.
type Nutzerbildantwort struct {
	Bilder map[string]string `json:"bilder"`

	// Badges sind die Abzeichen, je **Profil**kennung — Werkstück 9 (#1351).
	//
	// **In derselben Antwort und nicht in einer zweiten** (§7, Regel 2): sie
	// werden nebeneinander gezeigt, ein Abzeichen **über** dem Bild. Zwei
	// Rufe wären zwei Stände, und im Zwischenraum läge ein Abzeichen über
	// einem Gesicht, zu dem es nicht gehört.
	Badges map[string]string `json:"badges"`

	Fehler *Fehler `json:"fehler"`
}

// Nutzerbilder gibt zu jedem Nutzer mit Bild einen Daten-URI.
//
// **Ein Bild, das nicht zu holen ist, fehlt einfach** — dieselbe Zusage wie
// bei [App.Kartenbilder]. Eine Liste, die wegen einer beschädigten Datei gar
// nichts zeigte, wäre der schlechtere Tausch; dass die Datei kaputt ist,
// sagt der Dienst dort, wo es etwas bedeutet.
func (a *App) Nutzerbilder(ctx context.Context) Nutzerbildantwort {
	ctx = a.imProfil(ctx)

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return Nutzerbildantwort{
			Bilder: map[string]string{}, Badges: map[string]string{},
			Fehler: Uebersetze(ctx, ortfehler),
		}
	}

	zeilen, fehler := a.nutzer.Uebersicht(ctx)
	if fehler != nil {
		return Nutzerbildantwort{
			Bilder: map[string]string{}, Badges: map[string]string{},
			Fehler: Uebersetze(ctx, fehler),
		}
	}

	// **Nie `nil`, sondern eine leere Tabelle** (§3): `omitempty` ließe sie
	// sonst aus der Antwort verschwinden, und die Seite bliebe leer.
	bilder := make(map[string]string, len(zeilen))
	badges := map[string]string{}

	for _, zeile := range zeilen {
		if uri, da := a.datenURIVon(ablage, zeile.Nutzer.Bild); da {
			bilder[zahltext(zeile.Nutzer.ID)] = uri
		}

		for _, eines := range zeile.Profile {
			if uri, da := a.datenURIVon(ablage, eines.Badge); da {
				badges[zahltext(eines.ID)] = uri
			}
		}
	}

	return Nutzerbildantwort{Bilder: bilder, Badges: badges}
}

// datenURIVon liest ein Bild und packt es — `false`, wenn es keines gibt.
//
// **Ein Bild, das nicht zu holen ist, fehlt einfach** — dieselbe Zusage wie
// bei [App.Kartenbilder]. Eine Liste, die wegen einer beschädigten Datei gar
// nichts zeigte, wäre der schlechtere Tausch.
func (a *App) datenURIVon(ablage *service.Bildablage, name string) (string, bool) {
	bild, lesefehler := ablage.NutzerbildLesen(name)
	if lesefehler != nil || bild == nil {
		return "", false
	}

	uri, kodierfehler := a.alsDatenURI(bild)
	if kodierfehler != nil {
		return "", false
	}

	return uri, true
}

// NutzerbildSetzen legt das Bild eines Nutzers ab und trägt es ein.
//
// # Die Reihenfolge ist eine Entscheidung
//
// **Erst die neue Datei, dann der Bestand, dann weg mit der alten** —
// Befund 3 des Zweitblicks über Phase 6.1. Jedes Bild liegt unter einem
// eigenen Namen ([service.Bildablage.NutzerbildAblegen]); der Bestand wechselt
// den Verweis mit der Fassung, und **erst danach** geht das alte Bild.
// Scheitert der Bestand, geht das **neue** — es hat keinen Verweis —, und das
// alte steht, wie es stand.
func (a *App) NutzerbildSetzen(
	ctx context.Context, nutzerID, fassung, pfad string,
) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	kennung, geleseneFassung, befund := a.nutzerbildziel(nutzerID, fassung)
	if befund != nil {
		return EinNutzerAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	// **Erst lesen, dann ablegen.** Eine Datei, die kein Bild ist, wird hier
	// abgewiesen und nicht erst, nachdem sie in der Ablage liegt (§8).
	bild, lesefehler := bilddateiLesen(pfad)
	if lesefehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, lesefehler)}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	vorher, zeilenfehler := a.nutzerzeile(ctx, kennung)
	if zeilenfehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, zeilenfehler)}
	}

	alt := vorher.Bild

	name, ablegefehler := ablage.NutzerbildAblegen(kennung, bild)
	if ablegefehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, ablegefehler)}
	}

	vorher.Bild = name
	vorher.Fassung = geleseneFassung

	if _, fehler := a.nutzer.NutzerAendern(ctx, vorher); fehler != nil {
		// **Dasselbe Bild ist derselbe Name** — dann ist das neue das alte,
		// und es bleibt.
		if name != alt {
			bildWegraeumen(ctx, ablage, name)
		}

		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if alt != "" && alt != name {
		bildWegraeumen(ctx, ablage, alt)
	}

	return a.nutzerNach(ctx, kennung)
}

// NutzerbildWaehlen fragt nach einer Bilddatei und legt sie ab.
//
// **Der Dialog gehört hierher und nicht ins Frontend** — derselbe Grund wie
// bei [Bindungen.BildSetzen]: ein Fenster bekäme einen Namen und keinen Pfad.
//
// **Ein Abbruch ist kein Fehler.** Wer den Dialog schließt, hat nichts
// gewählt; eine Meldung darüber wäre ein Vorwurf für einen Handgriff, den
// jemand absichtlich getan hat. Was zurückkommt, ist eine leere Antwort.
func (a *App) NutzerbildWaehlen(
	ctx context.Context, nutzerID, fassung string,
) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	// **Vor dem Dialog gefragt** (ID-03): wer erst nach einer Datei fragt
	// und dann absagt, hat jemanden umsonst suchen lassen.
	if befund := a.nutzerVorab(ctx, nutzerID); befund != nil {
		return EinNutzerAntwort{Fehler: befund}
	}

	pfad, err := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  DialogtexteFuer(a.sprache).BildOeffnen,
		Filter: DialogtexteFuer(a.sprache).BildFilter,
		Muster: "*.png;*.jpg;*.jpeg;*.gif",
	})
	if err != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return EinNutzerAntwort{}
	}

	return a.NutzerbildSetzen(ctx, nutzerID, fassung, pfad)
}

// NutzerbildEntfernen nimmt das Bild eines Nutzers wieder weg.
//
// **Erst der Bestand, dann die Datei** — Befund 3 des Zweitblicks über Phase
// 6.1. Die Fassung weist ab, **bevor** etwas auf der Platte geschieht;
// scheitert danach das Wegräumen, bleibt eine Datei ohne Verweis — sie zeigt
// niemand mehr an, und das Protokoll nennt sie.
func (a *App) NutzerbildEntfernen(
	ctx context.Context, nutzerID, fassung string,
) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	kennung, geleseneFassung, befund := a.nutzerbildziel(nutzerID, fassung)
	if befund != nil {
		return EinNutzerAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	vorher, zeilenfehler := a.nutzerzeile(ctx, kennung)
	if zeilenfehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, zeilenfehler)}
	}

	alt := vorher.Bild
	vorher.Bild = ""
	vorher.Fassung = geleseneFassung

	if _, fehler := a.nutzer.NutzerAendern(ctx, vorher); fehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if alt != "" {
		bildWegraeumen(ctx, ablage, alt)
	}

	return a.nutzerNach(ctx, kennung)
}

// nutzerbildziel liest Kennung und Fassung eines Profilbildrufs.
func (a *App) nutzerbildziel(nutzerID, fassung string) (
	kennung, geleseneFassung int64, befund *Fehler,
) {
	kennung, befund = benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return 0, 0, befund
	}

	geleseneFassung, fassungsbefund := benannteKennungAus("fassung", fassung)
	if fassungsbefund != nil {
		return 0, 0, fassungsbefund
	}

	return kennung, geleseneFassung, nil
}

// nutzerbildablage gibt die Ablage der Profilbilder.
func (a *App) nutzerbildablage() (*service.Bildablage, error) {
	wurzel, err := a.bilderort()
	if err != nil {
		return nil, fmt.Errorf("%w: für Bilder ist kein Verzeichnis eingerichtet: %s",
			domain.ErrValidation, err.Error())
	}

	return service.NeueBildablage(wurzel), nil
}

// nutzerzeile gibt einen Nutzer, wie der Bestand ihn führt.
//
// **Weil [storage.Nutzerablage.NutzerAendern] alle Spalten schreibt.** Wer
// nur das Bild setzte und die übrigen Felder leer ließe, löschte Anrede,
// Namen und Kurzbiographie mit.
func (a *App) nutzerzeile(ctx context.Context, nutzerID int64) (domain.Nutzer, error) {
	zeilen, fehler := a.nutzer.Uebersicht(ctx)
	if fehler != nil {
		return domain.Nutzer{}, fehler
	}

	for _, zeile := range zeilen {
		if zeile.Nutzer.ID == nutzerID {
			return zeile.Nutzer, nil
		}
	}

	return domain.Nutzer{}, fmt.Errorf("%w: den Nutzer %d führt der Bestand nicht",
		domain.ErrNotFound, nutzerID)
}

// bildWegraeumen nimmt ein Bild weg, auf das kein Verweis mehr zeigt.
//
// **Gemeldet wird ins Protokoll und nicht an die Fläche** (§7, Regel 3): der
// Wechsel ist geschrieben, und eine Meldung „gescheitert" über einem Bild,
// das sich gerade geändert hat, widerspräche dem, was gilt. Die Datei, die
// liegen bleibt, zeigt niemand an; das Löschen des Nutzers räumt sie mit
// ([service.Bildablage.NutzerbildWegnehmen]).
func bildWegraeumen(ctx context.Context, ablage *service.Bildablage, name string) {
	if fehler := ablage.BildWegnehmen(name); fehler != nil {
		slog.WarnContext(ctx, "ein Bild ohne Verweis blieb liegen",
			"datei", name, "grund", fehler.Error())
	}
}
