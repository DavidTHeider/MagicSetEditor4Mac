package mcp

// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// # Drei Werkzeuge und ein bewusstes Nein
//
// Lesen darf ein Modell alles: was zurückgehen kann, was eine Sicherung
// kostet, und worüber nach einer Rückkehr nichts feststeht.
//
// **Quittieren darf es nicht.** Eine Quittung sagt „ich habe nachgesehen" —
// und nachsehen heißt hier: in die Kiste greifen und zählen. Ein Modell kann
// das nicht, und ein Werkzeug, mit dem es trotzdem behaupten könnte, es getan
// zu haben, machte aus der Unsicherheit eine Behauptung. Der Grund steht in
// `docs/FAEHIGKEITEN.md`.
//
// **Und zurückgespielt wird über MCP ohnehin nicht** ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
// ein Werkzeug schreibt nur, was es nicht überschreibt.

import (
	"context"
	"encoding/json"
)

// Ruecksprungart ist eine Art samt dem, was ihr im Weg steht.
type Ruecksprungart struct {
	Art       string `json:"art"`
	Hindernis string `json:"hindernis"`
}

// Vergleichszeile ist ein Zählstand nebeneinander.
type Vergleichszeile struct {
	Gegenstand  string `json:"gegenstand"`
	InSicherung int    `json:"inSicherung"`
	Jetzt       int    `json:"jetzt"`
	Verlust     int    `json:"verlust"`
	Unbekannt   bool   `json:"unbekannt"`
}

// Menge ist eine Stückzahl eines Gegenstands der Sammlung.
type Menge struct {
	Schluessel string `json:"schluessel"`
	Zustand    string `json:"zustand"`
	Sprache    string `json:"sprache"`

	// Foil ist der alte Wahrheitswert — abgekündigt seit T2 (0119).
	Foil bool `json:"foil"`

	// Finish ist die Ausführung: `nonfoil`, `foil` oder `etched`.
	Finish string `json:"finish"`

	Menge   int    `json:"menge"`
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
}

// Ruecksprunglage ist der Vergleich vor der Rückkehr.
type Ruecksprunglage struct {
	Datenstand         string            `json:"datenstand"`
	Zeilen             []Vergleichszeile `json:"zeilen"`
	Unklar             []Menge           `json:"unklar"`
	HandpruefungNoetig bool              `json:"handpruefungNoetig"`
}

// Rueckkehr ist eine verbuchte Rückkehr.
type Rueckkehr struct {
	ID              int64   `json:"id"`
	Zurueckgespielt string  `json:"zurueckgespielt"`
	Datenstand      string  `json:"datenstand"`
	Quelle          string  `json:"quelle"`
	Klaerung        string  `json:"klaerung"`
	GeklaertAm      string  `json:"geklaertAm"`
	Mengen          []Menge `json:"mengen"`
}

// Ruecksprungauskunft ist die Naht zur Anwendung.
type Ruecksprungauskunft interface {
	Ruecksprungarten(ctx context.Context) ([]Ruecksprungart, error)
	Ruecksprunglage(ctx context.Context, pfad string) (Ruecksprunglage, error)
	Rueckkehren(ctx context.Context) ([]Rueckkehr, error)
}

func ruecksprungwerkzeuge(naht Ruecksprungauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{
		{
			Name: "lesen.ruecksprungarten", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die VIER Arten, zurückzugehen — Appfassung, " +
				"Schema, Datenstand, Paketstand — und was jeder im Weg steht. " +
				"»hindernis« leer heißt: diese Anwendung kann es selbst. " +
				"»nachProzessende« heißt: ein Mensch tauscht das Bündel, wenn " +
				"das Programm nicht mehr läuft. »nurVorwaerts« heißt: es gibt " +
				"KEINE Rückwärtswanderung — der Weg zurück ist eine Sicherung. " +
				"ÄNDERT NICHTS.",
			Schema: objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				arten, fehler := naht.Ruecksprungarten(ctx)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(arten)
			},
		},
		{
			Name: "lesen.ruecksprunglage", Art: Lesend, Idempotent: true,
			Erklaerung: "Stellt eine Sicherungsdatei neben den laufenden " +
				"Bestand: was seit ihr dazukam, verschwindet beim " +
				"Zurückspielen (»verlust«) — und was sie MEHR hat, stünde " +
				"danach wieder da, ohne dass jemand weiß, ob es noch jemandem " +
				"gehört (»unklar«). »handpruefungNoetig« heißt, dass eine der " +
				"beiden Seiten ihre Sammlung nicht hergibt; dann ist KEINE " +
				"Aussage über unklare Mengen zu treffen. ÄNDERT NICHTS.",
			Schema: objekt(map[string]any{
				"pfad": text("Pfad der Sicherungsdatei"),
			}, "pfad"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var eingabe struct {
					Pfad string `json:"pfad"`
				}

				if fehler := json.Unmarshal(roh, &eingabe); fehler != nil {
					return "", fehler
				}

				lage, fehler := naht.Ruecksprunglage(ctx, eingabe.Pfad)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(lage)
			},
		},
		{
			Name: "lesen.rueckkehren", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt jede verbuchte Rückkehr dieses Bestands — wann, " +
				"auf welchen Datenstand, aus welcher Datei, und welche Mengen " +
				"sie unklar gelassen hat. »klaerung« ist »offen« oder " +
				"»quittiert«; eine offene sperrt ihre Mengen, bis ein MENSCH " +
				"nachgesehen hat. Sie verjährt nicht. ÄNDERT NICHTS.",
			Schema: objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				rueckkehren, fehler := naht.Rueckkehren(ctx)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(rueckkehren)
			},
		},
	}
}
