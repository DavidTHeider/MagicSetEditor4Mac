package mcp

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Grenzen dieses Prozesses als Werkzeug** — Werkstück O5 von Phase 5.2
// (#1429).
//
// Ein Modell, das weiß, was ein Ruf kostet und wie viele Arbeiten zugleich
// laufen, plant seine Rufe — statt am 429 zu lernen. Vermerk 0033, Regel 9:
// erraten ist keine Schnittstelle.

// Kontingent ist ein Eimer, wie das Werkzeug ihn nennt.
type Kontingent struct {
	Burst           int `json:"burst"`
	Nachfuellen     int `json:"nachfuellen"`
	FensterSekunden int `json:"fensterSekunden"`
}

// Grenzen ist die Antwort von `lesen.grenzen`.
type Grenzen struct {
	Parallel     map[string]int `json:"parallel"`
	AntwortBytes int64          `json:"antwortBytes"`
	Principal    Kontingent     `json:"principal"`
	Global       Kontingent     `json:"global"`
	Reserve      int            `json:"reserve"`
	Kosten       map[string]int `json:"kosten"`

	// Kontingente ist `an` oder `aus` — QA-08.AK-03 (#1739).
	Kontingente string `json:"kontingente"`

	// Verbindungen, Auftraege und Kostentafel — #2003.
	Verbindungen int                     `json:"verbindungen"`
	Auftraege    Auftragsgrenzen         `json:"auftraege"`
	Kostentafel  []domain.Kostenauskunft `json:"kostentafel"`
}

// Auftragsgrenzen sind die Grenzen des Auftragsbuchs im Dienst.
type Auftragsgrenzen struct {
	Laufend              int `json:"laufend"`
	Behalten             int `json:"behalten"`
	AufbewahrungSekunden int `json:"aufbewahrungSekunden"`
}

// Grenzenauskunft ist die Naht zur Anwendung.
type Grenzenauskunft interface {
	Grenzen(ctx context.Context) (Grenzen, error)
}

func grenzenwerkzeug(naht Grenzenauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.grenzen", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Grenzen dieses Prozesses: »kosten« je Kostenklasse " +
			"(lesen, schreiben, bild, rendern, ausfuhr, einfuhr, masse; ein Abbruch kostet " +
			"nichts), »principal« " +
			"und »global« als Guthaben mit Burst und Zulauf je Fenster, »reserve« " +
			"als das, was nur das Fenster nehmen darf, »parallel« je Ressourcenklasse " +
			"und »antwortBytes«. »verbindungen« nennt, wie viele Verbindungen ein Dienst " +
			"über HTTP zugleich annimmt, »auftraege« das Auftragsbuch des REST-Dienstes " +
			"(laufend, behalten, Aufbewahrung in Sekunden), »kostentafel« je Fähigkeit, " +
			"die mehr kostet als ein Lesen oder Schreiben — oder weniger, weil ihr " +
			"REST-Weg zwar POST heißt, aber nur liest —, Klasse, Kosten und ihre Wege " +
			"an REST und MCP, soweit sie buchen, dazu die Kommandozeile. " +
			"»kontingente« ist »an« oder »aus«: aus heißt, " +
			"dass die Guthaben in dieser kontrollierten Umgebung abgeschaltet sind — " +
			"Plätze, Größen und Fristen gelten trotzdem. ÄNDERT NICHTS. Ein Ruf über " +
			"dem Guthaben bekommt einen Fehler mit der Wartezeit in Sekunden.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			grenzen, fehler := naht.Grenzen(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(grenzen)
		},
	}}
}
