package service

// **Aktualisieren und zurück** — G04
// (#2251).
//
// # Eine vorige Fassung, nicht eine Geschichte
//
// **Eine Aktualisierung hebt den bisherigen Baum auf**, neben dem
// Paketordner, und **zurück tauscht beide** — danach liegt der neuere
// aufgehoben, und man kann wieder vor. Eine Ebene reicht, damit eine
// Aktualisierung nie etwas unwiderruflich nimmt; mehr wäre ein zweites
// Paketverzeichnis, das niemand aufräumt. Dieselbe Regel wie beim Ersetzen
// eines einzelnen Pakets (`paketersatz.go`).
//
// **Was das kostet, steht in der Vorschau**: ein aufgehobener
// `Full-Magic-Pack` wiegt über ein Gigabyte.
//
// # Woher der Bezug weiß, wie er bezogen wurde
//
// Aus `.mse4mac-bezug.json` im Baum — Adresse, Stand, Unterverzeichnis,
// Prüfsumme, geschrieben bei jedem Bezug. **Ein Baum von vorher hat keine**;
// dann gelten der Katalog und die Liste der Quellen, und das
// Unterverzeichnis ist das der Quelle oder leer.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// bezugsvermerkDatei steht im Wurzelverzeichnis jedes bezogenen Baums.
const bezugsvermerkDatei = ".mse4mac-bezug.json"

// Bezugsvermerk ist, wie ein Baum bezogen wurde.
type Bezugsvermerk struct {
	Adresse          string `json:"adresse"`
	Stand            string `json:"stand"`
	Unterverzeichnis string `json:"unterverzeichnis"`
	Pruefsumme       string `json:"pruefsumme"`
}

// Rueckweg ist, was nach einem Zurück an seinem Ort liegt.
type Rueckweg struct {
	Ort string
	Bezugsvermerk
}

// Aktualisieren bezieht `ziel` an die Stelle des bisherigen Baums und hebt
// diesen auf.
//
// **Die Vorschau wird neu gerechnet**, nicht übernommen: was jemand vor einer
// Minute bestätigt hat, kann heute anders aussehen — und verlangt sie eine
// Bestätigung, gibt es ohne keine Aktualisierung.
func (a *Paketaktualisierung) Aktualisieren(ctx context.Context, name, ziel string, bestaetigt bool) (*Bezugsergebnis, error) {
	vorschau, err := a.Vorschau(ctx, name, ziel)
	if err != nil {
		return nil, err
	}

	if vorschau.BestaetigungNoetig && !bestaetigt {
		return nil, domain.NewFieldError("bestaetigt", domain.SchluesselValidierung,
			"Die Aktualisierung von %q verlangt eine Bestätigung — die Vorschau nennt, warum.", name)
	}

	stand, err := a.bezogen(ctx, name, ziel)
	if err != nil {
		return nil, err
	}

	alt, err := a.vermerkVon(ctx, stand)
	if err != nil {
		return nil, err
	}

	// **Der Vermerk geht mit dem Baum** — ein Baum von vorher bekommt ihn
	// jetzt, damit das Zurück weiß, was es zurückholt.
	if err := vermerkSchreiben(stand.Ort, alt); err != nil {
		return nil, err
	}

	return a.bezieher.beziehen(ctx, alt.Adresse, ziel, alt.Unterverzeichnis, name, a.vorigeOrt(name))
}

// Zurueck tauscht einen Baum mit seinem aufgehobenen Stand.
func (a *Paketaktualisierung) Zurueck(_ context.Context, name string) (*Rueckweg, error) {
	if err := namePruefen(name); err != nil {
		return nil, err
	}

	vorige := a.vorigeOrt(name)
	ort := filepath.Join(a.verwalter.verwaltet, name)

	vermerk, err := vorigerVermerk(vorige)
	if err != nil {
		return nil, err
	}

	fertig, sperrfehler := alleinAmBaum(ort)
	if sperrfehler != nil {
		return nil, sperrfehler
	}

	defer fertig()

	if err := tauschen(ort, vorige, os.Rename); err != nil {
		return nil, fmt.Errorf("den vorigen Stand zurückholen: %w", err)
	}

	return &Rueckweg{Ort: ort, Bezugsvermerk: vermerk}, nil
}

// tauschen vertauscht zwei Verzeichnisse über einen dritten Namen.
//
// **Scheitert der zweite Zug, geht der erste zurück** — es soll nie ein
// Zustand bleiben, in dem keiner von beiden an seinem Ort liegt.
func tauschen(ort, vorige string, umbenennen func(alt, neu string) error) error {
	zwischen := vorige + ".tausch"

	if err := umbenennen(ort, zwischen); err != nil {
		return err
	}

	if err := umbenennen(vorige, ort); err != nil {
		return errors.Join(err, umbenennen(zwischen, ort))
	}

	// **Auch der dritte Zug hat seinen Rückweg** (Befund 1 des Zweitblicks
	// über Phase G, zweiter Teil): bis dahin lag der zurückgeholte Baum schon
	// am Ort, der andere unter `.tausch`, und gemeldet wurde ein Fehlschlag.
	// Scheitert der Rückweg selbst, steht er im Fehler.
	if err := umbenennen(zwischen, vorige); err != nil {
		return errors.Join(err, umbenennen(ort, vorige), umbenennen(zwischen, ort))
	}

	return nil
}

// vermerkVon gibt, wie ein Baum bezogen wurde — aus seiner Datei oder,
// fehlt sie, aus Katalog und Quellenliste.
func (a *Paketaktualisierung) vermerkVon(ctx context.Context, stand Bezugsstand) (Bezugsvermerk, error) {
	if vermerk, err := vorigerVermerk(stand.Ort); err == nil {
		return vermerk, nil
	}

	vermerk := Bezugsvermerk{Adresse: stand.Adresse, Stand: stand.Stand}

	for _, quelle := range bezug.Quellen() {
		if quelle.Adresse == stand.Adresse {
			vermerk.Unterverzeichnis = quelle.Unterverzeichnis
		}
	}

	pakete, err := a.verwalter.katalog.Pakete(ctx)
	if err != nil {
		return Bezugsvermerk{}, err
	}

	for _, paket := range pakete {
		if teile, imOrdner := pfadteileUnter(a.verwalter.verwaltet, paket.Pfad); imOrdner && teile[0] == stand.Name {
			vermerk.Pruefsumme = paket.Pruefsumme

			break
		}
	}

	return vermerk, nil
}

// vorigerVermerk liest den Vermerk eines Baums.
//
// **Begrenzt** (`msepaket.LeseDateiBegrenzt`): die Datei liegt im Baum, und
// einen Baum füllt ein fremdes Archiv. **Und ohne Pfad in der Meldung** (§8):
// der Ort geht ins Protokoll, der Satz an den Menschen.
func vorigerVermerk(ort string) (Bezugsvermerk, error) {
	inhalt, err := msepaket.LeseDateiBegrenzt(filepath.Join(ort, bezugsvermerkDatei))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Bezugsvermerk{}, fmt.Errorf("%w: es liegt kein aufgehobener Stand bereit", domain.ErrNotFound)
		}

		return Bezugsvermerk{}, domain.MitGrund(domain.ErrValidation, err,
			"Der Vermerk eines bezogenen Pakets lässt sich nicht lesen.")
	}

	var vermerk Bezugsvermerk
	if err := json.Unmarshal(inhalt, &vermerk); err != nil {
		return Bezugsvermerk{}, domain.MitGrund(domain.ErrValidation, err,
			"Der Vermerk eines bezogenen Pakets ist kein JSON.")
	}

	return vermerk, nil
}

// vermerkSchreiben legt den Vermerk in einen Baum.
func vermerkSchreiben(ort string, vermerk Bezugsvermerk) error {
	// Vier Zeichenketten lassen sich immer als JSON schreiben.
	inhalt, _ := json.MarshalIndent(vermerk, "", "  ") //nolint:errcheck // der Grund steht darüber

	if err := schreibeAtomar(filepath.Join(ort, bezugsvermerkDatei), inhalt); err != nil {
		return fmt.Errorf("den Vermerk schreiben: %w", err)
	}

	return nil
}
