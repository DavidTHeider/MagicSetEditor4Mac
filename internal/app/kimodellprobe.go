package app

// **Ein Modell messen, statt seinem Anbieter zu glauben** — Werkstück A9 von
// Phase 5.3
// (#1448).
//
// # Warum es diese Bindung gibt
//
// **Weil OpenAI seine Modelle nicht beschreibt** und ohne einen dritten Weg
// ein Schalter wäre, der nichts tut
// ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)).
//
// # Und warum sie Geld kostet
//
// **Weil eine Messung eine Anfrage ist.** Sie geht durch **dieselbe** Sperre
// wie ein Vorschlag: Schalter, »nur lokal«, Journal und Freigabe. Wer sie
// daran vorbeiführte, hätte einen kostenpflichtigen Ruf ohne Buchung — und
// genau das schließt A8 aus.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KIProbeDTO ist eine Messung, wie das Fenster sie sieht.
type KIProbeDTO struct {
	Anbieter   string `json:"anbieter"`
	Alias      string `json:"alias"`
	Faehigkeit string `json:"faehigkeit"`
	Datum      string `json:"datum"`
	Beleg      string `json:"beleg"`

	// Modell ist, was wirklich geantwortet hat — nicht der Alias.
	Modell string `json:"modell"`
}

// KIProbeAntwort trägt die Messungen samt Fehler.
type KIProbeAntwort struct {
	Proben []KIProbeDTO `json:"proben"`

	Fehler *Fehler `json:"fehler"`
}

// KIModellPruefen misst ein Modell und merkt sich das Ergebnis.
//
// **Eine geglückte Messung belegt zwei Fähigkeiten** — Text und feste Form —,
// eine gescheiterte schreibt nichts und sagt, woran es lag.
func (a *App) KIModellPruefen(
	ctx context.Context, anbieter, alias, nachweis string,
) KIProbeAntwort {
	ctx = a.imProfil(ctx)

	// **Die Rolle vor dem Modell** — dieselbe Frage wie bei
	// [App.KIVorschlagen] (ID-03, #1702).
	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		return KIProbeAntwort{Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, verweigert)}
	}

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return KIProbeAntwort{Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	// **Nicht über das Auftragswerk, und das ist der Punkt.**
	//
	// `KIAuftragswerk.Anlegen` verlangt, dass das Modell die Fähigkeit
	// **schon belegt** hat — und genau das soll hier erst gemessen werden.
	// Wer die Probe dort hindurchführte, hätte einen Kreis: ohne Beleg keine
	// Messung, ohne Messung kein Beleg.
	staende, _, fehler := a.frageAnbieter(ctx)
	if fehler != nil {
		return KIProbeAntwort{Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	quelle, adresse, fehler := a.quelleFuer(ctx, staende, anbieter)
	if fehler != nil {
		return KIProbeAntwort{Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	// **Die Buchung zuerst, wie bei jedem Auftrag** — eine Messung ist eine
	// Anfrage und keine Ausnahme (A8): ohne Journal und ohne die Freigabe
	// eines Menschen geht sie nicht hinaus.
	budget, budgetfehler := a.budgetFuer(ctx, anbieter, adresse,
		kideckung{nachweis: nachweis})
	if budgetfehler != nil {
		return KIProbeAntwort{
			Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, budgetfehler),
		}
	}

	posten, buchfehler := budget.Reservieren(ctx, "probe/"+anbieter+"/"+alias)
	if buchfehler != nil {
		return KIProbeAntwort{
			Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, buchfehler),
		}
	}

	// **Und er wird geschlossen, gleich wie es ausgeht** (#1542): eine
	// Messung ist eine Anfrage, und ihre Reservierung band bis zum
	// 2026-09-13 für immer Budget.
	defer a.kiPostenSchliessen(ctx, posten)

	ablage, ablagefehler := a.probenablage()
	if ablagefehler != nil {
		return KIProbeAntwort{
			Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, ablagefehler),
		}
	}

	proben, fehler := service.NeuesKIModellpruefwerk(ablage, a.uhr).Pruefen(
		ctx, profil, probequelle{quelle: quelle}, anbieter, alias)
	if fehler != nil {
		return KIProbeAntwort{Proben: []KIProbeDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return KIProbeAntwort{Proben: zuProbenDTO(proben)}
}

// KIModellVergessen nimmt die Messungen eines Modells zurück.
//
// **Weil eine Messung veralten kann, ohne falsch zu werden.** Wer neu prüfen
// will, soll die alte Auskunft wegnehmen können, statt sie stillschweigend
// überschrieben zu bekommen.
func (a *App) KIModellVergessen(
	ctx context.Context, anbieter, alias string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	ablage, fehler := a.probenablage()
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if fehler := ablage.Vergessen(ctx, profil, anbieter, alias); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// probequelle bringt die Laufzeitquelle auf die Gestalt des Prüfwerks.
//
// **Eine Hülle und keine zweite Fachlogik**: `service` soll `app` nicht
// kennen, und `KIRohantwort` liegt hier.
type probequelle struct{ quelle KILaufzeitquelle }

func (p probequelle) Fassung(ctx context.Context) (string, error) {
	return p.quelle.Fassung(ctx)
}

func (p probequelle) Vorschlagen(
	ctx context.Context, alias, auftrag string, schema map[string]any,
) (service.KIProbeantwort, error) {
	// **Die Kartenanweisung**, denn eine Messung stellt eine Kartenfrage:
	// gemessen wird, ob das Modell eine feste Form hält, und die Form ist
	// die eines Kartenvorschlags.
	roh, fehler := p.quelle.Vorschlagen(ctx, KIKartenanweisung, alias, auftrag, schema)
	if fehler != nil {
		return service.KIProbeantwort{}, fehler
	}

	return service.KIProbeantwort{
		Inhalt: roh.Inhalt, Abgeschnitten: roh.Abgeschnitten, Modell: roh.Modell,
	}, nil
}

// probenablage gibt die Ablage über dem offenen Bestand.
func (a *App) probenablage() (*storage.Modellprobenablage, error) {
	db := a.laufendeVerbindung()
	if db == nil {
		return nil, fehlenderBestand()
	}

	return storage.NeueModellprobenablage(db), nil
}

// zuProbenDTO macht aus den Messungen ihre DTOs.
func zuProbenDTO(proben []domain.KIModellprobe) []KIProbeDTO {
	alle := make([]KIProbeDTO, 0, len(proben))

	for _, probe := range proben {
		alle = append(alle, KIProbeDTO{
			Anbieter: probe.Anbieter, Alias: probe.Alias,
			Faehigkeit: string(probe.Faehigkeit), Datum: probe.Datum,
			Beleg: probe.Beleg, Modell: probe.Modell,
		})
	}

	return alle
}
