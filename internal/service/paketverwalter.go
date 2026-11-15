package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Paketablage führt den Katalog der installierten Pakete.
//
// **Die Schnittstelle steht hier, erfüllt wird sie in `storage`** (§2).
type Paketablage interface {
	Abgleichen(ctx context.Context, pakete []domain.Paket) error
	Pakete(ctx context.Context) ([]domain.Paket, error)
	PaketEntfernen(ctx context.Context, art, name string) error

	// **Und die Frage nach einem Muster** (#632): eine `package choice` baut
	// ihre Liste daraus.
	Musterablage
}

// Paketverwalter führt die installierten Vorlagenpakete.
//
// # Was er nicht tut
//
// **Er lädt nichts herunter.** Der Bezug aus dem Netz steht in
// [Paketbezieher] und beginnt mit einer ausdrücklichen Anweisung (#56); hier
// wird verwaltet, was auf dieser Platte liegt. Das ist die dritte Zusage aus
// #54, und sie
// ist strukturell gehalten: dieses Paket kennt `internal/bezug` nicht, und der
// Wächter `schichten.TestNurBezugKenntDasNetz` hält fest, dass niemand sonst
// im Projekt das Netz erreicht.
type Paketverwalter struct {
	katalog Paketablage

	// verwaltet ist das Verzeichnis, in das installiert und aus dem entfernt
	// wird. **Nur dieses eine**: die mitgelieferte Saat liegt im Bündel, und
	// ein Entfernen dort löschte einen Teil der Anwendung (#58).
	verwaltet string

	// wurzeln sind alle Orte, an denen Pakete liegen — bezogene, verwaltete
	// und die Saat. Der Katalog zeigt sie **alle**; er soll ja sagen, was da
	// ist.
	wurzeln []string

	// umbenennen verschiebt eine Datei oder ein Verzeichnis — `os.Rename`,
	// einsetzbar für die Fehlerwege von Ersetzen und Zurück (DOM-01.AK-02).
	umbenennen func(alt, neu string) error

	// setkoepfe liest die Sets dieses Profils — für die Nutzung in der
	// Übersicht (DOM-01.AK-01). Nil heißt: keine zu zählen.
	setkoepfe Setkopfleser

	// bau ist der Wegwerfort, an dem ein **neues** Paket entsteht, bevor es
	// installiert wird ([Paketverwalter.Anlegen]). Leer heißt: der des
	// Systems; gesetzt wird er nur von einem Test, der den Fehlerweg fährt.
	bau string

	// ablegen schreibt eine Datei ins entstehende Paket — `schreibeAtomar`.
	// **Ein Griff und kein fester Ruf**, wie `umbenennen`: seit jeder Bau
	// seinen eigenen Wegwerfort hat (`wegwerfpaket`), lässt sich der
	// Schreibfehler nicht mehr über einen vorbereiteten Pfad erzwingen.
	ablegen func(pfad string, inhalt []byte) error
}

// NeuerPaketverwalter baut den Verwalter über Katalog, Verwaltungsort und Wurzeln.
func NeuerPaketverwalter(katalog Paketablage, verwaltet string, wurzeln ...string) *Paketverwalter {
	return &Paketverwalter{
		katalog: katalog, verwaltet: verwaltet, wurzeln: wurzeln,
		umbenennen: os.Rename, ablegen: schreibeAtomar,
	}
}

// Auffrischen liest die Pakete von der Platte und gleicht den Katalog ab.
//
// **Befunde statt Abbruch** (§8): ein Verzeichnis auf `.mse-style` ohne lesbare
// Hauptdatei liegt in jeder gewachsenen Sammlung. Wer daran abbräche, hätte für
// **kein** Paket einen Katalog.
func (v *Paketverwalter) Auffrischen(ctx context.Context) ([]domain.Befund, error) {
	pakete, befunde := v.lesen(nil)

	if err := v.katalog.Abgleichen(ctx, pakete); err != nil {
		return befunde, err
	}

	return befunde, nil
}

// Liste gibt den Katalog aus dem Bestand.
//
// **Mit der lokalen Änderung** (DOM-01.AK-01): wer einen Abdruck trägt, wird
// mit dem verglichen, was heute dasteht.
func (v *Paketverwalter) Liste(ctx context.Context) ([]domain.Paket, error) {
	pakete, err := v.katalog.Pakete(ctx)
	if err != nil {
		return nil, err
	}

	return mitAenderung(pakete), nil
}

// Installieren legt ein Paket von der Platte in das Verwaltungsverzeichnis.
//
// **Ein vorhandenes Paket ist ein Konflikt und kein Überschreiben.** Wer eine
// Vorlage bearbeitet hat, verlöre sie sonst bei einer zweiten Installation, und
// nichts sagte es ihm.
func (v *Paketverwalter) Installieren(ctx context.Context, quelle string) (domain.Paket, error) {
	return v.einbinden(ctx, quelle, true)
}

// einbinden ist der gemeinsame Weg von Installieren und Anlegen —
// **`vermerken` sagt, ob die Quelle eine Herkunft ist**. Beim Installieren
// ist sie es; beim Anlegen ist sie ein Bauort, der gleich wieder
// verschwindet, und ein eigenes Paket hat keine Herkunft und keinen Abdruck
// (DOM-01.AK-01, #1696).
func (v *Paketverwalter) einbinden(ctx context.Context, quelle string, vermerken bool) (domain.Paket, error) {
	if err := v.ortPruefen(); err != nil {
		return domain.Paket{}, err
	}

	name := filepath.Base(filepath.Clean(quelle))

	art := msepaket.ArtAusName(name)
	if art == msepaket.ArtUnbekannt {
		return domain.Paket{}, domain.NewFieldError("quelle", domain.SchluesselValidierung,
			"%q trägt keine der sieben Paketendungen.", name)
	}

	// **Die Gestalt der Quelle zuerst**, und hier statt in `kopierePaket`: nur
	// hier ist der Fehlerweg erreichbar. Nach `paketLesen` steht fest, dass es
	// die Quelle gibt — ein `os.Stat` danach wäre ein Zweig, den kein Aufruf
	// betritt (§3).
	angaben, err := os.Stat(quelle)
	if err != nil {
		// **Der Systemfehler geht ins Protokoll, nicht in den Satz** (Befund 3
		// des Zweitblicks über Phase 4.13). „permission denied" oder „no such
		// file or directory" sagt dem, den es angeht, nichts — der Pfad, den
		// er selbst gewählt hat, sagt alles.
		return domain.Paket{}, errors.Join(
			domain.NewFieldError("quelle", domain.SchluesselValidierung,
				"%q ließ sich nicht lesen — prüfe, ob der Ordner noch dort "+
					"liegt und lesbar ist", quelle),
			domain.MitGrund(domain.ErrValidation, err, "die Quelle lesen"))
	}

	// **Nur Verzeichnisse**, und das ist eine Entscheidung mit Preis: ein Paket
	// als ZIP-Archiv wird abgewiesen. [msepaket.Oeffne] könnte es lesen, aber
	// [msepaket.Paketwurzeln] macht ein Archiv **nicht** zur Wurzel — ein so
	// installiertes Paket läge da und wäre für Stil, Bild und Schrift
	// unsichtbar. Lieber eine benannte Absage als ein Paket, das niemand findet.
	if !angaben.IsDir() {
		return domain.Paket{}, domain.NewFieldError("quelle", domain.SchluesselValidierung,
			"%q ist kein Verzeichnis. Ein Paket als Archiv wird nicht als Paketwurzel "+
				"gefunden und bliebe unsichtbar; wer es installieren will, packt es zuvor aus.",
			quelle)
	}

	ziel := filepath.Join(v.verwaltet, name)
	if _, err := os.Stat(ziel); err == nil {
		return domain.Paket{}, fmt.Errorf("%w: %q liegt bereits im Paketverzeichnis; "+
			"wer es ersetzen will, entfernt es zuerst", domain.ErrConflict, name)
	}

	// **Erst lesen, dann kopieren.** Ein Paket, das sich nicht öffnen lässt,
	// soll gar nicht erst im Verzeichnis landen — sonst stünde es danach da und
	// der Katalog nennte es als Befund.
	gelesen, fehler := v.paketLesen(quelle)
	if fehler != nil {
		return domain.Paket{}, fehler
	}

	if fehler := kopierePaket(quelle, ziel); fehler != nil {
		// **Und den halben Stand mitnehmen.** Ein abgebrochenes Kopieren
		// hinterließe ein halbes Paket unter dem Namen eines ganzen — derselbe
		// Grund wie beim Bezug.
		_ = os.RemoveAll(ziel) //nolint:errcheck // der eigentliche Befund ist der obige

		return domain.Paket{}, fehler
	}

	gelesen.Pfad = ziel

	// **Der Katalog wird ganz abgeglichen und nicht ergänzt.** Ein Eintrag
	// allein sagte nichts darüber, was sonst noch dasteht — und der Abgleich
	// ist der Weg, auf dem eine Herkunft überhaupt in den Bestand kommt.
	var vermerke []Vermerk
	if vermerken {
		gelesen.Herkunft = quelle
		vermerke = []Vermerk{{Ort: ziel, Herkunft: quelle}}
	}

	pakete, _ := v.lesen(vermerke)

	// **Ab hier hält kein Abbruch mehr auf** (§7, Regel 3: geschrieben ist
	// geschrieben). Das Paket **liegt** im Verzeichnis; ein `ctx`, der
	// dazwischen abbricht, meldete sonst einen Fehler über einen Zug, der
	// stattgefunden hat — und zurück ließe er ein installiertes Paket, das der
	// Katalog nicht kennt. Angestrichen vom Nachlauf über Phase 3.8
	// (#1169,
	// Befund 1); dieselbe Gestalt wie im Bestandsweg seit Phase 3.8b.
	if err := v.katalog.Abgleichen(context.WithoutCancel(ctx), pakete); err != nil {
		return domain.Paket{}, err
	}

	return *gelesen, nil
}

// Vermerken frischt den Katalog auf und schreibt Herkunft und Prüfsumme an
// alles, was unter einem Ort liegt.
//
// **Nach einem Bezug** (#56): `bezug` holt ein ganzes Repositorium, und die
// Summe gilt für alles darin. Sie „wird angezeigt und **vermerkt**" — angezeigt
// hat die Paketseite sie seit #139, und nach dem nächsten Start war sie fort.
func (v *Paketverwalter) Vermerken(ctx context.Context, vermerke ...Vermerk) ([]domain.Befund, error) {
	pakete, befunde := v.lesen(vermerke)

	if err := v.katalog.Abgleichen(ctx, pakete); err != nil {
		return befunde, err
	}

	return befunde, nil
}

// Entfernen nimmt ein Paket von der Platte und aus dem Katalog.
//
// # Warum die Bestätigung hier steht und nicht an der Wurzel
//
// `bestaetigterPfad` ist der Pfad, den der Rufende im Katalog **gesehen** hat.
// An der Wurzel verglichen, läse der Katalog **zweimal** — einmal zum
// Vergleichen, einmal zum Löschen.
//
// **Bestätigung und Wirkung bezögen sich damit nicht auf denselben Stand**
// (#1344).
// Wer dazwischen dasselbe Paket entfernt und ein anderes unter demselben Namen
// einbindet, ließe den alten Aufruf das neue löschen — und beide Vergleiche
// wären für sich richtig.
//
// Deshalb liest **eine** Lesung, und sie vergleicht und löscht.
//
// **Leer heißt: nicht bestätigt.** Das Fenster fragt in einem Dialog, die
// Kommandozeile zeigt ihren Aufruf; beide brauchen sie nicht
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
//
// # Was der Pfad hält und was nicht
//
// Er hält, **welcher Ort** geräumt wird. Er hält nicht, **welches Paket** dort
// liegt: wer zwischen Lesung und Ruf am selben Ort ein anderes einbindet,
// bekommt dieses gelöscht. Das ist kein Rest desselben Fehlers, sondern die
// Bedeutung der Bestätigung — der Rufende hat einen Ort bestätigt, und der
// wird geräumt.
func (v *Paketverwalter) Entfernen(
	ctx context.Context, art, name, bestaetigterPfad string,
) error {
	pakete, err := v.katalog.Pakete(ctx)
	if err != nil {
		return err
	}

	for _, paket := range pakete {
		if paket.Art != art || paket.Name != name {
			continue
		}

		if bestaetigterPfad != "" && paket.Pfad != bestaetigterPfad {
			return fmt.Errorf(
				"%w: der Pfad bestätigt das Paket nicht — %q liegt unter %q",
				domain.ErrConflict, name, paket.Pfad)
		}

		if !imVerzeichnis(paket.Pfad, v.verwaltet) {
			return domain.NewFieldError("pfad", domain.SchluesselValidierung,
				"%q liegt nicht im verwalteten Paketverzeichnis und wird deshalb nicht entfernt.",
				paket.Pfad)
		}

		// **Die Platte zuerst.** Bliebe der Eintrag und die Datei ginge, zeigte
		// der Katalog ein Paket, das es nicht gibt; andersherum liegt eine Datei
		// herum, die der nächste Abgleich ohnehin wieder aufnimmt — und das ist
		// die harmlosere Hälfte.
		if fehler := os.RemoveAll(paket.Pfad); fehler != nil {
			return fmt.Errorf("das Paket %q entfernen: %w", name, fehler)
		}

		// **Ab hier hält kein Abbruch mehr auf** — siehe `Installieren`. Die
		// Datei ist weg; ein Abbruch dazwischen ließe sie im Katalog stehen.
		return v.katalog.PaketEntfernen(context.WithoutCancel(ctx), art, name)
	}

	return fmt.Errorf("%w: das Paket %q der Art %q ist nicht installiert",
		domain.ErrNotFound, name, art)
}

// ortPruefen weist einen Verwaltungsort ab, der kein fester Ort ist.
//
// # Warum hier und nicht nur im Composition Root
//
// **Weil hier geschrieben wird.** `App.Starten` hebt den Fehler von
// `Paketverzeichnis()` auf und baut den Verwalter trotzdem — mit leerem Ort.
// `filepath.Join("", name)` wird daraus ein **relativer** Pfad, und installiert
// würde ins Arbeitsverzeichnis des Prozesses.
//
// Dieselbe Falle hat der Zweitblick über Phase 9 beim Paketbezieher gefunden
// (`wurzelPruefen`), und der über Phase 7 hier: **die Korrektur dort hat den
// Fall behandelt, nicht die Klasse.** Ein Weg, der schreibt, prüft seinen Ort
// selbst.
func (v *Paketverwalter) ortPruefen() error {
	if !filepath.IsAbs(v.verwaltet) {
		return domain.NewFieldError("ort", domain.SchluesselValidierung,
			"%q ist kein fester Ort für Pakete. Ohne einen wüsste niemand, wohin installiert wird.",
			v.verwaltet)
	}

	return nil
}

// imVerzeichnis sagt, ob ein Pfad unterhalb eines Verzeichnisses liegt.
//
// **Über `filepath.Rel` und nicht über ein Präfix.** `/pakete-alt` trägt
// `/pakete` als Zeichenkettenpräfix und liegt trotzdem nicht darin.
func imVerzeichnis(pfad, verzeichnis string) bool {
	rest, err := filepath.Rel(verzeichnis, pfad)
	if err != nil {
		return false
	}

	return rest != ".." && !strings.HasPrefix(rest, ".."+string(filepath.Separator))
}
