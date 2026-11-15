package mcp

// **Das Kostenjournal über MCP — lesend** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Warum ein Modell den Stand sehen soll
//
// **Weil es sonst gegen eine Wand läuft, die es nicht sieht.** Ein Modell, das
// einen bezahlten Auftrag anstößt und eine Absage bekommt, kann nichts damit
// anfangen; eines, das vorher weiß, wie viel übrig ist, kann sich einteilen.
//
// # Und warum es die Grenze nicht setzen darf
//
// **Weil die Grenze die Freigabe ist.** Wer sie selbst höhersetzen könnte,
// hätte keine — das ist dasselbe `bestaetigt`-Feld, das kein Nachweis sein
// soll ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).

import (
	"context"
	"encoding/json"
)

// Betrag ist Geld in der kleinsten Einheit seiner Währung.
type Betrag struct {
	Wert     int64  `json:"wert"`
	Waehrung string `json:"waehrung"`
}

// Haushaltsstand sind die vier Zahlen eines Kontos.
type Haushaltsstand struct {
	Konto      string `json:"konto"`
	Zeitraum   string `json:"zeitraum"`
	Grenze     Betrag `json:"grenze"`
	Reserviert Betrag `json:"reserviert"`
	Verbraucht Betrag `json:"verbraucht"`
	Unklar     Betrag `json:"unklar"`
	Verfuegbar Betrag `json:"verfuegbar"`
}

// Kostenauskunft ist die Naht zur Anwendung.
type Kostenauskunft interface {
	Kostenstand(ctx context.Context, konto, zeitraum string) (Haushaltsstand, error)
}

func kostenwerkzeug(naht Kostenauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.kostenstand", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, wie viel Geld an einem Anbieterkonto noch " +
			"freigegeben ist — **ändert nichts**. »zeitraum« ist »tag« oder " +
			"»monat«; leer heißt Tag. Alle Beträge stehen in der **kleinsten " +
			"Einheit** ihrer Währung: 1250 EUR-Cent sind 12,50 €. " +
			"»reserviert« ist angemeldet und noch offen, »verbraucht« " +
			"abgerechnet, »unklar« lief in eine Frist NACH der möglichen " +
			"Annahme — es zählt wie verbraucht, bis ein Mensch es klärt. " +
			"»verfuegbar« ist, was übrig bleibt. **Das ist eine lokale " +
			"Grenze**: wer denselben Schlüssel woanders benutzt, verbraucht " +
			"Kontingent, von dem dieses Journal nichts weiß. GESETZT WIRD " +
			"HIER NICHTS: eine Grenze ist die Freigabe eines Menschen.",
		Schema: objekt(map[string]any{
			"konto":    text("das Anbieterkonto"),
			"zeitraum": text("»tag« oder »monat«; leer heißt Tag"),
		}, "konto"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var eingabe struct {
				Konto    string `json:"konto"`
				Zeitraum string `json:"zeitraum"`
			}

			if fehler := json.Unmarshal(roh, &eingabe); fehler != nil {
				return "", fehler
			}

			stand, fehler := naht.Kostenstand(ctx, eingabe.Konto, eingabe.Zeitraum)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(stand)
		},
	}}
}
