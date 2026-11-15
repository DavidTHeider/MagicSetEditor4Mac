package rest

// **Marken über die Rückschleife** — Werkstück 3 von Phase 3.5
// (#1208).

import (
	"context"
	"net/http"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdmarken ist, was der Dienst von den Marken braucht.
type Fremdmarken interface {
	MarkeSetzen(ctx context.Context, setID, karteID int64, marke string) error
	MarkeEntfernen(ctx context.Context, setID, karteID int64, marke string) error
	KartenMarken(ctx context.Context, setID, karteID int64) ([]string, error)
	Setmarken(ctx context.Context, setID int64) ([]string, error)
}

// setmarkenZeigen nennt die Marken, die in einem Set vorkommen.
func (d *Dienst) setmarkenZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	marken, fehler := d.naehte.Marken.Setmarken(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"marken": markenliste(marken)})
}

// kartenmarkenZeigen nennt die Marken einer Karte.
func (d *Dienst) kartenmarkenZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, karteID, gut := markenstelle(schreiben, anfrage)
	if !gut {
		return
	}

	marken, fehler := d.naehte.Marken.KartenMarken(anfrage.Context(), setID, karteID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"marken": markenliste(marken)})
}

// markeSetzen hängt eine Marke an eine Karte.
//
// **`PUT` und nicht `POST`**: dieselbe Marke zweimal zu setzen ergibt denselben
// Zustand, und das ist die Zusage, die `PUT` gibt.
//
// **Die Antwort trägt die Normalform, nicht den Wert aus dem Weg.**
// `PUT …/marken/Fertig` legt `fertig` ab, und ein Klient, der die
// Schreibantwort in seinen Zustand übernimmt, zeigt dasselbe wie der nächste
// Leseaufruf — §16 verlangt dieselbe Antwort an jeder Fläche
// (#1278,
// Befund 3).
func (d *Dienst) markeSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, karteID, marke, gut := markenweg(schreiben, anfrage)
	if !gut {
		return
	}

	if fehler := d.naehte.Marken.MarkeSetzen(anfrage.Context(), setID, karteID, marke); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"marke": domain.Markennormalform(marke), "gesetzt": true,
	})
}

// markeEntfernen nimmt eine Marke wieder ab.
func (d *Dienst) markeEntfernen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, karteID, marke, gut := markenweg(schreiben, anfrage)
	if !gut {
		return
	}

	if fehler := d.naehte.Marken.MarkeEntfernen(
		anfrage.Context(), setID, karteID, marke); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"marke": domain.Markennormalform(marke), "entfernt": true,
	})
}

// markenstelle liest Set- und Kartenkennung aus dem Weg.
func markenstelle(schreiben http.ResponseWriter, anfrage *http.Request) (
	setID, karteID int64, gut bool,
) {
	setID, gut = kennung(schreiben, anfrage, "setID")
	if !gut {
		return 0, 0, false
	}

	karteID, gut = kennung(schreiben, anfrage, "karteID")
	if !gut {
		return 0, 0, false
	}

	return setID, karteID, true
}

// markenweg liest zusätzlich die Marke — **aus dem Weg**, nicht aus dem Rumpf.
//
// **Weil sie die Sache benennt, die angesprochen wird.** `PUT
// /sets/1/karten/7/marken/fertig` ist dieselbe Adresse wie das `DELETE`
// darauf; eine Marke im Rumpf hätte für das Löschen keinen Platz (ein `DELETE`
// hat keinen), und dann stünde sie an zwei verschiedenen Stellen.
func markenweg(schreiben http.ResponseWriter, anfrage *http.Request) (
	setID, karteID int64, marke string, gut bool,
) {
	setID, karteID, gut = markenstelle(schreiben, anfrage)
	if !gut {
		return 0, 0, "", false
	}

	// **`PathValue` gibt den Abschnitt schon entpackt** — „noch kein bild"
	// steht im Weg als `noch%20kein%20bild` und kommt hier als Klartext an.
	//
	// **Ein zweites Entpacken wäre falsch und obendrein unerreichbar**: `%25`
	// würde zu `%`, und eine Marke „100 %" käme kaputt an. Unerreichbar, weil
	// `net/http` einen Weg mit ungültiger Kodierung gar nicht erst annimmt —
	// ein `if err != nil` daneben wäre ein Zweig, den kein Aufruf betritt (§3).
	entpackt := anfrage.PathValue("marke")

	if strings.TrimSpace(entpackt) == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"die Marke fehlt — sie gehört in den Weg")

		return 0, 0, "", false
	}

	return setID, karteID, entpackt, true
}

// markenliste gibt eine leere Liste statt `null`.
func markenliste(marken []string) []string {
	if marken == nil {
		return []string{}
	}

	return marken
}
