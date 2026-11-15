// **Wer arbeitet, und wie man ein anderer wird** — Werkstück 17 von Phase 4.12
// (#1400),
// nach [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).
//
// **Keine fachliche Entscheidung hier** (§2): übersetzt wird zwischen Wails
// und dem Dienst.

package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// WerArbeitetAntwort nennt den Menschen, der gerade arbeitet.
//
// **Die Kennung und nicht der ganze Nutzer.** Wer die Angaben braucht, hat sie
// aus der Übersicht — und eine zweite Stelle, die einen Nutzer zusammensetzt,
// liefe mit der ersten auseinander.
type WerArbeitetAntwort struct {
	NutzerID string  `json:"nutzerID"`
	Fehler   *Fehler `json:"fehler"`
}

// WerArbeitet sagt, wer gerade arbeitet.
//
// # Warum es diese Bindung gibt
//
// **Weil die Antwort seit Werkstück 17 nicht mehr aus dem gezeigten Profil
// abzulesen ist.** Vorher hätte das Fenster `GezeigtesProfil().NutzerID`
// genommen; das ist jetzt der Eigner des Profils, in dem gearbeitet wird —
// und der ist ein anderer, sobald jemand in ein fremdes gewechselt ist.
func (a *App) WerArbeitet(ctx context.Context) WerArbeitetAntwort {
	ctx = a.imProfil(ctx)

	nutzerID, fehler := a.profilwechsel.WerArbeitet(ctx)
	if fehler != nil {
		return WerArbeitetAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return WerArbeitetAntwort{NutzerID: strconv.FormatInt(nutzerID, 10)}
}

// NutzerWechseln stellt um, wer arbeitet — gegen den Ausweis des Ziels.
//
// **Der Nachweis ist der zweite Faktor des Menschen, der man wird.** Ein
// leerer heißt „über den Sensor"; wer keinen Faktor führt, wird nicht gefragt.
// Warum das der Ausweis des **Ziels** ist und nicht der eigene, steht im
// Dienst.
func (a *App) NutzerWechseln(
	ctx context.Context, nutzerID, nachweis string,
) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return EinProfilAntwort{Fehler: befund}
	}

	gezeigt, fehler := a.nutzerWechseln(ctx, kennung, nachweis)
	if fehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Die Sitzung des Fensters geht mit** (ID-02) — sonst sperrte „Sperren"
	// danach den, der vorher arbeitete.
	if sitzung := a.sitzung.Load(); sitzung != nil {
		sitzung.Umgestellt(kennung)
	}

	return EinProfilAntwort{Profil: &ProfilDTO{
		ID:       strconv.FormatInt(gezeigt.ID, 10),
		NutzerID: strconv.FormatInt(gezeigt.NutzerID, 10),
		Name:     gezeigt.Name,
	}}
}

// nutzerWechseln stellt um und zieht das Profil nach.
//
// **Ohne die Sitzung**: [App.LokalAnmelden] ruft es unter ihrem Schloss, und
// ein Ruf an [service.Sitzung.Umgestellt] klemmte dort gegen sich selbst.
func (a *App) nutzerWechseln(ctx context.Context, kennung int64, nachweis string) (domain.Profil, error) {
	if fehler := a.profilwechsel.NutzerWechseln(ctx, kennung, nachweis); fehler != nil {
		return domain.Profil{}, fehler
	}

	// **Das neue Profil kommt zurück** — der Wechsel zieht den Zeiger mit, und
	// die Fläche schreibt es sich an. Wer es danach erst holte, hätte zwei
	// Rufe und zwei Stände, zwischen denen sich etwas ändern kann (§7,
	// Regel 2).
	gezeigt, fehler := a.profilwechsel.Gezeigt(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	// **Kennung und Leseart in einem Zug, unter der Umschaltsperre** — wie
	// beim Profilwechsel: wer ein anderer Mensch wird, wechselt ebenso das
	// Profil, und eine Bindung dazwischen nähme das neue Profil mit der alten
	// Verbindung.
	a.profilUmstellen(ctx, gezeigt.ID)

	return gezeigt, nil
}

// WerArbeitet gibt den Menschen, der gerade arbeitet.
func (b *Bindungen) WerArbeitet(ctx context.Context) WerArbeitetAntwort {
	return b.app(ctx).WerArbeitet(b.ruf(ctx))
}

// NutzerWechseln stellt um, wer arbeitet.
func (b *Bindungen) NutzerWechseln(ctx context.Context, nutzerID, nachweis string) EinProfilAntwort {
	return b.app(ctx).NutzerWechseln(b.ruf(ctx), nutzerID, nachweis)
}

// WerArbeitet sagt, wer arbeitet — für die Flächen ohne Fenster.
//
// **Auf der Nutzerfläche und nicht auf `Bindungen`**: Kommandozeile,
// MCP-Server und Dienst sehen die Anwendung durch sie (§16, dieselbe Fähigkeit
// überall).
func (f *Nutzerflaeche) WerArbeitet(ctx context.Context) WerArbeitetAntwort {
	return f.app.WerArbeitet(ctx)
}

// NutzerWechseln stellt um, wer arbeitet.
func (f *Nutzerflaeche) NutzerWechseln(
	ctx context.Context, nutzerID, nachweis string,
) EinProfilAntwort {
	return f.app.NutzerWechseln(ctx, nutzerID, nachweis)
}
