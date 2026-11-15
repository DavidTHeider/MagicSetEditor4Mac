package msepaket

import (
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	// Die Formate, die in MSE-Paketen vorkommen. **JPEG gehört dazu**:
	// `magic-m15.mse-style` setzt seinen Kartengrund über `card.jpg`, und wer
	// nur PNG entschlüsselte, zeichnete bei jeder M15-Karte den Hintergrund
	// nicht. GIF steht daneben, weil `image` es ohnehin mitbringt und ältere
	// Pakete es führen.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// ErrKeinBild meldet ein Bild, das in keiner der Wurzeln liegt.
var ErrKeinBild = errors.New("das Bild ist in keinem Paket zu finden")

// Bildquelle liest Bilder aus Paketen, die auf der Platte liegen.
//
// # Warum sie hier steht und nicht in `render`
//
// `render` sagt, **was** es braucht (`Bildquelle`), und weiß nicht, wo Pakete
// liegen — dieselbe Richtung wie `storage` zu `service` (AGENTS.md §2). Wegen
// Gos struktureller Typisierung erfüllt dieser Typ die Schnittstelle, ohne
// `render` zu importieren.
//
// # Was sie vorhält, und warum
//
// **Den entschlüsselten Stand**, nicht die Rohbytes: das Entschlüsseln ist der
// teure Teil. Gemessen an einem echten Kartenrahmen aus `M15-Magic-Pack`
// (`bildmessung_test.go`, hinter `fremdpakete`): **13 ms je Lauf**, 1,18 s für
// neunzig Karten — und das für **ein** Bild. Eine Karte trägt ein Dutzend
// Ebenen (#187).
//
// # Wann er ungültig wird
//
// **Ein Fehlschlag wird nicht gemerkt.** Wer ein Paket während der Sitzung
// bezieht, findet es beim nächsten Versuch — dieselbe Antwort, die
// `schriften.paketdatei` seit #140 gibt. Ein gemerkter Fehlschlag machte daraus
// einen, der bis zum Neustart gilt.
//
// **Ein Treffer bleibt dagegen stehen, bis der Speicher räumt.** Das ist der
// Preis, und er gehört ausgesprochen: wer eine Vorlagendatei bearbeitet,
// während die Anwendung läuft, sieht die Änderung nicht. Vertretbar ist das,
// weil Paketdateien **Vorlagen** sind und keine Arbeitsdateien — sie werden
// bezogen, nicht getippt. Wer doch eine ändert, startet neu.
//
// # Die Obergrenze
//
// Sie räumt **ganz**, nicht klug. Eine Verdrängung nach Alter verlangt
// Buchführung über jeden Zugriff, und der Nutzen wäre einer für einen Fall, den
// es nicht gibt: ein Druckbogen benutzt ein gutes Dutzend Bilder immer wieder.
// Wer die Grenze erreicht, hat keinen Arbeitssatz mehr, der hineinpasst — dann
// verschöbe eine Verdrängungsstrategie das Problem, statt es zu lösen.
type Bildquelle struct {
	// treffer und fehlgriffe zählen die Abfragen — OPS-08 (#1725).
	treffer, fehlgriffe atomic.Int64

	wurzeln []string

	// sperre schützt `gemerkt`. **Nicht `sync.Map`**: die lohnt sich bei
	// vielen Schreibern auf verschiedene Schlüssel, und hier schreibt jeder
	// Lauf dieselben paar Bilder.
	sperre  sync.RWMutex
	gemerkt map[string]image.Image

	// belegt und grenzeBytes zählen, was die gemerkten Bilder belegen —
	// PERF03 (#2353, `bildbelegung.go`).
	belegt, grenzeBytes int64
}

// GrenzeBildspeicher ist die Zahl der Bilder, die eine Quelle vorhält.
//
// **512, und die Zahl hat einen Grund.** Hier standen 64 — gut 50 MB bei
// voller Grenze, als die Zahl die einzige Grenze war. Seit PERF03 hält die
// **Bytegrenze** den Speicher (`bildbelegung.go`), und 64 räumte zu früh:
// über alle 3.788 Karten von `Black` sammelten sich 255 verschiedene Bilder,
// und die Prüfung eines ganzen Sets ging 28.837 Mal auf die Platte (PERF02,
// #2352,
// gemessen am 2026-10-15). Die Zahl bleibt eine zweite Grenze gegen sehr
// viele sehr kleine Bilder.
const GrenzeBildspeicher = 512

// GrenzeBildbytes ist, wie viel die gemerkten Bilder belegen dürfen — **128
// MiB**, PERF03 (#2353). Die vierundsechzig Rahmen von oben belegen rund 50
// MB; die Grenze greift also erst, wo ein Paket ungewöhnlich große Bilder
// bringt, und dann räumt sie ganz, wie an der Zahl.
const GrenzeBildbytes = 128 << 20

// NeueBildquelle baut die Quelle über einer Reihe von Paketwurzeln.
//
// **Die Reihenfolge ist die Aussage.** Wer ein Paket bezogen hat, meint das
// bezogene; die mitgelieferte Saat kommt danach.
//
// **Eine leere Wurzel fällt weg.** `app.Paketverzeichnis()` gibt einen leeren
// Pfad, solange niemand ein Paket bezogen hat, und `filepath.Join("", …)`
// machte daraus einen **relativen** Pfad: gesucht würde im Arbeitsverzeichnis
// des Prozesses. Dieselbe Falle hat der Zweitblick über Phase 9 schon beim
// Paketbezieher gefunden — hier wird sie gleich hier abgeräumt, statt sie jedem
// Aufrufer zu überlassen.
func NeueBildquelle(wurzeln ...string) *Bildquelle {
	return &Bildquelle{
		wurzeln: echteWurzeln(wurzeln),
		gemerkt: make(map[string]image.Image, GrenzeBildspeicher),

		grenzeBytes: GrenzeBildbytes,
	}
}

// echteWurzeln wirft leere Angaben weg — der Grund steht bei [NeueBildquelle].
func echteWurzeln(wurzeln []string) []string {
	echte := make([]string, 0, len(wurzeln))

	for _, wurzel := range wurzeln {
		if strings.TrimSpace(wurzel) != "" {
			echte = append(echte, wurzel)
		}
	}

	return echte
}

// ErrKeinPaket meldet eine Datei, die in keiner der Wurzeln liegt.
var ErrKeinPaket = errors.New("die Datei ist in keinem Paket zu finden")

// verbinde fasst die Fehlschläge aller Wurzeln zusammen.
func verbinde(versuche []error) error { return errors.Join(versuche...) }

// Bild liest das Bild zu einem Paketpfad.
//
// Der Verweis ist ein **Paketpfad** (`/probe-bild.mse-style/rahmen.png`), kein
// Dateisystempfad — er sagt Paket und Datei, und wo das Paket liegt, entscheidet
// erst diese Quelle.
func (b *Bildquelle) Bild(verweis string) (image.Image, error) {
	// **Das Set-Paket zuerst**, denn sein Ort steht im Verweis und wird nicht
	// gesucht ([SetpaketVerweis]). Ein Bild der Karte liegt dort und nirgends
	// sonst; die Wurzeln danach zu befragen fände im günstigen Fall nichts und
	// im ungünstigen ein gleichnamiges Bild aus einer fremden Vorlage.
	if ort, datei, ist := ZerlegeSetpaketverweis(verweis); ist {
		return b.ausSetpaket(verweis, ort, datei)
	}

	pfad, err := ZerlegePaketpfad(verweis)
	if err != nil {
		return nil, err
	}

	if gemerkt, da := b.nachschlagen(verweis); da {
		b.treffer.Add(1)

		return gemerkt, nil
	}

	b.fehlgriffe.Add(1)

	// **Jeder Fehlschlag wird aufgehoben, nicht verworfen.** „Nicht gefunden"
	// und „eine der Wurzeln ist unlesbar" sind zwei Lagen mit zwei Abhilfen;
	// wer nur die erste meldete, schickte jemanden auf die falsche Suche.
	versuche := make([]error, 0, len(b.wurzeln)+1)
	versuche = append(versuche, ErrKeinBild)

	for _, wurzel := range b.wurzeln {
		rohdaten, fehler := paketdatei(filepath.Join(wurzel, pfad.Paket), pfad.Datei)
		if fehler != nil {
			versuche = append(versuche, fehler)

			continue
		}

		bild, _, fehler := DekodiereBild(rohdaten)
		if fehler != nil {
			// **Hier wird nicht weitergesucht.** Die Datei ist da und kaputt;
			// eine gleichnamige aus der nächsten Wurzel zu nehmen, verdeckte
			// genau den Befund, den jemand beheben muss.
			return nil, fmt.Errorf("%s entschlüsseln: %w", verweis, fehler)
		}

		b.merken(verweis, bild)

		return bild, nil
	}

	// **Der Fehlschlag wird nicht gemerkt** — der Grund steht am Typ.
	return nil, fmt.Errorf("%s: %w", verweis, errors.Join(versuche...))
}

// nachschlagen sucht ein Bild im Speicher.
func (b *Bildquelle) nachschlagen(verweis string) (image.Image, bool) {
	b.sperre.RLock()
	defer b.sperre.RUnlock()

	bild, da := b.gemerkt[verweis]

	return bild, da
}

// merken legt ein Bild ab und räumt, wenn eine Grenze erreicht ist — die
// Zahl oder die Bytes.
//
// **Ein Bild über der ganzen Bytegrenze wird nicht gemerkt**: es räumte sonst
// alles, was hineinpasst, und läge danach allein über der Grenze.
func (b *Bildquelle) merken(verweis string, bild image.Image) {
	groesse := belegteBytes(bild)

	b.sperre.Lock()
	defer b.sperre.Unlock()

	if groesse > b.grenzeBytes {
		return
	}

	// **Zweimal dasselbe zählt einmal** — zwei Läufe, die es zugleich nicht
	// fanden, merken es beide.
	if vorher, da := b.gemerkt[verweis]; da {
		b.belegt -= belegteBytes(vorher)
	}

	if len(b.gemerkt) >= GrenzeBildspeicher || b.belegt+groesse > b.grenzeBytes {
		clear(b.gemerkt)
		b.belegt = 0
	}

	b.gemerkt[verweis] = bild
	b.belegt += groesse
}

// paketdatei liest eine Datei aus einem Paket am genannten Ort.
func paketdatei(ort, datei string) ([]byte, error) {
	baum, schliesser, err := paketbaum(ort)
	if err != nil {
		return nil, err
	}

	if schliesser != nil {
		defer func() { _ = schliesser.Close() }() //nolint:errcheck // gelesen ist gelesen
	}

	// **Ohne eigene Ausbruchsprüfung, und das ist kein Versehen.**
	// [ZerlegePaketpfad] hat `.` und `..` schon am rohen Verweis abgewiesen und
	// den Rest durch `path.Clean` geschickt; eine zweite Prüfung hier wäre ein
	// Zweig, den kein Aufruf betreten kann — also eine Behauptung darüber, was
	// im Ernstfall geschieht (§3). Der Rückhalt bleibt: `fs.FS.Open` weist
	// jeden Pfad ab, den `fs.ValidPath` nicht mag.
	inhalt, fehler := leseBegrenzt(baum, datei)
	if fehler != nil {
		return nil, fmt.Errorf("%s in %q lesen: %w", datei, ort, fehler)
	}

	return inhalt, nil
}

// Gibtes sagt, ob eine Datei in einem der Pakete liegt.
//
// **Ohne sie zu entschlüsseln.** Die Vorlagen fragen nach jeder Datei, nicht
// nur nach Bildern:
//
//	if exists_in_package(path) then path
//	else "/magic-modules.mse-include/loyalty/" + name
//
// `magic.mse-game/script`, Zeile 3614. Ein Leser, der das Bild dekodierte,
// sagte bei einer Textdatei „nein", obwohl sie dasteht.
//
// **Ein unbrauchbarer Verweis ist ein Nein, kein Fehler.** Die Vorlage fragt,
// um sich zu entscheiden; eine Ausnahme brächte sie um ihre Entscheidung
// (#270).
//
// **Diese Funktion steht in keiner der beiden portierbaren Quellen.** Weder
// MSE2 f9d9356 noch `fenhl/magic-set-generator` kennen `exists_in_package`;
// das Pack benutzt sie trotzdem. Was sie bedeutet, ist an ihrer Verwendung
// **erschlossen** und nicht nachgelesen — und das steht hier, damit niemand
// sie für belegt hält (§10).
func (b *Bildquelle) Gibtes(verweis string) bool {
	pfad, err := ZerlegePaketpfad(verweis)
	if err != nil {
		return false
	}

	if _, da := b.nachschlagen(verweis); da {
		return true
	}

	for _, wurzel := range b.wurzeln {
		if _, fehler := paketdatei(filepath.Join(wurzel, pfad.Paket), pfad.Datei); fehler == nil {
			return true
		}
	}

	return false
}
