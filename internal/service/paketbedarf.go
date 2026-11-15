package service

// **Was ein Set an Paketen braucht, und was davon fehlt** — DOM-01.AK-04
// (#1696):
// „Fehlende Abhängigkeiten lassen sich aus dem betroffenen Fachworkflow
// diagnostizieren und gezielt beheben."
//
// # Wo der Bedarf herkommt
//
// **Aus dem Setkopf** — `game:` nennt das Spiel, `stylesheet:` den Stil in
// der Kurzform, die die Stilsuche zu zwei Paketnamen macht
// (`stilpaketnamen`). **Und von dort durch `depends on:`**: was Spiel und
// Stil brauchen, braucht das Set mit, und so weiter, bis nichts Neues kommt.
// Jedes Paket wird dabei **einmal** geprüft; ein Kreis endet, statt zu
// kreisen.
//
// # Was MSE2 tut
//
// Es öffnet das Set **nicht**: ein fehlendes Paket ist beim Öffnen eine
// Ausnahme (`PackageNotFoundError`, `src/util/io/package.cpp` Zeile 80,
// Fassung f9d9356 vom 2021-07-06), und wer ein Set ohne sein Spiel hat,
// erfährt den Namen aus der Meldung. Hier öffnet es mit dem Behelfsplan
// (§8), und dieser Bedarf ist die Stelle, an der der Grund steht —
// Portierungsmarke **adapted**.

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Setbedarf nennt, was dem Set `setID` an Paketen fehlt oder zu alt ist.
//
// **Ein fehlendes Spiel oder ein fehlender Stil steht mit dem Namen des
// Sets als `Paket`** — das Set braucht es, kein anderes Paket. Für den Stil
// steht der Name, den die Stilsuche **zuerst** liest.
func (v *Paketverwalter) Setbedarf(ctx context.Context, setID int64) ([]Abhaengigkeitsbefund, error) {
	kopf, err := v.setkopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	return v.Kopfbedarf(ctx, kopf)
}

// Kopfbedarf ist [Paketverwalter.Setbedarf] für einen Kopf, der noch nicht im
// Bestand stehen muss — die Vorschau vor dem Einfahren fragt mit dem aus der
// Datei (FMT-03, #1692).
func (v *Paketverwalter) Kopfbedarf(ctx context.Context, kopf domain.SetHead) ([]Abhaengigkeitsbefund, error) {
	pakete, err := v.katalog.Pakete(ctx)
	if err != nil {
		return nil, err
	}

	da := make(map[string]domain.Paket, len(pakete))
	for _, paket := range pakete {
		da[paket.Dateiname()] = paket
	}

	befunde := []Abhaengigkeitsbefund{}
	wurzeln := []string{}

	spiel := kopf.Game + "." + string(msepaket.ArtSpiel)
	if _, gibt := da[spiel]; gibt {
		wurzeln = append(wurzeln, spiel)
	} else {
		befunde = append(befunde, Abhaengigkeitsbefund{Paket: kopf.Name, Braucht: spiel, Art: AbhaengigkeitFehlt})
	}

	if kopf.Style != "" {
		stil, gefunden := stilImKatalog(kopf, da)
		if gefunden {
			wurzeln = append(wurzeln, stil)
		} else {
			befunde = append(befunde, Abhaengigkeitsbefund{Paket: kopf.Name, Braucht: stil, Art: AbhaengigkeitFehlt})
		}
	}

	return append(befunde, bedarfDurch(wurzeln, da)...), nil
}

// stilImKatalog sucht den Stil eines Setkopfs unter beiden Namen, die die
// Stilsuche liest.
//
// **Ohne Treffer nennt es einen**, und zwar den, unter dem jemand sucht: den
// Wert, wie er dasteht, wenn er schon mit dem Spiel beginnt, sonst die
// Kurzform mit dem Spiel davor. Das ist **nur die Meldung** — gesucht wird
// wie in der Stilsuche unter beiden, und ein Präfix entscheidet dort nichts
// (der Grund steht bei [stilpaketnamen]).
func stilImKatalog(kopf domain.SetHead, da map[string]domain.Paket) (string, bool) {
	namen := stilpaketnamen(kopf.Style, kopf.Game)

	for _, name := range namen {
		if _, gibt := da[name+"."+string(msepaket.ArtStil)]; gibt {
			return name + "." + string(msepaket.ArtStil), true
		}
	}

	if strings.HasPrefix(kopf.Style, kopf.Game+"-") {
		return kopf.Style + "." + string(msepaket.ArtStil), false
	}

	return namen[0] + "." + string(msepaket.ArtStil), false
}

// bedarfDurch geht von den Wurzeln durch `depends on:` und prüft jeden
// Bedarf einmal.
func bedarfDurch(wurzeln []string, da map[string]domain.Paket) []Abhaengigkeitsbefund {
	befunde := []Abhaengigkeitsbefund{}
	gesehen := map[string]bool{}
	offen := append([]string{}, wurzeln...)

	for len(offen) > 0 {
		name := offen[0]
		offen = offen[1:]

		if gesehen[name] {
			continue
		}

		gesehen[name] = true

		for _, bedarf := range da[name].Braucht {
			if befund := pruefeBedarf(name, bedarf, da); befund != nil {
				befunde = append(befunde, *befund)
			}

			if _, gibt := da[bedarf.Paket]; gibt {
				offen = append(offen, bedarf.Paket)
			}
		}
	}

	return befunde
}

// setkopf liest den Kopf eines Sets — `ErrNotFound`, wenn es keines gibt
// oder die Fläche keine Setköpfe kennt.
func (v *Paketverwalter) setkopf(ctx context.Context, setID int64) (domain.SetHead, error) {
	if v.setkoepfe != nil {
		koepfe, err := v.setkoepfe.Setkoepfe(ctx)
		if err != nil {
			return domain.SetHead{}, err
		}

		for _, kopf := range koepfe {
			if kopf.ID == setID {
				return kopf, nil
			}
		}
	}

	return domain.SetHead{}, fmt.Errorf("%w: das Set %d führt der Bestand nicht", domain.ErrNotFound, setID)
}
