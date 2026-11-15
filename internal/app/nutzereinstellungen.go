// **Die Einstellungen, die dem Menschen gehören, an allen vier Flächen** —
// Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 9 und 11 der Kommentierung vom 2026-09-05.
//
// # Warum sie neben den Anwendungseinstellungen stehen und nicht in ihnen
//
// **Weil sie einer anderen Sache gehören.** `Einstellungen` gilt für das
// **Profil**, in dem gearbeitet wird — Sprache, Erscheinungsbild,
// Schriftgröße. Was hier steht, gilt für den **Menschen**: er nimmt es in
// jedes seiner Profile mit. Beides in eine Antwort zu werfen hieße, dem
// Fenster die Zuordnung zu überlassen, und dann stünde sie an zwei Stellen.
//
// # Und warum das Setzen die ganze Liste zurückgibt
//
// **§7, Regel 2.** Wer nach dem Setzen noch einmal läse, hätte zwei Antworten
// und zwei Stände — dieselbe Bauform wie bei den Nutzern und ihren Profilen.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nutzereinstellungsfuehrer ist die Naht zu `internal/service`.
type Nutzereinstellungsfuehrer interface {
	Alle(ctx context.Context, nutzerID int64) (map[string]string, error)
	Setzen(ctx context.Context, nutzerID int64, schluessel, wert string) error
}

// NutzereinstellungDTO ist eine Einstellung samt ihrer möglichen Werte.
//
// **Die Werte kommen mit** — dieselbe Begründung wie bei [EinstellungDTO]:
// ohne sie baute das Fenster die Auswahlliste selbst, und die erlaubten Werte
// stünden an zwei Stellen, von denen die im Fenster die ungeprüfte wäre.
type NutzereinstellungDTO struct {
	Schluessel string   `json:"schluessel"`
	Wert       string   `json:"wert"`
	Werte      []string `json:"werte"`
}

// NutzereinstellungenAntwort trägt die Einstellungen eines Menschen.
type NutzereinstellungenAntwort struct {
	Einstellungen []NutzereinstellungDTO `json:"einstellungen"`
	Fehler        *Fehler                `json:"fehler"`
}

// MitNutzereinstellungen setzt die Naht.
func (a *App) MitNutzereinstellungen(fuehrer Nutzereinstellungsfuehrer) {
	a.nutzereinstellungen = fuehrer
}

// Nutzereinstellungen gibt die Einstellungen eines Menschen.
func (a *App) Nutzereinstellungen(
	ctx context.Context, nutzerID string,
) NutzereinstellungenAntwort {
	ctx = a.imProfil(ctx)

	// **Ein Token liest keines Menschen Einstellungen** (QA-09, #1740) —
	// dieselbe Grenze wie bei den Konten: eine Maschine ist kein Gast.
	if fehler := keinTokenAmMenschen(ctx); fehler != nil {
		return NutzereinstellungenAntwort{
			Einstellungen: []NutzereinstellungDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return NutzereinstellungenAntwort{
			Einstellungen: []NutzereinstellungDTO{}, Fehler: befund,
		}
	}

	return a.nutzereinstellungenNach(ctx, kennung)
}

// NutzereinstellungSetzen schreibt eine fort und gibt alle zurück.
func (a *App) NutzereinstellungSetzen(
	ctx context.Context, nutzerID, schluessel, wert string,
) NutzereinstellungenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return NutzereinstellungenAntwort{
			Einstellungen: []NutzereinstellungDTO{}, Fehler: befund,
		}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return NutzereinstellungenAntwort{
			Einstellungen: []NutzereinstellungDTO{}, Fehler: Uebersetze(ctx, fremd),
		}
	}

	if fehler := a.nutzereinstellungen.Setzen(
		ctx, kennung, schluessel, wert); fehler != nil {
		return NutzereinstellungenAntwort{
			Einstellungen: []NutzereinstellungDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return a.nutzereinstellungenNach(ctx, kennung)
}

// nutzereinstellungenNach liest sie und baut die Antwort.
func (a *App) nutzereinstellungenNach(
	ctx context.Context, nutzerID int64,
) NutzereinstellungenAntwort {
	werte, fehler := a.nutzereinstellungen.Alle(ctx, nutzerID)
	if fehler != nil {
		return NutzereinstellungenAntwort{
			Einstellungen: []NutzereinstellungDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	// **Die Reihe kommt aus der Domäne**, nicht von hier: sonst entschiede
	// die Fassade, welche Einstellungen es gibt.
	gelesen := make([]NutzereinstellungDTO, 0, len(domain.Nutzereinstellungsreihe()))

	for _, schluessel := range domain.Nutzereinstellungsreihe() {
		gelesen = append(gelesen, NutzereinstellungDTO{
			Schluessel: schluessel,
			Wert:       werte[schluessel],
			Werte:      domain.ErlaubteNutzerwerte(schluessel),
		})
	}

	return NutzereinstellungenAntwort{Einstellungen: gelesen}
}
