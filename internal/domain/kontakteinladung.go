package domain

// **Wer mich aufnehmen soll, bekommt ein Stück Text** — NET-01
// (#1705).
//
// # Was sie von der Einladung aus N3 unterscheidet
//
// **Jene sagt, wo gewählt wird; diese sagt, wer dort sitzt.** Die Einladung
// aus [NeueEinladung] trägt eine Adresse und sonst nichts — wer sie bekam,
// musste die Kennung getrennt austauschen und ohne sie einen Kontakt von Hand
// anlegen. Diese hier bindet, was NET-01.AK-02 verlangt: die **erwartete
// Gegenstelle** (die Kennung **ist** der Schlüssel), den **Zweck**, die
// **Fassung** und den **Ablauf** — und sie ist unterschrieben, von genau dem
// Schlüssel, den sie nennt.
//
// **Deshalb eine eigene Marke**: `mse4mac-kontakt` statt `mse4mac`. Ein Feld,
// das eine Adresse erwartet, nimmt die eine an und weist die andere ab, und
// umgekehrt — zwei Texte, die ähnlich aussehen und Verschiedenes tun, wären
// die Verwechslung, vor der eine Marke schützt.
//
// # Warum sie kein kurzer Code ist
//
// **Weil ein kurzer Code ein eigenes Verfahren verlangte** (AK-02: „ein
// geeignetes etabliertes Sicherheitsverfahren sowie Raten- und
// Wiederholungsschutz") — ein PAKE, und damit eine Verbindung, über die
// geraten werden kann. Diese Einladung trägt den ganzen Schlüssel; zu raten
// gibt es nichts. Der **Wiederholungsschutz** ist die Nonce: eine
// angenommene Einladung wird kein zweites Mal angenommen. Geprüft wird sie im
// Dienst, weil er den Bestand kennt.
//
// # Warum die Unterschrift nicht hier geprüft wird
//
// **Weil `domain` keinen Schlüssel kennt** (§2). Hier steht, **was**
// unterschrieben wird — die Bytes samt Vorsatz; wer prüft, ist der Dienst mit
// `peer.Pruefe`. Und geprüft wird über die Bytes, **die ankamen**, nicht über
// eine neu geschriebene Fassung derselben Felder: zwei JSON-Schreiber müssen
// sich über Leerzeichen nicht einig sein.

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Kontakteinladungsmarke steht am Anfang jeder Kontakteinladung.
const Kontakteinladungsmarke = "mse4mac-kontakt"

// Kontakteinladungsfassung ist die Fassung, die dieses Programm ausgibt und
// annimmt.
const Kontakteinladungsfassung = 1

// KontakteinladungGueltig ist, wie lange eine neue Einladung gilt.
//
// **Ein Tag**: lang genug, dass eine Nachricht über Nacht ankommt, und kurz
// genug, dass eine vergessene in einem alten Verlauf nichts mehr öffnet.
const KontakteinladungGueltig = 24 * time.Hour

// KontakteinladungHoechstens ist die längste Gültigkeit, die angenommen wird.
//
// **Eine Grenze auf der annehmenden Seite**, weil die ausgebende sie nicht
// setzt, wenn sie nicht dieses Programm ist.
const KontakteinladungHoechstens = 7 * 24 * time.Hour

// kontakteinladungsname ist die längste Namenslänge in Zeichen.
const kontakteinladungsname = 64

// kontakteinladungsnonce ist die Länge der Nonce in Bytes.
const kontakteinladungsnonce = 16

// Kontakteinladung ist, was ein Gegenüber über mich wissen muss.
type Kontakteinladung struct {
	// Kennung ist der Schlüssel, der unterschrieben hat — und der erwartet
	// wird, wenn gewählt wird.
	Kennung string `json:"kennung"`

	// Name ist, wie ich mich nenne — **kein Nachweis**.
	Name string `json:"name"`

	// Zweck ist, wofür ich einlade; freigegeben wird er erst nach dem
	// Vergleich.
	Zweck Zweck `json:"zweck"`

	// Adresse ist, wo ich erreichbar bin — leer, wenn ich es nicht weiß.
	Adresse string `json:"adresse,omitempty"`

	Ablauf time.Time `json:"ablauf"`

	// Nonce macht die Einladung einmalig.
	Nonce string `json:"nonce"`

	// Vorher ist meine Kennung vor der letzten Rotation.
	//
	// **Eine Behauptung, kein Nachweis** — der alte Schlüssel ist fort und
	// kann nichts mehr unterschreiben. Sie dient einzig dem Hinweis
	// „Schlüssel geändert"; Vertrauen geht über sie nicht (AK-03).
	Vorher string `json:"vorher,omitempty"`

	// Einloesbar sagt, dass der Einladende eine Einlösung an [Adresse]
	// beantwortet ([Einloesung], NET-01b). **In den unterschriebenen Daten**
	// und nur geschrieben, wenn es gilt: eine ältere Fassung kennt das Feld
	// nicht, stellt also nie eine einlösbare aus — und wer eine annimmt,
	// versucht nichts, was drüben niemand beantwortet.
	Einloesbar bool `json:"einloesbar,omitempty"`
}

// Unterschrieben sind die Bytes einer Einladung und ihre Unterschrift.
type Unterschrieben struct {
	Daten        []byte
	Unterschrift []byte
}

// kontakteinladungsvorsatz trennt diese Unterschrift von jeder anderen
// desselben Schlüssels.
func kontakteinladungsvorsatz() []byte {
	return []byte(Kontakteinladungsmarke + "/" +
		strconv.Itoa(Kontakteinladungsfassung) + "\n")
}

// Kodiere schreibt die Einladung und lässt sie unterschreiben.
func (e Kontakteinladung) Kodiere(
	unterschreibe func(daten []byte) []byte,
) (string, error) {
	if fehler := e.pruefeFelder(); fehler != nil {
		return "", fehler
	}

	inhalt, fehler := json.Marshal(e)
	if fehler != nil {
		return "", MitGrund(ErrValidation, fehler, "eine Einladung schreiben")
	}

	daten := append(kontakteinladungsvorsatz(), inhalt...)
	b64 := base64.RawURLEncoding.EncodeToString

	return fmt.Sprintf("%s:%d:%s.%s", Kontakteinladungsmarke,
		Kontakteinladungsfassung, b64(inhalt), b64(unterschreibe(daten))), nil
}

// ZerlegeKontakteinladung liest eine Einladung — **ungeprüft**.
//
// Was sie sagt, gilt erst, wenn der Dienst die Unterschrift über
// [Unterschrieben.Daten] geprüft hat und [Kontakteinladung.Pruefe] nichts
// einzuwenden hat.
func ZerlegeKontakteinladung(
	text string,
) (Kontakteinladung, Unterschrieben, error) {
	teile := strings.SplitN(ohneLeerraum(text), ":", 3)
	if len(teile) != 3 || teile[0] != Kontakteinladungsmarke {
		return Kontakteinladung{}, Unterschrieben{}, fmt.Errorf(
			"%w: das ist keine Kontakteinladung dieses Programms",
			ErrValidation)
	}

	fassung, fehler := strconv.Atoi(teile[1])
	if fehler != nil || fassung < 1 {
		return Kontakteinladung{}, Unterschrieben{}, fmt.Errorf(
			"%w: die Einladung nennt keine Fassung", ErrValidation)
	}

	if fassung > Kontakteinladungsfassung {
		return Kontakteinladung{}, Unterschrieben{}, fmt.Errorf("%w (%d gegen %d)",
			ErrEinladungZuNeu, fassung, Kontakteinladungsfassung)
	}

	return zerlegeRumpf(teile[2])
}

// zerlegeRumpf liest Daten und Unterschrift.
func zerlegeRumpf(rumpf string) (Kontakteinladung, Unterschrieben, error) {
	kaputt := fmt.Errorf(
		"%w: die Einladung ist unvollständig — wurde sie ganz kopiert?",
		ErrValidation)

	daten, unterschrift, gefunden := strings.Cut(rumpf, ".")
	if !gefunden {
		return Kontakteinladung{}, Unterschrieben{}, kaputt
	}

	inhalt, fehler := base64.RawURLEncoding.DecodeString(daten)
	if fehler != nil {
		return Kontakteinladung{}, Unterschrieben{}, kaputt
	}

	// **Eine leere Unterschrift ist ein abgeschnittener Code** — nach dem
	// Punkt kopiert und nicht weiter. Sie scheiterte später an der Prüfung,
	// mit „ungültig" statt mit dem Satz, der hilft (FuzzZerlegeKontakteinladung,
	// QA-06).
	signatur, fehler := base64.RawURLEncoding.DecodeString(unterschrift)
	if fehler != nil || len(signatur) == 0 {
		return Kontakteinladung{}, Unterschrieben{}, kaputt
	}

	var einladung Kontakteinladung
	if fehler := json.Unmarshal(inhalt, &einladung); fehler != nil {
		return Kontakteinladung{}, Unterschrieben{}, kaputt
	}

	return einladung, Unterschrieben{
		Daten:        append(kontakteinladungsvorsatz(), inhalt...),
		Unterschrift: signatur,
	}, nil
}

// Pruefe hält die Einladung gegen das, was eine sein kann — und gegen die Uhr.
func (e Kontakteinladung) Pruefe(jetzt time.Time) error {
	if fehler := e.pruefeFelder(); fehler != nil {
		return fehler
	}

	if !jetzt.Before(e.Ablauf) {
		return fmt.Errorf(
			"%w: die Einladung ist abgelaufen — bitte um eine neue",
			ErrValidation)
	}

	if e.Ablauf.Sub(jetzt) > KontakteinladungHoechstens {
		return fmt.Errorf(
			"%w: die Einladung gilt länger als sieben Tage", ErrValidation)
	}

	return nil
}

// pruefeFelder hält alles außer der Uhr.
func (e Kontakteinladung) pruefeFelder() error {
	if e.Kennung == "" {
		return fmt.Errorf("%w: die Einladung nennt niemanden", ErrValidation)
	}

	if e.Vorher == e.Kennung {
		return fmt.Errorf(
			"%w: eine Kennung ist nicht ihre eigene Vorgängerin", ErrValidation)
	}

	if fehler := pruefeEinladungsname(e.Name); fehler != nil {
		return fehler
	}

	if fehler := PruefeZweck(e.Zweck); fehler != nil {
		return fehler
	}

	if fehler := PruefeAdresse(e.Adresse); fehler != nil {
		return fehler
	}

	if e.Einloesbar && e.Adresse == "" {
		return fmt.Errorf(
			"%w: eine einlösbare Einladung nennt, wo sie eingelöst wird",
			ErrValidation)
	}

	return pruefeNonce(e.Nonce)
}

// pruefeNonce hält die Nonce einer Einladung — auch dort, wo sie eingelöst
// wird.
func pruefeNonce(text string) error {
	nonce, fehler := base64.RawURLEncoding.DecodeString(text)
	if fehler != nil || len(nonce) != kontakteinladungsnonce {
		return fmt.Errorf(
			"%w: die Einladung trägt keine gültige Nonce", ErrValidation)
	}

	return nil
}

// pruefeEinladungsname weist ab, was ein Mensch nicht lesen kann.
//
// **Formatzeichen eingeschlossen**: ein U+202E kehrt die Schreibrichtung um,
// und ein Name, der anders aussieht, als er ist, wäre die Täuschung, vor der
// der Vergleich schützen soll.
func pruefeEinladungsname(name string) error {
	if !utf8.ValidString(name) || strings.TrimSpace(name) == "" ||
		utf8.RuneCountInString(name) > kontakteinladungsname {
		return fmt.Errorf(
			"%w: der Name einer Einladung hat ein bis %d Zeichen",
			ErrValidation, kontakteinladungsname)
	}

	if strings.IndexFunc(name, unsichtbar) >= 0 {
		return fmt.Errorf(
			"%w: der Name einer Einladung enthält ein unsichtbares Zeichen",
			ErrValidation)
	}

	return nil
}

// unsichtbar erkennt ein Steuer- oder Formatzeichen.
func unsichtbar(zeichen rune) bool {
	return unicode.IsControl(zeichen) || unicode.Is(unicode.Cf, zeichen)
}

// LesbarerName sagt, ob ein selbstgewählter Name einer ist, den ein Mensch
// lesen kann — **dieselbe Regel wie beim Namen einer Einladung**: ein bis 64
// Zeichen, gültiges UTF-8, kein Steuer- oder Formatzeichen. Für Namen, die
// ein Fremder ohne Prüfung schickt, etwa in einer Antwort auf die Suche im
// Netz (QA-06, #1737).
func LesbarerName(name string) bool {
	return pruefeEinladungsname(name) == nil
}

// ohneLeerraum nimmt alles heraus, was ein Messenger hineinbricht.
func ohneLeerraum(text string) string {
	return strings.Map(func(zeichen rune) rune {
		if unicode.IsSpace(zeichen) {
			return -1
		}

		return zeichen
	}, text)
}

// Kontaktaufnahme ist, was eine angenommene Einladung schreibt — in einem
// Zug.
//
// **Stand und Rechte stehen im Kontakt**, weil der Dienst sie entschieden
// hat: unbestätigt ohne Vergleich, bestätigt mit ihm, und Rechte nur dann.
type Kontaktaufnahme struct {
	Kontakt Peerkontakt
	Nonce   string
}
