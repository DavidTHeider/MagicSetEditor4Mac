// Diese Datei sichert den Bestand aus dem Fenster heraus.
//
// **Eine eigene Frage**: nicht, wie ein Set in eine Datei kommt, sondern wie
// die Datenbank selbst weggetragen wird — samt allem, was in ihr steht.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Bestandssicherer schreibt eine Sicherung des ganzen Bestands.
//
// **Die Schnittstelle steht im Dienst, weil der Anwendungsfall sie braucht**
// (§2) — hier steht nur der Name, unter dem der Composition Root sie setzt.
type Bestandssicherer = service.Bestandssicherer

// MitBestandssicherer setzt die Naht — für Tests und den Composition Root.
func (a *App) MitBestandssicherer(sicherer Bestandssicherer) {
	a.bestandssicherer = sicherer
}

// Sicherungsantwort trägt den Ort der Sicherung — oder einen Fehler.
//
// **Der Pfad kommt zurück.** Wer sichert, will wissen, wohin; die Meldung im
// Fenster soll ihn nennen können, ohne ihn zu erraten. Bei einem Abbruch ist er
// leer, und das ist die Unterscheidung zwischen „nichts getan" und „getan".
type Sicherungsantwort struct {
	Pfad   string  `json:"pfad"`
	Fehler *Fehler `json:"fehler"`
}

// BestandSichern fragt nach einem Ort und schreibt die Sicherung dorthin.
//
// # Warum der Dialog hier steht und nicht im Fenster
//
// Es ist der native Speichern-Dialog (§6): macOS-Eigenheiten werden bedient,
// nicht nachgeahmt. Er kennt zuletzt benutzte Orte, Favoriten und die Suche —
// eine nachgebaute Auswahl im Fenster wäre eine schlechtere Kopie.
//
// # Warum ein Abbruch kein Fehler ist
//
// Wails gibt für „Abbrechen" eine leere Zeichenkette **ohne** Fehler. Wer
// daraus einen Fehler machte, meldete jemandem etwas, das er selbst gewollt
// hat.
//
// # Und warum das Kennwort vor dem Dialog geprüft wird
//
// **Ein zu kurzes Kennwort gehört ans Feld** (§6) und nicht hinter einen
// Dialog, den jemand erst bestätigt, um dann zu erfahren, dass es nichts wird
// (OPS-06.AK-02).
func (a *App) BestandSichern(ctx context.Context, kennwort string, klartext bool) Sicherungsantwort {
	ctx = a.imProfil(ctx)

	wunsch := service.Sicherungswunsch{Flaeche: domain.FlaecheFenster, Kennwort: kennwort, Klartext: klartext}
	if fehler := a.bestandsweg().SicherungswunschPruefen(wunsch); fehler != nil {
		return Sicherungsantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	ctx = service.MitSicherungswunsch(ctx, wunsch)

	texte := DialogtexteFuer(a.sprache)

	ziel, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Speichern: true,
		Ablage:    AblageSicherung,
		Titel:     texte.BestandSpeichern,
		Vorschlag: texte.BestandVorschlag,
		Filter:    texte.BestandFilter,
		Muster:    "*" + service.Sicherungsendung,
	})
	if fehler != nil {
		return Sicherungsantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if ziel == "" {
		return Sicherungsantwort{}
	}

	if fehler := a.bestandsweg().Sichern(ctx, ziel); fehler != nil {
		return Sicherungsantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return Sicherungsantwort{Pfad: ziel}
}
