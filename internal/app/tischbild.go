package app

// **Das eigene Tischbild** — Werkstück G3 von Phase 5.5
// (#1460):
// „neutrale und rechtlich verwendbare Hintergründe **einschließlich Import**".
//
// # Warum es keinen Eintrag im Bestand bekommt
//
// **Weil der Dateiname aus der Kennung folgt.** Ein Profilbild trägt seinen
// Namen in der Nutzerzeile, weil es auch keines geben kann und die Zeile das
// sagt; hier genügt der Blick in die Ablage — liegt eine Datei, gibt es ein
// Bild. Ein zweiter Ort für dieselbe Auskunft liefe auseinander.
//
// # Und warum es dem Menschen gehört
//
// **Dieselbe Überlegung wie bei der Wahl aus den sechs Sätzen** (Frontend,
// `tisch.bild`): wer seinen Tisch so sehen will, sieht ihn so in jedem seiner
// Profile.

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// Tischbildpfad ist der Anfang jeder Tischbildadresse — [vorschauweg.Tischpfad].
const Tischbildpfad = vorschauweg.Tischpfad

// TischbildAntwort sagt, ob ein eigenes Bild vorliegt, und unter welcher
// Adresse.
type TischbildAntwort struct {
	// Adresse ist leer, solange kein eigenes Bild abgelegt ist.
	Adresse string `json:"adresse"`

	Fehler *Fehler `json:"fehler"`
}

// Tischbild sagt, ob dieser Mensch ein eigenes Tischbild hat.
func (a *App) Tischbild(ctx context.Context, nutzerID string) TischbildAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return TischbildAntwort{Fehler: befund}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	bild, lesefehler := ablage.TischbildLesen(tischbilddatei(kennung))
	if lesefehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, lesefehler)}
	}

	if bild == nil {
		return TischbildAntwort{}
	}

	return TischbildAntwort{Adresse: Tischbildpfad + nutzerID + ".png"}
}

// TischbildWaehlen fragt nach einer Bilddatei und legt sie ab.
//
// **Ein Abbruch ist kein Fehler** — dieselbe Zusage wie bei
// [App.NutzerbildWaehlen]: wer den Dialog schließt, hat nichts gewählt.
func (a *App) TischbildWaehlen(
	ctx context.Context, nutzerID string,
) TischbildAntwort {
	ctx = a.imProfil(ctx)

	// **Vor dem Dialog gefragt** (ID-03): wer erst nach einer Datei fragt
	// und dann absagt, hat jemanden umsonst suchen lassen.
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
		return a.Tischbild(ctx, nutzerID)
	}

	return a.tischbildSetzen(ctx, nutzerID, pfad)
}

// tischbildSetzen liest die Datei und legt sie ab.
//
// **Ohne eigene Frage nach dem Menschen** (ID-03): der einzige Weg hierher
// ist [App.TischbildWaehlen], und der fragt vor dem Dialog.
func (a *App) tischbildSetzen(
	ctx context.Context, nutzerID, pfad string,
) TischbildAntwort {
	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return TischbildAntwort{Fehler: befund}
	}

	// **Erst lesen, dann ablegen** — eine Datei, die kein Bild ist, wird hier
	// abgewiesen und nicht erst, nachdem sie in der Ablage liegt (§8).
	bild, lesefehler := bilddateiLesen(pfad)
	if lesefehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, lesefehler)}
	}

	ablage, ortfehler := a.nutzerbildablage()
	if ortfehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, ortfehler)}
	}

	if _, ablegefehler := ablage.TischbildAblegen(kennung, bild); ablegefehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, ablegefehler)}
	}

	return TischbildAntwort{Adresse: Tischbildpfad + nutzerID + ".png"}
}

// TischbildEntfernen nimmt das eigene Bild wieder weg.
//
// **Danach steht wieder einer der sechs Sätze da** — die Wahl bleibt, nur das
// Bild ist fort.
func (a *App) TischbildEntfernen(
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

	if wegfehler := ablage.TischbildWegnehmen(kennung); wegfehler != nil {
		return TischbildAntwort{Fehler: Uebersetze(ctx, wegfehler)}
	}

	return TischbildAntwort{}
}

// Tischbildleitung bedient [vorschauweg.Tischbildquelle].
//
// **Als eigener Typ und nicht als Methode auf [App]** — Wails bindet jede
// exportierte Methode ins Fenster, und eine, die Bytes zurückgibt, wäre dort
// die Gestalt, die dieser Weg gerade vermeidet.
type Tischbildleitung struct {
	app *App
}

// Tischbilder gibt die Quelle, die `main.go` an die Assetleitung hängt.
func Tischbilder(a *App) *Tischbildleitung {
	return &Tischbildleitung{app: a}
}

// Tischbild gibt das abgelegte Bild als PNG.
func (t *Tischbildleitung) Tischbild(nutzerID string) ([]byte, bool) {
	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return nil, false
	}

	ablage, ortfehler := t.app.nutzerbildablage()
	if ortfehler != nil {
		return nil, false
	}

	return ablage.TischbildBytes(kennung)
}

// tischbilddatei ist der Dateiname unter `nutzer/`.
//
// **Er steht hier noch einmal**, weil `service` ihn nicht hergibt: dort ist er
// ein Detail der Ablage. Dass beide dieselbe Form nennen, hält
// `tischbild_intern_test.go` fest.
func tischbilddatei(nutzerID int64) string {
	return "tischbild-" + strconv.FormatInt(nutzerID, 10) + ".png"
}

// Tischbild sagt dem Fenster, ob dieser Mensch ein eigenes Tischbild hat.
func (b *Bindungen) Tischbild(ctx context.Context, nutzerID string) TischbildAntwort {
	return b.app(ctx).Tischbild(b.ruf(ctx), nutzerID)
}

// TischbildWaehlen fragt nach einer Bilddatei und legt sie ab.
//
// **Der Dialog gehört hierher und nicht ins Frontend** — derselbe Grund wie
// bei [Bindungen.NutzerbildWaehlen]: ein Fenster bekäme einen Namen und
// keinen Pfad.
func (b *Bindungen) TischbildWaehlen(ctx context.Context, nutzerID string) TischbildAntwort {
	return b.app(ctx).TischbildWaehlen(b.ruf(ctx), nutzerID)
}

// **`tischbildSetzen` steht hier nicht**, obwohl es die Anwendung gibt.
//
// **Weil das Fenster keinen Rufer dafür hätte** — dieselbe Überlegung wie bei
// `NutzerbildSetzen`: das Bild kommt über den Dialog herein. Und die drei
// Flächen ohne Fenster haben keinen Tisch, also auch kein Tischbild; eine
// Bindung dorthin wäre eine Zusage ohne Einlösung (§2).

// TischbildEntfernen nimmt das eigene Bild wieder weg.
func (b *Bindungen) TischbildEntfernen(ctx context.Context, nutzerID string) TischbildAntwort {
	return b.app(ctx).TischbildEntfernen(b.ruf(ctx), nutzerID)
}
