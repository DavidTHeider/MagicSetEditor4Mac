// **Die Dateien eines installierten Pakets über die Rückschleife** —
// #1164.
//
// # Warum `PUT` und nicht `POST`
//
// Gesetzt wird eine benannte Datei auf einen Inhalt. Zweimal derselbe Aufruf
// mit demselben Stand ist **nicht** dasselbe — der zweite wird abgewiesen,
// weil der Stand nicht mehr stimmt. Das ist kein Widerspruch zu `PUT`: der
// Stand ist die Bedingung, nicht der Inhalt.

package rest

import (
	"context"
	"net/http"
)

// Paketwerkstatt ist, was der Dienst für Paketdateien braucht.
type Paketwerkstatt interface {
	DateienUndBeiwerk(ctx context.Context, paket string) (dateien, beiwerk []string, fehler error)
	Lesen(ctx context.Context, paket, datei string) (inhalt, stand string, fehler error)
	Schreiben(ctx context.Context, paket, datei, inhalt, erwartet string) (string, error)

	// Pruefen nennt Zeile, Spalte und Satz des ersten Syntaxfehlers in den
	// Skripten eines Paketdateitextes — **Zeile 0**, wenn keiner darin steht.
	Pruefen(inhalt string) (zeile, spalte int, meldung string)

	// Der Dateibaum (FMT-02, #1691) — **ohne Einfügen von der Platte**: über
	// das Netz wäre es ein Weg, eine beliebige Datei des Rechners in ein Paket
	// zu legen und über `GET …/dateien/{datei}` wieder hinauszulesen.
	//
	// **Mit `Datei` im Namen**: ein Double der Tests bedient auch die
	// Scryfall-Anlage, und die heißt schon `Anlegen`.
	DateiAnlegen(ctx context.Context, paket, datei, inhalt string) (string, error)
	DateiUmbenennen(ctx context.Context, paket, alt, neu, erwartet string) error
	DateiEntfernen(ctx context.Context, paket, datei, erwartet string) error
}

// paketdateien zählt die bearbeitbaren Dateien eines Pakets auf.
func (d *Dienst) paketdateien(schreiben http.ResponseWriter, anfrage *http.Request) {
	dateien, beiwerk, fehler := d.naehte.Werkstatt.DateienUndBeiwerk(anfrage.Context(),
		anfrage.PathValue("paket"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"dateien": leerStattNil(dateien), "beiwerk": leerStattNil(beiwerk)})
}

// paketdateiLesen gibt eine Paketdatei samt ihrem Stand.
func (d *Dienst) paketdateiLesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	inhalt, stand, fehler := d.naehte.Werkstatt.Lesen(anfrage.Context(),
		anfrage.PathValue("paket"), anfrage.PathValue("datei"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"inhalt": inhalt, "stand": stand})
}

// paketdateiPruefen prüft die Skripte eines Paketdateitextes.
//
// **`POST` und nicht `GET`.** Der zu prüfende Text steht im Rumpf — er kann
// mehrere hundert Zeilen haben, und in einer Adresse hätte er nichts zu
// suchen. Geändert wird dabei nichts; das steht in der Erklärung und in
// `docs/FAEHIGKEITEN.md`, denn die Methode allein sagt es hier nicht.
//
// **Der Text kommt mit und wird nicht gelesen**: eine Änderung soll sich
// prüfen lassen, bevor sie geschrieben wird. Das ist derselbe Weg wie im
// Fenster, wo beim Tippen geprüft wird.
func (d *Dienst) paketdateiPruefen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Inhalt string `json:"inhalt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	zeile, spalte, meldung := d.naehte.Werkstatt.Pruefen(rumpf.Inhalt)
	if zeile == 0 {
		ja(schreiben, http.StatusOK, map[string]any{"gefunden": false})

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"gefunden": true, "zeile": zeile, "spalte": spalte, "meldung": meldung,
	})
}

// paketdateiSchreiben setzt eine Paketdatei zurück.
func (d *Dienst) paketdateiSchreiben(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Inhalt string `json:"inhalt"`
		Stand  string `json:"stand"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	stand, fehler := d.naehte.Werkstatt.Schreiben(anfrage.Context(),
		anfrage.PathValue("paket"), anfrage.PathValue("datei"), rumpf.Inhalt, rumpf.Stand)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"stand": stand})
}

// leerStattNil gibt eine leere Liste statt nil — im JSON stünde sonst `null`.
func leerStattNil(liste []string) []string {
	if liste == nil {
		return []string{}
	}

	return liste
}
