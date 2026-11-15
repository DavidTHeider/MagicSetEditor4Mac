// Die Karten eines ganzen Satzes
// (#1033).
//
// **Eine eigene Datei, weil es eine eigene Frage ist**: `Suche` gibt **eine**
// Seite und sagt, dass es mehr gibt — hier wird geblättert.

package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Seitengrenze ist, wie viele Seiten höchstens geblättert werden.
//
// # Warum es eine gibt
//
// **Weil ein Dienst, der immer `has_more` sagt, sonst endlos liefe.** Das ist
// keine Unterstellung, sondern die Regel aus §8: fremde Daten sind der
// Normalfall, nicht der Ausnahmefall.
//
// **Vierzig Seiten sind 7.000 Karten.** Nachgelesen am 2026-08-24 unter
// `https://scryfall.com/docs/api/cards/search`: „This method is paginated,
// returning 175 cards at a time."
//
// # Warum vierzig und nicht zehn
//
// Hier standen zehn — 1.750 Karten — mit der Begründung, der größte
// Magic-Satz liege „weit darunter". **Das stimmt nicht**, und es ist am
// 2026-08-24 an `https://api.scryfall.com/sets` nachgemessen worden: von 1.047
// Sätzen liegen drei darüber.
//
//	plst   5.120  The List
//	prm    3.094  Magic Online Promos
//	sld    2.678  Secret Lair Drop
//
// Vierzig Seiten decken den größten mit Luft. Die Grenze bleibt, weil sie
// gegen einen Dienst schützt, der immer `has_more` sagt — sie sitzt jetzt nur
// über der Wirklichkeit statt darin.
//
// **Sie ist exportiert**, damit ein Test gegen sie prüfen kann, statt die Zahl
// zu wiederholen — eine zweite Zahl liefe beim ersten Nachziehen auseinander.
const Seitengrenze = 40

// Satz sind die Karten eines Satzes.
type Satz struct {
	Karten []*Karte

	// Gesamt ist, wie viele Karten Scryfall über alle Seiten nennt.
	//
	// **Sie steht neben der Liste**, damit ein Aufrufer erkennt, wie viele es
	// hätten sein sollen — die Liste allein sagt nur, wie viele ankamen.
	Gesamt int

	// Abgebrochen sagt, dass die [Seitengrenze] gegriffen hat und Scryfall
	// **weitere** Karten meldet.
	//
	// # Warum das ein eigenes Feld ist und kein Rechenweg
	//
	// Ein Aufrufer könnte `len(Karten) < Gesamt` rechnen, und das wäre eine
	// zweite Wahrheit neben dieser: `Gesamt` zählt, was Scryfall auf die Frage
	// antwortet, nicht was hier ankam. Wer aus zwei Zahlen einen Zustand
	// erschließt, erschließt ihn beim nächsten Randfall falsch.
	//
	// **Und der Befund daneben genügt nicht.** Ein Befund ist eine Meldung für
	// Menschen; ein Aufrufer, der daraus eine Entscheidung ableiten wollte,
	// müsste auf ihren Wortlaut prüfen
	// (#1042).
	Abgebrochen bool
}

// Setkarten holt alle Karten eines Satzes.
//
// # Warum das blättert und `Suche` nicht
//
// `Suche` gibt eine Seite und meldet `Weitere` — mit Absicht: eine Suche, die
// von selbst durch alle Seiten liefe, stellte ohne Zutun beliebig viele
// Anfragen, und die Bedingungen sagen, dass für **viele** Karten die
// Bulk-Datei zu nehmen ist.
//
// **Ein Satz ist der Fall, in dem man genau das will**, und er ist begrenzt:
// unter tausend Karten, also unter sechs Seiten. Das ist kein Massenabruf.
//
// # Die Drossel greift zwischen den Seiten
//
// Jede Seite geht durch [Klient.hole] und damit durch die Kartendrossel —
// 600 ms Abstand, unter den erlaubten 2/s (§16, Regel 4). Ein Satz mit 300
// Karten kostet damit zwei Anfragen und rund eine halbe Sekunde.
//
// # Warum die Folgeadresse nicht nachgebaut wird
//
// **Scryfall gibt sie**: „If there is a page beyond the current page, this
// field will contain a full API URI to that page." Wer sie selbst
// zusammensetzte, riete die Blätterung — und läge falsch, sobald der Dienst
// sie ändert.
func (k *Klient) Setkarten(ctx context.Context, code string) (
	satz Satz, befunde []domain.Befund, fehler error,
) {
	// **`e:` und nicht `set:`** — beides geht, `e` ist die kurze Form. Dazu
	// `unique=prints`, damit ein Satz seine Karten so führt, wie er sie
	// gedruckt hat, und `order=set` für die Sammlernummer.
	adresse := k.grundadresse + "/cards/search?q=" + url.QueryEscape("e:"+code) +
		"&unique=prints&order=set"

	for seite := 1; seite <= Seitengrenze; seite++ {
		roh, holfehler := k.hole(ctx, adresse, "die Karten dieses Satzes")
		if holfehler != nil {
			return Satz{}, nil, holfehler
		}

		gelesen, weiter, seitenbefunde, lesefehler := seiteLesen(roh)
		if lesefehler != nil {
			return Satz{}, nil, lesefehler
		}

		satz.Karten = append(satz.Karten, gelesen.Karten...)
		satz.Gesamt = gelesen.Gesamt
		befunde = append(befunde, seitenbefunde...)

		if weiter == "" {
			return satz, befunde, nil
		}

		adresse = weiter
	}

	// **Die Grenze sagt es** (§10): ein Aufrufer, der nur die Liste sieht,
	// hielte sie für vollständig — **im Befund für Menschen und im Feld für
	// den Aufrufer** (#1042).
	satz.Abgebrochen = true

	return satz, append(befunde, domain.Befund{
		Ort: "cards/search",
		Meldung: fmt.Sprintf(
			"nach %d Seiten abgebrochen — Scryfall meldet weitere Karten zu %q",
			Seitengrenze, code),
	}), nil
}

// seiteLesen zerlegt eine Listenantwort.
//
// **`object` ist Pflicht, `warnings` ist es nicht** (§16): ein fehlendes
// Pflichtfeld bricht laut, ein fehlendes Zusatzfeld gar nicht.
func seiteLesen(roh []byte) (satz Satz, weiter string, befunde []domain.Befund, fehler error) {
	var liste struct {
		Objekt    string            `json:"object"`
		Weitere   bool              `json:"has_more"`
		Naechste  string            `json:"next_page"`
		Gesamt    int               `json:"total_cards"`
		Warnungen []string          `json:"warnings"`
		Daten     []json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(roh, &liste); err != nil {
		return Satz{}, "", nil, domain.MitGrund(domain.ErrValidation, err,
			"die Satzantwort von Scryfall ist kein gültiges JSON")
	}

	if liste.Objekt != "list" {
		return Satz{}, "", nil, fmt.Errorf(
			"%w: die Satzantwort von Scryfall hat kein Feld %q — das ist ein Pflichtfeld",
			domain.ErrValidation, "object")
	}

	// **Die Warnungen der Antwort sind Befunde, nicht Fehler** — nachgelesen
	// am 2026-08-24 unter `https://scryfall.com/docs/api/lists`: „non-fatal
	// issues that the API discovered with your input […] they indicate that
	// the List will not contain all of the information you requested."
	for _, warnung := range liste.Warnungen {
		befunde = append(befunde, domain.Befund{Ort: "cards/search", Meldung: warnung})
	}

	satz.Gesamt = liste.Gesamt
	satz.Karten = make([]*Karte, 0, len(liste.Daten))

	for _, eine := range liste.Daten {
		gelesen, weitere, lesefehler := LiesKarte(eine)
		if lesefehler != nil {
			return Satz{}, "", nil, lesefehler
		}

		satz.Karten = append(satz.Karten, gelesen)
		befunde = append(befunde, weitere...)
	}

	if liste.Weitere {
		weiter = liste.Naechste
	}

	return satz, weiter, befunde, nil
}
