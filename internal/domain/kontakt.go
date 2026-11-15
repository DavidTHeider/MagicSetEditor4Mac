package domain

// **Ein lokaler Kontakt** — ID-04
// (#1703),
// Vermerk [0137](../../docs/entscheidungen/0137-kontakte-kommen-als-vcard.md).
//
// # Kein Gegenüber
//
// **Ein Kontakt ist ein Eintrag aus einem Adressbuch** — Name, Bild, wie man
// jemanden erreicht. Ein [Peerkontakt] ist ein anderer Rechner mit einem
// Schlüssel. Nichts an einem Kontakt ist geprüft: eine übernommene
// E-Mail-Adresse ist eine Angabe und kein nachgewiesener Kanal
// (ID-04.AK-04, dieselbe Regel wie 0066 für die Adresse eines Nutzers), und
// dieser Typ trägt kein Feld, das etwas anderes behaupten könnte.
//
// **Neu** und nicht portiert: MSE2 hat kein Adressbuch.

import (
	"bytes"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Kontaktwegart ist, wie man jemanden erreicht.
type Kontaktwegart string

// Die zwei Arten, die übernommen werden (ID-04.AK-02).
const (
	KontaktwegEmail   Kontaktwegart = "email"
	KontaktwegTelefon Kontaktwegart = "telefon"
)

// KontaktbildGrenze ist die größte Bilddatei, die ein Kontakt trägt —
// ein Adressbuchbild ist ein Vorschaubild, keine Fotografie.
const KontaktbildGrenze = 512 * 1024

// kontaktnamenGrenze ist die längste Form eines Namens in Zeichen.
const kontaktnamenGrenze = 200

// kontaktbildarten sind die Arten, die ein Fenster ohne Umweg zeigt.
var kontaktbildarten = []string{"image/jpeg", "image/png", "image/gif", "image/heic"}

// Kontaktweg ist eine Adresse oder Nummer.
type Kontaktweg struct {
	Art  Kontaktwegart
	Wert string
}

// Kontakt ist ein Mensch aus einem Adressbuch, lokal in diesem Profil.
type Kontakt struct {
	ID      int64
	Fassung int64

	Name string
	Wege []Kontaktweg

	// Bild und Bildart sind das Bild — beides leer ohne.
	Bild    []byte
	Bildart string

	// Ursprung ist die Kennung der Karte im Adressbuch (vCard `UID`) —
	// **woran ein zweiter Import denselben Menschen erkennt**. Leer, wenn
	// die Karte keine trug.
	Ursprung string

	// Uebernommen ist, wann er zuletzt übernommen wurde.
	Uebernommen time.Time
}

// Normalisieren glättet, was aus einer Datei kommt.
func (k *Kontakt) Normalisieren() {
	k.Name = strings.TrimSpace(k.Name)
	k.Ursprung = strings.TrimSpace(k.Ursprung)

	for stelle := range k.Wege {
		k.Wege[stelle].Wert = strings.TrimSpace(k.Wege[stelle].Wert)
	}
}

// Validate gibt einen Befund je Feld, das nicht stimmt.
func (k Kontakt) Validate() []*FieldError {
	var befunde []*FieldError

	if k.Name == "" || utf8.RuneCountInString(k.Name) > kontaktnamenGrenze {
		befunde = append(befunde, NewFieldError("name", SchluesselValidierung,
			"Ein Kontakt braucht einen Namen mit höchstens %d Zeichen.", kontaktnamenGrenze))
	}

	for _, weg := range k.Wege {
		if weg.Wert == "" || (weg.Art != KontaktwegEmail && weg.Art != KontaktwegTelefon) {
			befunde = append(befunde, NewFieldError("wege", SchluesselValidierung,
				"Ein Kontaktweg ist eine E-Mail-Adresse oder eine Telefonnummer und nicht leer."))

			break
		}
	}

	if len(k.Bild) > 0 && (!slices.Contains(kontaktbildarten, k.Bildart) || len(k.Bild) > KontaktbildGrenze) {
		befunde = append(befunde, NewFieldError("bild", SchluesselValidierung,
			"Ein Kontaktbild ist JPEG, PNG, GIF oder HEIC und höchstens %d KiB groß.", KontaktbildGrenze/1024))
	}

	return befunde
}

// Kontaktabgleich sagt, wie ein neuer Kontakt zu den vorhandenen steht.
type Kontaktabgleich string

// Die vier Antworten.
const (
	// KontaktNeu: es gibt ihn noch nicht.
	KontaktNeu Kontaktabgleich = "neu"

	// KontaktGleich: es gibt ihn mit denselben Angaben — nichts zu tun.
	KontaktGleich Kontaktabgleich = "gleich"

	// KontaktGeaendert: es gibt ihn, mit anderen Angaben.
	KontaktGeaendert Kontaktabgleich = "geaendert"

	// KontaktVermutlich: ohne gemeinsamen Ursprung, aber gleicher Name und
	// ein gemeinsamer Weg — **vermutlich** derselbe Mensch. Ob er es ist,
	// entscheidet ein Mensch.
	KontaktVermutlich Kontaktabgleich = "vermutlich"
)

// KontaktAbgleichen findet den vorhandenen Kontakt zu einem neuen.
//
// **Zuerst über den Ursprung**, denn der ist die Kennung der Karte; **dann
// über Name und Weg**, und dann nur als Vermutung — zwei Menschen können
// gleich heißen, und eine geteilte Familienadresse macht sie nicht zu einem.
func KontaktAbgleichen(vorhandene []Kontakt, neu Kontakt) (Kontaktabgleich, *Kontakt) {
	if neu.Ursprung != "" {
		for stelle := range vorhandene {
			if vorhandene[stelle].Ursprung != neu.Ursprung {
				continue
			}

			if gleicheAngaben(vorhandene[stelle], neu) {
				return KontaktGleich, &vorhandene[stelle]
			}

			return KontaktGeaendert, &vorhandene[stelle]
		}
	}

	// **Gleiche Angaben ohne Ursprung sind derselbe** (Nachlauf über ID-04,
	// Befund 1): zwei Karten, die nur einen Namen tragen, teilen keinen Weg,
	// und ohne diese Regel kam jede doppelt durch.
	for stelle := range vorhandene {
		ohneUrsprung := vorhandene[stelle].Ursprung == "" || neu.Ursprung == ""
		if ohneUrsprung && gleicheAngaben(vorhandene[stelle], neu) {
			return KontaktGleich, &vorhandene[stelle]
		}
	}

	for stelle := range vorhandene {
		if strings.EqualFold(vorhandene[stelle].Name, neu.Name) && gemeinsamerWeg(vorhandene[stelle], neu) {
			return KontaktVermutlich, &vorhandene[stelle]
		}
	}

	return KontaktNeu, nil
}

// gleicheAngaben vergleicht, was ein Import ändern würde.
func gleicheAngaben(a, b Kontakt) bool {
	return a.Name == b.Name && a.Bildart == b.Bildart && bytes.Equal(a.Bild, b.Bild) &&
		slices.Equal(a.Wege, b.Wege)
}

// gemeinsamerWeg sagt, ob zwei Kontakte eine Adresse oder Nummer teilen —
// **in Vergleichsform**: eine Adresse ohne Groß- und Kleinschreibung, eine
// Nummer nur aus Ziffern und einem führenden Plus.
func gemeinsamerWeg(a, b Kontakt) bool {
	for _, einer := range a.Wege {
		for _, anderer := range b.Wege {
			if einer.Art == anderer.Art && vergleichsform(einer) == vergleichsform(anderer) {
				return true
			}
		}
	}

	return false
}

func vergleichsform(weg Kontaktweg) string {
	if weg.Art == KontaktwegEmail {
		return strings.ToLower(weg.Wert)
	}

	return strings.Map(func(zeichen rune) rune {
		if unicode.IsDigit(zeichen) || zeichen == '+' {
			return zeichen
		}

		return -1
	}, weg.Wert)
}
