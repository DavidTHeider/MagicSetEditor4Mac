// **Die Ablage gezeichneter Symbole** —
// #890,
// Werkstück 4 von Phase X.
//
// # Warum sie im Bildverzeichnis liegt
//
// Weil `Beiwerk` beim Sichern **alles** von dort ins geschriebene Paket
// kopiert (`ausAblageKopieren`). Ein eigenes Verzeichnis daneben käme nie mit,
// und der Fehler fiele erst auf, wenn jemand die Datei woanders öffnet — bei
// grüner Pipeline.
//
// **Deshalb bekommt sie die Bildablage eingesetzt** und rechnet den Ort nicht
// selbst aus: eine zweite Rechnung liefe auseinander, und diese hier wäre die
// stille Hälfte.

package service

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// symbolrumpf ist der Namensanfang einer abgelegten Zeichnung.
//
// **Wie MSE es tut**: `package.newFileName(value->field().name, ".mse-symbol")`
// (`src/gui/symbol/window.cpp`, Zeile 253, Fassung f9d9356 vom 2021-07-06) —
// der Feldname als Präfix, die Endung dahinter. Das Setsymbolfeld heißt in
// jedem Spielpaket `symbol`.
const symbolrumpf = "symbol"

// Symbolablage schreibt und liest gezeichnete `.mse-symbol`-Dateien eines Sets.
type Symbolablage struct{ bilder *Bildablage }

// NeueSymbolablage baut die Ablage über der Bildablage.
func NeueSymbolablage(bilder *Bildablage) *Symbolablage {
	return &Symbolablage{bilder: bilder}
}

// Ablegen schreibt eine Zeichnung und gibt den Namen, unter dem sie im Set
// steht.
func (s *Symbolablage) Ablegen(
	setID int64, zeichnung msepaket.Symbolzeichnung, unter string,
) (string, error) {
	// **Dieselben Grenzen wie beim Lesen** — DOM-02.AK-04 (#1697): was
	// sich nicht einlesen ließe, soll auch nicht entstehen.
	if err := msepaket.PruefeSymbolumfang(zeichnung); err != nil {
		return "", err
	}

	ort, err := s.ort(setID)
	if err != nil {
		return "", err
	}

	if mkdirfehler := os.MkdirAll(ort, 0o750); mkdirfehler != nil {
		return "", fmt.Errorf("Symbolverzeichnis anlegen: %w", mkdirfehler)
	}

	name, namensfehler := abzulegenderName(ort, unter, s.bilder.schranke())
	if namensfehler != nil {
		return "", namensfehler
	}

	// **Über eine Teildatei, nicht unmittelbar** ([schreibeAtomar]).
	//
	// `os.WriteFile` öffnet mit `O_TRUNC`: bricht das Schreiben ab — volle
	// Platte, entzogenes Medium —, bleibt eine **halbe** `.mse-symbol` liegen.
	// Der nächste Versuch nähme `symbol2.mse-symbol`, und der Setschreiber
	// kopierte die beschädigte erste als Beiwerk mit. Angestrichen vom
	// Zweitblick über Phase X.
	if schreibfehler := schreibeAtomar(filepath.Join(ort, name),
		msepaket.SchreibeSymbolzeichnung(zeichnung)); schreibfehler != nil {
		return "", fmt.Errorf("Symboldatei schreiben: %w", schreibfehler)
	}

	return name, nil
}

// abzulegenderName gibt den Namen, unter dem geschrieben wird.
//
// **Ein Set hat ein Symbol, und dafür genügt eine Datei.** Wer `unter` nennt,
// überschreibt sie; wer nichts nennt, bekommt einen freien Namen.
//
// # Warum das nicht immer ein freier Name sein darf
//
// Bis zum 2026-08-25 war es einer, und zwar jedes Mal. Zeichnen, ins Set
// legen, weiterzeichnen, wieder ins Set legen — das gab `symbol2.mse-symbol`,
// `set.symbol` zeigte darauf, und `symbol1.mse-symbol` blieb liegen. Beim
// Sichern kopiert `ausAblageKopieren` **jede** Datei des Verzeichnisses; die
// verwaiste ging als Beiwerk mit ins Paket.
//
// **Der Name wird geprüft, nicht vertraut** — genau wie in [Symbolablage.Gestalt].
// Er kommt aus einem Feldwert, den eine fremde Datei trägt (§8); ein `..`
// darin schriebe aus dem Verzeichnis heraus.
//
// Angestrichen vom Nachlauf des Zweitblicks über Phase X, Befund 1 (#890).
func abzulegenderName(ort, unter string, hoechstens int) (string, error) {
	if strings.TrimSpace(unter) == "" {
		return freierName(ort, symbolrumpf, ".mse-symbol", hoechstens)
	}

	sauber, gepruefterfehler := gepruefterName(unter)
	if gepruefterfehler != nil {
		return "", gepruefterfehler
	}

	return sauber, nil
}

// gepruefterName weist einen Namen ab, der aus dem Verzeichnis führt.
func gepruefterName(name string) (string, error) {
	sauber := strings.TrimSpace(name)
	if sauber == "" || sauber != filepath.Base(sauber) || sauber == ".." {
		return "", fmt.Errorf(
			"%w: %q ist kein Name einer abgelegten Zeichnung", domain.ErrValidation, name)
	}

	return sauber, nil
}

// Gestalt liest eine abgelegte Zeichnung zurück.
//
// **Der Name wird geprüft, nicht vertraut.** Er kommt über eine Naht von außen
// — aus einem Feldwert, den eine fremde Datei trägt (§8). Ein Name mit einem
// Schrägstrich oder einem `..` darin führt aus dem Verzeichnis heraus, und ein
// Leser, der ihm folgte, läse eine beliebige Datei des Rechners.
func (s *Symbolablage) Gestalt(setID int64, name string) (msepaket.Symbolzeichnung, error) {
	ort, err := s.ort(setID)
	if err != nil {
		return msepaket.Symbolzeichnung{}, err
	}

	sauber, namensfehler := gepruefterName(name)
	if namensfehler != nil {
		return msepaket.Symbolzeichnung{}, namensfehler
	}

	zeichnung, _, lesefehler := msepaket.LiesSymbolzeichnung(filepath.Join(ort, sauber))
	// **Eine fehlende Datei ist „nicht gefunden"** (C01, #2226) — roh kam
	// sie als unerwarteter Fehler an, mit dem vollen Pfad im Protokoll.
	if errors.Is(lesefehler, fs.ErrNotExist) {
		return msepaket.Symbolzeichnung{}, fmt.Errorf("%w: das Symbol %q liegt nicht neben dem Set",
			domain.ErrNotFound, sauber)
	}

	if lesefehler != nil {
		return msepaket.Symbolzeichnung{}, lesefehler
	}

	return zeichnung, nil
}

// ort gibt das Verzeichnis eines Sets, oder sagt, dass es keines gibt.
func (s *Symbolablage) ort(setID int64) (string, error) {
	if s.bilder == nil {
		return "", fmt.Errorf("%w: für Symbole ist keine Ablage eingerichtet",
			domain.ErrValidation)
	}

	ort := s.bilder.Verzeichnis(setID)
	if ort == "" {
		return "", fmt.Errorf("%w: für Symbole ist kein Verzeichnis eingerichtet",
			domain.ErrValidation)
	}

	return ort, nil
}
