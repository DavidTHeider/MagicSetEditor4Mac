// Package schluesselbund legt Geheimnisse dort ab, wo macOS sie hütet.
//
// **Ein Blatt** (§2): es kennt `internal/domain` und sonst nichts aus diesem
// Projekt. Was hier liegt, sind Zugangsdaten fremder Dienste — Werkstück 2 von
// Phase 4.8 (#1321),
// entschieden in [0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)
// und [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
//
// **Warum nicht im Bestand:** die SQLite-Datei ist die, die Nutzende sichern
// und weitergeben. „Wer seine Datei kopiert, nimmt seine Einstellungen mit"
// steht als **Vorteil** in Migration 0010 — und für ein Geheimnis ist es der
// Nachteil.
package schluesselbund

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Hoechstlaenge ist, wie lang ein Geheimnis sein darf.
//
// **Gemessen am 2026-09-01**, nicht geschätzt: über den Weg unten laufen
// 1.024 Zeichen rund, 2.048 nicht — dort kam **nichts** zurück. Wo genau
// dazwischen die Grenze liegt, ist **nicht** nachgemessen; sie ist vermutlich
// die Länge der Befehlszeile und nicht die des Geheimnisses, denn das
// Geheimnis reist als Hex und wiegt damit doppelt.
//
// **Die Reichweite dieser Zahl** (§10): sie gilt für macOS 15 auf diesem
// Rechner. Ein Zugangsschlüssel wiegt in der Praxis unter hundert Zeichen; die
// Grenze ist da, damit nichts **still** verschwindet, nicht weil jemand sie
// erreichen wird.
const Hoechstlaenge = 1024

// Bund ist ein Schlüsselbund.
//
// **Der Ort ist ein Feld und keine Konstante**, damit ein Test nie in den
// echten Anmeldebund schreibt (§3). Leer heißt: der Bund, den macOS für
// Anmeldungen führt.
type Bund struct {
	ort   string
	frist time.Duration
}

// UeblicheFrist ist, wie lange ein Aufruf höchstens dauern darf.
//
// # Warum es sie gibt
//
// **`security` fragt bei manchen Fehlschlägen nach einem Passwort — und
// wartet dann.** Gemessen am 2026-09-02: ein Schreiben in einen Bund, den es
// nicht gibt, kam bei leerem `HOME` **nie** zurück; die Stufe `test` stand
// nach 600 Sekunden.
//
// **Ein Unterprozess ohne Frist ist derselbe Fehler wie `http.DefaultClient`
// ohne Timeout** (§16, Regel 3): im Betrieb hinge das Fenster, und abbrechen
// könnte nur, wer die Anwendung beendet.
//
// **Zehn Sekunden**, weil ein Schlüsselbundzugriff auf einer gesunden
// Maschine Millisekunden dauert — die Frist fängt den Hänger, nicht die
// Langsamkeit.
const UeblicheFrist = 10 * time.Second

// NeuerBund gibt einen Bund an einem bestimmten Ort.
func NeuerBund(ort string) *Bund {
	return &Bund{ort: ort, frist: UeblicheFrist}
}

// Bundort ist die Umgebungsvariable, die einen **anderen** Bund wählt.
//
// # Warum es sie gibt
//
// **Weil ein Lauf gegen die gebauten Programme sonst in den echten
// Anmeldebund schreibt.** Der Flächenlauf startet `mse`, `mse-mcp` und
// `mse-rest` als eigene Prozesse; ein Feld auf der `App` erreicht sie nicht,
// und ohne diesen Weg legte jeder Lauf Einträge im Schlüsselbund dieses
// Rechners ab (§3: ein Test, der nach draußen greift, ist grün, **weil** er
// wirklich funktioniert hat).
//
// **Und sie ist nicht nur für Läufe.** Dieselbe Form wie `MSE_BESTAND`: wer
// einen zweiten Bestand führt, gibt ihm einen zweiten Bund.
//
// **Leer heißt: der Anmeldebund**, wie bisher.
const Bundort = "MSE_SCHLUESSELBUND"

// Anmeldebund gibt den Bund, den macOS für Anmeldungen führt — oder den, den
// [Bundort] nennt.
func Anmeldebund() *Bund {
	return &Bund{ort: os.Getenv(Bundort), frist: UeblicheFrist}
}

// MitFrist setzt die Frist — für Tests, die nicht zehn Sekunden warten wollen.
func (b *Bund) MitFrist(frist time.Duration) *Bund {
	return &Bund{ort: b.ort, frist: frist}
}

// Schreibe legt ein Geheimnis ab — und ersetzt, was dort stand.
func (b *Bund) Schreibe(ctx context.Context, dienst, konto, geheimnis string) error {
	if err := pruefeNamen(dienst, konto); err != nil {
		return err
	}

	// **Ein leeres Geheimnis ist keines.** „Kein Schlüssel eingetragen" und
	// „ein leerer Schlüssel" wären im Bund nicht mehr zu unterscheiden, und
	// Vermerk 0063 baut auf dieser Unterscheidung auf. Wer nichts mehr
	// hinterlegt haben will, ruft [Bund.Loesche].
	if geheimnis == "" {
		return fmt.Errorf(
			"%w: ein leeres Geheimnis ist keines — zum Entfernen dient Loesche",
			domain.ErrValidation)
	}

	if len(geheimnis) > Hoechstlaenge {
		return fmt.Errorf(
			"%w: das Geheimnis ist %d Zeichen lang, höchstens %d gehen",
			domain.ErrValidation, len(geheimnis), Hoechstlaenge)
	}

	return b.schreibe(ctx, dienst, konto, geheimnis)
}

// Lies holt ein Geheimnis zurück.
func (b *Bund) Lies(ctx context.Context, dienst, konto string) (string, error) {
	if err := pruefeNamen(dienst, konto); err != nil {
		return "", err
	}

	return b.lies(ctx, dienst, konto)
}

// Loesche nimmt ein Geheimnis heraus.
//
// **Ein Eintrag, den es nicht gibt, ist kein Fehler.** Wer löscht, will, dass
// nichts mehr da ist — und das ist dann schon der Fall.
func (b *Bund) Loesche(ctx context.Context, dienst, konto string) error {
	if err := pruefeNamen(dienst, konto); err != nil {
		return err
	}

	return b.loesche(ctx, dienst, konto)
}

// pruefeNamen hält Dienst und Konto auf einem Alphabet, das keine Befehlszeile
// zerteilt.
//
// # Warum sie nicht wie das Geheimnis umkodiert werden
//
// **Weil ein Mensch sie lesen können soll.** Was hier abgelegt wird, steht in
// der Schlüsselbundverwaltung von macOS; wer dort nachsieht, soll „Scrydex"
// finden und nicht `53637279646578`. Das Geheimnis daneben darf unlesbar
// sein — es ist der Inhalt, nicht die Aufschrift.
//
// # Warum es überhaupt eine Prüfung braucht
//
// Beide reisen auf derselben Befehlszeile wie das Geheimnis, und `security`
// zerteilt sie selbst. **Gefunden beim ersten Lauf am 2026-09-01**: die
// Testfallnamen wurden als Kontonamen benutzt, und „mit Leerzeichen" ließ den
// Aufruf mit einer Nutzungsmeldung scheitern.
//
// **Sie sind Bezeichner dieses Programms und keine Eingabe.** Wer einen aus
// einer Eingabe baut, bekommt hier ein Nein statt einer verstümmelten
// Befehlszeile.
func pruefeNamen(dienst, konto string) error {
	for name, wert := range map[string]string{"Dienst": dienst, "Konto": konto} {
		if wert == "" {
			return fmt.Errorf("%w: der %s fehlt", domain.ErrValidation, name)
		}

		if strings.TrimFunc(wert, istErlaubt) != "" {
			return fmt.Errorf(
				"%w: der %s %q trägt Zeichen außer Buchstaben, Ziffern, "+
					"Punkt, Strich und Unterstrich",
				domain.ErrValidation, name, wert)
		}
	}

	return nil
}

// istErlaubt sagt, ob ein Zeichen in einem Namen stehen darf.
func istErlaubt(zeichen rune) bool {
	switch {
	case zeichen >= 'a' && zeichen <= 'z', zeichen >= 'A' && zeichen <= 'Z':
		return true
	case zeichen >= '0' && zeichen <= '9':
		return true
	case zeichen == '.', zeichen == '-', zeichen == '_':
		return true
	}

	return false
}
