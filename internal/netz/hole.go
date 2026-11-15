package netz

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ErrGedrosselt sagt, dass die Gegenstelle zu viele Anfragen gemeldet hat.
//
// **Er steht in `domain`** (seit dem 2026-08-24): sein Rat ist eine fachliche
// Auskunft — „warte" —, und `internal/service` muss ihn unterscheiden können,
// ohne den Transport zu kennen (§2). Hier bleibt der Name, damit ein Aufrufer
// dieses Pakets ihn nicht suchen muss.
var ErrGedrosselt = domain.ErrGedrosselt

// Kenner ist, womit sich dieses Programm vorstellt.
//
// **Pflicht, nicht Höflichkeit.** Nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`: „All HTTP requests to api.scryfall.com must
// include a User-Agent header", und „Your User-Agent header must be accurate to
// your usage context."
type Kenner struct {
	// Name ist der `User-Agent`.
	Name string
	// Annahme ist der `Accept`.
	//
	// **Auch er ist Pflicht**: „The Accept header must be present, but you can
	// provide a generic preference." `*/*` ist ausdrücklich erlaubt.
	Annahme string

	// Marke ist die Zugangsmarke — leer, wenn es keine gibt.
	//
	// **Sie steht am Kenner, weil ein Kenner genau das ist**: die Art, wie
	// dieser Klient sich vorstellt. Ein privates Repositorium antwortet ohne
	// sie mit 404, und das ist eine Auskunft über die Vorstellung, nicht über
	// den Inhalt.
	//
	// **Leer heißt: keine Kopfzeile.** Ein leeres `Authorization` ist etwas
	// anderes als keines — manche Dienste weisen es ab, statt die Anfrage als
	// unangemeldet zu behandeln.
	//
	// **Sie wird nirgends ausgegeben.** Keine Meldung dieses Pakets nennt
	// Kopfzeilen; was in einem Fehler steht, ist die Adresse.
	Marke string

	// Zusatz sind weitere Kopfzeilen, mit denen sich dieses Programm
	// vorstellt — Phase 4.8
	// (#1321).
	//
	// # Warum es sie neben [Kenner.Marke] gibt
	//
	// **Weil nicht jeder Dienst `Authorization` nimmt.** Scrydex verlangt
	// zwei eigene: `X-Api-Key` **und** `X-Team-ID` (nachgelesen am
	// 2026-09-02). Ein Kenner mit drei festen Feldern trüge das nicht, und
	// ein zweiter Weg neben [Hole] hätte die Hälfte seiner Zusagen —
	// Drossel, Frist, das 429 und die Rumpfgrenze.
	//
	// **Ein leerer Wert setzt keine Kopfzeile**, aus demselben Grund wie bei
	// der Marke: ein leeres `X-Api-Key` ist etwas anderes als keines.
	//
	// **Auch sie wird nirgends ausgegeben.**
	Zusatz map[string]string
}

// setzeKopf schreibt die Kopfzeilen, die einen Kenner ausmachen.
func (k Kenner) setzeKopf(gestellt *http.Request) {
	gestellt.Header.Set("User-Agent", k.Name)
	gestellt.Header.Set("Accept", k.Annahme)

	if k.Marke != "" {
		gestellt.Header.Set("Authorization", "Bearer "+k.Marke)
	}

	for name, wert := range k.Zusatz {
		if wert != "" {
			gestellt.Header.Set(name, wert)
		}
	}
}

// Die Statuscodes, die ein Dienstpaket deuten muss.
//
// **Sie stehen hier, damit kein Dienst `net/http` einbinden muss.** Das ist
// derselbe Grund wie bei [Klient]: die Aufzählung der Pakete, die das Netz
// erreichen dürfen, soll nicht mit jedem neuen Dienst wachsen.
const (
	StatusGut           = http.StatusOK
	StatusNichtGefunden = http.StatusNotFound
	StatusVerboten      = http.StatusForbidden
	StatusGedrosselt    = http.StatusTooManyRequests

	// **Vier weitere seit A8** (#1447): Anthropic unterscheidet einen
	// fehlenden Zugang, eine offene Rechnung und einen Widerspruch, und ein
	// Blatt soll dafür nicht `net/http` einbinden müssen.
	//
	// **Und die unbrauchbare Anfrage ist dort nicht immer eine**: eine
	// selbst gesetzte Ausgabengrenze antwortet mit **400**, nicht mit 429
	// (`internal/anthropic/fehler.go`).
	StatusOhneZugang  = http.StatusUnauthorized
	StatusOhneZahlung = http.StatusPaymentRequired
	StatusWiderspruch = http.StatusConflict
	StatusUnbrauchbar = http.StatusBadRequest
)

// Kopfzeilen sind die Kopfzeilen einer Antwort.
//
// **Ein eigener Name für `http.Header`**, damit ein Dienstpaket sie lesen
// kann, ohne `net/http` einzubinden — derselbe Grund wie bei den Statuscodes
// darüber. Der Wächter in `internal/schichten/netz_test.go` zählt die Pakete,
// die das Netz erreichen dürfen, und die Liste soll nicht mit jedem neuen
// Dienst wachsen.
type Kopfzeilen = http.Header

// Antwort ist, was von einer Anfrage zurückkommt.
type Antwort struct {
	Status int
	Rumpf  []byte
	// Kopf sind die Kopfzeilen der Antwort — für den, der `Retry-After` oder
	// eine Fassungsangabe braucht.
	Kopf http.Header
}

// Hole stellt eine Anfrage — gedrosselt, mit Fristen, und ohne ein 429 zu
// übergehen.
//
// # Was sie tut, und in welcher Reihenfolge
//
//  1. **Warten**, bis die Drossel es erlaubt. Gedrosselt wird **bevor** die
//     Grenze greift (§16, Regel 4), nicht als Antwort auf ein 429.
//  2. Die Anfrage stellen, mit `User-Agent` und `Accept`.
//  3. Ein **429 als solches melden** — nicht als Fehler der Eingabe.
//  4. Den Rumpf bis zu einer Grenze lesen.
//
// # Was sie ausdrücklich nicht tut
//
// **Sie wiederholt nicht.** Ein fachliches 4xx zu wiederholen ist sinnlos, und
// ein 429 zu wiederholen ist das, was die Bedingungen verbieten: „Continuing to
// overload the API after this point may result in a temporary or permanent ban
// of your application." Wer wiederholen will, tut es beim Aufrufer — er weiß,
// ob es sich lohnt, und er hat den `Retry-After` aus [Antwort.Kopf].
func Hole(
	ctx context.Context, klient *Klient, drossel *Drossel,
	kenner Kenner, adresse string, hoechstens int64,
) (Antwort, error) {
	return holen(ctx, klient, drossel, kenner, http.MethodGet, adresse, nil, "", hoechstens)
}

// SendeJSON stellt eine **POST**-Anfrage mit einem JSON-Rumpf.
//
// # Warum es sie neben [Hole] gibt
//
// `POST /cards/collection` fragt bis zu 75 Karten auf einmal — die Kennungen
// stehen im **Rumpf**, weil sie in keine Abfrage passen
// (#1348).
//
// **Und deshalb steht sie hier und nicht beim Aufrufer.** Alles, was [Hole]
// darüber zusagt — drosseln vor der Grenze, ein 429 als solches, die Sperre
// setzen, den Rumpf begrenzen —, gilt für einen POST genauso. Ein zweiter Weg
// daneben hätte die Hälfte davon, und niemand sähe welche.
//
// # Warum POST fest steht und nicht als Wert kommt
//
// Weil `internal/schichten/netz_test.go` es so verlangt und sein Kommentar den
// Grund nennt: „Ein Eintrag, der die Liste kürzer machen soll, nicht länger."
// Ein `methode string` zwänge jeden Aufrufer, `net/http` zu kennen — und
// `internal/scryfall` steht in dieser Aufzählung ausdrücklich **nicht**. Eine
// zweite Methode kommt dazu, wenn ein Aufrufer sie braucht, nicht vorher (§2).
func SendeJSON(
	ctx context.Context, klient *Klient, drossel *Drossel, kenner Kenner,
	adresse string, rumpf []byte, hoechstens int64,
) (Antwort, error) {
	return holen(ctx, klient, drossel, kenner, http.MethodPost, adresse,
		rumpf, "application/json", hoechstens)
}

// holen ist der gemeinsame Weg von [Hole] und [SendeJSON].
//
// **Er ist nicht ausgeführt**, und das ist der Punkt: die Methode steht in
// den beiden Türen davor, nicht beim Aufrufer.
func holen(
	ctx context.Context, klient *Klient, drossel *Drossel, kenner Kenner,
	methode, adresse string, rumpf []byte, inhaltstyp string, hoechstens int64,
) (Antwort, error) {
	// **Die Erlaubnis vor der Drossel.** Wer nicht fragen darf, soll nicht
	// erst warten — und eine Wartezeit vor einem Nein wäre eine Anfrage, die
	// nur nicht abgeschickt wurde (Phase 4.8).
	if fehler := klient.darfFragen(ctx, adresse); fehler != nil {
		return Antwort{}, fehler
	}

	if drossel != nil {
		if fehler := drossel.Warte(ctx); fehler != nil {
			return Antwort{}, fehler
		}
	}

	var koerper io.Reader = http.NoBody
	if rumpf != nil {
		koerper = bytes.NewReader(rumpf)
	}

	gestellt, err := http.NewRequestWithContext(ctx, methode, adresse, koerper)
	if err != nil {
		return Antwort{}, fmt.Errorf("die Anfrage an %s bauen: %w", adresse, err)
	}

	kenner.setzeKopf(gestellt)

	if inhaltstyp != "" {
		gestellt.Header.Set("Content-Type", inhaltstyp)
	}

	antwort, fehler := klient.unten.Do(gestellt)
	if fehler != nil {
		return Antwort{}, klient.mitProxyhinweis(gestellt,
			deuteTransport(gestellt.Context(), gestellt.URL, klient.fristen, fehler))
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // eine gelesene Antwort zu schließen kann nichts mehr retten

	// **Ein `407` ist nie die Antwort des Dienstes** — nur ein Proxy sendet es
	// (RFC 9110, 15.5.8), und der Aufrufer hielte es sonst für dessen Status.
	if antwort.StatusCode == http.StatusProxyAuthRequired {
		return Antwort{Status: antwort.StatusCode, Kopf: antwort.Header},
			klient.proxyanmeldungAus(gestellt, antwort)
	}

	if antwort.StatusCode == http.StatusTooManyRequests {
		// **Das 429 sperrt die Drossel wirklich** (Zweitblick über Phase V,
		// 2026-08-24).
		//
		// Scryfall schreibt: „It is not acceptable to ignore HTTP 429
		// responses", und wiederholte 429 können zur dauerhaften Sperre führen
		// — ein Meldungstext allein ist keine Antwort darauf.
		sperreSetzen(drossel, antwort.Header)

		// **Und der Rumpf reist mit** — seit dem 2026-09-11 (A8, #1447).
		//
		// **Zwei verschiedene Dinge tragen bei Anthropic denselben
		// Status** — eine Drossel, die vorübergeht, und die Sperre am
		// Monatsdeckel, die es nicht tut —, und der Unterschied steht im
		// Rumpf ([0099](../../docs/entscheidungen/0099-was-anthropic-belegt.md)).
		//
		// **Am Fehler ändert das nichts**: er bleibt, die Drossel bleibt
		// gesperrt, und niemand bekommt etwas, das wie ein Erfolg aussieht.
		// **Ein Lesefehler ebenso wenig**: der leere Rumpf ist dann die
		// Auskunft, und das 429 ist die Meldung.
		gedrosselt, _ := lies(antwort.Body, hoechstens) //nolint:errcheck // siehe oben

		return Antwort{
				Status: antwort.StatusCode, Rumpf: gedrosselt, Kopf: antwort.Header,
			},
			fmt.Errorf("%w: %s — %s", ErrGedrosselt, ohneGeheimnis(gestellt.URL),
				wartehinweis(antwort.Header))
	}

	rumpf, lesefehler := lies(antwort.Body, hoechstens)
	if lesefehler != nil {
		return Antwort{Status: antwort.StatusCode, Kopf: antwort.Header},
			deuteRumpf(gestellt.Context(), gestellt.URL, klient.fristen.Stillstand, lesefehler)
	}

	return Antwort{Status: antwort.StatusCode, Rumpf: rumpf, Kopf: antwort.Header}, nil
}

// WarteZeit liest `Retry-After` — in Sekunden oder als Zeitpunkt.
//
// # Warum beide Formen
//
// RFC 9110 lässt beides zu, und wer nur die eine liest, hält die Kopfzeile bei
// der anderen für nicht vorhanden — und wartet gar nicht. Das ist genau das
// Übergehen, das die Bedingungen verbieten.
//
// **Der zweite Rückgabewert sagt, ob etwas dastand.** Eine Null ist nicht
// dasselbe wie „keine Angabe": bei Null darf man sofort wieder, bei keiner
// Angabe muss man selbst entscheiden.
func WarteZeit(kopf http.Header, jetzt time.Time) (time.Duration, bool) {
	roh := kopf.Get("Retry-After")
	if roh == "" {
		return 0, false
	}

	if sekunden, err := strconv.Atoi(roh); err == nil {
		if sekunden < 0 {
			return 0, true
		}

		// Sättigend (J05, #2269): wer länger warten lässt, als eine Dauer
		// fasst, meint „sehr lange" und nicht eine negative Wartezeit.
		return domain.ZeitspanneHoechstens(int64(sekunden), time.Second), true
	}

	zeitpunkt, err := http.ParseTime(roh)
	if err != nil {
		return 0, false
	}

	if dauer := zeitpunkt.Sub(jetzt); dauer > 0 {
		return dauer, true
	}

	return 0, true
}

// wartehinweis sagt, wie lange zu warten ist — für die Meldung.
func wartehinweis(kopf http.Header) string {
	dauer, gefunden := WarteZeit(kopf, time.Now())
	if !gefunden {
		return "bitte später noch einmal versuchen"
	}

	return fmt.Sprintf("bitte in %s noch einmal versuchen", dauer)
}

// lies holt einen Strom bis zur Grenze — und sagt, wenn er darüber liegt.
//
// # Warum nicht `io.LimitReader(…, hoechstens)`
//
// Weil er **kappt statt zu melden**
// (#898). Für
// den Aufrufer sähe ein zu großer Rumpf aus wie ein vollständiger, der an
// einer unglücklichen Stelle endet.
//
// **Ein Byte mehr als erlaubt** ist der ganze Kniff: kommt es, war der Strom
// länger.
func lies(strom io.Reader, hoechstens int64) ([]byte, error) {
	inhalt, err := io.ReadAll(io.LimitReader(strom, hoechstens+1))
	if err != nil {
		return nil, fmt.Errorf("die Antwort lesen: %w", err)
	}

	if int64(len(inhalt)) > hoechstens {
		return nil, domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"Die Antwort ist größer als %d Byte.", hoechstens)
	}

	return inhalt, nil
}

// SperreOhneAngabe ist die Sperrzeit, wenn `Retry-After` fehlt.
//
// **30 Sekunden**, nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`: ein 429 sperrt dort 30 Sekunden. Die
// Kopfzeile ist nicht Pflicht; ohne sie eine Sekunde zu warten hieße, die
// Grenze zu übergehen.
//
// **Sie steht hier und nicht in `internal/scryfall`**, obwohl die Zahl von dort
// kommt: dieses Paket entscheidet, was ohne Angabe geschieht, und eine
// Voreinstellung, die jeder Dienst neu setzen müsste, wäre keine.
const SperreOhneAngabe = 30 * time.Second

// sperreSetzen übersetzt ein 429 in einen Sperrzustand.
//
// **Ohne Drossel wird nichts gesperrt.** Ein Aufrufer ohne Drossel — der
// Bilddownload etwa, dessen Adressen kein Budget haben — hat keinen Ort für
// die Sperre. Ihm eine unterzuschieben hieße, ihn stumm zu machen.
func sperreSetzen(drossel *Drossel, kopf http.Header) {
	if drossel == nil {
		return
	}

	// **Die Uhr der Drossel entscheidet**, nicht `time.Now()` — deshalb
	// [Drossel.SperreFuer] mit einer Dauer. `Retry-After` in seiner
	// Zeitpunktform braucht trotzdem eine Gegenwart; genommen wird die der
	// Drossel.
	dauer, gefunden := WarteZeit(kopf, drossel.Jetzt())
	if !gefunden {
		dauer = SperreOhneAngabe
	}

	drossel.SperreFuer(dauer)
}
