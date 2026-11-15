// **Freigaben über MCP** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Drei Werkzeuge, alle erweiternd** (§16).
//
// # Warum ein Sprachmodell Freigaben verwalten darf
//
// **Weil es nichts vergeben kann, was der Mensch dahinter nicht schon hat.**
// Die Rolle des arbeitenden Menschen entscheidet, und die steht fest, bevor
// ein Werkzeug gerufen wird — ein Modell kann sie nicht wählen.

package mcp

import (
	"context"
	"encoding/json"
)

// Freigabenpflege ist die Naht zu den Freigaben.
type Freigabenpflege interface {
	Freigaben(ctx context.Context, profilID string) (Freigabenstand, error)
	FreigabeSetzen(ctx context.Context, profilID, nutzerID, rolle string) (
		Freigabenstand, error)
	FreigabeLoeschen(ctx context.Context, profilID, nutzerID string) (
		Freigabenstand, error)
}

// Freigabenstand ist, wer an einem Profil was darf.
type Freigabenstand struct {
	Freigaben []Freigabezeile `json:"freigaben"`

	// Sichtbarkeit ist `privat` oder `bestand`.
	Sichtbarkeit string `json:"sichtbarkeit"`

	// EigeneRolle sagt, was der arbeitende Mensch darf — leer heißt: nichts.
	EigeneRolle string `json:"eigeneRolle"`

	// Nachladefehler sagt: die Änderung **gilt**, nur die Liste danach war
	// nicht zu lesen — etwa nach dem eigenen Herabstufen (§7, Regel 3;
	// #1970). **Erweiternd im Sinne von 0033**: wo nichts fehlt, steht es
	// nicht da.
	Nachladefehler *Nachladefehler `json:"nachladefehler,omitempty"`
}

// Freigabezeile ist die Rolle eines Nutzers.
type Freigabezeile struct {
	NutzerID string `json:"nutzerID"`
	Name     string `json:"name"`
	Rolle    string `json:"rolle"`
	Angelegt string `json:"angelegt"`
}

// freigabenwerkzeuge gibt das lesende und die zwei schreibenden.
func freigabenwerkzeuge(pflege Freigabenpflege) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{freigabenlesen(pflege)},
		[]Werkzeug{freigabesetzen(pflege), freigabeloeschen(pflege)}
}

func freigabenlesen(pflege Freigabenpflege) Werkzeug {
	return Werkzeug{
		Name: "lesen.freigaben", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt, wer an einem Profil welche Rolle hat — und unter " +
			"»eigeneRolle«, was der Mensch darf, für den dieser Server läuft. " +
			"Drei Rollen, aufsteigend: »leser« SIEHT nur, dass es das Profil " +
			"gibt; »mitarbeiter« wechselt hinein und arbeitet darin; " +
			"»verwalter« vergibt selbst Freigaben. Der Eigner ist Verwalter " +
			"seines Profils, OHNE dass eine Zeile dasteht.",
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

			stand, err := pflege.Freigaben(ctx, argumente.ProfilID)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}
}

func freigabesetzen(pflege Freigabenpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.freigabe", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: gibt einem anderen Nutzer eine Rolle " +
			"an einem Profil — oder schreibt seine bisherige fort. NUR ein " +
			"»verwalter« darf das; alles andere wird ABGEWIESEN. Ein Profil " +
			"lässt sich nicht an seinen eigenen Eigner freigeben: er " +
			"verwaltet es ohnehin. **Ein »leser« kommt NICHT in das Profil " +
			"hinein** — dafür braucht es »mitarbeiter«.",
		Schema: objekt(map[string]any{
			"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
			"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
			"rolle":    text("»leser«, »mitarbeiter« oder »verwalter«."),
		}, "profilID", "nutzerID", "rolle"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				ProfilID string `json:"profilID"`
				NutzerID string `json:"nutzerID"`
				Rolle    string `json:"rolle"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			stand, err := pflege.FreigabeSetzen(
				ctx, argumente.ProfilID, argumente.NutzerID, argumente.Rolle)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}
}

func freigabeloeschen(pflege Freigabenpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.freigabe.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: ZIEHT eine Freigabe ZURÜCK. Wer sie " +
			"hatte, kommt danach nicht mehr in das Profil. Was er darin " +
			"angelegt hat, bleibt. NUR ein »verwalter« darf das.",
		Schema: objekt(map[string]any{
			"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
			"nutzerID": text("Die Kennung des Nutzers aus »lesen.freigaben«."),
		}, "profilID", "nutzerID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				ProfilID string `json:"profilID"`
				NutzerID string `json:"nutzerID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			stand, err := pflege.FreigabeLoeschen(
				ctx, argumente.ProfilID, argumente.NutzerID)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}
}
