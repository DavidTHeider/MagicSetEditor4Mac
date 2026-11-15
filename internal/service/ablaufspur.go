package service

// **Die Ablaufspur** — OPS-05.AK-02/03
// (#1722).
//
// # Was eine Zeile ist
//
// **Ein Ruf über eine Fläche**: wann, welche Vorgangskennung, welcher Weg,
// wie lange, wie er ausging. Kein Rumpf, keine Parameter, kein roher Pfad —
// der Weg ist das Muster (`GET /sets/{kennung}`) oder der Name des
// Werkzeugs. Mit der Kennung findet man die Protokollzeilen, die derselbe Ruf
// schreibt — bei Dienst und Werkzeugserver auf deren Standardfehlerausgabe.
//
// # Warum ein Ring und keine Datei
//
// **Weil AK-03 keinen unbegrenzten lokalen Datenbestand will**, und ein Ring
// kann keiner werden: er hält höchstens `Spurgrenze` Einträge, keinen älter
// als die Aufbewahrung, und er lebt so lange wie der Prozess. Wer ihn
// behalten will, sichert die Metriken des Prozesses, der ihn führt.
//
// # Abgetastet wird genau
//
// **Bei zehn Prozent jeder zehnte Ruf**, gezählt und nicht gewürfelt: zwei
// Läufe über dieselben Rufe geben dieselbe Spur, und ein Test kann sie
// nachzählen.

import (
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Spurpunkt ist ein Eintrag der Spur.
type Spurpunkt struct {
	Zeit    time.Time              `json:"zeit"`
	Vorgang domain.Vorgangskennung `json:"vorgang"`
	Flaeche domain.Dienstmodus     `json:"flaeche"`
	Weg     string                 `json:"weg"`
	Art     string                 `json:"art"`
	DauerMs float64                `json:"dauerMs"`
	Ausgang domain.Rufausgang      `json:"ausgang"`
}

// Spurauskunft sagt, was die Spur hält — und was sie nicht hält.
type Spurauskunft struct {
	An            bool  `json:"an"`
	Abtastung     int   `json:"abtastung"`
	Grenze        int   `json:"grenze"`
	AufbewahrungS int64 `json:"aufbewahrungS"`

	// Gesehen, Aufgenommen, Verdraengt und Abgelaufen sagen, wie vollständig
	// die Punkte sind — eine volle Spur sähe sonst aus wie eine vollständige.
	Gesehen     int64 `json:"gesehen"`
	Aufgenommen int64 `json:"aufgenommen"`
	Verdraengt  int64 `json:"verdraengt"`
	Abgelaufen  int64 `json:"abgelaufen"`

	Punkte []Spurpunkt `json:"punkte"`
}

// ablaufspur hält die Punkte in einem **Ring fester Größe**: `anfang` ist der
// älteste, `anzahl` die belegten Plätze. Bis zum Zweitblick über Phase 6.4
// (Befund 15) stand hier ein Slice, und eine volle Spur kopierte bei jedem
// aufgenommenen Ruf alle Einträge — unter der Sperre, die jeder Ruf nimmt.
var ablaufspur = struct {
	sperre sync.Mutex
	regel  domain.Messregel
	ring   []Spurpunkt
	anfang int
	anzahl int
	jetzt  func() time.Time
	zahlen Spurauskunft
}{jetzt: time.Now}

// SpurEinrichten setzt die Regel und leert die Spur.
func SpurEinrichten(regel domain.Messregel) {
	ablaufspur.sperre.Lock()
	defer ablaufspur.sperre.Unlock()

	ablaufspur.regel = regel
	ablaufspur.ring, ablaufspur.anfang, ablaufspur.anzahl = nil, 0, 0

	// **Nur eine eingeschaltete Spur belegt ihren Platz** — der Ring ist bei
	// der Vorgabe tausend Einträge groß, und aus soll nichts kosten.
	if regel.Spur {
		ablaufspur.ring = make([]Spurpunkt, max(regel.Spurgrenze, 1))
	}
	ablaufspur.zahlen = Spurauskunft{}
}

// SpurUhrSetzen setzt die Uhr der Spur — für Läufe.
func SpurUhrSetzen(jetzt func() time.Time) {
	ablaufspur.sperre.Lock()
	defer ablaufspur.sperre.Unlock()

	ablaufspur.jetzt = jetzt
}

func spurZuruecksetzen() {
	ablaufspur.sperre.Lock()
	defer ablaufspur.sperre.Unlock()

	ablaufspur.regel = domain.Messregel{}
	ablaufspur.ring, ablaufspur.anfang, ablaufspur.anzahl = nil, 0, 0
	ablaufspur.zahlen = Spurauskunft{}
	ablaufspur.jetzt = time.Now
}

// spure nimmt einen Ruf auf, wenn die Spur an ist und die Abtastung ihn trifft.
func spure(vorgang domain.Vorgangskennung, punkt Laufzeitpunkt) {
	ablaufspur.sperre.Lock()
	defer ablaufspur.sperre.Unlock()

	regel := ablaufspur.regel
	if !regel.Spur {
		return
	}

	ablaufspur.zahlen.Gesehen++

	// **Genau abgetastet**: aufgenommen wird, sobald der Anteil der
	// gesehenen Rufe eine ganze Zahl weiterspringt.
	gesehen := ablaufspur.zahlen.Gesehen
	if gesehen*int64(regel.Abtastung)/100 == (gesehen-1)*int64(regel.Abtastung)/100 {
		return
	}

	jetzt := ablaufspur.jetzt()
	abgelaufenEntfernen(jetzt)

	art := ArtLesend
	if punkt.Schreibend {
		art = ArtSchreibend
	}

	// **Voll verdrängt der älteste** — er rückt aus dem Ring, statt dass
	// die übrigen umziehen.
	if ablaufspur.anzahl == len(ablaufspur.ring) {
		ablaufspur.anfang = (ablaufspur.anfang + 1) % len(ablaufspur.ring)
		ablaufspur.anzahl--
		ablaufspur.zahlen.Verdraengt++
	}

	ablaufspur.ring[(ablaufspur.anfang+ablaufspur.anzahl)%len(ablaufspur.ring)] = Spurpunkt{
		Zeit: jetzt, Vorgang: vorgang, Flaeche: punkt.Flaeche, Weg: punkt.Weg, Art: art,
		DauerMs: float64(punkt.Dauer) / float64(time.Millisecond), Ausgang: punkt.Ausgang,
	}
	ablaufspur.anzahl++
	ablaufspur.zahlen.Aufgenommen++
}

// abgelaufenEntfernen nimmt, was älter ist als die Aufbewahrung — unter der
// Sperre.
func abgelaufenEntfernen(jetzt time.Time) {
	grenze := jetzt.Add(-ablaufspur.regel.Aufbewahrung)

	for ablaufspur.anzahl > 0 && ablaufspur.ring[ablaufspur.anfang].Zeit.Before(grenze) {
		ablaufspur.anfang = (ablaufspur.anfang + 1) % len(ablaufspur.ring)
		ablaufspur.anzahl--
		ablaufspur.zahlen.Abgelaufen++
	}
}

// Spurstand gibt die Spur samt ihren Zahlen — Kopien.
func Spurstand() Spurauskunft {
	ablaufspur.sperre.Lock()
	defer ablaufspur.sperre.Unlock()

	regel := ablaufspur.regel
	auskunft := ablaufspur.zahlen
	auskunft.An = regel.Spur
	auskunft.Punkte = []Spurpunkt{}

	if !regel.Spur {
		return auskunft
	}

	abgelaufenEntfernen(ablaufspur.jetzt())

	auskunft = ablaufspur.zahlen
	auskunft.An, auskunft.Abtastung, auskunft.Grenze = true, regel.Abtastung, regel.Spurgrenze
	auskunft.AufbewahrungS = int64(regel.Aufbewahrung / time.Second)
	auskunft.Punkte = make([]Spurpunkt, ablaufspur.anzahl)

	for i := range auskunft.Punkte {
		auskunft.Punkte[i] = ablaufspur.ring[(ablaufspur.anfang+i)%len(ablaufspur.ring)]
	}

	return auskunft
}
