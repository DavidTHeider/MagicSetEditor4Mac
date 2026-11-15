package domain

// **Der Stand einer Karte** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Wofür es ihn gibt
//
// **Damit ein Vorschlag zu dem Stand gehört, aus dem er stammt.** A6 friert
// beim Anlegen eine Revision ein und vergleicht sie beim Übernehmen
// ([service.KIAuftragswerk]) — das ist §7, Regel 2, über eine Frist von
// Stunden statt Millisekunden.
//
// # Warum gerechnet und nicht abgelesen
//
// **Weil eine Karte keinen Zeitstempel trägt.** Der Bestand führt Felder und
// Werte; wann sie zuletzt jemand angefasst hat, steht nirgends. Eine Spalte
// dafür wäre eine Wanderung für eine Frage, die eine Prüfsumme genauso
// beantwortet — und die Prüfsumme ist die ehrlichere Antwort: sie sagt, ob
// sich **etwas geändert hat**, nicht, ob jemand gespeichert hat.

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
)

// Stand ist die Prüfsumme über Kennung und Felder dieser Karte.
//
// **Die Reihenfolge zählt nicht**: die Felder einer Karte sind eine Menge, und
// ein Stand, der sich mit der Sortierung änderte, meldete einen Wechsel, wo
// keiner war.
//
// **Und die Trennung zählt doch**: Name und Wert werden durch ein Zeichen
// getrennt, das in keinem von beiden vorkommt. Ohne es gäben `{ab: c}` und
// `{a: bc}` denselben Stand.
func (c CardRow) Stand() string {
	teile := make([]string, 0, len(c.Values)+1)
	teile = append(teile, strconv.FormatInt(c.ID, 10))

	for _, wert := range c.Values {
		teile = append(teile, wert.Field+"\x1f"+wert.Text)
	}

	// **Die Kennung steht vorn und wird mitsortiert** — sie ist eine Zahl
	// und landet ohnehin vor jedem Feldnamen, der mit einem Buchstaben
	// anfängt; wichtig ist nur, dass zwei Karten mit gleichem Inhalt und
	// verschiedener Kennung nicht denselben Stand bekommen.
	slices.Sort(teile[1:])

	summe := sha256.Sum256([]byte(strings.Join(teile, "\x1e")))

	return hex.EncodeToString(summe[:])
}
