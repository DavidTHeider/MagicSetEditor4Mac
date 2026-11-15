package rest

// **Die Spalten der Kartenliste und ihre Ordnung über die Rückschleife** —
// Werkstück 3 von Phase 3.7
// (#1168).

import (
	"context"
	"net/http"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdliste ist, was der Dienst von der Kartenliste braucht.
type Fremdliste interface {
	Listenaufbau(ctx context.Context, setID int64) (domain.Kartenlistenaufbau, error)

	ListenspaltenSetzen(ctx context.Context, setID int64, wahl []domain.Spaltenwahl) error
	ListensortierungSetzen(ctx context.Context, setID int64, ordnung domain.Kartenordnung) error

	// KartenFiltern gibt die Kennungen der Karten, die durch den Schnellfilter
	// kommen — Werkstück U5 von Phase 5.1 (#1416).
	KartenFiltern(ctx context.Context, setID int64, anfrage string) ([]int64, error)
}

// kartenliste gibt Spalten, Zeilen und die wirksame Ordnung.
func (d *Dienst) kartenliste(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	aufbau, fehler := d.naehte.Liste.Listenaufbau(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"spalten":     listenspalten(aufbau.Spalten),
		"zeilen":      listenzeilen(aufbau.Zeilen),
		"sortierfeld": aufbau.Ordnung.Feld,
		"aufsteigend": aufbau.Ordnung.Aufsteigend,
		"ohneVorlage": aufbau.OhneVorlage,
	})
}

// listenspaltenSetzen schreibt die Wahl fort — **die ganze Wahl**.
func (d *Dienst) listenspaltenSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Spalten []struct {
			Feld     string `json:"feld"`
			Sichtbar bool   `json:"sichtbar"`
		} `json:"spalten"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	wahl := make([]domain.Spaltenwahl, 0, len(rumpf.Spalten))

	// **Die Stelle kommt aus der Reihenfolge der Liste** — dieselbe Regel wie
	// am Fenster und über MCP, damit alle drei dasselbe meinen.
	for stelle, eine := range rumpf.Spalten {
		wahl = append(wahl, domain.Spaltenwahl{
			Feld: eine.Feld, Sichtbar: eine.Sichtbar, Stelle: stelle + 1,
		})
	}

	if fehler := d.naehte.Liste.ListenspaltenSetzen(
		anfrage.Context(), setID, wahl); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"spalten": len(wahl)})
}

// listensortierungSetzen schreibt die Ordnung fort.
func (d *Dienst) listensortierungSetzen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Feld        string `json:"feld"`
		Aufsteigend bool   `json:"aufsteigend"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	fehler := d.naehte.Liste.ListensortierungSetzen(anfrage.Context(), setID,
		domain.Kartenordnung{Feld: rumpf.Feld, Aufsteigend: rumpf.Aufsteigend})
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"sortierfeld": rumpf.Feld, "aufsteigend": rumpf.Aufsteigend,
	})
}

// listenspalten macht aus den Spalten eine Liste — **immer eine**.
func listenspalten(spalten []domain.Listenspalte) []map[string]any {
	liste := make([]map[string]any, 0, len(spalten))

	for _, spalte := range spalten {
		liste = append(liste, map[string]any{
			"feld":       spalte.Feld,
			"aufschrift": spalte.Aufschrift,
			"sichtbar":   spalte.Sichtbar,
		})
	}

	return liste
}

// listenzeilen macht aus den Zeilen eine Liste — ebenso.
func listenzeilen(zeilen []domain.Listenzeile) []map[string]any {
	liste := make([]map[string]any, 0, len(zeilen))

	for _, zeile := range zeilen {
		liste = append(liste, map[string]any{
			"karteID": zeile.KarteID,
			"nummer":  zeile.Nummer + 1,
			"werte":   zeile.Werte,
		})
	}

	return liste
}

// kartenfilter legt den Schnellfilter über die Karten eines Sets — Werkstück
// U5 von Phase 5.1 (#1416). Die Anfrage kommt als `?anfrage=`; leer heißt
// alle.
func (d *Dienst) kartenfilter(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	kennungen, fehler := d.naehte.Liste.KartenFiltern(
		anfrage.Context(), setID, anfrage.URL.Query().Get("anfrage"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ids := make([]string, 0, len(kennungen))
	for _, id := range kennungen {
		ids = append(ids, strconv.FormatInt(id, 10))
	}

	ja(schreiben, http.StatusOK, map[string]any{"kartenIDs": ids})
}
