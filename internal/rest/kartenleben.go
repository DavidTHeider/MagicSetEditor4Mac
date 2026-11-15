// Der Lebenszyklus einer Karte und der Volltext über REST
// (#1107).
//
// **Sechs Fähigkeiten, die der MCP-Server längst konnte.** Die Nähte dahinter
// sind dieselben — `app.Schreibnaht` und `storage.NeuerSetLeser` —; gefehlt
// hat allein, dass jemand sie von hier aus ruft. Das ist die leere Zelle, die
// #1107 zählt: keine Entscheidung gegen die Fähigkeit, sondern gar keine.

package rest

import (
	"net/http"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// karteAnlegen legt eine leere Karte an.
//
// **201 und die Kennung**, wie bei jeder anderen Anlage hier: der Aufrufer
// braucht sie, um im nächsten Zug ein Feld zu setzen.
func (d *Dienst) karteAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	neue, fehler := d.naehte.Schreiber.KarteAnlegen(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{
		"karteID": strconv.FormatInt(neue, 10),
	})
}

// karteKopieren legt eine Karte samt ihren Werten noch einmal an.
func (d *Dienst) karteKopieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	neue, fehler := d.naehte.Schreiber.KarteKopieren(anfrage.Context(), setID, karteID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{
		"karteID": strconv.FormatInt(neue, 10),
	})
}

// karteLoeschen nimmt eine Karte aus dem Set.
//
// **200 mit leerem Rumpf und nicht 204.** Jede andere Antwort dieses Dienstes
// ist ein JSON-Objekt; ein einzelner Weg ohne Rumpf zwänge jeden Aufrufer zu
// einer Sonderbehandlung für genau ihn.
func (d *Dienst) karteLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	// **`?trotzdem=true` bestätigt ein Löschen, das einen Deckplatz mitnimmt**
	// (#1369).
	// Ohne die Angabe weist der Dienst eine Karte ab, die in einem Deck liegt,
	// und nennt die Decks.
	trotzdem := anfrage.URL.Query().Get("trotzdem") == "true"

	if fehler := d.naehte.Schreiber.KarteLoeschen(
		anfrage.Context(), setID, karteID, trotzdem); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}

// kartenTauschen dreht zwei Karten in der Reihenfolge.
func (d *Dienst) kartenTauschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	// **Die Kennungen als Zeichenketten im Rumpf**, wie sie auch aus jeder
	// Antwort dieses Dienstes kommen: eine Kartenkennung als JSON-Zahl wird
	// jenseits von 2^53 stillschweigend ungenau.
	var rumpf struct {
		Eine   string `json:"eine"`
		Andere string `json:"andere"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	eine, guteEine := ausRumpf(schreiben, "eine", rumpf.Eine)
	if !guteEine {
		return
	}

	andere, guteAndere := ausRumpf(schreiben, "andere", rumpf.Andere)
	if !guteAndere {
		return
	}

	if fehler := d.naehte.Schreiber.KartenTauschen(
		anfrage.Context(), setID, eine, andere); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}

// felder nennt die Feldnamen, die in einem Set vorkommen.
func (d *Dienst) felder(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	namen, fehler := d.naehte.Leser.Felder(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Nie `null`.** Ein `nil` aus dem Speicher wird sonst zu `null` im JSON,
	// und ein Aufrufer bekommt statt einer leeren Liste einen Wert, den er
	// nicht durchlaufen kann (§3, „Zwischen den Prüfebenen").
	liste := make([]string, 0, len(namen))
	liste = append(liste, namen...)

	ja(schreiben, http.StatusOK, map[string]any{"felder": liste})
}

// suche findet Kartenstellen über den Volltext.
//
// **`set` ist wahlfrei**, wie im Fenster: ohne es geht die Suche über den
// ganzen Bestand. Ein Pflichtargument machte den häufigen Fall — „wo steht das
// überhaupt?" — zum umständlichen.
func (d *Dienst) suche(schreiben http.ResponseWriter, anfrage *http.Request) {
	text := anfrage.URL.Query().Get("text")
	if text == "" {
		nein(schreiben, http.StatusBadRequest, "validierung", "text fehlt")

		return
	}

	frage := domain.SearchQuery{Text: text, Field: anfrage.URL.Query().Get("feld")}

	if roh := anfrage.URL.Query().Get("set"); roh != "" {
		gelesen, gut := ausRumpf(schreiben, "set", roh)
		if !gut {
			return
		}

		frage.SetID = gelesen
	}

	ergebnis, fehler := d.naehte.Leser.Suche(anfrage.Context(), frage)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Und ob gekürzt wurde, reist mit** — eine stille Kürzung sieht aus wie
	// ein vollständiges Ergebnis.
	ja(schreiben, http.StatusOK, map[string]any{
		"treffer":  alsTreffer(ergebnis.Hits),
		"gekuerzt": ergebnis.Truncated,
	})
}

// alsTreffer übersetzt die gefundenen Stellen.
func alsTreffer(hits []domain.SearchHit) []map[string]any {
	zeilen := make([]map[string]any, 0, len(hits))

	for _, treffer := range hits {
		zeilen = append(zeilen, map[string]any{
			"karteID": strconv.FormatInt(treffer.CardID, 10),
			"setID":   strconv.FormatInt(treffer.SetID, 10),
			"nummer":  treffer.Number,
			"feld":    treffer.Field,
			"wert":    treffer.Value,
		})
	}

	return zeilen
}

// ausRumpf liest eine Kennung, die nicht im Weg stand.
//
// **Neben [kennung] und nicht statt ihr**: jene liest aus dem Pfad, diese aus
// Rumpf oder Abfrage. Beide klagen gleich, damit ein Aufrufer nicht zwei
// Meldungsformen unterscheiden muss.
func ausRumpf(schreiben http.ResponseWriter, name, roh string) (int64, bool) {
	zahl, fehler := strconv.ParseInt(roh, 10, 64)
	if fehler != nil || zahl < 1 {
		nein(schreiben, http.StatusBadRequest, "validierung",
			name+" muss eine Kennung sein, nicht "+strconv.Quote(roh))

		return 0, false
	}

	return zahl, true
}
