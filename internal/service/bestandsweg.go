// **Der Sicherungsweg** — sichern und zurückspielen, mit Reihenfolge,
// Rücknahme, Fassungsregel und Marke
// (#1175).
//
// # Warum er hier steht und nicht mehr in `internal/app`
//
// **Weil er entscheidet.** §2 weist Anwendungsfälle, Transaktionsgrenzen und
// fachliche Regeln `internal/service` zu und verbietet fachliche
// Entscheidungen in der Fassade. Der ganze Weg lag seit
// #940
// trotzdem dort und entschied: was gemeinsam gesichert wird, welche
// Archivfassungen zulässig sind, in welcher Reihenfolge der zweiteilige
// Tausch läuft und wann ein abgebrochener nachgeholt wird.
//
// **Der Schichtenwächter erlaubte die Kante `app → sicherung` ausdrücklich**
// — er hielt damit die Ausnahme und nicht die Regel. Sie ist mit diesem
// Umbau weggefallen.
//
// # Was in der Fassade bleibt
//
// **Die Orte und der Lebenszyklus.** `internal/app` ist der einzige Platz,
// der Datenbankpfad und Bilderverzeichnis zugleich kennt, und der einzige,
// der die Verbindung öffnet und schließt. Beides wird **hineingereicht**,
// nicht hier entschieden — dieselbe Gestalt wie beim Ausschneiden der
// Zwischenablage, wo das Ablegen hineingereicht wird
// (#1244).
//
// **Und der Dateidialog.** Er ist der Grund, warum Sichern und Zurückspielen
// bis heute nur im Fenster stehen (§16); er bleibt in `app`, weil es ihn
// ohne Fenster nicht gibt.

package service

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Sicherungsendung ist die Dateiendung einer Sicherung.
//
// **Sie steht hier, damit `internal/app` sie für den Dateidialog nennen kann,
// ohne `internal/sicherung` zu kennen** — die Kante, die dieser Umbau
// wegnimmt. Ein Muster im Dialog ist keine fachliche Entscheidung; woraus es
// besteht, schon.
const Sicherungsendung = sicherung.Endung

// Sicherungsorte sind die Plätze, die nur die Fassade kennt.
//
// **Zwei Angaben und keine Abkürzung**: das Bilderverzeichnis ist eine
// Funktion, weil es sich beim ersten eingesetzten Bild erst herstellt und ein
// Test es umlenken können muss (§3).
type Sicherungsorte struct {
	// Datenbank ist der Pfad der laufenden Bestandsdatei.
	Datenbank string

	// Bilder gibt das Verzeichnis der eingesetzten Kartenbilder.
	Bilder func() (string, error)
}

// Bestandsweg führt Sicherung und Rückspielung.
type Bestandsweg struct {
	orte     Sicherungsorte
	pruefer  Sicherungspruefer
	allein   Alleinpruefung
	sicherer Bestandssicherer

	// schliessen und oeffnen sind der Lebenszyklus der Verbindung.
	//
	// **Sie gehören der Fassade** — sie hält die Verbindung, und ein Dienst,
	// der sie selbst öffnete, entschiede über den Prozess statt über den
	// Vorgang. Hier stehen sie als Naht: der Weg braucht sie in einer
	// bestimmten **Reihenfolge**, und die ist die fachliche Regel.
	schliessen func() error
	oeffnen    func(ctx context.Context) error

	// fassung ist die Fassung, die ins Archiv geschrieben wird.
	fassung string

	// chiffre ist die Verschlüsselung — leer, wo eine Fläche sie nicht kennt.
	chiffre Bestandschiffre

	// Die drei Nähte des Rücksprungvergleichs (O7, #1431) — leer, wo eine
	// Fläche ihn nicht kennt.
	schau Bestandsschau
	eigen Eigenstand
	buch  Rueckkehrbuch

	// uhr sagt, wann eine Rückkehr geschah.
	//
	// **Eine Naht und kein `time.Now()` mitten im Weg**: ein Lauf, der den
	// Zeitpunkt der Epoche prüft, prüfte sonst die Uhr dieser Maschine.
	uhr func() time.Time

	// pakete und paketordner nennen die Vorlagenpakete für die Sicherung —
	// OPS-06.AK-01, siehe [Bestandsweg.MitPaketen].
	pakete      Paketquelle
	paketordner func() (string, error)

	// vermerke führt die Vermerke der Sicherungen — OPS01, siehe
	// [Bestandsweg.MitSicherungsbuch].
	vermerke Sicherungsbuch
}

// NeuerBestandsweg baut ihn.
func NeuerBestandsweg(
	orte Sicherungsorte, pruefer Sicherungspruefer, allein Alleinpruefung,
	sicherer Bestandssicherer, schliessen func() error,
	oeffnen func(ctx context.Context) error, fassung string,
) *Bestandsweg {
	return &Bestandsweg{
		orte: orte, pruefer: pruefer, allein: allein, sicherer: sicherer,
		schliessen: schliessen, oeffnen: oeffnen, fassung: fassung,
		uhr: time.Now,

		// **Die Vorgabe der zwei Leser sagt „ich kann nicht nachsehen"** —
		// der Grund steht bei [unbekannterStand].
		schau: func(context.Context, string) (domain.Bestandsstand, error) {
			return unbekannterStand(), nil
		},
		eigen: func(context.Context) domain.Bestandsstand { return unbekannterStand() },
	}
}

// Bestandssicherer schreibt den laufenden Bestand als Datei.
//
// **Die Schnittstelle steht hier, weil dieser Anwendungsfall sie braucht**
// (§2). Erfüllt wird sie von `storage.Sichern`; verdrahtet im Composition Root.
type Bestandssicherer interface {
	Sichern(ctx context.Context, ziel string) error
}

// MitBestandssicherer setzt die Naht — für Tests und den Composition Root.
func (b *Bestandsweg) MitBestandssicherer(sicherer Bestandssicherer) { b.sicherer = sicherer }
