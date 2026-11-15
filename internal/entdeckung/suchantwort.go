// **Was eine Antwort bedeutet** — herausgelöst am 2026-09-21, als `suche.go`
// mit der Wahl über alle Schnittstellen (NET-02.NAT-01,
// #1706)
// über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Nebenan steht, **wie gefragt wird** — über welche Schnittstellen, mit
// welcher Frist, und was bei einem Fehler geschieht. Hier steht, **was eine
// Antwort bedeutet**: was sie behaupten muss, damit sie zählt, und woran ein
// gescheiterter Ruf lag.

package entdeckung

import (
	"errors"
	"io"
	"log"
	"strings"
	"syscall"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/hashicorp/mdns"
)

// sammle liest die Antworten und prüft, was sie behaupten.
func sammle(antworten <-chan *mdns.ServiceEntry) Ergebnis {
	var ergebnis Ergebnis

	gesehen := map[peer.Kennung]bool{}

	for antwort := range antworten {
		fund, gilt := deute(antwort)

		// **Zweimal dieselbe Kennung ist einmal** — mDNS antwortet über
		// IPv4 und IPv6, und beide Wege führen zu demselben Rechner.
		if !gilt || gesehen[fund.Kennung] {
			if !gilt {
				ergebnis.Verworfen++
			}

			continue
		}

		gesehen[fund.Kennung] = true
		ergebnis.Funde = append(ergebnis.Funde, fund)
	}

	return ergebnis
}

// deute macht aus einer Antwort einen Fund — oder nichts.
func deute(antwort *mdns.ServiceEntry) (Fund, bool) {
	fund := Fund{Port: antwort.Port}

	for _, feld := range antwort.InfoFields {
		switch {
		case strings.HasPrefix(feld, kennungsmarke):
			fund.Kennung = peer.Kennung(strings.TrimPrefix(feld, kennungsmarke))
		case strings.HasPrefix(feld, namensmarke):
			fund.Name = strings.TrimPrefix(feld, namensmarke)
		}
	}

	// **Ein Name, den kein Mensch lesen kann, fällt weg — der Fund nicht.**
	// Er steht in der Liste „wer ist da", und ein U+202E verdrehte, was dort
	// steht; gefunden von `FuzzDeute` (QA-06, #1737). Dieselbe Regel wie beim
	// Namen einer Einladung, und die Kennung trägt den Fund auch ohne ihn.
	if !domain.LesbarerName(fund.Name) {
		fund.Name = ""
	}

	// **Die Kennung wird gelesen und nicht geglaubt.** Wer auf
	// `_mse4mac._tcp` etwas antwortet, das keine ist, hat hier nichts
	// verloren — und ein Eintrag mit leerem Schlüssel wäre einer, den
	// niemand je bestätigen kann.
	//
	// **Geprüft über den Fingerabdruck**, weil [peer.Fingerabdruck] für
	// eine unlesbare Kennung die leere Zeichenkette gibt: dieselbe Prüfung
	// wie beim Aufnehmen eines Kontakts, und damit dieselbe Antwort.
	if peer.Fingerabdruck(fund.Kennung) == "" || fund.Port < 1 {
		return Fund{}, false
	}

	fund.Wirt = wirtAus(antwort)
	if fund.Wirt == "" {
		return Fund{}, false
	}

	// **Der Name kommt dazu, wenn es einen gibt** (#1633) — und **nicht**,
	// wenn er schon der Wirt ist: dann stünde dieselbe Angabe zweimal, und
	// `Merkadresse` gäbe eine Auskunft, die sich als zweite ausgäbe.
	if name := namensform(antwort.Host); name != "" && name != fund.Wirt {
		fund.Wirtname = name
	}

	return fund, true
}

// stilleLogger schluckt, was die Bibliothek sagen will.
//
// # Warum das kein Verschlucken ist
//
// **Weil sie über jede fremde Antwort im Netz schriebe.** `hashicorp/mdns`
// protokolliert Antworten, die es nicht deuten kann — und auf einem WLAN sind
// das Drucker, Fernseher und Lautsprecher. Was **dieses** Programm davon
// angeht, steht in [Ergebnis.Verworfen], und das ist eine Zahl, die ein Mensch
// sieht.
//
// **Und der Protokollwächter verlangt es**: `internal/entdeckung` ist keine
// angemeldete Protokollstelle, und eine Bibliothek, die auf `log.Default()`
// schreibt, ginge an ihm vorbei.
func stilleLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

// wirtAus gibt die Adresse, unter der **gewählt** wird.
//
// **IPv4 vor IPv6 und beides vor dem Namen.** Ein `.local`-Name müsste noch
// einmal aufgelöst werden, und wer schon eine Adresse hat, braucht das nicht.
//
// **Was sich zu **merken** lohnt, ist eine andere Frage** — sie beantwortet
// [Fund.Merkadresse] (#1633).
func wirtAus(antwort *mdns.ServiceEntry) string {
	if antwort.AddrV4 != nil {
		return antwort.AddrV4.String()
	}

	if antwort.AddrV6 != nil {
		return antwort.AddrV6.String()
	}

	return namensform(antwort.Host)
}

// namensform gibt den `.local`-Namen ohne den Punkt der Drahtform.
//
// **Der Punkt gehört zum Protokoll und nicht zum Namen.** `probe.local.` ist
// eine absolute DNS-Angabe; wer sie so weiterreichte, schriebe in eine
// Kontaktadresse einen Namen, den niemand tippt.
func namensform(wirt string) string {
	return strings.TrimSuffix(wirt, ".")
}

// deuteSuchfehler sagt, was ein gescheiterter Ruf bedeutet.
//
// # Warum eine Deutung und nicht der rohe Fehler
//
// **Weil der rohe Fehler niemandem hilft.** Die Sichtprüfung am 2026-09-14
// zeigte „Es ist ein unerwarteter Fehler aufgetreten", und der Grund stand
// nur im Protokoll:
//
//	write udp4 0.0.0.0:52316->224.0.0.251:5353: sendto: no route to host
//
// `224.0.0.251` ist die Multicastadresse von mDNS. macOS gibt für sie
// `EHOSTUNREACH`, solange die Anwendung die Erlaubnis **Lokales Netzwerk**
// nicht hat — und das ist mit zwei Klicks erledigt, wenn man es weiß.
//
// # Warum über `errors.Is` und nicht über den Text
//
// **Weil ein Zeichenkettenvergleich hier dreimal falsch wäre** (§8): die
// Meldung ist plattformabhängig, sie ist übersetzt, und sie ändert sich mit
// dem Betriebssystem. `syscall.EHOSTUNREACH` ist dieselbe Zahl in jeder
// Fassung.
//
// **`ENETUNREACH` zählt mit**: ein Rechner ohne Route für Multicast — etwa
// ohne aktive Schnittstelle — gibt sie statt der ersten, und der Rat ist
// derselbe.
//
// **Seit NET-02 (#1706) steht `ENETUNREACH` für sich** ([domain.ErrOhneNetzroute]):
// „kein Weg ins Netz" ist belegt, und der Rat ist ein anderer. Für
// `EHOSTUNREACH` und `EPERM` bleibt es bei zwei Lesarten — fehlende
// Erlaubnis **oder** ein Netz ohne Multicast —, und der Satz dazu nennt
// beide, statt eine als die übliche auszugeben.
func deuteSuchfehler(fehler error) error {
	return deuteNetzfehler(fehler, "im Netz nach Gegenüber fragen")
}

// deuteNetzfehler deutet einen Fehler der Suche **und** der Ankündigung —
// derselbe Weg für beide (S13.NAT-03, #1706).
func deuteNetzfehler(fehler error, was string) error {
	switch {
	case errors.Is(fehler, syscall.ENETUNREACH):
		return domain.MitGrund(domain.ErrOhneNetzroute, fehler, "%s", was)

	case errors.Is(fehler, syscall.EHOSTUNREACH), errors.Is(fehler, syscall.EPERM):
		return domain.MitGrund(domain.ErrOhneNetzerlaubnis, fehler, "%s", was)
	}

	return domain.MitGrund(domain.ErrValidation, fehler, "%s", was)
}
