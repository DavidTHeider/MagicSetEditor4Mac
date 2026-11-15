package domain_test

// **Was der Abgleich bei zwanzigtausend Zeilen kostet** — Werkstück `T4` von
// Phase 5.6
// (#1469).
//
// # Warum hier eine Zeit steht und sonst nirgends
//
// **Weil das Issue eine Messung verlangt** und nicht ein Gefühl: „keine
// O(n*m)-Vollpaarung" ist eine Aussage über das Wachstum, und die einzige
// Art, sie zu halten, ist, sie zu messen. Die Schranke ist **großzügig** —
// sie trennt eine Rechnung über das Produkt beider Listen von einer über
// ihre Summe, und nicht zwei Läufe unter Last voneinander.
//
// **Die Maschine steht dabei** (§10, „die Reichweite einer Messung"): gemessen
// am 2026-09-18 auf einem Apple-Silicon-Mac, allein. Wer sie anderswo fährt,
// bekommt andere Zahlen — die **Trennung** bleibt, weil zwischen einer
// Vollpaarung und einem Verzeichnis nicht Prozente liegen, sondern
// Größenordnungen.

import (
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dieMessgroessen sind die Zahlen aus dem Issue.
const (
	kleineMessung = 5_000
	grosseMessung = 20_000

	// dieSchranke fängt, was auf jedem Rechner zu langsam ist — **und sonst
	// nichts**.
	//
	// **Dreißig Sekunden, und das ist Absicht.** Der erste Anlauf stand auf
	// einer halben Sekunde und wurde in der Pipeline rot: dort läuft die
	// Teststufe mit Race-Detektor und Abdeckung, und dieselbe Probe kostete
	// 1,15 s statt 24 ms. Eine Schranke in dieser Größenordnung misst die
	// Instrumentierung und die Auslastung, nicht die Rechnung.
	//
	// **Was die Rechnung misst, ist das Verhältnis** — es hält in beiden
	// Bedingungen, weil beide Hälften derselben Messung unter derselben Last
	// laufen. Gemessen am 2026-09-18, Apple-Silicon-Mac, allein:
	//
	//	           ohne Instrumentierung   mit -race und Abdeckung
	//	 5.000                     3,2 ms                  31,7 ms
	//	20.000                    11,8 ms                 120,2 ms
	//	12.500 Bez.               ~30 ms                  111,5 ms
	//	50.000 Bez.               23,8 ms                 531,5 ms
	//
	// Die Verhältnisse sind 3,7 und 3,8 beziehungsweise 4,8 — in beiden
	// Bedingungen die Summe und nicht das Produkt. Die Schranke hier ist nur
	// noch die Notbremse: eine Vollpaarung bräuchte auch instrumentiert ein
	// Vielfaches davon.
	dieSchranke = 30 * time.Second
)

// vieleAngebote baut n Angebotszeilen mit je eigenem Druck.
func vieleAngebote(anzahl int) []domain.Angebotszeile {
	zeilen := make([]domain.Angebotszeile, 0, anzahl)

	for lauf := range anzahl {
		kennung := "druck-" + strconv.Itoa(lauf)

		zeilen = append(zeilen, domain.Angebotszeile{
			Druck: kennung, Konzept: "konzept-" + strconv.Itoa(lauf),
			Menge: 4, Zustand: domain.ZustandNearMint, Sprache: "de",
			Finish: domain.FinishNonfoil, Bezeichnung: kennung,
		})
	}

	return zeilen
}

// vieleWuensche baut n Wunschzeilen, die auf dieselben Drucke zeigen.
func vieleWuensche(anzahl int) []domain.Wunschzeile {
	zeilen := make([]domain.Wunschzeile, 0, anzahl)

	for lauf := range anzahl {
		kennung := "druck-" + strconv.Itoa(lauf)

		zeilen = append(zeilen, domain.Wunschzeile{
			Druck: kennung, Konzept: "konzept-" + strconv.Itoa(lauf),
			Menge: 2, Zustand: domain.ZustandNearMint, Sprache: "de",
			Finish: domain.FinishNonfoil, Art: domain.WunschGenau,
			Rang: domain.RangNormal, Bezeichnung: kennung,
		})
	}

	return zeilen
}

func TestDerAbgleichWaechstMitDerSummeUndNichtMitDemProdukt(t *testing.T) {
	// **Kein `t.Parallel()`** — zwei Messungen, die sich die Maschine teilen,
	// messen die Maschine.

	klein, gross := gemesseneGroessen(t, kleineMessung, grosseMessung)

	t.Logf("%d: %v · %d: %v", kleineMessung, klein, grosseMessung, gross)

	// **Das Verhältnis ist die Aussage, nicht die Zahl.** Viermal so viele
	// Zeilen kosten bei einer Rechnung über die Summe etwa viermal so viel
	// und bei einer über das Produkt etwa sechzehnmal. Die Schranke liegt
	// dazwischen — sie trennt die zwei Wachstumsarten und nicht zwei
	// Rechner voneinander.
	//
	// **Gemessen am 2026-09-18** auf einem Apple-Silicon-Mac, allein: über
	// ein Verzeichnis 5.000 → wenige Millisekunden, 20.000 → wenige
	// Millisekunden mehr. Die Vollpaarung davor brauchte für dieselbe Probe
	// 137 ms und 1,9 s — Faktor **vierzehn** bei viermal so vielen Zeilen.
	const hoechstesVerhaeltnis = 8

	if verhaeltnis := float64(gross) / float64(max(klein, 1)); verhaeltnis > hoechstesVerhaeltnis {
		t.Errorf("viermal so viele Zeilen kosteten das %.1f-fache — "+
			"das ist das Wachstum einer Vollpaarung", verhaeltnis)
	}

	// **Und eine großzügige Schranke daneben**, damit eine Rechnung, die auf
	// jedem Rechner zu langsam ist, nicht durchkommt, nur weil sie
	// gleichmäßig zu langsam ist.
	if gross > dieSchranke {
		t.Errorf("der große Abgleich brauchte %v — die Schranke ist %v",
			gross, dieSchranke)
	}
}

// gemessen fährt einen Abgleich und gibt, was er gekostet hat.
// **Die schnellste von drei Läufen und nicht die erste.**
//
// # Warum das Kleinste und nicht der Mittelwert
//
// **Weil Rauschen nur addiert.** Ein Lauf, der zwischendurch die Maschine
// abgeben muss, dauert länger; keiner dauert deshalb kürzer. Das Minimum ist
// damit die beste Schätzung dafür, was diese Rechnung kostet — ein
// Mittelwert nähme das Rauschen mit hinein und ein Höchstwert nur das
// Rauschen.
//
// # Und warum das nötig wurde
//
// **Weil die Pipeline die Maschine teilt.** Am 2026-09-22 ist dieser Lauf
// dort rot geworden: 31,6 ms gegen 261 ms, Verhältnis 8,3 bei einer
// Schranke von 8. Derselbe Rechner, unmittelbar danach und allein, maß
// dreimal 3,3–4,7 ms gegen 10,7–18,3 ms — Verhältnis 3,2. **Gemessen wurde
// die Auslastung und nicht die Rechnung** (§10), und ein Wächter, der unter
// Last etwas anderes prüft als allein, prüft nichts Verlässliches.
//
// **Seit ENG-03 misst er Rechenzeit** ([rechenzeit]) **und abwechselnd**
// ([gemesseneGroessen]): das Minimum hielt gleichmäßige Last aus, aber keine,
// die zwischen zwei Messungen wechselt — und seit der Racelauf in Teilen
// geht, wechselt sie ständig.
func gemessen(
	t *testing.T, angebot []domain.Angebotszeile, wuensche []domain.Wunschzeile, groesse int,
) time.Duration {
	t.Helper()

	begonnen := rechenzeit(t)
	treffer, befunde := domain.Tauschkandidaten(angebot, wuensche)
	gedauert := rechenzeit(t) - begonnen

	// **Erst das Ergebnis, dann die Zeit.** Eine schnelle Rechnung, die
	// nichts findet, ist keine Leistung.
	if len(treffer) != groesse {
		t.Fatalf("es kamen %d Treffer statt %d", len(treffer), groesse)
	}

	if len(befunde) != 0 {
		t.Fatalf("es kamen %d Befunde", len(befunde))
	}

	return gedauert
}

// gemesseneGroessen misst beide Größen nach der Messregel aus [abwechselnd].
func gemesseneGroessen(t *testing.T, klein, gross int) (kleineZeit, grosseZeit time.Duration) {
	t.Helper()

	kleinesAngebot, kleineWuensche := vieleAngebote(klein), vieleWuensche(klein)
	grossesAngebot, grosseWuensche := vieleAngebote(gross), vieleWuensche(gross)

	return abwechselnd(t,
		func() time.Duration { return gemessen(t, kleinesAngebot, kleineWuensche, klein) },
		func() time.Duration { return gemessen(t, grossesAngebot, grosseWuensche, gross) })
}

// abwechselnd misst zwei Läufe **im Wechsel**, fünf Runden lang, und gibt je
// den schnellsten — so stehen beide unter derselben Last, statt dass der
// kleine eine ruhige und der große eine volle Maschine erwischt.
//
// **Die Messregel beider Verhältniswächter dieser Datei steht hier und nur
// hier**: Rechenzeit ([rechenzeit], in den Läufen selbst), abwechselnd, ohne
// Speicherbereinigung. Sie lief auseinander, solange jeder Wächter seine
// eigene Schleife hatte — der Abgleich maß seit ENG-03 so, die
// fünfzigtausend Beziehungen daneben weiter mit der Wanduhr, dreimal
// hintereinander (Befund 3 des Zweitblicks über Phase 6.2).
//
// **Der Prozess gehört ihr dabei allein**: kein Wächter ruft `t.Parallel`,
// und die abgeschaltete Speicherbereinigung gilt für alle.
func abwechselnd(t *testing.T, erster, zweiter func() time.Duration) (ersteZeit, zweiteZeit time.Duration) {
	t.Helper()

	// **Ohne Speicherbereinigung** während der Messung: der große Lauf löste
	// Sammelzyklen aus, der kleine keinen, und deren Arbeit auf anderen Kernen
	// zählt in der Rechenzeit mit — ohne Last und abwechselnd gemessen kam
	// das 8,9-Fache heraus. Gemessen wird die Rechnung, nicht die Sammlung.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	runtime.GC()

	ersteZeit, zweiteZeit = time.Duration(1<<62-1), time.Duration(1<<62-1)

	for range 5 {
		ersteZeit = min(ersteZeit, erster())
		zweiteZeit = min(zweiteZeit, zweiter())
	}

	return ersteZeit, zweiteZeit
}

// **Fünfzigtausend Beziehungen** — das dritte Maß des Issues: viele Wünsche
// auf wenige Drucke, so dass jeder Druck mehrfach gefragt wird.
func TestFuenfzigtausendBeziehungen(t *testing.T) {
	const drucke = 500

	angebot := vieleAngebote(drucke)
	wenige, viele := vieleBeziehungen(drucke, 12_500), vieleBeziehungen(drucke, 50_000)

	klein, gross := abwechselnd(t,
		func() time.Duration { return beziehungenGemessen(t, angebot, wenige, drucke) },
		func() time.Duration { return beziehungenGemessen(t, angebot, viele, drucke) })

	t.Logf("12.500: %v · 50.000: %v (über %d Drucke)", klein, gross, drucke)

	// **Viermal so viele Beziehungen, höchstens das Achtfache** — dieselbe
	// Aussage wie oben, hier über die Zahl der **Wünsche** je Druck.
	const hoechstesVerhaeltnis = 8

	if verhaeltnis := float64(gross) / float64(max(klein, 1)); verhaeltnis > hoechstesVerhaeltnis {
		t.Errorf("viermal so viele Beziehungen kosteten das %.1f-fache",
			verhaeltnis)
	}

	if gross > dieSchranke {
		t.Errorf("der Abgleich brauchte %v — die Schranke ist %v",
			gross, dieSchranke)
	}
}

// vieleBeziehungen baut Wünsche, die reihum auf wenige Drucke zeigen.
func vieleBeziehungen(drucke, beziehungen int) []domain.Wunschzeile {
	wuensche := make([]domain.Wunschzeile, 0, beziehungen)

	for lauf := range beziehungen {
		kennung := "druck-" + strconv.Itoa(lauf%drucke)

		wuensche = append(wuensche, domain.Wunschzeile{
			Druck: kennung, Konzept: "konzept-" + strconv.Itoa(lauf%drucke),
			Menge: 1, Zustand: domain.ZustandNearMint, Sprache: "de",
			Finish: domain.FinishNonfoil, Art: domain.WunschGenau,
			Rang: domain.RangNormal, Bezeichnung: kennung + "/" + strconv.Itoa(lauf),
		})
	}

	return wuensche
}

// beziehungenGemessen fährt einen Abgleich mit vielen Wünschen je Druck und
// gibt seine Rechenzeit.
//
// **Hier maß bis zum Zweitblick über Phase 6.2 die Wanduhr**, dreimal
// hintereinander — und wurde am 2026-09-23 in der Vorprüfung rot, 8,1-fach
// bei einer Schranke von 8, während derselbe Lauf allein dreimal grün war.
// Behoben war damals der Fall bei [gemessen], nicht die Klasse (§10); jetzt
// misst er wie jener und wiederholt über [abwechselnd].
func beziehungenGemessen(t *testing.T, angebot []domain.Angebotszeile, wuensche []domain.Wunschzeile, drucke int) time.Duration {
	t.Helper()

	begonnen := rechenzeit(t)
	treffer, _ := domain.Tauschkandidaten(angebot, wuensche)
	gedauert := rechenzeit(t) - begonnen

	// **Jede Zeile liegt viermal da** — gedeckt werden können 4 je Druck, der
	// Rest bleibt offen. Das ist die Aussage: ein Stück Papier wird nicht
	// hundertmal vergeben.
	if len(treffer) != drucke*4 {
		t.Fatalf("es kamen %d Treffer statt %d", len(treffer), drucke*4)
	}

	return gedauert
}
