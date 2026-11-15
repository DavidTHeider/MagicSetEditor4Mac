package rest

// **Karten kopieren und einfügen über die Rückschleife** — Werkstück 1 von
// Phase 3.7
// (#1168).

import (
	"context"
	"net/http"
	"strings"
)

// Fremdablage ist, was der Dienst vom Kopieren und Einfügen braucht.
type Fremdablage interface {
	AblageKopieren(ctx context.Context, setID int64, karteIDs []int64) (string, error)
	AblageAusschneiden(ctx context.Context, setID int64, karteIDs []int64) (string, error)
	AblageEinfuegen(ctx context.Context, setID int64, inhalt string) (
		karten, befunde []string, fehler error)
}

// ablageKopieren gibt die genannten Karten als Text.
//
// # Warum POST und nicht GET
//
// **Weil die Kartenliste in den Rumpf gehört.** Eine Auswahl von dreihundert
// Karten in einer Abfrage wäre eine Adresse, die keine Zwischenstelle
// weiterreicht — und ein `GET` mit Rumpf ist keiner. Geändert wird trotzdem
// nichts; das steht in der Antwort und nicht in der Methode.
func (d *Dienst) ablageKopieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.inDieAblage(schreiben, anfrage, d.naehte.Ablage.AblageKopieren)
}

// ablageAusschneiden gibt sie als Text und entfernt sie danach.
func (d *Dienst) ablageAusschneiden(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.inDieAblage(schreiben, anfrage, d.naehte.Ablage.AblageAusschneiden)
}

// inDieAblage ist der gemeinsame Rumpf beider Wege.
func (d *Dienst) inDieAblage(
	schreiben http.ResponseWriter, anfrage *http.Request,
	hole func(context.Context, int64, []int64) (string, error),
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Karten []int64 `json:"karten"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if len(rumpf.Karten) == 0 {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"es wurde keine Karte genannt — sie gehören als \"karten\" in den Rumpf")

		return
	}

	inhalt, fehler := hole(anfrage.Context(), setID, rumpf.Karten)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"inhalt": inhalt})
}

// ablageEinfuegen legt die Karten aus einem Text im Set an.
func (d *Dienst) ablageEinfuegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Inhalt string `json:"inhalt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if strings.TrimSpace(rumpf.Inhalt) == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"der Inhalt fehlt — er gehört als \"inhalt\" in den Rumpf")

		return
	}

	karten, befunde, fehler := d.naehte.Ablage.AblageEinfuegen(
		anfrage.Context(), setID, rumpf.Inhalt)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"karten": alsListe(karten), "befunde": alsListe(befunde),
	})
}
