package vorschauweg

// **Das Fach der Spielbilder** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum es neben [Fach] steht und nicht an seiner Stelle
//
// **Weil beide etwas anderes zählen.** [Fach] hält [Faecher] Bilder, weil nie
// mehr als eine Vorschau auf dem Bildschirm steht. Ein Spieltisch zeigt
// sechzig Karten, und die Frage ist nicht „wie viele", sondern „wie viel
// Speicher": zwischen einer Miniatur und einem Betrachtungsbild liegt der
// Faktor hundert.
//
// # Und warum der Schlüssel nicht der Inhalt ist
//
// **Weil er vor dem Zeichnen feststehen muss.** [Fach] adressiert über den
// Abdruck der **Bytes** — die hat es, wenn es ablegt. Hier lautet die Frage
// umgekehrt: *muss* überhaupt gezeichnet werden? Ein Schlüssel aus den
// Eingaben beantwortet sie, ein Abdruck der Ausgabe nicht. Woraus er gebaut
// wird, entscheidet `internal/service`; hier steht nur, dass er einer ist.

import (
	"bytes"
	"strings"
	"sync"
)

// Spielpfad ist der Anfang jeder Spielbildadresse.
//
// **Ein eigener und nicht [Pfad]**: dieselbe Leitung, zwei Fächer — und die
// Adresse sagt, welches gemeint ist. Ein gemeinsamer Vorsatz zwänge den Weg,
// beide zu fragen, und ein Schlüssel, der in beiden steht, gäbe zwei Bilder.
const Spielpfad = "/spielbild/"

// Texturfach hält gerasterte Spielbilder unter ihrem Schlüssel.
//
// **Es verdrängt nach Bytes und nach letztem Gebrauch**, nicht nach Zahl und
// nicht nach Einfügereihenfolge — die Gründe stehen an [Texturfach.Hole].
type Texturfach struct {
	schranke sync.Mutex

	budget int
	belegt int

	bilder map[string]spielbild

	// reihe ist die Reihenfolge des letzten Gebrauchs — vorn das Älteste.
	reihe []string

	// treffer und fehlgriffe zählen die Nachfragen — OPS-08 (#1725).
	treffer, fehlgriffe int64
}

// spielbild ist ein abgelegtes Bild samt seiner Maße.
//
// # Warum die Maße hier liegen und nicht beim Fragenden
//
// **Weil sonst ein Treffer anders antwortet als ein Fehltreffer.** Wer
// zeichnet, kennt die Maße des Bildes, das dabei entstand; wer im Fach
// fündig wird, kennt sie nicht und müsste sie **ausrechnen** — und eine
// gerechnete Zahl und eine gemessene gehen um einen Bildpunkt auseinander.
// Gemessen am 2026-09-15: derselbe Rücken meldete beim ersten Ruf 512 × 512
// und beim zweiten 512 × 714. Das Issue verlangt „unveränderter Inhalt
// trifft ihn", und ein Treffer, der andere Zahlen nennt, ist keiner.
type spielbild struct {
	bytes []byte

	breite int
	hoehe  int
}

// NeuesTexturfach baut eines mit einem Budget in Bytes.
func NeuesTexturfach(budget int) *Texturfach {
	return &Texturfach{budget: budget, bilder: map[string]spielbild{}}
}

// Hinterlege legt ein Bild unter seinem Schlüssel ab und gibt seine Adresse.
//
// **Auch ein zu großes bekommt eine Adresse.** Sonst stünde der Aufrufer vor
// der Frage, ob er noch einmal zeichnen soll, und zeichnete jedes Mal. Die
// Adresse zeigt dann ins Leere, und [Weg] antwortet mit 404 — eine sichtbare
// Diagnose statt eines leeren Bildes (#1459).
func (t *Texturfach) Hinterlege(
	schluessel string, bild []byte, breite, hoehe int,
) string {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	t.entferne(schluessel)

	if len(bild) <= t.budget {
		// **Eine Kopie in genauer Größe** (E06, #2278): `bild` ist ein Stück
		// des Puffers, in den gepackt wurde, und dessen Array ist beim
		// Wachsen verdoppelt — gezählt wurde `len`, gehalten `cap`, bis zum
		// Doppelten des Budgets.
		t.bilder[schluessel] = spielbild{
			bytes: bytes.Clone(bild), breite: breite, hoehe: hoehe,
		}
		t.belegt += len(bild)
		t.reihe = append(t.reihe, schluessel)

		t.raeumeAuf()
	}

	return Spielpfad + schluessel + ".png"
}

// Adresse gibt die Adresse zu einem Schlüssel — und ob es sie gibt.
//
// **Ohne den Gebrauch zu zählen.** Sie beantwortet „liegt es schon da?", und
// wer fragt, benutzt es gleich darauf über [Texturfach.Hole]; ein zweites
// Hochrücken machte aus einer Frage eine Benutzung.
func (t *Texturfach) Adresse(schluessel string) (
	adresse string, breite, hoehe int, da bool,
) {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	bild, da := t.bilder[schluessel]
	if !da {
		t.fehlgriffe++

		return "", 0, 0, false
	}

	t.treffer++

	return Spielpfad + schluessel + ".png", bild.breite, bild.hoehe, true
}

// Hole gibt ein abgelegtes Bild und rückt es ans Ende der Reihe.
//
// # Warum das Holen zählt und nicht nur das Hinterlegen
//
// **Weil ein Spieltisch immer wieder dieselben Karten zeigt.** Ein Bild, das
// im Fach liegt, wird **nicht noch einmal hinterlegt** — es wird geholt. Wer
// nur das Ablegen zählte, verdrängte genau die Bilder, die am häufigsten
// gebraucht werden. Dieselbe Lehre hat [Fach] schon einmal gekostet
// (#1278,
// Befund 1), dort für das zweite Ablegen.
func (t *Texturfach) Hole(schluessel string) ([]byte, bool) {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	bild, da := t.bilder[schluessel]
	if !da {
		return nil, false
	}

	t.reihe = ohne(t.reihe, schluessel)
	t.reihe = append(t.reihe, schluessel)

	return bild.bytes, true
}

// Belegt sagt, wie viele Bytes das Fach hält.
//
// **Für den Wächter über das Budget** — ohne ihn wäre ein Fach, das nicht mehr
// verdrängt, von außen nicht von einem zu unterscheiden, das es tut.
func (t *Texturfach) Belegt() int {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	return t.belegt
}

// Raeume wirft alles weg.
//
// **Für den Profilwechsel** (#1459): wer wechselt, sieht die Karten eines
// anderen — und ein Fach, das sie noch hielte, gäbe sie unter ihrer Adresse
// weiter heraus.
func (t *Texturfach) Raeume() {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	t.bilder = map[string]spielbild{}
	t.reihe = nil
	t.belegt = 0
}

// Vergiss nimmt ein Bild heraus, auf das keine Karte mehr zeigt (E06,
// #2278) — der Dienst weiß das, das Fach nicht: es zählt nach Bytes, nicht
// nach Verweisen.
func (t *Texturfach) Vergiss(schluessel string) {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	t.entferne(schluessel)
}

// entferne nimmt einen Schlüssel heraus — **unter der Schranke**.
func (t *Texturfach) entferne(schluessel string) {
	if bild, da := t.bilder[schluessel]; da {
		t.belegt -= len(bild.bytes)

		delete(t.bilder, schluessel)
	}

	t.reihe = ohne(t.reihe, schluessel)
}

// raeumeAuf verdrängt, bis das Budget wieder passt — **unter der Schranke**.
func (t *Texturfach) raeumeAuf() {
	for t.belegt > t.budget && len(t.reihe) > 0 {
		aeltestes := t.reihe[0]

		t.belegt -= len(t.bilder[aeltestes].bytes)

		delete(t.bilder, aeltestes)

		t.reihe = t.reihe[1:]
	}
}

// spielbildmarke gibt den Schlüssel aus einer Spielbildadresse.
func spielbildmarke(pfad string) string {
	return strings.TrimSuffix(strings.TrimPrefix(pfad, Spielpfad), ".png")
}
