// **Die Preise über MCP** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Der Vorbehalt steht in jeder Erklärung
//
// Die Quelle sagt über ihre Preise ausdrücklich: „daily estimates …
// **Absolutely no guarantee** is made for any price information" (nachgelesen
// am 2026-08-27). Ein Modell, das eine Zahl weitergibt, soll wissen, was sie
// ist — deshalb steht es dort, wo es liest, und nicht nur im Vermerk.
//
// # Und eines der vier geht ins Netz
//
// `schreiben.preiseholen` fragt die Gegenstelle; die anderen drei lesen, was
// dasteht. Der Unterschied ist wichtig genug, um in der Art zu stehen: **Holen
// ist schreibend**, denn es legt Momentaufnahmen ab.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Preiswerkstatt ist, was die Werkzeuge der Preise brauchen.
type Preiswerkstatt interface {
	PreiseHolen(ctx context.Context, printing string) ([]Preissicht, []string, error)
	Preise(ctx context.Context, printing string) ([]Preissicht, error)
	Preisverlauf(ctx context.Context, reihe Preisreihe) ([]Preissicht, error)
	Portfolio(ctx context.Context) (Portfoliosicht, error)

	// Preisbulk holt eine ganze Preisdatei einer Quelle (A3, #1442).
	Preisbulk(ctx context.Context, quelle, umfang string) (Bulkbilanzsicht, error)

	// Marktangebote gibt die Posten eines Marktplatzes zu einem Druck
	// (A4, #1443).
	Marktangebote(ctx context.Context, wunsch Marktwunsch) (Marktsicht, error)

	// Anbieterstand nennt das Kontingent an einem Anbieterkonto (A5, #1444).
	Anbieterstand(ctx context.Context, dienst string) (Anbietersicht, error)

	// Fremdpreisverlauf holt den Verlauf eines Drucks (A5, #1444).
	Fremdpreisverlauf(ctx context.Context, wunsch Verlaufsbezug) (Verlaufssicht, error)

	// KIModelle nennt, was über KI-Modelle feststeht (A6, #1445).
	KIModelle(ctx context.Context) (KIKatalogsicht, error)

	// KIVorschlag lässt ein Modell Felder einer Karte vorschlagen (A7, #1446).
	KIVorschlag(ctx context.Context, bezug KIVorschlagsbezug) (
		KIVorschlagssicht, error)

	// KIBatch lässt ein Modell für mehrere Karten vorschlagen (A13.1, #1452).
	//
	// **Es schreibt nichts** — übernommen wird je Vorschlag.
	KIBatch(ctx context.Context, bezug KIBatchbezug) (KIBatchsicht, error)

	// KIGruppe übernimmt mehrere Vorschläge **als eine Gruppe** —
	// Werkstück A13.1b (#1452). Ganz oder gar nicht.
	KIGruppe(ctx context.Context, bezug KIGruppenbezug) (KIGruppensicht, error)

	// KIErkennung liest ein Kartenfoto — Werkstück A13.3 (#1452).
	// Sie schreibt an keiner Karte und legt kein Exemplar an.
	KIErkennung(ctx context.Context, bezug KIErkennungsbezug) (
		KIErkennungssicht, error)

	// KIEinschaetzung lässt ein Modell eine Karte beurteilen (A13.2a, #1452).
	//
	// **Es ändert weder Karte noch Sammlung noch Spielstand** — es bucht
	// einen Kostenposten und fragt.
	KIEinschaetzung(ctx context.Context, bezug KIEinschaetzungsbezug) (
		KIEinschaetzungssicht, error)

	// KIModellPruefen misst, was ein Modell kann (A9, #1448).
	KIModellPruefen(ctx context.Context, bezug KIProbenbezug) (
		KIProbensicht, error)

	// KIModellVergessen nimmt die Messungen eines Modells zurück (A9, #1448).
	KIModellVergessen(ctx context.Context, bezug KIProbenbezug) error

	// KIGeraetemodellFuerSitzung schaltet Apples Gerätemodell für diesen
	// Server ab oder an (AI-05, #2042) und gibt den neuen Stand.
	KIGeraetemodellFuerSitzung(ctx context.Context, an bool) (bool, error)
}

// Bulkbilanzsicht ist, was aus einem Bulkbezug wurde.
type Bulkbilanzsicht struct {
	Quelle  string   `json:"quelle"`
	Umfang  string   `json:"umfang"`
	Drucke  int      `json:"drucke"`
	Staende int      `json:"staende"`
	Befunde []string `json:"befunde"`
}

// Preissicht ist ein Preisstand, wie diese Fläche ihn sieht.
type Preissicht struct {
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441) — als
	// Zeichenkette, leer für keine. Erweiternd.
	DruckID string `json:"druckID"`

	// **Die sieben Teile der Reihe** (A3, #1442) — `markt`, `art`, `zustand`
	// und `sprache` sind **erweiternd** dazugekommen.
	Quelle   string `json:"quelle"`
	Markt    string `json:"markt"`
	Art      string `json:"art"`
	Waehrung string `json:"waehrung"`
	Finish   string `json:"finish"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Lage sagt, was feststeht: `genannt`, `weggefallen` oder `ungeprueft`.
	Lage string `json:"lage"`

	// Wert ist leer, wo die Lage nicht `genannt` ist.
	Wert string `json:"wert"`

	// Beobachtet ist der Tag, den die Quelle nennt — leer für keinen.
	Beobachtet string `json:"beobachtet"`
	Erhoben    string `json:"erhoben"`

	// AlterStunden sagt, wie alt der Stand ist.
	AlterStunden int `json:"alterStunden"`
}

// Preisreihe benennt, welche Reihe ein Verlauf meint.
//
// **Ohne Vorgaben** (A3, #1442): „wenn nichts dasteht, dann Scryfall" wäre
// eine Preisstrategie, und das Issue verbietet sie ausdrücklich.
type Preisreihe struct {
	Printing string `json:"printing"`
	Quelle   string `json:"quelle"`
	Markt    string `json:"markt"`
	Art      string `json:"art"`
	Waehrung string `json:"waehrung"`
	Finish   string `json:"finish"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`
}

// FehlendeTeile nennt, was eine Reihe nicht sagt — leer heißt vollständig.
//
// **`zustand` und `sprache` gehören nicht dazu**: leer ist dort eine Angabe
// („unbekannt"), und heute füllt sie keine Quelle.
func (r Preisreihe) FehlendeTeile() string {
	fehlend := []string{}

	for name, wert := range map[string]string{
		"quelle": r.Quelle, "markt": r.Markt, "art": r.Art,
		"waehrung": r.Waehrung, "finish": r.Finish,
	} {
		if strings.TrimSpace(wert) == "" {
			fehlend = append(fehlend, name)
		}
	}

	slices.Sort(fehlend)

	return strings.Join(fehlend, ", ")
}

// Portfoliozeilensicht ist die Summe einer Währung.
type Portfoliozeilensicht struct {
	Waehrung       string `json:"waehrung"`
	Wert           string `json:"wert"`
	Exemplare      int    `json:"exemplare"`
	AeltesterPreis string `json:"aeltesterPreis"`
	AlterStunden   int    `json:"alterStunden"`
	Quelle         string `json:"quelle"`
}

// Portfoliosicht ist die Übersicht.
type Portfoliosicht struct {
	Zeilen    []Portfoliozeilensicht `json:"zeilen"`
	OhnePreis int                    `json:"ohnePreis"`
}

// vorbehalt steht in jeder Erklärung, die eine Zahl herausgibt.
const vorbehalt = "PREISE SIND SCHÄTZWERTE: Scryfall nennt sie »daily estimates " +
	"and/or market values provided by our affiliates« und sagt ausdrücklich, " +
	"dass für keine Preisangabe eine Gewähr übernommen wird. Jede Zahl trägt " +
	"deshalb ihre Quelle, ihren Zeitpunkt und ihr Alter in Stunden."

// preiswerkzeuge gibt die vier Werkzeuge der Preise.
func preiswerkzeuge(werkstatt Preiswerkstatt) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.preise", Art: Lesend,
			Erklaerung: "Gibt die zuletzt ABGELEGTEN Preise eines Drucks — je Währung " +
				"und Ausführung einen. ÄNDERT NICHTS und fragt NICHT im Netz. " +
				vorbehalt,
			Schema: objekt(map[string]any{
				"printing": text("Die Scryfall-Kennung des Drucks."),
			}, "printing"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				printing, err := einDruck(roh)
				if err != nil {
					return "", err
				}

				staende, fehler := werkstatt.Preise(ctx, printing)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Staende": ohnePreisNil(staende)})
			},
		},
		{
			Name: "lesen.preisverlauf", Art: Lesend,
			Erklaerung: "Gibt eine Preisreihe in der Zeit — ältester zuerst. " +
				"ÄNDERT NICHTS und fragt NICHT im Netz. " + vorbehalt,
			Schema: objekt(map[string]any{
				"printing": text("Die Scryfall-Kennung des Drucks."),
				"waehrung": text("Der Währungscode, etwa »EUR«, »USD« oder »TIX«."),
				"finish":   text("nonfoil, foil oder etched."),
			}, "printing", "waehrung", "finish"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				// **Die ganze Reihe und keine Vorgabe** (A3, #1442): welche
				// Reihe gemeint ist, sagt der Aufrufer.
				var reihe Preisreihe

				if err := lies(roh, &reihe); err != nil {
					return "", err
				}

				if fehlend := reihe.FehlendeTeile(); fehlend != "" {
					return "", fmt.Errorf(
						"%w: die Preisreihe nennt %s nicht",
						domain.ErrValidation, fehlend)
				}

				staende, fehler := werkstatt.Preisverlauf(ctx, reihe)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Staende": ohnePreisNil(staende)})
			},
		},
		{
			Name: "lesen.portfolio", Art: Lesend,
			Erklaerung: "Rechnet die eigene Sammlung zusammen — je Währung eine Summe. " +
				"ÄNDERT NICHTS und fragt NICHT im Netz. Währungen werden NICHT " +
				"umgerechnet: dafür bräuchte es einen Kurs, den dieses Programm " +
				"nicht hat. »ohnePreis« sagt, wie viele Exemplare ohne Preis " +
				"geblieben sind — ohne diese Zahl sagt die Summe wenig. " + vorbehalt,
			Schema: objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				uebersicht, fehler := werkstatt.Portfolio(ctx)
				if fehler != nil {
					return "", fehler
				}

				if uebersicht.Zeilen == nil {
					uebersicht.Zeilen = []Portfoliozeilensicht{}
				}

				return alsJSON(uebersicht)
			},
		},
		{
			// **Keine Idempotenz.** Innerhalb der Zwischenspeicherfrist käme
			// derselbe Stand mit demselben Zeitpunkt, und die Ablage
			// übergeht ihn — danach ist ein zweiter Ruf ein **neuer** Abruf
			// und legt eine neue Momentaufnahme ab. Genau das soll er auch;
			// eine Zusage „ohne zusätzliche Wirkung" wäre hier falsch.
			Name: "schreiben.preiseholen", Art: Schreibend,
			Erklaerung: "ÄNDERT DEN BESTAND und FRAGT SCRYFALL: holt die Preise eines " +
				"Drucks und legt sie als Momentaufnahmen ab. Es ändert keine " +
				"Karte und kein Set. Das ist der einzige Weg hier, der ins NETZ " +
				"geht — für dreihundert Karten sind es dreihundert Abrufe. " +
				"»befunde« nennt Preise, die nicht zu deuten waren; die übrigen " +
				"kommen trotzdem an. " + vorbehalt,
			Schema: objekt(map[string]any{
				"printing": text("Die Scryfall-Kennung des Drucks."),
			}, "printing"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				printing, err := einDruck(roh)
				if err != nil {
					return "", err
				}

				staende, befunde, fehler := werkstatt.PreiseHolen(ctx, printing)
				if fehler != nil {
					return "", fehler
				}

				if befunde == nil {
					befunde = []string{}
				}

				return alsJSON(map[string]any{
					"Staende": ohnePreisNil(staende), "Befunde": befunde,
				})
			},
		},
		{
			// **Ein Ruf, ein Gigabyte** (A3, #1442). Anders als
			// `schreiben.preiseholen` fragt dieser **einmal** und bringt die
			// Preise **aller** Drucke, die der Bestand kennt — MTGJSON
			// antwortet nicht je Karte, sondern mit einer Datei.
			//
			// **Keine Idempotenz**, aus demselben Grund wie nebenan: ein
			// zweiter Ruf ist ein neuer Abruf und legt neue Momentaufnahmen
			// ab.
			Name: "schreiben.preise.bulk", Art: Schreibend,
			Erklaerung: "ÄNDERT DEN BESTAND und LÄDT EINE GROSSE DATEI: holt " +
				"die Preise ALLER Drucke dieses Bestands bei einer Quelle auf " +
				"einmal. »umfang« ist »heute« (der laufende Tag) oder " +
				"»neunzigTage« (die letzten neunzig, über ein Gigabyte). " +
				"Die Quelle wird GENANNT und nicht erraten; heute gibt es " +
				"»mtgjson«. Es ändert keine Karte und kein Set. " + vorbehalt,
			Schema: objekt(map[string]any{
				"quelle": text("Die Preisquelle — heute nur »mtgjson«."),
				"umfang": text("»heute« oder »neunzigTage«."),
			}, "quelle", "umfang"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Quelle string `json:"quelle"`
					Umfang string `json:"umfang"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				bilanz, fehler := werkstatt.Preisbulk(ctx,
					argumente.Quelle, argumente.Umfang)
				if fehler != nil {
					return "", fehler
				}

				if bilanz.Befunde == nil {
					bilanz.Befunde = []string{}
				}

				return alsJSON(bilanz)
			},
		},
		marktwerkzeug(werkstatt),
		anbieterstandwerkzeug(werkstatt),
		fremdverlaufwerkzeug(werkstatt),
		kimodellewerkzeug(werkstatt),
		kivorschlagwerkzeug(werkstatt),
		kibatchwerkzeug(werkstatt),
		kigruppenwerkzeug(werkstatt),
		kierkennungswerkzeug(werkstatt),
		kieinschaetzungswerkzeug(werkstatt),
		kimodellpruefenwerkzeug(werkstatt),
		kimodellvergessenwerkzeug(werkstatt),
		kigeraetsitzungwerkzeug(werkstatt),
	}
}

// einDruck liest das einzige Argument `printing`.
func einDruck(roh json.RawMessage) (string, error) {
	var argumente struct {
		Printing string `json:"printing"`
	}

	if err := lies(roh, &argumente); err != nil {
		return "", err
	}

	return argumente.Printing, nil
}

// ohnePreisNil macht aus einer fehlenden Liste eine leere (§3).
func ohnePreisNil(staende []Preissicht) []Preissicht {
	if staende == nil {
		return []Preissicht{}
	}

	return staende
}
