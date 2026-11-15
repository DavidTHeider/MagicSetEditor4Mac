// Die Naht zu den Sammeldaten
// (#888).
//
// **Ohne Bestand**, wie die Auskünfte: eine Sammeldatei zu holen liest nichts
// aus der Datenbank.

package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammeldaten ist der Weg zu den Sammeldateien von Scryfall.
type Sammeldaten struct {
	quelle Sammelquelle
	pfad   string
}

// Sammelquelle ist, was die Naht vom Klienten braucht.
type Sammelquelle interface {
	Verzeichnis(ctx context.Context) ([]Sammeleintrag, error)
	Holen(ctx context.Context, eintrag Sammeleintrag, ziel string) error
}

// Sammeleintrag ist eine angebotene Sammeldatei.
type Sammeleintrag struct {
	Art          string `json:"art"`
	Name         string `json:"name"`
	Beschreibung string `json:"beschreibung"`
	Adresse      string `json:"-"`
	Bytes        int64  `json:"bytes"`
	Stand        string `json:"stand"`
}

// Sammeldatennaht baut den Weg.
func Sammeldatennaht(wahl Dienstwahl) (*Sammeldaten, error) {
	// **Beide auf einmal, ein Zweig** — dieselbe Überlegung wie in
	// [Zeichennaht]: sie scheitern an derselben Ursache, dem fehlenden
	// Benutzerverzeichnis. Ein zweites `if` darauf wäre ein Zweig, den kein
	// Aufruf betreten kann (§3).
	quelle, quellfehler := Sammelquellenaht(wahl)

	pfad, pfadfehler := Bulkpfad()

	if fehler := errors.Join(quellfehler, pfadfehler); fehler != nil {
		return nil, fehler
	}

	return &Sammeldaten{quelle: quelle, pfad: pfad}, nil
}

// Stand sagt, was liegt und was angeboten wird.
//
// **Ohne Netz sagt er trotzdem, was liegt.** Ein Ausfall des Verzeichnisses
// darf nicht die Auskunft über die eigene Platte verhindern — das wäre die
// Umkehrung von §16, Regel 5.
func (s *Sammeldaten) Stand(ctx context.Context) (Sammeldatenstand, error) {
	stand := Sammeldatenstand{Art: bulkart, Pfad: s.pfad}

	if angaben, fehler := os.Stat(s.pfad); fehler == nil {
		stand.Vorhanden = true
		stand.Bytes = angaben.Size()
		stand.Geholt = angaben.ModTime().UTC().Format(time.RFC3339)
	}

	eintraege, fehler := s.quelle.Verzeichnis(ctx)
	if fehler != nil {
		// **Der Ausfall des Verzeichnisses ist hier kein Fehler**, und der
		// Linter hat recht, dass das ungewöhnlich aussieht (nilerr).
		//
		// Der Grund steht über der Methode: die Auskunft über die **eigene
		// Platte** darf nicht daran hängen, dass Scryfall antwortet. Wer ohne
		// Netz nachsieht, ob die Sammeldatei bereitliegt, bekommt die Antwort
		// — nur die angebotene Größe fehlt dann.
		//nolint:nilerr // §16 Regel 5: ohne Netz arbeitet die Anwendung weiter
		return stand, nil
	}

	for _, einer := range eintraege {
		if einer.Art == bulkart {
			stand.Angeboten = einer.Bytes
			stand.Stand = einer.Stand
		}
	}

	return stand, nil
}

// Holen lädt die Sammeldatei und legt sie ab.
func (s *Sammeldaten) Holen(ctx context.Context) (Sammeldatenstand, error) {
	eintraege, fehler := s.quelle.Verzeichnis(ctx)
	if fehler != nil {
		return Sammeldatenstand{}, fehler
	}

	for _, einer := range eintraege {
		if einer.Art != bulkart {
			continue
		}

		if holfehler := s.quelle.Holen(ctx, einer, s.pfad); holfehler != nil {
			return Sammeldatenstand{}, holfehler
		}

		return s.Stand(ctx)
	}

	return Sammeldatenstand{}, fmt.Errorf(
		"%w: Scryfall bietet keine Sammeldatei der Art %q an",
		domain.ErrNotFound, bulkart)
}

// SammeldatenAntwort ist der Stand, wie das Fenster ihn sieht.
type SammeldatenAntwort struct {
	Art       string `json:"art"`
	Pfad      string `json:"pfad"`
	Vorhanden bool   `json:"vorhanden"`

	// MB sind die Größen in Mebibyte, gerundet — das Fenster soll nicht rechnen.
	MB          float64 `json:"mb"`
	AngebotenMB float64 `json:"angebotenMB"`

	Stand string `json:"stand"`

	// Geholt ist, wann die liegende Datei geschrieben wurde (F04, #2246) —
	// [Sammeldatenstand.Geholt].
	Geholt string `json:"geholt"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// ScryfallSammeldaten sagt, ob die Sammeldatei bereitliegt.
func (a *App) ScryfallSammeldaten(ctx context.Context) SammeldatenAntwort {
	ctx = a.imProfil(ctx)

	return a.sammeldaten(ctx, false)
}

// ScryfallSammeldatenHolen lädt die Sammeldatei.
func (a *App) ScryfallSammeldatenHolen(ctx context.Context) SammeldatenAntwort {
	ctx = a.imProfil(ctx)

	return a.sammeldaten(ctx, true)
}

// sammeldaten führt beide Wege — sie unterscheiden sich in einem Aufruf.
func (a *App) sammeldaten(ctx context.Context, holen bool) SammeldatenAntwort {
	naht, fehler := Sammeldatennaht(a.wahl())
	if fehler != nil {
		return SammeldatenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	stand, standfehler := naht.Stand(ctx)
	if holen {
		beenden := a.auftragBeginnen("sammeldaten", "")
		stand, standfehler = naht.Holen(ctx)
		beenden(standfehler)
	}

	if standfehler != nil {
		return SammeldatenAntwort{Fehler: Uebersetze(ctx, standfehler)}
	}

	return SammeldatenAntwort{
		Art: stand.Art, Pfad: stand.Pfad, Vorhanden: stand.Vorhanden,
		MB:          float64(stand.Bytes) / 1e6,
		AngebotenMB: float64(stand.Angeboten) / 1e6,
		Stand:       stand.Stand,
		Geholt:      stand.Geholt,
	}
}

// SammeldatenUeber baut den Weg über einer gesetzten Quelle — **nur für Tests**.
//
// **Sie steht hier und nicht in einer Testdatei**, weil `Sammeldaten` seine
// Felder nicht exportiert: ein Test im Paket `app_test` könnte sie sonst nicht
// setzen. Dieselbe Naht wie [MitScryfalladresse].
//
// **Der Pfad kommt vom Aufrufer.** Der erste Anlauf holte ihn über [Bulkpfad]
// und führte einen Fehlerzweig darauf — den kein Test betreten kann, weil ein
// Test, der ihn brauchte, kein Zuhause hätte und dann schon vorher scheiterte.
func SammeldatenUeber(
	t interface{ Helper() }, quelle Sammelquelle, pfad string,
) *Sammeldaten {
	t.Helper()

	return &Sammeldaten{quelle: quelle, pfad: pfad}
}
