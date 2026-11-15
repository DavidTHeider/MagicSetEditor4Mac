package rest

// **Die vier Verwaltungswege einer Partie am Dienst** — PLAY-03.AK-05
// (#1711).
//
// # Warum die Ausfuhr hier einen Text gibt und keinen Pfad
//
// **Weil am anderen Ende der Leitung ein fremder Rechner steht.** Ein Ort
// auf **dieser** Platte wäre dort keine Auskunft — er wäre eine Anweisung an
// jemanden, der sie nicht ausführen kann (§16). Das Fenster und die
// Kommandozeile stehen an derselben Platte und bekommen deshalb dort einen
// Pfad.
//
// # Und warum die Einfuhr den ganzen Zustand annimmt
//
// **Weil eine Partie ohne ihre verdeckten Teile eine andere Lage wäre.** Wer
// hier eine Datei hereinreicht, legt einen **neuen Übungszustand** an — er
// setzt keine fremde Partie fort. Die Kennung ist neu, der Verlauf bleibt
// drüben, und Einwilligungen kommen nicht mit.

import "net/http"

// Partieverwaltungsangaben ist der Ausgang eines dieser vier Wege.
type Partieverwaltungsangaben struct {
	// Kennung ist die Partie, die dabei entstanden ist.
	Kennung string `json:"kennung"`

	// Inhalt ist die Partie als Text — nur bei der Ausfuhr.
	Inhalt string `json:"inhalt,omitempty"`

	// Befunde sind Auffälligkeiten der eingelesenen Datei.
	//
	// **Eine leere Liste und kein `nil`** — „nichts war auffällig" ist eine
	// Antwort, und sie sieht anders aus als keine.
	Befunde []string `json:"befunde"`
}

// partieLoeschen nimmt eine Partie aus dem Bestand.
func (d *Dienst) partieLoeschen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	angaben, fehler := d.naehte.Partien.PartieLoeschen(
		anfrage.Context(), anfrage.PathValue("partieID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}

// partieKopieren legt sie als neuen Übungszustand daneben.
func (d *Dienst) partieKopieren(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	angaben, fehler := d.naehte.Partien.PartieKopieren(
		anfrage.Context(), anfrage.PathValue("partieID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, angaben)
}

// partieAusfuhr gibt die Partie als Text.
func (d *Dienst) partieAusfuhr(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	angaben, fehler := d.naehte.Partien.PartieAusfuhr(
		anfrage.Context(), anfrage.PathValue("partieID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}

// partieEinfuhr legt eine Partie aus einem Text an.
func (d *Dienst) partieEinfuhr(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var rumpf struct {
		Inhalt string `json:"inhalt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	inhalt, gut := pflichttext(schreiben, "inhalt", rumpf.Inhalt)
	if !gut {
		return
	}

	angaben, fehler := d.naehte.Partien.PartieEinfuhr(anfrage.Context(), inhalt)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, angaben)
}
