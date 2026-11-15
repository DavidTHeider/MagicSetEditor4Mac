package service

// **Die zeitweiligen Dateien des Teilens** — NET-04.AK-04
// (#1708),
// entschieden in
// [0155](../../docs/entscheidungen/0155-teilen-ist-eine-kopie.md).
//
// # Warum ein eigener Ordner je Teilung
//
// **Weil der Dienst, der die Datei nimmt, sie später liest.** Mail legt den
// Anhang beim Verfassen an, AirDrop beim Senden; wer die Datei gleich nach dem
// Wählen des Ziels löschte, schickte nichts. Deshalb bleibt sie, bis jemand
// „Fertig" sagt — oder bis das nächste Aufräumen sie findet.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Teilstelle legt die Dateien des Teilens ab und räumt sie wieder weg.
type Teilstelle struct {
	wurzel string
	jetzt  func() time.Time
}

// NeueTeilstelle baut eine Stelle unter `wurzel`.
func NeueTeilstelle(wurzel string, jetzt func() time.Time) *Teilstelle {
	return &Teilstelle{wurzel: wurzel, jetzt: jetzt}
}

// Ordner legt den Ordner einer neuen Teilung an — nur für den Eigentümer.
func (s *Teilstelle) Ordner() (kennung, ordner string, err error) {
	roh := make([]byte, 16)

	// **Ohne Fehlerzweig**: seit Go 1.24 kehrt `crypto/rand.Read` nie mit
	// einem Fehler zurück, sondern hält das Programm an.
	_, _ = rand.Read(roh) //nolint:errcheck // der Grund steht darüber

	kennung = hex.EncodeToString(roh)
	ordner = filepath.Join(s.wurzel, kennung)

	if err := os.MkdirAll(ordner, 0o700); err != nil {
		return "", "", fmt.Errorf("den Teilordner anlegen: %w", err)
	}

	return kennung, ordner, nil
}

// Ablegen schreibt eine Datei in den Ordner einer Teilung.
//
// **Der Name wird zum Namen und nie zum Weg** — `filepath.Base` nimmt nur
// seinen letzten Teil.
func (s *Teilstelle) Ablegen(
	kennung, dateiname string, inhalt []byte,
) (pfad string, groesse int64, err error) {
	if err := domain.PruefeTeilkennung(kennung); err != nil {
		return "", 0, err
	}

	pfad = filepath.Join(s.wurzel, kennung, filepath.Base(dateiname))

	// **Über eine Nachbardatei** ([schreibeAtomar]), und die legt mit `0600`
	// an — nur der Eigentümer liest die Kopie.
	if err := schreibeAtomar(pfad, inhalt); err != nil {
		return "", 0, fmt.Errorf("die Datei ablegen: %w", err)
	}

	return pfad, int64(len(inhalt)), nil
}

// Verwerfen räumt eine Teilung weg.
func (s *Teilstelle) Verwerfen(kennung string) error {
	if err := domain.PruefeTeilkennung(kennung); err != nil {
		return err
	}

	if err := os.RemoveAll(filepath.Join(s.wurzel, kennung)); err != nil {
		return fmt.Errorf("die Teilung verwerfen: %w", err)
	}

	return nil
}

// Aufraeumen räumt Teilungen weg, die älter sind als `alter` — was ein
// Absturz oder ein vergessenes „Fertig" liegen ließ.
//
// **Nur, was eine Kennung trägt**: ein Fremdkörper im Teilordner gehört
// jemand anderem, und ein Aufräumen, das ihn nähme, räumte zu viel.
func (s *Teilstelle) Aufraeumen(alter time.Duration) (int, error) {
	eintraege, err := os.ReadDir(s.wurzel)
	if os.IsNotExist(err) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("den Teilordner lesen: %w", err)
	}

	grenze := s.jetzt().Add(-alter)
	geraeumt := 0

	for _, eintrag := range eintraege {
		if !eintrag.IsDir() || domain.PruefeTeilkennung(eintrag.Name()) != nil {
			continue
		}

		info, infofehler := eintrag.Info()
		if infofehler != nil || info.ModTime().After(grenze) {
			continue
		}

		if s.Verwerfen(eintrag.Name()) == nil {
			geraeumt++
		}
	}

	return geraeumt, nil
}
