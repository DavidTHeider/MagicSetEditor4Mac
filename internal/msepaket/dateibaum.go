// **Dateien im Paket anlegen, umbenennen und entfernen** — FMT-02.AK-01
// (#1691).
//
// # Keine Entsprechung in MSE2 — **new**
//
// MSE2 hat keinen Paketeditor: gesucht in `src/util/io/package.cpp` (die
// Klasse `Package` schreibt ein **offenes** Paket als Ganzes, `saveAs`,
// `saveToDirectory`, `saveToZipfile`) und in `src/gui/`, Fassung f9d9356 vom
// 2021-07-06. Die Regeln hier sind deshalb die des Schreibwegs
// (`schreibweg.go`) und keine neuen:
//
//   - **dieselbe Sperre** am Paket ([beimSchreiben]),
//   - **der Stand wird geprüft**, wo einer genannt ist,
//   - **was verschwindet, liegt danach in der Sicherung** — `<paket>.bak/`,
//     neben dem Paket und nicht darin.
//
// # Nur Verzeichnispakete
//
// **Ein gepacktes Paket wird nicht umgebaut.** Jede dieser Operationen wäre
// dort ein neues Archiv; wer an einem gepackten Paket arbeiten will, legt
// eine eigene Kopie an, und die ist ein Verzeichnis (`service`, `Kopieren`).
//
// # Die Hauptdatei bleibt
//
// Ohne `style`, `game` oder `export-template` ist ein Paket keines mehr —
// `kopfLesen` scheitert, und die Werkstatt fände es nicht wieder. Sie lässt
// sich deshalb weder entfernen noch umbenennen.

package msepaket

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// DateiAnlegen legt eine neue Datei im Paket an — eine vorhandene wird nicht
// überschrieben ([domain.ErrConflict]).
//
// **Über die Nachbardatei**, wie der Schreibweg: wer mittendrin abbricht,
// hinterlässt keine halbe Datei unter dem neuen Namen.
func DateiAnlegen(ctx context.Context, ort, name string, inhalt []byte) error {
	return dateiAnlegen(ctx, ort, name, inhalt, echteHilfen())
}

func dateiAnlegen(ctx context.Context, ort, name string, inhalt []byte, hilfen schreibhilfen) error {
	sauber, fertig, err := amVerzeichnis(ctx, ort, name)
	if err != nil {
		return err
	}

	defer fertig()

	ziel := filepath.Join(ort, filepath.FromSlash(sauber))
	if fehler := frei(ziel, sauber, ort); fehler != nil {
		return fehler
	}

	if fehler := os.MkdirAll(filepath.Dir(ziel), 0o750); fehler != nil {
		return fmt.Errorf("den Ordner für %s anlegen: %w", sauber, fehler)
	}

	neben, fehler := hilfen.nachbar(filepath.Dir(ziel))
	if fehler != nil {
		return fmt.Errorf("die Nachbardatei zu %s anlegen: %w", sauber, fehler)
	}

	defer func() { _ = os.Remove(neben.Name()) }() //nolint:errcheck // nach dem Tausch ist sie fort

	if fehler := ablegen(neben, inhalt); fehler != nil {
		return fmt.Errorf("%s schreiben: %w", sauber, fehler)
	}

	return hilfen.umbenennen(neben.Name(), ziel)
}

// DateiUmbenennen gibt einer Datei einen neuen Namen, auch in einem anderen
// Ordner des Pakets.
//
// `erwartet` ist der Stand, den der Aufrufer gelesen hat; **leer heißt ohne
// Prüfung** — für Beiwerk wie ein Bild, das die Werkstatt nicht liest und
// für das sie deshalb keinen Stand hat.
func DateiUmbenennen(ctx context.Context, ort, alt, neu, erwartet string) error {
	return dateiUmbenennen(ctx, ort, alt, neu, erwartet, echteHilfen())
}

func dateiUmbenennen(ctx context.Context, ort, alt, neu, erwartet string, hilfen schreibhilfen) error {
	neuSauber, err := paketpfad(neu)
	if err != nil {
		return err
	}

	altSauber, fertig, err := amVerzeichnis(ctx, ort, alt)
	if err != nil {
		return err
	}

	defer fertig()

	quelle, fehler := vorhanden(ort, altSauber, erwartet)
	if fehler != nil {
		return fehler
	}

	ziel := filepath.Join(ort, filepath.FromSlash(neuSauber))
	if fehler := frei(ziel, neuSauber, ort); fehler != nil {
		return fehler
	}

	if fehler := os.MkdirAll(filepath.Dir(ziel), 0o750); fehler != nil {
		return fmt.Errorf("den Ordner für %s anlegen: %w", neuSauber, fehler)
	}

	return hilfen.umbenennen(quelle, ziel)
}

// DateiEntfernen nimmt eine Datei aus dem Paket — **in die Sicherung**, wo
// sie sich zurückholen lässt. `erwartet` wie bei [DateiUmbenennen].
func DateiEntfernen(ctx context.Context, ort, name, erwartet string) error {
	return dateiEntfernen(ctx, ort, name, erwartet, echteHilfen())
}

func dateiEntfernen(ctx context.Context, ort, name, erwartet string, hilfen schreibhilfen) error {
	sauber, fertig, err := amVerzeichnis(ctx, ort, name)
	if err != nil {
		return err
	}

	defer fertig()

	quelle, fehler := vorhanden(ort, sauber, erwartet)
	if fehler != nil {
		return fehler
	}

	sicherung := filepath.Join(ort+Sicherungsendung, filepath.FromSlash(sauber))
	if fehler := os.MkdirAll(filepath.Dir(sicherung), 0o700); fehler != nil {
		return fmt.Errorf("den Rückweg zu %s anlegen: %w", sauber, fehler)
	}

	return hilfen.umbenennen(quelle, sicherung)
}

// amVerzeichnis prüft Name und Paket und nimmt die Schreibsperre.
func amVerzeichnis(ctx context.Context, ort, name string) (sauber string, fertig func(), fehler error) {
	sauber, fehler = paketpfad(name)
	if fehler != nil {
		return "", nil, fehler
	}

	angaben, statfehler := os.Stat(ort)
	if statfehler != nil {
		return "", nil, domain.MitGrund(domain.ErrNotFound, statfehler,
			"das Paket %q ist nicht zu finden", ort)
	}

	if !angaben.IsDir() {
		return "", nil, fmt.Errorf("%w: %q ist gepackt — Dateien anlegen, umbenennen und entfernen "+
			"geht an einer eigenen Kopie", domain.ErrValidation, filepath.Base(ort))
	}

	fertig, fehler = beimSchreiben(ctx, ort)
	if fehler != nil {
		return "", nil, fehler
	}

	return sauber, fertig, nil
}

// vorhanden gibt den Pfad einer Datei, die umbenannt oder entfernt werden
// darf: sie steht da, ist kein Ordner und nicht die Hauptdatei, und ihr Stand
// ist der erwartete.
func vorhanden(ort, name, erwartet string) (string, error) {
	if name == ArtAusName(filepath.Base(ort)).Hauptdatei() {
		return "", fmt.Errorf("%w: %s ist die Hauptdatei — ohne sie ist %q kein Paket mehr",
			domain.ErrValidation, name, filepath.Base(ort))
	}

	pfad := filepath.Join(ort, filepath.FromSlash(name))

	angaben, err := os.Lstat(pfad)
	if err != nil {
		return "", domain.MitGrund(domain.ErrNotFound, err, "%s steht nicht in %q", name, ort)
	}

	if angaben.IsDir() {
		return "", fmt.Errorf("%w: %s ist ein Ordner und keine Datei", domain.ErrValidation, name)
	}

	if erwartet == "" {
		return pfad, nil
	}

	da, lesefehler := LeseDateiBegrenzt(pfad)
	if lesefehler != nil {
		return "", fmt.Errorf("%s in %q lesen: %w", name, ort, lesefehler)
	}

	return pfad, standPruefen(da, name, ort, erwartet)
}

// frei sagt, ob unter einem Namen noch nichts steht.
//
// **Drei Antworten, nicht zwei**: steht dort etwas, ist das ein Konflikt;
// fehlt es, ist der Name frei; und scheitert schon die Frage — `style/neu`,
// wo `style` eine Datei ist —, dann ist das weder das eine noch das andere.
func frei(ziel, name, ort string) error {
	_, err := os.Lstat(ziel)

	switch {
	case err == nil:
		return fmt.Errorf("%w: %s steht in %q schon da", domain.ErrConflict, name, filepath.Base(ort))
	case errors.Is(err, fs.ErrNotExist):
		return nil
	default:
		// **Der Grund fürs Protokoll, der Satz für Menschen** (0075) — ein
		// `ENOTDIR` mit Pfad gehört nicht in die Oberfläche.
		return domain.MitGrund(domain.ErrValidation, err,
			"%s lässt sich in %q nicht anlegen — ein Teil des Wegs ist eine Datei", name, filepath.Base(ort))
	}
}
