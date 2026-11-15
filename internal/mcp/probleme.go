package mcp

// **Die Probleme eines ganzen Sets als Werkzeuge** — PERF02
// (#2352).
//
// **Zwei Werkzeuge und nicht eines mit Schalter**: der Stand ist billig, die
// Prüfung zeichnet jede Karte. Die Kosten eines Rufs hängen am Namen des
// Wegs (`domain.Kostentafel`), und ein Schalter im Argument hätte die
// Prüfung so billig gemacht wie das Lesen. **Lesend beide**: die Prüfung
// zeichnet Karten und schreibt nichts.

import (
	"context"
	"encoding/json"
)

// Problemquelle ist, was die Werkzeuge brauchen.
type Problemquelle interface {
	Problemstand(ctx context.Context, setID int64) (Problemstand, error)
	ProblemePruefen(ctx context.Context, setID int64) (Problemstand, error)
}

// Kartenproblem sind die Befunde einer Karte.
//
// **Eine eigene Gestalt und nicht `app.KartenproblemDTO`**: `internal/mcp`
// ist ein Blatt und kennt `app` nicht (§2).
type Kartenproblem struct {
	KarteID string   `json:"karteID"`
	Befunde []string `json:"befunde"`
}

// Problemstand ist der Stand der Prüfung eines Sets.
type Problemstand struct {
	Urteil   string          `json:"urteil"`
	Zahl     int             `json:"zahl"`
	Sicher   bool            `json:"sicher"`
	Geprueft int             `json:"geprueft"`
	Gesamt   int             `json:"gesamt"`
	Beendet  string          `json:"beendet"`
	Alter    int             `json:"alter"`
	Karten   []Kartenproblem `json:"karten"`
}

// urteilssatz erklärt einem Modell, wie weit eine Zahl gilt — **ausführbares
// öffentliches Verhalten**: es entscheidet danach, ob es ihr glaubt.
const urteilssatz = "»urteil« sagt, wie weit der Stand gilt: aktuell, veraltet (das " +
	"Set, eine Schrift oder ein Paket hat sich seitdem geändert), unvollstaendig " +
	"(abgebrochen) oder ungeprueft. Nur bei »sicher« ist »zahl« eine Zahl, die gilt."

// problemwerkzeuge gibt das lesende und das prüfende Werkzeug. **Die Quelle
// wird erst beim Ruf gefragt** — eine Naht, die ein Test nicht setzt, macht
// die Liste sonst schon beim Bauen kaputt.
func problemwerkzeuge(quelle Problemquelle) []Werkzeug {
	return []Werkzeug{
		problemwerkzeug("lesen.set.probleme", "Nennt den letzten Stand der Prüfung eines "+
			"Sets auf Probleme, ohne zu prüfen. ÄNDERT NICHTS. "+urteilssatz,
			func(ctx context.Context, setID int64) (Problemstand, error) { return quelle.Problemstand(ctx, setID) }),
		problemwerkzeug("lesen.set.probleme.pruefen", "Prüft jede Karte eines Sets, wie "+
			"die Vorschau sie zeichnet — fehlende Schriften, Bilder, Stile, Skriptfehler — "+
			"und nennt die mit Problemen. ÄNDERT NICHTS. Bei einem großen Set bis zu einer "+
			"Minute. "+urteilssatz,
			func(ctx context.Context, setID int64) (Problemstand, error) {
				return quelle.ProblemePruefen(ctx, setID)
			}),
	}
}

func problemwerkzeug(name, erklaerung string, lesen func(context.Context, int64) (Problemstand, error)) Werkzeug {
	return Werkzeug{
		Name: name, Art: Lesend, Erklaerung: erklaerung,
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := kennung(argumente)
			if err != nil {
				return "", err
			}

			stand, fehler := lesen(ctx, wahl)
			if fehler != nil {
				return "", fehler
			}

			stand.Karten = liste(stand.Karten)

			return alsJSON(stand)
		},
	}
}
