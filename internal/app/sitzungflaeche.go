// **Die Sitzung an der Kommandozeile** — ID-02
// (#1701).
//
// **Ohne Fenster gibt es keine laufende Sitzung**, wohl aber eine
// festgehaltene: ein gesperrtes Fenster bleibt es über ⌘Q hinweg, und die
// Kommandozeile sieht denselben Stand. **Deshalb dieselbe Antwort**: gesperrt
// wechselt auch hier kein Kennwort, und die Wiederherstellung über den Zettel
// geht hier wie dort.
//
// **Sperren, Entsperren, An- und Abmelden gibt es hier nicht** — sie gelten
// dem Fenster, und wer an der Kommandozeile sitzt, ist beim Betriebssystem
// angemeldet (`docs/FAEHIGKEITEN.md`).

package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// festeSitzung nimmt die festgehaltene Sitzung für den, der arbeitet, auf.
func (a *App) festeSitzung(ctx context.Context) (*service.Sitzung, error) {
	wer, fehler := a.profilwechsel.WerArbeitet(ctx)
	if fehler != nil {
		return nil, fehler
	}

	sitzung := service.NeueSitzung(a.sitzungsablage, a.faktor)
	if fehler := sitzung.Beginnen(ctx, wer); fehler != nil {
		return nil, fehler
	}

	return sitzung, nil
}

// mitFesterSitzung führt einen Schritt an ihr aus.
func (a *App) mitFesterSitzung(ctx context.Context, schritt func(*service.Sitzung) error) SchreibAntwort {
	sitzung, fehler := a.festeSitzung(ctx)
	if fehler == nil {
		fehler = schritt(sitzung)
	}

	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// Sitzungsstand sagt, wie die festgehaltene Sitzung steht.
func (f *Nutzerflaeche) Sitzungsstand(ctx context.Context) SitzungsstandAntwort {
	sitzung, fehler := f.app.festeSitzung(ctx)
	if fehler != nil {
		return SitzungsstandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	zustand, nutzerID := sitzung.Stand()

	return SitzungsstandAntwort{Zustand: string(zustand), NutzerID: strconv.FormatInt(nutzerID, 10)}
}

// KennwortWechseln setzt ein neues Kennwort für den, der arbeitet.
func (f *Nutzerflaeche) KennwortWechseln(ctx context.Context, alt, neu string) SchreibAntwort {
	return f.app.mitFesterSitzung(ctx, func(sitzung *service.Sitzung) error {
		return sitzung.KennwortWechseln(ctx, alt, neu)
	})
}

// KennwortWiederherstellen setzt ein vergessenes Kennwort über einen Code vom
// Zettel.
func (f *Nutzerflaeche) KennwortWiederherstellen(ctx context.Context, nutzerID, code, neu string) SchreibAntwort {
	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return f.app.mitFesterSitzung(ctx, func(sitzung *service.Sitzung) error {
		return sitzung.Wiederherstellen(ctx, kennung, code, neu)
	})
}
