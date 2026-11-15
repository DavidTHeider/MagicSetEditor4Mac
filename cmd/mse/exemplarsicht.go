// **Was die Kommandozeile als JSON herausgibt** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum es diese Gestalt gibt
//
// **Weil `mse --json sammlung` bis heute die Wails-DTO ausgeschüttet hat.**
// Nach [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md),
// Regel 5 ist die Wails-Naht **innen** — sie darf sich mit jeder Fassung
// ändern —, die Kommandozeile aber **außen**. Wer beides über einen Typ
// führt, hat die innere Naht eingefroren oder die äußere gebrochen; welches
// von beidem, merkt er erst hinterher.
//
// **Angestrichen beim Umbau auf die Ausführung**: `foil` fiel aus der DTO, und
// damit aus der Ausgabe eines Befehls, den jemand in ein Skript geschrieben
// haben kann.
//
// MCP und REST haben ihre eigene Gestalt seit jeher (`mcp.Exemplarsicht`,
// `rest.Exemplarsicht`); hier steht die dritte.
package main

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"

// exemplarsicht ist ein Exemplar, wie die Kommandozeile es schreibt.
type exemplarsicht struct {
	ID       int64  `json:"id"`
	Fassung  int64  `json:"fassung"`
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Foil ist der **alte** Wahrheitswert — abgeleitet aus
	// [exemplarsicht.Finish] und seit T2 abgekündigt
	// ([0119](../../docs/entscheidungen/0119-eine-ausfuehrung-statt-eines-wahrheitswerts.md)).
	Foil bool `json:"foil"`

	// Finish ist die physische Ausführung: `nonfoil`, `foil` oder `etched`.
	Finish string `json:"finish"`

	Erwerbsdatum    string `json:"erwerbsdatum"`
	Erwerbspreis    string `json:"erwerbspreis"`
	Erwerbswaehrung string `json:"erwerbswaehrung"`
	Notiz           string `json:"notiz"`
	DruckID         string `json:"druckID"`
	Vorlaeufig      bool   `json:"vorlaeufig"`
	Name            string `json:"name"`
	Setcode         string `json:"setcode"`
	Nummer          string `json:"nummer"`
	Lagerort        string `json:"lagerort"`
	Wunsch          bool   `json:"wunsch"`

	// Wunschrang und Wunschart gelten **nur an einem Wunsch** (T4, #1469) —
	// `niedrig`, `normal`, `hoch`, `sehrhoch` und `genau`, `jederdruck`.
	// An einer besessenen Zeile stehen sie auf ihrer Vorgabe.
	Wunschrang  string `json:"wunschrang"`
	Wunschart   string `json:"wunschart"`
	Bezeichnung string `json:"bezeichnung"`
}

// alsExemplarsichten übersetzt die DTOs in die Gestalt dieser Fläche.
//
// **Feld für Feld und nicht über JSON** — dieselbe Begründung wie bei MCP und
// REST: ein Umweg über `Marshal` verlöre beim ersten Feld, dessen Marken
// auseinanderlaufen, genau dieses Feld, und zwar still.
func alsExemplarsichten(dtos []app.ExemplarDTO) []exemplarsicht {
	sichten := make([]exemplarsicht, 0, len(dtos))

	for _, dto := range dtos {
		sichten = append(sichten, exemplarsicht{
			ID: dto.ID, Fassung: dto.Fassung, Printing: dto.Printing,
			Menge: dto.Menge, Zustand: dto.Zustand, Sprache: dto.Sprache,
			Foil:            dto.Finish == "foil",
			Finish:          dto.Finish,
			Erwerbsdatum:    dto.Erwerbsdatum,
			Erwerbspreis:    dto.Erwerbspreis,
			Erwerbswaehrung: dto.Erwerbswaehrung,
			Notiz:           dto.Notiz,
			DruckID:         dto.DruckID,
			Vorlaeufig:      dto.Vorlaeufig,
			Name:            dto.Name,
			Setcode:         dto.Setcode,
			Nummer:          dto.Nummer,
			Lagerort:        dto.Lagerort,
			Wunsch:          dto.Wunsch,
			Wunschrang:      dto.Wunschrang,
			Wunschart:       dto.Wunschart,
			Bezeichnung:     dto.Bezeichnung,
		})
	}

	return sichten
}
