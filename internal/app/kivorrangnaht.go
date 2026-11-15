package app

// **Vorrang und Rückfall** — AI-05.AK-05 und AK-06
// (#2042).
//
// # Wer entscheidet
//
// **Die Regel steht in `domain`** ([domain.KIReihenfolge]): eine kleine
// Textaufgabe fragt zuerst das Gerätemodell, dann die lokalen Anbieter, nie
// von selbst die Cloud. **Hier steht nur, woher die Lage kommt** — aus dem
// Katalog, wie bei jedem anderen Vorschlag — und dass der nächste drankommt,
// wenn der vorige nicht geliefert hat ([App.vorschlagVersuch]).
//
// # Was das Gerät dafür fragt
//
// **Ob es eingerichtet ist, nur, wenn es eingeschaltet ist** — und das ist
// die Erkennung über `fm`, für eine Minute gemerkt. Ein abgeschaltetes Gerät
// startet keinen Prozess, auch keinen zum Nachsehen (AK-03, AK-04).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vorschlagGewaehlt fragt den genannten Anbieter — und bietet nach einer
// Ablehnung die Rückfälle an.
func (a *App) vorschlagGewaehlt(ctx context.Context, frage kifrage, bezug KIVorschlagsbezug) KIVorschlagAntwort {
	staende, anbieter, fehler := a.frageAnbieter(ctx)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	werk, quelle, fehler := a.werkzeugAus(ctx, staende, frage.anbieter, kideckung{nachweis: bezug.Nachweis})
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	antwort := a.vorschlagMit(ctx, kiwerkzeugsatz{werk: werk, quelle: quelle}, frage)
	antwort.Weg = string(domain.KIWegGewaehlt)

	if antwort.Ausgang == KIAusgangAbgelehnt {
		antwort.Rueckfaelle = domain.KIRueckfaelle(frage.anbieter,
			domain.IstKleineTextaufgabe(bezug.Nur, bezug.Auftrag, bezug.Belegfrage),
			anbieterlagen(anbieter))
	}

	return antwort
}

// vorschlagMitVorrang fragt der Reihe nach, bis einer liefert.
func (a *App) vorschlagMitVorrang(ctx context.Context, frage kifrage, bezug KIVorschlagsbezug) KIVorschlagAntwort {
	if fehler := domain.PruefeVorrangwunsch(bezug.Alias); fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	staende, anbieter, fehler := a.frageAnbieter(ctx)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	reihe := domain.KIReihenfolge(domain.KIVorrang,
		domain.IstKleineTextaufgabe(bezug.Nur, bezug.Auftrag, bezug.Belegfrage),
		anbieterlagen(anbieter))
	if len(reihe) == 0 {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: kein lokaler Anbieter ist bereit — das Gerätemodell oder Ollama "+
				"einschalten, oder einen Anbieter ausdrücklich wählen", domain.ErrValidation))}
	}

	var letzte KIVorschlagAntwort

	for _, versuch := range reihe {
		antwort, weiter := a.vorrangVersuch(ctx, staende, frage, bezug, versuch.Anbieter)
		antwort.Weg = string(versuch.Weg)

		if !weiter {
			return antwort
		}

		letzte = antwort
	}

	return letzte
}

// vorrangVersuch fragt einen Anbieter der Reihe mit seinem ersten Modell,
// das strukturiert antworten kann.
func (a *App) vorrangVersuch(
	ctx context.Context, staende []anbieterstand, frage kifrage, bezug KIVorschlagsbezug, name string,
) (KIVorschlagAntwort, bool) {
	alias, wahlfehler := modellFuer(staende, name, a.vorrangmodellwahl(ctx))
	if wahlfehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, wahlfehler)}, true
	}

	if alias == "" {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: %s nennt kein Modell, das strukturiert antworten kann", domain.ErrValidation, name))}, true
	}

	werk, quelle, fehler := a.werkzeugAus(ctx, staende, name, kideckung{nachweis: bezug.Nachweis})
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}, true
	}

	frage.anbieter, frage.alias = name, alias

	return a.vorschlagVersuch(ctx, kiwerkzeugsatz{werk: werk, quelle: quelle}, frage)
}

// anbieterlagen gibt die Lage jedes Anbieters für die Regel.
func anbieterlagen(anbieter []kianbieter) []domain.KIAnbieterlage {
	lagen := make([]domain.KIAnbieterlage, 0, len(anbieter))

	for _, einer := range anbieter {
		lage := domain.KIAnbieterlage{
			Name: einer.name, Ort: ortVon(einer.name, einer.adresse),
			Bereit: einer.eingeschaltet && einer.bauen != nil,
		}

		// **Eingeschaltet heißt beim Gerät schon eingerichtet**
		// ([App.geraeteanbieterAus]).
		lage.Geraet = einer.name == geraeteanbieter

		lagen = append(lagen, lage)
	}

	return lagen
}
