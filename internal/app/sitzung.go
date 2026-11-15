// **Die lokale Sitzung am Fenster** — ID-02
// (#1701).
//
// **Keine fachliche Entscheidung hier** (§2): wann gesperrt, entsperrt,
// abgemeldet werden darf, entscheidet [service.Sitzung]. Hier steht nur, dass
// eine verschlossene Sitzung **kein Profil** in den Kontext bekommt
// ([App.imProfilOhneSperre]) — und damit keine Abfrage den Bestand eines
// Profils erreicht, gleich, was das Fenster fragt.
//
// **Nur das Fenster hat eine Sitzung.** Kommandozeile, Dienst und
// Werkzeugserver laufen unter dem Konto von macOS und bringen ihre eigene
// Anmeldung mit (REST: Zugangsschlüssel); was sie von hier können, steht in
// `docs/FAEHIGKEITEN.md`.

package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SitzungsstandAntwort sagt, wie es um die Sitzung steht — und was der
// Sperrbildschirm anbieten kann.
type SitzungsstandAntwort struct {
	Zustand  string `json:"zustand"`
	NutzerID string `json:"nutzerID"`

	// Kennwort sagt, ob der Nutzer eines hat; Sensor, ob er auf diesem
	// Rechner **lokal entsperren** kann (Touch ID — kein Passkey, AK-02).
	Kennwort bool `json:"kennwort"`
	Sensor   bool `json:"sensor"`

	Fehler *Fehler `json:"fehler"`
}

// sitzungBeginnen nimmt beim Start den festgehaltenen Zustand auf.
//
// **Ein Fehler lässt die Sitzung abgemeldet und nicht offen** (Zweitblick
// über ID-02, Befund 1): ein zeitweiliger Lesefehler höbe sonst eine
// festgehaltene Sperre auf. Abgemeldet heißt: die Anmeldung steht da, und
// wer ein Kennwort hat, braucht es — wer keines hat, kommt herein wie
// vorher.
func (a *App) sitzungBeginnen(ctx context.Context, ablage service.Sitzungsablage) {
	ctx = a.imProfilTrotzSitzung(ctx)
	sitzung := service.NeueSitzung(ablage, a.faktor)

	wer, fehler := a.profilwechsel.WerArbeitet(ctx)
	if fehler == nil {
		fehler = sitzung.Beginnen(ctx, wer)
	}

	if fehler != nil {
		slog.WarnContext(ctx, "Die Sitzung ließ sich nicht aufnehmen", "grund", fehler)
	}

	a.sitzung.Store(sitzung)
}

// sitzungVerschlossen sagt, ob die Sitzung des Fensters gesperrt oder
// beendet ist.
func (a *App) sitzungVerschlossen() bool {
	sitzung := a.sitzung.Load()
	if sitzung == nil {
		return false
	}

	zustand, _ := sitzung.Stand()

	return !zustand.ErreichtDenBestand()
}

// ohneSitzung ist die Antwort, wenn es keine gibt — ohne Fenster oder vor dem
// Start.
func ohneSitzung(ctx context.Context) *Fehler {
	return Uebersetze(ctx, fmt.Errorf("%w: eine Sitzung gibt es nur am Fenster", domain.ErrConflict))
}

// Sitzungsstand sagt, ob gearbeitet, gesperrt oder abgemeldet ist.
func (a *App) Sitzungsstand(ctx context.Context) SitzungsstandAntwort {
	ctx = a.imProfil(ctx)

	sitzung := a.sitzung.Load()
	if sitzung == nil {
		return SitzungsstandAntwort{Zustand: string(domain.SitzungAngemeldet)}
	}

	zustand, nutzerID := sitzung.Stand()
	antwort := SitzungsstandAntwort{Zustand: string(zustand), NutzerID: strconv.FormatInt(nutzerID, 10)}

	if nutzerID == 0 {
		return antwort
	}

	auskunft, fehler := sitzung.Auskunft(ctx)
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	antwort.Kennwort, antwort.Sensor = auskunft.Kennwort, auskunft.Sensor

	return antwort
}

// Sperren sperrt das Fenster — und beendet, was am Profil hängt.
//
// **Die Dienste zuerst** (Zweitblick über ID-02, Befund 6): ein REST- oder
// MCP-Dienst, ein Horchposten und offene Peersitzungen, die das Fenster
// gestartet hat, bedienten sonst weiter Anfragen für einen Bestand, den
// gerade niemand sehen darf — dieselben Schritte wie beim Profilwechsel.
func (a *App) Sperren(ctx context.Context) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.inDerSitzung(ctx, func(sitzung *service.Sitzung) error {
		if fehler := sitzung.Sperren(ctx); fehler != nil {
			return fehler
		}

		a.profilgebundenesRaeumen()

		return nil
	})
}

// Entsperren löst die Sperre mit dem Kennwort.
func (a *App) Entsperren(ctx context.Context, kennwort string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.inDerSitzung(ctx, func(sitzung *service.Sitzung) error { return sitzung.Entsperren(ctx, kennwort) })
}

// EntsperrenMitSensor löst die Sperre über Touch ID.
func (a *App) EntsperrenMitSensor(ctx context.Context) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.inDerSitzung(ctx, func(sitzung *service.Sitzung) error { return sitzung.EntsperrenUeberDenSensor(ctx) })
}

// Abmelden beendet die Sitzung — und mit ihr, was am Profil hängt.
func (a *App) Abmelden(ctx context.Context) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.inDerSitzung(ctx, func(sitzung *service.Sitzung) error {
		if fehler := sitzung.Abmelden(ctx); fehler != nil {
			return fehler
		}

		a.profilgebundenesRaeumen()

		return nil
	})
}

// LokalAnmelden meldet einen Nutzer dieses Rechners an — Kennwort, dann
// **streng** der zweite Faktor, dann stellen die Zeiger um.
//
// **Die Umstellung prüft nicht noch einmal** (Zweitblick über ID-02, Befund
// 2): der Nutzerwechsel kehrte für den, der schon arbeitete, ohne Prüfung
// zurück, und seine Prüfung lässt sich abschalten. **Die Zeiger schreibt die
// Sitzung, zusammen mit ihrem Zustand** (Nachlauf, Befund 2); hier stehen nur
// das Profil und der Schritt im Prozess — beide trotz verschlossener Sitzung,
// denn sie geht erst danach auf.
func (a *App) LokalAnmelden(ctx context.Context, nutzerID, kennwort, nachweis string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return a.inDerSitzung(ctx, func(sitzung *service.Sitzung) error {
		return sitzung.Anmelden(ctx, kennung, kennwort, nachweis, service.Anmeldeziel{
			Profil: func(c context.Context, nutzerID int64) (int64, error) {
				return a.profilwechsel.Anmeldeprofil(a.imProfilTrotzSitzung(c), nutzerID)
			},
			Nachziehen: func(c context.Context, profilID int64) {
				a.profilUmstellen(a.imProfilTrotzSitzung(c), profilID)
			},
		})
	})
}

// KennwortWechseln setzt ein neues Kennwort für den, der arbeitet.
func (a *App) KennwortWechseln(ctx context.Context, alt, neu string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.inDerSitzung(ctx, func(sitzung *service.Sitzung) error { return sitzung.KennwortWechseln(ctx, alt, neu) })
}

// KennwortWiederherstellen setzt ein vergessenes Kennwort über einen Code vom
// Zettel — und sonst nichts (AK-04).
func (a *App) KennwortWiederherstellen(ctx context.Context, nutzerID, code, neu string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return a.inDerSitzung(ctx, func(sitzung *service.Sitzung) error {
		return sitzung.Wiederherstellen(ctx, kennung, code, neu)
	})
}

// inDerSitzung führt einen Schritt an der Sitzung des Fensters aus.
func (a *App) inDerSitzung(ctx context.Context, schritt func(*service.Sitzung) error) SchreibAntwort {
	sitzung := a.sitzung.Load()
	if sitzung == nil {
		return SchreibAntwort{Fehler: ohneSitzung(ctx)}
	}

	if fehler := schritt(sitzung); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Und die Menüleiste folgt**: gesperrt zeigte sie sonst weiter die
	// zuletzt benutzten Sets dessen, der gesperrt hat — ohne Profil ist die
	// Liste leer ([App.zuletztBenutzte]), entsperrt ist sie wieder da.
	a.menueNeu(a.imProfil(ctx))

	return SchreibAntwort{}
}
