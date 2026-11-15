package domain

// **Die digitale Deckleihe** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Was sie ist und was sie ausdrücklich nicht ist
//
// **Sie ist eine Nutzungsfreigabe und keine Eigentumsübertragung.** Wer
// verleiht, behält sein Deck; wer leiht, darf damit **antreten** — nicht
// besitzen, nicht handeln, nicht in seine Sammlung übernehmen. Der physische
// Bestand ändert sich durch sie um **kein** Exemplar, und das Issue sagt es in
// jeder seiner Abnahmen.
//
// **Und sie ist kein Kopierschutz.** Das Issue grenzt es selbst ab: „DRM,
// beweisbare Löschung beim Empfänger und Verleih physischer Exemplare gehören
// nicht zu dieser digitalen Leihe." Wer die Bytes eines Abzugs einmal hat, hat
// sie; was diese Leihe regelt, ist, **womit dieses Programm eine Partie
// startet** — und das ist eine Zusage unter Mitspielern, keine Sperre gegen
// einen Angreifer. Eine Oberfläche, die mehr verspräche, verspräche etwas, das
// niemand halten kann (§10).
//
// # Warum sie Peerkennungen trägt und keine Profile
//
// **Weil ein Profil dieses Rechners drüben nichts bedeutet.** Das Issue
// verlangt es wörtlich: „Lokale OwnerProfileID/BorrowerProfileID werden nicht
// über die Grenze übertragen." Eine Peerkennung **trägt den Schlüssel**; sie
// ist dieselbe Angabe auf beiden Seiten, und gegen sie prüft die Unterschrift.
//
// # Warum der Abzug über seinen Abdruck gebunden wird
//
// **Weil er sich nicht ändern darf.** [Partiedeck.Abdruck] ist der Hash über
// den eingefrorenen Abzug; wer das Ursprungsdeck danach ändert oder löscht,
// ändert diesen Hash nicht — und genau das ist die zweite Abnahme des Issues.
// Die Leihe nennt **den Abdruck** und nicht die Deckkennung: eine Kennung
// zeigte auf etwas, das sich bewegt.

import (
	"fmt"
	"strings"
	"time"
)

// Leihart sagt, woran eine Leihe hängt.
type Leihart string

// Die zwei Arten, an die eine Leihe gebunden werden kann.
//
// **Zwei und nicht beliebig viele.** Das Issue nennt „Session-/Turnierbindung";
// beides sind Dinge, die dieses Programm **kennt** und deren Kennung es beim
// Start vorliegen hat. Eine dritte Art wäre eine Bindung an etwas, das niemand
// prüfen kann — und eine Prüfung, die nichts prüft, ist keine (§10).
const (
	// LeiheFuerPartie gilt einer Partie.
	LeiheFuerPartie Leihart = "partie"

	// LeiheFuerTurnier gilt einem Turnier — und damit jeder Partie darin.
	LeiheFuerTurnier Leihart = "turnier"
)

// Leihbindung sagt, **woran** eine Leihe hängt.
type Leihbindung struct {
	Art Leihart

	// Woran ist die Kennung — **leer heißt: die nächste dieser Art**.
	//
	// # Warum es das gibt
	//
	// **Weil eine Partie erst entsteht, wenn beide dasitzen.** Die Abnahme
	// sagt „Besitzer verleiht für **nächste** Partie"; wer erst die Kennung
	// abwarten müsste, könnte nicht im Voraus verleihen.
	//
	// **Was das kostet, gehört ausgesprochen**: eine offene Bindung passt auf
	// jede Partie dieser Art, bis sie **verbraucht** ist — und dass sie
	// verbraucht ist, merkt sich die Seite des Entleihers. Ein geändertes
	// Programm könnte das übergehen; das ist derselbe Satz wie oben über den
	// Kopierschutz, und er gilt hier wie dort.
	Woran string
}

// Leihrecht sagt, **wozu** eine Leihe berechtigt.
type Leihrecht string

// Die zwei Rechte, und beide werden geprüft.
//
// **Zwei und keine Liste von Wünschen.** Jedes Recht hier hat in diesem
// Programm eine Stelle, an der es nachgesehen wird; ein drittes, das niemand
// abfragte, stünde als Zusage da, die ihr Rumpf nicht hält (§10).
const (
	// RechtSpielen erlaubt, mit dem Abzug an einer Partie anzutreten.
	RechtSpielen Leihrecht = "spielen"

	// RechtEinreichen erlaubt, ihn als Turnierdeckliste einzureichen.
	RechtEinreichen Leihrecht = "einreichen"
)

// Leihmodus sagt, was mit einer **laufenden** Nutzung geschieht, wenn die
// Leihe endet.
//
// # Warum das vorher vereinbart wird
//
// **Weil das Issue heimliches Wegnehmen ausschließt**: „Ablauf/Widerruf
// verhindert neue Starts und betrifft laufende Nutzung nur nach vorab
// vereinbartem Modus, **niemals heimliches Entfernen von Spielobjekten**."
// Eine Partie, aus der mitten im Zug Karten verschwinden, ist kein Spiel — und
// sie wäre für beide Seiten unerklärlich.
type Leihmodus string

// Die zwei Modi.
const (
	// LaufZuEnde lässt eine laufende Partie zu Ende laufen.
	LaufZuEnde Leihmodus = "zuende"

	// LaufAbbruch beendet sie — **sichtbar und mit Grund**.
	LaufAbbruch Leihmodus = "abbruch"
)

// Deckleihe ist die Freigabe eines eingefrorenen Abzugs an einen Peer.
type Deckleihe struct {
	Kennung string

	// Verleiher und Entleiher sind **Peerkennungen** — der Grund steht im
	// Kopf dieser Datei.
	Verleiher string
	Entleiher string

	// Abdruck ist der Hash des Abzugs ([Partiedeck.Abdruck]).
	Abdruck string

	Bindung Leihbindung
	Rechte  []Leihrecht

	Ausgestellt time.Time
	Bis         time.Time

	Modus Leihmodus

	// Unterschrift ist die des Verleihers über [Deckleihe.Unterschriftsform],
	// hexadezimal. **Sie geht nicht in ihre eigene Form ein.**
	Unterschrift string
}

// Pruefe sagt, ob eine Leihe eine ist.
//
// **Nur die Gestalt und nicht die Unterschrift.** Wer den Schlüssel prüft,
// braucht `internal/peer`, und `domain` kennt es nicht (§2). Die Unterschrift
// prüft `service`, und zwar **bevor** es diese Prüfung ruft.
func (l Deckleihe) Pruefe() error {
	if fehler := l.pruefeBeteiligte(); fehler != nil {
		return fehler
	}

	if l.Abdruck == "" {
		return fmt.Errorf(
			"%w: eine Leihe braucht den Abdruck des Abzugs, den sie meint",
			ErrValidation)
	}

	if fehler := l.pruefeBindung(); fehler != nil {
		return fehler
	}

	if fehler := l.pruefeRechte(); fehler != nil {
		return fehler
	}

	return l.pruefeFristen()
}

// pruefeBeteiligte hält Kennung, Verleiher und Entleiher gegeneinander.
func (l Deckleihe) pruefeBeteiligte() error {
	if l.Kennung == "" {
		return fmt.Errorf("%w: eine Leihe braucht eine Kennung", ErrValidation)
	}

	if l.Verleiher == "" || l.Entleiher == "" {
		return fmt.Errorf(
			"%w: eine Leihe braucht zwei Seiten", ErrValidation)
	}

	// **Sich selbst zu verleihen ist keine Leihe.** Wer es hinnähme, hätte
	// ein Dokument, das nichts erlaubt, was nicht ohnehin gilt — und jemanden,
	// der glaubt, er habe etwas geregelt.
	if l.Verleiher == l.Entleiher {
		return fmt.Errorf(
			"%w: %q verleiht an sich selbst", ErrValidation, l.Verleiher)
	}

	return nil
}

// pruefeBindung hält die Art gegen die Aufzählung.
func (l Deckleihe) pruefeBindung() error {
	if l.Bindung.Art != LeiheFuerPartie && l.Bindung.Art != LeiheFuerTurnier {
		return fmt.Errorf("%w: %q ist keine Art, an die sich binden lässt",
			ErrValidation, l.Bindung.Art)
	}

	return nil
}

// pruefeRechte hält jedes Recht gegen die Aufzählung.
func (l Deckleihe) pruefeRechte() error {
	if len(l.Rechte) == 0 {
		return fmt.Errorf(
			"%w: eine Leihe ohne Recht erlaubt nichts — dann gibt es sie nicht",
			ErrValidation)
	}

	for _, recht := range l.Rechte {
		if recht != RechtSpielen && recht != RechtEinreichen {
			return fmt.Errorf("%w: %q ist kein Recht dieser Leihe",
				ErrValidation, recht)
		}
	}

	return nil
}

// pruefeFristen hält Ausstellung, Ablauf und Modus.
func (l Deckleihe) pruefeFristen() error {
	if l.Ausgestellt.IsZero() || l.Bis.IsZero() {
		return fmt.Errorf(
			"%w: eine Leihe braucht einen Anfang und ein Ende — eine ohne "+
				"Ablauf wäre eine Übertragung", ErrValidation)
	}

	if !l.Bis.After(l.Ausgestellt) {
		return fmt.Errorf(
			"%w: diese Leihe läuft ab, bevor sie gilt", ErrValidation)
	}

	if l.Modus != LaufZuEnde && l.Modus != LaufAbbruch {
		return fmt.Errorf("%w: %q ist kein Modus für eine laufende Nutzung",
			ErrValidation, l.Modus)
	}

	return nil
}

// Darf sagt, ob dieses Recht in der Leihe steht.
func (l Deckleihe) Darf(recht Leihrecht) bool {
	for _, gegeben := range l.Rechte {
		if gegeben == recht {
			return true
		}
	}

	return false
}

// Gilt hält eine Leihe gegen den Fall, in dem sie benutzt werden soll.
//
// # Die vier Abweisungen aus der Abnahme
//
// „anderer Peer/falsche Session/falscher Hash/abgelaufene Signatur nicht" —
// und die vierte ist hier die **Frist**: ob der Schlüssel stimmt, hat `service`
// vorher geprüft.
//
// **Die Reihenfolge ist die Auskunft.** Wer zuerst den Abdruck prüfte,
// antwortete einem fremden Peer mit „falsches Deck" statt mit „nicht für
// dich".
func (l Deckleihe) Gilt(
	jetzt time.Time, entleiher, abdruck string, bindung Leihbindung,
	recht Leihrecht,
) error {
	if entleiher != l.Entleiher {
		return fmt.Errorf("%w: diese Leihe gilt %s und nicht %s",
			ErrValidation, l.Entleiher, entleiher)
	}

	if jetzt.Before(l.Ausgestellt) {
		return fmt.Errorf("%w: diese Leihe gilt erst ab %s",
			ErrValidation, l.Ausgestellt.Format(time.RFC3339))
	}

	if !jetzt.Before(l.Bis) {
		return fmt.Errorf("%w: diese Leihe ist am %s abgelaufen",
			ErrValidation, l.Bis.Format(time.RFC3339))
	}

	if abdruck != l.Abdruck {
		return fmt.Errorf(
			"%w: diese Leihe gilt einem anderen Abzug", ErrValidation)
	}

	if fehler := l.pruefePassendeBindung(bindung); fehler != nil {
		return fehler
	}

	if !l.Darf(recht) {
		return fmt.Errorf("%w: diese Leihe erlaubt kein %q",
			ErrValidation, recht)
	}

	return nil
}

// pruefePassendeBindung hält die Bindung gegen den Fall.
//
// **Eine leere Kennung passt auf jede dieser Art** — aber nie auf eine andere
// Art: wer eine Partie geliehen hat, hat kein Turnier geliehen.
func (l Deckleihe) pruefePassendeBindung(bindung Leihbindung) error {
	if bindung.Art != l.Bindung.Art {
		return fmt.Errorf("%w: diese Leihe gilt einer Bindung der Art %q",
			ErrValidation, l.Bindung.Art)
	}

	if l.Bindung.Woran != "" && l.Bindung.Woran != bindung.Woran {
		return fmt.Errorf("%w: diese Leihe gilt %q und nicht %q",
			ErrValidation, l.Bindung.Woran, bindung.Woran)
	}

	return nil
}

// Unterschriftsform ist, was unterschrieben wird.
//
// # Warum eine gefügte Zeichenkette und kein JSON
//
// **Weil JSON zwei Fassungen derselben Angabe kennt** — Einrückung,
// Feldreihenfolge, die Schreibweise einer Zeit. Wer darüber unterschriebe,
// müsste eine kanonische Form verabreden, und **wie** sie aussieht, wäre eine
// zweite Verabredung, die niemand liest. Hier steht sie: jedes Feld, in fester
// Reihenfolge, durch ein Nullbyte getrennt.
//
// **Das Nullbyte, weil keines der Felder eines enthalten kann.** Sie sind
// Kennungen, Hexzahlen, Aufzählungswerte und Zeitpunkte; ein Trenner, der in
// einem Feld vorkommen kann, ließe zwei verschiedene Leihen dieselbe Form
// haben.
//
// **Die Unterschrift selbst geht nicht ein** — ein Feld, das in seinen eigenen
// Hash einginge, ließe sich nie ausrechnen. Derselbe Satz wie bei
// [Ereignis.Kette].
func (l Deckleihe) Unterschriftsform() []byte {
	rechte := make([]string, 0, len(l.Rechte))
	for _, recht := range l.Rechte {
		rechte = append(rechte, string(recht))
	}

	return []byte(strings.Join([]string{
		"deckleihe",
		l.Kennung,
		l.Verleiher,
		l.Entleiher,
		l.Abdruck,
		string(l.Bindung.Art),
		l.Bindung.Woran,
		strings.Join(rechte, ","),
		l.Ausgestellt.UTC().Format(time.RFC3339Nano),
		l.Bis.UTC().Format(time.RFC3339Nano),
		string(l.Modus),
	}, "\x00"))
}

// Leihstand ist eine [Deckleihe] samt dem, was dieser Rechner dazu weiß.
//
// # Warum das getrennt vom Dokument steht
//
// **Weil das Dokument über die Grenze geht und dies nicht.** Ob eine Leihe
// hier schon verbraucht oder widerrufen ist, ist eine Auskunft **dieses**
// Rechners; sie mitzuschicken hieße, dem Gegenüber einen Stand zu senden, den
// es nicht prüfen kann — und den es beim nächsten Neustart anders sähe.
type Leihstand struct {
	Leihe Deckleihe

	// Abzug ist der eingefrorene Abzug — der Grund steht im Kopf.
	Abzug Partiedeck

	// Verliehen sagt, ob **dieser** Rechner sie ausgestellt hat.
	Verliehen bool

	// Verbraucht ist die Partie oder das Turnier, an der sie eingelöst wurde.
	//
	// **Leer heißt: noch nicht.** Eine Leihe mit offener Bindung
	// ([Leihbindung.Woran] leer) wird damit beim ersten Start
	// festgenagelt; eine zweite Partie bekommt sie nicht.
	Verbraucht string

	// Widerrufen sagt, dass der Verleiher sie zurückgenommen hat.
	Widerrufen bool
}
