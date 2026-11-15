// **Zugangsdaten eintragen, prüfen und entfernen** — Werkstück 8 von
// Phase 4.8 (#1321).
//
// # Warum es diese Bindungen gibt
//
// **Sonst kann niemand einen Schlüssel hinterlegen.** Die Klienten sind
// gebaut, die Schalter stehen da — und ohne einen Weg, den Zugang
// einzutragen, bleiben zwei der vier Dienste für immer „ohne Zugang".
//
// # Und warum das Geheimnis nie zurückkommt
//
// **Gelesen wird nur, ob eines da ist.** Ein Weg, der den Schlüssel ins
// Fenster zurückgäbe, machte aus dem Schlüsselbund eine Durchreiche: der Wert
// stünde im JSON der Bindung, im Speicher des Browsers und in jedem
// Mitschnitt. Was das Fenster braucht, ist die **Antwort auf eine Frage** —
// „liegt einer?" —, nicht der Wert.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// Geheimnisspeicher ist der Teil des Schlüsselbunds, den die Bindungen
// brauchen.
type Geheimnisspeicher interface {
	Geheimnisleser

	Schreibe(ctx context.Context, dienst, konto, geheimnis string) error
	Loesche(ctx context.Context, dienst, konto string) error
}

// MitSchluesselbund setzt den Bund — für Tests und den Composition Root.
func (a *App) MitSchluesselbund(bund Geheimnisspeicher) { a.bund = bund }

// bundOderAnmeldebund gibt den gesetzten Bund — oder den von macOS.
func (a *App) bundOderAnmeldebund() Geheimnisspeicher {
	if a.bund == nil {
		return schluesselbund.Anmeldebund()
	}

	return a.bund
}

// ZugangsstandDTO sagt, welche Angaben ein Dienst braucht und welche liegen.
type ZugangsstandDTO struct {
	Dienst string `json:"dienst"`

	// Angaben sind die Namen der Angaben — leer, wenn der Dienst keine braucht.
	Angaben []string `json:"angaben"`

	// Hinterlegt sind die, die im Schlüsselbund liegen.
	//
	// **Namen und keine Werte.** Siehe den Kopf dieser Datei.
	Hinterlegt []string `json:"hinterlegt"`

	// Vollstaendig sagt, ob alle nötigen Angaben da sind.
	//
	// **Es ist nicht `len(Hinterlegt) == len(Angaben)`** für das Fenster:
	// diese Rechnung soll nicht an zwei Stellen stehen (§2).
	Vollstaendig bool `json:"vollstaendig"`
}

// ZugaengeAntwort trägt den Stand aller Dienste oder einen Fehler.
type ZugaengeAntwort struct {
	Zugaenge []ZugangsstandDTO `json:"zugaenge"`
	Fehler   *Fehler           `json:"fehler"`
}

// Zugaenge sagt, für welche Dienste ein Zugang hinterlegt ist.
func (a *App) Zugaenge(ctx context.Context) ZugaengeAntwort {
	ctx = a.imProfil(ctx)

	staende := make([]ZugangsstandDTO, 0, len(domain.Dienste()))

	for _, dienst := range domain.Dienste() {
		stand, err := a.zugangsstand(ctx, dienst)
		if err != nil {
			// **Eine leere Liste auch hier** (§3) — dieselbe Zusage wie auf
			// dem Erfolgsweg, und sie galt bis zum 2026-09-14 nur dort.
			return ZugaengeAntwort{
				Zugaenge: []ZugangsstandDTO{},
				Fehler:   Uebersetze(ctx, err),
			}
		}

		staende = append(staende, stand)
	}

	return ZugaengeAntwort{Zugaenge: staende}
}

// zugangsstand sammelt den Stand eines Dienstes.
func (a *App) zugangsstand(ctx context.Context, dienst domain.Dienst) (
	ZugangsstandDTO, error,
) {
	// **Eine leere Liste und kein `nil`** (§3): `nil` wird zu `null` im JSON,
	// und das Fenster bekäme dann **gar keine** Liste statt einer leeren.
	// Genau diese Klasse steht in den Leitlinien als das, was zwischen den
	// Prüfebenen durchfällt — hier hat ein Wächter sie gefangen.
	angaben := dienst.Zugangsangaben()
	if angaben == nil {
		angaben = []string{}
	}

	stand := ZugangsstandDTO{
		Dienst:       string(dienst),
		Angaben:      angaben,
		Hinterlegt:   []string{},
		Vollstaendig: true,
	}

	for _, angabe := range angaben {
		wert, err := geheimnis(ctx, a.bundOderAnmeldebund(),
			domain.ZugangKonto(dienst, angabe))
		if err != nil {
			return ZugangsstandDTO{}, err
		}

		if wert == "" {
			stand.Vollstaendig = false

			continue
		}

		stand.Hinterlegt = append(stand.Hinterlegt, angabe)
	}

	return stand, nil
}

// ZugangSetzen legt eine Angabe im Schlüsselbund ab.
func (a *App) ZugangSetzen(
	ctx context.Context, dienst, angabe, geheimnis string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	gewaehlt, befund := gepruefteAngabe(dienst, angabe)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx,
		a.bundOderAnmeldebund().Schreibe(ctx, Schluesselbunddienst,
			domain.ZugangKonto(gewaehlt, angabe), geheimnis))}
}

// ZugangLoeschen nimmt eine Angabe heraus.
func (a *App) ZugangLoeschen(ctx context.Context, dienst, angabe string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	gewaehlt, befund := gepruefteAngabe(dienst, angabe)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx,
		a.bundOderAnmeldebund().Loesche(ctx, Schluesselbunddienst,
			domain.ZugangKonto(gewaehlt, angabe)))}
}

// gepruefteAngabe prüft, dass es den Dienst gibt und er diese Angabe braucht.
//
// **Beides und nicht nur das erste.** „Scryfall braucht keinen Schlüssel" ist
// eine andere Auskunft als „diesen Dienst gibt es nicht" — und wer eine
// Angabe für einen Dienst ablegte, der sie nie liest, bekäme ein Geheimnis im
// Bund, das nichts tut (§8).
func gepruefteAngabe(dienst, angabe string) (domain.Dienst, *Fehler) {
	gewaehlt := domain.Dienst(dienst)

	angaben := gewaehlt.Zugangsangaben()
	if len(angaben) == 0 {
		return "", Nutzerfehler(domain.NewFieldError("dienst",
			domain.SchluesselValidierung,
			"Der Dienst »%s« kennt keine Zugangsdaten.", dienst))
	}

	for _, eine := range angaben {
		if eine == angabe {
			return gewaehlt, nil
		}
	}

	return "", Nutzerfehler(domain.NewFieldError("angabe",
		domain.SchluesselValidierung,
		"Der Dienst »%s« kennt die Angabe »%s« nicht.", dienst, angabe))
}

// NimmtDenAnmeldebund sagt, ob der Anmeldebund gewählt würde.
//
// **Nur für den Lauf, der das belegt.** Er darf den Anmeldebund nicht
// anfassen (§3) — also **wählt** er ihn und sieht nach, dass einer
// herauskommt, statt hineinzuschreiben.
//
// **Und das geht gefahrlos**: `schluesselbund.Anmeldebund()` baut nur den
// Verweis; `security` läuft erst, wenn jemand liest oder schreibt.
func (a *App) NimmtDenAnmeldebund() bool {
	return a.bundOderAnmeldebund() != nil && a.bund == nil
}

// **Die drei Bindungen ans Fenster** — hier und nicht in `bindungen.go`.
//
// **Der Schnitt folgt der Frage** (§2): dort steht, was das Fenster **rufen**
// kann, hier, was ein **Zugang** ist und wie er verwaltet wird. Und
// `bindungen.go` stand mit ihnen bei 412 Zeilen.

// Zugaenge sagt, für welche Dienste ein Zugang hinterlegt ist.
//
// **Sie gibt Namen und keine Werte.** Siehe den Kopf dieser Datei.
func (b *Bindungen) Zugaenge(ctx context.Context) ZugaengeAntwort {
	return b.app(ctx).Zugaenge(b.ruf(ctx))
}

// ZugangSetzen legt eine Angabe im Schlüsselbund ab.
func (b *Bindungen) ZugangSetzen(ctx context.Context, dienst, angabe, geheimnis string) SchreibAntwort {
	return b.app(ctx).ZugangSetzen(b.ruf(ctx), dienst, angabe, geheimnis)
}

// ZugangLoeschen nimmt eine Angabe heraus.
func (b *Bindungen) ZugangLoeschen(ctx context.Context, dienst, angabe string) SchreibAntwort {
	return b.app(ctx).ZugangLoeschen(b.ruf(ctx), dienst, angabe)
}
