package service

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Sperre über einem Bezugsbaum** —
// #1190.
//
// # Zwei Sperren, die einander nicht sahen
//
// `msepaket.DateiSchreiben` nimmt eine Sperre neben der **Paketdatei** —
// `…/data/magic-m15.mse-style.schreibsperre` — und hält sie über Lesen,
// Prüfen und Tauschen. Alle vier Flächen der Paketwerkstatt gehen da durch.
//
// `Paketbezieher.Beziehen` tauscht dagegen einen **Baum** aus:
// `pakete/M15-Magic-Pack/`. Die beiden liegen auf verschiedenen Ebenen und
// wussten nichts voneinander — wer gerade eine Datei darunter bearbeitete,
// verlor sie ohne Meldung.
//
// Seit dem 2026-08-28 wird der bisherige Stand beiseitegelegt statt gelöscht
// (#1170,
// Befund 3). Für jemanden, der gerade schreibt, ist das derselbe Verlust: die
// Datei, die er kennt, liegt danach woanders.
//
// # Warum die Ordnung hier steht und nicht in `msepaket`
//
// **Weil hier die Wurzeln bekannt sind.** `msepaket` weiß, wie ein Paket auf
// der Platte aussieht; **wo** die Bezugsbäume liegen, entscheidet der
// Composition Root und reicht es dem Dienst. Genau so steht es im Zuschnitt
// des Issues.
//
// # Die Richtung
//
//   - Der **Bezug** nimmt die **ausschließliche** Sperre am Bezugsziel: er
//     tauscht den ganzen Baum.
//   - Ein **Schreiber** nimmt zusätzlich eine **geteilte** am umgebenden
//     Bezugsbaum: mehrere dürfen gleichzeitig schreiben, ein Bezug nicht
//     dazwischen.
//
// # Was sie nicht hält
//
// **Einen Schreiber, der die Wurzel nicht kennt.** Wer eine Paketdatei
// unmittelbar über `msepaket` schreibt, ohne über [MitBezugssperre] zu gehen,
// nimmt die Baumsperre nicht — die Flächen tun es, ein fremdes Werkzeug nicht.
// Das ist dieselbe Grenze wie bei der Schreibsperre selbst: sie ist eine
// **Absprache**, keine Zwangsmaßnahme des Dateisystems.

// Bezugssperrendung hängt an den Ort eines bezogenen Pakets.
//
// **Neben dem Baum und nicht darin.** Eine Datei **im** Baum verschwände mit
// ihm beim Tausch — und damit die Sperre mitten im Zug.
const Bezugssperrendung = ".bezugssperre"

// bezugsfrist ist, wie lange ein Bezug auf laufende Schreiber wartet.
//
// **Dieselbe Frist wie beim Schreiben eines Pakets** (`msepaket.Schreibfrist`)
// und aus demselben Grund: wer länger braucht, schreibt nicht mehr.
var bezugsfrist = 10 * time.Second

// bezugstakt ist der Abstand zwischen zwei Versuchen.
var bezugstakt = 20 * time.Millisecond

// MitBezugsfrist setzt die Wartezeit — **für Tests**.
//
// **Eine Naht und kein Parameter**: die Frist ist eine Eigenschaft des
// Betriebs und keine Wahl des Aufrufers; ein Lauf, der zehn Sekunden wartete,
// wäre keiner, den jemand fährt.
func MitBezugsfrist(frist time.Duration) (zurueck func()) {
	vorher := bezugsfrist
	bezugsfrist = frist

	return func() { bezugsfrist = vorher }
}

// alleinAmBaum nimmt die ausschließliche Sperre über einem Bezugsziel.
func alleinAmBaum(ort string) (fertig func(), fehler error) {
	return amBaum(ort, func(griff *dateisperre.Griff) error {
		return griff.AlleinBis(context.Background(), bezugsfrist, bezugstakt)
	})
}

// GeteiltAmBaum nimmt die geteilte Sperre über einem Bezugsziel.
//
// **Für Schreiber**: mehrere dürfen gleichzeitig an Dateien darunter
// arbeiten; ein Bezug kommt in dieser Zeit nicht dazwischen.
//
// **Der Ort ist der Baum, nicht die Datei.** Wer `…/pakete/M15-Magic-Pack`
// nennt, sperrt gegen einen Bezug desselben; wer die Datei nennt, sperrt gegen
// nichts.
func GeteiltAmBaum(ctx context.Context, ort string) (fertig func(), fehler error) {
	return amBaum(ort, func(griff *dateisperre.Griff) error {
		return griff.TeilenBis(ctx, bezugsfrist, bezugstakt)
	})
}

// amBaum ist der gemeinsame Rumpf beider Richtungen.
func amBaum(
	ort string, nimm func(*dateisperre.Griff) error,
) (fertig func(), fehler error) {
	griff, err := dateisperre.Oeffnen(ort + Bezugssperrendung)
	if err != nil {
		return nil, fmt.Errorf("die Bezugssperre zu %q nehmen: %w", ort, err)
	}

	if wartefehler := nimm(griff); wartefehler != nil {
		_ = griff.Schliessen() //nolint:errcheck // der Griff geht ohnehin weg

		return nil, domain.MitGrund(domain.ErrConflict, wartefehler,
			"an %q schreibt gerade jemand anderes — versuche es gleich noch einmal",
			ort)
	}

	return func() {
		// **Die Sperrdatei bleibt liegen.**
		//
		// Hier stand ein `os.Remove`, mit der Begründung, die Datei sei ein
		// Hilfsmittel und kein Bestandteil des Pakets. Das stimmt — und die
		// Sperre **zerfiel** daran.
		//
		// `flock` hängt an der **Inode**, nicht am Pfad. Ein Wartender hält
		// die alte Inode schon offen; wird die Datei entfernt, legt der
		// nächste Öffner am selben Pfad eine **neue** an und sperrt die. Dann
		// halten zwei Prozesse gleichzeitig „die" Sperre, und Paketbezug und
		// Paketwerkstatt schreiben zugleich in denselben Baum.
		//
		// Angestrichen vom Nachlauf über die Nacharbeit
		// (#1282,
		// Befund 5).
		//
		// **Der Preis ist eine leere Datei neben jedem bezogenen Paket.** Das
		// ist billiger als eine Sperre, die unter Last keine ist.
		_ = griff.Schliessen() //nolint:errcheck // die Sperre fällt mit dem Griff
	}, nil
}
