// **Plan, Ersetzen und Zurück** — DOM-01.AK-02 und AK-03
// (#1696).
//
// # Warum eine eigene Datei
//
// `paketpflege.go` bindet ein, legt an, übernimmt und entfernt; hier steht,
// was **vor** einer Änderung zu sehen ist und wie eine Fassung gegen eine
// andere getauscht wird. Die Werkzeuge hängen an derselben Naht
// ([Paketpflege]) und erscheinen mit ihr — **erweiternd** nach §16.

package mcp

import (
	"context"
	"encoding/json"
)

// Paketplan ist, was eine Paketänderung bewirken würde.
type Paketplan struct {
	Vorgang   string          `json:"vorgang"`
	Paket     string          `json:"paket"`
	Quelle    string          `json:"quelle"`
	Von       string          `json:"von"`
	Zu        string          `json:"zu"`
	Geaendert bool            `json:"geaendert"`
	Moeglich  bool            `json:"moeglich"`
	Sets      []string        `json:"sets"`
	Konflikte []Paketkonflikt `json:"konflikte"`
}

// Paketkonflikt ist ein Bedarf, der nach der Änderung nicht mehr erfüllt
// wäre.
type Paketkonflikt struct {
	Paket    string `json:"paket"`
	Braucht  string `json:"braucht"`
	Art      string `json:"art"`
	Verlangt string `json:"verlangt"`
	Da       string `json:"da"`
}

// paketplanung gibt die vier Werkzeuge rund um den Plan.
func paketplanung(pflege Paketpflege) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.set.paketbedarf", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt, was einem Set an Paketen FEHLT oder zu alt ist: sein Spiel, " +
				"sein Stil und alles, was diese über »depends on« brauchen, Stufe für " +
				"Stufe. Eine leere Liste heißt: nichts fehlt. Beheben lässt es sich mit " +
				"»schreiben.paket.installieren« oder »schreiben.paket.beziehen« — " +
				"vorher zeigt »lesen.paket.plan«, was es bewirkt.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			}, "setID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				setID, err := kennung(roh)
				if err != nil {
					return "", err
				}

				befunde, err := pflege.PaketBedarf(ctx, setID)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"befunde": liste(befunde)})
			},
		},
		{
			Name: "lesen.paket.plan", Art: Lesend, Idempotent: true, Pfadfelder: []string{"quelle"},
			Erklaerung: "Sagt VOR einer Paketänderung, was sie bewirkt, und ändert " +
				"nichts: welche Sets das Paket benutzen, von welcher Fassung zu welcher " +
				"es geht, ob es lokal geändert ist und welche Abhängigkeiten danach " +
				"nicht mehr erfüllt wären. »vorgang« ist »entfernen« und »zurueck« " +
				"(mit »art« und »name«) oder »einbinden« (mit »quelle«; gibt es das " +
				"Paket schon, heißt der Vorgang im Ergebnis »ersetzen«). NICHTS wird " +
				"mitgenommen: ein Konflikt wird genannt, nicht gelöst.",
			Schema: objekt(map[string]any{
				"vorgang": text("»entfernen«, »einbinden« oder »zurueck«."),
				"art":     text("Beim Entfernen und Zurück: die Art aus »lesen.pakete«."),
				"name":    text("Beim Entfernen und Zurück: der Name aus »lesen.pakete«."),
				"quelle":  text("Beim Einbinden: der Pfad zum Paketverzeichnis."),
			}, "vorgang"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Vorgang string `json:"vorgang"`
					Art     string `json:"art"`
					Name    string `json:"name"`
					Quelle  string `json:"quelle"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				plan, err := pflege.PaketPlan(
					ctx, argumente.Vorgang, argumente.Art, argumente.Name, argumente.Quelle)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"plan": plan})
			},
		},
		{
			Name: "schreiben.paket.ersetzen", Art: Schreibend, Pfadfelder: []string{"quelle"},
			Erklaerung: "ÄNDERT den Rechner: legt eine neue Fassung über ein " +
				"eingebundenes Paket gleichen Namens. Die bisherige wird als VORIGE " +
				"Fassung aufgehoben und lässt sich mit »schreiben.paket.zurueck« " +
				"wiederholen. Ist das Paket seit dem Einbinden lokal geändert, wird " +
				"es nur mit »bestaetigt« ersetzt — setze es erst, wenn ein Mensch " +
				"den Plan aus »lesen.paket.plan« gesehen hat.",
			Schema: objekt(map[string]any{
				"quelle":     text("Der Pfad zur neuen Fassung, etwa …/magic-m15.mse-style."),
				"bestaetigt": jaNein("Wahr, wenn eine lokale Änderung überschrieben werden darf; sie bleibt als vorige Fassung erhalten."),
			}, "quelle"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Quelle     string `json:"quelle"`
					Bestaetigt bool   `json:"bestaetigt"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				paket, err := pflege.PaketErsetzen(ctx, argumente.Quelle, argumente.Bestaetigt)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"paket": paket})
			},
		},
		{
			Name: "schreiben.paket.zurueck", Art: Schreibend,
			Erklaerung: "ÄNDERT den Rechner: tauscht die aktuelle Fassung eines " +
				"Pakets gegen die vorige, die ein Ersetzen aufgehoben hat. Der Tausch " +
				"ist selbst umkehrbar: ein zweiter Ruf holt die ersetzte wieder. Ohne " +
				"vorige Fassung geschieht nichts.",
			Schema: objekt(map[string]any{
				"art":  text("Die Art aus »lesen.pakete«, etwa mse-style."),
				"name": text("Der Name aus »lesen.pakete«, etwa magic-m15."),
			}, "art", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Art  string `json:"art"`
					Name string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				paket, err := pflege.PaketZurueck(ctx, argumente.Art, argumente.Name)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"paket": paket})
			},
		},
	}
}
