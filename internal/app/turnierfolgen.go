package app

// **Die Folgen einer Berichtigung** — DOM-04.AK-02
// (#1699).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KorrekturfolgenAntwort trägt, was eine Berichtigung verschöbe.
type KorrekturfolgenAntwort struct {
	Folgen *KorrekturfolgenDTO `json:"folgen"`
	Fehler *Fehler             `json:"fehler"`
}

// KorrekturfolgenDTO ist `domain.Korrekturfolgen` für die Oberfläche.
type KorrekturfolgenDTO struct {
	Runde          int               `json:"runde"`
	SpaetereRunden []int             `json:"spaetereRunden"`
	Verschoben     []PlatzwechselDTO `json:"verschoben"`
}

// PlatzwechselDTO ist ein Teilnehmer vor und nach der Berichtigung.
type PlatzwechselDTO struct {
	TeilnehmerID  string `json:"teilnehmerID"`
	Name          string `json:"name"`
	PlatzVorher   int    `json:"platzVorher"`
	PlatzNachher  int    `json:"platzNachher"`
	PunkteVorher  int    `json:"punkteVorher"`
	PunkteNachher int    `json:"punkteNachher"`
}

// TurnierKorrekturfolgen sagt, was ein neues Ergebnis verschöbe — **ohne es
// einzutragen**. `stand` leer heißt: zurücknehmen, wie beim Eintragen.
func (a *App) TurnierKorrekturfolgen(
	ctx context.Context, turnierID, paarungID string, stand *ErgebnisDTO,
) KorrekturfolgenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return KorrekturfolgenAntwort{Fehler: befund}
	}

	paarung, paarungsbefund := benannteKennungAus("paarungID", paarungID)
	if paarungsbefund != nil {
		return KorrekturfolgenAntwort{Fehler: paarungsbefund}
	}

	var ergebnis *domain.Ergebnis

	if stand != nil {
		ergebnis = &domain.Ergebnis{
			SpieleA: stand.SpieleA, SpieleB: stand.SpieleB, Unentschieden: stand.Unentschieden,
		}
	}

	folgen, err := a.turniere.Korrekturfolgen(ctx, kennung, paarung, ergebnis)
	if err != nil {
		return KorrekturfolgenAntwort{Fehler: Uebersetze(ctx, err)}
	}

	dto := KorrekturfolgenDTO{
		Runde:          folgen.Runde,
		SpaetereRunden: append([]int{}, folgen.SpaetereRunden...),
		Verschoben:     make([]PlatzwechselDTO, 0, len(folgen.Verschoben)),
	}

	for _, wechsel := range folgen.Verschoben {
		dto.Verschoben = append(dto.Verschoben, PlatzwechselDTO{
			TeilnehmerID: kennungtext(wechsel.TeilnehmerID), Name: wechsel.Name,
			PlatzVorher: wechsel.PlatzVorher, PlatzNachher: wechsel.PlatzNachher,
			PunkteVorher: wechsel.PunkteVorher, PunkteNachher: wechsel.PunkteNachher,
		})
	}

	return KorrekturfolgenAntwort{Folgen: &dto}
}
