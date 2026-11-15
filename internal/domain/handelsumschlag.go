package domain

// **Der Kopf einer Handelsnachricht** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Warum die Arten eine Aufzählung sind
//
// **Weil eine Fläche, die jede Zeichenkette als Art annimmt, ein frei
// wählbarer Aufrufkanal ist.** Ein Gegenüber soll handeln können und sonst
// nichts — keine lokale Methode wählen, keine Datei benennen, kein Skript
// anstoßen. Zwölf Arten, und was nicht darunter fällt, endet vor der Naht.
//
// # Warum der Zeitstempel nicht unterschrieben wird
//
// **Weil er ein Metadatum ist und kein Replay-Schutz.** Eine Uhr, die
// vorgeht, ist kein Angriff; eine, die nachgeht, darf keine Nachricht
// ungültig machen. Gegen die zweite Zustellung steht der
// **Idempotenzschlüssel**, und der hängt an dem, was die Nachricht bewirkt.
//
// # Und warum die Kennung nicht im Schlüssel steht
//
// **Weil zwei verschiedene Nachrichten derselben Operation gegen dieselbe
// Revision ein Konflikt sind und nicht zwei Vorgänge.** Wer die Kennung
// hineinnähme, ließe beide durch — und der zweite überschriebe den ersten,
// ohne dass irgendwo stünde, dass es zwei gab.

import (
	"strconv"
	"strings"
	"time"
)

// Handelsprotokollfassung ist die Fassung, die dieses Programm spricht.
//
// **Eine Zahl und kein Bereich.** Wer eine fremde Fassung durchließe, deutete
// Felder, deren Bedeutung er nicht kennt; und wer sie stillschweigend als die
// eigene läse, täte es ohne es zu merken.
const Handelsprotokollfassung = 1

// Handelsnachrichtenart ist eine der zwölf erlaubten Arten.
type Handelsnachrichtenart string

// Die zwölf Arten — der ganze Wortschatz dieses Protokolls.
const (
	// NachrichtHallo nennt Fassung und Fähigkeiten.
	NachrichtHallo Handelsnachrichtenart = "hallo"

	// NachrichtSichtanfrage bittet um die freigegebene Sicht (T3).
	NachrichtSichtanfrage Handelsnachrichtenart = "sichtanfrage"

	// NachrichtSicht ist die Antwort darauf, auf derselben Leitung — P11
	// (#2379).
	// **Erweiternd**: eine Fassung, die sie nicht kennt, bekommt sie nie,
	// denn sie antwortet nur auf eine Frage, die jene nicht stellt.
	NachrichtSicht Handelsnachrichtenart = "sicht"

	NachrichtVorschlag      Handelsnachrichtenart = "vorschlag"
	NachrichtGegenvorschlag Handelsnachrichtenart = "gegenvorschlag"
	NachrichtAnnahme        Handelsnachrichtenart = "annahme"
	NachrichtAblehnung      Handelsnachrichtenart = "ablehnung"

	// NachrichtVorbereitung bindet beide Seiten, bevor etwas übergeht.
	NachrichtVorbereitung Handelsnachrichtenart = "vorbereitung"

	// NachrichtUebergabe bestätigt, dass ein Teil übergegangen ist.
	NachrichtUebergabe Handelsnachrichtenart = "uebergabe"

	// NachrichtCommitstand sagt, was die andere Seite verbucht hat.
	NachrichtCommitstand Handelsnachrichtenart = "commitstand"

	NachrichtAbbruch Handelsnachrichtenart = "abbruch"
	NachrichtBeleg   Handelsnachrichtenart = "beleg"
)

// alleNachrichtenarten ist die Aufzählung in fester Reihenfolge.
var alleNachrichtenarten = []Handelsnachrichtenart{
	NachrichtHallo, NachrichtSichtanfrage, NachrichtSicht,
	NachrichtVorschlag, NachrichtGegenvorschlag,
	NachrichtAnnahme, NachrichtAblehnung,
	NachrichtVorbereitung, NachrichtUebergabe, NachrichtCommitstand,
	NachrichtAbbruch, NachrichtBeleg,
}

// Handelsnachrichtenarten gibt die zwölf — als Kopie.
func Handelsnachrichtenarten() []Handelsnachrichtenart {
	kopie := make([]Handelsnachrichtenart, len(alleNachrichtenarten))
	copy(kopie, alleNachrichtenarten)

	return kopie
}

// IstHandelsnachrichtenart sagt, ob es die Art gibt.
func IstHandelsnachrichtenart(art Handelsnachrichtenart) bool {
	for _, kandidat := range alleNachrichtenarten {
		if kandidat == art {
			return true
		}
	}

	return false
}

// Handelsumschlag ist der Kopf, den jede Handelsnachricht trägt.
//
// **Der Kopf und nicht die Nutzlast.** Was eine Nachricht **sagt**, hängt an
// ihrer Art; was sie **ist**, steht hier — und nur das lässt sich prüfen,
// bevor irgendjemand die Nutzlast deutet.
type Handelsumschlag struct {
	// Fassung ist die Protokollfassung des Absenders.
	Fassung int

	Art Handelsnachrichtenart

	// NachrichtID ist eindeutig je Absender — für die Zuordnung einer
	// Antwort, nicht für die Idempotenz.
	NachrichtID string

	// Vorgang ist die Kennung des Handels (T1).
	Vorgang string

	// Basisrevision ist die Revision, gegen die diese Nachricht gilt.
	Basisrevision int

	// Abdruck ist der Angebotsabdruck, den sie meint — die Kennung der
	// Revision ([Handelsangebot.Abdruck]).
	Abdruck string

	// Folge ist die Sequenz je Absender und Vorgang — **ab eins**.
	Folge int64

	// Absender ist die Peerkennung, und sie wird gegen den
	// **authentifizierten** Sender gehalten, nicht geglaubt.
	Absender string

	// Unterschrift ist die Signatur über [Handelsumschlag.Unterschriftsform].
	Unterschrift string

	// Gesendet ist die Uhr des Absenders — **ein Metadatum**. Sie steht nicht
	// in der Unterschrift und schützt vor nichts.
	Gesendet time.Time
}

// Validate gibt einen Befund je Feld, das nicht stimmt.
func (u Handelsumschlag) Validate() []*FieldError {
	var befunde []*FieldError

	if u.Fassung != Handelsprotokollfassung {
		befunde = append(befunde, NewFieldError("fassung", SchluesselValidierung,
			"Diese Fassung des Handelsprotokolls ist %d, nicht %d.",
			Handelsprotokollfassung, u.Fassung))
	}

	if !IstHandelsnachrichtenart(u.Art) {
		befunde = append(befunde, NewFieldError("art", SchluesselValidierung,
			"Die Nachrichtenart %q gibt es nicht.", u.Art))
	}

	befunde = append(befunde, u.pflichtfelder()...)

	if u.Basisrevision < 0 {
		befunde = append(befunde, NewFieldError("basisrevision",
			SchluesselValidierung, "Eine Revision ist nicht negativ."))
	}

	if u.Folge < 1 {
		befunde = append(befunde, NewFieldError("folge", SchluesselValidierung,
			"Die Folge zählt ab eins; %d gibt es nicht.", u.Folge))
	}

	return befunde
}

// pflichtfelder prüft die vier, die dastehen müssen.
func (u Handelsumschlag) pflichtfelder() []*FieldError {
	var befunde []*FieldError

	for _, pflicht := range []struct {
		feld, wert, satz string
	}{
		{"nachrichtID", u.NachrichtID, "Eine Nachricht hat eine Kennung."},
		{"vorgang", u.Vorgang, "Eine Nachricht gehört zu einem Vorgang."},
		{"absender", u.Absender, "Eine Nachricht hat einen Absender."},
		{"unterschrift", u.Unterschrift, "Eine Nachricht ist unterschrieben."},
	} {
		if strings.TrimSpace(pflicht.wert) == "" {
			befunde = append(befunde, NewFieldError(pflicht.feld,
				SchluesselPflichtfeld, "%s", pflicht.satz))
		}
	}

	return befunde
}

// Idempotenzschluessel bindet Vorgang, Operation und Revision.
//
// **Diese drei und nicht die Nachrichtenkennung.** Zweimal dieselbe Operation
// gegen dieselbe Revision ist **eine** Wirkung; wer sie an der Kennung
// festmachte, ließe eine zweite Nachricht mit anderem Inhalt als zweiten
// Vorgang durch — und die Wirkung geschähe zweimal.
//
// **Das Profil steht nicht darin, und das ist kein Versehen.** Jede
// Wurzelabfrage klemmt ohnehin auf ein Profil
// ([0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md));
// es hier ein zweites Mal zu binden hieße, dieselbe Zusage an zwei Stellen zu
// führen — und die zweite läuft irgendwann von der ersten weg.
func (u Handelsumschlag) Idempotenzschluessel() string {
	return strings.Join([]string{
		u.Vorgang,
		string(u.Art),
		strconv.Itoa(u.Basisrevision),
	}, "\x00")
}

// Unterschriftsform gibt die Bytes, über die unterschrieben wird.
//
// **Feste Reihenfolge, Nullbyte dazwischen, und eine Marke vorweg** — genau
// wie bei [Handelsangebot.Unterschriftsform]. Ohne den Trenner ergäben „ab"
// und „c" dieselbe Form wie „a" und „bc", und eine Unterschrift ließe sich
// auf eine andere Nachricht übertragen; ohne die Marke ließe sich diese Form
// als die eines Angebots ausgeben.
func (u Handelsumschlag) Unterschriftsform() []byte {
	return []byte(strings.Join([]string{
		"handelsumschlag",
		strconv.Itoa(u.Fassung),
		string(u.Art),
		u.NachrichtID,
		u.Vorgang,
		strconv.Itoa(u.Basisrevision),
		u.Abdruck,
		strconv.FormatInt(u.Folge, 10),
		u.Absender,
	}, "\x00"))
}
