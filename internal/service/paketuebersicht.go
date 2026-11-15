package service

// **Die Übersicht über die Pakete** — DOM-01.AK-01
// (#1696).
//
// Je Paket, **woher** es stammt ([domain.Paket.HerkunftsartIn]), **wer** es
// braucht ([domain.GebrauchtVon]) und **wie viele Sets** es benutzen — das,
// was der Inspector des Artboards `Pakete` unter „Eingebunden · 3 Sets
// benutzen es" und „Hängt ab von" zeigt, und woran die Liste filtert.
//
// # Wie gezählt wird
//
// **Über die Köpfe der Sets dieses Profils**: ein Spiel nutzt jedes Set, das
// es als `game:` nennt; ein Stil jedes, dessen `stylesheet:` ihn meint —
// in der Kurzform oder mit vollem Namen, wie die Stilsuche es liest
// (`stilpaketnamen`). **Andere Arten zählen nicht**: welches Set eine
// Symbolschrift oder eine Exportvorlage braucht, steht in keinem Kopf, und
// eine Null dort hieße „keines", wo es „unbekannt" ist.
//
// Portierungsmarke **new** — MSE2s Paketliste nennt keine Nutzung
// (`src/gui/package_update_list.cpp`, Fassung f9d9356 vom 2021-07-06).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Paketzeile ist ein Paket samt dem, was die Übersicht über es weiß.
type Paketzeile struct {
	Paket        domain.Paket
	Herkunftsart domain.Paketherkunftsart

	// GebrauchtVon nennt die Pakete, die dieses in `depends on:` führen.
	GebrauchtVon []string

	// Benutzt zählt die Sets dieses Profils, die es nennen — nur bei Spielen
	// und Stilen, sonst null (siehe oben).
	Benutzt int
}

// MitSetkoepfen setzt die Naht zu den Sets — für die Nutzung. **Ohne sie
// zählt die Übersicht nichts**, statt zu scheitern: eine Fläche ohne
// Bestand hat keine Sets.
func (v *Paketverwalter) MitSetkoepfen(leser Setkopfleser) *Paketverwalter {
	v.setkoepfe = leser

	return v
}

// Herkunftsart sagt, woher ein Paket stammt — dieselbe Antwort wie in der
// Übersicht, aber **ohne sie zu lesen**: sie hängt am Paket und am
// verwalteten Verzeichnis. Ein Schreibweg, dessen Nachlesen scheitert, nennt
// sie trotzdem (Zweitblick über Phase B, Befund 2).
func (v *Paketverwalter) Herkunftsart(paket domain.Paket) domain.Paketherkunftsart {
	return paket.HerkunftsartIn(v.verwaltet)
}

// Uebersicht gibt die Liste samt Herkunftsart, Bedarf und Nutzung.
func (v *Paketverwalter) Uebersicht(ctx context.Context) ([]Paketzeile, error) {
	pakete, err := v.Liste(ctx)
	if err != nil {
		return nil, err
	}

	nutzung, err := v.nutzung(ctx)
	if err != nil {
		return nil, err
	}

	zeilen := make([]Paketzeile, 0, len(pakete))

	for _, paket := range pakete {
		zeilen = append(zeilen, Paketzeile{
			Paket:        paket,
			Herkunftsart: v.Herkunftsart(paket),
			GebrauchtVon: domain.GebrauchtVon(pakete, paket),
			Benutzt:      len(nutzung[paket.Dateiname()]),
		})
	}

	return zeilen, nil
}

// nutzung nennt je Paketdateiname die Sets, die es nennen — mit Namen, denn
// der Plan vor einer Änderung zeigt sie (DOM-01.AK-02).
func (v *Paketverwalter) nutzung(ctx context.Context) (map[string][]string, error) {
	nutzung := map[string][]string{}

	if v.setkoepfe == nil {
		return nutzung, nil
	}

	koepfe, err := v.setkoepfe.Setkoepfe(ctx)
	if err != nil {
		return nil, err
	}

	for _, kopf := range koepfe {
		nutzung[kopf.Game+".mse-game"] = append(nutzung[kopf.Game+".mse-game"], kopf.Name)

		if kopf.Style == "" {
			continue
		}

		// **Beide Namen, die die Stilsuche liest** — die Kurzform mit dem
		// Spiel davor und den Wert, wie er dasteht; sie sind nie gleich.
		for _, name := range stilpaketnamen(kopf.Style, kopf.Game) {
			nutzung[name+".mse-style"] = append(nutzung[name+".mse-style"], kopf.Name)
		}
	}

	return nutzung, nil
}
