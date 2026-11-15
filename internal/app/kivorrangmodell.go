package app

// **Welches Ollama-Modell der Vorrang fragt, und dass es am Formular steht**
// — #2054, aus AI-02 (#1743). Die Regel steht in
// [domain.KIVorrangmodellAus]; hier wird sie mit den Ständen der Anbieter
// und der Einstellung zusammengeführt.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vorrangmodellwahl liest die Wahl — leer heißt: automatisch, und ein
// unlesbarer Stand ebenso, denn dann gilt, was ohne Wahl gilt.
func (a *App) vorrangmodellwahl(ctx context.Context) string {
	if a.einstellungen == nil {
		return ""
	}

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return ""
	}

	return stand[domain.KIVorrangmodell].Wirksam
}

// modellFuer gibt das Modell eines Anbieters, das der Vorrang fragt.
//
// **Die Wahl gilt nur für Ollama**: das Gerätemodell hat eines, und eine
// Cloud fragt der Vorrang nie von selbst.
func modellFuer(staende []anbieterstand, name, gewaehlt string) (string, error) {
	if name != ollamaanbieter {
		gewaehlt = ""
	}

	for _, stand := range staende {
		if stand.name == name {
			return domain.KIVorrangmodellAus(stand.modelle, gewaehlt)
		}
	}

	return "", nil
}

// vorrangreihe nennt, wen der Vorrang für eine **kleine** Textaufgabe fragt
// — Name und Flavortext am Formular —, als `anbieter/modell`; und was dabei
// nicht stimmt, als Befund.
func vorrangreihe(
	staende []anbieterstand, anbieter []kianbieter, gewaehlt string,
) (reihe []string, befunde []domain.Befund) {
	reihe = []string{}

	for _, versuch := range domain.KIReihenfolge(domain.KIVorrang, true, anbieterlagen(anbieter)) {
		alias, fehler := modellFuer(staende, versuch.Anbieter, gewaehlt)
		if fehler != nil {
			befunde = append(befunde, domain.Befund{Art: domain.ArtUmgebung, Meldung: alsSatz(fehler)})

			continue
		}

		if alias != "" {
			reihe = append(reihe, versuch.Anbieter+"/"+alias)
		}
	}

	return reihe, befunde
}

// alsSatz nimmt einem Fehler der Domäne seine Art vom Anfang — im Katalog
// steht ein Satz, kein „nicht gefunden: …".
func alsSatz(fehler error) string {
	satz := fehler.Error()
	for _, art := range []error{domain.ErrNotFound, domain.ErrValidation} {
		satz = strings.TrimPrefix(satz, art.Error()+": ")
	}

	if satz == "" {
		return satz
	}

	return strings.ToUpper(satz[:1]) + satz[1:] + "."
}
