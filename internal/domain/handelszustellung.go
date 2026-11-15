package domain

// **Was mit einer zweiten Zustellung geschieht** — Werkstück `T5` von Phase
// 5.6 (#1470).
//
// # Warum es drei Ausgänge gibt und nicht zwei
//
// **Weil „schon gesehen" zwei verschiedene Dinge heißt.** Dieselbe Nachricht
// noch einmal ist eine **Wiederholung**: das Netz hat verdoppelt, und die
// richtige Antwort ist die von vorhin. Dieselbe **Operation** mit anderem
// Inhalt ist etwas anderes — zwei Seiten haben gegen dieselbe Revision
// verschiedene Dinge gewollt, und das gehört einem Menschen gesagt.
//
// **Die zweite stillschweigend gewinnen zu lassen wäre das Gefährlichste von
// allem**: der erste Stand verschwände, und niemand erführe, dass es ihn gab.
//
// # Warum die Uhr hier nicht entscheidet
//
// **Weil sie nicht unterschrieben ist** ([Handelsumschlag.Gesendet]). Wer aus
// ihr einen Replay-Schutz machte, verließe sich auf ein Feld, das jeder
// ändern kann. Was wirklich schützt, ist zweierlei: der
// **Idempotenzschlüssel**, solange sein Eintrag steht, und darunter die
// **Revision** — ein Vorschlag gegen eine Revision, die weitergezogen ist,
// wird von [Handelsangebot.Nimmt] abgewiesen, gleich wie alt er ist.

import "time"

// Idempotenzfrist ist, wie lange ein behandelter Schlüssel stehen bleibt.
//
// **Dreißig Tage, und die Zahl ist eine Entscheidung.** Sie muss länger sein,
// als eine Nachricht unterwegs sein darf: wer den Eintrag früher wegräumt,
// lässt dieselbe Wirkung ein zweites Mal zu, sobald jemand eine alte
// Nachricht noch einmal schickt.
//
// **Darunter liegt die Revision als zweite Sicherung.** Sie ist die
// eigentliche: ein Grabstein ist eine Abkürzung, die Revision eine Regel.
const Idempotenzfrist = 30 * 24 * time.Hour

// Inboxeintrag ist, was von einer behandelten Nachricht bleibt.
//
// **Das Ergebnis steht dabei**, und das ist der ganze Zweck: eine zweite
// Zustellung bekommt dieselbe Antwort, ohne dass irgendetwas ein zweites Mal
// geschieht.
type Inboxeintrag struct {
	// Schluessel ist der Idempotenzschlüssel
	// ([Handelsumschlag.Idempotenzschluessel]).
	Schluessel string

	// Nutzlastabdruck ist der Hash über Umschlag **und** Nutzlast.
	//
	// **Über beides.** Ein Abdruck nur über den Kopf machte zwei Nachrichten
	// mit verschiedenem Inhalt ununterscheidbar — und genau deren Unterschied
	// ist der Konflikt.
	Nutzlastabdruck string

	// Ergebnis ist die Antwort, die schon gegeben wurde.
	Ergebnis string

	// Behandelt ist die **eigene** Uhr, nicht die des Absenders.
	Behandelt time.Time
}

// Abgelaufen sagt, ob dieser Eintrag weggeräumt werden darf.
//
// **Auf der Frist steht er noch.** Eine Grenze, die ihren eigenen Wert
// einschließt, räumt eine Nachricht weg, die im selben Augenblick ankommt.
func (e Inboxeintrag) Abgelaufen(jetzt time.Time) bool {
	return jetzt.Sub(e.Behandelt) > Idempotenzfrist
}

// Zustellungsart sagt, was mit einer Zustellung zu tun ist.
type Zustellungsart string

// Die drei Ausgänge.
const (
	// ZustellungNeu heißt: noch nicht behandelt — ausführen.
	ZustellungNeu Zustellungsart = "neu"

	// ZustellungBekannt heißt: dieselbe Nachricht noch einmal — die Antwort
	// von vorhin, und **nichts** geschieht ein zweites Mal.
	ZustellungBekannt Zustellungsart = "bekannt"

	// ZustellungKonflikt heißt: dieselbe Operation gegen dieselbe Revision,
	// aber ein anderer Inhalt — das gehört einem Menschen gesagt.
	ZustellungKonflikt Zustellungsart = "konflikt"
)

// Zustellung sagt, was mit dieser Nachricht zu tun ist.
//
// **Der Aufrufer bringt den vorigen Eintrag mit**, statt dass diese Regel ihn
// suchte: sie gehört in die **eine** Transaktion, in der auch die Wirkung
// steht, und eine Domäne, die nachschlägt, kennt keine Transaktion.
func Zustellung(vorher *Inboxeintrag, nutzlastabdruck string) Zustellungsart {
	if vorher == nil {
		return ZustellungNeu
	}

	if vorher.Nutzlastabdruck == nutzlastabdruck {
		return ZustellungBekannt
	}

	return ZustellungKonflikt
}
