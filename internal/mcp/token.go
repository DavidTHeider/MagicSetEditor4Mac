// **Zugriffstoken über MCP** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Vier Werkzeuge, alle erweiternd** (§16): sie kommen hinzu, keines ändert
// ein bestehendes.
//
// # Warum ein Sprachmodell Token verwalten darf
//
// **Weil Verwalten und Vorzeigen zweierlei sind.** Ein Token anzulegen heißt,
// einem Automaten einen Zugang einzurichten — dieselbe Art Aufgabe wie ein
// Profil anzulegen. Wer es **vorzeigt**, tut das am REST-Zugang, und dorthin
// reicht MCP nicht.
//
// **Der Klartext kommt trotzdem genau einmal heraus**, und das steht in der
// Erklärung des Werkzeugs: ein Modell, das ihn nicht weiterreicht, hat ihn
// verloren.

package mcp

import (
	"context"
	"encoding/json"
)

// Tokenpflege ist die Naht zu den Zugriffstoken.
type Tokenpflege interface {
	Token(ctx context.Context, profilID string) ([]Tokenzeile, error)
	TokenAnlegen(ctx context.Context, angaben Tokenangaben) (Tokenzeile, string, error)
	TokenAendern(ctx context.Context, angaben Tokenangaben) (Tokenzeile, error)
	TokenLoeschen(ctx context.Context, tokenID string) error

	// **Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von Phase 5.2
	// (#1426). Die Rotation fehlt: sie gäbe ein Geheimnis zurück.
	TokenAlleWiderrufen(ctx context.Context, profilID string) (int64, error)
	Tokenablaufwahlen(ctx context.Context) ([]Ablaufwahlzeile, error)
}

// Tokenzeile ist ein Zugriffstoken, wie MCP es ausgibt — **ohne sein
// Geheimnis**.
type Tokenzeile struct {
	ID       string `json:"id"`
	ProfilID string `json:"profilID"`

	Bezeichnung string `json:"bezeichnung"`
	Principal   string `json:"principal"`
	Scope       string `json:"scope"`
	Ablauf      string `json:"ablauf"`
	Angelegt    string `json:"angelegt"`
	Abgelaufen  bool   `json:"abgelaufen"`

	// **Werkstück O2 von Phase 5.2** (#1426): wann zuletzt angenommen, und
	// bis wann ein voriges Geheimnis nach einer Rotation noch gilt.
	LetzteNutzung string `json:"letzteNutzung"`
	VorgaengerBis string `json:"vorgaengerBis"`

	Fassung string `json:"fassung"`
}

// Ablaufwahlzeile ist ein benannter Vorschlag für den Ablauf eines Tokens.
type Ablaufwahlzeile struct {
	Name   string `json:"name"`
	Ablauf string `json:"ablauf"`
}

// Tokenangaben ist, was sich an einem Token setzen lässt.
type Tokenangaben struct {
	TokenID  string `json:"tokenID"`
	ProfilID string `json:"profilID"`

	Bezeichnung string `json:"bezeichnung"`
	Principal   string `json:"principal"`
	Scope       string `json:"scope"`
	Ablauf      string `json:"ablauf"`

	Fassung string `json:"fassung"`
}

// tokenwerkzeuge gibt das lesende und die drei schreibenden.
func tokenwerkzeuge(pflege Tokenpflege) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{tokenlisten(pflege), tokenablaufwahlen(pflege)}, []Werkzeug{
		tokenanlegen(pflege), tokenaendern(pflege), tokenloeschen(pflege),
		tokenalleWiderrufen(pflege),
	}
}

// tokenablaufwahlen nennt die Vorschläge für den Ablauf — Werkstück O2 von
// Phase 5.2 (#1426).
func tokenablaufwahlen(pflege Tokenpflege) Werkzeug {
	return Werkzeug{
		Name: "lesen.token.ablaufwahlen", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die vier Vorschläge für den Ablauf eines Tokens — " +
			"»tag«, »monat«, »jahr«, »nie« — mit dem fertig gerechneten Zeitpunkt " +
			"in RFC 3339, den »schreiben.token.anlegen« als »ablauf« nimmt. " +
			"»nie« hat einen leeren Zeitpunkt.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			wahlen, err := pflege.Tokenablaufwahlen(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"wahlen": wahlen})
		},
	}
}

// tokenalleWiderrufen zieht alle Token eines Profils zurück — Werkstück O2.
//
// **Die Rotation fehlt hier mit Absicht**: sie gäbe ein Geheimnis zurück, und
// das stünde danach im Verlauf eines Sprachmodells — derselbe Grund, aus dem
// `KontoEintragen` fehlt.
func tokenalleWiderrufen(pflege Tokenpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.token.widerrufen.alle", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: ZIEHT ALLE Zugriffstoken eines Profils " +
			"ZURÜCK — ENDGÜLTIG, auch das, mit dem diese Sitzung vielleicht " +
			"läuft. Antwortet mit der Zahl der gezogenen Token; null ist kein " +
			"Fehler.",
		Schema: objekt(map[string]any{
			"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
		}, "profilID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				ProfilID string `json:"profilID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			gezogen, err := pflege.TokenAlleWiderrufen(ctx, argumente.ProfilID)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"gezogen": gezogen})
		},
	}
}

func tokenlisten(pflege Tokenpflege) Werkzeug {
	return Werkzeug{
		Name: "lesen.token", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Zugriffstoken eines Profils — ihre Kennung, " +
			"Bezeichnung, Erlaubnis und ihren Ablauf. NICHT ihre Geheimnisse: " +
			"die gibt es nur einmal, beim Anlegen. »abgelaufen« sagt, ob der " +
			"Ablauf JETZT überschritten ist.",
		Schema: objekt(map[string]any{
			"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
		}, "profilID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				ProfilID string `json:"profilID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			zeilen, err := pflege.Token(ctx, argumente.ProfilID)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"token": zeilen})
		},
	}
}

func tokenanlegen(pflege Tokenpflege) Werkzeug {
	// **Ohne Idempotenzzusage wie jedes schreibende Werkzeug** — und hier aus
	// einem **zweiten**, stärkeren Grund: ein zweiter Ruf legt ein zweites
	// Token mit einem **anderen** Geheimnis an. Wer nach einer
	// Zeitüberschreitung wiederholt, hätte danach zwei, und das erste kennte
	// niemand mehr. Dieselbe Gestalt wie bei `schreiben.karte.anlegen`
	// (#895); der allgemeine Grund steht an
	// `TestKeinSchreibendesWerkzeugSagtIdempotenzZu`.
	return Werkzeug{
		Name: "schreiben.token.anlegen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt ein Zugriffstoken für ein Profil " +
			"an, mit dem sich ein Automat am REST-Zugang ausweist. Der " +
			"KLARTEXT steht NUR in dieser einen Antwort, unter »klartext« — " +
			"danach gibt es ihn nirgends mehr; wer ihn verliert, legt ein " +
			"neues an. »scope« ist »lesen« oder »schreiben«; »schreiben« " +
			"schließt Lesen ein. »ablauf« ist ein Zeitpunkt in RFC 3339 — leer " +
			"heißt: läuft nicht ab.",
		Schema: objekt(map[string]any{
			"profilID":    text("Die Kennung des Profils aus »lesen.nutzer«."),
			"bezeichnung": text("Wofür es da ist — etwa »Nachtlauf«."),
			"principal": text("Für wen es arbeitet — ein Mensch oder eine " +
				"Maschine, etwa »Baurechner«."),
			"scope": text("»lesen« oder »schreiben«."),
			"ablauf": text("Zeitpunkt in RFC 3339, etwa »2027-01-01T00:00:00Z« — " +
				"»lesen.token.ablaufwahlen« rechnet Vorschläge aus. Leer heißt: " +
				"läuft nicht ab."),
		}, "profilID", "bezeichnung", "principal", "scope"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente Tokenangaben

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			angelegt, klartext, err := pflege.TokenAnlegen(ctx, argumente)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"token": angelegt, "klartext": klartext})
		},
	}
}

func tokenaendern(pflege Tokenpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.token.aendern", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: setzt Bezeichnung, Erlaubnis und Ablauf " +
			"eines vorhandenen Tokens. Das GEHEIMNIS bleibt dasselbe — wer ein " +
			"neues will, legt ein neues Token an und zieht das alte zurück. Es " +
			"SETZT und verschmilzt nicht.",
		Schema: objekt(map[string]any{
			"tokenID":     text("Die Kennung des Tokens aus »lesen.token«."),
			"bezeichnung": text("Wofür es da ist."),
			"principal":   text("Für wen es arbeitet."),
			"scope":       text("»lesen« oder »schreiben«."),
			"ablauf":      text("Zeitpunkt in RFC 3339. Leer heißt: läuft nicht ab."),
			"fassung":     text("Die zuletzt gelesene Fassung aus »lesen.token«."),
		}, "tokenID", "bezeichnung", "principal", "scope", "fassung"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente Tokenangaben

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			geaendert, err := pflege.TokenAendern(ctx, argumente)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"token": geaendert})
		},
	}
}

func tokenloeschen(pflege Tokenpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.token.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: ZIEHT ein Zugriffstoken ZURÜCK — " +
			"ENDGÜLTIG. Wer es benutzt hat, kommt danach nicht mehr herein. " +
			"Was damit angelegt wurde, bleibt im Bestand.",
		Schema: objekt(map[string]any{
			"tokenID": text("Die Kennung des Tokens aus »lesen.token«."),
		}, "tokenID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				TokenID string `json:"tokenID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pflege.TokenLoeschen(ctx, argumente.TokenID); err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"geloescht": argumente.TokenID})
		},
	}
}
