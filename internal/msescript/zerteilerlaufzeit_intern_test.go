package msescript

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// **Der Zerteiler wächst mit der Länge, nicht mit ihrem Quadrat.**
//
// # Der Befund, aus dem dieser Test entstanden ist
//
// `magic.mse-game` bringt nach dem Einbinden seiner Module 582 KB Skript mit.
// Das zu zerteilen kostete **19 Sekunden** — und ein Stilplan, der es als
// Umgebung braucht, kostete 40. Gemessen über vier Größen: 10 % → 0,28 s,
// 25 % → 1,65 s, 50 % → 6,26 s, 100 % → 19,4 s. Doppelte Länge, vierfache Zeit.
//
// Die Ursache stand in `lexer.operator`: sie wandelte bei **jedem** Merkmal den
// gesamten Rest der Eingabe in eine Zeichenkette, um Operatorpräfixe zu
// vergleichen. Ein O(n)-Schritt je Merkmal ergibt O(n²) über die Datei
// (#195).
//
// # Warum als Verhältnis und nicht als Sekundenzahl
//
// Eine absolute Schranke wäre auf einer langsameren Maschine rot, ohne dass
// etwas kaputt ist. Das **Verhältnis** zweier Läufe auf derselben Maschine sagt
// dagegen etwas über die Ordnung — und genau darum geht es.
//
// **Was er nicht prüft:** ob der Zerteiler schnell *genug* ist. Er prüft die
// Ordnung; eine langsame Konstante ginge durch.
func TestDerZerteilerWaechstNichtQuadratisch(t *testing.T) {
	// **Operatoren im Text**, denn genau ihretwegen entstand die Quadratik.
	// Ein Skript aus lauter Namen liefe auch mit dem alten Lexer linear.
	einheit := "a := 1 + 2 * 3 - 4 / 5; b := a >= 2 and a <= 9;\n"

	// **Groß genug, dass die Uhr mehr misst als sich selbst.**
	//
	// Hier standen 400 und 1600 Einheiten, und der kurze Lauf kostete damit
	// 0,6 Millisekunden. In dieser Größenordnung schlägt die Zeitscheibe des
	// Betriebssystems durch: lief nebenher ein zweites Testpaket, streute das
	// Verhältnis auf 7 und der Wächter wurde rot, ohne dass etwas kaputt war.
	//
	// **Vervierfacht statt die Schranke gelockert.** Die Schranke ist die
	// Aussage — wer sie hochsetzt, damit ein flatteriger Test grün wird,
	// schwächt die Zusage statt die Messung. Bei 6400 Einheiten liegt der
	// lange Lauf im zweistelligen Millisekundenbereich, und die Zeitscheibe
	// fällt nicht mehr ins Gewicht.
	kurz := strings.Repeat(einheit, 1600)
	lang := strings.Repeat(einheit, 6400)

	kurzeZeit, langeZeit := zerteilzeiten(t, kurz, lang)

	// Viermal so lang darf gut das Sechsfache kosten — quadratisch wären es
	// sechzehn. Der Abstand zwischen 6 und 16 ist die Aussage.
	if langeZeit > 6*kurzeZeit {
		t.Errorf("%d Einheiten: %v, %d Einheiten: %v — das wächst überlinear",
			1600, kurzeZeit, 6400, langeZeit)
	}
}

// zerteilzeiten misst beide Längen **abwechselnd**, fünf Runden lang, und gibt
// je die schnellste — in **Rechenzeit** und nicht an der Wanduhr (siehe
// [rechenzeit]).
//
// **Den schnellsten**, weil ein Einbruch durch eine fremde Last auf der
// Maschine nach oben streut und nie nach unten.
//
// **Abwechselnd**, weil die Last in der Pipeline wechselt, während die Teile
// des Racelaufs fertig werden (ENG-03, #1728). Drei kurze Läufe hintereinander
// und dann drei lange erwischten zwei verschiedene Maschinen: die kurzen eine
// ruhige, die langen eine volle — und das Verhältnis maß den Wechsel. Auch die
// Rechenzeit schützt davor nicht ganz, denn unter Last rechnet ein Lauf auf
// einem langsameren Kern.
//
// **Und am Zerteiler selbst, nicht an [Zerteile].** Seit #683 merkt sich
// `Zerteile` jeden Baum; der zweite und dritte Lauf waren damit Speichertreffer,
// und der schnellste von dreien maß das Nachschlagen eines Schlüssels — 1,2 µs
// für 1.600 Einheiten. Unter Last streute dessen Verhältnis über sechs, und
// der Wächter wurde rot für etwas, das er gar nicht hielt (gefunden in der
// Vorprüfung zu QA-03, 2026-09-23). Gemessen wird deshalb [zerteileFrisch].
func zerteilzeiten(t *testing.T, kurz, lang string) (kurzeZeit, langeZeit time.Duration) {
	t.Helper()

	kurzeZeit, langeZeit = time.Duration(1<<62-1), time.Duration(1<<62-1)

	// **Ohne Speicherbereinigung** während der Messung — dieselbe Ursache wie
	// im Tauschabgleich (`internal/domain/tauschmessung_test.go`): ein langer
	// Lauf löst Sammelzyklen aus, ein kurzer vielleicht keinen, und ihre
	// Arbeit auf anderen Kernen zählt in der Rechenzeit mit.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	runtime.GC()

	for range 5 {
		kurzeZeit = min(kurzeZeit, zerteilzeit(t, kurz))
		langeZeit = min(langeZeit, zerteilzeit(t, lang))
	}

	return kurzeZeit, langeZeit
}

// zerteilzeit misst ein Zerteilen in Rechenzeit.
func zerteilzeit(t *testing.T, quelle string) time.Duration {
	t.Helper()

	anfang := rechenzeit(t)

	if _, err := zerteileFrisch(quelle); err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	return rechenzeit(t) - anfang
}
