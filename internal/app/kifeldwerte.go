package app

// **Was aus der Antwort eines Modells wird** — die Felder, ihr Vergleich mit
// dem Bestand und die Auswahl dessen, was durchkam. Herausgelöst aus
// `kivorschlag.go`, als AI-05 (#2042) dort Vorrang und Rückfall brauchte
// und die Datei an der Zeilengrenze stand (§2).

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// feldwerteAus liest die Antwort des Modells.
//
// **Ein Wert kann Text oder Liste sein** — ein Modell antwortet auf ein
// einfaches Feld mit einer Zeichenkette und auf ein Mehrfachfeld mit einem
// Array. Wer nur eines von beiden läse, verlöre die Hälfte jeder Antwort.
func feldwerteAus(roh KIRohantwort) (map[string][]string, error) {
	if roh.Abgeschnitten {
		return nil, fmt.Errorf(
			"%w: die Antwort brach am Fensterende ab — was dasteht, ist eine "+
				"halbe Antwort und sieht aus wie eine ganze",
			domain.ErrValidation)
	}

	var gelesen map[string]json.RawMessage

	if fehler := json.Unmarshal([]byte(roh.Inhalt), &gelesen); fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Antwort von %s war kein JSON-Objekt", roh.Modell)
	}

	felder := make(map[string][]string, len(gelesen))

	for name, wert := range gelesen {
		werte, fehler := werteAus(wert)
		if fehler != nil {
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"den Wert für %q lesen", name)
		}

		// **Ein leerer Wert ist kein Vorschlag** — N1 des Nachlaufs über
		// Phase 6.10: Z1 hielt nur den Rückfall des Vorrangs davon ab; am
		// gewählten Anbieter, im Stapel und im Einzelzug blieb er ein
		// Vorschlag, und `uebernehmen` schrieb ihn. Hier geht jede Antwort
		// durch, und hier fällt er heraus.
		if mitWerten := ohneLeere(werte); len(mitWerten) > 0 {
			felder[name] = mitWerten
		}
	}

	return felder, nil
}

// ohneLeere lässt Werte weg, die nur aus Leerraum bestehen.
func ohneLeere(werte []string) []string {
	behalten := make([]string, 0, len(werte))

	for _, wert := range werte {
		if strings.TrimSpace(wert) != "" {
			behalten = append(behalten, wert)
		}
	}

	return behalten
}

// werteAus macht aus einem rohen Wert eine Liste.
func werteAus(roh json.RawMessage) ([]string, error) {
	var liste []string

	if fehler := json.Unmarshal(roh, &liste); fehler == nil {
		return liste, nil
	}

	var einer string

	if fehler := json.Unmarshal(roh, &einer); fehler != nil {
		return nil, fehler
	}

	return []string{einer}, nil
}

// zuFeldwerten macht aus der Karte die Liste für das Fenster.
//
// **Sortiert** — eine Karte hat keine Reihenfolge, und eine Oberfläche, deren
// Felder bei jedem Ruf anders stehen, ist nicht zu bedienen.
func zuFeldwerten(felder map[string][]string) []KIFeldwertDTO {
	namen := make([]string, 0, len(felder))
	for name := range felder {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	werte := make([]KIFeldwertDTO, 0, len(namen))
	for _, name := range namen {
		werte = append(werte, KIFeldwertDTO{Feld: name, Werte: felder[name]})
	}

	return werte
}

// ohneBeanstandete gibt die Felder, zu denen kein Befund steht.
func ohneBeanstandete(
	felder map[string][]string, befunde []domain.Befund,
) map[string][]string {
	beanstandet := map[string]bool{}
	for _, befund := range befunde {
		beanstandet[befund.Datei] = true
	}

	gut := map[string][]string{}

	for name, werte := range felder {
		if !beanstandet[name] {
			gut[name] = werte
		}
	}

	return gut
}

// feldnamenAus nennt die Felder eines Vorschlags.
func feldnamenAus(felder []KIFeldwertDTO) []string {
	namen := make([]string, 0, len(felder))
	for _, feld := range felder {
		namen = append(namen, feld.Feld)
	}

	return namen
}

// mitBisherigem hängt an jeden Vorschlag den Wert, der heute im Feld steht.
//
// **Bestenfalls-Arbeit** (§7, Regel 3): lässt sich die Karte nicht lesen,
// bleibt der Vergleich leer — und der Vorschlag steht trotzdem da. Ein
// Vorschlag ohne Vergleich ist weniger wert als einer mit; gar keiner ist
// schlechter als beides.
func (a *App) mitBisherigem(
	ctx context.Context, karteID int64, felder []KIFeldwertDTO,
) []KIFeldwertDTO {
	if a.kartenleser == nil {
		return felder
	}

	zeile, fehler := a.kartenleser.KarteLesen(ctx, karteID)
	if fehler != nil {
		return felder
	}

	bisher := make(map[string]string, len(zeile.Values))
	for _, wert := range zeile.Values {
		bisher[wert.Field] = wert.Text
	}

	for i := range felder {
		felder[i].Bisher = bisher[felder[i].Feld]
	}

	return felder
}
