// **Der Profilwechsel am Fenster und an den drei Flächen** — Werkstück 3 von
// Phase 4.12
// (#1351).

package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Profilwechsler ist die Naht zum Wechsel.
//
// **Eine eigene neben [Nutzerfuehrer]**, weil sie eine andere Frage
// beantwortet: jene sagt, wen es gibt, diese, wer gerade arbeitet.
type Profilwechsler interface {
	Wechseln(ctx context.Context, profilID int64, nachweis string) (domain.Profil, error)
	Gezeigt(ctx context.Context) (domain.Profil, error)

	// Arbeitsprofil gibt das Profil, in dem **dieser Ruf** arbeitet — A14
	// (#1453).
	//
	// **Neben [Profilwechsler.Gezeigt] und nicht statt seiner**: jenes ist
	// die Wahl eines Menschen am Fenster, an ihr hängen der Nutzerwechsel,
	// die Einstellungsübernahme und die Frage, ob der Bestand nur lesend
	// aufgeht. Dieses hier ist die Wahl **dieses Aufrufs** — ein Token, ein
	// `--profil`.
	Arbeitsprofil(ctx context.Context) (domain.Profil, error)

	// WerArbeitet gibt den Menschen, der arbeitet — Werkstück 17 von Phase
	// 4.12. **Nicht der Eigner des gezeigten Profils**: wer in ein fremdes
	// wechselt, bleibt, wer er ist
	// (#1400).
	WerArbeitet(ctx context.Context) (int64, error)

	// NutzerWechseln stellt um, wer arbeitet — gegen den Ausweis des Ziels.
	NutzerWechseln(ctx context.Context, nutzerID int64, nachweis string) error

	// Anmeldeprofil nennt das Profil, auf das eine Anmeldung zeigt — **es
	// liest nur**; die Zeiger stellt `Sitzung.Anmelden` (ID-02).
	Anmeldeprofil(ctx context.Context, nutzerID int64) (int64, error)

	// RolleImGezeigten sagt, welche Rolle der arbeitende Mensch im gezeigten
	// Profil hat — Werkstück 16 von Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// **Daran hängt, ob der Bestand nur lesend geöffnet wird.**
	RolleImGezeigten(ctx context.Context) (domain.Rolle, error)
}

// ProfilWechseln stellt das Profil um, in dem dieses Fenster arbeitet.
//
// # Zwei Stellen, nicht eine
//
// **Der Bestand merkt sich den Zeiger, das Fenster sein Profil.** Ein Wechsel,
// der nur den Bestand umstellt, ließe dieses Fenster **weiter im alten Profil
// arbeiten** — bis zum nächsten Start, und dann unerklärlich. Deshalb steht
// hier ein `a.profil.Store` neben dem Ruf in den Fachkern.
//
// **Und die Reihenfolge zählt.** Erst der Bestand, dann das Fenster: scheitert
// das Schreiben, bleibt beides beim alten. Andersherum zeigte das Fenster ein
// Profil, das der Bestand nicht kennt (§7, Regel 3).
//
// # Was hier **nicht** steht
//
// **Keine Entscheidung über den zweiten Faktor.** Der Nachweis reist
// unverändert durch; **ob** einer verlangt wird und **ob** er stimmt,
// entscheidet `internal/service` (§2). Ein leerer heißt „über den Sensor" —
// und wer keinen Faktor hat, wird nicht gefragt.
//
// **Und kein Aufräumen der Oberfläche.** Dass ein Wechsel Formular,
// Rückfrage und geladene Nebenlisten mitnimmt (§7, Regel 1), entscheidet das
// Fenster — dort, wo der Zustand liegt.
func (a *App) ProfilWechseln(
	ctx context.Context, profilID, nachweis string,
) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("profilID", profilID)
	if befund != nil {
		return EinProfilAntwort{Fehler: befund}
	}

	// **Erst fragen, dann wechseln** — Werkstück 16 von Phase 4.12. Ein
	// Wechsel, der stattfindet und danach abgelehnt wird, hätte den Zeiger
	// schon bewegt.
	if abgewiesen := a.leserwechselAbgewiesen(ctx, kennung); abgewiesen != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, abgewiesen)}
	}

	gewechselt, fehler := a.profilwechsel.Wechseln(ctx, kennung, nachweis)
	if fehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Kennung und Nähte wechseln zusammen** — Befund 1 des Zweitblicks über
	// die Phasen 4.0 bis 4.13 (2026-09-07).
	//
	// Hier stand das `Store` allein, und der Umbau der Verbindungen kam
	// danach. Wails-Bindungen laufen parallel: eine, die dazwischen lief, nahm
	// das **neue** Profil und die **alte** Verbindung — wer als Leser
	// hineinging, schrieb noch einen Augenblick lang.
	//
	// **Und dieselbe Umstellung steht an zwei weiteren Stellen** — der
	// Nutzerwechsel und das Löschen eines Profils. Seit dem Nachlauf zum
	// selben Zweitblick gehen alle drei durch [App.profilUmstellen]; dort
	// steht auch, was die Sperre hält und was nicht.
	a.profilUmstellen(ctx, gewechselt.ID)

	return EinProfilAntwort{Profil: &ProfilDTO{
		ID:       strconv.FormatInt(gewechselt.ID, 10),
		NutzerID: strconv.FormatInt(gewechselt.NutzerID, 10),
		Name:     gewechselt.Name,
	}}
}

// GezeigtesProfil sagt, in welchem Profil **dieser Ruf** arbeitet.
//
// **Aus dem Bestand und nicht aus `a.profil`**: die Seite braucht den Namen,
// und das Feld trägt nur die Kennung. Wer sie ausgäbe, zwänge jede Fläche, den
// Namen selbst zu suchen — und die Kommandozeile hätte ihn gar nicht.
//
// # Warum das Arbeitsprofil und nicht der Zeiger
//
// **Weil ein Ruf mit Token nicht im gezeigten Profil arbeitet** — Werkstück
// A14 von Phase 5.3
// (#1453).
// `Tokenwache.Pruefe` hängt das Profil des Tokens an den Kontext, und jede
// Abfrage des Bestands liest von dort; nur diese Antwort kam bis A14 vom
// **Zeiger**. Über MCP hieß das: `lesen.sets` gab die Sets des Tokenprofils
// und `lesen.profil` den Namen des gezeigten — während die
// Werkzeugbeschreibung zusagt, jede andere Antwort gelte „für DIESES Profil
// und kein anderes".
//
// **Der Name bleibt, was er ist.** Am Fenster und an der Kommandozeile ohne
// `--profil` sind beide dasselbe: [service.Nutzerverwaltung.Arbeitsprofil]
// fällt auf den Zeiger zurück, wenn der Ruf keines trägt.
func (a *App) GezeigtesProfil(ctx context.Context) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	gezeigt, fehler := a.profilwechsel.Arbeitsprofil(ctx)
	if fehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return EinProfilAntwort{Profil: &ProfilDTO{
		ID:       strconv.FormatInt(gezeigt.ID, 10),
		NutzerID: strconv.FormatInt(gezeigt.NutzerID, 10),
		Name:     gezeigt.Name,
	}}
}

// MitProfilwechsel setzt die Naht.
func (a *App) MitProfilwechsel(wechsler Profilwechsler) { a.profilwechsel = wechsler }
