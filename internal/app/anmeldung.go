// **Die Anmeldung am Fenster und an den drei Flächen** — Werkstück 4 von
// Phase 4.12
// (#1351).

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// AnbieterkontoDTO ist ein Anbieterkonto, wie das Fenster es sieht.
type AnbieterkontoDTO struct {
	Anbieter string `json:"anbieter"`

	// Anzeigename ist, wie der Anbieter diesen Menschen nennt — leer erlaubt.
	Anzeigename string `json:"anzeigename"`
}

// AnbieterAntwort trägt die Anbieter, über die man sich anmelden kann.
type AnbieterAntwort struct {
	Anbieter []string `json:"anbieter"`

	// Staende sagt je Anbieter, ob er eingerichtet ist und wogegen sein Weg
	// geprüft ist — ID-05.AK-01 (#1704).
	Staende []AnmeldeanbieterDTO `json:"staende"`

	Fehler *Fehler `json:"fehler"`
}

// AnmeldeanbieterDTO ist ein Anmeldeanbieter samt seinem Stand.
type AnmeldeanbieterDTO struct {
	Anbieter string `json:"anbieter"`

	// Eingerichtet sagt, ob der Anmeldeweg beginnen kann — `client_id` im
	// Bestand und, wo der Anbieter es verlangt, das Geheimnis im Bund. Ohne
	// das geht kein Browser auf.
	Eingerichtet bool `json:"eingerichtet"`

	// Geprueft ist `simuliert` oder `realGeprueft` ([domain.Pruefstand]).
	Geprueft string `json:"geprueft"`
}

// KontenAntwort trägt die Konten eines Nutzers.
type KontenAntwort struct {
	Konten []AnbieterkontoDTO `json:"konten"`
	Fehler *Fehler            `json:"fehler"`
}

// Kontenfuehrer ist die Naht zu den Konten im Bestand.
type Kontenfuehrer interface {
	Konten(ctx context.Context, nutzerID int64) ([]domain.Anbieterkonto, error)
	KontoLoesen(ctx context.Context, nutzerID int64, anbieter string) error
}

// MitAnmeldung setzt die Nähte.
func (a *App) MitAnmeldung(naht *Anmeldenaht, konten Kontenfuehrer) {
	a.anmeldung = naht
	a.konten = konten
}

// Anbieter nennt die Anbieter, über die man sich anmelden kann.
//
// **Ohne Kontakt nach draußen**: die Liste steht in `domain`, und wer sie
// abfragt, hat noch nichts angestoßen.
//
// **Und je Anbieter sein Stand** (ID-05.AK-01): eingerichtet heißt, der
// Anmeldeweg **kann beginnen** — Kennung, und wo nötig Geheimnis (Nachlauf,
// Befund 4); geprüft sagt, ob der Weg je gegen den echten Anbieter lief oder
// nur gegen eine Nachbildung. **Unlesbare Einstellungen
// heißen „nicht eingerichtet"** — die Liste bleibt eine Auskunft ohne Kontakt
// nach draußen, und die Anmeldung sagt beim Klick, was fehlt.
func (a *App) Anbieter(ctx context.Context) AnbieterAntwort {
	ctx = a.imProfil(ctx)

	var werte map[string]string
	if a.einstellungen != nil {
		werte, _ = a.einstellungen.Alle(ctx) //nolint:errcheck // siehe oben
	}

	anbieter := domain.Anbieter()
	staende := make([]AnmeldeanbieterDTO, 0, len(anbieter))

	for _, name := range anbieter {
		staende = append(staende, AnmeldeanbieterDTO{
			Anbieter:     name,
			Eingerichtet: a.startbereit(ctx, name, werte[domain.AnbieterKennung(name)]),
			Geprueft:     string(domain.AnbieterPruefstand(name)),
		})
	}

	return AnbieterAntwort{Anbieter: anbieter, Staende: staende}
}

// startbereit sagt, ob ein Anbieter mit dem, was vorliegt, beginnen kann.
//
// **Dieselben Zutaten wie beim Beginnen** (`zugangsquelleUeber`): die
// Kennung aus dem Bestand, das Geheimnis aus dem Bund — und ein unlesbares
// Geheimnis ist dort wie hier keines. **Ohne Kennung wird der Bund nicht
// gefragt**: wer nichts eingerichtet hat, bekommt keinen Dialog.
func (a *App) startbereit(ctx context.Context, name, kennung string) bool {
	if strings.TrimSpace(kennung) == "" {
		return false
	}

	geheimnis, _ := a.bundOderAnmeldebund().Lies(ctx, Anbieterdienst, name) //nolint:errcheck // siehe oben

	return anbieterbau[name](kennung, geheimnis).Startbereit() == nil
}

// Anmelden führt den Anmeldefluss und gibt den Nutzer, der danach arbeitet.
//
// # Was hier geschieht, geschieht auf Anweisung
//
// **Der Browser geht auf, und ein fremder Dienst wird gefragt.** Beides
// geschieht nur hier und nur, wenn jemand es verlangt hat — §16, Regel 5: ohne
// Netz arbeitet die Anwendung weiter, und ohne Anmeldung erst recht.
func (a *App) Anmelden(ctx context.Context, anbieter string) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	nutzerID, fehler := a.anmeldung.Anmelden(ctx, anbieter)
	if fehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.nutzerNach(ctx, nutzerID)
}

// nutzerNach gibt den Nutzer samt Profilen, wie die Übersicht ihn führt.
//
// **Über die Übersicht und nicht über eine eigene Abfrage**: die Seite zeigt
// nach der Anmeldung dieselbe Gestalt wie vorher, und zwei Wege dorthin wären
// zwei Gestalten.
func (a *App) nutzerNach(ctx context.Context, nutzerID int64) EinNutzerAntwort {
	zeilen, fehler := a.nutzer.Uebersicht(ctx)
	if fehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	for _, zeile := range zeilen {
		if zeile.Nutzer.ID == nutzerID {
			gelesen := nutzerDTO(zeile)

			return EinNutzerAntwort{Nutzer: &gelesen}
		}
	}

	// **Der Fall ist keiner, den es geben kann** — und deshalb steht er hier
	// trotzdem: die Anmeldung hat gerade einen Nutzer geliefert, und wenn die
	// Übersicht ihn nicht führt, ist etwas zwischen beiden auseinander. Eine
	// leere Antwort sähe aus wie „niemand angemeldet" (§8).
	return EinNutzerAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
		"%w: den angemeldeten Nutzer %d führt der Bestand nicht",
		domain.ErrNotFound, nutzerID))}
}

// Anbieterkonten gibt die Konten eines Nutzers.
func (a *App) Anbieterkonten(ctx context.Context, nutzerID string) KontenAntwort {
	ctx = a.imProfil(ctx)

	// **Ein Token liest keines Menschen Konten** (QA-09, #1740): diese
	// Auskunft steht einem Gast im gezeigten Profil offen, aber ein Token ist
	// eine Maschine und kein Gast — sie bekäme sonst die verknüpften Konten
	// eines fremden Menschen (AK-03).
	if fehler := keinTokenAmMenschen(ctx); fehler != nil {
		return KontenAntwort{Konten: []AnbieterkontoDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	// **Immer eine Liste, nie `null`** — auch auf dem Fehlerweg. Der Typ sagt
	// `[]AnbieterkontoDTO`, und das Fenster darf darüber laufen, ohne vorher zu
	// prüfen. Angestrichen von `leerelistenlaufzeit_test.go`, und §3 nennt
	// genau diesen Fall: was zwischen den Prüfebenen liegt, sieht keine Seite.
	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return KontenAntwort{Konten: []AnbieterkontoDTO{}, Fehler: befund}
	}

	konten, fehler := a.konten.Konten(ctx, kennung)
	if fehler != nil {
		return KontenAntwort{Konten: []AnbieterkontoDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	gelesen := make([]AnbieterkontoDTO, 0, len(konten))
	for _, eines := range konten {
		gelesen = append(gelesen, AnbieterkontoDTO{
			Anbieter: eines.Anbieter, Anzeigename: eines.Anzeigename,
		})
	}

	return KontenAntwort{Konten: gelesen}
}

// KontoLoesen nimmt eine fremde Kennung von einem Nutzer.
//
// # Der Rückweg gehört dazu
//
// Das Issue sagt es ausdrücklich: „ein Nutzer, dessen Anbieterkonto wegfällt,
// darf nicht ausgesperrt sein." Ohne diesen Weg wäre die Anmeldung eine
// Einbahnstraße.
func (a *App) KontoLoesen(ctx context.Context, nutzerID, anbieter string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	if fehler := a.konten.KontoLoesen(ctx, kennung, anbieter); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}
