package openai

// **Was der Dienst über seine Grenzen sagt** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum sie an jeder Antwort hängen und nicht im Quelltext
//
// **Weil die Dokumentation sie nicht nennt.** Nachgelesen am 2026-09-12 unter
// <https://developers.openai.com/api/docs/guides/rate-limits>: die Grenzen
// stehen je Konto und Tarifstufe in den Kontoeinstellungen, nicht in der
// Dokumentation. Was jede Antwort mitbringt, sind sechs Kopfzeilen — und die
// sind die einzige belastbare Auskunft.
//
// # Dieselbe Gestalt wie bei Anthropic, andere Namen
//
// `x-ratelimit-limit-requests` statt `anthropic-ratelimit-requests-limit`.
// Die Reihenfolge der Wörter ist vertauscht, und **Zeichen** heißen hier
// `tokens` ohne Trennung nach Ein- und Ausgabe: es gibt eine Zahl für beide
// zusammen.

import (
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Grenzen sind die Angaben einer Antwort über Kontingent und Frist.
type Grenzen struct {
	// Warten ist `retry-after` in Sekunden — 0 heißt: keine Angabe.
	//
	// **Und „keine Angabe" ist bei einem 429 eine Aussage**: die Sperre
	// wegen erschöpften Guthabens trägt sie nicht und geht nicht vorüber
	// ([dieGeldgruende]).
	Warten time.Duration

	// Anfragen sind die drei Zahlen zur Anfragegrenze.
	Anfragen Kontingent

	// Zeichen sind dieselben für die Zeichen — **eine** Zahl für Ein- und
	// Ausgabe zusammen, anders als bei Anthropic.
	Zeichen Kontingent
}

// Kontingent ist Grenze, Rest und die Frist bis zur Auffüllung.
type Kontingent struct {
	Grenze int
	Rest   int

	// Frisch ist, **wie lange** es noch dauert — nicht, wann es so weit ist.
	//
	// **Das ist der Unterschied zu Anthropic**, und er ist leicht zu
	// übersehen: dort steht ein Zeitpunkt nach RFC 3339, hier eine Dauer wie
	// `6m0s` oder `1s`. Wer sie als Zeitpunkt läse, bekäme die Nullzeit.
	Frisch time.Duration

	// Genannt sagt, ob der Dienst überhaupt etwas dazu gesagt hat.
	Genannt bool
}

// grenzenAus liest sie aus den Kopfzeilen.
func grenzenAus(kopf netz.Kopfzeilen) Grenzen {
	return Grenzen{
		Warten:   wartenAus(kopf),
		Anfragen: kontingentAus(kopf, "requests"),
		Zeichen:  kontingentAus(kopf, "tokens"),
	}
}

// wartenAus liest `retry-after`.
//
// **In Sekunden**, und nur bei einer vorübergehenden Drossel: „The minimum
// number of seconds to wait before retrying a temporary rate-limit error,
// when present."
func wartenAus(kopf netz.Kopfzeilen) time.Duration {
	sekunden, fehler := strconv.Atoi(kopf.Get("retry-after"))
	if fehler != nil || sekunden < 0 {
		return 0
	}

	// Sättigend (J05, #2269): mehr, als eine Dauer fasst, ist das meiste
	// und keine umgelaufene, negative Wartezeit.
	return domain.ZeitspanneHoechstens(int64(sekunden), time.Second)
}

// kontingentAus liest die drei Kopfzeilen einer Art.
func kontingentAus(kopf netz.Kopfzeilen, art string) Kontingent {
	grenze, hatGrenze := zahlAus(kopf, "x-ratelimit-limit-"+art)
	rest, hatRest := zahlAus(kopf, "x-ratelimit-remaining-"+art)

	return Kontingent{
		Grenze: grenze, Rest: rest,
		Frisch:  dauerAus(kopf.Get("x-ratelimit-reset-" + art)),
		Genannt: hatGrenze || hatRest,
	}
}

// zahlAus liest eine Kopfzeile als Zahl.
func zahlAus(kopf netz.Kopfzeilen, name string) (int, bool) {
	roh := kopf.Get(name)
	if roh == "" {
		return 0, false
	}

	zahl, fehler := strconv.Atoi(roh)
	if fehler != nil {
		return 0, false
	}

	return zahl, true
}

// dauerAus liest eine Dauer in Gos Schreibweise — `6m0s`, `1s`, `88ms`.
//
// **Unlesbar heißt null**, und null heißt „keine Angabe": eine geratene
// Dauer wäre schlechter als keine.
func dauerAus(roh string) time.Duration {
	dauer, fehler := time.ParseDuration(roh)
	if fehler != nil || dauer < 0 {
		return 0
	}

	return dauer
}

// Verbrauch ist, was ein Ruf gekostet hat — in Zeichen.
//
// **Nicht in Geld.** Ein Tarif steht nicht in der Antwort, und ihn
// auszurechnen hieße, einen Preis zu erfinden, der sich ändern kann, ohne
// dass dieses Programm es merkt.
type Verbrauch struct {
	Eingabe      int               `json:"input_tokens"`
	Ausgabe      int               `json:"output_tokens"`
	Zusammen     int               `json:"total_tokens"`
	Eingabeteile verbrauchseingabe `json:"input_tokens_details"`
	Ausgabeteile verbrauchsausgabe `json:"output_tokens_details"`
}

// verbrauchseingabe sind die Teile der Eingabe.
type verbrauchseingabe struct {
	AusDemZwischenspeicher int `json:"cached_tokens"`
}

// verbrauchsausgabe sind die Teile der Ausgabe.
//
// **`reasoning_tokens` zählen mit und sind unsichtbar**: sie erscheinen in
// keiner Antwort, die jemand liest, und werden trotzdem bezahlt.
type verbrauchsausgabe struct {
	ZumDenken int `json:"reasoning_tokens"`
}
