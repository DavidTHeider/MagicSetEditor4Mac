// **Der Ordner, in dem ein Dateidialog aufgeht** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 8, und seit OPS-02
// (#1719)
// vier davon.
//
// # Warum das an einer Stelle steht und nicht an sechzehn
//
// **Weil es sechzehn Dialoge gibt und eine Frage.** Ein Bild, ein Set, eine
// Sicherung, MTG JSON, eine Sammlungsdatei — sie alle gehen dorthin auf, wo
// jemand seine Sachen liegen hat. Den Pfad an jeder Rufstelle zu holen hieße,
// ihn sechzehnmal zu vergessen zu können.
//
// **Deshalb ein Mantel um die Naht** — und zwar als **Methode** und nicht als
// ausgetauschtes Feld. Der erste Anlauf ersetzte `a.dateiwahl` beim Start; der
// nächste Ruf von [App.MitDateiwahl] zog den Mantel wieder aus, und kein
// Zweig sagte es. Eine Naht, die man setzen kann, muss man auch setzen
// dürfen.
//
// # Welcher Ordner
//
// **Beim Öffnen der fürs Hereinholen, beim Sichern der fürs Hinausgeben** —
// es sei denn, die Handlung nennt ihre Ablage: Sicherungen und Diagnose
// haben eigene. Bis OPS-02 ging ein Sicherndialog auf, wo das System ihn
// öffnete; „ein Druckbogen gehört neben sein Set" stand hier als Grund, und
// keiner der vierzehn Sicherndialoge setzte seinen Ort selbst.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// waehleDatei öffnet den Dialog — im eingestellten Ordner.
//
// **Ein Fehler beim Lesen ist kein Grund, keinen Dialog zu öffnen.** Dann geht
// er dort auf, wo das Betriebssystem ihn öffnet — genau wie ohne Einstellung
// (§8: der Ausfall einer Bequemlichkeit hält die Arbeit nicht an).
//
// **Ein Ordner, den es nicht mehr gibt, ist etwas anderes** (AK-03): beim
// Sichern hält er an, mit einem Satz, der das Laufwerk nennt — sonst landete
// eine Sicherung still an einem anderen Ort. Beim Öffnen nicht: dort wird
// nichts gespeichert, und wer ein Laufwerk abgezogen hat, holt seine Datei
// vielleicht von woanders.
func (a *App) waehleDatei(ctx context.Context, auswahl Dateiauswahl) (string, error) {
	if angabe := a.vorgabeordner(ctx, ordnerschluessel(auswahl)...); angabe != "" {
		ordner, fehler := geprueft(angabe)
		if fehler != nil && auswahl.Speichern {
			return "", fehler
		}

		auswahl.Startverzeichnis = ordner
	}

	return a.dateiwahl(ctx, auswahl)
}

// ordnerschluessel sagt, welche Einstellungen für einen Dialog gelten — die
// erste, die etwas nennt.
//
// **Die Bildausgabe hat zwei** (AP-10, #2367): den zuletzt gewählten Ordner
// wie MSE2s `default_export_dir`, und ohne ihn den fürs Hinausgeben.
func ordnerschluessel(auswahl Dateiauswahl) []string {
	switch {
	case auswahl.Ablage == AblageSicherung:
		return []string{domain.VorgabeSicherungspfad}
	case auswahl.Ablage == AblageDiagnose:
		return []string{domain.VorgabeDiagnosepfad}
	case auswahl.Ablage == AblageBilder:
		return []string{domain.EinstellungBildordner, domain.VorgabeAusfuhrpfad}
	case auswahl.Speichern:
		return []string{domain.VorgabeAusfuhrpfad}
	default:
		return []string{domain.VorgabeHochladepfad}
	}
}

// vorgabeordner gibt die erste Angabe dieser Ordnereinstellungen — leer, wenn
// keine dasteht oder sich nichts lesen lässt.
func (a *App) vorgabeordner(ctx context.Context, schluessel ...string) string {
	if a.einstellungen == nil {
		return ""
	}

	werte, fehler := a.einstellungen.Alle(a.imProfil(ctx))
	if fehler != nil {
		return ""
	}

	for _, einer := range schluessel {
		if werte[einer] != "" {
			return werte[einer]
		}
	}

	return ""
}
