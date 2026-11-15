package service

// **Eigene Übersetzungen** — LOC-02
// (#1688).
//
// # Kopieren heißt hier verweisen
//
// Das Issue verlangt einen Assistenten, der „eine Basissprache kopiert".
// **Angelegt wird hier ein Paket, das auf seine Basis verweist**, und jeder
// Satz, der fehlt, zeigt die Basis. Das ist dieselbe Oberfläche wie nach
// einer Kopie — mit einem Unterschied, und für ihn gibt es die Abweichung:
// eine Kopie machte einen unübersetzten Satz ununterscheidbar von einem
// übersetzten, und der Assistent könnte nicht mehr sagen, was noch fehlt
// (AK-01). Vermerk [0147](../../docs/entscheidungen/0147-eigene-uebersetzungen.md).
//
// # Geprüft wird hier, an jeder Fläche gleich
//
// Gegen den deutschen Katalog, den `internal/app/sprachvorlage.json` flach
// trägt. Fenster, Kommandozeile, REST und MCP gehen durch dieselben
// Methoden; keine prüft für sich.
//
// **new** — gesucht in `src/data/locale.cpp` und `src/gui/` (Fassung f9d9356
// vom 2021-07-06): MSE2 lädt `.mse-locale`-Pakete, legt aber keine an und
// prüft keine gegen eine Vorlage.

import (
	"context"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sprachpaketablage hält die eigenen Übersetzungen.
type Sprachpaketablage interface {
	Sprachpakete(ctx context.Context) ([]domain.Sprachpaket, error)
	Sprachpaket(ctx context.Context, kennung string) (domain.Sprachpaket, error)
	SprachpaketAnlegen(ctx context.Context, paket domain.Sprachpaket) error
	SprachpaketSchreiben(ctx context.Context, paket domain.Sprachpaket) error

	// SprachpaketAendern liest, ändert und schreibt **in einem Zug** —
	// sonst überschriebe ein zweiter Prozess, was der erste schrieb.
	SprachpaketAendern(ctx context.Context, kennung string, aendere func(*domain.Sprachpaket) error) error
	SprachpaketLoeschen(ctx context.Context, kennung string) error
}

// Sprachpakete ist der Anwendungsfall der eigenen Übersetzungen.
type Sprachpakete struct {
	// **Ausgeben schreibt eine Datei** — die hält eine lesende Verbindung
	// nicht auf (Vermerk 0078).
	wache

	ablage   Sprachpaketablage
	vorlage  domain.Katalogvorlage
	programm string
	uhr      func() time.Time
}

// NeueSprachpakete baut ihn — mit der Vorlage, gegen die geprüft wird, und
// dem Namen des Programms für eine ausgegebene Datei.
func NeueSprachpakete(
	ablage Sprachpaketablage, vorlage domain.Katalogvorlage, programm string, uhr func() time.Time,
) *Sprachpakete {
	return &Sprachpakete{ablage: ablage, vorlage: vorlage, programm: programm, uhr: uhr}
}

// MitPruefer setzt den Riegel vor dem Schreiben einer Datei.
func (s *Sprachpakete) MitPruefer(pruefer Pruefer) *Sprachpakete {
	s.pruefer = pruefer

	return s
}

// Sprachpaketauskunft ist ein Paket samt seinem Abgleich.
type Sprachpaketauskunft struct {
	Paket    domain.Sprachpaket
	Abgleich domain.Sprachabgleich
}

// Sprachpaketangaben sind die Angaben, die sich ändern sollen — `nil` heißt:
// bleibt.
type Sprachpaketangaben struct {
	Name, Stand, Lizenz, Herkunft *string
}

// Liste gibt jedes Paket mit seinem Abgleich — **ohne seine Sätze**: bei
// acht Paketen wären es sonst dreißigtausend, und die Liste nennt nur
// Zahlen.
func (s *Sprachpakete) Liste(ctx context.Context) ([]Sprachpaketauskunft, error) {
	pakete, err := s.ablage.Sprachpakete(ctx)
	if err != nil {
		return nil, err
	}

	liste := make([]Sprachpaketauskunft, 0, len(pakete))

	for _, paket := range pakete {
		auskunft := s.auskunft(paket)
		auskunft.Paket.Texte = nil
		liste = append(liste, auskunft)
	}

	return liste, nil
}

// Zeige gibt ein Paket mit seinen Sätzen.
func (s *Sprachpakete) Zeige(ctx context.Context, kennung string) (Sprachpaketauskunft, error) {
	paket, err := s.ablage.Sprachpaket(ctx, kennung)
	if err != nil {
		return Sprachpaketauskunft{}, err
	}

	return s.auskunft(paket), nil
}

// Anlegen legt eine Übersetzung über einer Basis an — als Entwurf, ohne
// einen Satz.
func (s *Sprachpakete) Anlegen(ctx context.Context, kennung, name, basis string) (Sprachpaketauskunft, error) {
	paket := domain.Sprachpaket{
		Kennung: kennung, Name: strings.TrimSpace(name), Basis: basis,
		Stand: domain.StandEntwurf, Texte: map[string]domain.Uebersetzungstext{},
	}

	if err := paket.Pruefe(); err != nil {
		return Sprachpaketauskunft{}, err
	}

	if err := s.ablage.SprachpaketAnlegen(ctx, paket); err != nil {
		return Sprachpaketauskunft{}, err
	}

	return s.Zeige(ctx, kennung)
}

// Aendern setzt Angaben und Sätze — **ganz oder gar nicht**: ein falscher
// Satz unter vielen schreibt keinen davon.
//
// Ein leerer Satz nimmt die Übersetzung zurück; dann gilt wieder die Basis.
// Die Quelle eines Satzes ist der deutsche Satz **von heute** — wer
// übersetzt, übersetzt gegen ihn.
func (s *Sprachpakete) Aendern(
	ctx context.Context, kennung string, angaben Sprachpaketangaben, texte map[string]string,
) (Sprachpaketauskunft, error) {
	err := s.ablage.SprachpaketAendern(ctx, kennung, func(paket *domain.Sprachpaket) error {
		angaben.anwenden(paket)

		if err := paket.Pruefe(); err != nil {
			return err
		}

		// **Geordnet**: zwei falsche Sätze geben zweimal denselben Befund.
		for _, pfad := range slices.Sorted(maps.Keys(texte)) {
			text := texte[pfad]
			if text == "" {
				delete(paket.Texte, pfad)

				continue
			}

			if fehler := s.vorlage.PruefeUebersetzung(pfad, text); fehler != nil {
				return fehler
			}

			paket.Texte[pfad] = domain.Uebersetzungstext{Text: text, Quelle: s.vorlage.Quelle(pfad)}
		}

		return nil
	})
	if err != nil {
		return Sprachpaketauskunft{}, err
	}

	return s.Zeige(ctx, kennung)
}

// Loeschen entfernt eine Übersetzung. **Wer sie benutzt hat, sieht danach
// ihre Basis** — die Einstellung zeigt ins Leere, und das Fenster fällt
// zurück, statt leer zu stehen.
func (s *Sprachpakete) Loeschen(ctx context.Context, kennung string) error {
	return s.ablage.SprachpaketLoeschen(ctx, kennung)
}

// auskunft stellt ein Paket neben die Vorlage.
func (s *Sprachpakete) auskunft(paket domain.Sprachpaket) Sprachpaketauskunft {
	return Sprachpaketauskunft{Paket: paket, Abgleich: s.vorlage.Abgleich(paket)}
}

// anwenden setzt, was gesetzt ist.
func (a Sprachpaketangaben) anwenden(paket *domain.Sprachpaket) {
	if a.Name != nil {
		paket.Name = strings.TrimSpace(*a.Name)
	}

	if a.Stand != nil {
		paket.Stand = domain.Sprachstand(*a.Stand)
	}

	if a.Lizenz != nil {
		paket.Lizenz = strings.TrimSpace(*a.Lizenz)
	}

	if a.Herkunft != nil {
		paket.Herkunft = strings.TrimSpace(*a.Herkunft)
	}
}
