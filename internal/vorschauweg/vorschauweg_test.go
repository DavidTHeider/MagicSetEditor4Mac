package vorschauweg_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Die Adresse führt zu den Bytes, die abgelegt wurden.**
func TestEineAbgelegteAdresseFuehrtZumBild(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	adresse := fach.Hinterlege([]byte("erstes Bild"))

	if !strings.HasPrefix(adresse, vorschauweg.Pfad) {
		t.Fatalf("Adresse %q liegt nicht unter %q", adresse, vorschauweg.Pfad)
	}

	if !strings.HasSuffix(adresse, ".png") {
		t.Errorf("Adresse %q endet nicht auf .png — ein Browser rät sonst am "+
			"Inhalt herum", adresse)
	}

	schreiben := hole(t, &fach, adresse)

	if schreiben.Code != http.StatusOK {
		t.Fatalf("Code %d", schreiben.Code)
	}

	if schreiben.Body.String() != "erstes Bild" {
		t.Errorf("abgeholt %q", schreiben.Body.String())
	}

	if art := schreiben.Header().Get("Content-Type"); art != "image/png" {
		t.Errorf("Content-Type %q — ein Browser zeigt das nicht als Bild", art)
	}

	// **Nicht im Speicher der Webansicht** (E06, #2278): die Marke ist je
	// Bild eine andere, und mit `immutable` behielt WebKit jede Antwort —
	// im Zwei-Stunden-Lauf wuchs der Webprozess dabei um 3,5 MB je Minute.
	// Das Bild hält das Fach in Go; ein zweiter Abruf ist eine Kopie.
	if frist := schreiben.Header().Get("Cache-Control"); frist != "no-store" {
		t.Errorf("Cache-Control %q, erwartet no-store", frist)
	}
}

// **Zwei Bilder bekommen zwei Adressen.**
//
// Ohne diese Zusage zeigte das Fenster beim schnellen Durchgehen der
// Kartenliste die vorletzte Karte: der Browser nähme für dieselbe Adresse
// seinen zwischengespeicherten Stand — erst recht mit `immutable` darüber.
func TestJedesBildBekommtEineEigeneAdresse(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	erste := fach.Hinterlege([]byte("A"))
	zweite := fach.Hinterlege([]byte("B"))

	if erste == zweite {
		t.Fatalf("beide unter %q — dann zeigt der Browser das alte", erste)
	}

	if inhalt := hole(t, &fach, erste).Body.String(); inhalt != "A" {
		t.Errorf("unter der ersten Adresse steht %q", inhalt)
	}

	if inhalt := hole(t, &fach, zweite).Body.String(); inhalt != "B" {
		t.Errorf("unter der zweiten Adresse steht %q", inhalt)
	}
}

// **Eine Adresse, deren Fach weitergerückt ist, gibt 404.**
//
// Nicht ein leeres PNG: das sähe aus wie eine Karte ohne Inhalt, und niemand
// unterschiede es von einem Rasterfehler.
func TestEineVerdraengteAdresseGibtVierNullVier(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	erste := fach.Hinterlege([]byte("A"))

	// **Ein Bild mehr, als das Fach hält.** [vorschauweg.Faecher] ist die
	// Zusage; eine feste Zahl hier ginge auseinander, sobald sie sich ändert.
	//
	// **Jedes mit eigenem Inhalt**, denn die Adresse folgt ihm: achtmal
	// dieselben Bytes wären **eine** Ablage und verdrängten nichts.
	var letzte string

	for i := range vorschauweg.Faecher {
		letzte = fach.Hinterlege([]byte{byte(i), 'x'})
	}

	if code := hole(t, &fach, erste).Code; code != http.StatusNotFound {
		t.Errorf("Code %d für eine verdrängte Adresse — erwartet 404", code)
	}

	// **Und das jüngste ist noch da.** Ohne diese Hälfte wäre der Lauf auch
	// grün, wenn das Fach **alles** vergäße.
	if code := hole(t, &fach, letzte).Code; code != http.StatusOK {
		t.Errorf("Code %d für das zuletzt abgelegte Bild", code)
	}

	if fach.Zahl() != vorschauweg.Faecher {
		t.Errorf("%d Bilder im Fach, zugesagt sind %d",
			fach.Zahl(), vorschauweg.Faecher)
	}
}

// **Ein Weg, der nie etwas gesehen hat, gibt 404** — der Nullwert trägt.
func TestEinLeeresFachGibtVierNullVier(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	if code := hole(t, &fach, vorschauweg.Pfad+"1.png").Code; code != http.StatusNotFound {
		t.Errorf("Code %d aus einem leeren Fach", code)
	}
}

// **Und ein Weg außerhalb des Pfads ebenso.**
//
// Wails reicht diesem Weg **jede** Anfrage weiter, die die eingebettete
// Oberfläche nicht kennt — nicht nur die unter [vorschauweg.Pfad].
func TestEinFremderWegGibtVierNullVier(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	fach.Hinterlege([]byte("A"))

	if code := hole(t, &fach, "/etwas/anderes").Code; code != http.StatusNotFound {
		t.Errorf("Code %d für einen fremden Weg", code)
	}
}

// **Das Fach verträgt Ablegen und Holen zugleich.**
//
// Wails gibt jedem Bindungsaufruf eine eigene Goroutine, und der Abruf läuft
// auf der Goroutine der Webansicht — nebenher zum nächsten Rasterlauf.
//
// **Nur unter `-race` sagt dieser Lauf etwas.** `scripts/ci.sh test` fährt mit;
// ohne ihn ginge auch ein Fach ohne Schranke durch.
func TestDasFachVertraegtGleichzeitigesAblegenUndHolen(t *testing.T) {
	t.Parallel()

	var (
		fach     vorschauweg.Fach
		weg      = vorschauweg.Neu(&fach)
		zusammen sync.WaitGroup
	)

	for i := range 8 {
		zusammen.Add(1)

		go func() {
			defer zusammen.Done()

			// **Mehr Ablagen als Fächer**, damit das Verdrängen mitläuft: es
			// schreibt die Ablage **und** die Reihe, und ein Leser greift
			// dazwischen.
			for range vorschauweg.Faecher * 2 {
				adresse := fach.Hinterlege([]byte{byte(i)})

				schreiben := httptest.NewRecorder()
				weg.ServeHTTP(schreiben, anfrage(t, adresse))
			}
		}()
	}

	zusammen.Wait()

	if fach.Zahl() > vorschauweg.Faecher {
		t.Errorf("%d Bilder im Fach, zugesagt sind höchstens %d",
			fach.Zahl(), vorschauweg.Faecher)
	}
}

// **Dieselben Bytes bekommen dieselbe Adresse — auch in einem anderen Fach.**
//
// # Warum das die eigentliche Zusage ist
//
// Die Antwort trägt `Cache-Control: immutable`. Das ist nur dann keine Lüge,
// wenn eine Adresse **nie** einen anderen Inhalt trägt — über Prozesse hinweg,
// nicht nur innerhalb eines Laufs.
//
// Hier stand eine laufende Zahl, und mit dem nächsten Start fing sie wieder
// bei 1 an. Im Fenster stand am 2026-08-28 im Namensfeld `Fireball` und in der
// Vorschau `Jamuraan Lion`; nachgemessen gab der Server unter
// `/vorschau/5.png` `547c405a…` und der Browser zeigte `6f59fce8…`.
//
// **Zwei Fächer stehen hier für zwei Läufe.** Was ein Test in einem Prozess
// über Prozessgrenzen sagen kann, ist genau das: die Adresse hängt am Inhalt
// und an nichts sonst — kein Zähler, kein Zeitpunkt, keine Kennung des Fachs.
func TestDieAdresseFolgtDemBild(t *testing.T) {
	t.Parallel()

	var einFach, zweitesFach vorschauweg.Fach

	// **Im zweiten Fach zuerst etwas anderes**, damit ein Zähler auseinander
	// liefe: ohne diese Zeile gäbe auch eine laufende Zahl zweimal dieselbe
	// Adresse, und der Lauf sagte nichts.
	zweitesFach.Hinterlege([]byte("etwas ganz anderes"))

	hier := einFach.Hinterlege([]byte("dasselbe Bild"))
	dort := zweitesFach.Hinterlege([]byte("dasselbe Bild"))

	if hier != dort {
		t.Errorf("dasselbe Bild unter %q und %q — dann trägt eine Adresse "+
			"zwei Inhalte, und `immutable` ist eine Lüge", hier, dort)
	}

	if anderes := einFach.Hinterlege([]byte("ein anderes Bild")); anderes == hier {
		t.Errorf("zwei Bilder unter %q — dann zeigt der Browser das alte", anderes)
	}
}

// **Und dasselbe Bild zweimal abzulegen verbraucht kein zweites Fach.**
//
// Ohne diese Zusage stünde eine Marke mehrfach in der Reihe, und das
// Verdrängen löschte einen Eintrag, den ein späterer Platz noch führt — die
// Adresse wäre tot, obwohl die Ablage sie noch zählt.
func TestDasselbeBildZweimalBrauchtEinFach(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	erste := fach.Hinterlege([]byte("A"))
	zweite := fach.Hinterlege([]byte("A"))

	if erste != zweite {
		t.Fatalf("%q und %q für dasselbe Bild", erste, zweite)
	}

	if fach.Zahl() != 1 {
		t.Errorf("%d Bilder im Fach für ein einziges", fach.Zahl())
	}

	if fach.Zahl() != 1 {
		t.Errorf("%d Bilder im Fach für ein einziges", fach.Zahl())
	}
}

// **Und ein wiedergezeigtes Bild fällt nicht zu früh hinaus.**
//
// # Der Schaden, den dieser Lauf fängt
//
// Der erste Anlauf prüfte nur, dass zweimal dasselbe Bild **ein** Fach belegt.
// Die Gegenprobe — das Verdrängen die Reihe doppelt weiterrücken lassen —
// blieb **grün**, und der Grund war: `len(bilder)` stimmte weiter, nur die
// **Reihe** trug die Marke zweimal.
//
// Sichtbar wird das erst bei dem Muster, das im Fenster der Normalfall ist:
// Karte A ansehen, Karte B, zurück zu A. Die Reihe trüge dann `[a, b, a]`, und
// schon **sechs** weitere Karten schöben A hinaus — obwohl es die drittjüngste
// Ansicht war und ihr `<img>` noch dasteht.
//
// **Eine Gegenprobe, die nichts ändert, ist keine** (§10). Diese hier ändert
// etwas.
func TestEinWiedergezeigtesBildFaelltNichtZuFruehHinaus(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	// **A, B, A** — hin und zurück, wie in der Kartenliste.
	a := fach.Hinterlege([]byte("Karte A"))
	b := fach.Hinterlege([]byte("Karte B"))
	fach.Hinterlege([]byte("Karte A"))

	// **Bis das Fach voll ist — und dann noch eines.**
	//
	// # Der Fehler, den erst die zweite Fassung dieses Laufs fing
	//
	// Der erste Anlauf füllte **genau** bis [vorschauweg.Faecher] und war grün,
	// obwohl A in der Reihe weiter ganz vorne stand: das Wiederablegen gab ihm
	// keinen neuen Platz. Beim **nächsten** Bild fiel A hinaus — vor B, obwohl
	// A jünger war.
	//
	// Angestrichen vom Zweitblick über die Nacharbeit
	// (#1278,
	// Befund 1). „Das Fach ist gerade erst voll" prüfte die Zusage einen
	// Schritt zu früh.
	for i := range vorschauweg.Faecher - 1 {
		fach.Hinterlege([]byte{byte(i), 'z'})
	}

	if code := hole(t, &fach, a).Code; code != http.StatusOK {
		t.Errorf("Code %d für Karte A — sie war die drittjüngste Ansicht und "+
			"muss vor B hinausfallen, nicht davor", code)
	}

	if code := hole(t, &fach, b).Code; code != http.StatusNotFound {
		t.Errorf("Code %d für Karte B — sie war die älteste und musste "+
			"hinausfallen", code)
	}

	if fach.Zahl() != vorschauweg.Faecher {
		t.Errorf("%d Bilder im Fach, erwartet %d", fach.Zahl(), vorschauweg.Faecher)
	}
}

// hole ruft eine Adresse über den echten Weg ab.
func hole(t *testing.T, fach *vorschauweg.Fach, adresse string) *httptest.ResponseRecorder {
	t.Helper()

	schreiben := httptest.NewRecorder()
	vorschauweg.Neu(fach).ServeHTTP(schreiben, anfrage(t, adresse))

	return schreiben
}

func anfrage(t *testing.T, adresse string) *http.Request {
	t.Helper()

	return httptest.NewRequestWithContext(t.Context(), http.MethodGet, adresse, http.NoBody)
}
