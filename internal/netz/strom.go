// **Ein Strom statt eines Rumpfes**
// (#1100).
//
// # Warum es beides gibt
//
// [Hole] liest die ganze Antwort in den Arbeitsspeicher und deckelt sie — das
// ist für eine JSON-Antwort von wenigen Kilobyte genau richtig, und der Deckel
// ist die Absicherung gegen eine Gegenstelle, die nicht aufhört.
//
// Für ein Abbild von **188 MB** taugt es nicht: der Deckel müsste größer sein
// als jede Datei, die je kommt, und der Speicher trüge sie ganz. [Strom] gibt
// den Rumpf deshalb **ungelesen** heraus; wer ihn nimmt, schreibt ihn weg und
// zählt dabei mit.
//
// # Und warum durch dieselbe Naht
//
// Weil ein zweiter Weg ins Netz die Drossel nicht kennte — und das ist die
// Umgehung, die §16 Satz 4 verbietet. Dieselbe Drossel, dieselbe Deutung eines
// 429, derselbe `User-Agent`.

package netz

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Stromantwort ist eine Antwort, deren Rumpf noch offen ist.
//
// **Wer sie bekommt, schließt sie.** Ein nicht geschlossener Rumpf hält die
// Verbindung, und bei einer Datei dieser Größe hält er sie lange.
type Stromantwort struct {
	Status int

	// Laenge ist `Content-Length` — **−1**, wenn die Gegenstelle keine nennt.
	//
	// **Nicht 0.** Eine Null wäre eine Länge, und ein Fortschritt rechnete
	// daraus einen Anteil von unendlich. Eine Gegenstelle mit
	// `Transfer-Encoding: chunked` nennt keine, und das ist erlaubt.
	Laenge int64

	Rumpf io.ReadCloser
	Kopf  http.Header
}

// Strom stellt eine Anfrage und gibt den Rumpf ungelesen heraus.
func Strom(
	ctx context.Context, klient *Klient, drossel *Drossel,
	kenner Kenner, adresse string,
) (Stromantwort, error) {
	// **Die Erlaubnis vor der Drossel** — dieselbe Reihenfolge und derselbe
	// Grund wie in [holen]: wer nicht fragen darf, soll nicht erst warten.
	//
	// **Hier hat sie bis zum 2026-09-11 gefehlt.** Ein Strom ging hinaus,
	// obwohl der Dienst abgeschaltet war; aufgefallen ist es am Bulkbezug der
	// Preise (A3, #1442), dem ersten Dienst **mit Schalter**, der strömt.
	if fehler := klient.darfFragen(ctx, adresse); fehler != nil {
		return Stromantwort{}, fehler
	}

	if drossel != nil {
		if fehler := drossel.Warte(ctx); fehler != nil {
			return Stromantwort{}, fehler
		}
	}

	gestellt, err := http.NewRequestWithContext(ctx, http.MethodGet, adresse, http.NoBody)
	if err != nil {
		return Stromantwort{}, fmt.Errorf("die Anfrage an %s bauen: %w", adresse, err)
	}

	kenner.setzeKopf(gestellt)

	antwort, fehler := klient.unten.Do(gestellt)
	if fehler != nil {
		return Stromantwort{}, fmt.Errorf("%s fragen: %w", adresse,
			klient.mitProxyhinweis(gestellt, fehler))
	}

	// **Ein 429 sperrt die Drossel wirklich** — dieselbe Behandlung wie in
	// [Hole], und aus demselben Grund: ein Meldungstext ist keine Antwort auf
	// „zu viele Anfragen".
	if antwort.StatusCode == http.StatusTooManyRequests {
		_ = antwort.Body.Close() //nolint:errcheck // der Rumpf eines 429 trägt nichts, was zu retten wäre

		sperreSetzen(drossel, antwort.Header)

		return Stromantwort{Status: antwort.StatusCode, Kopf: antwort.Header},
			fmt.Errorf("%w: %s — %s", ErrGedrosselt, adresse,
				wartehinweis(antwort.Header))
	}

	return Stromantwort{
		Status: antwort.StatusCode,
		Laenge: antwort.ContentLength,
		Rumpf:  antwort.Body,
		Kopf:   antwort.Header,
	}, nil
}
