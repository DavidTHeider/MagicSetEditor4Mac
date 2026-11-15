// **Der portable Umschlag über den Dienst** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Drei Wege, und zwei davon sind getrennt, obwohl sie zusammengehören
//
// **Ansehen und Übernehmen bleiben zwei Aufrufe** — #1457 verlangt die
// fachliche Übernahmefreigabe, und über HTTP ist sie derselbe zweite Ruf, den
// ein Mensch im Fenster als zweiten Knopf sieht. Ein `POST`, das erst zeigte
// und dann schriebe, nähme sie entgegen, nachdem sie gegenstandslos geworden
// wäre.
//
// # Und warum hier ein Pfad steht und kein Rumpf mit Bytes
//
// **Weil dieser Dienst lokal läuft** ([0046](../../docs/entscheidungen/0046-der-dienst-bindet-auf-die-schleife.md)).
// Er bindet auf die Schleifenadresse; wer ihn ruft, sitzt an derselben
// Maschine und kann eine Datei nennen. Ein Umschlag im Rumpf wäre ein
// Mehrfachformular für ein Problem, das es hier nicht gibt.

package rest

import (
	"context"
	"net/http"
)

// Umschlagwerkstatt ist, was die drei Endpunkte dieser Datei brauchen.
type Umschlagwerkstatt interface {
	UmschlagSchnueren(ctx context.Context, art, kennung, ziel string) (Umschlagangaben, error)
	UmschlagAnsehen(ctx context.Context, pfad string) (Umschlagangaben, error)
	UmschlagUebernehmen(ctx context.Context, pfad string) (Umschlagannahme, error)
}

// Umschlagangaben sind, was ein Umschlag über sich hergibt.
type Umschlagangaben struct {
	// Urheber ist die Peerkennung dessen, der unterschrieben hat.
	//
	// **Sie ist kein Vertrauen** — sie sagt **wer**, nicht ob man ihm glauben
	// sollte.
	Urheber       string           `json:"urheber,omitempty"`
	Fingerabdruck string           `json:"fingerabdruck,omitempty"`
	Objekte       []Umschlagobjekt `json:"objekte"`
}

// Umschlagobjekt ist ein Stück im Umschlag.
type Umschlagobjekt struct {
	Art         string `json:"art"`
	Bezeichnung string `json:"bezeichnung"`
	Groesse     int64  `json:"groesse"`

	// Braucht nennt, was dieses Objekt zum Vollständigsein braucht.
	//
	// **Es steht hier, weil #1457 es verlangt**: „Import zeigt
	// Objekte/**Abhängigkeiten**/Konflikte". Eine Fläche, die es wegließe,
	// zeigte weniger als das Fenster.
	Braucht []string `json:"braucht"`
}

// Umschlagannahme ist, was aus einer Übernahme wird.
type Umschlagannahme struct {
	SetID   string   `json:"setID"`
	Karten  int      `json:"karten"`
	Befunde []string `json:"befunde"`
}

// umschlagSchnueren legt einen Gegenstand als Umschlag ab.
//
// # Warum `setID` bleibt
//
// **Weil es die Zusage war** (0033, Regel 6 und 8). `T7`
// (#1472)
// bringt vier weitere Arten, und `kennung` heißt jetzt, was gemeint ist;
// `setID` wird weiter gelesen und bedeutet `art: "set"`. Wegfallen kann es
// frühestens mit 1.0.0 — die Änderung ist **erweiternd** und nicht brechend.
func (d *Dienst) umschlagSchnueren(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Art     string `json:"art"`
		Kennung string `json:"kennung"`
		SetID   string `json:"setID"`
		Ziel    string `json:"ziel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	// **Die alte Marke gewinnt nicht** — wer beide schickt, meint die neue;
	// sonst bekäme ein Aufrufer, der umgestellt hat und die alte vergaß, eine
	// andere Antwort als die, die er liest.
	//
	// **Und die Klage nennt die Marke, die er geschickt hat.** Wer `setID`
	// schickt und dort nichts hineinschreibt, soll `setID` lesen — eine
	// Meldung über ein Feld, das er nicht kennt, sagt ihm nichts.
	feld := "kennung"

	if rumpf.Kennung == "" {
		rumpf.Kennung, rumpf.Art, feld = rumpf.SetID, "set", "setID"
	}

	kennung, gut := pflichttext(schreiben, feld, rumpf.Kennung)
	if !gut {
		return
	}

	ziel, gut := pflichttext(schreiben, "ziel", rumpf.Ziel)
	if !gut {
		return
	}

	angaben, fehler := d.naehte.Umschlag.UmschlagSchnueren(anfrage.Context(),
		rumpf.Art, kennung, ziel)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, angaben)
}

// umschlagSchauen sagt, was in einem Umschlag steht — **ohne** ihn zu nehmen.
func (d *Dienst) umschlagSchauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	pfad, gut := pflichttext(schreiben, "pfad", anfrage.URL.Query().Get("pfad"))
	if !gut {
		return
	}

	angaben, fehler := d.naehte.Umschlag.UmschlagAnsehen(anfrage.Context(), pfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}

// umschlagUebernehmen schreibt einen Umschlag in den Bestand.
func (d *Dienst) umschlagUebernehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Pfad string `json:"pfad"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	pfad, gut := pflichttext(schreiben, "pfad", rumpf.Pfad)
	if !gut {
		return
	}

	annahme, fehler := d.naehte.Umschlag.UmschlagUebernehmen(anfrage.Context(), pfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, annahme)
}
