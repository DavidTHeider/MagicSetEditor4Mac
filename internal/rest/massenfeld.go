// **Ein Feld auf vielen Karten setzen — über den Dienst**
// (#1390).
//
// # Warum ein eigener Endpunkt und keine Schleife über `PUT …/felder/{feld}`
//
// **Weil die Atomarität die Fähigkeit ist.** Wer ein Feld über zweihundert
// Karten ändert, macht sonst zweihundert Anfragen; bricht die Verbindung in
// der Mitte, ist das Set halb geändert, und der Aufrufer weiß nicht, welche
// Hälfte. Dieser Weg schreibt in **einer** Transaktion.

package rest

import (
	"net/http"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// felderSetzen setzt ein Feld auf mehreren Karten eines Sets.
//
// **`POST` und nicht `PUT`.** `PUT` sagt „dieser eine Ort trägt danach diesen
// Wert"; hier tragen ihn viele, und der Ort im Weg ist das **Set**. Das
// Gegenstück im Fenster heißt Massenänderung und ist ein Vorgang, kein Ort.
func (d *Dienst) felderSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	if !gutesSet {
		return
	}

	var rumpf struct {
		KarteIDs []int64              `json:"karteIDs"`
		Feld     string               `json:"feld"`
		Wert     string               `json:"wert"`
		Erwartet []domain.Feldvorwert `json:"erwartet"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	// **Mit Erwartung geht es den vergleichenden Weg** (U6, #1417): je Karte
	// der Altwert aus der Vorschau, und der Bestand vergleicht, bevor er
	// schreibt. Eine Erwartung für eine Karte, die nicht in `karteIDs` steht,
	// ist ein Bedienfehler — sonst schriebe der Ruf Karten, die er nicht nennt.
	var (
		vorgang string
		fehler  error
	)

	switch {
	case len(rumpf.Erwartet) == 0:
		vorgang, fehler = d.naehte.Schreiber.FelderSetzen(
			anfrage.Context(), setID, rumpf.KarteIDs, rumpf.Feld, rumpf.Wert)
	case !erwartungInAuswahl(rumpf.Erwartet, rumpf.KarteIDs):
		nein(schreiben, http.StatusBadRequest, "validierung",
			"»erwartet« nennt eine Karte, die nicht in »karteIDs« steht.")

		return
	default:
		vorgang, fehler = d.naehte.Schreiber.FelderSetzenMitErwartung(
			anfrage.Context(), setID, rumpf.Erwartet, rumpf.Feld, rumpf.Wert)
	}

	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Die Zahl und nicht die Liste.** Welche Karten gemeint waren, weiß der
	// Aufrufer — er hat sie geschickt. Was er nicht wusste, ist, ob alle
	// durchgingen; genau das steht hier.
	//
	// **Und der Vorgang** (#1393): mit ihm nimmt
	// `POST /sets/{setID}/zuruecknehmen` genau diese Änderung zurück.und
	// nicht den jüngsten Schritt des Sets.
	ja(schreiben, http.StatusOK, map[string]any{
		"geaendert": len(rumpf.KarteIDs), "vorgang": vorgang,
	})
}

// felderVorschau sagt, was ein Feldwert auf den Karten hieße — `POST`, weil
// die Frage einen Rumpf braucht, und trotzdem lesend (U6, #1417).
func (d *Dienst) felderVorschau(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	if !gutesSet {
		return
	}

	var rumpf struct {
		KarteIDs []int64 `json:"karteIDs"`
		Feld     string  `json:"feld"`
		Wert     string  `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorschau, fehler := d.naehte.Schreiber.FelderVorschau(
		anfrage.Context(), setID, rumpf.KarteIDs, rumpf.Feld, rumpf.Wert)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorschauDTO(vorschau))
}

// vorschauDTO übersetzt die Vorschau: Kennungen als Zeichenketten wie überall
// an dieser Fläche, und `zeilen` nie `null`.
func vorschauDTO(vorschau domain.Massenvorschau) map[string]any {
	zeilen := make([]map[string]any, 0, len(vorschau.Zeilen))
	for _, zeile := range vorschau.Zeilen {
		zeilen = append(zeilen, map[string]any{
			"karteID":    strconv.FormatInt(zeile.KarteID, 10),
			"name":       zeile.Name,
			"alt":        zeile.Alt,
			"gabAlt":     zeile.GabAlt,
			"neu":        zeile.Neu,
			"ausschluss": zeile.Ausschluss,
		})
	}

	return map[string]any{"feld": vorschau.Feld, "wert": vorschau.Wert, "zeilen": zeilen}
}

// erwartungInAuswahl sagt, ob jede Erwartung eine Karte der Auswahl nennt.
func erwartungInAuswahl(erwartet []domain.Feldvorwert, karteIDs []int64) bool {
	genannt := make(map[int64]bool, len(karteIDs))
	for _, karteID := range karteIDs {
		genannt[karteID] = true
	}

	for _, vorwert := range erwartet {
		if !genannt[vorwert.KarteID] {
			return false
		}
	}

	return true
}
