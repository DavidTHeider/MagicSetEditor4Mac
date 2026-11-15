package service

// **Der Vertragslauf für Einbettung und Neuordnung** — Werkstück A6 von
// Phase 5.3 (#1445).
//
// # Wofür es ihn gibt
//
// **Damit ein Anbieter geprüft werden kann, bevor jemand ihm glaubt.** Das
// Issue verlangt einen „unabhängigen Embedding-/Rerank-Contract-Test **ohne
// Generierungsprovider**" — also einen Lauf, der nur die beiden Fähigkeiten
// braucht und nichts sonst. Wer in A7 oder später eine Laufzeit anbindet,
// fährt ihn und bekommt eine Liste von Befunden statt eines Gefühls.
//
// # Befund und Fehler sind zwei Dinge
//
// **Ein Befund sagt: der Anbieter hält seinen Vertrag nicht.** Ein Fehler
// sagt: es kam gar nichts. Wer beides gleich behandelte, ließe einen
// ausgefallenen Dienst wie einen vertragsbrüchigen aussehen — und umgekehrt.

import (
	"context"
	"fmt"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Einbettungsquelle ist, was der Vertragslauf von einem Anbieter braucht.
//
// **Eine Methode.** Mehr ist für eine Einbettung nicht nötig, und mehr zu
// verlangen hieße, dass ein reiner Einbetter hier nicht geprüft werden kann.
type Einbettungsquelle interface {
	Einbetten(ctx context.Context, texte []string) ([][]float32, error)
}

// Neuordnungsquelle ist dasselbe für eine Neuordnung.
type Neuordnungsquelle interface {
	Neuordnen(ctx context.Context, frage string, kandidaten []string) (
		[]domain.KIBewertung, error)
}

// dieProbetexte sind die Texte, mit denen geprüft wird.
//
// **Zwei verschiedene**, damit ein Anbieter, der immer denselben Vektor gibt,
// auffällt.
var dieProbetexte = []string{"Blitzschlag", "Gegenzauber"}

// PruefeEinbettung fährt den Vertragslauf gegen eine Einbettungsquelle.
//
// **Der Vertrag wird zuerst selbst geprüft**: ein Lauf gegen einen Vertrag,
// der keine Dimension nennt, prüfte nichts und wäre trotzdem grün.
func PruefeEinbettung(
	ctx context.Context, quelle Einbettungsquelle, vertrag domain.KIEinbettungsvertrag,
) ([]domain.Befund, error) {
	if fehler := vertrag.Pruefe(); fehler != nil {
		return nil, fehler
	}

	fragen := make([]string, 0, len(dieProbetexte))
	dokumente := make([]string, 0, len(dieProbetexte))

	for _, text := range dieProbetexte {
		fragen = append(fragen, vertrag.FuerFrage(text))
		dokumente = append(dokumente, vertrag.FuerDokument(text))
	}

	erste, fehler := quelle.Einbetten(ctx, fragen)
	if fehler != nil {
		return nil, fehler
	}

	zweite, fehler := quelle.Einbetten(ctx, fragen)
	if fehler != nil {
		return nil, fehler
	}

	alsDokument, fehler := quelle.Einbetten(ctx, dokumente)
	if fehler != nil {
		return nil, fehler
	}

	befunde := formPruefen(vertrag, erste)
	befunde = append(befunde, wiederholbarkeitPruefen(erste, zweite)...)

	return append(befunde, anweisungPruefen(vertrag, erste, alsDokument)...), nil
}

// formPruefen prüft Zahl, Länge, Zahlen und Normalisierung.
func formPruefen(
	vertrag domain.KIEinbettungsvertrag, vektoren [][]float32,
) []domain.Befund {
	var befunde []domain.Befund

	if len(vektoren) != len(dieProbetexte) {
		return []domain.Befund{{
			Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Auf %d Texte kamen %d Vektoren zurück.",
				len(dieProbetexte), len(vektoren)),
		}}
	}

	for stelle, vektor := range vektoren {
		if fehler := vertrag.PruefeVektor(vektor); fehler != nil {
			befunde = append(befunde, domain.Befund{
				Datei: dieProbetexte[stelle], Art: domain.ArtUmgebung,
				Meldung: fehler.Error(),
			})

			continue
		}

		if !vertrag.Normalisiert {
			continue
		}

		if betrag := betragVon(vektor); math.Abs(betrag-1) > 1e-4 {
			befunde = append(befunde, domain.Befund{
				Datei: dieProbetexte[stelle], Art: domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"Der Vertrag sagt normalisierte Vektoren zu, dieser hat "+
						"die Länge %.4f.", betrag),
			})
		}
	}

	return befunde
}

// wiederholbarkeitPruefen hält zwei Läufe desselben Textes gegeneinander.
//
// **Eine Einbettung ist keine Erzeugung.** Derselbe Text muss denselben
// Vektor geben — sonst zeigt ein Index morgen woandershin als heute, ohne
// dass jemand etwas geändert hat.
func wiederholbarkeitPruefen(erste, zweite [][]float32) []domain.Befund {
	if len(erste) != len(zweite) {
		return []domain.Befund{{
			Art:     domain.ArtUmgebung,
			Meldung: "Zweimal dieselbe Anfrage gab verschieden viele Vektoren.",
		}}
	}

	for stelle := range erste {
		if len(erste[stelle]) != len(zweite[stelle]) {
			continue
		}

		for i := range erste[stelle] {
			if erste[stelle][i] != zweite[stelle][i] {
				return []domain.Befund{{
					Datei: dieProbetexte[stelle], Art: domain.ArtUmgebung,
					Meldung: "Zweimal derselbe Text gab zwei verschiedene " +
						"Vektoren — ein Index darauf zeigte morgen woandershin.",
				}}
			}
		}
	}

	return nil
}

// anweisungPruefen sieht nach, ob Frage und Dokument auseinandergehen.
//
// **Nur wo der Vertrag es zusagt**: nennt er keine verschiedenen Anweisungen,
// dürfen die Vektoren gleich sein — dann ist es kein Befund, sondern die
// Abwesenheit einer Zusage.
func anweisungPruefen(
	vertrag domain.KIEinbettungsvertrag, alsFrage, alsDokument [][]float32,
) []domain.Befund {
	if vertrag.Frageanweisung == vertrag.Dokumentanweisung {
		return nil
	}

	if len(alsFrage) != len(alsDokument) {
		return nil
	}

	for stelle := range alsFrage {
		if !gleich(alsFrage[stelle], alsDokument[stelle]) {
			return nil
		}
	}

	return []domain.Befund{{
		Art: domain.ArtUmgebung,
		Meldung: "Der Vertrag nennt für Frage und Dokument verschiedene " +
			"Anweisungen, und die Quelle gibt beide Male denselben Vektor — " +
			"die Anweisung wirkt nicht.",
	}}
}

// PruefeNeuordnung fährt den Vertragslauf gegen eine Neuordnungsquelle.
func PruefeNeuordnung(
	ctx context.Context, quelle Neuordnungsquelle, frage string, kandidaten []string,
) ([]domain.Befund, error) {
	bewertungen, fehler := quelle.Neuordnen(ctx, frage, kandidaten)
	if fehler != nil {
		return nil, fehler
	}

	// **Ein Formfehler ist ein Befund und kein Fehler.** Der Anbieter hat
	// geantwortet, seine Antwort passt nur nicht zu dem, wonach gefragt wurde;
	// ein Fehler hieße „es kam nichts", und das wäre eine andere Auskunft.
	//
	// **Deshalb wird er umgewandelt und nicht durchgereicht** — so steht
	// nirgends ein `nil` neben einer Ursache, und niemand muss raten, ob das
	// Absicht war.
	if befunde := formfehlerAlsBefund(bewertungen, len(kandidaten)); befunde != nil {
		return befunde, nil
	}

	bekannt := map[string]bool{}
	for _, kandidat := range kandidaten {
		bekannt[kandidat] = true
	}

	var befunde []domain.Befund

	for _, bewertung := range bewertungen {
		if !bekannt[bewertung.Kandidat] {
			befunde = append(befunde, domain.Befund{
				Datei: bewertung.Kandidat, Art: domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"Bewertet wurde %q — dieser Kandidat war nicht dabei.",
					bewertung.Kandidat),
			})
		}
	}

	return befunde, nil
}

// formfehlerAlsBefund macht aus einem Formfehler einen Befund — oder nil.
func formfehlerAlsBefund(bewertungen []domain.KIBewertung, kandidaten int) []domain.Befund {
	fehler := domain.PruefeBewertungen(bewertungen, kandidaten)
	if fehler == nil {
		return nil
	}

	return []domain.Befund{{Art: domain.ArtUmgebung, Meldung: fehler.Error()}}
}

// betragVon gibt die Länge eines Vektors.
func betragVon(vektor []float32) float64 {
	quadrat := float64(0)

	for _, wert := range vektor {
		quadrat += float64(wert) * float64(wert)
	}

	return math.Sqrt(quadrat)
}

// gleich sagt, ob zwei Vektoren Stelle für Stelle gleich sind.
func gleich(einer, anderer []float32) bool {
	if len(einer) != len(anderer) {
		return false
	}

	for i := range einer {
		if einer[i] != anderer[i] {
			return false
		}
	}

	return true
}
