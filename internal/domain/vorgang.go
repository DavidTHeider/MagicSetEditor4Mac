// **Die Kennung eines Vorgangs und sein Ausgang** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Wozu
//
// Ein Fehler, den jemand sieht, und die Zeile, die ihn erklärt, sind zwei
// Dinge an zwei Orten. Ohne eine gemeinsame Kennung verbindet sie nur die
// Uhrzeit — und die reicht nicht, sobald zwei Vorgänge nebeneinander laufen.
//
// # Eine, nicht drei
//
// Das Issue verbietet ausdrücklich „drei bedeutungsgleiche Pflicht-IDs". Es
// gibt deshalb **eine** Kennung: sie entsteht an der Grenze, an der ein Ruf in
// diese Anwendung kommt, liegt im `context`, steht in jeder Protokollzeile
// dieses Rufs und reist mit seiner Antwort zurück. Keine Spans — die
// Ablaufspur aus OPS-05 (#1722) ist eine Zeile je Ruf mit dieser Kennung,
// kein zweiter Name für denselben Vorgang.

package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// Vorgangskennung benennt **einen** Ruf in diese Anwendung.
type Vorgangskennung string

// vorgangsbytes ist die Länge der Kennung in Bytes; hexadezimal steht sie mit
// der doppelten Zahl Zeichen da.
//
// **Vier Bytes, also acht Zeichen.** Sie soll jemand vorlesen und in einem
// Protokoll suchen können; eine UUID kann das nicht. Was sie leisten muss,
// ist die Unterscheidung der Rufe **einer Sitzung**, nicht die Eindeutigkeit
// über alle Rechner: bei 4.294.967.296 Möglichkeiten und tausend Rufen am Tag
// liegt die Wahrscheinlichkeit einer Doppelung unter 0,012 % (Geburtstag,
// n²/2N). Und selbst dann sind es zwei Vorgänge zu verschiedenen Zeiten —
// eine Kennung ist ein Griff ins Protokoll, kein Schlüssel im Bestand.
const vorgangsbytes = 4

// NeueVorgangskennung würfelt eine Kennung.
//
// **Ohne Fehlerrückgabe, und das ist nachgelesen**: `crypto/rand.Read` „never
// returns an error, and always fills b entirely" (`crypto/rand`, Go 1.26) —
// es gerät in Panik, wenn der Zufall des Betriebssystems ausfällt, und dann
// ist eine Protokollkennung das kleinere Problem.
func NeueVorgangskennung() Vorgangskennung {
	roh := make([]byte, vorgangsbytes)
	_, _ = rand.Read(roh) //nolint:errcheck // siehe oben

	return Vorgangskennung(hex.EncodeToString(roh))
}

// Ausgang ist, wie ein geprüfter Ruf ausgegangen ist.
//
// **Zwei Werte, und das ist der Punkt**: er ist eine Dimension der Messung,
// und eine Messdimension mit unbegrenztem Wertevorrat ist der Fehler, den
// [Messpunkt] verhindert.
type Ausgang string

const (
	// AusgangGewaehrt: der Riegel hat den Ruf durchgelassen.
	AusgangGewaehrt Ausgang = "gewaehrt"

	// AusgangAbgewiesen: der Riegel hat ihn abgewiesen.
	AusgangAbgewiesen Ausgang = "abgewiesen"
)

// Ausgaenge gibt beide.
func Ausgaenge() []Ausgang { return []Ausgang{AusgangGewaehrt, AusgangAbgewiesen} }

// VorgangskennungAus liest eine Kennung, die ein Aufrufer mitbringt — OPS-05.AK-02
// (#1722).
//
// **Nur in der Gestalt, die dieses Programm selbst vergibt**, und etwas
// Spielraum darüber: acht bis zweiunddreißig Kleinbuchstaben aus dem
// Sechzehnersystem. So kann ein Werkzeug davor seine eigene Kennung
// weiterreichen, ohne dass beliebiger Text in jede Protokollzeile gelangt —
// ein Zeilenumbruch darin fälschte eine Zeile.
func VorgangskennungAus(text string) (Vorgangskennung, bool) {
	if len(text) < 2*vorgangsbytes || len(text) > 32 {
		return "", false
	}

	for _, zeichen := range text {
		if (zeichen < '0' || zeichen > '9') && (zeichen < 'a' || zeichen > 'f') {
			return "", false
		}
	}

	return Vorgangskennung(text), true
}

// Rufausgang ist, wie ein Ruf über eine Fläche ausging — OPS-05.AK-01.
//
// **Drei Werte und kein Statuscode**: die Fehlerrate, nach der das Issue
// fragt, trennt, wer den Fehler verschuldet hat. Abgewiesen ist ein Ruf, den
// der Aufrufer anders stellen müsste; gescheitert einer, an dem der Dienst
// nichts ändern konnte.
type Rufausgang string

// Die drei Ausgänge.
const (
	RufGelungen    Rufausgang = "gelungen"
	RufAbgewiesen  Rufausgang = "abgewiesen"
	RufGescheitert Rufausgang = "gescheitert"
)

// Rufausgaenge gibt alle drei.
func Rufausgaenge() []Rufausgang { return []Rufausgang{RufGelungen, RufAbgewiesen, RufGescheitert} }
