// Die Zeilengestalten, mit denen der Stolpertreiber antwortet.
//
// **Getrennt vom Treiber, weil es eine andere Frage ist.** Der Treiber sagt,
// *wo* gestolpert wird; diese Typen sagen, *was* zurückkommt, solange er es
// nicht tut. Eine Antwort muss zur Frage passen — der Verlauf liest den neuen
// Wert mit **einer** Spalte, die Feldänderung ihre Vorprüfung mit **zwei**, der
// Verlauf seinen Eintrag mit **acht**. Ein Treiber, der immer dasselbe liefert,
// scheitert am Scan statt an der Stelle, die geprüft werden soll.
package storage

import (
	"database/sql/driver"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kennungszeilen gibt eine Zeile mit der Kennung 1 — genug, damit eine
// Vorprüfung eine Karte für vorhanden hält.
type kennungszeilen struct{ gegeben bool }

func (r *kennungszeilen) Columns() []string { return []string{"id"} }
func (r *kennungszeilen) Close() error      { return nil }

func (r *kennungszeilen) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = int64(1)

	return nil
}

// verlaufszeile gibt einen Verlaufseintrag, wie `naechster` ihn abfragt.
//
// Neun Spalten, und die beiden Wege sind gültiges JSON — sonst käme ein Lauf
// nie bis zum Schreiben, weil er vorher am Zerlegen scheiterte.
//
// **Die neunte kam mit der Herkunft** (#888). Sie ist der Grund, aus dem
// `verlaufsspalten` an **einer** Stelle steht: eine vergessene Spalte fällt
// erst beim Scan auf — und hier fiel sie nicht als Fehler auf, sondern als
// sechs plötzlich unbetretene Fehlerzweige in `verlaufbewegen.go`. Wer diese
// Liste ändert, ändert sie hier mit.
type verlaufszeile struct{ gegeben bool }

// **Zehn Spalten, seit der Verlauf seinen Zeitpunkt trägt** (U6, #1417).
// Mit neun scheiterte jede Bewegung schon beim Lesen — und die Stolpertests
// „beim Schreiben" und „beim Bestätigen" waren grün aus dem falschen Grund
// (§10): ein Fehler ist ein Fehler, aber nicht der, um den es ging.
func (r *verlaufszeile) Columns() []string {
	return []string{
		"id", "karte_id", "handlung", "hinweg", "rueckweg",
		"zurueckgenommen", "urheber", "gruppe", "herkunft", "zeitpunkt",
	}
}

func (r *verlaufszeile) Close() error { return nil }

func (r *verlaufszeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = int64(1)
	werte[1] = int64(1)
	werte[2] = "feld"
	// **Der Hinweg ohne Stelle, der Rückweg mit** — und das ist Absicht.
	//
	// `wegGehen` hat seit #160 zwei Schreibwege: anhängen, wenn der Eintrag
	// keine Stelle führt (so sahen alle Zeilen vor #160 aus), und an die
	// Stelle setzen, wenn er eine führt. `TestDasBewegenMeldetJedenWeg` geht
	// beide Richtungen; mit dieser Aufteilung betritt es damit **beide**
	// Schreibwege. Trüge sie beidseitig dasselbe, bliebe einer der beiden
	// Fehlerzweige unbetreten — und niemand sähe es.
	werte[3] = `{"feld":"name","wert":"neu","gabEs":true}`
	werte[4] = `{"feld":"name","wert":"alt","gabEs":true,"reihenfolge":2}`

	if verlaufOhneFeld {
		werte[4] = `{"feld":"name","gabEs":false}`
	}

	werte[5] = int64(0)
	werte[6] = string(domain.UrheberFenster)
	werte[7] = ""
	werte[8] = ""

	werte[9] = "2026-09-09T02:00:00Z"

	return nil
}

// einspaltigeZeile gibt eine Zeile mit einem Wert.
type einspaltigeZeile struct{ gegeben bool }

func (r *einspaltigeZeile) Columns() []string { return []string{"wert"} }
func (r *einspaltigeZeile) Close() error      { return nil }

func (r *einspaltigeZeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = "neu"

	return nil
}

// nachwertzeile gibt den neuen Wert und seine Stelle.
//
// Die Gestalt, die `verlaufen` abfragt: `SELECT wert, reihenfolge FROM
// card_values …`. Sie war einspaltig, bis der Verlauf die Reihenfolge
// mitführte (#160).
type nachwertzeile struct{ gegeben bool }

func (r *nachwertzeile) Columns() []string { return []string{"wert", "reihenfolge"} }
func (r *nachwertzeile) Close() error      { return nil }

func (r *nachwertzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = "neu"
	werte[1] = int64(0)

	return nil
}

// vorwertzeile gibt Kennung, alten Wert und alte Stelle.
//
// Die Gestalt, die `feldaenderung.go` abfragt: `(SELECT 1 FROM cards …),
// (SELECT wert FROM card_values …), (SELECT reihenfolge FROM card_values …)`.
// Damit kommt ein Lauf bis zum Ändern — und erst dort stolpert er.
//
// **Sie hieß `zweispaltigeZeile` und war zwei Spalten breit**, bis der Verlauf
// die Reihenfolge mitführte (#160). Der Lauf scheiterte danach am **Scan**
// statt an der Stelle, um die es ging, und die Fehlerwegtests waren grün, ohne
// ihren Fall auszuführen — zum dritten Mal in diesem Repository. Wer eine
// Abfrage um eine Spalte erweitert, zieht diese Zeile mit; der Name nennt
// deshalb ihre Aufgabe und nicht mehr ihre Breite.
type vorwertzeile struct{ gegeben bool }

func (r *vorwertzeile) Columns() []string { return []string{"da", "wert", "reihenfolge"} }
func (r *vorwertzeile) Close() error      { return nil }

func (r *vorwertzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = int64(1)
	werte[1] = "alter Wert"
	werte[2] = int64(0)

	return nil
}

// krummezeile liefert **eine** Zeile, deren Werte nicht in die Zielfelder
// passen.
//
// Damit lässt sich der Zweig prüfen, den ein `Scan` in einer Leseschleife hat:
// die Zeile ist da, aber unbrauchbar. Mit einer echten SQLite-Datenbank ist das
// nicht herzustellen — `STRICT` sorgt dafür, dass alles hineinpasst, was
// hineinkam. Im Betrieb entsteht der Fall an einem Schema, das jemand von Hand
// verändert hat.
type krummezeile struct{ gegeben bool }

func (r *krummezeile) Columns() []string {
	return []string{"a", "b", "c", "d", "e", "f", "g", "h"}
}

func (r *krummezeile) Close() error { return nil }

func (r *krummezeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	// **NULL, nicht „keine Zahl".** Eine Zeichenkette scheitert nur an einem
	// Zahlfeld; `database/sql` macht aus jedem anderen Wert klaglos eine
	// Zeichenkette. Ein Scan über acht Zeichenkettenfelder wäre damit geglückt,
	// und der Fehlerzweig bliebe ungeprüft. NULL scheitert an jedem Zielfeld.
	//
	// **Acht, weil der Verlauf acht liest.** Es waren sechs, solange er sechs
	// las; `urheber` und `gruppe` (#52) hätten diese Zeile fortan an der
	// **Anzahl** scheitern lassen — derselbe Zweig aus einem anderen Grund, und
	// die Umwandlung wäre ungeprüft geblieben, ohne dass ein Lauf rot wird. Wer
	// eine Abfrage um eine Spalte erweitert, zieht diese Zeile mit.
	//
	// Wo eine Abfrage **weniger** als acht Spalten liest, scheitert der Scan
	// weiterhin an der Anzahl. Das gehört dazugesagt, statt es als Prüfung der
	// Umwandlung auszugeben.
	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// stolperrows liefert keine Zeile und meldet danach einen Fehler — genau der
// Fall, den `zeilen.Err()` abfangen soll.
type stolperrows struct{ leer bool }

func (r *stolperrows) Columns() []string { return []string{"name"} }
func (r *stolperrows) Close() error      { return nil }

func (r *stolperrows) Next([]driver.Value) error {
	if r.leer {
		return io.EOF
	}

	return errGestolpert
}

// zaehlfehler ist ein Ergebnis, dessen Zählung scheitert.
//
// **Für den einen Zweig, den ein echter Treiber nicht hergibt**: `modernc`
// zählt immer, und `StilSetzen` entscheidet an dieser Zahl, ob es das Set
// überhaupt gibt. Ein Fehlerpfad, den kein Test betritt, ist eine Behauptung
// darüber, was im Ernstfall geschieht (§3).
type zaehlfehler struct{}

func (zaehlfehler) LastInsertId() (int64, error) { return 0, errGestolpert }
func (zaehlfehler) RowsAffected() (int64, error) { return 0, errGestolpert }

// kartenstellezeile beantwortet `SELECT reihenfolge FROM cards` — **eine**
// Spalte (#621).
//
// **Ohne sie bekäme die Abfrage die achtspaltige Verlaufszeile** und scheiterte
// am Scan statt an der Stelle, um die es geht — genau die Falle, vor der die
// Kommentare in `stolperabfragen_test.go` warnen.
type kartenstellezeile struct{ gegeben bool }

func (r *kartenstellezeile) Columns() []string { return []string{"reihenfolge"} }
func (r *kartenstellezeile) Close() error      { return nil }

func (r *kartenstellezeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = int64(0)

	return nil
}

// kartenanlegezeile beantwortet `INSERT INTO cards … RETURNING id, reihenfolge`
// — **zwei** Spalten (#625). Ohne sie bekäme das Kopieren die einspaltige
// Stellenzeile und scheiterte am Scan statt an der Stelle, um die es geht.
type kartenanlegezeile struct{ gegeben bool }

func (r *kartenanlegezeile) Columns() []string { return []string{"id", "reihenfolge"} }
func (r *kartenanlegezeile) Close() error      { return nil }

func (r *kartenanlegezeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0], werte[1] = int64(2), int64(1)

	return nil
}

// kartenfeldzeilen beantwortet `SELECT feld, wert, reihenfolge FROM card_values`
// — **drei** Spalten.
type kartenfeldzeilen struct{ gegeben bool }

func (r *kartenfeldzeilen) Columns() []string {
	return []string{"feld", "wert", "reihenfolge"}
}
func (r *kartenfeldzeilen) Close() error { return nil }

func (r *kartenfeldzeilen) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = "name"
	werte[1] = "Probe"
	werte[2] = int64(0)

	return nil
}

// krummesKartenfeld gibt drei Spalten mit dem falschen Typ — damit scheitert
// der `Scan` **in** der Schleife, nicht die Abfrage davor.
type krummesKartenfeld struct{ gegeben bool }

func (r *krummesKartenfeld) Columns() []string {
	return []string{"feld", "wert", "reihenfolge"}
}
func (r *krummesKartenfeld) Close() error { return nil }

func (r *krummesKartenfeld) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = "name"
	werte[1] = "Probe"

	// **Eine Zeichenkette, wo eine Zahl erwartet wird.** `Scan` in `int64`
	// scheitert daran, und genau der Zweig ist gemeint.
	werte[2] = "keine Zahl"

	return nil
}

// stolperndesKartenende meldet nach einer vollständigen Schleife einen Fehler
// — der Fall, den `rows.Err()` fängt und den man nie sieht.
type stolperndesKartenende struct{}

func (r *stolperndesKartenende) Columns() []string {
	return []string{"feld", "wert", "reihenfolge"}
}
func (r *stolperndesKartenende) Close() error { return nil }

func (r *stolperndesKartenende) Next([]driver.Value) error { return errGestolpert }

// stellenzeilen beantwortet `SELECT id, set_id, reihenfolge FROM cards`
// — **zwei** Karten desselben Sets (#628).
type stellenzeilen struct{ gegeben int }

func (r *stellenzeilen) Columns() []string { return []string{"id", "set_id", "reihenfolge"} }
func (r *stellenzeilen) Close() error      { return nil }

func (r *stellenzeilen) Next(werte []driver.Value) error {
	if r.gegeben >= 2 {
		return io.EOF
	}

	werte[0] = int64(r.gegeben + 1)
	werte[1] = int64(7)
	werte[2] = int64(r.gegeben)
	r.gegeben++

	return nil
}

// krummeStellenzeile gibt eine Zeile, die sich nicht scannen lässt.
type krummeStellenzeile struct{ gegeben bool }

func (r *krummeStellenzeile) Columns() []string { return []string{"id", "set_id", "reihenfolge"} }
func (r *krummeStellenzeile) Close() error      { return nil }

func (r *krummeStellenzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0], werte[1], werte[2] = "keine Zahl", int64(7), int64(0)

	return nil
}

// stolperndesStellenende scheitert erst am Ende der Zeilen.
type stolperndesStellenende struct{ gegeben bool }

func (r *stolperndesStellenende) Columns() []string {
	return []string{"id", "set_id", "reihenfolge"}
}
func (r *stolperndesStellenende) Close() error { return nil }

func (r *stolperndesStellenende) Next(werte []driver.Value) error {
	if r.gegeben {
		return errGestolpert
	}

	r.gegeben = true
	werte[0], werte[1], werte[2] = int64(1), int64(7), int64(0)

	return nil
}
