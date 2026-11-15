// **Die Kennungen eines Kontos** — Werkstück 6 von Phase 4.12
// (#1351),
// nach der Kommentierung vom 2026-09-05.
//
// # Vier Arten, ein Nutzer
//
// | Art | Kennung | Nachweis |
// | --- | --- | --- |
// | **ohne Konto** | keine | keiner — der Nutzer aus Wanderung 0034 |
// | **anonym** | eine Zufallskennung als Name | keiner |
// | **E-Mail** | die Adresse | Kennwort |
// | **Telefon** | die Nummer | Kennwort |
//
// **Die vier Anbieter stehen daneben** ([Anbieterkonto]): dort steht, wer
// jemand **anderswo** ist; hier, wer er **hier** ist.

package domain

import (
	"fmt"
	"strings"
	"unicode"
)

// Hoechstadresse ist, wie lang eine Adresse sein darf.
//
// **254 Zeichen**, die Grenze eines Umschlagpfads aus RFC 5321 §4.5.3.1.3. Sie
// steht hier nicht als Zusage über die Grammatik, sondern als Riegel gegen
// eine Eingabe, die kein Postfach mehr sein kann.
const Hoechstadresse = 254

// PruefeEmail sagt, ob diese Zeichenkette eine Adresse sein kann.
//
// # Warum so wenig geprüft wird
//
// **Weil mehr falsch wäre.** Die gültigen Adressen sind in RFC 5322 so weit
// gefasst, dass jede kürzere Regel echte Adressen abweist — und wer eine
// abgewiesen bekommt, die sein Postfach hat, hält das Programm für kaputt.
// Geprüft wird der **Tippfehler** und nicht die Grammatik: ein `@`, etwas auf
// beiden Seiten, kein Leerraum.
//
// **Was das nicht sagt**: dass es die Adresse gibt. Das sagt nur eine
// Zustellung, und die liefert dieses Programm nicht mit — siehe die
// Kommentierung am Werkstück.
func PruefeEmail(adresse string) error {
	if adresse == "" {
		return fmt.Errorf("%w: eine Adresse fehlt", ErrValidation)
	}

	if adresse != strings.TrimSpace(adresse) {
		return fmt.Errorf(
			"%w: eine Adresse trägt keine Leerzeichen am Rand", ErrValidation)
	}

	if len(adresse) > Hoechstadresse {
		return fmt.Errorf("%w: eine Adresse hat höchstens %d Zeichen",
			ErrValidation, Hoechstadresse)
	}

	if strings.ContainsFunc(adresse, unicode.IsSpace) {
		return fmt.Errorf("%w: eine Adresse trägt keine Leerzeichen", ErrValidation)
	}

	vorne, hinten, hatAt := strings.Cut(adresse, "@")
	if !hatAt || vorne == "" || hinten == "" || strings.Contains(hinten, "@") {
		return fmt.Errorf(
			"%w: eine Adresse hat genau ein @ und etwas auf beiden Seiten",
			ErrValidation)
	}

	return nil
}

// Die Grenzen einer Rufnummer nach E.164.
const (
	// Mindestnummer ist, wie kurz eine Nummer sein darf — Ziffern ohne Plus.
	Mindestnummer = 6

	// Hoechstnummer ist die Grenze aus E.164: fünfzehn Ziffern.
	Hoechstnummer = 15
)

// PruefeTelefon sagt, ob diese Zeichenkette eine Rufnummer sein kann.
//
// **In E.164-Gestalt, aber ohne Länderprüfung**: welche Vorwahlen es gibt,
// ändert sich, und eine Liste im Programm wäre am Tag ihrer Aufnahme veraltet.
// Geprüft wird, dass da Ziffern stehen und höchstens ein Plus davor.
func PruefeTelefon(nummer string) error {
	if nummer == "" {
		return fmt.Errorf("%w: eine Rufnummer fehlt", ErrValidation)
	}

	if nummer != strings.TrimSpace(nummer) {
		return fmt.Errorf(
			"%w: eine Rufnummer trägt keine Leerzeichen am Rand", ErrValidation)
	}

	ziffern := strings.TrimPrefix(nummer, "+")

	for _, zeichen := range ziffern {
		if !unicode.IsDigit(zeichen) {
			return fmt.Errorf(
				"%w: eine Rufnummer trägt nur Ziffern und darf ein + davor haben",
				ErrValidation)
		}
	}

	if len(ziffern) < Mindestnummer || len(ziffern) > Hoechstnummer {
		return fmt.Errorf("%w: eine Rufnummer hat %d bis %d Ziffern",
			ErrValidation, Mindestnummer, Hoechstnummer)
	}

	return nil
}
