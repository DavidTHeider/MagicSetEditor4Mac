// **Die Drossel endet nicht am Prozess** — Befund Z7-03 des
// Audits für 0.3.0 (Entwicklungsarchiv), Phase 3.2
// (#1163).
//
// # Der Befund
//
// Innerhalb eines Prozesses teilen sich alle Zugänge ein [Drosselwerk] — das
// steht und ist geprüft. Fenster, MCP-Server und Dienst sind aber **drei**
// Prozesse, und jeder hatte sein eigenes. Wer alle drei laufen hat, schickt
// das Dreifache des Budgets, das `DRITTANBIETER.md` zusagt.
//
// # Die Absprache ist eine Nachbardatei
//
// Sie liegt beim Zwischenspeicher der Gegenstelle und trägt **zwei**
// Zeitpunkte:
//
//   - wann zuletzt jemand losgeschickt hat — daraus wird der Abstand,
//   - bis wann **niemand** senden darf — das kommt aus einem 429.
//
// Wer schicken will, nimmt sie ausschließlich, liest beide, wartet und
// schreibt den neuen Sendezeitpunkt.
//
// **Die Sperre aus einem 429 gehört dazu** und war es zuerst nicht: Scryfall
// schreibt „It is not acceptable to ignore HTTP 429 responses", und wiederholte
// 429 können zur dauerhaften Sperre führen. Ein zweiter Prozess, der während
// `Retry-After` weitersendet, ignoriert sie — auch wenn der erste sie brav
// einhält (Zweitblick über Phase 3.2).
//
// **Gehalten wird über das Warten.** Das ist der Punkt: gäbe sie die Sperre
// vor dem Warten zurück, läsen zwei Prozesse denselben Zeitpunkt und
// schickten gemeinsam los. Ein Abstand ist kurz — 100 ms bei Scryfall —, und
// länger als der Abstand wartet niemand.
//
// # Was sie nicht kann
//
//   - **Sie ist beratend** (siehe `internal/dateisperre`). Ein anderes
//     Programm auf demselben Rechner hält sich nicht daran.
//   - **Sie zählt keine Anfragen.** Sie hält einen **Abstand** und eine
//     **Sperre**; wie viele in einer Stunde hinausgehen, sagt sie nicht.
//   - **Ohne sie geht es weiter.** Wer die Nachbardatei nicht anlegen kann,
//     drosselt wie vorher — je Prozess. Das ist schlechter als mit, und eine
//     Anwendung, die deswegen keine Karte mehr sucht, wäre schlechter als
//     beides.

package netz

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// Prozessgrenzfrist ist, wie lange auf die Absprache gewartet wird.
//
// **Großzügig gegenüber dem Abstand**: ein Halter gibt sie nach höchstens
// einem Abstand zurück. Bei 100 ms und einer Frist von einer Minute stehen
// damit sechshundert Wartende in der Schlange, bevor jemand aufgibt — mehr
// Prozesse hat niemand.
//
// **Eine Minute und nicht fünf Sekunden.** Der erste Anlauf nahm fünf, und
// der Zweitblick hat angestrichen, warum das zu knapp ist: bei hoher
// Gleichzeitigkeit gäbe gerade dann jemand auf, wenn die gemeinsame Grenze am
// meisten zählt — **fail-open genau unter Last**.
var Prozessgrenzfrist = time.Minute

// Prozessgrenztakt ist der Abstand zwischen zwei Versuchen.
var Prozessgrenztakt = 5 * time.Millisecond

// Prozessgrenze ist die Absprache über einen Abstand, prozessübergreifend.
type Prozessgrenze struct {
	// ort gibt die Nachbardatei — **bei jedem Aufruf neu**.
	//
	// **Eine Funktion und keine Zeichenkette.** Der Ort hängt am
	// Benutzerverzeichnis, und eine Drossel wird **einmal** gebaut und dann
	// wiederverwendet: eine festgehaltene Zeichenkette bliebe für immer die
	// des ersten Aufrufs. Im Betrieb fällt das nicht auf — es gibt ein
	// Zuhause. In den Läufen schrieb die zweite Drossel in das Zuhause der
	// ersten, und das war das **echte** dieses Rechners (Nachlauf über
	// Phase 3.2, #1163).
	ort     func() string
	abstand time.Duration
	jetzt   func() time.Time
	schlafe func(time.Duration)
}

// NeueProzessgrenze baut die Absprache über eine Nachbardatei.
//
// **Uhr und Schlaf werden eingesetzt** (§16), aus demselben Grund wie bei
// [Drossel]: ein Lauf, der auf Uhren wartet, flackert unter Last (§3).
func NeueProzessgrenze(
	pfad string, abstand time.Duration,
	jetzt func() time.Time, schlafe func(time.Duration),
) *Prozessgrenze {
	return NeueProzessgrenzeAn(func() string { return pfad }, abstand, jetzt, schlafe)
}

// NeueProzessgrenzeAn baut sie über eine Naht auf den Ort.
//
// **Für Aufrufer, deren Ort sich ändern kann** — warum, steht bei
// [Prozessgrenze.ort].
func NeueProzessgrenzeAn(
	ort func() string, abstand time.Duration,
	jetzt func() time.Time, schlafe func(time.Duration),
) *Prozessgrenze {
	if jetzt == nil {
		jetzt = time.Now
	}

	return &Prozessgrenze{ort: ort, abstand: abstand, jetzt: jetzt, schlafe: schlafe}
}

// pfad gibt die Nachbardatei — leer, wenn es keine gibt.
func (p *Prozessgrenze) pfad() string {
	if p == nil || p.ort == nil {
		return ""
	}

	return p.ort()
}

// Warte hält den Abstand ein — über alle Prozesse, die mitmachen.
func (p *Prozessgrenze) Warte(ctx context.Context) error {
	// **Höchstens so viele Runden.** Jede folgt auf eine abgewartete Sperre;
	// wer danach wieder eine vorfindet, hat sie von einem anderen Prozess
	// bekommen. Ohne Schranke drehte der Weg bei einer stehenden Uhr für
	// immer — im Betrieb kommt nie eine zweite Runde vor.
	const hoechstensRunden = 8

	for range hoechstensRunden {
		gesperrtBis, fehler := p.eineRunde(ctx)
		if fehler != nil || gesperrtBis.IsZero() {
			return fehler
		}

		// **Ohne Griff abgewartet.** Eine Sperre kann Minuten dauern; sie mit
		// der Absprache in der Hand abzuwarten hieße, alle anderen Prozesse
		// so lange auszusperren — und die gäben nach [Prozessgrenzfrist] auf
		// und sendeten **ohne** Absprache.
		if wartefehler := p.warte(ctx, gesperrtBis.Sub(p.jetzt())); wartefehler != nil {
			return wartefehler
		}
	}

	return nil
}

// eineRunde nimmt die Absprache und hält den Abstand.
//
// Gibt einen Zeitpunkt zurück, wenn statt des Abstands eine **Sperre** gilt —
// dann hat der Aufrufer sie ohne Griff abzuwarten.
func (p *Prozessgrenze) eineRunde(ctx context.Context) (gesperrtBis time.Time, fehler error) {
	pfad := p.pfad()
	if pfad == "" {
		return time.Time{}, nil
	}

	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		// **Ohne Absprache geht es weiter.** Der Aufrufer hat seine eigene
		// Drossel; sie ist dann so gut wie vor Phase 3.2.
		return time.Time{}, nil
	}

	defer func() { _ = griff.Schliessen() }() //nolint:errcheck // die Sperre fällt mit dem Griff

	if sperrfehler := griff.AlleinBis(ctx, Prozessgrenzfrist, Prozessgrenztakt); sperrfehler != nil {
		// **Ein beendeter Kontext ist etwas anderes als eine abgelaufene
		// Frist.** Das eine heißt „niemand will die Antwort mehr", das andere
		// „die Absprache ist nicht zu haben" — und nur beim zweiten geht es
		// ohne sie weiter (Zweitblick über Phase 3.2).
		return time.Time{}, ctx.Err()
	}

	return p.abstandHalten(ctx, griff)
}

// abstandHalten liest beide Zeitpunkte, wartet und schreibt den neuen.
//
// **Die Sperre wird ohne Griff abgewartet** und der Abstand mit.
//
// Der Unterschied ist die Dauer: ein Abstand ist kurz — 100 ms bei Scryfall —,
// eine Sperre aus einem 429 kann Minuten dauern. Wer sie **mit** dem Griff
// abwartete, hielte die Absprache so lange; die anderen Prozesse gäben nach
// [Prozessgrenzfrist] auf und sendeten **ohne** sie. Genau dann wäre die
// gemeinsame Grenze weg, wenn sie am meisten zählt (Nachlauf über Phase 3.2,
// #1163).
func (p *Prozessgrenze) abstandHalten(
	ctx context.Context, griff *dateisperre.Griff,
) (time.Time, error) {
	zuletzt, gesperrtBis := p.lesen(griff)
	jetzt := p.jetzt()

	// **Eine Sperre geht vor dem Abstand.** Sie ist die Antwort der
	// Gegenstelle und damit die strengere Auskunft — dieselbe Reihenfolge wie
	// in [Drossel.Warte], und aus demselben Grund.
	//
	// **Und sie wird nicht hier abgewartet**, sondern beim Aufrufer, ohne
	// Griff. Warum, steht bei [Prozessgrenze.Warte].
	if gesperrtBis.After(jetzt) {
		return gesperrtBis, nil
	}

	// **Ein Zeitpunkt aus der Zukunft ist keiner.** Er entsteht bei einer
	// verstellten Uhr oder einer veränderten Datei — und ohne diese Zeile
	// wartete der nächste Aufruf so lange, wie die Zukunft entfernt ist.
	if zuletzt.After(jetzt) {
		zuletzt = jetzt
	}

	if verstrichen := jetzt.Sub(zuletzt); verstrichen < p.abstand {
		if fehler := p.warte(ctx, p.abstand-verstrichen); fehler != nil {
			return time.Time{}, fehler
		}

		jetzt = p.jetzt()
	}

	p.schreiben(griff, jetzt, gesperrtBis)

	return time.Time{}, nil
}

// Sperren trägt eine Sperre aus einem 429 in die Absprache ein.
//
// **Damit sie alle Prozesse trifft** und nicht nur den, der die Antwort bekam.
// Ohne sie sendet ein zweiter Prozess während `Retry-After` weiter — und
// ignoriert das 429 damit gemeinsam, auch wenn der erste es einhält.
func (p *Prozessgrenze) Sperren(bis time.Time) {
	pfad := p.pfad()
	if pfad == "" {
		return
	}

	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		return
	}

	defer func() { _ = griff.Schliessen() }() //nolint:errcheck // die Sperre fällt mit dem Griff

	// **Gewartet und nicht einmal versucht.** Der erste Anlauf nahm die
	// Absprache ohne Warten — war sie gerade belegt, wurde die Sperre **gar
	// nicht** eingetragen, und die anderen Prozesse sandten weiter. Ein 429
	// ist die Antwort der Gegenstelle; sie zu verlieren, weil ein anderer
	// gerade schreibt, ist der teuerste Fall (Nachlauf über Phase 3.2).
	if fehler := griff.AlleinBis(context.Background(),
		Prozessgrenzfrist, Prozessgrenztakt); fehler != nil {
		// **Und wenn sie auch dann nicht zu haben ist**, bleibt die Sperre im
		// eigenen Prozess. Sie dort zu halten ist besser als sie zu verlieren.
		return
	}

	zuletzt, vorherige := p.lesen(griff)

	// **Die längere gilt.** Zwei 429 kurz hintereinander sagen nicht, dass
	// die zweite Sperre die erste ablöst — sie sagen, dass die Gegenstelle
	// mindestens so lange nichts will.
	if vorherige.After(bis) {
		bis = vorherige
	}

	p.schreiben(griff, zuletzt, bis)
}

// lesen holt beide vermerkten Zeitpunkte.
//
// **Eine unlesbare Datei ist zweimal die Nullzeit** — „lange her" und „keine
// Sperre". Sie entsteht beim ersten Mal, und ein Fehler darauf hielte den
// ersten Aufruf jedes frischen Rechners auf.
func (p *Prozessgrenze) lesen(griff *dateisperre.Griff) (zuletzt, gesperrtBis time.Time) {
	roh := make([]byte, 16)

	// **Genau acht Bytes, sonst gar keine.** Ein kürzerer Inhalt füllt den
	// Rest mit dem, was zufällig im Puffer stand — und daraus wird ein
	// Zeitpunkt, den niemand geschrieben hat. Angestrichen beim ersten Lauf
	// gegen eine Datei mit sieben Zeichen darin: er wartete.
	gelesen, err := griff.LiesVonVorn(roh)
	if (err != nil && !errors.Is(err, io.EOF)) || gelesen != len(roh) {
		return time.Time{}, time.Time{}
	}

	return alsZeit(roh[:8]), alsZeit(roh[8:])
}

// alsZeit macht aus acht Bytes einen Zeitpunkt — die Nullzeit bei Unsinn.
func alsZeit(roh []byte) time.Time {
	nanos := int64(binary.BigEndian.Uint64(roh)) // #nosec G115 -- der Wert kam als int64 heraus
	if nanos <= 0 {
		return time.Time{}
	}

	return time.Unix(0, nanos)
}

// schreiben vermerkt den Zeitpunkt für den nächsten.
//
// **Ein Fehler bleibt ohne Folge.** Er hieße, dass der nächste Aufruf die
// Nullzeit liest und nicht wartet — schlechter, aber kein Grund, eine Anfrage
// scheitern zu lassen, die längst hinausgegangen ist.
func (p *Prozessgrenze) schreiben(griff *dateisperre.Griff, wann, gesperrtBis time.Time) {
	roh := make([]byte, 16)
	binary.BigEndian.PutUint64(roh, alsNanos(wann))
	binary.BigEndian.PutUint64(roh[8:], alsNanos(gesperrtBis))

	_ = griff.SchreibVonVorn(roh) //nolint:errcheck // siehe darüber
}

// alsNanos macht aus einem Zeitpunkt acht Bytes — null bei der Nullzeit.
func alsNanos(wann time.Time) uint64 {
	if wann.IsZero() {
		return 0
	}

	return uint64(wann.UnixNano()) // #nosec G115 -- Zeit nach 1970
}

// warte hält an — oder gibt auf, wenn der Kontext endet.
func (p *Prozessgrenze) warte(ctx context.Context, dauer time.Duration) error {
	if p.schlafe != nil {
		p.schlafe(dauer)

		return nil
	}

	uhr := time.NewTimer(dauer)
	defer uhr.Stop()

	select {
	case <-uhr.C:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
