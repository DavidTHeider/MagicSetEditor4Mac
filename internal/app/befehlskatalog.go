package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// BefehlDTO ist ein Befehl, wie das Fenster ihn zeigt.
type BefehlDTO struct {
	Name        string         `json:"name"`
	Faehigkeit  string         `json:"faehigkeit"`
	CLI         string         `json:"cli"`
	REST        string         `json:"rest"`
	Erklaerung  string         `json:"erklaerung"`
	Schreibend  bool           `json:"schreibend"`
	Idempotent  bool           `json:"idempotent"`
	Gefaehrlich bool           `json:"gefaehrlich"`
	Schema      map[string]any `json:"schema"`
}

// BefehlskatalogAntwort ist der Umschlag fürs Fenster.
type BefehlskatalogAntwort struct {
	Befehle []BefehlDTO `json:"befehle"`
	Fehler  *Fehler     `json:"fehler"`
}

// Befehlskatalog nennt jeden Befehl, sortiert nach Namen.
func (a *App) Befehlskatalog(ctx context.Context) BefehlskatalogAntwort {
	ctx = a.imProfil(ctx)

	if len(a.befehle) == 0 {
		return BefehlskatalogAntwort{
			Befehle: []BefehlDTO{},
			Fehler:  Uebersetze(ctx, fmt.Errorf("%w: die Kommandozentrale hat keine Befehle — keine Quelle beim Start", domain.ErrNotFound)),
		}
	}

	// **Was abgeschaltet ist, bietet die Palette nicht an** (Befund 1 des
	// Gesamtreviews über 0.6.0, #1660) — der Stand wird einmal gelesen,
	// nicht je Befehl.
	offen := a.befehleOffen(ctx)

	befehle := make([]BefehlDTO, 0, len(a.befehle))
	for _, befehl := range a.befehle {
		if !offen(befehl.Name) {
			continue
		}

		verweis := befehlsverweise[befehl.Name]
		befehle = append(befehle, BefehlDTO{
			Name: befehl.Name, Faehigkeit: verweis.Faehigkeit, CLI: verweis.CLI, REST: verweis.REST,
			Erklaerung: befehl.Erklaerung, Schreibend: befehl.Schreibend, Idempotent: befehl.Idempotent,
			Gefaehrlich: gefaehrlich(befehl), Schema: schemaOder(befehl.Schema),
		})
	}

	sort.Slice(befehle, func(i, j int) bool { return befehle[i].Name < befehle[j].Name })

	return BefehlskatalogAntwort{Befehle: befehle}
}

// gefaehrlich sagt, ob ein Befehl vor dem Ruf bestätigt werden muss.
//
// **Schreibend und nicht idempotent** — ein zweiter Ruf legt eine zweite Karte
// an —, oder ein Name, der löscht, widerruft oder entfernt: das ist die Klasse,
// nicht ein Fall.
func gefaehrlich(befehl Befehl) bool {
	if !befehl.Schreibend {
		return false
	}

	if !befehl.Idempotent {
		return true
	}

	for _, wort := range []string{"loeschen", "widerrufen", "entfernen", "zuruecknehmen", "aussteigen"} {
		if strings.Contains(befehl.Name, wort) {
			return true
		}
	}

	return false
}

// Befehlskatalog ist die Fensterbindung.
func (b *Bindungen) Befehlskatalog(ctx context.Context) BefehlskatalogAntwort {
	return b.app(ctx).Befehlskatalog(b.ruf(ctx))
}

// schemaOder gibt einem Werkzeug ohne Argumente ein leeres Objektschema.
// Ein nil-Verbund würde als `null` serialisiert, und das Fenster liest
// `schema.properties` — gefunden in der Sichtprüfung am 2026-09-09, weil die
// Probe jedem Befehl ein Objekt gab und der Betrieb nicht (§3).
func schemaOder(schema map[string]any) map[string]any {
	if schema == nil {
		return map[string]any{"type": "object"}
	}

	return schema
}

// befehleOffen sagt je Befehl, ob sein Werkzeugweg offen ist.
//
// **Ohne Einstellungen oder bei einem Lesefehler ist alles offen** — ein
// Schalter sichert nichts (Vermerk 0158), und vor dem Ruf fragt die Wache
// ohnehin noch einmal.
func (a *App) befehleOffen(ctx context.Context) func(name string) bool {
	if a.einstellungen == nil {
		return func(string) bool { return true }
	}

	werte, fehler := a.einstellungen.Alle(ctx)
	if fehler != nil {
		return func(string) bool { return true }
	}

	stand := domain.SchalterstandAus(werte)
	tafel := domain.EingebauteWegetafel()

	return func(name string) bool {
		return tafel.Pruefe(domain.FlaecheMCP, name, stand) == nil
	}
}
