package mtgapi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Adresse ist die Grundadresse des Dienstes.
//
// **Nachgelesen am 2026-09-02** unter `https://docs.magicthegathering.io/`:
// „https://api.magicthegathering.io/<version>/<resource>", Fassung `v1`.
const Adresse = "https://api.magicthegathering.io/v1"

// Seitengroesse ist, wie viele Karten eine Seite trägt.
//
// **Vorgabe und Höchstwert sind beide 100** (nachgelesen am 2026-09-02). Wer
// mehr verlangte, bekäme nicht mehr — und behauptete es in seinem eigenen
// Code.
const Seitengroesse = 100

// Abstand ist, wie oft gefragt werden darf.
//
// # Die Dokumentation sagt 5.000, der Dienst sagt 1.000
//
// **Gemessen am 2026-09-02** gegen `api.magicthegathering.io`:
//
//	ratelimit-limit: 1000
//	ratelimit-remaining: 997
//
// `https://docs.magicthegathering.io/` nennt am selben Tag „5000 requests per
// hour". **Der Dienst hat recht und die Seite nicht** — und genau dafür gibt
// es den Lauf hinter `netzlauf`: er ist rot geworden, als hier noch 5.000
// stand.
//
// **Also fünffach weniger, als die Doku verspricht.** Wer der Seite geglaubt
// hätte, wäre mit 3.600 Anfragen je Stunde gegen eine Grenze von 1.000
// gelaufen — und hätte es erst am 403 gemerkt.
//
// # Warum fünf Sekunden
//
// 1.000 je Stunde ist eine je 3,6 Sekunden. Gedrosselt wird **unter** der
// genannten Zahl (§16, Regel 4); fünf Sekunden lassen 720 zu, mit 28 %
// Abstand — dasselbe Verhältnis, das vorher gegen die falsche Zahl stand.
//
// **Der Abstand ist kein Konto.** Er hält den Dauerstrom unter der Grenze;
// wer nach einer Pause stößt, könnte theoretisch mehr. Dafür gibt es den
// Restwert aus der Antwort ([Stand]), und wer ihn sinken sieht, weiß es.
const Abstand = 5 * time.Second

// hoechstens ist, wie groß eine Antwort werden darf.
//
// **Hundert Karten mit Regeltext und Legalitäten.** Eine Antwort, die
// darüber liegt, ist keine Kartenliste mehr — und ein Leser ohne Grenze ist
// eine offene Tür.
const hoechstens = 8 << 20

// Stand ist, was der Dienst über die Grenze sagt.
//
// **`Bekannt` ist kein Luxus.** „Null übrig" heißt „gleich ist Schluss";
// „unbekannt" heißt „der Dienst hat nichts gesagt". Wer beides zusammenwürfe,
// drosselte grundlos.
type Stand struct {
	Bekannt bool
	Grenze  int
	Uebrig  int
}

// Klient fragt magicthegathering.io.
//
// **Kein `*http.Client`**, sondern `netz.Klient`: die Aufzählung der Pakete,
// die das Netz erreichen dürfen, soll nicht mit jedem neuen Dienst wachsen
// (`internal/schichten/netz_test.go`).
type Klient struct {
	grundadresse string
	klient       *netz.Klient
	drossel      *netz.Drossel
	kenner       netz.Kenner
}

// NeuerKlient baut einen Klienten.
func NeuerKlient(grundadresse string, klient *netz.Klient) *Klient {
	return &Klient{
		grundadresse: strings.TrimSuffix(grundadresse, "/"),
		klient:       klient,
		drossel:      netz.NeueDrossel(Abstand, time.Now, time.Sleep),
		kenner: netz.Kenner{
			// **Die Anwendung nennt sich** (§16, Regel 2). Bei
			// magicthegathering.io ist es keine Pflicht — es ist die
			// Mindesthöflichkeit, und ein Dienst, der eine Last sieht, soll
			// wissen, woher sie kommt.
			Name:    "MagicSetEditor4Mac",
			Annahme: "application/json",
		},
	}
}

// MitDrossel setzt die Drossel — **für Tests**, die nicht sekundenweise warten.
func (k *Klient) MitDrossel(drossel *netz.Drossel) *Klient {
	k.drossel = drossel

	return k
}

// Karten fragt `/cards` mit einer fertigen Abfrage.
//
// **Die Abfrage kommt vom Aufrufer.** Welche Felder es gibt, steht beim
// Dienst; sie hier nachzubauen hieße, eine zweite Aufzählung zu führen, die
// mit der ersten auseinanderläuft.
func (k *Klient) Karten(ctx context.Context, abfrage string) (
	[]Karte, Stand, error,
) {
	adresse := fmt.Sprintf("%s/cards?%s&pageSize=%d",
		k.grundadresse, abfrage, Seitengroesse)

	antwort, fehler := netz.Hole(ctx, k.klient, k.drossel, k.kenner, adresse, hoechstens)
	if fehler != nil {
		return nil, Stand{}, fmt.Errorf("magicthegathering.io fragen: %w", fehler)
	}

	stand := standAus(antwort)

	if err := pruefeStatus(antwort, stand); err != nil {
		return nil, stand, err
	}

	karten, _, lesefehler := LiesKarten(antwort.Rumpf)
	if lesefehler != nil {
		return nil, stand, lesefehler
	}

	return karten, stand, nil
}

// pruefeStatus deutet den Statuscode.
//
// # Warum 403 hier zweimal vorkommt
//
// **Nachgelesen am 2026-09-02**: über der Grenze antwortet der Dienst mit
// **403 Forbidden** und „Rate Limit Exceeded" — nicht mit 429. Derselbe Code
// heißt sonst „du darfst hier nicht", und die beiden Auskünfte haben zwei
// Abhilfen: warten oder aufhören (§8).
//
// **Unterschieden wird am Rumpf**, weil der Dienst nichts anderes anbietet.
//
// # Und warum hier kein Fall für 429 steht
//
// **Weil `netz.Hole` ihn schon beantwortet** — mit `ErrGedrosselt` und, was
// hier gefehlt hätte, dem **Sperren der Drossel**. Ein zweiter Fall daneben
// wäre ein Zweig, den kein Aufruf betreten kann (§3), und er machte eine
// schwächere Zusage als der, den es gibt.
func pruefeStatus(antwort netz.Antwort, stand Stand) error {
	switch antwort.Status {
	case netz.StatusGut:
		return nil

	case netz.StatusNichtGefunden:
		return fmt.Errorf(
			"%w: magicthegathering.io kennt dazu nichts", domain.ErrNotFound)

	case netz.StatusVerboten:
		if !strings.Contains(string(antwort.Rumpf), "Rate Limit Exceeded") {
			return fmt.Errorf(
				"%w: magicthegathering.io weist die Anfrage ab: %s",
				domain.ErrValidation, kurz(antwort.Rumpf))
		}

		return fmt.Errorf(
			"%w: die Grenze von magicthegathering.io ist erreicht — %s",
			domain.ErrGedrosselt, standtext(stand))
	}

	return fmt.Errorf(
		"%w: magicthegathering.io antwortet mit %d",
		domain.ErrValidation, antwort.Status)
}

// standAus liest die beiden Kopfzeilen.
//
// **Nachgelesen am 2026-09-02**: `Ratelimit-Limit` und `Ratelimit-Remaining`.
func standAus(antwort netz.Antwort) Stand {
	grenze, grenzeGut := zahl(antwort.Kopf.Get("Ratelimit-Limit"))
	uebrig, uebrigGut := zahl(antwort.Kopf.Get("Ratelimit-Remaining"))

	// **Beide oder keiner.** Ein halber Stand ist keiner: „5.000 Grenze, und
	// wie viele übrig sind, weiß ich nicht" hilft niemandem.
	if !grenzeGut || !uebrigGut {
		return Stand{}
	}

	return Stand{Bekannt: true, Grenze: grenze, Uebrig: uebrig}
}

// zahl liest eine Kopfzeile als Zahl.
func zahl(roh string) (int, bool) {
	wert, err := strconv.Atoi(strings.TrimSpace(roh))

	return wert, err == nil
}

// standtext macht den Stand lesbar — auch wenn es keinen gibt.
func standtext(stand Stand) string {
	if !stand.Bekannt {
		return "der Dienst nennt keinen Stand"
	}

	return fmt.Sprintf("%d von %d Anfragen übrig", stand.Uebrig, stand.Grenze)
}

// kurz gibt den Anfang eines Rumpfes — für eine Meldung.
//
// **Nicht den ganzen.** Eine Fehlermeldung mit acht Megabyte darin ist keine.
func kurz(rumpf []byte) string {
	const hoechstens = 200

	text := strings.TrimSpace(string(rumpf))
	if len(text) > hoechstens {
		return text[:hoechstens] + " …"
	}

	return text
}

// Abfrage baut eine Abfrage aus Feld und Wert.
//
// **Sie steht hier, damit niemand von Hand kodiert.** Ein Kartenname mit
// Leerzeichen oder Apostroph gehört kodiert, und wer es vergisst, bekommt
// eine Antwort auf eine andere Frage.
func Abfrage(feld, wert string) string {
	return feld + "=" + url.QueryEscape(wert)
}
