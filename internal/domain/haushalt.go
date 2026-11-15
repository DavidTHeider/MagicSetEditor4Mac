package domain

import (
	"math"
	"time"
)

// **Ressourcenbudgets ohne gegenseitige Aussperrung** — Werkstück O5 von
// Phase 5.2 (#1429).
//
// # Drei Fragen, drei Begriffe
//
//   - **Wie viel darf einer?** Ein [Kontingent] je Principal und eines für
//     alle, gerechnet als [Guthaben]: ein Eimer mit Decke (`Burst`), der sich
//     gleichmäßig nachfüllt. Jeder Ruf kostet nach seiner [Kostenklasse].
//   - **Wie viele zugleich?** Je [Ressourcenklasse] eine Parallelitätsgrenze —
//     zwei Rasterungen, eine Einfuhr, eine Ausfuhr. Was hier steht, teilen
//     Fenster, Kommandozeile, MCP und REST; wer eine Fläche umgeht, umgeht
//     die Grenze nicht.
//   - **Wer rät, wartet länger.** Eine [Sperrlage] je vorgezeigtem Nachweis:
//     drei freie Versuche, dann verdoppelt sich die Sperre bis zum Deckel.
//     **Je Nachweis, nicht je Rechner** — sonst sperrte ein Angreifer den
//     legitimen Aufrufer mit aus, wie es die eine Drossel bis O5 tat.
//
// **Keine Uhr in diesem Paket.** Jede Rechnung nimmt `jetzt` entgegen; wer
// sie ruft, bringt die Uhr mit (§16) — und ein Sprung rückwärts füllt nichts
// ab und macht nichts negativ.
//
// # Warum nicht das feste Fenster der REST-Drossel
//
// Die Drossel aus Phase 4 zählt und setzt zurück; das reicht gegen das Raten
// einer Marke. Ein Budget muss mehr: ein kurzer Schub darf durch, ein
// Dauerlauf nicht, und die Auskunft „wann wieder" muss stimmen — `Retry-After`
// war bis O5 eine feste 60, gleich wie viel vom Fenster schon um war.

// Ressourcenklasse ist, was eine Arbeit an der Maschine bindet.
type Ressourcenklasse string

const (
	RessourceRendern Ressourcenklasse = "rendern"
	RessourceEinfuhr Ressourcenklasse = "einfuhr"
	RessourceAusfuhr Ressourcenklasse = "ausfuhr"
	RessourceBild    Ressourcenklasse = "bild"
	RessourceMasse   Ressourcenklasse = "masse"

	// RessourceSuche ist eine Frage an den Retrievalindex — Werkstück A15
	// von Phase 5.3.
	//
	// **Sie steht neben dem Indexbau und nicht in ihm.** Das Issue verlangt
	// es als eigene Zeile: „Interaktive Suche/Editor dürfen durch Rebuild
	// nicht verhungern." Eine gemeinsame Klasse hieße, dass ein laufender
	// Neubau jede Suche blockiert.
	RessourceSuche Ressourcenklasse = "suche"

	// RessourceIndexbau ist der Aufbau eines Retrievalindex.
	RessourceIndexbau Ressourcenklasse = "indexbau"
)

// Ressourcenklassen sind alle, in fester Reihenfolge.
//
// **Die Aufzählung ist die Grenze.** Eine Klasse, die hier fehlt, ist im
// [Haushalt] **unbegrenzt** — sie bekommt keinen Kanal und damit auch keine
// Warteschlange. Wer eine neue Art von Arbeit einführt, trägt sie hier und in
// [Vorgabegrenzen] ein; ein Wächter zählt beide gegeneinander.
func Ressourcenklassen() []Ressourcenklasse {
	return []Ressourcenklasse{
		RessourceRendern, RessourceEinfuhr, RessourceAusfuhr, RessourceBild,
		RessourceMasse, RessourceSuche, RessourceIndexbau,
	}
}

// Kostenklasse ist, was ein Ruf vom Guthaben nimmt.
type Kostenklasse string

const (
	KostenLesen     Kostenklasse = "lesen"
	KostenSchreiben Kostenklasse = "schreiben"
	KostenBild      Kostenklasse = "bild"
	KostenRendern   Kostenklasse = "rendern"
	KostenAusfuhr   Kostenklasse = "ausfuhr"
	KostenEinfuhr   Kostenklasse = "einfuhr"
	KostenMasse     Kostenklasse = "masse"
)

// kosten je Klasse.
//
// **Die Zahlen sind Verhältnisse, keine Messungen**: ein Schreiben wiegt zwei
// Lesen, eine Rasterung fünf, eine Einfuhr acht, ein Massenvorgang zehn. Was
// zählt, ist, dass ein Dauerlauf teurer Rufe das Budget schneller leert als
// einer billiger — die Verhältnisse stehen in der Dokumentation, und wer sie
// ändert, ändert sie hier und dort.
//
// **Ein Abbruch kostet nichts** (#2003, `kostentafel.go`): er nimmt Last,
// statt welche zu machen.
var kosten = map[Kostenklasse]int{
	KostenLesen: 1, KostenSchreiben: 2, KostenBild: 3, KostenRendern: 5,
	KostenAusfuhr: 8, KostenEinfuhr: 8, KostenMasse: 10, KostenAbbruch: 0,
}

// Kostenklassen sind alle, in fester Reihenfolge.
func Kostenklassen() []Kostenklasse {
	return []Kostenklasse{
		KostenLesen, KostenSchreiben, KostenBild, KostenRendern,
		KostenAusfuhr, KostenEinfuhr, KostenMasse, KostenAbbruch,
	}
}

// Kosten sind die Einheiten, die ein Ruf dieser Klasse nimmt — eine unbekannte
// Klasse kostet, was Lesen kostet.
func (k Kostenklasse) Kosten() int {
	if wert, bekannt := kosten[k]; bekannt {
		return wert
	}

	return kosten[KostenLesen]
}

// Kontingent ist ein Eimer: `Burst` Einheiten als Decke, `Nachfuellen`
// Einheiten je `Fenster` als Zulauf.
type Kontingent struct {
	Burst       int
	Nachfuellen int
	Fenster     time.Duration
}

// jeSekunde ist der Zulauf als Rate.
func (k Kontingent) jeSekunde() float64 {
	if k.Fenster <= 0 || k.Nachfuellen <= 0 {
		return 0
	}

	return float64(k.Nachfuellen) / k.Fenster.Seconds()
}

// Dauer ist, wie lange der Zulauf für `einheiten` braucht — das Fenster, wenn
// es keinen Zulauf gibt: eine Zahl muss es sein, und „nie" ist keine.
func (k Kontingent) Dauer(einheiten float64) time.Duration {
	if einheiten <= 0 {
		return 0
	}

	rate := k.jeSekunde()
	if rate <= 0 {
		return k.Fenster
	}

	return time.Duration(math.Ceil(einheiten / rate * float64(time.Second)))
}

// Guthaben ist der Stand eines Kontingents zu einer Zeit.
type Guthaben struct {
	kontingent Kontingent
	stand      float64
	zeit       time.Time
}

// NeuesGuthaben beginnt voll — ein neuer Principal darf seinen Burst.
func NeuesGuthaben(kontingent Kontingent, jetzt time.Time) Guthaben {
	return Guthaben{kontingent: kontingent, stand: float64(kontingent.Burst), zeit: jetzt}
}

// nachfuellen bringt den Stand auf `jetzt` — geklemmt an die Decke, und ein
// Sprung der Uhr nach hinten füllt nichts ab.
func (g Guthaben) nachfuellen(jetzt time.Time) Guthaben {
	if vergangen := jetzt.Sub(g.zeit); vergangen > 0 {
		g.stand = math.Min(float64(g.kontingent.Burst), g.stand+vergangen.Seconds()*g.kontingent.jeSekunde())
	}

	g.zeit = jetzt

	return g
}

// Stand ist, was jetzt bereitsteht.
func (g Guthaben) Stand(jetzt time.Time) float64 {
	return g.nachfuellen(jetzt).stand
}

// Abbuchen nimmt `kosten` Einheiten, wenn sie da sind — sonst sagt es, wie
// lange es dauert, bis sie da wären.
//
// **Kosten über der Decke gehen nie durch**; die Wartezeit ist dann das
// Fenster, denn eine Zahl muss es sein, und „nie" ist keine.
func (g *Guthaben) Abbuchen(kosten int, jetzt time.Time) (bool, time.Duration) {
	if kosten <= 0 {
		return true, 0
	}

	*g = g.nachfuellen(jetzt)

	if kosten > g.kontingent.Burst {
		return false, g.kontingent.Fenster
	}

	if g.stand >= float64(kosten) {
		g.stand -= float64(kosten)

		return true, 0
	}

	return false, g.kontingent.Dauer(float64(kosten) - g.stand)
}

// Die Sperre gegen das Raten.
const (
	// SperreFreieVersuche sind Fehlversuche ohne Folge — ein Tippfehler ist
	// kein Angriff.
	SperreFreieVersuche = 3
	// SperreGrund ist die erste Sperre; jede weitere verdoppelt sie.
	SperreGrund = time.Second
	// SperreHoechstens ist der Deckel.
	SperreHoechstens = 5 * time.Minute
)

// Sperrlage ist, wie oft ein Nachweis schon danebenlag — und bis wann er warten muss.
type Sperrlage struct {
	Fehlversuche int
	GesperrtBis  time.Time
}

// Fehlgeschlagen zählt einen Versuch und setzt die Sperre neu.
func (s Sperrlage) Fehlgeschlagen(jetzt time.Time) Sperrlage {
	s.Fehlversuche++

	if s.Fehlversuche <= SperreFreieVersuche {
		return s
	}

	// Verdoppeln bis zum Deckel — und die Schiebung bleibt klein, damit sie
	// nie überläuft.
	stufe := s.Fehlversuche - SperreFreieVersuche - 1
	dauer := SperreHoechstens

	if stufe < 20 {
		dauer = min(SperreGrund<<stufe, SperreHoechstens)
	}

	s.GesperrtBis = jetzt.Add(dauer)

	return s
}

// Gesperrt sagt, ob der Nachweis warten muss — und wie lange noch, nie negativ.
func (s Sperrlage) Gesperrt(jetzt time.Time) (bool, time.Duration) {
	if s.GesperrtBis.IsZero() || !jetzt.Before(s.GesperrtBis) {
		return false, 0
	}

	return true, s.GesperrtBis.Sub(jetzt)
}

// Gelungen löscht die Lage — ein richtiger Nachweis hat nichts zu büßen.
func (s Sperrlage) Gelungen() Sperrlage {
	return Sperrlage{}
}

// Haushaltsgrenzen sind alle Grenzen an einer Stelle.
type Haushaltsgrenzen struct {
	// Parallel ist, wie viele Arbeiten einer Klasse zugleich laufen.
	Parallel map[Ressourcenklasse]int
	// Antwort ist die größte Antwort in Bytes, die eine Fläche hinausgibt.
	Antwort int64
	// Principal ist das Kontingent je Aufrufer, Global das aller zusammen.
	Principal Kontingent
	Global    Kontingent
	// Reserve ist, was vom globalen Guthaben nur das Fenster nehmen darf —
	// die Schutzreserve, die den Editor arbeitsfähig hält, wenn ein
	// Maschinenzugang das Übrige aufbraucht.
	Reserve int
	// OhneKontingente schaltet die Guthaben ab — Principal, Global und mit
	// ihnen die Reserve (QA-08.AK-03, `kontingentschalter.go`). **Die Null
	// heißt: sie gelten**; wer Grenzen von Hand baut, bekommt den Schutz,
	// ohne an ihn zu denken.
	OhneKontingente bool

	// Verbindungen ist, wie viele Verbindungen ein Dienst über HTTP zugleich
	// annimmt — `mse-rest` und `mse-mcp`; eine darüber wartet
	// (`internal/rest/gedraenge.go`).
	Verbindungen int

	// Auftraege sind die Grenzen des Auftragsbuchs im Dienst
	// (`internal/rest/auftraege.go`, FMT-03.AK-02).
	Auftraege Auftragsgrenzen
}

// Auftragsgrenzen sind die Grenzen des Auftragsbuchs: wie viele zugleich
// laufen, wie viele beendete es behält und wie lange.
//
// **Sie stehen hier und nicht beim Buch** (#2003): `grenzen` nennt sie an
// jeder Fläche, und das Buch liest dieselben Zahlen.
type Auftragsgrenzen struct {
	Laufend      int
	Behalten     int
	Aufbewahrung time.Duration
}

// Vorgabegrenzen sind die Grenzen dieser Fassung.
//
// **Gewählt, nicht gemessen** — nach dem Zweck: zwei Rasterungen zugleich, weil
// ein zweiter Kern da ist und die Vorschau nicht auf den Druckbogen warten
// soll; eine Einfuhr und eine Ausfuhr, weil beide den Bestand anfassen; 120
// Einheiten je Principal und Minute halten die Parität zur alten Drossel (60
// Schreiben oder 120 Lesen); 600 global mit 120 Reserve fürs Fenster.
func Vorgabegrenzen() Haushaltsgrenzen {
	return Haushaltsgrenzen{
		Parallel: map[Ressourcenklasse]int{
			RessourceRendern: 2, RessourceEinfuhr: 1, RessourceAusfuhr: 1,
			RessourceBild: 4, RessourceMasse: 1,
			// **Vier Suchen und ein Indexbau.** Eine Suche liest und rechnet
			// ein Skalarprodukt je Zeile; ein Indexbau schreibt und hält den
			// Einbetter besetzt. Die Zahlen sind **gewählt** wie die darüber
			// — was sie nicht sind, ist eine RAM-Zusage: ein Zähler hält
			// keinen Speicher frei (Issue, Werkstück 4).
			RessourceSuche: 4, RessourceIndexbau: 1,
		},
		Antwort:   32 << 20,
		Principal: Kontingent{Burst: 120, Nachfuellen: 120, Fenster: time.Minute},
		Global:    Kontingent{Burst: 600, Nachfuellen: 600, Fenster: time.Minute},
		Reserve:   120,

		// **Sechzehn Verbindungen**, nach dem Zweck: ein Dienst hat einen
		// Aufrufer, und mehr als eine Handvoll gleichzeitiger Aufträge ist
		// keine Benutzung, sondern ein Versehen oder ein Angriff. Nach oben
		// begrenzen sie den Schaden halb geschriebener Anfragen — die
		// erreichen kein Guthaben (Zweitblick über Phase U).
		Verbindungen: 16,

		// **Vier Aufträge laufen zugleich** — die Fachläufe dahinter gehen
		// ohnehin einer nach dem anderen (`Parallel`), und jeder hält
		// Speicher für seine Antwort. Beendete bleiben eine Stunde,
		// höchstens sechzehn.
		Auftraege: Auftragsgrenzen{Laufend: 4, Behalten: 16, Aufbewahrung: time.Hour},
	}
}
