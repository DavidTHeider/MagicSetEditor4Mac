package service

// **Was eine Peerverbindung ist — und was sie nicht ist** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Warum die Schnittstellen hier stehen und nicht beim Transport
//
// **Weil `service` sie braucht und nicht der Adapter sie anbietet** (§2). Wer
// sie beim Transport definierte, hätte die Fachschicht an libp2p gebunden —
// und N3 (Pion) wäre dann eine zweite Fachschicht.
//
// **Und weil es zwei Transporte gibt.** Derselbe Ablauf läuft über beide; was
// sie unterscheidet, ist, wie Bytes über die Leitung gehen, und das steht
// hinter diesen drei Zeilen.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Zustellung sagt, was mit einer Nachricht geschieht, wenn es eng wird.
type Zustellung int

const (
	// Zuverlaessig heißt: sie kommt an, oder der Absender erfährt es.
	//
	// **Der Absender wartet**, wenn die Warteschlange voll ist — das ist der
	// Rückstau (»backpressure«) aus dem Issue. Ohne ihn wüchse sie, bis der
	// Speicher es merkt.
	Zuverlaessig Zustellung = iota

	// Verwerfbar heißt: der neuere Stand ersetzt den älteren.
	//
	// **Für Präsenz.** „Ich sehe gerade Karte 12 an" ist in dem Moment
	// wertlos, in dem „ich sehe Karte 13 an" vorliegt. Wer sie zuverlässig
	// zustellte, hielte eine Leitung für einen Stand auf, den niemand mehr
	// will.
	Verwerfbar
)

// Verbindung ist, was ein Transportadapter liefert.
//
// **Bytes und sonst nichts.** Was eine Nachricht bedeutet, entscheidet
// `internal/peer`; was sie darf, entscheidet [Kontaktwerk]. Ein Adapter, der
// eine Hülle läse, wäre eine dritte Stelle mit derselben Regel.
type Verbindung interface {
	// Kanal ist die Bindung dieser einen Verbindung.
	//
	// **Wer da ist, sagt der Handschlag.** Hier stand: „und das ist alles,
	// was sie über die Gegenseite sagt — ein Transport, der es ebenfalls
	// sagte, wäre eine zweite Quelle für dieselbe Frage." Die zweite Quelle
	// gab es aber schon — das Zeugnis im TLS-Handschlag —, und niemand hielt
	// beide gegeneinander (SEC-16). Sie entscheidet nicht, **wer** da ist;
	// sie darf dem Handschlag nur nicht widersprechen ([Gegenstelle]).
	//
	// **Je Verbindung eine andere.** Darauf ruht der Schutz gegen das
	// Wiedereinspielen über Verbindungen hinweg; der Satz steht bei
	// [peer.Nonceliste], und hier steht seine Forderung.
	Kanal() string

	// Gegenstelle ist die Kennung aus dem Zeugnis der Gegenseite — leer, wo
	// der Transport keines kennt (eine Röhre im selben Prozess). Ist sie
	// gesetzt, muss der Handschlag **dieselbe** Kennung nennen (SEC-16,
	// #2145).
	Gegenstelle() peer.Kennung

	// Sende schickt eine fertige Nachricht.
	Sende(ctx context.Context, nachricht []byte) error

	// Empfange liest die nächste Nachricht.
	//
	// # Warum sie hier steht, obwohl sie erst später dazukam
	//
	// **Weil ein Handschlag ein Wechsel ist.** Diese Schnittstelle hatte in
	// N1.5 vier Methoden und konnte nur schreiben; beim ersten wirklichen
	// Gebrauch — dem Handschlag aus N1.3 — fiel auf, dass damit keine
	// Sitzung zustande kommt. Eine Leitung, auf der man nur senden kann,
	// trägt keinen Wechsel (N2,
	// #1455).
	//
	// **Ein Ruf, eine Nachricht.** Wer einen Strom läse, bekäme irgendeine
	// Aufteilung; die Rahmung liegt beim Adapter.
	Empfange(ctx context.Context) ([]byte, error)

	// Schliesse beendet die Verbindung.
	Schliesse() error
}

// Waehler baut Verbindungen — libp2p (N2) oder Pion (N3).
type Waehler interface {
	Waehle(
		ctx context.Context, kennung peer.Kennung, protokoll string,
	) (Verbindung, error)
}

// Ausweisstelle gibt die Identität, mit der dieses Profil unterschreibt.
//
// **Eine Schnittstelle und kein [Peerwerk]**: gebraucht wird ein Ausweis und
// keine Fläche — und ein Werk, das ein anderes ganz kennt, ist die Naht, an
// der später beide hängen.
type Ausweisstelle interface {
	Ausweis(ctx context.Context) (peer.Identitaet, error)
}

// Die beiden Rollen eines Handschlags.
//
// **Sie benennen, wer angefangen hat, und sonst nichts.** Was eine Rolle
// **darf**, steht im Zweck; wer hier eine dritte braucht, braucht einen
// eigenen Zweck.
const (
	RolleAnbietend = "anbietend"
	RolleAnnehmend = "annehmend"
)

// Rechteauskunft ist, was die Sitzungsverwaltung von den Kontakten braucht.
//
// **Eine Schnittstelle und kein [Kontaktwerk]**: gebraucht wird ein Satz und
// keine Fläche — und ein Werk, das ein anderes ganz kennt, ist die Naht, an
// der später beide hängen.
type Rechteauskunft interface {
	Darf(
		ctx context.Context, kennung peer.Kennung, zweck domain.Zweck,
	) (bool, error)
}

// Sitzungsstand ist, was über eine offene Sitzung feststeht.
type Sitzungsstand struct {
	Kennung       string
	Gegenueber    peer.Kennung
	Fingerabdruck string
	Zweck         domain.Zweck
	Kanal         string
	Seit          time.Time

	// Wartend ist, was noch nicht hinausgegangen ist.
	Wartend int

	// Verworfen zählt, was die Warteschlange fallen ließ.
	//
	// **Sichtbar und nicht still.** Eine verworfene Präsenznachricht ist
	// richtig; eine Zahl, die niemand sieht, machte aus „richtig" ein
	// „unbemerkt".
	Verworfen int
}
