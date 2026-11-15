package domain

// **Ein fremder Dienst, der nicht antwortet — in sechs Fällen** — D02
// (#2233).
//
// Bis dahin gingen Transportfehler roh hinaus, und das Fenster zeigte für
// jeden „Es ist ein unerwarteter Fehler aufgetreten". **Die sechs verlangen
// verschiedenen Rat**: ein Name, der sich nicht auflöst, ist eine Frage an
// Netz, DNS oder Proxy; ein Port, an dem niemand annimmt, eine an den Dienst;
// keine Antwort binnen der Frist heißt warten und noch einmal; eine Antwort,
// die stockt, eine Verbindung, die unterwegs hängt; eine Verbindung, die
// abreißt, eine Gegenstelle, die mittendrin auflegt; ein Proxy, der eine
// Anmeldung verlangt, eine an seine Einstellung. Den sechsten gibt es seit
// #2487:
// bis dahin hieß er „unerwarteter Fehler".
//
// **Alle sechs umschließen [ErrGegenstelle]** — wer nur wissen will, dass es
// am fremden Dienst lag, fragt danach. Eingeordnet werden sie in
// `internal/netz`, nach Fehlertyp und Phase, nicht nach Meldungstext (§8).

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrGegenstelle heißt: ein fremder Dienst war nicht zu erreichen oder
	// antwortete nicht.
	ErrGegenstelle = errors.New("der fremde Dienst ist nicht erreichbar")

	// ErrNameUnaufloesbar heißt: sein Name ließ sich nicht auflösen.
	ErrNameUnaufloesbar = fmt.Errorf("%w: sein Name ließ sich nicht auflösen", ErrGegenstelle)

	// ErrNichtErreichbar heißt: an seiner Adresse nimmt niemand an.
	ErrNichtErreichbar = fmt.Errorf("%w: an seiner Adresse nimmt niemand an", ErrGegenstelle)

	// ErrKeineAntwort heißt: er antwortete nicht binnen der Frist.
	ErrKeineAntwort = fmt.Errorf("%w: er antwortete nicht rechtzeitig", ErrGegenstelle)

	// ErrAntwortStockt heißt: seine Antwort blieb mittendrin stehen.
	ErrAntwortStockt = fmt.Errorf("%w: seine Antwort blieb stehen", ErrGegenstelle)

	// ErrVerbindungAbgerissen heißt: er nahm die Verbindung an und schloss
	// sie, bevor seine Antwort ganz da war — vor dem Kopf oder mitten im
	// Rumpf.
	ErrVerbindungAbgerissen = fmt.Errorf("%w: er brach die Verbindung ab", ErrGegenstelle)

	// ErrProxyAnmeldung heißt: der Proxy dazwischen verlangt eine Anmeldung
	// (`407`). **Der Wirt ist dann der Proxy**, nicht der Dienst.
	ErrProxyAnmeldung = fmt.Errorf("%w: der Proxy verlangt eine Anmeldung", ErrGegenstelle)
)

// Gegenstellenfehler trägt, **wer** nicht antwortete und welche Frist griff —
// für den Satz des Fensters (`{wirt}`) und fürs Protokoll.
//
// **Der Wirt ist Name und Port**, ohne Pfad und ohne Abfrage: in einer
// Abfrage steht bei manchem Dienst ein Schlüssel, und der gehört in keine
// Meldung.
type Gegenstellenfehler struct {
	Wirt  string
	Frist time.Duration

	// Bis ist das Ende einer Drosselung — Nullwert, wenn die Gegenstelle
	// keines nennt (R01, #2333).
	// **Als Wert und nicht im Satz**: das Fenster übersetzt den Satz weg und
	// setzt die Uhrzeit in seinen eigenen.
	Bis time.Time

	innen error
}

// NeuerGegenstellenfehler verbindet einen der sechs Fälle mit Wirt, Frist und
// der technischen Ursache.
func NeuerGegenstellenfehler(fall, grund error, wirt string, frist time.Duration) error {
	satz := wirt
	if frist > 0 {
		satz = fmt.Sprintf("%s, nach %v", wirt, frist)
	}

	return &Gegenstellenfehler{Wirt: wirt, Frist: frist, innen: MitGrund(fall, grund, "%s", satz)}
}

// NeueDrosselung meldet, dass eine Gegenstelle bis `bis` keine Anfrage
// annimmt — [ErrGedrosselt] mit Wirt und Ende.
func NeueDrosselung(grund error, wirt string, bis time.Time) error {
	return &Gegenstellenfehler{
		Wirt: wirt, Bis: bis,
		innen: MitGrund(ErrGedrosselt, grund, "%s nimmt bis %s keine Anfrage an",
			wirt, bis.Local().Format("15:04")),
	}
}

func (g *Gegenstellenfehler) Error() string { return g.innen.Error() }

// Unwrap gibt Fall **und** Grund weiter — über [MitGrund].
func (g *Gegenstellenfehler) Unwrap() error { return g.innen }
