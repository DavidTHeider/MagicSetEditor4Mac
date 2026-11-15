package domain

// **Wo eine KI-Laufzeit lauscht** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Warum die Adresse den Ort entscheidet
//
// **Weil eine Adresse ein Nein beweisen kann und ein Ja nicht.** Das Issue
// sagt beides:
//
//   - „**Loopback allein beweist keine lokale Inferenz**" — auf `127.0.0.1`
//     kann ein Stellvertreter sitzen, der weiterreicht. Deshalb heißt
//     Schleifenadresse hier **nicht** „lokal", sondern nur: dieses Programm
//     hat keinen Gegenbeweis.
//   - „Remote-Ollama nur ausdrückliche Konfiguration und **eindeutige
//     Einstufung**" — eine Adresse, die **nicht** auf der Schleife liegt, ist
//     ein Gegenbeweis. Was dort rechnet, rechnet woanders.
//
// Die Einstufung ist damit einseitig, und genau so steht sie hier.

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// PruefeKIAdresse sagt, ob dieser Text eine brauchbare Adresse ist.
//
// **Leer ist erlaubt** und heißt: die Vorgabe der Laufzeit.
func PruefeKIAdresse(wert string) error {
	if strings.TrimSpace(wert) == "" {
		return nil
	}

	zerlegt, fehler := url.Parse(wert)
	if fehler != nil {
		return MitGrund(ErrValidation, fehler, "%q ist keine Adresse", wert)
	}

	if zerlegt.Scheme != "http" && zerlegt.Scheme != "https" {
		return fmt.Errorf("%w: %q nennt kein http oder https",
			ErrValidation, wert)
	}

	if zerlegt.Host == "" {
		return fmt.Errorf("%w: %q nennt keinen Rechner", ErrValidation, wert)
	}

	return nil
}

// KIOrtFuerAdresse stuft ein, wo eine Laufzeit unter dieser Adresse rechnet.
//
// **Einseitig, und der Kopf sagt warum**: nicht auf der Schleife heißt
// [KIOrtCloud], auf der Schleife heißt [KIOrtLokal] — und das zweite ist die
// Abwesenheit eines Gegenbeweises und kein Beleg.
func KIOrtFuerAdresse(adresse string) KIOrt {
	if strings.TrimSpace(adresse) == "" {
		return KIOrtLokal
	}

	zerlegt, fehler := url.Parse(adresse)
	if fehler != nil {
		return KIOrtCloud
	}

	if aufDerSchleife(zerlegt.Hostname()) {
		return KIOrtLokal
	}

	return KIOrtCloud
}

// aufDerSchleife sagt, ob dieser Rechnername diese Maschine meint.
//
// # Warum von Hand und nicht über `net`
//
// **Weil `internal/domain` das Netz nicht kennen darf** — die Zusage „von
// sich aus geschieht nichts" hängt an dieser Grenze, und ein Wächter
// (`internal/schichten/netz_test.go`) hält sie. `net.ParseIP` lüde nichts,
// aber die Grenze ist eine **Aufzählung**, keine Abwägung je Funktion.
//
// # Was diese Prüfung deckt
//
// **Die drei Schreibweisen, die vorkommen**: `localhost`, alles in
// `127.0.0.0/8` und `::1`. Die Klammern einer IPv6-Adresse hat `url.Hostname`
// schon entfernt.
//
// # Und was sie **nicht** deckt — in der sicheren Richtung
//
// **Exotische Schreibweisen derselben Adresse** — `2130706433`, `0x7f000001`,
// `127.1`. Sie gelten hier als **nicht** lokal, und das ist der Ausgang, den
// man haben will: unter »nur lokal« wird dort nicht gefragt. Ein Fehler in
// diese Richtung kostet eine Anfrage, die jemand anders schreiben muss; der
// umgekehrte kostete die Zusage.
func aufDerSchleife(rechner string) bool {
	if strings.EqualFold(rechner, "localhost") || rechner == "::1" {
		return true
	}

	teile := strings.Split(rechner, ".")
	if len(teile) != 4 || teile[0] != "127" {
		return false
	}

	for _, teil := range teile {
		zahl, fehler := strconv.Atoi(teil)
		if fehler != nil || zahl < 0 || zahl > 255 {
			return false
		}
	}

	return true
}
