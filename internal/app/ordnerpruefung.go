package app

// **Die Platte nach einem Ordner fragen** — OPS-02.AK-01 und AK-03
// (#1719).
//
// # Was gefragt wird
//
// **Gibt es ihn, ist er ein Ordner, darf diese Anwendung hineinschreiben, und
// ist dort noch Platz?** Vier Fragen, und jede hat ihren eigenen Satz — „geht
// nicht" hülfe niemandem, der ein Laufwerk abgezogen hat.
//
// # Was nicht geschieht
//
// **Geschrieben wird nichts.** Ob ein Ordner beschreibbar ist, fragt
// `access(2)` und keine Probedatei: eine Einstellung, die über den Dienst
// gesetzt wird, soll in einem fremden Ordner keine Spur hinterlassen (AK-04).

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mindestplatz ist, was ein Ordner mindestens frei haben muss — eine
// Sicherung des Bestands ist selten kleiner, und darunter scheitert das
// nächste Schreiben mit einer Meldung, die niemand dem Ordner zuordnet.
const mindestplatz int64 = 16 << 20

// PruefeOrdner fragt die Platte nach der Angabe einer Ordnereinstellung —
// `~` gedehnt zum Zuhause dessen, der die Anwendung ausführt.
func PruefeOrdner(angabe string) error {
	_, fehler := geprueft(angabe)

	return fehler
}

// pruefeOrdnerFuer fragt die Platte — und über einen offenen Server vorher,
// ob der Ordner einer ist, den ein Server als seinen gelten lässt.
//
// **Befund 2 des Zweitblicks über Phase 6.4**: über den Dienst ließ sich
// `~/Library/LaunchAgents` einstellen, und der Pfadvertrag galt danach dort.
// Und die Frage an die Platte verriet, was in einem verborgenen Winkel liegt
// — „ist eine Datei" heißt „gibt es". Die Regel davor weist beides ab, mit
// demselben Satz für alles, was dort liegen könnte.
func pruefeOrdnerFuer(ctx context.Context, angabe string) error {
	return pruefeOrdnerUeber(ctx, angabe, domain.Laufwerke)
}

// pruefeOrdnerUeber ist [pruefeOrdnerFuer] mit einem eigenen Ort für
// Laufwerke — für Läufe, die keines einhängen können.
func pruefeOrdnerUeber(ctx context.Context, angabe, laufwerke string) error {
	if modus := DienstmodusAus(ctx); modus == domain.ModusDienst || modus == domain.ModusMCP {
		zuhause, heimfehler := os.UserHomeDir()

		// **Aufgelöst, wie der Pfadvertrag beim Gebrauch** — Nachlauf über
		// Phase 6.4, N1: `/Volumes/Macintosh HD` ist ein Verweis auf `/`, und
		// über ihn galt `~/Library/LaunchAgents` als Laufwerk.
		if fehler := errors.Join(heimfehler, domain.PruefeDienstordnerUnter(
			aufgeloest(domain.OrdnerAus(angabe, zuhause)), aufgeloest(zuhause), aufgeloest(laufwerke))); fehler != nil {
			return fehler
		}
	}

	return PruefeOrdner(angabe)
}

// geprueft gibt den gedehnten Ordner, wenn er taugt.
func geprueft(angabe string) (string, error) {
	return pruefeOrdnerMit(angabe, os.UserHomeDir, freierPlatz)
}

func pruefeOrdnerMit(
	angabe string,
	heim func() (string, error),
	platz func(string) (int64, error),
) (string, error) {
	zuhause, fehler := heim()
	if fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler, "das Zuhause bestimmen")
	}

	ordner := domain.OrdnerAus(angabe, zuhause)

	angaben, fehler := os.Stat(ordner)
	if errors.Is(fehler, fs.ErrNotExist) {
		return "", ordnerbefund(domain.SchluesselOrdnerFehlt, ordner, "",
			"Der Ordner »%s« ist nicht erreichbar. Schließe das Laufwerk an oder wähle einen anderen Ordner.")
	}

	if fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler, "den Ordner ansehen")
	}

	if !angaben.IsDir() {
		return "", ordnerbefund(domain.SchluesselOrdnerIstDatei, ordner, "",
			"»%s« ist eine Datei und kein Ordner. Wähle den Ordner, in dem sie liegt, oder einen anderen.")
	}

	if !schreibbar(ordner) {
		return "", ordnerbefund(domain.SchluesselOrdnerGesperrt, ordner, "",
			"In »%s« darf diese Anwendung nicht schreiben. Wähle einen anderen Ordner oder gib diesen frei.")
	}

	// **Ein Platz, der sich nicht bestimmen lässt, hält nicht an** — auf
	// einer Plattform ohne `Statfs` gibt es keine Zahl, und eine erfundene
	// wäre schlimmer als keine.
	if frei, platzfehler := platz(ordner); platzfehler == nil && frei < mindestplatz {
		return "", ordnerbefund(domain.SchluesselOrdnerVoll, ordner, groesse(frei),
			"Auf dem Laufwerk von »%s« sind nur noch %s frei. Schaffe Platz oder wähle einen anderen Ordner.")
	}

	return ordner, nil
}

// ordnerbefund baut den Befund eines Ordners — mit Schlüssel und Werten, damit
// das Fenster den Satz in seiner Sprache zeigt.
func ordnerbefund(schluessel domain.Schluessel, ordner, frei, vorlage string) *domain.FieldError {
	werte := map[string]string{"ordner": ordner}
	argumente := []any{ordner}

	if frei != "" {
		werte["frei"] = frei
		argumente = append(argumente, frei)
	}

	return domain.NewFieldError("wert", schluessel, vorlage, argumente...).MitWerten(werte)
}

// groesse schreibt eine Byteanzahl für Menschen.
func groesse(bytes int64) string {
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
}
