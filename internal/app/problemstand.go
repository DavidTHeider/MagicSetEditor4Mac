package app

// **Die Probleme eines ganzen Sets, für jede Fläche** — PERF02
// (#2352).
//
// Der Stand sagt mit, **wie weit er gilt**: aktuell, veraltet, unvollständig
// oder ungeprüft, und wann er entstand. Eine Zahl
// ist nur dann `sicher`, wenn der Stand aktuell ist — das falsche Null aus dem
// Issue gibt es damit nicht.

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KartenproblemDTO sind die Befunde einer Karte — als Sätze.
type KartenproblemDTO struct {
	KarteID string   `json:"karteID"`
	Befunde []string `json:"befunde"`
}

// ProblemstandDTO ist der Stand der Prüfung eines Sets.
type ProblemstandDTO struct {
	// Urteil ist `ungeprueft`, `aktuell`, `veraltet` oder `unvollstaendig`.
	Urteil string `json:"urteil"`
	// Zahl sind die Karten mit Problemen — nur bei `sicher` eine, die gilt.
	Zahl     int  `json:"zahl"`
	Sicher   bool `json:"sicher"`
	Geprueft int  `json:"geprueft"`
	Gesamt   int  `json:"gesamt"`
	// Beendet ist der Zeitpunkt des Stands (RFC 3339) und Alter seine Zeit in
	// Sekunden — leer und null, solange nie geprüft wurde.
	Beendet string             `json:"beendet"`
	Alter   int                `json:"alter"`
	Karten  []KartenproblemDTO `json:"karten"`
}

// problemstandAlsDTO übersetzt einen Stand — gemessen am heutigen Abdruck.
func problemstandAlsDTO(stand domain.Problemstand, heute string, jetzt time.Time) ProblemstandDTO {
	zahl, sicher := stand.Zahl(heute)
	dto := ProblemstandDTO{
		Urteil: string(stand.Urteil(heute)), Zahl: zahl, Sicher: sicher,
		Geprueft: stand.Geprueft, Gesamt: stand.Gesamt, Karten: []KartenproblemDTO{},
	}

	if !stand.Beendet.IsZero() {
		dto.Beendet = stand.Beendet.UTC().Format(time.RFC3339)
		dto.Alter = int(jetzt.Sub(stand.Beendet).Seconds())
	}

	for _, karte := range stand.Karten {
		dto.Karten = append(dto.Karten, KartenproblemDTO{KarteID: strconv.FormatInt(karte.KarteID, 10), Befunde: karte.Befunde})
	}

	return dto
}

// problemstandUeber liest den Stand eines Sets aus einer Prüfung.
func problemstandUeber(ctx context.Context, pruefung *service.Problempruefung, setID int64, jetzt time.Time) (ProblemstandDTO, error) {
	stand, heute, err := pruefung.Stand(ctx, setID)
	if err != nil {
		return ProblemstandDTO{}, err
	}

	return problemstandAlsDTO(stand, heute, jetzt), nil
}

// Problemweg ist die Problemprüfung für die Flächen ohne Fenster —
// Kommandozeile, Werkzeugserver und Dienst. **Das Fenster folgt mit AP-8**
// (#2365),
// das die Zahl auf der Übersicht zeigt; eine Bindung ohne Aufrufer hält
// `bindungsaufrufer_test.go` ab.
type Problemweg struct {
	pruefung *service.Problempruefung
	uhr      func() time.Time
}

// ProblemwegUeber baut den Weg über dem Zeichenwerk des Bestands — dem, das
// die Fläche auch zum Zeichnen nimmt ([Zeichennaht]). **Es bleibt je
// Bestand**, und mit ihm die Stände: ein Dienst, der eben prüfte, nennt den
// Stand beim nächsten Ruf.
func ProblemwegUeber(db *sql.DB) (*Problemweg, error) {
	werk, err := Zeichennaht(db)
	if err != nil {
		return nil, err
	}

	return &Problemweg{pruefung: werk.Problempruefung(), uhr: time.Now}, nil
}

// Stand gibt den Stand der Prüfung eines Sets — **ohne zu prüfen**.
func (w *Problemweg) Stand(ctx context.Context, setID int64) (ProblemstandDTO, error) {
	return problemstandUeber(ctx, w.pruefung, setID, w.uhr())
}

// Pruefen prüft ein Set und gibt den Stand **dieses Laufs**. `melde` hört
// nach jeder Karte, wie viele geprüft sind. **Ein Abbruch gibt den Fehler und
// den unvollständigen Stand** — wer ihn zeigt, zeigt ihn nicht als „0
// Probleme".
//
// **Nicht der gemeinsame Speicher** (Zweitblick über Phase 3, Befund 4): dort
// stand, was ein anderer Lauf ablegte, und ein abgebrochener nannte den
// vollständigen Stand des vorigen.
func (w *Problemweg) Pruefen(ctx context.Context, setID int64, melde func(geprueft, gesamt int)) (ProblemstandDTO, error) {
	stand, err := w.pruefung.Pruefen(ctx, setID, melde)

	// **Ohne Lauf ungeprüft** — die Eingaben ließen sich nicht lesen, und
	// der Stand dieses Rufs ist der eines Sets, das niemand prüfte.
	if stand.SetID == 0 {
		return problemstandAlsDTO(domain.Problemstand{}, "", w.uhr()), err
	}

	// **Der heutige Abdruck ohne den Abbruch des Rufs**, damit ein
	// abgebrochener Lauf seinen unvollständigen Stand noch nennt.
	_, heute, standfehler := w.pruefung.Stand(context.WithoutCancel(ctx), setID)

	return problemstandAlsDTO(stand, heute, w.uhr()), errors.Join(err, standfehler)
}
