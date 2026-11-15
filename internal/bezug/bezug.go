// Package bezug holt Vorlagenpakete auf den Rechner der Nutzenden.
//
// # Warum es das gibt
//
// Fremde Vorlagensammlungen kommen **nicht** ins Repository
// ([0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md), §11): sie
// stehen unter keiner Lizenz, wiegen über ein Gigabyte und enthalten geschützte
// Grafiken sowie fremde Windows-Binärdateien. Bezogen wird auf dem Rechner der
// Nutzenden, und **sie** entscheiden das.
//
// # Was dieses Paket ist
//
// Ein Blatt (§2): es kennt `internal/domain` und sonst nichts aus dem Projekt.
// Es weiß, wie ein Archiv über HTTP kommt und wie es sicher ausgepackt wird —
// nicht, was ein MSE-Paket ist.
package bezug

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Repositorium beschreibt, woher ein Paket kommt.
type Repositorium struct {
	// Adresse ist der Ort des Archivs, ohne den Stand.
	Adresse string

	// Marken ist der Ort, an dem sich eine Freigabe nachschlagen lässt.
	Marken string

	// Stand ist eine Freigabe oder ein Commit — **kein Zweig**.
	Stand string

	// Unterverzeichnis ist die Paketwurzel im Archiv. Leer: die Wurzel selbst.
	Unterverzeichnis string
}

// Ergebnis ist der Ausgang eines Bezugs.
type Ergebnis struct {
	Stand      string
	Pruefsumme string

	// Uebergangen sind die Dateien, die **nicht** ausgepackt wurden.
	//
	// Aufgezählt und nicht verschwiegen: wer nicht erfährt, dass etwas fehlt,
	// sucht später den Fehler im Paket.
	Uebergangen []string
}

// hoechstens ist die Grenze, bis zu der ein Archiv gelesen wird.
//
// **Zwei Gigabyte**, weil `Full-Magic-Pack` über eines wiegt und ein Bezug
// nicht daran scheitern soll — und weil ein Archiv ohne Grenze den
// Arbeitsspeicher füllt, bevor jemand merkt, dass die Gegenstelle sendet, was
// sie will.
//
// **Eine Veränderliche und keine Konstante** (#898). Sie ändert sich im
// Betrieb nie; ein Test dreht sie herunter, um das Verhalten **an** der Grenze
// zu fahren, ohne zwei Gigabyte zu erzeugen. Wer sie im Produktivcode
// zuweist, hat einen Fehler gefunden.
var hoechstens int64 = 2 << 30

// hoechstensEintrag ist die Grenze für eine **einzelne** Datei im Archiv.
//
// **Zwei Gigabyte**, dieselbe Zahl wie für den Bezug: eine Datei kann nicht
// größer sein als das Archiv, das sie trägt.
//
// **Und ein Überschreiten ist ein Fehler**, kein Abschneiden. Bis Phase 3.2
// deckelte `io.LimitReader` hier und `io.Copy` meldete nichts; auf der Platte
// lag eine Datei, die aussah wie eine ganze
// (#1163,
// Befund Z7-01).
var hoechstensEintrag int64 = 2 << 30

// hoechstensGesamt ist die Grenze für **alle** Einträge zusammen.
//
// **Vier Gigabyte**, das Doppelte des Bezugs: `Full-Magic-Pack` wiegt
// ausgepackt mehr als verdichtet, und ein Bezug soll daran nicht scheitern.
//
// **Ohne sie war ein Archiv unbegrenzt**: tausend Einträge zu je zwei
// Gigabyte sind einzeln zulässig.
var hoechstensGesamt int64 = 4 << 30

// hoechstensEintraege ist die Grenze für die **Zahl** der Einträge.
//
// Sie fängt, was ein Bytebudget durchlässt: sehr viele sehr kleine Einträge.
// Jeder kostet einen Systemaufruf und einen Verzeichniseintrag.
//
// **Eine Million**: `Full-Magic-Pack` führt rund 40.000 Dateien.
var hoechstensEintraege = 1 << 20

// commitform ist die Gestalt eines Commit-Hashes.
var commitform = regexp.MustCompile(`^[0-9a-f]{40}$`)

// nichtAuspacken sind die Endungen, die nicht ins Ziel kommen.
//
// **Die Regel steht hier und nicht in einem Kommentar** (#139). `M15-Magic-Pack`
// liefert `magicseteditor.exe`, `mse.exe`, `mse.com` und einen Deinstallierer
// mit — ausführbaren Code unbekannter Herkunft, den niemand angefordert hat.
var nichtAuspacken = []string{
	".exe", ".com", ".bat", ".cmd", ".msi", ".scr",
	".dll", ".dylib", ".so",
	".sh", ".bash", ".zsh", ".command",
	".app", ".pkg", ".dmg",
}

// Herkunft sagt, wessen Inhalte ein Bezug holen würde.
//
// **Zwei Angaben, kein Satz.** Der erste Anlauf gab deutschen Fließtext zurück
// und die Oberfläche zeigte ihn — in beiden Sprachen denselben. §9 verlangt das
// Gegenteil: Oberflächentexte stehen in den Katalogen, und was von hier kommt,
// sind die **Werte**, die darin eingesetzt werden. Der Zweitblick über Phase 9
// hat es angestrichen.
type Herkunft struct {
	// Quelle ist der Teil der Adresse, den jemand wiedererkennt.
	Quelle string

	// Stand ist die Freigabe oder der Commit.
	Stand string
}

// Woher sagt, wessen Inhalte geholt würden.
//
// **Vor dem Bezug, nicht danach** (#139): wer nicht weiß, woher eine Vorlage
// kommt, kann ihre Rechtslage nicht einschätzen. Der Satz dazu steht in den
// Sprachkatalogen; hier stehen die Angaben, die er nennt.
func Woher(quelle Repositorium) Herkunft {
	return Herkunft{Quelle: kurzname(quelle.Adresse), Stand: quelle.Stand}
}

// kurzname macht aus einer Adresse den Teil, den jemand wiedererkennt.
func kurzname(adresse string) string {
	ohneSchema := adresse
	if _, rest, gefunden := strings.Cut(adresse, "://"); gefunden {
		ohneSchema = rest
	}

	teile := strings.Split(strings.TrimSuffix(ohneSchema, "/"), "/")
	if len(teile) >= 3 {
		return strings.Join(teile[len(teile)-2:], "/")
	}

	return ohneSchema
}

// Hole bezieht ein Paket und packt es nach `ziel`.
func Hole(ctx context.Context, quelle Repositorium, ziel, kenner string) (*Ergebnis, error) {
	if err := standPruefen(ctx, quelle, kenner); err != nil {
		return nil, err
	}

	strom, err := stroemen(ctx, quelle.Adresse+"/"+quelle.Stand, kenner)
	if err != nil {
		return nil, err
	}

	defer func() { _ = strom.Close() }() //nolint:errcheck // nur gelesen

	// **Die Summe entsteht beim Lesen** (#1163, Befund Z7-01). Bis Phase 3.2
	// lag das ganze Archiv erst im Arbeitsspeicher — bis zu zwei Gigabyte —,
	// und danach lief `sha256` ein zweites und `gzip` ein drittes Mal darüber.
	rechner := sha256.New()

	mitlesend := io.TeeReader(strom, rechner)

	uebergangen, err := auspacken(mitlesend, quelle.Unterverzeichnis, ziel)
	if err != nil {
		return nil, err
	}

	// **Der Rest wird leergelesen.** Die Prüfsumme gilt dem **ganzen** Rumpf,
	// nicht dem, was tar gebraucht hat: ein Archiv mit Anhängsel ergäbe sonst
	// eine andere Summe als vor Phase 3.2, und sie steht im Paketkatalog.
	if _, fehler := io.Copy(io.Discard, mitlesend); fehler != nil {
		return nil, fmt.Errorf("Paket beziehen: %w", fehler)
	}

	return &Ergebnis{
		Stand:       quelle.Stand,
		Pruefsumme:  "sha256:" + hex.EncodeToString(rechner.Sum(nil)),
		Uebergangen: uebergangen,
	}, nil
}

// standPruefen weist ab, was kein Stand ist.
//
// **Ein Zweig ist keine Fassung.** `main` ändert sich; eine Prüfsumme darauf
// wäre eine Momentaufnahme, und zwei Bezüge derselben Angabe ergäben
// verschiedene Pakete. Anerkannt sind deshalb genau zwei Formen: ein
// vierzigstelliger Commit-Hash — der ist unveränderlich — und eine Marke, die
// die Gegenstelle als Freigabe kennt.
func standPruefen(ctx context.Context, quelle Repositorium, kenner string) error {
	if quelle.Stand == "" {
		return domain.NewFieldError("stand", domain.SchluesselPflichtfeld,
			"Es ist kein Stand angegeben. Ein Zweig ist keiner — bezogen wird eine Freigabe oder ein Commit.")
	}

	if commitform.MatchString(quelle.Stand) {
		return nil
	}

	adresse := quelle.Marken + quelle.Stand

	// **Über `netz.Hole`, seit dem 2026-08-23** (#1005). Die Kopfzeilen, die
	// Fristen und die Erkennung eines 429 stehen dort für alle Dienste; hier
	// bleibt, was **dieser** Weg bedeutet.
	antwort, err := netz.Hole(ctx, klient(), nil, kennung(kenner), adresse, hoechstens)
	if err != nil {
		// **Die Adresse ist das Feld, nicht der Stand** (#2000): gesperrt ist
		// das Repositorium, und wer es ändert, ändert die Adresse.
		if fehler := alsZielbefund(err, "adresse", quelle.Adresse); fehler != nil {
			return fehler
		}

		// **Auch hier** (#917): sie läuft **vor** dem Archivbezug, und ohne
		// diese Übersetzung las eine gedrosselte Person „das ist keine
		// Freigabe" — und suchte den Fehler an ihrer Eingabe.
		if fehler := alsDrosselbefund(err, "stand", adresse); fehler != nil {
			return fehler
		}

		return fmt.Errorf("Stand nachschlagen: %w", err)
	}

	if antwort.Status != netz.StatusGut {
		return domain.NewFieldError("stand", domain.SchluesselValidierung,
			"%q ist keine Freigabe dieses Repositoriums. Ein Zweig ist kein Stand: "+
				"er ändert sich, und dieselbe Angabe ergäbe morgen ein anderes Paket.",
			quelle.Stand)
	}

	return nil
}
