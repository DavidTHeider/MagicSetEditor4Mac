// Pakete, Kennzahlen, Verlauf und Einstellungen über REST
// (#1107).

package rest

import "net/http"

// pakete frischt den Katalog auf und gibt ihn.
func (d *Dienst) pakete(schreiben http.ResponseWriter, anfrage *http.Request) {
	liste, befunde, fehler := d.naehte.Pakete.Liste(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"pakete": ohneNil(liste), "befunde": ohneNil(befunde),
	})
}

// paketBeziehen holt ein Vorlagenpaket.
//
// **Der Stand ist Pflicht und kein Zweig** (§11): ein Zweig zeigt morgen auf
// etwas anderes, und derselbe Aufruf ergäbe ein anderes Paket.
func (d *Dienst) paketBeziehen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Adresse          string `json:"adresse"`
		Stand            string `json:"stand"`
		Unterverzeichnis string `json:"unterverzeichnis"`
		Name             string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	ergebnis, fehler := d.naehte.Pakete.Beziehen(
		anfrage.Context(), rumpf.Adresse, rumpf.Stand, rumpf.Unterverzeichnis, rumpf.Name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{
		"ort": ergebnis.Ort, "stand": ergebnis.Stand,
		"pruefsumme": ergebnis.Pruefsumme, "uebergangen": ohneNil(ergebnis.Uebergangen),
	})
}

// statistik zählt ein Set aus.
func (d *Dienst) statistik(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	zahlen, fehler := d.naehte.Zahlen.Statistik(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"karten": zahlen.Karten, "felder": zahlen.Felder, "leere": zahlen.Leere,
		"belegung": ohneNil(zahlen.Belegung),
	})
}

// verteilung sagt, wie viele Karten welchen Wert eines Felds tragen.
func (d *Dienst) verteilung(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	werte, fehler := d.naehte.Zahlen.Verteilung(
		anfrage.Context(), setID, anfrage.PathValue("feld"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"verteilung": ohneNil(werte)})
}

// zuruecknehmen macht die jüngste noch geltende Änderung rückgängig.
func (d *Dienst) zuruecknehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.bewege(schreiben, anfrage, true)
}

// verlauf nennt die Schritte eines Sets — der jüngste zuerst (U6, #1417).
func (d *Dienst) verlauf(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	schritte, fehler := d.naehte.Verlauf.Verlauf(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	if schritte == nil {
		schritte = []Verlaufsschritt{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"schritte": schritte})
}

// wiederherstellen führt die jüngste zurückgenommene Änderung wieder aus.
func (d *Dienst) wiederherstellen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.bewege(schreiben, anfrage, false)
}

// bewege führt beide Richtungen — sie unterscheiden sich nur im Aufruf.
//
// **200 und nicht 204, auch wenn nichts zu bewegen war.** „Der Verlauf ist am
// Ende" ist eine Antwort und kein Fehler; sie steht in `bewegt`.
func (d *Dienst) bewege(schreiben http.ResponseWriter, anfrage *http.Request, zurueck bool) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	richtung := d.naehte.Verlauf.Wiederherstellen
	if zurueck {
		richtung = d.naehte.Verlauf.Zuruecknehmen
	}

	// **Der Vorgang steht in der Abfrage und nicht im Rumpf** (#1393): beide
	// Wege sind `POST` ohne Rumpf, und einer, der nur manchmal einen hat,
	// wäre zwei Verträge in einem Pfad. Fehlt er, gilt der jüngste Schritt.
	bewegung, fehler := richtung(
		anfrage.Context(), setID, anfrage.URL.Query().Get("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"bewegt": bewegung.Bewegt, "karteID": bewegung.KarteID,
		"feld": bewegung.Feld, "umfang": bewegung.Umfang, "urheber": bewegung.Urheber,
	})
}

// einstellungen nennt jede Einstellung samt ihren erlaubten Werten.
func (d *Dienst) einstellungen(schreiben http.ResponseWriter, anfrage *http.Request) {
	alle, fehler := d.naehte.Einstellungen.Alle(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"einstellungen": ohneNil(alle)})
}

// einstellungSetzen schreibt eine Einstellung fort.
func (d *Dienst) einstellungSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Wert string `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if fehler := d.naehte.Einstellungen.EinstellungSetzen(
		anfrage.Context(), anfrage.PathValue("schluessel"), rumpf.Wert); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}
