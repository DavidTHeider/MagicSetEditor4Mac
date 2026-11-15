package app

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// OrteAntwort nennt die vorgeschlagenen MSE2-Verzeichnisse.
//
// **Nur die, die es gibt.** Ein Vorschlag auf ein Verzeichnis, das nicht
// existiert, wäre eine Aufforderung, es zu suchen — und genau das soll die
// Anwendung den Nutzenden abnehmen.
type OrteAntwort struct {
	Orte []string `json:"orte"`
}

// UebernahmeAntwort ist der Ausgang einer Übernahme.
type UebernahmeAntwort struct {
	Uebernommen []PaketDTO `json:"uebernommen"`

	// Uebergangen sind die Pakete, die schon dalagen.
	Uebergangen []string `json:"uebergangen"`
	Befunde     []string `json:"befunde"`

	// Abgebrochen sagt, dass niemand ein Verzeichnis gewählt hat.
	Abgebrochen bool    `json:"abgebrochen"`
	Fehler      *Fehler `json:"fehler"`
}

// MSE2Vorschlaege gibt die üblichen Orte, an denen MSE2 seine Vorlagen hat.
//
// **Die leere Liste wird ausdrücklich gebaut** — und zwar hier, nicht beim
// Lieferanten.
//
// `nil` ginge als `null` durch die Serialisierung, und `Uebernahme.tsx` prüft
// `daten !== undefined && daten.length > 0`: bei `null` bricht die Seite. Die
// Naht ist der Ort dafür, weil die Zusage „eine Liste kommt als Liste" der
// Fassade gehört und nicht dem, der sie gerade füllt.
//
// **Heute ist das Vorsorge, kein behobener Fehler**, und das gehört
// dazugesagt: `MSE2Orte` gibt in beiden Zweigen eine gebaute Liste zurück, nie
// `nil`. Der Vertragstest aus
// #64 hat die
// Stelle gefunden, weil er die Naht mit einem `nil`-Lieferanten befährt — und
// genau ein solcher Lieferant ist der nächste, den jemand einsetzt.
func (a *App) MSE2Vorschlaege() OrteAntwort {
	orte := a.mse2orte()
	if orte == nil {
		orte = []string{}
	}

	return OrteAntwort{Orte: orte}
}

// VorlagenUebernehmen holt die Vorlagen aus einem Verzeichnis.
func (a *App) VorlagenUebernehmen(ctx context.Context, ort string) UebernahmeAntwort {
	ctx = a.imProfil(ctx)

	leer := UebernahmeAntwort{
		Uebernommen: []PaketDTO{},
		Uebergangen: []string{},
		Befunde:     []string{},
	}

	if strings.TrimSpace(ort) == "" {
		leer.Abgebrochen = true

		return leer
	}

	ergebnis, err := a.pakete.Uebernehmen(ctx, ort)

	return a.uebernahmeAntwort(ctx, leer, ergebnis, err)
}

// uebernahmeAntwort trägt ein Ergebnis in die Antwort — **auch neben einem
// Fehler**: was übernommen ist, bleibt übernommen (§7, Regel 3), und die
// Antwort sagt beides. Dieselbe Gestalt für die Übernahme aus MSE2 und für
// einen Installer (FMT-02.AK-04, #1691).
func (a *App) uebernahmeAntwort(ctx context.Context, antwort UebernahmeAntwort, ergebnis *service.Uebernahme, err error) UebernahmeAntwort {
	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)
	}

	if ergebnis == nil {
		return antwort
	}

	// **Mit der Zeile der Übersicht** — wie jedes geschriebene Paket (C01).
	alsDTO := a.uebersichtFuer(ctx)
	for _, paket := range ergebnis.Uebernommen {
		antwort.Uebernommen = append(antwort.Uebernommen, alsDTO(paket))
	}

	antwort.Uebergangen = append(antwort.Uebergangen, ergebnis.Uebergangen...)

	for _, befund := range ergebnis.Befunde {
		antwort.Befunde = append(antwort.Befunde, befund.String())
	}

	return antwort
}

// MitMSE2Orten setzt die Naht zu den Vorschlägen.
//
// Als Naht und aus demselben Grund wie `paketort`: ein Test, der gegen die
// echten Verzeichnisse dieses Rechners liefe, wäre auf einer zweiten Maschine
// grün oder rot, je nachdem, wer dort MSE2 installiert hat
// ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)).
func (a *App) MitMSE2Orten(orte func() []string) { a.mse2orte = orte }
