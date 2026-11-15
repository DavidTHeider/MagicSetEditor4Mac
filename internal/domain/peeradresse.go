package domain

// **Die Adresse eines Gegenübers** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Wofür es sie gibt
//
// **Für Netze, über die mDNS nicht läuft** — nicht über ein Tailnet, nicht
// über ein VPN, nicht über eine Portweiterleitung, und in manchen WLANs gar
// nicht. Dort ist das Gegenüber **erreichbar** und nur nicht **auffindbar**;
// wer die Adresse kennt, wählt sie direkt.
//
// # Warum sie hier geprüft wird und nicht an der Leitung
//
// **Weil die Form eine Regel ist und die Erreichbarkeit eine Messung.** „Sieht
// das aus wie eine Adresse" beantwortet diese Schicht, ohne ein einziges Paket
// zu schicken; „antwortet dort jemand" beantwortet nur ein Versuch. Wer beides
// hier beantworten wollte, hätte eine Domäne, die ins Netz greift.
//
// # Warum ohne `net.SplitHostPort`
//
// **Weil `internal/domain` das Netz nicht erreichen darf** (§2, und der
// Wächter dazu steht in `internal/schichten/netz_test.go`). `net` trüge den
// Weg dorthin mit herein — für eine Zerlegung, die vier Zeilen lang ist.

import (
	"fmt"
	"strconv"
	"strings"
)

// Hoechsteadresse ist, wie lang eine Adresse werden darf.
//
// **253 Zeichen ist die Grenze eines Namens im DNS**, dazu ein Doppelpunkt und
// fünf Ziffern für den Port. Eine längere ist keine Adresse, sondern eine
// Eingabe, die jemand verwechselt hat.
const Hoechsteadresse = 253 + 1 + 5

// PruefeAdresse hält eine Adresse gegen das, was eine sein kann.
//
// **Leer ist erlaubt und heißt: keine.** Der Weg „Adresse von Hand" ist der
// Zusatz und nicht die Regel; ein Kontakt ohne Adresse ist der Normalfall.
//
// **Ein Wirt ist Pflicht.** `:41234` sagt nicht, **wo** — dorthin lässt sich
// nicht wählen. Zum **Horchen** ist genau das die übliche Form; dafür gibt es
// [PruefeHorchadresse].
func PruefeAdresse(adresse string) error {
	if adresse == "" {
		return nil
	}

	// **Eine Einladung ist auch eine Antwort auf „wo?"** — N3
	// (#1456).
	//
	// Im Adressfeld eines Kontakts steht entweder eine Adresse im eigenen
	// Netz oder eine Einladung über ein fremdes; beides beantwortet dieselbe
	// Frage, und zwei Felder daneben wären zwei Antworten auf eine.
	//
	// **Gefunden hat das ein Lauf und nicht das Lesen**: der Wähler bog
	// längst auf den Fernweg ab, und der Schreibweg ließ eine Einladung gar
	// nicht erst herein. Ein Paket, das trägt, und ein Weg, der nicht
	// durchgeht — die Naht aus §10.
	if IstEinladung(adresse) {
		_, fehler := ZerlegeEinladung(Einladung(adresse))

		return fehler
	}

	return pruefeAdresse(adresse, false)
}

// PruefeHorchadresse hält eine Adresse gegen das, was sich horchen lässt.
//
// # Warum es zwei Prüfungen sind und nicht eine
//
// **Weil `:41234` beides zugleich ist**: die übliche Form, auf **jeder**
// Schnittstelle dieses Rechners zu horchen — und keine Adresse, zu der sich
// wählen ließe. Eine Prüfung für beide Fragen beantwortet eine davon falsch.
//
// **Und sie hat es getan.** Bis zum 2026-09-14 gab es nur [PruefeAdresse], und
// sie wies `:41234` ab, während Katalog, Hilfetext und `FAEHIGKEITEN.md` genau
// diese Form empfahlen. Gefunden hat es der erste Lauf, der wirklich eine
// Horchadresse angab.
//
// **Leer ist hier nicht erlaubt.** Beim Kontakt heißt leer „ich kenne keine",
// und das ist eine Aussage; hier hieße es „horche nirgends", und dafür gibt es
// das Beenden.
func PruefeHorchadresse(adresse string) error {
	if adresse == "" {
		return fmt.Errorf(
			"%w: ohne Adresse lässt sich nicht horchen — etwa „:41234“ für "+
				"jede Schnittstelle dieses Rechners", ErrValidation)
	}

	return pruefeAdresse(adresse, true)
}

// pruefeAdresse ist beiden gemeinsam — bis auf den Wirt.
func pruefeAdresse(adresse string, ohneWirt bool) error {
	if len(adresse) > Hoechsteadresse {
		return fmt.Errorf(
			"%w: eine Adresse hat höchstens %d Zeichen und nicht %d",
			ErrValidation, Hoechsteadresse, len(adresse))
	}

	wirt, port, gilt := zerlege(adresse)
	if !gilt {
		return fmt.Errorf(
			"%w: %q ist keine Adresse — sie braucht einen Wirt und einen "+
				"Port, etwa „192.168.1.7:41234“ oder „[2001:db8::1]:41234“",
			ErrValidation, adresse)
	}

	// **Der leere Wirt ist nur beim Horchen erlaubt** — und nur er; ein
	// Wirt, der aus Leerzeichen besteht, bleibt auch dort einer zu viel.
	if !ohneWirt || wirt != "" {
		if fehler := pruefeWirt(wirt, adresse); fehler != nil {
			return fehler
		}
	}

	return pruefePort(port, adresse, ohneWirt)
}

// ZerlegeAdresse gibt Wirt und Port einer geprüften Adresse.
//
// # Warum das hier steht und nicht bei `net.SplitHostPort`
//
// **Weil `internal/app` das Netz nicht erreichen darf** (§2): `net` bringt
// den Weg dorthin mit, und der Wächter in `internal/schichten/netz_test.go`
// zählt nach. Die Zerlegung ist dieselbe wie in [PruefeAdresse] — sie steht
// schon da, und zwei Zerleger für dieselbe Form liefen auseinander.
func ZerlegeAdresse(adresse string) (wirt string, port int, fehler error) {
	if fehler := PruefeHorchadresse(adresse); fehler != nil {
		return "", 0, fehler
	}

	wirt, text, _ := zerlege(adresse)

	// **Der Fehler wird nicht geprüft, und das ist belegt**: `PruefeHorchadresse`
	// darüber hat denselben Weg genommen und dabei `strconv.Atoi` bestanden.
	port, _ = strconv.Atoi(text) //nolint:errcheck // siehe darüber

	return wirt, port, nil
}

// zerlege trennt Wirt und Port am **letzten** Doppelpunkt.
//
// **Am letzten und nicht am ersten**: eine IPv6-Adresse hat viele, und sie
// steht deshalb in eckigen Klammern — genau wie in einer URL.
func zerlege(adresse string) (wirt, port string, gilt bool) {
	trenner := strings.LastIndex(adresse, ":")
	if trenner < 0 {
		return "", "", false
	}

	wirt, port = adresse[:trenner], adresse[trenner+1:]

	// **Eine geklammerte Adresse wird ausgepackt** — und eine halb
	// geklammerte ist keine.
	if strings.HasPrefix(wirt, "[") {
		if !strings.HasSuffix(wirt, "]") {
			return "", "", false
		}

		return wirt[1 : len(wirt)-1], port, true
	}

	// **Und eine ungeklammerte darf keinen zweiten Doppelpunkt haben.** Wer
	// „2001:db8::1:41234“ durchließe, hielte den letzten Block für einen
	// Port — und wählte etwas anderes, als dastand.
	if strings.Contains(wirt, ":") {
		return "", "", false
	}

	return wirt, port, true
}

// pruefeWirt hält den Wirt gegen das Wenige, was sich hier sagen lässt.
//
// **Mehr als „nicht leer und ohne Leerraum“ steht hier nicht**, und das ist
// Absicht: ein Name im lokalen Netz (`mein-mac.local`), eine IPv4, eine IPv6
// und ein Name aus dem DNS sehen verschieden aus, und eine Regel, die alle
// vier trifft, träfe auch alles andere. Was wirklich gilt, sagt die Auflösung.
func pruefeWirt(wirt, ganze string) error {
	if wirt == "" {
		return fmt.Errorf(
			"%w: %q hat keinen Wirt vor dem Doppelpunkt",
			ErrValidation, ganze)
	}

	// **Und keine Klammer** — eine geklammerte Adresse ist schon ausgepackt;
	// was jetzt noch eine trägt, war `[[]:0` (FuzzZerlegeAdresse, QA-06).
	if strings.ContainsAny(wirt, " \t\n\r/\\?#@[]") {
		return fmt.Errorf(
			"%w: %q ist kein Wirt — er trägt ein Zeichen, das dort nicht "+
				"hingehört", ErrValidation, wirt)
	}

	return nil
}

// pruefePort hält den Port gegen seinen Bereich.
//
// # Warum die Null beim Horchen gilt und beim Wählen nicht
//
// **Weil sie dort eine Bitte ist und hier eine Adresse wäre.** `:0` heißt beim
// Öffnen „such mir einen freien aus" — welchen, sagt das Betriebssystem und
// danach [Horcher.Adresse]. Zu `:0` zu wählen hieße, an die Tür mit der
// Nummer null zu klopfen.
//
// **Der zweite Fund derselben Sorte** (2026-09-14): eine Prüfung für zwei
// Fragen beantwortet eine davon falsch — zuerst beim Wirt, dann beim Port.
func pruefePort(port, ganze string, freierZaehlt bool) error {
	nummer, fehler := strconv.Atoi(port)
	if fehler != nil {
		return fmt.Errorf(
			"%w: %q hat keinen Port hinter dem Doppelpunkt, sondern %q",
			ErrValidation, ganze, port)
	}

	untergrenze := 1
	if freierZaehlt {
		untergrenze = 0
	}

	if nummer < untergrenze || nummer > 65535 {
		return fmt.Errorf(
			"%w: %d ist kein Port — er liegt zwischen %d und 65535",
			ErrValidation, nummer, untergrenze)
	}

	return nil
}
