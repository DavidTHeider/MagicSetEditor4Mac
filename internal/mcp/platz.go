// **Was Platz belegt, und wie er frei wird** — Werkstück 4 von Phase 4.13
// (#1352).
//
// # Wofür ein Sprachmodell das braucht
//
// **Damit es auf die Frage „warum ist meine Platte voll?" antworten kann,
// statt zu raten.** Am 2026-09-07 lagen auf dem Rechner des Auftraggebers
// dreißig Sicherungen vor Migrationen — zusammen über zwei Gigabyte, und
// nichts sagte es.

package mcp

import (
	"context"
	"encoding/json"
)

// Platzauskunft ist, was die Werkzeuge dieser Datei brauchen.
type Platzauskunft interface {
	Platz(ctx context.Context) ([]Platzposten, error)
	PlatzFreigeben(ctx context.Context, posten string) (Platzgeraeumt, error)
}

// Platzgeraeumt ist, was nach dem Wegräumen kommt: die Zählung — und der
// Grund, aus dem sie fehlen kann (§7, Regel 3; #1970).
type Platzgeraeumt struct {
	Posten         []Platzposten
	Nachladefehler *Nachladefehler
}

// Platzposten ist ein Posten samt seinem Gewicht.
//
// **Die Marken sind die Zusage** — siehe [Satz].
type Platzposten struct {
	Posten   string `json:"Posten"`
	Dateien  int    `json:"Dateien"`
	Bytes    int64  `json:"Bytes"`
	Raeumbar bool   `json:"Raeumbar"`
}

// platzwerkzeuge gibt das lesende und das schreibende Werkzeug.
func platzwerkzeuge(auskunft Platzauskunft) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{lesenPlatz(auskunft)}, []Werkzeug{schreibenPlatzFreigeben(auskunft)}
}

// lesenPlatz zählt, was wie viel belegt.
func lesenPlatz(auskunft Platzauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.platz", Art: Lesend,
		// **Sechs Posten, wie die Domäne sie führt** — Befund 7 des Zweitblicks
		// über Phase 4.13 (2026-09-07). Dieser Text ist **ausführbares
		// öffentliches Verhalten**: ein Sprachmodell liest ihn und berät danach.
		Erklaerung: "Sagt, wie viel Platz was belegt — der Bestand, die " +
			"eingesetzten Bilder, die bezogenen Pakete, die Sicherungen vor " +
			"Migrationen, das Protokoll und unter »sonstiges« alles Übrige im " +
			"Anwendungsordner. ÄNDERT NICHTS. »Raeumbar« sagt, ob sich ein " +
			"Posten mit »schreiben.platz.freigeben« wegräumen lässt — das sind " +
			"die Sicherungen und das Protokoll, und auch dort bleibt etwas " +
			"stehen: bei den Sicherungen der Weg zurück, beim Protokoll die " +
			"laufende Datei. Die übrigen gehören zu den Daten, mit denen die " +
			"Anwendung arbeitet, oder sind Dateien, die sie nicht kennt.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			posten, fehler := auskunft.Platz(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(liste(posten))
		},
	}
}

// schreibenPlatzFreigeben räumt einen Posten weg.
func schreibenPlatzFreigeben(auskunft Platzauskunft) Werkzeug {
	return Werkzeug{
		Name: "schreiben.platz.freigeben", Art: Schreibend,
		// **Zwei Sicherungen bleiben, nicht eine** — der Nachlauf zum
		// Zweitblick über Phase 4.13 (2026-09-07): die jüngste kann nach zwei
		// Aktualisierungen von einer neueren Fassung stammen. Ein Modell liest
		// diesen Text und berät danach.
		Erklaerung: "Räumt einen räumbaren Posten weg und gibt die Zählung " +
			"danach. ÄNDERT: es LÖSCHT DATEIEN. Bei den Sicherungen bleiben " +
			"**zwei** stehen: die jüngste überhaupt und die jüngste, die " +
			"diese Fassung öffnen kann — nach zwei Aktualisierungen ist das " +
			"nicht dieselbe Datei. Beim Protokoll bleibt die laufende Datei. " +
			"Ein Posten, der nicht räumbar ist, wird abgewiesen.",
		Schema: objekt(map[string]any{
			"posten": text("Der Posten aus »lesen.platz«, etwa »vorsicherungen«."),
		}, "posten"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Posten string `json:"posten"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			geraeumt, fehler := auskunft.PlatzFreigeben(ctx, argumente.Posten)
			if fehler != nil {
				return "", fehler
			}

			text, _ := alsJSON(liste(geraeumt.Posten)) //nolint:errcheck // es gibt keinen, siehe [alsJSON]

			return nachgeladen(text, geraeumt.Nachladefehler)
		},
	}
}
