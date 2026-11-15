package rest

// **Die eigene Peeridentität im Dienst** — Werkstück N1 von Phase 5.4
// (#1454).
//
// Drei Wege, und die Methode sagt, welcher was tut:
//
//	GET  /peer            womit sich dieses Profil ausweist — ohne Nebenwirkung
//	POST /peer            die erste Identität anlegen, wenn es keine gibt
//	POST /peer/rotation   sie ablösen; jeder Kontakt muss neu prüfen
//
// **Zwei `POST` und kein `PUT`**: keiner von beiden ist idempotent im Sinne
// der Wiederholung — der zweite legt jedes Mal eine neue Identität an, und
// das ist seine Aufgabe.

import (
	"context"
	"net/http"
)

// Peernaht ist die Naht zu `internal/app`.
type Peernaht interface {
	Peeridentitaet(ctx context.Context) (Peeridentitaet, error)
	PeeridentitaetSicherstellen(ctx context.Context) (Peeridentitaet, error)
	PeeridentitaetRotieren(ctx context.Context) (Peeridentitaet, error)
}

// Peeridentitaet ist die Identität, wie der Dienst sie nennt.
type Peeridentitaet struct {
	Stand               string   `json:"stand"`
	Kennung             string   `json:"kennung"`
	Fingerabdruck       string   `json:"fingerabdruck"`
	Seit                string   `json:"seit"`
	Vorher              string   `json:"vorher"`
	VorherFingerabdruck string   `json:"vorherFingerabdruck"`
	Befunde             []string `json:"befunde"`
}

// peerZeigen beantwortet `GET /peer`.
func (d *Dienst) peerZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.peerAntworten(schreiben, anfrage, d.naehte.Peer.Peeridentitaet)
}

// peerAnlegen beantwortet `POST /peer`.
func (d *Dienst) peerAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.peerAntworten(schreiben, anfrage, d.naehte.Peer.PeeridentitaetSicherstellen)
}

// peerRotieren beantwortet `POST /peer/rotation`.
func (d *Dienst) peerRotieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.peerAntworten(schreiben, anfrage, d.naehte.Peer.PeeridentitaetRotieren)
}

// peerAntworten fährt einen der drei Wege und schreibt das Ergebnis.
func (d *Dienst) peerAntworten(
	schreiben http.ResponseWriter, anfrage *http.Request,
	fahre func(context.Context) (Peeridentitaet, error),
) {
	identitaet, fehler := fahre(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Eine leere Liste und kein `null`** — sonst verschwände sie aus der
	// Antwort, und wer sie liest, hielte „verloren" für einen Zustand ohne
	// Grund (§3, „zwischen den Prüfebenen").
	if identitaet.Befunde == nil {
		identitaet.Befunde = []string{}
	}

	ja(schreiben, http.StatusOK, identitaet)
}
