// Package netz ist der Transport zu fremden Diensten
// (#1002).
//
// # Was hier liegt
//
// Die **erste** der drei Schichten aus AGENTS.md §16: HTTP, Fristen,
// Wiederholung, Drossel, Zwischenspeicher. Das fremde DTO und die Zuordnung
// auf `domain` liegen im Paket des jeweiligen Dienstes — hier weiß niemand,
// was eine Karte ist.
//
// # Warum es dieses Paket gibt
//
// Weil es sonst zwei Kopien gäbe. `internal/bezug` hat den Transport nach
// Phase R gelöst — drei Fristen, ein benannter `User-Agent`, ein 429, das als
// solches gemeldet wird —, und ein zweiter Dienst hätte ihn abgeschrieben.
// Was daraus wird, steht in `msepaket/paket.go`: „Zwei Kopien dieser
// Fallunterscheidung liefen an dem Tag auseinander, an dem eine von ihnen eine
// dritte Form lernte."
//
// **Ein Blatt** (§2): es kennt `domain` und sonst nichts aus dem Projekt.
package netz

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// Drossel hält den Mindestabstand zwischen zwei ausgehenden Anfragen.
//
// # Warum sie wartet und nicht abweist
//
// Sie ist das Gegenstück zu der in `internal/rest`, und sie tut das Gegenteil.
// Dort kommt eine Anfrage **herein**: wer zu viele schickt, bekommt ein 429,
// und abweisen ist die Antwort. Hier geht eine Anfrage **hinaus**, und der
// Aufrufer ist dieses Programm — es soll langsamer werden, nicht scheitern.
// Ein Abweisen machte aus einer Grenze der Gegenstelle einen Fehler bei uns.
//
// # Warum ein Mindestabstand und kein Zeitfenster
//
// Weil die Gegenstelle es so schreibt. Scryfall nennt „2/second (500ms)" und
// „10/second (100ms)" — nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api/rate-limits`. Ein Zeitfenster ließe zehn
// Anfragen in einer Zehntelsekunde durch und danach eine Sekunde nichts; das
// ist im Mittel dasselbe und im Augenblick das Doppelte.
type Drossel struct {
	abstand time.Duration
	jetzt   func() time.Time
	schlafe func(time.Duration)

	schranke sync.Mutex
	zuletzt  time.Time

	// gesperrtBis ist der Zeitpunkt, bis zu dem nichts hinausgeht.
	//
	// **Er kommt von einem 429.** Die Gegenstelle hat gesagt, dass sie nicht
	// mehr will; ein Mindestabstand ist dann zu wenig. Scryfall schreibt: „It
	// is not acceptable to ignore HTTP 429 responses", und wiederholte 429
	// können zur dauerhaften Sperre führen.
	//
	// **Die Nullzeit heißt: keine Sperre.**
	gesperrtBis time.Time

	// grenze ist die prozessübergreifende Absprache — nil, wenn es keine
	// gibt (#1163, Befund Z7-03).
	grenze *Prozessgrenze

	// werk ist das Werk, aus dem diese Drossel stammt — nil, wenn sie
	// einzeln gebaut wurde.
	//
	// **Dafür, dass eine Sperre alle Klassen trifft** (#1037). Die Abstände
	// sind je Klasse verschieden und bleiben es; die **Sperre** gilt für die
	// Gegenstelle als Ganzes, und `DRITTANBIETER.md` sagt genau das zu:
	// „Bis dahin geht aus diesem Prozess keine Anfrage hinaus — auch nicht
	// aus einem anderen Zugang."
	werk *Drosselwerk
}

// NeueDrossel baut eine Drossel mit einem Mindestabstand.
//
// **Uhr und Schlaf werden eingesetzt** (§16). Sonst wartete jeder Test
// wirklich, und ein Lauf, der auf Uhren wartet, flackert unter Last (§3).
// `schlafe` darf `nil` sein — dann wird echt gewartet.
func NeueDrossel(abstand time.Duration, jetzt func() time.Time, schlafe func(time.Duration)) *Drossel {
	return &Drossel{abstand: abstand, jetzt: jetzt, schlafe: schlafe}
}

// Sperre hält alles an, bis der Zeitpunkt vorbei ist.
//
// # Warum eine Sperre und nicht ein größerer Abstand
//
// Weil sie eine **andere** Aussage ist. Der Abstand ist unsere Vorsicht; die
// Sperre ist die Antwort der Gegenstelle. Sie in den Abstand zu rechnen hieße,
// künftige Anfragen dauerhaft zu bremsen, weil eine einzige zu früh kam.
//
// # Die längere gilt
//
// Zwei Anfragen können gleichzeitig ein 429 bekommen, mit verschiedenen
// `Retry-After`. Die kürzere zuletzt zu nehmen hieße, die längere zu
// unterlaufen — und genau das würde die Gegenstelle als Ignorieren lesen.
func (d *Drossel) Sperre(bis time.Time) {
	d.schranke.Lock()
	defer d.schranke.Unlock()

	if bis.After(d.gesperrtBis) {
		d.gesperrtBis = bis
	}
}

// Gesperrt sagt, ob gerade nichts hinausgehen darf.
//
// **Für Tests und für einen Aufrufer, der es wissen will** — der Wartepfad
// prüft es ohnehin selbst.
func (d *Drossel) Gesperrt() bool {
	d.schranke.Lock()
	defer d.schranke.Unlock()

	return d.gesperrtBis.After(d.jetzt())
}

// Jetzt gibt die Gegenwart nach der Uhr dieser Drossel.
//
// **Nötig, weil `Retry-After` auch ein Zeitpunkt sein darf.** Um daraus eine
// Dauer zu machen, braucht der Aufrufer dieselbe Gegenwart, mit der die Drossel
// rechnet — sonst rechnet er gegen eine andere Uhr.
func (d *Drossel) Jetzt() time.Time { return d.jetzt() }

// SperreFuer hält alles für eine Dauer an — gerechnet auf **ihrer** Uhr.
//
// # Warum eine Dauer und kein Zeitpunkt
//
// **Weil die Uhr der Drossel gilt und nicht die des Aufrufers.** Der erste
// Anlauf rechnete den Zeitpunkt mit `time.Now()` aus, und ein Test mit
// gestellter Uhr sah eine Sperre in der Vergangenheit — sie wirkte nicht, und
// der Lauf war rot. `Retry-After` nennt ohnehin eine Dauer.
// **Und sie gilt für das ganze Werk** (#1037). Ein 429 ist eine Aussage der
// Gegenstelle über **sich**, nicht über eine Endpunktklasse; wer nur die eine
// sperrte, ließe die andere weiterlaufen — und genau das las die Gegenstelle
// als Ignorieren. Angestrichen vom Nachlauf über Phase V: die Zusage stand in
// `DRITTANBIETER.md`, der Code hielt sie nicht.
//
// **Eine Drossel ohne Werk sperrt nur sich selbst.** Sie hat keinen Ort für
// mehr, und ihr einen unterzuschieben hieße, eine Zusage zu erfinden.
func (d *Drossel) SperreFuer(dauer time.Duration) {
	d.SperreGegenstelle(d.jetzt().Add(dauer))
}

// Warte hält den Abstand ein — und gibt auf, wenn der Kontext endet.
//
// # Warum der Abbruch hier hineingehört
//
// Eine Drossel, die nach einem `⌃C` noch eine halbe Sekunde stillsteht, hält
// den ganzen Vorgang auf. Der Aufrufer hat gesagt, dass er nicht mehr will.
//
// # Was sie nicht tut
//
// **Sie zählt nicht mit, wie viele warten.** Bei zwanzig gleichzeitigen
// Anfragen reihen sich alle hintereinander ein, und die letzte wartet zehn
// Sekunden — das ist richtig so, aber es heißt, dass ein Aufrufer mit einer
// Frist sie hier verbraucht. Wer das nicht will, gibt eine kürzere Frist mit.
func (d *Drossel) Warte(ctx context.Context) error {
	d.schranke.Lock()
	defer d.schranke.Unlock()

	jetzt := d.jetzt()

	// **Eine Sperre geht vor dem Abstand.** Sie ist die Antwort der
	// Gegenstelle und damit die strengere Auskunft.
	if d.gesperrtBis.After(jetzt) {
		if fehler := d.warte(ctx, d.gesperrtBis.Sub(jetzt)); fehler != nil {
			return fehler
		}

		jetzt = d.jetzt()
	}

	// **Der erste Aufruf wartet nicht.** `zuletzt` ist dann die Nullzeit, und
	// der Abstand ist längst um.
	if verstrichen := jetzt.Sub(d.zuletzt); verstrichen < d.abstand {
		if fehler := d.warte(ctx, d.abstand-verstrichen); fehler != nil {
			return fehler
		}

		jetzt = d.jetzt()
	}

	d.zuletzt = jetzt

	// **Und danach die Absprache mit den anderen Prozessen** (#1163, Befund
	// Z7-03).
	//
	// **Danach und nicht davor**: die Schranke dieser Drossel hält schon; wer
	// die Nachbardatei zuerst nähme, hielte sie, während er auf die eigene
	// wartet — und blockierte damit die anderen Prozesse für eine Zeit, die
	// nichts mit dem Abstand zu tun hat.
	//
	// **Beide zusammen ergeben den größeren Abstand**, und das ist die
	// richtige Antwort: jede für sich hält, was sie zusagt.
	return d.grenze.Warte(ctx)
}

// warte hält an — oder gibt auf, wenn der Kontext endet.
func (d *Drossel) warte(ctx context.Context, dauer time.Duration) error {
	if d.schlafe != nil {
		d.schlafe(dauer)

		return nil
	}

	uhr := time.NewTimer(dauer)
	defer uhr.Stop()

	select {
	case <-uhr.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("beim Warten auf die Drossel: %w", ctx.Err())
	}
}

// Drosselwerk gibt für denselben Schlüssel dieselbe Drossel.
//
// # Warum es das gibt
//
// **Das Budget gilt für Host und IP, nicht für eine Instanz in diesem
// Programm.** Vor dem 2026-08-24 baute das Fenster **vier** Klienten, jeden
// mit eigener Drossel; MCP und REST bauten ihre Naht **je Aufruf** neu. Weil
// der erste Aufruf einer Drossel absichtlich nicht wartet, sah eine solche
// Drossel praktisch nie einen zweiten — die Grenze war wirkungslos.
//
// Angestrichen vom Zweitblick über Phase V
// (#888).
//
// # Was es nicht kann
//
// **Über Prozessgrenzen hinweg gilt es nicht.** Fenster, `mse-mcp` und
// `mse-rest` sind eigene Prozesse; wer alle drei gleichzeitig fährt, hat
// dreimal das Budget. Ein gemeinsamer Zustand dafür hieße eine Datei oder ein
// Schloss auf der Platte — das ist eine eigene Entscheidung und keine, die
// hier nebenbei fällt.
//
// **Der Schlüssel wird vom Aufrufer gewählt.** Dieses Paket weiß nicht, welche
// Wege zu welcher Klasse gehören; das weiß der Dienst.
type Drosselwerk struct {
	jetzt   func() time.Time
	schlafe func(time.Duration)

	// absprache gibt das Verzeichnis der prozessübergreifenden Absprache —
	// leer oder nil, wenn es keine gibt (#1163, Befund Z7-03).
	//
	// **Eine Funktion und keine Zeichenkette.** Der Ort entsteht aus dem
	// Benutzerverzeichnis, und ein Paketwert wertete ihn beim **Laden** des
	// Pakets aus — also vor jedem `t.Setenv("HOME", …)`. Die Tests legten
	// damit Nachbardateien im **echten** Zuhause an und blieben grün: genau
	// der Fall, den §3 unter „ein Test, der nach draußen greift" führt.
	// Angestrichen vom Nachlauf über Phase 3.2.
	absprache func() string

	schranke sync.Mutex
	drosseln map[string]*Drossel
}

// MitAbsprache legt die prozessübergreifende Absprache in ein Verzeichnis.
//
// # Warum das Werk und nicht die einzelne Drossel
//
// Weil das Budget der Gegenstelle gehört und nicht einer Endpunktklasse. Wer
// es setzt, setzt es für alle Klassen dieses Werks — und je Klasse entsteht
// eine eigene Nachbardatei, denn die **Abstände** sind verschieden.
//
// **Nur vor dem ersten `Fuer`.** Danach stünden Drosseln mit und ohne
// Absprache nebeneinander, und welche man bekommt, hinge an der Reihenfolge.
func (w *Drosselwerk) MitAbsprache(verzeichnis func() string) *Drosselwerk {
	w.absprache = verzeichnis

	return w
}

// NeuesDrosselwerk baut ein Werk.
//
// **Uhr und Schlaf werden eingesetzt** (§16) — dieselbe Begründung wie bei
// [NeueDrossel].
func NeuesDrosselwerk(jetzt func() time.Time, schlafe func(time.Duration)) *Drosselwerk {
	return &Drosselwerk{jetzt: jetzt, schlafe: schlafe, drosseln: map[string]*Drossel{}}
}

// Abspracheort gibt das Verzeichnis der prozessübergreifenden Absprache.
func (w *Drosselwerk) Abspracheort() string {
	if w.absprache == nil {
		return ""
	}

	return w.absprache()
}

// Fuer gibt die Drossel eines Schlüssels — dieselbe bei jedem Aufruf.
//
// **Der Abstand des ersten Aufrufs gilt.** Ein zweiter mit anderem Abstand
// bekommt die vorhandene Drossel; sie zu ändern hieße, dass der letzte
// Aufrufer über alle bestimmt.
func (w *Drosselwerk) Fuer(schluessel string, abstand time.Duration) *Drossel {
	w.schranke.Lock()
	defer w.schranke.Unlock()

	if vorhanden, schon := w.drosseln[schluessel]; schon {
		return vorhanden
	}

	neue := NeueDrossel(abstand, w.jetzt, w.schlafe)
	neue.werk = w
	neue.grenze = w.grenzeFuer(schluessel, abstand)
	w.drosseln[schluessel] = neue

	return neue
}

// SperreAlle sperrt jede Drossel dieses Werks.
//
// # Warum das ganze Werk und nicht nur die eine
//
// **Weil ein 429 eine Aussage der Gegenstelle über sich ist**, nicht über eine
// Endpunktklasse. Scryfall schreibt: „It is not acceptable to ignore HTTP 429
// responses" — und wer die zweite Klasse weiterlaufen ließe, ignorierte es
// halb (#1037).
//
// **Die Abstände bleiben getrennt.** Sie sind je Klasse verschieden und dürfen
// es sein; zusammengeworfen bremsten sie die allgemeinen Wege auf ein Fünftel
// dessen, was erlaubt ist. Zwei Zusagen, nicht eine.
//
// # Warum kein Verklemmen
//
// Die Schranke des Werks schützt die **Karte**, die einer Drossel ihren
// **Zustand**. [Drossel.Sperre] greift nie auf das Werk zu — die Reihenfolge
// ist damit immer dieselbe.
func (w *Drosselwerk) SperreAlle(bis time.Time) {
	// **Und über Prozessgrenzen hinweg** (#1163, Befund Z7-03). Ohne das
	// sendet ein zweiter Prozess während `Retry-After` weiter — und ignoriert
	// das 429 damit gemeinsam, auch wenn dieser es brav einhält.
	//
	// **Je Klasse eine Nachbardatei**, weil es je Klasse eine gibt; die
	// Sperre gilt trotzdem für alle, denn sie wird in jede eingetragen.
	w.schranke.Lock()
	drosseln := make([]*Drossel, 0, len(w.drosseln))

	for _, eine := range w.drosseln {
		drosseln = append(drosseln, eine)
	}
	w.schranke.Unlock()

	for _, eine := range drosseln {
		eine.grenze.Sperren(bis)
	}

	for _, eine := range drosseln {
		eine.Sperre(bis)
	}
}

// grenzeFuer baut die Absprache einer Klasse — nil ohne Verzeichnis.
//
// **Der Schlüssel wird gehasht.** Er kommt vom Dienst und darf alles sein, was
// ein Dienst für eine Klasse hält; ein `/` darin schriebe die Nachbardatei
// woanders hin.
func (w *Drosselwerk) grenzeFuer(schluessel string, abstand time.Duration) *Prozessgrenze {
	if w.absprache == nil {
		return nil
	}

	summe := sha256.Sum256([]byte(schluessel))
	name := hex.EncodeToString(summe[:8]) + ".drossel"

	// **Der Ort wird bei jedem Aufruf gefragt.** Eine Drossel wird einmal
	// gebaut und dann wiederverwendet; ein festgehaltener Ort bliebe für
	// immer der des ersten Aufrufs. Warum das zählt, steht bei
	// [Prozessgrenze.ort].
	return NeueProzessgrenzeAn(func() string {
		ort := w.Abspracheort()
		if ort == "" {
			return ""
		}

		return filepath.Join(ort, name)
	}, abstand, w.jetzt, w.schlafe)
}
