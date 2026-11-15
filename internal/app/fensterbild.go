package app

// **Das eigene Fensterbild** — UX-09.AK-01
// (#1681):
// *„Nutzerhintergründe besitzen Vorschau, Skalierung und eine sofort
// erreichbare Rücksetzung."*
//
// **Derselbe Weg wie beim Tischbild** — Dialog in Go, Schranke statt
// Zuschnitt, eine Adresse statt Bytes — und dieselbe Antwortgestalt
// ([TischbildAntwort]): eine Adresse, leer, solange keines abgelegt ist.
// **Ob es erscheint und wie groß**, sagt die Einstellung `hintergrund` des
// Profils; hier liegt nur das Bild.
//
// **Nur am Fenster**, wie das Tischbild: die drei Flächen ohne Fenster haben
// keinen Hintergrund, und eine Bindung dorthin wäre eine Zusage ohne
// Einlösung (§2).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// Fensterbildpfad ist der Anfang jeder Fensterbildadresse.
const Fensterbildpfad = vorschauweg.Fensterpfad

// Fensterbild sagt, ob dieser Mensch ein eigenes Fensterbild hat.
func (a *App) Fensterbild(ctx context.Context, nutzerID string) TischbildAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return TischbildAntwort{Fehler: befund}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	if _, da := ablage.FensterbildBytes(kennung); !da {
		return TischbildAntwort{}
	}

	return TischbildAntwort{Adresse: Fensterbildpfad + nutzerID + ".png"}
}

// FensterbildWaehlen fragt nach einer Bilddatei und legt sie ab.
//
// **Ein Abbruch ist kein Fehler** — wer den Dialog schließt, bekommt zurück,
// was vorher galt.
func (a *App) FensterbildWaehlen(
	ctx context.Context, nutzerID string,
) TischbildAntwort {
	ctx = a.imProfil(ctx)

	// **Vor dem Dialog gefragt** (ID-03), wie beim Tischbild.
	if befund := a.nutzerVorab(ctx, nutzerID); befund != nil {
		return TischbildAntwort{Fehler: befund}
	}

	pfad, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  DialogtexteFuer(a.sprache).BildOeffnen,
		Filter: DialogtexteFuer(a.sprache).BildFilter,
		Muster: "*.png;*.jpg;*.jpeg;*.gif",
	})
	if fehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if pfad == "" {
		return a.Fensterbild(ctx, nutzerID)
	}

	return a.fensterbildSetzen(ctx, nutzerID, pfad)
}

// fensterbildSetzen liest die Datei und legt sie ab — **erst lesen, dann
// ablegen**, damit eine Datei ohne Bild gar nicht erst in der Ablage liegt.
func (a *App) fensterbildSetzen(
	ctx context.Context, nutzerID, pfad string,
) TischbildAntwort {
	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return TischbildAntwort{Fehler: befund}
	}

	bild, lesefehler := bilddateiLesen(pfad)
	if lesefehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, lesefehler)}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	if _, ablegefehler := ablage.FensterbildAblegen(kennung, bild); ablegefehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, ablegefehler)}
	}

	return TischbildAntwort{Adresse: Fensterbildpfad + nutzerID + ".png"}
}

// FensterbildEntfernen nimmt das eigene Fensterbild wieder weg.
func (a *App) FensterbildEntfernen(
	ctx context.Context, nutzerID string,
) TischbildAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return TischbildAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	if wegfehler := ablage.FensterbildWegnehmen(kennung); wegfehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, wegfehler)}
	}

	return TischbildAntwort{}
}

// Fensterbildleitung bedient [vorschauweg.Fensterbildquelle] — als eigener
// Typ, damit Wails keine Methode mit Bytes ins Fenster bindet.
type Fensterbildleitung struct {
	app *App
}

// Fensterbilder gibt die Quelle, die der Bildweg trägt.
func Fensterbilder(a *App) *Fensterbildleitung {
	return &Fensterbildleitung{app: a}
}

// Fensterbild gibt das abgelegte Bild als PNG.
func (f *Fensterbildleitung) Fensterbild(nutzerID string) ([]byte, bool) {
	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return nil, false
	}

	ablage, ortfehler := f.app.nutzerbildablage()
	if ortfehler != nil {
		return nil, false
	}

	return ablage.FensterbildBytes(kennung)
}

// Fensterbild sagt dem Fenster, ob dieser Mensch ein eigenes Fensterbild hat.
func (b *Bindungen) Fensterbild(ctx context.Context, nutzerID string) TischbildAntwort {
	return b.app(ctx).Fensterbild(b.ruf(ctx), nutzerID)
}

// FensterbildWaehlen fragt nach einer Bilddatei und legt sie ab.
func (b *Bindungen) FensterbildWaehlen(ctx context.Context, nutzerID string) TischbildAntwort {
	return b.app(ctx).FensterbildWaehlen(b.ruf(ctx), nutzerID)
}

// FensterbildEntfernen nimmt das eigene Fensterbild wieder weg.
func (b *Bindungen) FensterbildEntfernen(ctx context.Context, nutzerID string) TischbildAntwort {
	return b.app(ctx).FensterbildEntfernen(b.ruf(ctx), nutzerID)
}
