package service

// **Ein Set als Archiv, ohne den Arbeitsort zu verschieben** — NET-04
// (#1708).
//
// # Warum nicht [Setschreiber.Schreibe]
//
// **Weil Schreiben den Arbeitsort wechselt** (#226), und das ist dort
// richtig: nach „Sichern unter" gilt das neue Ziel. Wer ein Set **teilt**,
// gibt eine Kopie weiter — ein Arbeitsort in einem Ordner für Mailanhänge
// wäre der Ort, an dem das nächste ⌘S landet und den niemand meint.
//
// # Warum ein Archiv
//
// **Weil ein Mailanhang eine Datei ist und kein Verzeichnis.** Ein
// `.mse-set` darf beides sein (`docs/FORMATE.md`); MSE2 und dieses Programm
// öffnen das Archiv wie den Ordner.

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Abzug schreibt ein Set als Archiv an `archiv` — und lässt den Arbeitsort,
// wo er war.
//
// **Ohne Schreibrecht am Bestand**: ein Abzug ändert nichts daran, er gibt
// her, was dasteht — dieselbe Wirkung wie ein Lesen.
func (s *Setschreiber) Abzug(ctx context.Context, setID int64, archiv string) error {
	quelle, baum, err := s.baumVon(ctx, setID)
	if err != nil {
		return err
	}

	if befund := s.fuellen(ctx, setID, baum); befund != nil {
		return befund
	}

	// **Gebaut wird neben dem Ziel und nicht darin** — ein Verzeichnis unter
	// dem Namen des Archivs stünde im Weg, sobald es entsteht.
	bau, err := os.MkdirTemp(filepath.Dir(archiv), ".abzug-*")
	if err != nil {
		return fmt.Errorf("den Bauort anlegen: %w", err)
	}

	defer func() { _ = os.RemoveAll(bau) }() //nolint:errcheck // aufgeräumt wird bestenfalls

	if _, err := s.ablegen(quelle, baum, setID, bau); err != nil {
		return err
	}

	return packen(bau, archiv, os.ReadFile)
}

// packen legt, was unter `bau` liegt, als Archiv an `archiv`.
//
// **Der Leser wird hereingereicht**, damit ein Lauf den Fehlerweg betreten
// kann: unter einem Bauort, den dieser Prozess eben angelegt hat, scheitert
// ein Lesen sonst nicht.
func packen(bau, archiv string, lesen func(string) ([]byte, error)) error {
	dateien, err := einsammeln(bau, lesen)
	if err != nil {
		return err
	}

	if err := msepaket.ArchivAnlegen(archiv, dateien); err != nil {
		return fmt.Errorf("das Archiv anlegen: %w", err)
	}

	return nil
}

// einsammeln liest jede Datei unter `bau` — mit ihrem Weg darunter als Namen.
func einsammeln(bau string, lesen func(string) ([]byte, error)) (map[string][]byte, error) {
	dateien := map[string][]byte{}

	if err := filepath.WalkDir(bau, func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil || eintrag.IsDir() {
			return fehler
		}

		inhalt, lesefehler := lesen(pfad)
		if lesefehler != nil {
			return lesefehler
		}

		name, _ := filepath.Rel(bau, pfad) //nolint:errcheck // pfad liegt unter bau
		dateien[filepath.ToSlash(name)] = inhalt

		return nil
	}); err != nil {
		return nil, fmt.Errorf("den Abzug einsammeln: %w", err)
	}

	return dateien, nil
}

// Setname ist der Name eines Sets — für den Namen seines Archivs.
func (s *Setschreiber) Setname(ctx context.Context, setID int64) (string, error) {
	zeile, err := s.leser.Setzeile(ctx, setID)
	if err != nil {
		return "", fmt.Errorf("das Set lesen: %w", err)
	}

	return zeile.Name, nil
}
