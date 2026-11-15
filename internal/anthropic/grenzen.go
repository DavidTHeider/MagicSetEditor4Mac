package anthropic

// **Was der Dienst über seine Grenzen sagt** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Warum sie an jeder Antwort hängen
//
// **Weil das Issue sie weitergeben will**: „Anbieterspezifische
// Kosten-/Usage-/Rate-Metadaten an 5.2 weitergeben." Das Kostenjournal aus
// O12 rechnet in Geld; was hier ankommt, sind **Zeichen** und **Anteile** —
// und wer daraus Geld machte, erfände einen Tarif.
//
// **Nachgelesen am 2026-09-11** unter
// <https://platform.claude.com/docs/en/api/rate-limits>.

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
	// **Und „keine Angabe" ist bei einem 429 eine Aussage**: die Sperre am
	// Ausgabendeckel trägt sie nicht und geht nicht vorüber
	// ([dieAusgabensperre]).
	Warten time.Duration

	// Anfragen sind die drei Zahlen zur Anfragegrenze.
	Anfragen Kontingent

	// Eingabezeichen und Ausgabezeichen sind dieselben je Zeichenart.
	Eingabezeichen Kontingent
	Ausgabezeichen Kontingent
}

// Kontingent ist Grenze, Rest und Zeitpunkt der Auffüllung.
type Kontingent struct {
	Grenze int
	Rest   int

	// Frisch ist, wann wieder aufgefüllt ist — die Nullzeit heißt: keine
	// Angabe.
	Frisch time.Time

	// Genannt sagt, ob der Dienst überhaupt etwas dazu gesagt hat.
	//
	// **Ohne dieses Feld sähe „keine Angabe" aus wie „null übrig"** — und
	// das ist der Unterschied zwischen „gleich gesperrt" und „weiß nicht".
	Genannt bool
}

// grenzenAus liest sie aus den Kopfzeilen.
func grenzenAus(kopf netz.Kopfzeilen) Grenzen {
	return Grenzen{
		Warten:         wartenAus(kopf),
		Anfragen:       kontingentAus(kopf, "requests"),
		Eingabezeichen: kontingentAus(kopf, "input-tokens"),
		Ausgabezeichen: kontingentAus(kopf, "output-tokens"),
	}
}

// wartenAus liest `retry-after`.
//
// **In Sekunden.** Die Dokumentation sagt es so: „The number of seconds to
// wait until you can retry the request."
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
//
// **Über ein Präfix und nicht über neun Namen**: `anthropic-ratelimit-<art>-`
// trägt `limit`, `remaining` und `reset`, und die Namen unterscheiden sich
// allein in der Art.
func kontingentAus(kopf netz.Kopfzeilen, art string) Kontingent {
	praefix := "anthropic-ratelimit-" + art + "-"

	grenze, hatGrenze := zahlAus(kopf, praefix+"limit")
	rest, hatRest := zahlAus(kopf, praefix+"remaining")

	// **Der Zeitpunkt ist RFC 3339**, nicht eine Zahl von Sekunden — das ist
	// der Unterschied zu `retry-after`, und er ist leicht zu übersehen.
	//
	// **Ein unlesbarer bleibt die Nullzeit**, und die heißt „keine Angabe"
	// ([Kontingent.Frisch]) — derselbe Ausgang wie eine fehlende Kopfzeile.
	frisch, _ := time.Parse(time.RFC3339, kopf.Get(praefix+"reset")) //nolint:errcheck // siehe oben

	return Kontingent{
		Grenze: grenze, Rest: rest, Frisch: frisch,
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

// Verbrauch ist, was ein Ruf gekostet hat — in Zeichen.
//
// **Nicht in Geld.** Ein Tarif steht nicht in der Antwort, und ihn
// auszurechnen hieße, einen Preis zu erfinden, der sich ändern kann, ohne
// dass dieses Programm es merkt.
type Verbrauch struct {
	Eingabe int `json:"input_tokens"`
	Ausgabe int `json:"output_tokens"`

	// ZwischenspeicherGeschrieben und ZwischenspeicherGelesen sind die zwei
	// Zahlen des Prompt-Zwischenspeichers.
	//
	// **Sie zählen verschieden**: das Schreiben zählt gegen die
	// Eingabegrenze, das Lesen bei den meisten Modellen **nicht**. Wer nur
	// `input_tokens` läse, sähe bei zwischengespeichertem Inhalt einen
	// Bruchteil dessen, was wirklich hineinging.
	ZwischenspeicherGeschrieben int `json:"cache_creation_input_tokens"`
	ZwischenspeicherGelesen     int `json:"cache_read_input_tokens"`
}

// Eingabezeichen gibt die Zeichen, die wirklich hineingingen.
//
// **Nachgelesen am 2026-09-11**: „`total_input_tokens = cache_read_input_tokens
// + cache_creation_input_tokens + input_tokens`" — `input_tokens` sind allein
// die **nach** der letzten Zwischenspeichermarke.
func (v Verbrauch) Eingabezeichen() int {
	return v.Eingabe + v.ZwischenspeicherGeschrieben + v.ZwischenspeicherGelesen
}
