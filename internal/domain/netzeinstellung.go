package domain

// **Wie dieses Programm ins Netz geht** — Werkstück N5 von Phase 5.4
// (#1571).
//
// # Warum die Schlüssel hier stehen und nicht bei `internal/netz`
//
// **Weil eine Einstellung eine Einstellung ist.** Sie geht durch denselben
// Weg wie jede andere — geprüft in `domain`, abgelegt in `storage`, gesetzt
// über `EinstellungSetzen` —, und damit erreicht sie **alle vier Flächen**,
// ohne dass jemand vier Wege baut. Das ist der Punkt von Werkstück 2 des
// Issues: `mse`, `mse-mcp` und `mse-rest` sind Klienten wie das Fenster.
//
// # Und warum leer etwas bedeutet
//
// **Leer heißt: die Umgebung gilt.** Das ist keine Bequemlichkeit, sondern
// die Antwort auf die Frage, die A18 offengelassen hat
// (#1533):
// „Die Einstellung schlägt die Umgebung, falls eine gesetzt ist." Wer nichts
// setzt, bekommt, was `HTTPS_PROXY` sagt; wer etwas setzt, schlägt sie.
//
// **Ein gesetzter Wert, der leer sein soll, ist deshalb nicht darstellbar** —
// und das ist richtig: „kein Proxy, auch wenn die Umgebung einen nennt"
// schreibt man als `NetzProxyAus` mit `*`.

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Die Schlüssel der Netzeinstellungen.
const (
	// NetzProxyHTTP ist der Proxy für `http://`-Adressen.
	//
	// **Leer heißt: `HTTP_PROXY` gilt.**
	NetzProxyHTTP = "netz.proxy.http"

	// NetzProxyHTTPS ist der Proxy für `https://`-Adressen.
	//
	// **Getrennt vom vorigen**, weil `httpproxy` sie getrennt führt und weil
	// viele Netze nur für eines davon einen haben.
	NetzProxyHTTPS = "netz.proxy.https"

	// NetzProxyAus nennt, was **nicht** über den Proxy geht — die Gestalt
	// von `NO_PROXY`: Rechnernamen und Adressbereiche, mit Komma getrennt,
	// `*` für alles.
	//
	// **Sie ist die Hälfte der Fähigkeit** (#1571): wer einen Proxy setzt
	// und Ollama im eigenen Netz betreibt, schickte die Einbettung sonst
	// hindurch. Die Schleifenadresse bleibt von sich aus außen vor — ein
	// Rechner **nebenan** nicht.
	NetzProxyAus = "netz.proxy.aus"

	// NetzFristVerbindung ist die Zeit für Verbindungsaufbau und TLS, in
	// **Sekunden**.
	//
	// **Drei Fristen und nicht eine** — der Grund steht bei `netz.Fristen`:
	// eine Gesamtfrist, die für eine 74-MB-Datei reicht, fängt keinen
	// hängenden Server mehr.
	NetzFristVerbindung = "netz.frist.verbindung"

	// NetzFristKopf ist die Zeit, die eine Gegenstelle für ihren Antwortkopf
	// hat, in Sekunden.
	NetzFristKopf = "netz.frist.kopf"

	// NetzFristStillstand ist die Zeit, die ein Antwortrumpf ohne
	// Fortschritt haben darf, in Sekunden.
	NetzFristStillstand = "netz.frist.stillstand"

	// NetzSchnittstelle ist die Netzwerkschnittstelle, über die Suche und
	// Ankündigung im eigenen Netz gehen — **leer heißt: automatisch**, jede
	// taugliche (NET-02.AK-03,
	// #1706).
	NetzSchnittstelle = "netz.schnittstelle"
)

// HoechsteNetzfrist ist die Obergrenze einer eingestellten Frist.
//
// **Eine Stunde, und sie ist eine Entscheidung dieses Prüfers.** Eine Frist,
// die höher steht, ist keine mehr: sie fängt keinen hängenden Server, und
// niemand wartet so lange vor einem Fenster. Wer einen Bezug über eine sehr
// langsame Leitung fährt, braucht die **Stillstandsfrist** hoch und nicht die
// Kopffrist.
const HoechsteNetzfrist = time.Hour

// PruefeProxyadresse weist ab, was kein Proxy sein kann.
//
// **Leer ist gültig** — das ist der Normalfall und heißt „die Umgebung gilt".
//
// **Und ein Schema ist Pflicht.** `httpproxy` nimmt `proxy:3128` entgegen und
// deutet es als **Rechner mit Port**; wer `http://` vergisst, bekommt
// stillschweigend etwas anderes, als er meinte. Hier endet es mit einer
// Meldung statt mit einem Rätsel (§8).
func PruefeProxyadresse(wert string) error {
	if strings.TrimSpace(wert) == "" {
		return nil
	}

	zerlegt, fehler := url.Parse(wert)
	if fehler != nil {
		return MitGrund(ErrValidation, fehler, "%q ist keine Proxyadresse", wert)
	}

	if zerlegt.Scheme != "http" && zerlegt.Scheme != "https" {
		return fmt.Errorf(
			"%w: %q nennt kein http oder https — ein Proxy ohne Schema wird "+
				"als Rechnername gelesen", ErrValidation, wert)
	}

	if zerlegt.Host == "" {
		return fmt.Errorf("%w: %q nennt keinen Rechner", ErrValidation, wert)
	}

	// **Kein Kennwort in den Einstellungen** — S13.NAT-02 (#1706): es läge
	// unverschlüsselt im Bestand und stünde in jeder Anzeige. Wer einen
	// Proxy mit Anmeldung braucht, setzt ihn über die Umgebung.
	if _, mitKennwort := zerlegt.User.Password(); mitKennwort {
		return fmt.Errorf("%w: ein Kennwort gehört nicht in die Proxyadresse — "+
			"entferne es und setze den Proxy mit Anmeldung über HTTPS_PROXY",
			ErrValidation)
	}

	return nil
}

// ProxyadresseOhneKennwort verdeckt das Kennwort einer Proxyadresse.
//
// **Bewusst ohne `url.Parse`**: eine Adresse aus der Umgebung ist nicht
// geprüft, und ein Kennwort mit `/` oder ohne Schema lässt sich nicht
// zerlegen — es stünde dann unverdeckt da. Gesucht wird das letzte `@`;
// eine Adresse mit `@` im Pfad wird dabei zu viel verdeckt, und das ist die
// richtige Seite für einen Fehler.
func ProxyadresseOhneKennwort(wert string) string {
	anfang := 0
	if stelle := strings.Index(wert, "://"); stelle >= 0 {
		anfang = stelle + len("://")
	}

	rest := wert[anfang:]

	at := strings.LastIndex(rest, "@")
	if at < 0 {
		return wert
	}

	name, _, mitKennwort := strings.Cut(rest[:at], ":")
	if !mitKennwort {
		return wert
	}

	return wert[:anfang] + name + ":xxxxx" + rest[at:]
}

// Einstellungsanzeige ist der Wert einer Einstellung, wie er gezeigt wird —
// bei einem Proxy ohne Kennwort.
func Einstellungsanzeige(schluessel, wert string) string {
	if schluessel == NetzProxyHTTP || schluessel == NetzProxyHTTPS {
		return ProxyadresseOhneKennwort(wert)
	}

	return wert
}

// PruefeNetzfrist weist ab, was keine Frist in Sekunden ist.
//
// **Null ist nicht erlaubt.** In Go heißt eine Frist von null „gar keine", und
// eine Einstellung, die eine Sicherung stillschweigend abschaltet, ist die
// Gestalt, vor der §8 warnt. Wer keine Frist will, hebt sie an.
func PruefeNetzfrist(wert string) error {
	if strings.TrimSpace(wert) == "" {
		return nil
	}

	sekunden, fehler := strconv.Atoi(wert)
	if fehler != nil {
		return MitGrund(ErrValidation, fehler,
			"%q ist keine Zahl von Sekunden", wert)
	}

	if sekunden <= 0 {
		return fmt.Errorf("%w: eine Frist von %d Sekunden ist keine — "+
			"in Go heißt null »gar keine«", ErrValidation, sekunden)
	}

	// **Über [Zeitspanne] und nicht über das Produkt** (J05, #2269): das lief
	// bei 9.223.372.037 Sekunden ins Negative und lag unter der Grenze.
	if frist, passt := Zeitspanne(int64(sekunden), time.Second); !passt || frist > HoechsteNetzfrist {
		return fmt.Errorf("%w: %d Sekunden sind mehr als die Obergrenze von "+
			"%d", ErrValidation, sekunden, int(HoechsteNetzfrist.Seconds()))
	}

	return nil
}

// Netzfrist deutet einen eingestellten Wert — die Vorgabe, wo keiner steht.
//
// **Hier und nicht in `app`**, damit jede Fläche dieselbe Zahl bekommt: die
// Kommandozeile liest dieselben Einstellungen wie das Fenster, und zwei
// Deutungen desselben Werts wären zwei Antworten auf eine Frage (§16).
func Netzfrist(wert string, vorgabe time.Duration) time.Duration {
	sekunden, fehler := strconv.Atoi(strings.TrimSpace(wert))
	if fehler != nil || sekunden <= 0 {
		return vorgabe
	}

	frist, passt := Zeitspanne(int64(sekunden), time.Second)
	if !passt {
		return vorgabe
	}

	return frist
}

// schnittstellenname ist die Gestalt eines Namens unter macOS: ein
// Buchstabe, dann Buchstaben und Ziffern — `en0`, `bridge100`, `utun3`.
// **Höchstens fünfzehn Zeichen**, so viel fasst `IFNAMSIZ` samt Null.
var schnittstellenname = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{0,14}$`)

// PruefeSchnittstellenname urteilt über eine Schnittstelle — leer ist die
// Automatik.
//
// **Ob es sie gibt, prüft das nicht**: eine Schnittstelle kommt und geht mit
// dem Netz, und ein Name, der heute fehlt, gilt morgen wieder. Wer einen
// fehlenden nennt, bekommt die Automatik und einen Befund.
func PruefeSchnittstellenname(wert string) error {
	if wert == "" || schnittstellenname.MatchString(wert) {
		return nil
	}

	return NewFieldError("wert", SchluesselValidierung,
		"»%s« ist kein Schnittstellenname — erwartet wird einer wie en0 oder bridge100.", wert)
}
