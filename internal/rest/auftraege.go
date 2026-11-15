package rest

// **Das Auftragsbuch des Dienstes** — FMT-03.AK-02
// (#1692).
//
// # Was ein Auftrag ist
//
// Ein Lauf eines der [auftragswege], um den jemand mit `Prefer:
// respond-async` bat (`auftragswege.go`). Das Buch nimmt ihn an, fährt ihn
// über einen [Auftragslauf] nebenher und hält danach seine Antwort — Status,
// Art und Rumpf, so wie der Weg sie gleich gegeben hätte.
//
// # Warum es eine Naht ist und kein Feld des Dienstes
//
// **Der Dienst hält keinen Zustand** außer Marke und Drossel, und jeder Weg
// ruft genau eine Naht — darauf baut die Vier-Klassen-Probe jedes Endpunkts.
// Das Buch ist deshalb eine Naht wie die übrigen; die Wurzel setzt dieses
// ein, ein Test sein Doppel.
//
// # Was es hält, und wie lange
//
// Es lebt im Prozess, wie das Auftragsbuch des Fensters: nach einem Neustart
// gibt es keinen Auftrag mehr, und das ist die Wahrheit. **Vier laufen
// zugleich** — die Fachläufe dahinter gehen ohnehin einer nach dem anderen
// (`domain.Vorgabegrenzen`), und jeder hält Speicher für seine Antwort.
// Beendete bleiben eine Stunde, **höchstens sechzehn**. **Ein Auftrag gehört
// dem Principal, der ihn anstieß** — ein anderer sieht ihn nicht, auch nicht
// mit der richtigen Kennung, und die Kennung ist gewürfelt.

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Grenzen des Auftragsbuchs — **aus `domain.Vorgabegrenzen`** (#2003):
// dort stehen sie bei den übrigen Grenzen, und `grenzen` nennt sie an jeder
// Fläche. Bis dahin standen sie nur hier und im Satz einer Ablehnung.
var (
	gleichzeitigeAuftraege = domain.Vorgabegrenzen().Auftraege.Laufend
	behalteneAuftraege     = domain.Vorgabegrenzen().Auftraege.Behalten
	aufbewahrung           = domain.Vorgabegrenzen().Auftraege.Aufbewahrung
)

// Die Stände eines Auftrags.
const (
	auftragLaeuft      = "laeuft"
	auftragFertig      = "fertig"
	auftragGescheitert = "gescheitert"
	auftragAbgebrochen = "abgebrochen"
)

// ErrZuVieleAuftraege sagt, dass schon so viele laufen, wie das Buch hält.
var ErrZuVieleAuftraege = fmt.Errorf("es laufen schon %d Aufträge — bitte warten, bis einer fertig ist",
	gleichzeitigeAuftraege)

// Auftragseingabe ist die Anfrage eines Auftrags — für den Lauf und seine
// Wiederholung.
type Auftragseingabe struct {
	Weg, Methode, Pfad, Art string
	Rumpf                   []byte
}

// Auftragsergebnis ist die aufgenommene Antwort.
type Auftragsergebnis struct {
	Status int
	Art    string

	// Ort ist `Location` — der Ausgang einer Ausgabe ohne Ziel.
	Ort   string
	Rumpf []byte
}

// Auftragslauf fährt einen Auftrag und gibt seine Antwort; `melde` hört den
// Fortschritt.
type Auftragslauf func(ctx context.Context, eingabe Auftragseingabe, melde domain.Fortschrittsmelder) Auftragsergebnis

// restauftrag ist ein Auftrag im Buch.
type restauftrag struct {
	id, principal, wiederholungVon string
	eingabe                        Auftragseingabe

	stand             string
	fortschritt       domain.Fortschritt
	begonnen, beendet time.Time
	ergebnis          *Auftragsergebnis
	abbrechen         context.CancelFunc
	abbruchVerlangt   bool
}

// Auftragsbuch führt die Aufträge eines Dienstes — die jüngsten zuerst.
type Auftragsbuch struct {
	sperre    sync.Mutex
	uhr       func() time.Time
	auftraege []*restauftrag

	// geschlossen und laeufe gehören zum Dienstende (BK-07b, #2116,
	// `auftragsende.go`).
	geschlossen bool
	laeufe      sync.WaitGroup
}

// NeuesAuftragsbuch baut das Buch mit einer Uhr.
func NeuesAuftragsbuch(uhr func() time.Time) *Auftragsbuch { return &Auftragsbuch{uhr: uhr} }

// Annehmen trägt einen Auftrag ein und fährt ihn nebenher.
func (b *Auftragsbuch) Annehmen(ctx context.Context, eingabe Auftragseingabe, lauf Auftragslauf) (AuftragDTO, error) {
	return b.starte(ctx, eingabe, "", lauf)
}

// Wiederholen fährt die Anfrage eines beendeten Auftrags als neuen.
//
// **Mit dem Kontext dieses Rufs**: Nachweis und Profil sind die von jetzt,
// nicht die von damals.
func (b *Auftragsbuch) Wiederholen(ctx context.Context, id string, lauf Auftragslauf) (AuftragDTO, error) {
	gefunden, fehler := b.eigener(ctx, id)
	if fehler != nil {
		return AuftragDTO{}, fehler
	}

	if gefunden.stand == auftragLaeuft {
		return AuftragDTO{}, fmt.Errorf("%w: der Auftrag läuft noch — ein zweiter Lauf schriebe an dasselbe Ziel",
			domain.ErrConflict)
	}

	return b.starte(ctx, gefunden.eingabe, gefunden.id, lauf)
}

// starte trägt ein, prüft die Zahl und fährt los.
//
// **Mit den Werten der Anfrage, aber ohne ihr Ende**: Profil, Principal und
// Vorgangskennung reisen mit, und die Verbindung darf schließen, ohne dass
// der Auftrag stirbt. Er endet mit seinem eigenen Abbruch.
func (b *Auftragsbuch) starte(ctx context.Context, eingabe Auftragseingabe, von string, lauf Auftragslauf) (AuftragDTO, error) {
	laufkontext, abbrechen := context.WithCancel(context.WithoutCancel(ctx))

	b.sperre.Lock()

	if b.geschlossen {
		b.sperre.Unlock()
		abbrechen()

		return AuftragDTO{}, errBuchGeschlossen
	}

	b.aufraeumen()

	laufend := 0

	for _, anderer := range b.auftraege {
		if anderer.stand == auftragLaeuft {
			laufend++
		}
	}

	if laufend >= gleichzeitigeAuftraege {
		b.sperre.Unlock()
		abbrechen()

		return AuftragDTO{}, ErrZuVieleAuftraege
	}

	auftrag := &restauftrag{
		id: rand.Text(), principal: principalAus(ctx), wiederholungVon: von, eingabe: eingabe,
		stand: auftragLaeuft, begonnen: b.uhr(), abbrechen: abbrechen,
	}
	b.auftraege = append([]*restauftrag{auftrag}, b.auftraege...)
	dto := auftragAlsDTO(*auftrag)
	b.laeufe.Add(1)
	b.sperre.Unlock()

	go func() {
		defer b.laeufe.Done()

		// **Ein Zusammenbruch reißt den Dienst nicht mit** — hier fängt
		// `net/http` nichts mehr, denn die Goroutine gehört nicht ihm.
		defer func() {
			if grund := recover(); grund != nil {
				b.beende(auftrag, zusammengebrochen(grund))
			}
		}()

		b.beende(auftrag, lauf(laufkontext, eingabe, func(stand domain.Fortschritt) { b.melde(auftrag, stand) }))
	}()

	return dto, nil
}

// zusammengebrochen ist die Antwort eines Laufs, der nicht zu Ende kam.
func zusammengebrochen(grund any) Auftragsergebnis {
	meldung := "der Auftrag ist zusammengebrochen"

	// **Eine zu große Antwort bricht wie am Draht ab** — `aufnahme` hält
	// dieselbe Grenze wie `begrenzt`.
	if grund == http.ErrAbortHandler { //nolint:errorlint // ein Wert aus recover, kein umschlossener Fehler
		meldung = "die Antwort des Auftrags ist größer als erlaubt"
	}

	antwort := neueAufnahme(1 << 12)
	nein(antwort, http.StatusInternalServerError, "intern", meldung)

	return antwort.ergebnis()
}

// aufraeumen vergisst, was zu alt ist oder über die Zahl geht — unter der Sperre.
func (b *Auftragsbuch) aufraeumen() {
	behalten, beendete := b.auftraege[:0], 0

	for _, auftrag := range b.auftraege {
		if auftrag.stand != auftragLaeuft {
			beendete++

			if beendete > behalteneAuftraege || b.uhr().Sub(auftrag.beendet) > aufbewahrung {
				continue
			}
		}

		behalten = append(behalten, auftrag)
	}

	b.auftraege = behalten
}

// eigener gibt den Auftrag des Rufers — eine Kopie.
func (b *Auftragsbuch) eigener(ctx context.Context, id string) (restauftrag, error) {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	b.aufraeumen()

	for _, auftrag := range b.auftraege {
		if auftrag.id == id && auftrag.principal == principalAus(ctx) {
			return *auftrag, nil
		}
	}

	return restauftrag{}, fmt.Errorf("%w: den Auftrag %q gibt es nicht", domain.ErrNotFound, id)
}

// Auftraege gibt die Aufträge des Rufers.
func (b *Auftragsbuch) Auftraege(ctx context.Context) ([]AuftragDTO, error) {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	b.aufraeumen()

	eigene := []AuftragDTO{}

	for _, auftrag := range b.auftraege {
		if auftrag.principal == principalAus(ctx) {
			eigene = append(eigene, auftragAlsDTO(*auftrag))
		}
	}

	return eigene, nil
}

// Auftrag gibt einen Auftrag des Rufers.
func (b *Auftragsbuch) Auftrag(ctx context.Context, id string) (AuftragDTO, error) {
	gefunden, fehler := b.eigener(ctx, id)

	return auftragAlsDTO(gefunden), fehler
}

// Ergebnis gibt die Antwort eines beendeten Auftrags.
func (b *Auftragsbuch) Ergebnis(ctx context.Context, id string) (Auftragsergebnis, error) {
	gefunden, fehler := b.eigener(ctx, id)
	if fehler != nil {
		return Auftragsergebnis{}, fehler
	}

	if gefunden.ergebnis == nil {
		return Auftragsergebnis{}, fmt.Errorf("%w: der Auftrag läuft noch", domain.ErrConflict)
	}

	return *gefunden.ergebnis, nil
}

// Abbrechen hält einen laufenden Auftrag an — oder vergisst einen beendeten.
// Der Wert sagt, ob er lief.
func (b *Auftragsbuch) Abbrechen(ctx context.Context, id string) (bool, error) {
	if _, fehler := b.eigener(ctx, id); fehler != nil {
		return false, fehler
	}

	b.sperre.Lock()
	defer b.sperre.Unlock()

	for _, auftrag := range b.auftraege {
		if auftrag.id == id && auftrag.stand == auftragLaeuft {
			auftrag.abbruchVerlangt = true
			auftrag.abbrechen()

			return true, nil
		}
	}

	b.auftraege = slices.DeleteFunc(b.auftraege, func(auftrag *restauftrag) bool { return auftrag.id == id })

	return false, nil
}

// melde setzt den Fortschritt eines Auftrags.
func (b *Auftragsbuch) melde(auftrag *restauftrag, stand domain.Fortschritt) {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	auftrag.fortschritt = stand
}

// statusAbgebrochen ist der Status, den [fehlerhaft] einem Abbruch gibt.
const statusAbgebrochen = 499

// beende schließt einen Auftrag mit seiner Antwort.
func (b *Auftragsbuch) beende(auftrag *restauftrag, ergebnis Auftragsergebnis) {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	auftrag.ergebnis, auftrag.beendet = &ergebnis, b.uhr()

	switch {
	case auftrag.abbruchVerlangt || ergebnis.Status == statusAbgebrochen:
		auftrag.stand = auftragAbgebrochen
	case ergebnis.Status < http.StatusBadRequest:
		auftrag.stand = auftragFertig
	default:
		auftrag.stand = auftragGescheitert
	}

	// **Der Kontext wird freigegeben** — auch ohne Abbruch hält er sonst,
	// was an ihm hängt.
	auftrag.abbrechen()
}

// istZuViele sagt, ob ein Fehler die volle Zahl meint.
func istZuViele(fehler error) bool { return errors.Is(fehler, ErrZuVieleAuftraege) }
