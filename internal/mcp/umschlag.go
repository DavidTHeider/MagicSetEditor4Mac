// **Den portablen Umschlag am MCP-Server** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Drei Werkzeuge, und zwei davon schreiben verschiedene Dinge
//
// Der Maßstab ist die **Umkehrbarkeit**
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
// „Ein Werkzeug schreibt nur, was es nicht überschreibt."
//
//   - **Schauen** ändert nichts und ist lesend.
//   - **Schnüren** legt eine **neue** Datei an — an einen Ort, an dem nichts
//     liegt, wie `schreiben.bestand.sichern`.
//   - **Übernehmen** legt ein **neues** Set an. Es überschreibt keines: die
//     Kennungen werden neu vergeben, und das ist die Zusage von #1457.
//
// **Und die fachliche Übernahmefreigabe bleibt getrennt** — hier ist sie der
// zweite Ruf. Wer beides zusammenlegte, hätte ein Schauen, das schreibt.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Umschlagwerkstatt ist, was die drei Werkzeuge dieser Datei brauchen.
type Umschlagwerkstatt interface {
	UmschlagSchnueren(ctx context.Context, art, kennung, ziel string) (Umschlagangaben, error)
	UmschlagAnsehen(ctx context.Context, pfad string) (Umschlagangaben, error)
	UmschlagUebernehmen(ctx context.Context, pfad string) (Umschlagannahme, error)
}

// Umschlagangaben sind, was ein Umschlag über sich hergibt.
type Umschlagangaben struct {
	// Urheber ist die Peerkennung dessen, der unterschrieben hat.
	//
	// **Sie ist kein Vertrauen.** Wer sie liest, weiß, **wer** unterschrieben
	// hat — nicht, ob man ihm etwas glauben sollte.
	Urheber       string           `json:"urheber,omitempty"`
	Fingerabdruck string           `json:"fingerabdruck,omitempty"`
	Objekte       []Umschlagobjekt `json:"objekte"`
}

// Umschlagobjekt ist ein Stück im Umschlag.
type Umschlagobjekt struct {
	Art         string `json:"art"`
	Bezeichnung string `json:"bezeichnung"`
	Groesse     int64  `json:"groesse"`

	// Braucht nennt, was dieses Objekt zum Vollständigsein braucht.
	//
	// **Es steht hier, weil #1457 es verlangt**: „Import zeigt
	// Objekte/**Abhängigkeiten**/Konflikte". Eine Fläche, die es wegließe,
	// zeigte weniger als das Fenster.
	Braucht []string `json:"braucht"`
}

// Umschlagannahme ist, was aus einer Übernahme wird.
type Umschlagannahme struct {
	SetID   string   `json:"setID"`
	Karten  int      `json:"karten"`
	Befunde []string `json:"befunde"`
}

// umschlagwerkzeuge baut die drei.
func umschlagwerkzeuge(quelle Umschlagwerkstatt) (lesend, schreibend []Werkzeug) {
	if quelle == nil {
		return nil, nil
	}

	lesend = []Werkzeug{
		{
			Name: "lesen.umschlag", Art: Lesend, Idempotent: true,
			Erklaerung: "Sagt, was in einem .msepackage liegt: wer unterschrieben hat " +
				"und welche Objekte darin stehen. ÄNDERT NICHTS — weder am Bestand " +
				"noch an der Datei. Die genannte Kennung sagt, WER unterschrieben hat, " +
				"nicht, ob man ihm trauen sollte.",
			Schema: objekt(map[string]any{
				"pfad": text("Der Pfad des .msepackage."),
			}, "pfad"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				pfad, err := einPfad(roh)
				if err != nil {
					return "", err
				}

				angaben, err := quelle.UmschlagAnsehen(ctx, pfad)
				if err != nil {
					return "", err
				}

				return alsJSON(angaben)
			},
		},
	}

	schreibend = []Werkzeug{
		{
			Name: "schreiben.umschlag.schnueren", Art: Schreibend,
			Erklaerung: "ÄNDERT die Platte: schreibt ein Set als signierten .msepackage " +
				"an einen Ort, an dem noch nichts liegt. Am Bestand ändert sich nichts.",
			Schema: objekt(map[string]any{
				"setID": text("Die Kennung des Sets."),
				"ziel":  text("Der Pfad, an den geschrieben wird — dort darf nichts liegen."),
			}, "setID", "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Art     string `json:"art"`
					Kennung string `json:"kennung"`
					SetID   string `json:"setID"`
					Ziel    string `json:"ziel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				// **Die alte Marke bleibt lesbar** (0033, Regel 6): wer nur
				// `setID` schickt, meint ein Set — die Änderung ist
				// erweiternd (T7, #1472).
				if argumente.Kennung == "" {
					argumente.Kennung, argumente.Art = argumente.SetID, "set"
				}

				if argumente.Kennung == "" || argumente.Ziel == "" {
					return "", fmt.Errorf(
						"%w: es braucht eine Kennung und ein Ziel", domain.ErrValidation)
				}

				angaben, err := quelle.UmschlagSchnueren(ctx, argumente.Art,
					argumente.Kennung, argumente.Ziel)
				if err != nil {
					return "", err
				}

				return alsJSON(angaben)
			},
		},
		{
			Name: "schreiben.umschlag.uebernehmen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt den Inhalt eines .msepackage als NEUES " +
				"Set an. Es überschreibt nichts — die Kennungen werden neu vergeben. " +
				"Vorher »lesen.umschlag« rufen: was hier hereinkommt, stammt von einem " +
				"Fremden.",
			Schema: objekt(map[string]any{
				"pfad": text("Der Pfad des .msepackage."),
			}, "pfad"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				pfad, err := einPfad(roh)
				if err != nil {
					return "", err
				}

				annahme, err := quelle.UmschlagUebernehmen(ctx, pfad)
				if err != nil {
					return "", err
				}

				return alsJSON(annahme)
			},
		},
	}

	return lesend, schreibend
}

// einPfad liest das eine Argument, das zwei dieser Werkzeuge haben.
func einPfad(roh json.RawMessage) (string, error) {
	var argumente struct {
		Pfad string `json:"pfad"`
	}

	if err := lies(roh, &argumente); err != nil {
		return "", err
	}

	if argumente.Pfad == "" {
		return "", fmt.Errorf("%w: es wurde kein Pfad genannt", domain.ErrValidation)
	}

	return argumente.Pfad, nil
}
