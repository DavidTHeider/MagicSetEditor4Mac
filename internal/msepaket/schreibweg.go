// **Eine Datei im Paket zurückschreiben** — Werkstück 2 von Phase 3.3
// (#1164).
//
// # Portiert aus MSE2
//
// `Package::saveToZipfile` und `Package::saveToDirectory`
// (`src/util/io/package.cpp`, Zeilen 402–491, Fassung f9d9356 vom
// 2021-07-06). Drei Züge, und ihre Reihenfolge ist die ganze Zusage:
//
//  1. In eine **Nachbardatei** schreiben.
//  2. Den alten Stand nach `.bak` umbenennen.
//  3. Die Nachbardatei an seinen Platz umbenennen.
//
// Unveränderte Einträge eines Archivs werden **als Einträge** kopiert
// (`newZip->CopyEntry`), nicht ausgepackt und neu gepackt. Das ist die
// Rundlaufzusage: was niemand angefasst hat, hat danach dieselben Bytes.
//
// # Was hier anders ist als in MSE2 — **adapted**
//
// **Ein Verzeichnispaket bekommt auch eine `.bak`.** MSE2 löscht dort die
// alte Datei und schiebt die neue darüber (`saveToDirectory`, Zeilen 417–423).
// Das Issue verlangt „nichts überschreiben, was nicht zurückzuholen ist", und
// der Unterschied liegt im Anlass: wer in MSE2 eine Vorlage ändert, tut es mit
// einem Texteditor und hat dessen Sicherung; wer sie hier im Fenster ändert,
// hätte gar nichts.
//
// **Sie liegt neben dem Paket und nicht darin** — `<paket>.mse-style.bak/`.
// Warum, steht bei [insVerzeichnis]: eine Sicherung im Paketverzeichnis wäre
// eine Datei des Pakets.
//
// **Der erwartete Stand wird geprüft.** MSE2 kennt keinen Konfliktweg für
// Paketdateien — es hat auch keinen Editor dafür (gesucht in `src/gui/set/*`,
// `src/gui/symbol/*`, in jeder Klasse mit `public Packaged` und in jedem
// Aufrufer von `save()`/`saveAs()`; geschrieben werden `.mse-set` über
// `Package::save` und `.mse-symbol` über einen schlichten `Writer`). Ein
// Fenster, das eine Datei geöffnet hält, während jemand anderes sie ändert,
// ist dagegen der Normalfall — §7, Regel 2.
//
// # Warum Prüfen und Schreiben denselben Griff benutzen
//
// Zwischen einem Prüfen, das sein Paket wieder schließt, und einem Schreiben,
// das es neu öffnet, liegt ein Fenster, in dem jemand anderes schreiben kann.
// Der Konfliktschutz sagte dann ja zu einem Stand, der beim Schreiben schon
// nicht mehr dasteht.

package msepaket

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sicherungsendung ist, woran der vorige Stand zu finden ist.
//
// **`.bak` und nicht etwas Eigenes**: MSE2 legt seine Sicherung so ab, und wer
// beide Programme benutzt, soll nicht zwei Endungen kennen müssen.
const Sicherungsendung = ".bak"

// DateiSchreiben ersetzt eine Datei im Paket.
//
// `erwartet` ist die SHA-256 des Inhalts, den der Aufrufer gelesen hat, in
// Kleinbuchstaben. Stimmt sie nicht mehr mit dem, was dasteht, gibt es
// [domain.ErrConflict] — **und es ist nichts geschrieben**.
func DateiSchreiben(ctx context.Context, ort, name string, inhalt []byte, erwartet string) error {
	return dateiSchreiben(ctx, ort, name, inhalt, erwartet, echteHilfen())
}

// dateiSchreiben ist [DateiSchreiben] mit eingesetzten Griffen.
//
// **Der Weg als Ganzes wird damit betretbar, nicht nur [austauschen].** Die
// Begründung dort gilt eine Ebene höher genauso: einen Fehlschlag mitten im
// Tausch liefert kein Dateisystem auf Bestellung — und was danach im
// Paketverzeichnis liegt, entscheidet sich nicht in `austauschen`, sondern
// zwischen Sicherung, Nachbardatei und Aufräumen. Was daran hängt, steht in
// `fehlerinjektion_intern_test.go` und `ablegeabbruch_intern_test.go` (#1289).
func dateiSchreiben(ctx context.Context, ort, name string, inhalt []byte,
	erwartet string, hilfen schreibhilfen,
) error {
	sauber, err := paketpfad(name)
	if err != nil {
		return err
	}

	angaben, statfehler := os.Stat(ort)
	if statfehler != nil {
		return domain.MitGrund(domain.ErrNotFound, statfehler,
			"das Paket %q ist nicht zu finden", ort)
	}

	// **Lesen, Prüfen und Tauschen unter einer Sperre.** Warum, steht in
	// `schreibsperre.go` — ohne sie kamen von sechs gleichzeitigen Schreibern
	// zwei bis vier durch.
	fertig, sperrfehler := beimSchreiben(ctx, ort)
	if sperrfehler != nil {
		return sperrfehler
	}

	defer fertig()

	if angaben.IsDir() {
		return insVerzeichnis(ort, sauber, inhalt, erwartet, hilfen)
	}

	return insArchiv(ort, sauber, inhalt, erwartet, hilfen)
}

// standPruefen weist ab, was inzwischen jemand anderes geändert hat.
func standPruefen(da []byte, name, ort, erwartet string) error {
	roh := sha256.Sum256(da)
	if gefunden := hex.EncodeToString(roh[:]); gefunden != erwartet {
		return fmt.Errorf("%w: %s in %q hat sich seit dem Öffnen geändert",
			domain.ErrConflict, name, ort)
	}

	return nil
}

// insVerzeichnis schreibt eine Datei in ein Verzeichnispaket.
//
// **Die Nachbardatei entsteht über [os.CreateTemp]** und nicht unter einem
// ausgerechneten Namen: zwei Läufe zugleich träfen sich sonst darin, und was
// am Ende dastünde, wäre die Mischung zweier Stände (§7, Regel 4).
func insVerzeichnis(ort, name string, inhalt []byte, erwartet string,
	hilfen schreibhilfen,
) error {
	ziel := filepath.Join(ort, filepath.FromSlash(name))

	// **Mit Budget** (Befund 4 des Zweitblicks über Phase 6.1): verglichen
	// wird mit dem, was dasteht, und das kann seit dem Öffnen alles Mögliche
	// geworden sein.
	da, lesefehler := LeseDateiBegrenzt(ziel)
	if errors.Is(lesefehler, ErrZuGross) {
		return lesefehler
	}

	if lesefehler != nil {
		return domain.MitGrund(domain.ErrNotFound, lesefehler,
			"%s in %q ließ sich nicht lesen", name, ort)
	}

	if fehler := standPruefen(da, name, ort, erwartet); fehler != nil {
		return fehler
	}

	// **Der vorige Stand kommt neben das Paket, nicht hinein.**
	//
	// Gefunden am 2026-08-26 durch den Lauf über 526 bezogene Pakete
	// (`fremdschreibweg_test.go`): eine `style.bak` **im** Verzeichnis ist
	// danach eine Datei **des Pakets** — `Dateien()` zählte sie mit, aus 24
	// wurden 25. Bei einer `.mse-symbol-font`, deren Dateien alle etwas
	// bedeuten, wäre das kein Schönheitsfehler.
	//
	// `<paket>.mse-style.bak/` trägt keine Paketendung mehr und ist damit für
	// [ArtAusName] kein Paket — dieselbe Lage wie beim Archiv, wo `.bak` an
	// die Datei tritt.
	sicherung := filepath.Join(ort+Sicherungsendung, filepath.FromSlash(name))
	if fehler := os.MkdirAll(filepath.Dir(sicherung), 0o700); fehler != nil {
		return fmt.Errorf("den Rückweg zu %s anlegen: %w", name, fehler)
	}

	// **Die Nachbardatei bleibt im Paket.** `os.Rename` ist nur innerhalb
	// eines Dateisystems atomar, und das Paketverzeichnis ist der einzige Ort,
	// von dem das sicher gilt. Ein Absturz mitten im Zug lässt eine
	// `.mse-schreibt-…` liegen — MSE2 hinterlässt an derselben Stelle eine
	// `.tmp` (`saveToZipfile`, Zeile 447).
	neben, err := hilfen.nachbar(filepath.Dir(ziel))
	if err != nil {
		return fmt.Errorf("die Nachbardatei zu %s anlegen: %w", name, err)
	}

	// **Weggeräumt und geschlossen wird auf jedem Weg.** Nach einem geglückten
	// Umbenennen gibt es die Nachbardatei nicht mehr, und `os.Remove` sagt das
	// als `ErrNotExist`; ein zweites `Close` gibt `ErrClosed`. Beides ist kein
	// Grund für eine Meldung — ein **offener Griff** nach einem Fehlschlag
	// dagegen schon, und den verhindert diese Zeile.
	defer func() {
		_ = neben.Close()           //nolint:errcheck // siehe darüber
		_ = os.Remove(neben.Name()) //nolint:errcheck // dito
	}()

	// **Schreiben, Schließen und die beiden Umbenennungen auf einem Weg.**
	//
	// Sie scheitern an derselben Ursache — der Ort nimmt die Datei nicht —,
	// und getrennte Zweige wären welche, die kein Lauf betritt: ein Griff auf
	// eine eben angelegte Datei scheitert beim Schreiben und beim Schließen
	// nicht auf eine Art, die sich herbeiführen ließe (§3). Dasselbe Muster
	// steht in `internal/netz/speicher.go` und dort mit derselben Begründung.
	fehler := ablegen(neben, inhalt)
	if fehler == nil {
		fehler = austauschen(ziel, sicherung, neben.Name(), hilfen.umbenennen)
	}

	if fehler != nil {
		return fmt.Errorf("%s in %q einsetzen: %w", name, ort, fehler)
	}

	return nil
}

// austauschen setzt die Nachbardatei an den Platz des Ziels.
//
// # Zwei Umbenennungen, und die zweite kann scheitern
//
// Erst geht das Ziel in die Sicherung, dann die Nachbardatei an seinen Platz.
// Scheitert der **zweite** Zug, war der erste schon getan: das Ziel gibt es
// nicht mehr, die Nachbardatei räumt der Aufrufer weg — und übrig bliebe ein
// Paket ohne diese Datei, während der vorige Stand unter `.bak` liegt und
// niemand ihn dort sucht.
//
// **Deshalb wird zurückgeholt.** Gelingt auch das nicht, sagt die Meldung, wo
// der vorige Stand liegt; das ist dann das Einzige, was noch hilft.
//
// **Und das ist die Grenze dieser Rücknahme**: nach einem gescheiterten
// **dritten** Zug fehlt das Ziel weiterhin, und der vorige Stand liegt unter
// `.bak`. Weiter kommt man mit Umbenennungen nicht — wer hier ankommt, hat ein
// Dateisystem, das zwei aufeinanderfolgende Umbenennungen in **einem**
// Verzeichnis ablehnt. Angestrichen vom Nachlauf des Zweitblicks über
// Phase 3.3.
//
// **`umbenennen` kommt herein**, damit dieser Weg zu betreten ist: einen
// Fehlschlag genau zwischen zwei Umbenennungen kann kein Dateisystem auf
// Bestellung liefern (§3). Angestrichen vom Zweitblick durch Codex über
// Phase 3.3.
func austauschen(ziel, sicherung, neben string, umbenennen func(alt, neu string) error) error {
	if fehler := umbenennen(ziel, sicherung); fehler != nil {
		return fehler
	}

	fehler := umbenennen(neben, ziel)
	if fehler == nil {
		return nil
	}

	if zurueck := umbenennen(sicherung, ziel); zurueck != nil {
		return fmt.Errorf("%w — und der vorige Stand liegt jetzt in %q: %w",
			fehler, sicherung, zurueck)
	}

	return fehler
}

// insArchiv schreibt eine Datei in ein Archivpaket.
//
// **Jeder andere Eintrag wird über [zip.Writer.Copy] übernommen** — roh, ohne
// Auspacken und ohne neues Packen. Das ist MSEs `CopyEntry` und der Grund,
// warum ein Rundlauf danach noch stimmt.
func insArchiv(ort, name string, inhalt []byte, erwartet string,
	hilfen schreibhilfen,
) error {
	lesen, griff, err := oeffneArchiv(ort)
	if err != nil {
		return fmt.Errorf("Paket %q als Archiv öffnen: %w", ort, err)
	}

	offen := true

	defer func() {
		if offen {
			_ = griff.Close() //nolint:errcheck // nur gelesen
		}
	}()

	da, lesefehler := gelesen(lesen, name)
	if errors.Is(lesefehler, ErrZuGross) {
		return lesefehler
	}

	if lesefehler != nil {
		return domain.MitGrund(domain.ErrNotFound, lesefehler,
			"%s in %q ließ sich nicht lesen", name, ort)
	}

	if fehler := standPruefen(da, name, ort, erwartet); fehler != nil {
		return fehler
	}

	neben, tempfehler := hilfen.nachbar(filepath.Dir(ort))
	if tempfehler != nil {
		return fmt.Errorf("die Nachbardatei zu %q anlegen: %w", ort, tempfehler)
	}

	// **Weggeräumt und geschlossen wird auf jedem Weg.** Nach einem geglückten
	// Umbenennen gibt es die Nachbardatei nicht mehr, und `os.Remove` sagt das
	// als `ErrNotExist`; ein zweites `Close` gibt `ErrClosed`. Beides ist kein
	// Grund für eine Meldung — ein **offener Griff** nach einem Fehlschlag
	// dagegen schon, und den verhindert diese Zeile.
	defer func() {
		_ = neben.Close()           //nolint:errcheck // siehe darüber
		_ = os.Remove(neben.Name()) //nolint:errcheck // dito
	}()

	// **Erst das Archiv schließen, dann umbenennen.** Unter Windows ist eine
	// offene Datei nicht zu ersetzen (`docs/PORTIERUNG.md`); unter macOS ginge
	// es, und genau deshalb fiele es hier nie auf.
	fehler := archivBauen(neben, lesen, name, inhalt)
	if fehler == nil {
		fehler = ablegen(neben, nil)
	}

	if fehler == nil {
		offen = false
		fehler = hilfen.schliessen(griff)
	}

	if fehler == nil {
		fehler = austauschen(ort, ort+Sicherungsendung, neben.Name(), hilfen.umbenennen)
	}

	if fehler != nil {
		return fmt.Errorf("%s in %q einsetzen: %w", name, ort, fehler)
	}

	return nil
}

// gelesen holt einen Eintrag aus einem geöffneten Archiv — mit demselben
// Budget wie der Leser: ein Eintrag, der ausgepackt mehr wird, als sein Kopf
// sagt, hält an der Grenze an (Befund 4 des Zweitblicks über Phase 6.1).
func gelesen(archiv *zip.Reader, name string) ([]byte, error) {
	return leseBegrenzt(archiv, name)
}

// archivBauen schreibt das neue Archiv.
//
// **Es nimmt einen [io.Writer] und keine Datei**, damit seine Fehlerwege
// betretbar sind: ein Schreiber, der ablehnt, führt jeden einzelnen herbei.
// Dieselbe Gestalt wie `sicherung.hineinschreiben` und aus demselben Grund.
func archivBauen(nach io.Writer, lesen *zip.Reader, name string, inhalt []byte) error {
	packer := zip.NewWriter(nach)

	for _, eintrag := range lesen.File {
		if eintrag.Name == name {
			continue
		}

		if err := packer.Copy(eintrag); err != nil {
			return fmt.Errorf("%s übernehmen: %w", eintrag.Name, err)
		}
	}

	// **Anlegen und Schreiben auf einem Weg** — dieselbe Überlegung wie in
	// `sicherung.teilSchreiben`: `Create` meldet, was beim Abschluss des
	// **vorigen** Eintrags schiefging, und beim ersten also nie.
	ziel, err := packer.Create(name)
	if err == nil {
		_, err = ziel.Write(inhalt)
	}

	if err == nil {
		err = packer.Close()
	}

	if err != nil {
		return fmt.Errorf("%s ins Archiv legen: %w", name, err)
	}

	return nil
}
