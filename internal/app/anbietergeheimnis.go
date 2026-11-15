package app

// **Das Geheimnis eines Anmeldeanbieters** —
// #1401,
// Befund 3 des Zweitblicks über Phase 4.12.
//
// # Wofür es diese drei Bindungen gibt
//
// **GitHub verlangt neben der `client_id` ein `client_secret`**
// (`anmeldung.GitHub`, `GeheimnisPflicht: true`), und **keine Bindung schrieb
// eines**. Gelesen wurde es aus dem Schlüsselbund; hineingeschrieben werden
// musste es mit der Schlüsselbundverwaltung oder `security`. Damit war die
// Anmeldung über GitHub praktisch nicht einzurichten.
//
// **§6 sagt, ein leerer Zustand erklärt den nächsten Schritt** — nicht: er
// verweist auf ein anderes Programm.
//
// # Der Zuschnitt kommt von nebenan
//
// **Dieselben drei Fragen wie bei `Zugaenge`** (`zugangbindungen.go`): was
// braucht es, was liegt vor, und hinein damit. **Und dieselbe Zusage: der Wert
// kommt nie zurück.** Ein Weg, der ihn ins Fenster gäbe, machte aus dem
// Schlüsselbund eine Durchreiche — er stünde im JSON der Bindung, im Speicher
// des Browsers und in jedem Mitschnitt.
//
// # Warum nicht über `domain.Dienst`
//
// **Weil das etwas anderes ist.** `Dienst` sind die Kartendienste — Scryfall,
// Scrydex, MTGJSON —, und ihre Geheimnisse liegen unter einem anderen
// Bundnamen ([Schluesselbunddienst] gegen [Anbieterdienst]). Sie
// zusammenzulegen hieße, zwei Fragen mit einer Liste zu beantworten.

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// AnbietergeheimnisDTO sagt, was ein Anbieter braucht und was vorliegt.
type AnbietergeheimnisDTO struct {
	Anbieter string `json:"anbieter"`

	// Pflicht und Verboten kommen aus `internal/anmeldung` und sind dort je
	// Anbieter nachgelesen.
	Pflicht  bool `json:"pflicht"`
	Verboten bool `json:"verboten"`

	// Hinterlegt sagt, **ob** eines liegt — nie welches.
	Hinterlegt bool `json:"hinterlegt"`
}

// AnbietergeheimnisseAntwort trägt die Liste.
type AnbietergeheimnisseAntwort struct {
	Anbieter []AnbietergeheimnisDTO `json:"anbieter"`
	Fehler   *Fehler                `json:"fehler"`
}

// Anbietergeheimnisse sagt je Anbieter, was er braucht und was vorliegt.
func (a *App) Anbietergeheimnisse(ctx context.Context) AnbietergeheimnisseAntwort {
	ctx = a.imProfil(ctx)

	// **Eine leere Liste und kein `nil`** (§3): `nil` wird zu `null` im JSON,
	// und das Fenster bekäme dann gar keine Liste statt einer leeren.
	liste := []AnbietergeheimnisDTO{}

	for _, name := range domain.Anbieter() {
		eintrag, fehler := a.anbietergeheimnis(ctx, name)
		if fehler != nil {
			// **Auch hier, und das ist der Punkt.** Der Satz darüber stand
			// da, seit es diese Bindung gibt — und galt nur für den
			// Erfolgsweg: der Fehlerzweig gab `null`. Wer nur den einen
			// prüft, hat die Hälfte ausgelassen (§10). Angestrichen hat es
			// die Sperre gegen den Anmeldebund am 2026-09-14: vorher kam
			// hier nie ein Fehler an, weil der Schlüsselbund dieses
			// Rechners immer antwortete.
			return AnbietergeheimnisseAntwort{
				Anbieter: []AnbietergeheimnisDTO{},
				Fehler:   Uebersetze(ctx, fehler),
			}
		}

		liste = append(liste, eintrag)
	}

	return AnbietergeheimnisseAntwort{Anbieter: liste}
}

// anbietergeheimnis baut einen Eintrag.
//
// **Ohne Wache auf den Namen**, und das ist eine Aussage: gerufen wird nur mit
// den Namen aus [domain.Anbieter], und dass [anbieterbau] genau diese führt,
// hält `TestJederAnbieterHatEinenBauplan` daneben. Ein `if` darauf wäre ein
// Zweig, den kein Aufruf betritt — also eine Behauptung darüber, was im
// Ernstfall geschieht (§3), und der Wächter sagt mehr als sie.
func (a *App) anbietergeheimnis(ctx context.Context, name string) (
	AnbietergeheimnisDTO, error,
) {
	// **Der Bauplan wird leer gerufen**: gefragt sind seine zwei Schalter,
	// nicht seine Zugangsdaten.
	gebaut := anbieterbau[name]("", "")

	wert, lesefehler := a.bundOderAnmeldebund().Lies(ctx, Anbieterdienst, name)
	if lesefehler != nil && !errors.Is(lesefehler, domain.ErrNotFound) {
		return AnbietergeheimnisDTO{}, lesefehler
	}

	return AnbietergeheimnisDTO{
		Anbieter:   name,
		Pflicht:    gebaut.GeheimnisPflicht,
		Verboten:   gebaut.GeheimnisVerboten,
		Hinterlegt: wert != "",
	}, nil
}

// AnbietergeheimnisSetzen legt das Geheimnis eines Anbieters im Bund ab.
//
// **Wer eines verbietet, bekommt keines.** Microsoft weist einen öffentlichen
// Klienten mit Geheimnis ab; ein Eintrag, der die Anmeldung scheitern lässt,
// ist schlimmer als keiner — und er wäre schwer zu finden, weil ihn niemand
// sieht.
func (a *App) AnbietergeheimnisSetzen(
	ctx context.Context, anbieter, geheimnis string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	bauplan, kennt := anbieterbau[anbieter]
	if !kennt {
		return SchreibAntwort{Fehler: Uebersetze(ctx, unbekannterAnbieter(anbieter))}
	}

	if bauplan("", "").GeheimnisVerboten {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: %s weist ein Geheimnis von einem öffentlichen Klienten ab — "+
				"trage nur die Kennung ein", domain.ErrValidation, anbieter))}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx,
		a.bundOderAnmeldebund().Schreibe(ctx, Anbieterdienst, anbieter, geheimnis))}
}

// AnbietergeheimnisLoeschen nimmt es wieder heraus.
func (a *App) AnbietergeheimnisLoeschen(
	ctx context.Context, anbieter string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if _, kennt := anbieterbau[anbieter]; !kennt {
		return SchreibAntwort{Fehler: Uebersetze(ctx, unbekannterAnbieter(anbieter))}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx,
		a.bundOderAnmeldebund().Loesche(ctx, Anbieterdienst, anbieter))}
}

// unbekannterAnbieter ist die Abweisung, die alle drei teilen.
func unbekannterAnbieter(anbieter string) error {
	return fmt.Errorf("%w: %q ist keiner der Anbieter %v",
		domain.ErrValidation, anbieter, domain.Anbieter())
}

// **Und dieselben drei am Fenster — und nur dort.**
//
// **Der Grund ist der Schlüsselbund**, und er steht schon bei `Zugaenge`
// nebenan: macOS fragt beim ersten Zugriff eines Programms nach Erlaubnis. Ein
// Befehl, der öffnet, arbeitet und endet, stellte diese Frage bei **jedem**
// Aufruf; das Fenster stellt sie einmal. Wer sie ohne Fenster beantworten
// will, tut es mit `security` — der Weg steht in `docs/ADMINISTRATION.md`.

// Anbietergeheimnisse sagt es dem Fenster.
func (b *Bindungen) Anbietergeheimnisse(ctx context.Context) AnbietergeheimnisseAntwort {
	return b.app(ctx).Anbietergeheimnisse(b.ruf(ctx))
}

// AnbietergeheimnisSetzen legt es ab.
func (b *Bindungen) AnbietergeheimnisSetzen(ctx context.Context, anbieter, geheimnis string) SchreibAntwort {
	return b.app(ctx).AnbietergeheimnisSetzen(b.ruf(ctx), anbieter, geheimnis)
}

// AnbietergeheimnisLoeschen nimmt es heraus.
func (b *Bindungen) AnbietergeheimnisLoeschen(ctx context.Context, anbieter string) SchreibAntwort {
	return b.app(ctx).AnbietergeheimnisLoeschen(b.ruf(ctx), anbieter)
}
