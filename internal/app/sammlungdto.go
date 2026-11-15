// **Die Übersetzung zwischen Exemplar und DTO**
// (#1165).
//
// # Warum der Preis als Zeichenkette geht
//
// Im Bestand steht er **in der kleinsten Einheit seiner Währung als Ganzzahl**: 4,17 € sind 417, 150 ¥ sind 150. JSON
// kennt nur `number`, und das ist in JavaScript ein `double` — genau die
// Gleitkommazahl, die [domain.GeldAus] verhindert. Was über die Naht
// geht, ist deshalb die Form, in der jemand ihn tippt.
//
// **Und die Umrechnung steht an einer Stelle**, nicht in jeder Fläche.

package app

import (
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zuExemplarDTOs macht aus Exemplaren ihre DTOs.
func zuExemplarDTOs(exemplare []*domain.Exemplar) []ExemplarDTO {
	dtos := make([]ExemplarDTO, 0, len(exemplare))

	for _, exemplar := range exemplare {
		dtos = append(dtos, zuExemplarDTO(exemplar))
	}

	return dtos
}

// zuExemplarDTO bildet ein Exemplar ab.
//
// **Herausgezogen, weil es einen zweiten Aufrufer gibt** (#1292): die
// Dublettenzeile trägt **ein** Exemplar, keine Liste.
func zuExemplarDTO(exemplar *domain.Exemplar) ExemplarDTO {
	return ExemplarDTO{
		ID:              exemplar.ID,
		Fassung:         exemplar.Fassung,
		Printing:        exemplar.Printing,
		DruckID:         kennungtext(exemplar.DruckID),
		Vorlaeufig:      exemplar.Vorlaeufig(),
		Menge:           exemplar.Menge,
		Zustand:         string(exemplar.Zustand),
		Sprache:         exemplar.Sprache,
		Finish:          string(exemplar.Ausfuehrung()),
		Erwerbsdatum:    exemplar.Erwerbsdatum,
		Erwerbspreis:    preistext(exemplar.Erwerbspreis, exemplar.Erwerbswaehrung),
		Erwerbswaehrung: exemplar.Erwerbswaehrung,
		Notiz:           exemplar.Notiz,
		Name:            exemplar.Name,
		Setcode:         exemplar.Setcode,
		Nummer:          exemplar.Nummer,
		Lagerort:        exemplar.Lagerort,
		Wunsch:          exemplar.Wunsch,
		Wunschrang:      string(exemplar.Dringlichkeit()),
		Wunschart:       string(exemplar.Druckbindung()),
		Bezeichnung:     exemplar.Bezeichnung(),
		Herkunft:        alsHerkunft(exemplar.Herkunft),
	}
}

// ausExemplarDTO macht aus einem DTO ein Exemplar.
//
// **Ein unlesbarer Preis ist ein Feldbefund**, kein stiller Nullwert: wer sich
// vertippt hat, soll es am Feld sehen (§6) und nicht später an einer Summe,
// die um vier Euro danebenliegt.
func ausExemplarDTO(dto ExemplarDTO) (domain.Exemplar, error) {
	preis, err := preiswert(dto.Erwerbspreis, dto.Erwerbswaehrung)
	if err != nil {
		return domain.Exemplar{}, err
	}

	druckID, druckfehler := druckzeileAus(dto.DruckID)
	if druckfehler != nil {
		return domain.Exemplar{}, druckfehler
	}

	return domain.Exemplar{
		ID:              dto.ID,
		Fassung:         dto.Fassung,
		Printing:        dto.Printing,
		DruckID:         druckID,
		Menge:           dto.Menge,
		Zustand:         domain.Zustand(dto.Zustand),
		Sprache:         dto.Sprache,
		Finish:          domain.Finish(dto.Finish),
		Erwerbsdatum:    dto.Erwerbsdatum,
		Erwerbspreis:    preis,
		Erwerbswaehrung: dto.Erwerbswaehrung,
		Notiz:           dto.Notiz,
		Name:            dto.Name,
		Setcode:         dto.Setcode,
		Nummer:          dto.Nummer,
		Lagerort:        dto.Lagerort,
		Wunsch:          dto.Wunsch,
		Wunschrang:      domain.Wunschrang(dto.Wunschrang),
		Wunschart:       domain.Wunschart(dto.Wunschart),
	}, nil
}

// preistext macht aus der kleinsten Einheit die Form, in der jemand einen Preis liest.
//
// **Null wird zur leeren Zeichenkette**, nicht zu `"0.00"`: „kein Preis
// erfasst" und „hat nichts gekostet" sind zwei Aussagen, und ein Formularfeld,
// das von allein `0.00` zeigt, lädt zum Falschlesen ein.
// **Mit den Stellen seiner Währung** ([domain.GeldText]) — LOC-03.AK-02
// (#1689).
// Hier stand `%d.%02d`, und das war für jede Währung außer den
// zweistelligen falsch: 150 Yen standen im Feld als »1.50«.
func preistext(betrag int64, waehrung string) string {
	if betrag == 0 {
		return ""
	}

	return domain.GeldText(betrag, waehrung)
}

// preiswert macht aus der getippten Form die kleinste Einheit der Währung.
//
// **Über [domain.GeldAus] und nicht über `PreisAusText`** — LOC-03.AK-02
// (#1689):
// jene liest immer zwei Nachkommastellen und weiß nichts von der Währung.
// Für den Yen erfand sie damit eine Genauigkeit, die es nicht gibt, und für
// den Dinar verlor sie eine Stelle.
func preiswert(text, waehrung string) (int64, error) {
	if strings.TrimSpace(text) == "" {
		return 0, nil
	}

	wert, err := domain.GeldAus(text, waehrung)
	if err != nil {
		return 0, domain.NewFieldError("erwerbspreis", domain.SchluesselValidierung,
			"%q ist kein Preis: %s", text, kurzgefasst(err))
	}

	return wert, nil
}

// kurzgefasst nimmt den Satz eines Fehlers ohne seinen Umschlag.
//
// **Sonst stünde `Die Eingabe ist nicht gültig: "4,17" ist keine Zahl` im
// Feld** — die Hälfte davon sagt der Rahmen schon.
//
// **Ohne Zweig, und das ist Absicht.** Der erste Anlauf fragte, ob überhaupt
// ein Doppelpunkt dasteht — und der Fall ohne einen war keiner: was hier
// ankommt, kommt aus [domain.GeldAus], und das umschließt immer
// [domain.ErrValidation]. `TrimPrefix` lässt einen Satz ohne diesen Anfang
// unverändert und braucht dafür keine Verzweigung (§3).
func kurzgefasst(err error) string {
	return strings.TrimPrefix(err.Error(), domain.ErrValidation.Error()+": ")
}

// druckzeileAus liest die Zeile des Drucks — leer heißt keine.
//
// **Ein Feldbefund wie beim Preis**: eine Kennung, die keine Zahl ist, kommt
// nur von einem Aufrufer, der die Naht falsch bedient, und der soll es am
// Feld sehen.
func druckzeileAus(text string) (int64, error) {
	if text == "" {
		return 0, nil
	}

	kennung, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, domain.NewFieldError("druckID", domain.SchluesselValidierung,
			"Die Kennung %q ist keine Zahl.", text)
	}

	return kennung, nil
}

// AusfuehrungAus entscheidet, welche Ausführung ein Aufrufer gemeint hat.
//
// **Sie steht hier und nicht an jeder Fläche.** MCP, REST und die
// Kommandozeile führen dieselben zwei Felder; drei Flächen, die den Vorrang je
// selbst rechnen, rechnen ihn irgendwann verschieden (§16).
//
// **`finish` schlägt `foil`.** Der alte Wahrheitswert entscheidet nur, solange
// niemand die Ausführung nennt — sonst sagte ein Gegenüber, das beide schickt,
// zwei Dinge zugleich, und welches gälte, hinge an der Reihenfolge im Code.
//
// **Was er nicht kann, ist `etched`**: ein Gegenüber, das nur ihn kennt und
// eine geätzte Zeile zurückschreibt, macht sie zu `nonfoil`. Der Preis steht
// im Vermerk [0119](../../docs/entscheidungen/0119-eine-ausfuehrung-statt-eines-wahrheitswerts.md).
func AusfuehrungAus(finish string, foil bool) string {
	if finish != "" {
		return finish
	}

	return string(domain.FinishAus(foil))
}
