package service

// **Eine eigene Übersetzung als Datei** — LOC-02.AK-03 und AK-04
// (#1688).
//
// # Was hinausgeht
//
// **Herkunft, Lizenz und Stand, dazu die Sätze mit ihren Quellen — und sonst
// nichts.** Keine Einstellung, kein Profil, kein Geheimnis: die Datei hat
// eine feste Gestalt ([sprachpaketdatei]), und ein Feld, das hinzukommt,
// steht dort und im Lauf, der die Felder zählt.
//
// # Was hereinkommt
//
// **Nichts ersetzt still**: ein Paket, das es schon gibt, braucht die
// ausdrückliche Wahl. Ein Pfad, den diese Fassung nicht kennt, bleibt im
// Paket — er stammt aus einer neueren, und wer es weitergibt, verliert ihn
// nicht.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// hoechstensSprachpaketdatei ist die größte Datei, die eingelesen wird. Der
// ganze deutsche Katalog wog am 2026-09-29 rund 350 KB; mit Quellen und
// einer Sprache mit längeren Wörtern bleibt ein Vielfaches Platz.
const hoechstensSprachpaketdatei = 4 << 20

// sprachpaketdatei ist die Gestalt der Datei.
type sprachpaketdatei struct {
	Format     string                              `json:"format"`
	Kennung    string                              `json:"kennung"`
	Name       string                              `json:"name"`
	Basis      string                              `json:"basis"`
	Stand      string                              `json:"stand"`
	Lizenz     string                              `json:"lizenz"`
	Herkunft   string                              `json:"herkunft"`
	Programm   string                              `json:"programm"`
	Ausgegeben string                              `json:"ausgegeben"`
	Texte      map[string]domain.Uebersetzungstext `json:"texte"`
}

// Ausgeben schreibt eine Übersetzung als Datei und nennt ihren Namen.
//
// **Ohne Lizenz geht nichts hinaus** — wer die Datei bekommt, wüsste nicht,
// was er damit darf.
func (s *Sprachpakete) Ausgeben(ctx context.Context, kennung string) (name string, inhalt []byte, err error) {
	paket, err := s.ablage.Sprachpaket(ctx, kennung)
	if err != nil {
		return "", nil, err
	}

	if paket.Lizenz == "" {
		return "", nil, domain.NewFieldError("lizenz", domain.SchluesselSprachpaketOhneLizenz,
			"Nenne eine Lizenz, bevor du die Übersetzung weitergibst — etwa CC0-1.0.")
	}

	inhalt, _ = json.MarshalIndent(sprachpaketdatei{ //nolint:errcheck // Zeichenketten und Karten
		Format: domain.Sprachpaketformat, Kennung: paket.Kennung, Name: paket.Name,
		Basis: paket.Basis, Stand: string(paket.Stand), Lizenz: paket.Lizenz,
		Herkunft: paket.Herkunft, Programm: s.programm,
		Ausgegeben: s.uhr().UTC().Format(time.RFC3339), Texte: paket.Texte,
	}, "", " ")

	return paket.Kennung + ".sprachpaket.json", append(inhalt, '\n'), nil
}

// AusgebenNach schreibt eine Übersetzung an einen Ort — atomar und
// überschreibend: ob dort etwas liegen darf, entscheidet die Fläche.
func (s *Sprachpakete) AusgebenNach(ctx context.Context, kennung, ziel string) (string, error) {
	if err := s.darfSchreiben(ctx); err != nil {
		return "", err
	}

	name, inhalt, err := s.Ausgeben(ctx, kennung)
	if err != nil {
		return "", err
	}

	return name, NeueSprachaustauschdatei().Schreibe(ziel, inhalt)
}

// EinlesenAus liest eine Übersetzung von einem Ort.
func (s *Sprachpakete) EinlesenAus(ctx context.Context, pfad string, ersetzen bool) (Sprachpaketauskunft, error) {
	inhalt, err := NeueSprachaustauschdatei().Lies(pfad)
	if err != nil {
		return Sprachpaketauskunft{}, err
	}

	return s.Einlesen(ctx, inhalt, ersetzen)
}

// Einlesen nimmt eine Übersetzung aus einer Datei auf.
func (s *Sprachpakete) Einlesen(ctx context.Context, inhalt []byte, ersetzen bool) (Sprachpaketauskunft, error) {
	paket, err := s.lesen(inhalt)
	if err != nil {
		return Sprachpaketauskunft{}, err
	}

	switch _, fehler := s.ablage.Sprachpaket(ctx, paket.Kennung); {
	case fehler == nil && !ersetzen:
		return Sprachpaketauskunft{}, domain.NewFieldError("kennung", domain.SchluesselDoppelt,
			"Die Übersetzung %q gibt es schon — lies sie mit „ersetzen“ ein, wenn diese gelten soll.",
			paket.Kennung)
	case fehler == nil:
		err = s.ablage.SprachpaketSchreiben(ctx, paket)
	case errors.Is(fehler, domain.ErrNotFound):
		err = s.ablage.SprachpaketAnlegen(ctx, paket)
	default:
		err = fehler
	}

	if err != nil {
		return Sprachpaketauskunft{}, err
	}

	return s.Zeige(ctx, paket.Kennung)
}

// lesen macht aus einer Datei ein geprüftes Paket — oder den ersten Befund.
func (s *Sprachpakete) lesen(inhalt []byte) (domain.Sprachpaket, error) {
	if len(inhalt) > hoechstensSprachpaketdatei {
		return domain.Sprachpaket{}, dateibefund("Diese Datei ist zu groß für eine Übersetzung.")
	}

	datei := sprachpaketdatei{Texte: map[string]domain.Uebersetzungstext{}}

	leser := json.NewDecoder(bytes.NewReader(inhalt))
	leser.DisallowUnknownFields()

	if err := leser.Decode(&datei); err != nil {
		return domain.Sprachpaket{}, dateibefund("Diese Datei ist keine Übersetzung dieses Programms.")
	}

	if datei.Format != domain.Sprachpaketformat {
		if strings.HasPrefix(datei.Format, "mse4mac/sprachpaket/") {
			return domain.Sprachpaket{}, domain.NewFieldError("datei", domain.SchluesselSprachpaketNeuer,
				"Diese Übersetzung stammt aus einem neueren Programm (%s) — hole dir die neue Fassung.",
				datei.Format).MitWerten(map[string]string{"format": datei.Format})
		}

		return domain.Sprachpaket{}, dateibefund("Diese Datei ist keine Übersetzung dieses Programms.")
	}

	paket := domain.Sprachpaket{
		Kennung: datei.Kennung, Name: datei.Name, Basis: datei.Basis,
		Stand: domain.Sprachstand(datei.Stand), Lizenz: datei.Lizenz, Herkunft: datei.Herkunft,
		Texte: datei.Texte,
	}

	if err := paket.Pruefe(); err != nil {
		return domain.Sprachpaket{}, err
	}

	for _, pfad := range slices.Sorted(maps.Keys(paket.Texte)) {
		// **Ein Satz ohne Pfad gehört zu keiner Stelle** — gefunden von der
		// vollen Kampagne (C02, #2227, `FuzzSprachpaketEinlesen`): unter dem
		// leeren Pfad trug der Befund ein leeres Feld, und einer aus
		// Leerraum ging durch.
		if strings.TrimSpace(pfad) == "" {
			return domain.Sprachpaket{}, domain.NewFieldError("texte", domain.SchluesselValidierung,
				"Ein Satz ohne Pfad gehört zu keiner Stelle des Programms.")
		}

		text := paket.Texte[pfad].Text

		fehler := domain.PruefeFremdeUebersetzung(pfad, text)
		if s.vorlage.Quelle(pfad) != "" {
			fehler = s.vorlage.PruefeUebersetzung(pfad, text)
		}

		if fehler != nil {
			return domain.Sprachpaket{}, fehler
		}
	}

	return paket, nil
}

// dateibefund ist ein Befund an der Datei als Ganzer.
func dateibefund(satz string) error {
	return domain.NewFieldError("datei", domain.SchluesselValidierung, "%s", satz)
}
