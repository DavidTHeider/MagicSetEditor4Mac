// Package vorschauweg liefert gerasterte Kartenbilder an das eigene Fenster aus.
//
// **Der Bildweg** — #1270.
//
// # Der Befund
//
// Eine Vorschau brauchte im Fenster **0,4–1,0 s**, in Go aber nur **53 ms**.
// Gemessen am 2026-08-28, im laufenden Fenster:
//
//	Datenstand   (winzige Antwort)      0,3 ms
//	Kartenzeilen (107 Zeilen JSON)     15,8 ms
//	Vorschau     (242 KB Data-URI)    554,0 ms
//
// **Die Naht ist nicht langsam — die Datenmenge ist es.** Ein PNG von 182 KB
// wird als Base64 zu 242 KB, wandert als **JSON-Zeichenkette** durch die
// Brücke, wird dort entwertet und noch einmal kopiert. Über die Größe gemessen:
// 86 KB → 397 ms, 268 KB → 873 ms, 694 KB → 1.648 ms.
//
// # Die Abhilfe
//
// **Das Bild geht gar nicht durch die Naht.** Wails betreibt für die
// eingebettete Oberfläche ohnehin eine Assetleitung; was sie nicht kennt,
// reicht sie an diesen Handler weiter. Die Antwort auf `Vorschau` trägt dann
// eine **Adresse** statt der Bytes, und der Browser holt sie über denselben Weg
// wie jede andere Datei — ohne Base64, ohne JSON, ohne zweite Kopie.
//
// Nachgemessen im Fenster nach dem Umbau: **72–75 ms** für eine Vorschau,
// davon 2–4 ms für das Bild.
//
// # Warum es nicht `bildweg` heißt
//
// Weil `service.Bildweg` schon etwas anderes ist: der Weg zu einem **Bild von
// Scryfall**. Zwei Wege mit demselben Namen wären in jedem Review eine
// Rückfrage. Dieser hier trägt **Vorschauen** und sonst nichts.
//
// # Warum ein eigenes Paket
//
// **Weil `internal/app` das Netz nicht kennen darf** (#54,
// `internal/schichten/netz_test.go`). Ein `http.Handler` ist ohne `net/http`
// nicht zu schreiben, und der Wächter prüft strukturell — mit Absicht: wer das
// Netz nicht erreichen **kann**, lädt nicht.
//
// Der Eintrag in seiner Ausnahmeliste ist damit so klein wie möglich: dieses
// Paket, nicht die ganze Fassade. Stünde `app` dort, hätte **jede** künftige
// Zeile in den Bindungen den Weg nach draußen.
//
// **Was dieses Paket nicht tut**: es macht keine Verbindung auf und wartet auf
// keine. Es hat keinen Klienten, keinen Lauscher und keine Adresse; es formt
// Bytes zu einer Antwort. Dass das so bleibt, hält
// [TestDasBlattMachtKeineVerbindungAuf].
//
// # Warum ein Fach und kein zweites Rastern
//
// Der Handler könnte die Karte selbst zeichnen. Dann liefe der Planbau
// **zweimal**: einmal für die Antwort — sie trägt Feldflächen, Vorgaben und
// Befunde, und die kommen aus dem Plan — und einmal für das Bild. 28 ms
// doppelt sind teurer als ein paar Byte im Speicher.
//
// # Warum nur wenige Fächer
//
// Ein Bild ist 182 KB. Acht sind anderthalb Megabyte, und mehr als eine
// Vorschau steht nie gleichzeitig auf dem Bildschirm — die Kartenliste hat
// ihren eigenen, kleineren Weg. Was hinausfällt, ist eine Adresse, die niemand
// mehr abruft.
package vorschauweg

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
)

// Faecher ist die Zahl der Bilder, die auf ihren Abruf warten.
//
// **Acht und nicht eines.** Wer schnell durch die Kartenliste geht, stößt
// mehrere Läufe an; das Fenster zeigt am Ende einen davon, aber die anderen
// Adressen stehen dann schon in einem `<img>`, das gerade lädt. Ein einziges
// Fach machte aus jedem zweiten Bild ein 404.
const Faecher = 8

// Pfad ist der Anfang jeder Bildadresse.
//
// **Mit führendem Schrägstrich**, damit die Adresse gegen den Ursprung des
// Fensters aufgelöst wird und nicht gegen die Seite, auf der jemand gerade
// steht.
const Pfad = "/vorschau/"

// Fach hält die zuletzt gerasterten Bilder.
//
// **Sein Nullwert ist gebrauchsfertig.**
type Fach struct {
	schranke sync.Mutex
	bilder   map[string][]byte

	// reihe ist die Einfügereihenfolge — das älteste fällt zuerst hinaus.
	reihe []string

	// treffer und fehlgriffe zählen die Abholungen — OPS-08 (#1725).
	treffer, fehlgriffe int64
}

// Hinterlege legt ein Bild ab und gibt seine Adresse.
//
// **Dasselbe Bild bekommt dieselbe Adresse** — siehe [marke].
func (f *Fach) Hinterlege(bild []byte) string {
	kennung := marke(bild)

	f.schranke.Lock()
	defer f.schranke.Unlock()

	if f.bilder == nil {
		f.bilder = map[string][]byte{}
	}

	// **Ein zweites Mal dasselbe Bild rückt ans Ende der Reihe** — es steht
	// dort nur **einmal**.
	//
	// Zwei Zusagen in einer Bewegung:
	//
	//  1. **Keine Marke doppelt.** Sonst löschte das Verdrängen einen Eintrag,
	//     den ein späterer Platz noch führt — die Adresse wäre tot, obwohl die
	//     Ablage sie noch zählt.
	//  2. **Wer wieder gezeigt wird, ist jung.** Der erste Anlauf ließ die
	//     Marke stehen, wo sie war: nach `A, B, A` fiel **A** vor B hinaus,
	//     obwohl A die jüngere Ansicht war. Genau das ist das Muster der
	//     Kartenliste — hin und wieder zurück. Angestrichen vom Zweitblick
	//     über die Nacharbeit
	//     (#1278,
	//     Befund 1).
	f.reihe = ohne(f.reihe, kennung)

	// **Eine Kopie in genauer Größe** — derselbe Grund wie im
	// [Texturfach]: `bild` hängt am Puffer, in den gepackt wurde (E06).
	f.bilder[kennung] = bytes.Clone(bild)
	f.reihe = append(f.reihe, kennung)

	for len(f.reihe) > Faecher {
		delete(f.bilder, f.reihe[0])
		f.reihe = f.reihe[1:]
	}

	return Pfad + kennung + ".png"
}

// ohne gibt die Reihe ohne eine Marke.
//
// **An Ort und Stelle und ohne neue Ablage**: die Reihe ist höchstens
// [Faecher] lang, und diese Funktion läuft bei jeder Vorschau.
func ohne(reihe []string, marke string) []string {
	behalten := reihe[:0]

	for _, eine := range reihe {
		if eine != marke {
			behalten = append(behalten, eine)
		}
	}

	return behalten
}

// marke ist die Adresse eines Bildes — **aus seinem Inhalt**.
//
// # Warum nicht eine laufende Zahl
//
// Hier stand eine, und sie war falsch. Die Antwort trug damals
// `Cache-Control: immutable` (seit E06 `no-store`), und die Begründung daneben lautete: „dieselbe
// Marke trägt nie einen anderen Inhalt". **Das galt nur innerhalb eines
// Prozesses.** Mit dem nächsten Start fing der Zähler wieder bei 1 an, und
// `/vorschau/1.png` bedeutete etwas anderes — während der Browser die alte
// Antwort für ewig gültig hielt.
//
// **Im laufenden Fenster gesehen, nicht überlegt** (§3): am 2026-08-28 stand
// im Namensfeld `Fireball` und in der Vorschau `Jamuraan Lion`. Nachgemessen
// gab der Server unter `/vorschau/5.png` `547c405a…`, der Browser zeigte
// `6f59fce8…` — dieselbe Adresse, zwei Bilder.
//
// # Warum der Inhalt und nicht eine Prozesskennung
//
// Eine Kennung je Lauf hätte den Fall auch behoben. Der Inhalt macht die
// Zusage aber **wahr statt wahrscheinlich**: dieselbe Adresse trägt dasselbe
// Bild, über Prozesse, Neustarts und zwei Fächer hinweg — nicht, weil zwei
// Zufallszahlen selten gleich sind, sondern weil die Adresse aus dem Bild
// folgt.
//
// **Und sie spart einen Abruf**: wer zwischen zwei Karten hin- und herwechselt,
// bekommt zweimal dieselbe Adresse, und der Browser hat sie schon.
//
// # Was sie kostet
//
// SHA-256 über 182 KB — auf dieser Maschine unter einer Zehntelmillisekunde,
// gegen 53 ms für den ganzen Weg. Sechzehn Hexstellen sind 64 Bit; für die
// paar Bilder eines Fachs ist ein Zusammenstoß kein Fall, über den man
// nachdenkt.
//
// **Sie ist keine Sicherheitsaussage.** Wer die Adresse errät, bekommt ein
// Bild, das er ohnehin auf dem Bildschirm hat: die Leitung liegt im Fenster
// und hört auf keinem Netz.
func marke(bild []byte) string {
	summe := sha256.Sum256(bild)

	return hex.EncodeToString(summe[:8])
}

// Hole gibt ein abgelegtes Bild; der zweite Wert sagt, ob es noch da ist.
func (f *Fach) Hole(marke string) ([]byte, bool) {
	f.schranke.Lock()
	defer f.schranke.Unlock()

	bild, da := f.bilder[marke]
	if da {
		f.treffer++
	} else {
		f.fehlgriffe++
	}

	return bild, da
}

// Zahl sagt, wie viele Bilder das Fach hält.
//
// **Für den Wächter über die Schranke.** Ohne ihn wäre ein Fach, das unter Last
// nicht mehr verdrängt, von außen nicht zu unterscheiden von einem, das es tut.
func (f *Fach) Zahl() int {
	f.schranke.Lock()
	defer f.schranke.Unlock()

	return len(f.bilder)
}

// Weg liefert die Bilder eines Fachs aus.
//
// **Ein eigener Typ und kein `http.Handler` im Rückgabewert von
// [Neu].** So kann `internal/app` den Weg herausgeben und `main.go` ihn an
// Wails hängen, ohne dass eines von beiden `net/http` einführen müsste — der
// Typ wird nirgends genannt, nur benutzt.
type Weg struct {
	fach *Fach

	// texturen ist das Fach der Spielbilder — **nil heißt: keines**.
	//
	// Die Kommandozeile und der Dienst bauen keinen Spieltisch; ihr Weg trägt
	// deshalb keines, und ein Abruf unter [Spielpfad] endet dort mit 404
	// statt mit einem Sturz.
	texturen *Texturfach

	// tischbilder gibt den eigenen Hintergrund eines Menschen — **nil heißt:
	// keinen** (G3, #1460).
	//
	// Derselbe Grund wie bei `texturen`: wer ohne Fenster arbeitet, hat keinen
	// Tisch, und ein Abruf unter [Tischpfad] endet dort mit 404.
	tischbilder Tischbildquelle

	// fensterbilder gibt das eigene Fensterbild eines Menschen — **nil
	// heißt: keines** (UX-09, #1681).
	fensterbilder Fensterbildquelle
}

// Neu gibt den Weg zu einem Fach.
func Neu(fach *Fach) *Weg {
	return &Weg{fach: fach}
}

// MitTexturen hängt das Fach der Spielbilder daneben (G2, #1459).
func (w *Weg) MitTexturen(texturen *Texturfach) *Weg {
	w.texturen = texturen

	return w
}

// ServeHTTP beantwortet einen Bildabruf.
//
// **Zwei Fächer an einer Leitung** — welches gemeint ist, sagt der Vorsatz der
// Adresse. Ein gemeinsamer zwänge den Weg, beide zu fragen, und ein Schlüssel,
// der in beiden stünde, gäbe zwei Bilder.
func (w *Weg) ServeHTTP(schreiben http.ResponseWriter, anfrage *http.Request) {
	if strings.HasPrefix(anfrage.URL.Path, Spielpfad) {
		w.spielbild(schreiben, anfrage)

		return
	}

	if strings.HasPrefix(anfrage.URL.Path, Tischpfad) {
		w.tischbild(schreiben, anfrage)

		return
	}

	if strings.HasPrefix(anfrage.URL.Path, Fensterpfad) {
		w.fensterbild(schreiben, anfrage)

		return
	}

	bild, da := w.fach.Hole(vorschaumarke(anfrage.URL.Path))
	if !da {
		// **404 und keine leere Antwort.** Ein Bild, dessen Fach schon
		// weitergerückt ist, gehört zu einer Vorschau, die niemand mehr
		// zeigt; ein leeres PNG sähe aus wie eine Karte ohne Inhalt.
		http.NotFound(schreiben, anfrage)

		return
	}

	schreiben.Header().Set("Content-Type", "image/png")

	// **Nicht im Speicher der Webansicht** (E06, #2278). Hier stand
	// `immutable`, weil dieselbe Marke nie einen anderen Inhalt trägt — das
	// gilt weiter, aber jede Bearbeitung gibt eine **neue** Marke, und WebKit
	// behielt jede alte Antwort bis an die Grenze seines Speichers: im
	// Zwei-Stunden-Lauf vom 2026-10-11 wuchs der Webprozess am Ende um 3,5 MB
	// je Minute. Das Bild hält [Fach]; ein zweiter Abruf ist eine Kopie.
	schreiben.Header().Set("Cache-Control", "no-store")

	//nolint:errcheck // ein abgebrochener Abruf ist der Normalfall beim Tippen
	_, _ = schreiben.Write(bild)
}

// spielbild beantwortet den Abruf eines Spielbildes.
func (w *Weg) spielbild(schreiben http.ResponseWriter, anfrage *http.Request) {
	if w.texturen == nil {
		http.NotFound(schreiben, anfrage)

		return
	}

	bild, da := w.texturen.Hole(spielbildmarke(anfrage.URL.Path))
	if !da {
		// **404 und keine leere Antwort** — derselbe Grund wie eine Zeile
		// weiter oben, und hier wiegt er schwerer: ein leeres PNG auf einem
		// Spieltisch sähe aus wie eine Karte ohne Inhalt, und niemand suchte
		// nach der Ursache (#1459: „fehlende Ressource hat sichtbare
		// Diagnose").
		http.NotFound(schreiben, anfrage)

		return
	}

	schreiben.Header().Set("Content-Type", "image/png")

	// **Nicht im Speicher der Webansicht** — derselbe Grund wie bei der
	// Vorschau: eine geänderte Karte bekommt einen neuen Schlüssel, und das
	// Bild hält [Texturfach] (E06, #2278).
	schreiben.Header().Set("Cache-Control", "no-store")

	//nolint:errcheck // ein abgebrochener Abruf ist der Normalfall am Tisch
	_, _ = schreiben.Write(bild)
}

// Vor setzt den Bildweg vor einen anderen Weg — die Weiche vor dem
// Assetserver von Wails 3 (ENG-02,
// #1727).
//
// # Warum eine Weiche und kein Rückfall
//
// **Unter Wails 2 fragte der Assetserver zuerst die eingebettete Oberfläche**
// und reichte, was dort nicht lag, an diesen Weg. Wails 3 hat dafür keinen
// Rückfall, sondern eine Middleware vor seinem eigenen Weg; die Weiche sieht
// deshalb auf den Vorsatz. **Und sie gilt auch beim Entwickeln**: dort reicht
// Wails jede Anfrage an Vite weiter, und Vite beantwortete einen unbekannten
// Weg mit `index.html` — die Weiche steht **vor** diesem Weiterreichen.
//
// # Warum hier und nicht in der Hülle
//
// **Weil die Vorsätze hier stehen** und `frontend/vite.config.ts` sie von
// hier liest — eine zweite Liste liefe weg. Und weil ein Weg über HTTP ohne
// `net/http` nicht zu schreiben ist: stünde die Weiche in `internal/fenster`,
// hätte die Hülle den Import, den der Netzwächter (#54) aus gutem Grund
// abweist. Als Methodenwert passt sie in `application.Middleware`, ohne dass
// die Hülle den Typ nennen muss.
func (w *Weg) Vor(weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if gehoert(anfrage.URL.Path) {
			w.ServeHTTP(schreiben, anfrage)

			return
		}

		weiter.ServeHTTP(schreiben, anfrage)
	})
}

// gehoert sagt, ob ein Weg dem Bildweg gehört — einer der drei Vorsätze.
func gehoert(weg string) bool {
	for _, vorsatz := range []string{Pfad, Spielpfad, Tischpfad, Fensterpfad} {
		if strings.HasPrefix(weg, vorsatz) {
			return true
		}
	}

	return false
}
