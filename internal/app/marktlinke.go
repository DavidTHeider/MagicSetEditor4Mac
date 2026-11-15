package app

// **Die Wege nach draußen am Fenster** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Zwei Rufe und nicht einer
//
// **Weil Ansehen und Öffnen zwei Dinge sind.** Die Liste entsteht, ohne dass
// etwas geschieht; das Öffnen verlässt dieses Programm. Wer beides
// zusammenlegte, öffnete einen Browser, weil jemand eine Karte aufgeschlagen
// hat.
//
// # Und warum die Adresse beim Öffnen noch einmal geprüft wird
//
// **Weil sie dazwischen durch fremde Hände geht.** Sie stand in einer
// Antwort, lag im Fenster, kam über REST zurück — und was von dort kommt, ist
// nicht das, was hier gebaut wurde, sondern das, was jemand schickt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MarktlinkDTO ist ein Weg nach draußen.
type MarktlinkDTO struct {
	// Anbieter ist `cardmarket`, `cardtrader`, `tcgplayer`, `cardkingdom`,
	// `gatherer` oder `mtggoldfish`.
	Anbieter string `json:"anbieter"`

	// Art ist `produkt`, `suche` oder `startseite` — **die Beschriftung
	// hängt daran**: ein Suchlink verspricht nicht, was ein Produktlink
	// hält.
	Art string `json:"art"`

	// Herkunft ist `kennung`, `name` oder `keine`.
	Herkunft string `json:"herkunft"`

	Adresse string `json:"adresse"`
}

// MarktlinkAntwort trägt die Wege samt ihrem Fehler.
type MarktlinkAntwort struct {
	Links  []MarktlinkDTO `json:"links"`
	Fehler *Fehler        `json:"fehler"`
}

// Marktlinks gibt je Anbieter den genauesten Weg, der zu haben ist.
//
// **`printing` darf leer sein** — eine eigene Karte hat keinen kanonischen
// Druck und bekommt deshalb keine geratenen Produktlinks.
func (a *App) Marktlinks(
	ctx context.Context, printing, name string,
) MarktlinkAntwort {
	ctx = a.imProfil(ctx)

	links, fehler := a.linkwerk().Linksfuer(ctx, printing, name)
	if fehler != nil {
		return MarktlinkAntwort{
			Links: []MarktlinkDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return MarktlinkAntwort{Links: alsMarktlinks(links)}
}

// MarktlinkOeffnen macht einen geprüften Weg im Standardbrowser auf.
//
// **Der Anbieter kommt mit**, denn ohne ihn wüsste die Prüfung nicht, gegen
// welchen Wirt sie prüft.
func (a *App) MarktlinkOeffnen(
	ctx context.Context, anbieter, adresse string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.linkwerk().Oeffne(
		ctx, a.browseroeffner, anbieter, adresse); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// alsMarktlinks macht aus den Wegen ihre Gestalt für das Fenster.
func alsMarktlinks(links []domain.Marktlink) []MarktlinkDTO {
	dtos := make([]MarktlinkDTO, 0, len(links))

	for _, link := range links {
		dtos = append(dtos, MarktlinkDTO{
			Anbieter: string(link.Anbieter), Art: string(link.Art),
			Herkunft: string(link.Herkunft), Adresse: link.Adresse,
		})
	}

	return dtos
}
