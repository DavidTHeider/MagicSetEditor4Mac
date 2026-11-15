package service

// **Eine Partie als Datei auf der Platte** — PLAY-03.AK-05
// (#1711) —
// **und ein Turnier** (DOM-04.AK-03,
// #1699).
// Zwei Formate, ein Weg: sie unterscheiden sich in der Grenze und im Namen,
// den eine Meldung trägt, und in nichts sonst.
//
// # Warum das hier steht und nicht in `app`
//
// **Weil eine Datei kein Fenster braucht.** Der Dialog, der nach einem Ort
// fragt, gehört zur Fläche; das Schreiben und Lesen gehört dorthin, wo es
// geprüft werden kann — dieselbe Trennung wie bei [Tischbildausfuhr].
//
// # Warum gelesen wird, was die Domäne ohnehin begrenzt
//
// **Weil die Grenze sonst zu spät käme.** [domain.ZerlegePartie] weist eine
// zu große Datei ab — **nachdem** sie im Arbeitsspeicher liegt. Wer eine
// Datei von fünf Gigabyte zuerst liest und dann ablehnt, hat die Ablehnung
// teuer bezahlt (§16).

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Austauschdatei legt ein Dokument ab und holt es zurück.
//
// **Zwei Felder und kein Zustand**: was abgelegt wird — für die Meldung —
// und wie groß es höchstens sein darf.
type Austauschdatei struct {
	was        string
	hoechstens int
}

// NeuePartieaustauschdatei baut sie für eine Partie.
func NeuePartieaustauschdatei() Austauschdatei {
	return Austauschdatei{was: "die Partie", hoechstens: domain.PartieaustauschHoechstens}
}

// NeueTurnieraustauschdatei baut sie für ein Turnier.
func NeueTurnieraustauschdatei() Austauschdatei {
	return Austauschdatei{was: "das Turnier", hoechstens: domain.TurnieraustauschHoechstens}
}

// NeueSprachaustauschdatei baut sie für eine eigene Übersetzung (LOC-02).
func NeueSprachaustauschdatei() Austauschdatei {
	return Austauschdatei{was: "die Übersetzung", hoechstens: hoechstensSprachpaketdatei}
}

// Schreibe legt den Inhalt an den Pfad.
//
// **Atomar**, wie jede andere Datei, die dieses Programm schreibt: eine halb
// geschriebene Partie sähe aus wie eine ganze.
func (d Austauschdatei) Schreibe(pfad string, inhalt []byte) error {
	if pfad == "" {
		return fmt.Errorf("%w: es fehlt der Ort", domain.ErrValidation)
	}

	if fehler := schreibeAtomar(pfad, inhalt); fehler != nil {
		return fmt.Errorf("%s schreiben: %w", d.was, fehler)
	}

	return nil
}

// Lies holt den Inhalt — höchstens so viel, wie das Format hergibt.
//
// **Ein Byte mehr als die Grenze wird gelesen und nicht mehr**: so
// unterscheidet sich „genau voll" von „zu groß", ohne dass die ganze Datei
// dafür im Arbeitsspeicher landet.
func (d Austauschdatei) Lies(pfad string) ([]byte, error) {
	if pfad == "" {
		return nil, fmt.Errorf("%w: es fehlt der Ort", domain.ErrValidation)
	}

	// #nosec G304 -- der Ort kommt von der Fläche: im Fenster aus dem
	// Dateidialog, an der Kommandozeile als Argument. Ein Programm, das nur
	// feste Pfade öffnet, könnte keine Datei einlesen.
	datei, fehler := os.Open(pfad) //nolint:gosec // siehe darüber
	if fehler != nil {
		if errors.Is(fehler, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: die Datei %q gibt es nicht",
				domain.ErrNotFound, pfad)
		}

		return nil, fmt.Errorf("%s lesen: %w", d.was, fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	roh, fehler := io.ReadAll(
		io.LimitReader(datei, int64(d.hoechstens)+1))
	if fehler != nil {
		return nil, fmt.Errorf("%s lesen: %w", d.was, fehler)
	}

	return roh, nil
}
