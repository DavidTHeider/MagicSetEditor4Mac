// **Pakete einbinden, entfernen und aus MSE2 übernehmen** —
// #1291.
//
// # Warum sie eine eigene Datei haben
//
// `bestandswerkzeuge.go` beantwortet „was liegt im Bestand"; hier steht, was
// **auf die Platte** kommt und was von ihr verschwindet. Das ist eine andere
// Frage, und die Zusagen sind andere (§2,
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Paketpflege ist, was die Werkzeuge dieser Datei und die aus
// `paketplan.go` brauchen.
type Paketpflege interface {
	// PaketInstallieren bindet ein Paket ein, das schon auf der Platte liegt.
	PaketInstallieren(ctx context.Context, quelle string) (Paketzeile, error)

	// PaketAnlegen schreibt ein neues Paket aus einem Gerüst
	// (#1350).
	PaketAnlegen(ctx context.Context, art, name, vollname, spiel string) (Paketzeile, error)

	// PaketKopieren legt eine eigene Kopie eines Pakets an — FMT-02
	// (#1691).
	PaketKopieren(ctx context.Context, art, name, neu string) (Paketzeile, error)

	// PaketEntfernen nimmt ein Paket von der Platte und aus dem Katalog.
	//
	// **`pfad` ist die Bestätigung** und wird durchgereicht, nicht hier
	// geprüft: die Wache gehört dorthin, wo die Naht zusammengesetzt wird
	// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)),
	// denn sie ist eine Aussage über die **Platte**. Dieselbe Gestalt wie
	// `freierOrt` beim Zeichnen und Sichern.
	PaketEntfernen(ctx context.Context, art, name, pfad string) error

	// PaketUebernehmen holt alle Vorlagenpakete aus einem Verzeichnis.
	PaketUebernehmen(ctx context.Context, ort string) (Paketuebernahme, error)

	// PaketWeitergeben und PaketEinlesen gehen über `.mse-installer` —
	// FMT-02.AK-04 (#1691).
	PaketWeitergeben(ctx context.Context, pakete []string, ziel string) (string, error)
	PaketEinlesen(ctx context.Context, datei string) (Paketuebernahme, error)

	// MSE2Orte nennt die üblichen Orte einer MSE2-Installation.
	MSE2Orte() []string

	// PaketPlan sagt vor einer Änderung, was sie bewirkt — `entfernen`,
	// `einbinden` (aus `quelle`) oder `zurueck` (DOM-01, #1696).
	PaketPlan(ctx context.Context, vorgang, art, name, quelle string) (Paketplan, error)

	// PaketErsetzen legt eine neue Fassung über die vorhandene, die als
	// vorige aufgehoben wird. Ein lokal geändertes Paket braucht
	// `bestaetigt`.
	PaketErsetzen(ctx context.Context, quelle string, bestaetigt bool) (Paketzeile, error)

	// PaketZurueck tauscht die aktuelle Fassung gegen die vorige.
	PaketZurueck(ctx context.Context, art, name string) (Paketzeile, error)

	// PaketBedarf nennt, was einem Set an Paketen fehlt oder zu alt ist —
	// DOM-01.AK-04.
	PaketBedarf(ctx context.Context, setID int64) ([]Paketkonflikt, error)
}

// Paketuebernahme ist der Ausgang einer Übernahme.
type Paketuebernahme struct {
	Uebernommen []Paketzeile `json:"uebernommen"`
	Uebergangen []string     `json:"uebergangen"`
	Befunde     []string     `json:"befunde"`
}

// paketpflege gibt die Werkzeuge rund um installierte Pakete.
func paketpflege(pflege Paketpflege) []Werkzeug {
	if pflege == nil {
		return nil
	}

	return append([]Werkzeug{
		{
			Name: "lesen.mse2orte", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die üblichen Orte einer MSE2-Installation auf diesem " +
				"Rechner. VORGESCHLAGEN, nicht gelesen: was dort liegt, erfährt die " +
				"Anwendung erst über »schreiben.paket.uebernehmen«. Eine leere Liste " +
				"heißt: hier ist kein MSE2 installiert.",
			Schema: objekt(map[string]any{}),
			Ruf: func(context.Context, json.RawMessage) (string, error) {
				orte := pflege.MSE2Orte()
				if orte == nil {
					orte = []string{}
				}

				return alsJSON(map[string]any{"orte": orte})
			},
		},
		{
			Name: "schreiben.paket.installieren", Art: Schreibend, Pfadfelder: []string{"quelle"},
			Erklaerung: "ÄNDERT den Rechner: bindet ein Vorlagenpaket ein, das schon auf " +
				"der Platte liegt — ein Verzeichnis mit einer der sieben MSE-Endungen. " +
				"Ein Paket, das schon eingebunden ist, wird abgewiesen und nicht " +
				"überschrieben.",
			Schema: objekt(map[string]any{
				"quelle": text("Der Pfad zum Paketverzeichnis, etwa …/magic-m15.mse-style."),
			}, "quelle"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Quelle string `json:"quelle"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				paket, err := pflege.PaketInstallieren(ctx, argumente.Quelle)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"paket": paket})
			},
		},
		{
			Name: "schreiben.paket.anlegen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Rechner: legt ein NEUES Vorlagenpaket aus einem " +
				"Gerüst an — ein Spiel (mse-game) mit drei Kartenfeldern, einen Stil " +
				"(mse-style), der die Felder SEINES Spiels auslegt, ein Include " +
				"(mse-include) oder eine Exportvorlage (mse-export-template), die je " +
				"Karte den Namen ausgibt. Es landet im verwalteten Paketverzeichnis " +
				"und steht danach im Katalog. Stil und Exportvorlage BRAUCHEN »spiel«.",
			Schema: objekt(map[string]any{
				"art":      text("»mse-game«, »mse-style«, »mse-include« oder »mse-export-template«."),
				"name":     text("Der Kurzname; er wird zum Verzeichnisnamen. Buchstaben, Ziffern, Bindestriche, Unterstriche."),
				"vollname": text("Der Name in der Paketliste. Leer heißt: wie »name«."),
				"spiel":    text("Für Stil und Exportvorlage: der Name des Spielpakets aus »lesen.pakete«."),
			}, "art", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Art      string `json:"art"`
					Name     string `json:"name"`
					Vollname string `json:"vollname"`
					Spiel    string `json:"spiel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				paket, err := pflege.PaketAnlegen(
					ctx, argumente.Art, argumente.Name, argumente.Vollname, argumente.Spiel)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"paket": paket})
			},
		},
		{
			Name: "schreiben.paket.uebernehmen", Art: Schreibend, Pfadfelder: []string{"ort"},
			Erklaerung: "ÄNDERT den Rechner: holt alle Vorlagenpakete aus einem " +
				"Verzeichnis — etwa dem data-Ordner einer MSE2-Installation aus " +
				"»lesen.mse2orte«. Was schon eingebunden ist, wird ÜBERGANGEN und " +
				"nicht überschrieben.",
			Schema: objekt(map[string]any{
				"ort": text("Das Verzeichnis, dessen Pakete geholt werden."),
			}, "ort"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Ort string `json:"ort"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				// **Der Teilerfolg reist mit dem Fehler** — dieselbe Klasse
				// wie beim MTG-JSON-Einlesen (Befund 1 des Zweitblicks über
				// Phase 4.4). `Paketweg.Uebernehmen` kopiert die Pakete und
				// frischt **danach** den Katalog auf; scheitert das, **liegen
				// die Pakete schon da**. Wer nur den Fehler bekäme, ließe sie
				// unerkannt zurück (§7 Regel 3).
				uebernahme, err := pflege.PaketUebernehmen(ctx, argumente.Ort)
				if err != nil {
					// **Die Namen und nicht ihre Zahl** — Befund 2 des
					// Nachlaufs über Phase 4.4. „2 Pakete" sagt einem
					// Gegenüber, dass etwas kopiert wurde, aber nicht, was es
					// prüfen oder wegräumen muss.
					return "", fmt.Errorf("%w — übernommen wurden bereits %s",
						err, namenVon(uebernahme.Uebernommen))
				}

				return alsJSON(uebernahme)
			},
		},
		{
			Name: "schreiben.paket.entfernen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Rechner UNUMKEHRBAR: löscht ein eingebundenes " +
				"Vorlagenpaket von der Platte. Es gibt keinen Papierkorb. Zur " +
				"Bestätigung ist der Pfad aus »lesen.pakete« mitzugeben; stimmt er " +
				"nicht, geschieht nichts.",
			Schema: objekt(map[string]any{
				"art":  text("Die Art aus »lesen.pakete«, etwa mse-style."),
				"name": text("Der Name aus »lesen.pakete«, etwa magic-m15."),
				"pfad": text("Der Pfad aus »lesen.pakete« — die Bestätigung."),
			}, "art", "name", "pfad"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Art  string `json:"art"`
					Name string `json:"name"`
					Pfad string `json:"pfad"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pflege.PaketEntfernen(
					ctx, argumente.Art, argumente.Name, argumente.Pfad); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"entfernt": argumente.Art + "/" + argumente.Name,
				})
			},
		},
	}, append(append(append(paketplanung(pflege), paketkopie(pflege)), paketweitergabe(pflege)...),
		paketbezugwerkzeuge(pflege)...)...)
}

// namenVon nennt die Pakete einer Übernahme — Art und Name, durch Komma
// getrennt.
//
// **Keine leere Aufzählung**: „übernommen wurden bereits" ohne etwas dahinter
// läse sich wie ein abgeschnittener Satz.
func namenVon(pakete []Paketzeile) string {
	if len(pakete) == 0 {
		return "keine"
	}

	namen := make([]string, 0, len(pakete))
	for _, paket := range pakete {
		namen = append(namen, paket.Art+"/"+paket.Name)
	}

	return strings.Join(namen, ", ")
}
