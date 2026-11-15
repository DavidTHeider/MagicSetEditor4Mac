package app

// Die **Wortlisten** eines Feldes — auswählen statt tippen
// (#1074).
//
// # Warum das Fenster nicht selbst rechnet
//
// Der Zustand eines Menüs hängt an drei Dingen: am Wert (welche Stelle gehört
// zu welcher Liste), am Spielpaket (was steht in der Liste) und an der Karte
// (was rechnet ein `script:` aus). Das zweite und dritte kennt nur `service`.
//
// **Und der neue Wert kommt von dort zurück, nicht aus dem Fenster.** Wer im
// Frontend `Legendary ` vor `Creature` setzte, baute MSEs `select` ein zweites
// Mal nach — mit den Präfixen in der Reihenfolge des Klickens statt in der der
// Liste.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Wortlistenbauer ist der Teil des [Planbauer], den diese Bindungen brauchen.
//
// **Eine eigene Schnittstelle und kein zweiter Empfänger**: sie wird in
// `karteneditor.go` in [Planbauer] eingebettet, sodass es weiter **eine**
// Zusage über den Planbauer gibt.
type Wortlistenbauer interface {
	// Kartenwortlisten gibt, was ein Feld zur Auswahl anbietet.
	Kartenwortlisten(ctx context.Context, karte *domain.CardRow, feld string) (
		[]service.Wortlistenbereich, []domain.Befund, error,
	)

	// Kartenwortwahl gibt den **rohen** Feldwert nach einer Wahl.
	//
	// **Der neue Wert kommt von dort und nicht aus dem Fenster**: wer ihn im
	// Frontend zusammensetzte, baute MSEs `select` ein zweites Mal nach.
	Kartenwortwahl(
		ctx context.Context, karte *domain.CardRow, feld string, bereich int, weg []int,
	) (string, []domain.Befund, error)
}

// WortmenueeintragDTO ist eine Zeile im Auswahlmenü.
type WortmenueeintragDTO struct {
	// Name ist das Wort — **unübersetzt** (§9): es steht so im Paket und wird
	// so in die Karte geschrieben.
	Name string `json:"name"`

	// IstPraefix macht aus der Zeile ein Kästchen statt eines Punktes: sie
	// lässt sich mit anderen verbinden, statt sie zu ersetzen.
	IstPraefix bool `json:"istPraefix"`

	// ZeileDarunter zieht einen Strich unter die Zeile.
	ZeileDarunter bool `json:"zeileDarunter"`

	// Aktiv sagt, dass der Wert des Feldes diesen Eintrag enthält.
	Aktiv bool `json:"aktiv"`

	// Kinder sind das Untermenü einer Gruppe — leer bei einem gewöhnlichen
	// Eintrag.
	//
	// **Keine eigene Marke für »ist eine Gruppe«**: eine Gruppe ist genau ein
	// Eintrag mit Kindern, und zwei Angaben darüber liefen auseinander.
	Kinder []WortmenueeintragDTO `json:"kinder"`
}

// WortlistenbereichDTO ist eine Stelle im Feld mit ihrem Menü.
type WortlistenbereichDTO struct {
	// Liste ist der Name der Wortliste — für die Beschriftung, wenn der Text
	// leer ist.
	Liste string `json:"liste"`

	// Text ist, was an dieser Stelle steht.
	Text string `json:"text"`

	// Kaestchen sagt, dass das Menü eine Spalte für Kästchen braucht.
	Kaestchen bool `json:"kaestchen"`

	// Gewaehlt ist die Zeile, auf der die Auswahl steht — **-1**, wenn keine
	// passt ([service.KeineWahl]).
	Gewaehlt int `json:"gewaehlt"`

	Eintraege []WortmenueeintragDTO `json:"eintraege"`
}

// WortlistenAntwort trägt die Bereiche eines Feldes oder einen Fehler.
type WortlistenAntwort struct {
	// Bereiche ist **eine leere Liste und kein `nil`** bei einem Feld ohne
	// Wortliste: über die Wails-Bindung stünde dort sonst `null`, und das
	// Frontend prüfte auf zwei Dinge statt auf eines.
	Bereiche []WortlistenbereichDTO `json:"bereiche"`

	Fehler *Fehler `json:"fehler"`
}

// Wortlisten gibt, was ein Feld zur Auswahl anbietet.
func (a *App) Wortlisten(ctx context.Context, karteID, feld string) WortlistenAntwort {
	ctx = a.imProfil(ctx)

	leer := []WortlistenbereichDTO{}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return WortlistenAntwort{Bereiche: leer, Fehler: befund}
	}

	karte, err := a.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		return WortlistenAntwort{Bereiche: leer, Fehler: Uebersetze(ctx, err)}
	}

	// **Die Befunde bleiben hier.** Eine fehlende Wortliste ist ein Mangel des
	// Pakets und steht an der Vorschau; sie ein zweites Mal an jedem Feld zu
	// zeigen wäre dieselbe Meldung sechzigmal.
	bereiche, _, err := a.planbauer.Kartenwortlisten(ctx, karte, feld)
	if err != nil {
		return WortlistenAntwort{Bereiche: leer, Fehler: Uebersetze(ctx, err)}
	}

	return WortlistenAntwort{Bereiche: bereicheDTO(bereiche)}
}

// WortWaehlen schreibt den Wert, der nach einem Klick dasteht.
//
// **Rechnen und Schreiben in einem Aufruf** (§7, Regel 4): dazwischen dürfte
// niemand tippen. Zwei Bindungen — eine, die den neuen Wert gibt, und eine, die
// ihn ablegt — ließen genau dieses Fenster offen.
func (a *App) WortWaehlen(
	ctx context.Context, setID, karteID, feld string, bereich int, weg []int,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if befund := domain.PruefeFeldname(feld); befund != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, befund)}
	}

	karte, err := a.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	wert, _, err := a.planbauer.Kartenwortwahl(ctx, karte, feld, bereich, weg)
	if err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	_, _, err = a.schreiber.FeldSetzenMitVorwert(ctx, setkennung, kennung, feld, wert)

	return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
}

// bereicheDTO setzt die Bereiche um.
func bereicheDTO(bereiche []service.Wortlistenbereich) []WortlistenbereichDTO {
	umgesetzt := make([]WortlistenbereichDTO, 0, len(bereiche))

	for _, bereich := range bereiche {
		umgesetzt = append(umgesetzt, WortlistenbereichDTO{
			Liste:     bereich.Liste,
			Text:      bereich.Text,
			Kaestchen: bereich.Menue.Kaestchen,
			Gewaehlt:  bereich.Menue.Gewaehlt,
			Eintraege: eintraegeDTO(bereich.Menue.Eintraege),
		})
	}

	return umgesetzt
}

// eintraegeDTO setzt ein Menü samt seiner Untermenüs um.
func eintraegeDTO(eintraege []service.Wortmenueeintrag) []WortmenueeintragDTO {
	umgesetzt := make([]WortmenueeintragDTO, 0, len(eintraege))

	for _, eintrag := range eintraege {
		var kinder []WortmenueeintragDTO
		if eintrag.Untermenue != nil {
			kinder = eintraegeDTO(eintrag.Untermenue.Eintraege)
		}

		umgesetzt = append(umgesetzt, WortmenueeintragDTO{
			Name:          eintrag.Name,
			IstPraefix:    eintrag.IstPraefix,
			ZeileDarunter: eintrag.ZeileDarunter,
			Aktiv:         eintrag.Aktiv,
			Kinder:        kinder,
		})
	}

	return umgesetzt
}
