// **Die Stände der bezogenen Bäume über MCP** — G04
// (#2251).
//
// Dieselbe Auskunft und dieselben Schritte wie im Fenster (§16). **Eine
// eigene Schnittstelle neben [Paketpflege]**: wer sie nicht erfüllt, bekommt
// die vier Werkzeuge nicht — eine Pflege ohne Bezüge bleibt, wie sie war.

package mcp

import (
	"context"
	"encoding/json"
)

// Bezugspflege ist, was die Werkzeuge dieser Datei brauchen.
type Bezugspflege interface {
	PaketBezuege(ctx context.Context, fragen bool) ([]Bezugsstand, error)
	PaketBezugVorschau(ctx context.Context, name, stand string) (Bezugsvorschau, error)
	PaketBezugAktualisieren(ctx context.Context, name, stand string, bestaetigt bool) (Paketbezug, error)
	PaketBezugZurueck(ctx context.Context, name string) (Paketbezug, error)
}

// Bezugsstand ist ein bezogener Baum samt Lage seiner Quelle.
type Bezugsstand struct {
	Name         string `json:"name"`
	Adresse      string `json:"adresse"`
	Stand        string `json:"stand"`
	Geprueft     string `json:"geprueft"`
	Pakete       int    `json:"pakete"`
	Vorige       bool   `json:"vorige"`
	VorigerStand string `json:"vorigerStand"`
	Lage         string `json:"lage"`
	Spitze       string `json:"spitze"`
	SpitzeZeit   string `json:"spitzeZeit"`
	Vorsprung    int    `json:"vorsprung"`
	Grund        string `json:"grund,omitempty"`
}

// Bezugsvorschau ist, was eine Aktualisierung täte.
type Bezugsvorschau struct {
	Name               string           `json:"name"`
	Von                string           `json:"von"`
	Zu                 string           `json:"zu"`
	Lage               string           `json:"lage"`
	Vorsprung          int              `json:"vorsprung"`
	Rueckstand         int              `json:"rueckstand"`
	Pakete             []Paketaenderung `json:"pakete"`
	Uebrige            int              `json:"uebrige"`
	Gekuerzt           bool             `json:"gekuerzt"`
	Sets               []string         `json:"sets"`
	BestaetigungNoetig bool             `json:"bestaetigungNoetig"`
	Gruende            []string         `json:"gruende"`
	VorigeErsetzt      bool             `json:"vorigeErsetzt"`
}

// Paketaenderung ist ein Paket, das eine Aktualisierung berührt.
type Paketaenderung struct {
	Paket     string `json:"paket"`
	Neu       int    `json:"neu"`
	Geaendert int    `json:"geaendert"`
	Entfernt  int    `json:"entfernt"`
}

// paketbezugwerkzeuge gibt die vier Werkzeuge — oder keines.
func paketbezugwerkzeuge(pflege Paketpflege) []Werkzeug {
	bezuege, erfuellt := pflege.(Bezugspflege)
	if !erfuellt {
		return nil
	}

	return []Werkzeug{
		{
			Name: "lesen.paket.bezuege", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die bezogenen Vorlagenbäume: Quelle, bezogener Commit, der " +
				"Commit, gegen den diese Anwendung geprüft ist, und ob ein voriger Stand " +
				"aufgehoben ist. Mit »fragen« wird die Quelle nach ihrem neuesten Commit " +
				"gefragt — das geht ins Netz (GitHub); ohne bleibt alles auf dem Rechner. " +
				"Lage: ungefragt, aktuell, neuer, abweichend oder unbekannt (dann mit Grund).",
			Schema: objekt(map[string]any{
				"fragen": jaNein("Wahr, um die Quellen zu fragen."),
			}),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Fragen bool `json:"fragen"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				staende, err := bezuege.PaketBezuege(ctx, argumente.Fragen)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"bezuege": staende})
			},
		},
		{
			Name: "lesen.paket.bezug.vorschau", Art: Lesend, Idempotent: true,
			Erklaerung: "Sagt, was eine Aktualisierung eines bezogenen Baums auf einen Commit " +
				"täte: welche Pakete sich ändern, welche Sets sie nennen, und ob eine " +
				"Bestätigung nötig ist (setVerliert, nichtVoraus, gekuerzt). Fragt GitHub. " +
				"Ein Zweig wird abgewiesen — nur ein Commit.",
			Schema: objekt(map[string]any{
				"name":  text("Der Name aus »lesen.paket.bezuege«, etwa M15-Magic-Pack."),
				"stand": text("Der Commit (40 Zeichen), etwa die »spitze« aus »lesen.paket.bezuege«."),
			}, "name", "stand"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Name  string `json:"name"`
					Stand string `json:"stand"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				vorschau, err := bezuege.PaketBezugVorschau(ctx, argumente.Name, argumente.Stand)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"vorschau": vorschau})
			},
		},
		{
			Name: "schreiben.paket.bezug.aktualisieren", Art: Schreibend,
			Erklaerung: "ÄNDERT den Rechner: lädt einen bezogenen Baum an einem neuen Commit " +
				"aus dem Netz und setzt ihn an die Stelle des bisherigen, der als voriger " +
				"Stand aufgehoben wird (»schreiben.paket.bezug.zurueck«). Verlangt die " +
				"Vorschau eine Bestätigung, geschieht ohne »bestaetigt« nichts — setze es " +
				"erst, wenn ein Mensch die Vorschau gesehen hat.",
			Schema: objekt(map[string]any{
				"name":       text("Der Name aus »lesen.paket.bezuege«."),
				"stand":      text("Der Commit (40 Zeichen)."),
				"bestaetigt": jaNein("Wahr, wenn ein Mensch die Gründe der Vorschau gesehen hat."),
			}, "name", "stand"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Name       string `json:"name"`
					Stand      string `json:"stand"`
					Bestaetigt bool   `json:"bestaetigt"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				ergebnis, err := bezuege.PaketBezugAktualisieren(ctx, argumente.Name, argumente.Stand, argumente.Bestaetigt)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"bezug": ergebnis})
			},
		},
		{
			Name: "schreiben.paket.bezug.zurueck", Art: Schreibend,
			Erklaerung: "ÄNDERT den Rechner: tauscht einen bezogenen Baum mit seinem " +
				"aufgehobenen vorigen Stand. Der Tausch ist selbst umkehrbar: ein zweiter " +
				"Ruf holt den neueren wieder. Ohne aufgehobenen Stand geschieht nichts.",
			Schema: objekt(map[string]any{
				"name": text("Der Name aus »lesen.paket.bezuege«."),
			}, "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Name string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				ergebnis, err := bezuege.PaketBezugZurueck(ctx, argumente.Name)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"bezug": ergebnis})
			},
		},
	}
}
