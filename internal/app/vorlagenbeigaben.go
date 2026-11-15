package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was eine Exportvorlage neben ihre Ausgabe legen darf** —
// #657.
//
// # Der Befund, der das ausgelöst hat
//
// `write_image_file` gibt es seit
// #216 samt
// seiner Naht — und **niemand hat sie gesetzt**. `MitBildschreiber` kam in
// keiner Zeile Produktivcode vor, nur in Tests: die Fähigkeit war gebaut und
// unerreichbar. Fünf der neunzehn Vorlagen von `M15-Magic-Pack` scheiterten in
// der laufenden Anwendung an „der Name `write_image_file` ist hier nicht
// bekannt", während der Test danebenlag und grün war (§10).
//
// # Wo die Ausbruchsgrenze liegt
//
// **Hier**, denn hier entsteht der Pfad. Die Vorlage nennt einen **Namen**;
// welchen Ort er bekommt, entscheidet der Aufrufer — das Ausgabeverzeichnis
// steht fest, bevor das Skript läuft.
//
//	if (!ei.allow_writes_outside) {
//	  … throw ScriptError(_("Not a relative filename: ") + rel_name);
//	}
//
// `get_export_full_path`, `src/script/functions/export.cpp`, Zeilen 350–367,
// Fassung f9d9356 vom 2021-07-06. MSE prüft dieselbe Sache an derselben
// Stelle; `allow_writes_outside` gibt es hier nicht, und es soll es nicht
// geben.

// beigaben legt die Dateien einer Vorlage neben ihre Ausgabe.
type beigaben struct {
	// ziel ist das Verzeichnis der Ausgabe — von außen und nie aus der Vorlage.
	ziel string

	// vorlagenpaket ist der Ort, aus dem `copy_file` liest.
	vorlagenpaket string

	setID       int64
	setschreibt Speicherer

	// ablage ist gesetzt, wenn dies ein **Probelauf** ist
	// (#2015):
	// dann gehen Namen und Inhalt durch dieselben Prüfungen wie im echten Lauf,
	// und der Schritt, der auf die Platte ginge, trägt sich hier ein.
	ablage *service.Probeablage
}

// AusVorlage gibt eine Datei des Vorlagenpakets.
//
// **Über `msepaket.Oeffne` und nicht über `os.ReadFile`**: eine Vorlage darf
// ein Verzeichnis **oder** ein Archiv sein, und der Paketleser kennt beide.
// Seine Pfadprüfung gilt dabei mit — ein `..` im Namen kommt gar nicht erst
// bis zur Platte.
func (b beigaben) AusVorlage(name string) ([]byte, error) {
	paket, err := msepaket.Oeffne(b.vorlagenpaket)
	if err != nil {
		return nil, err
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	return paket.Datei(name)
}

// Lege schreibt Bytes neben die Ausgabe und gibt den Dateinamen.
func (b beigaben) Lege(datei string, inhalt []byte) (string, error) {
	name, err := harmloserName(datei)
	if err != nil {
		return "", err
	}

	// **Der Probelauf endet hier**, vor dem Schreiben: Name und Länge sind das,
	// was er weiß.
	if b.ablage != nil {
		b.ablage.Lege(name, len(inhalt))

		return name, nil
	}

	if err := service.SchreibeAtomar(filepath.Join(b.ziel, name), inhalt); err != nil {
		return "", err
	}

	return name, nil
}

// LegeSet schreibt das Set neben die Ausgabe und gibt den Namen.
//
// # Ein Archiv und kein Verzeichnis
//
//	set->saveCopy(out_path);
//
// `src/script/functions/export.cpp`, Zeile 436, und `saveCopy` ist
//
//	void Package::saveCopy(const String& name) {
//	  saveToZipfile(name, true, true);
//	  clearKeepFlag();
//	}
//
// `src/util/io/package.cpp`, Zeilen 112–115. Beides Fassung f9d9356 vom
// 2021-07-06. **MSE2 schreibt ein ZIP**, und eine Vorlage, die das Set neben
// ihre Ausgabe legt, hängt seinen Namen in eine Zeile — dort steht dann eine
// Datei und kein Ordner.
//
// # Warum hier ein Verzeichnis stand, und warum der Grund nicht trug
//
// Der Kommentar an `service.ausfuhrlauf.setleger` berief sich auf
// [0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md):
// ein ZIP-Schreiber weiche die Byte-Zusage auf. **Der Vermerk deckt diesen Fall
// nicht.** Er heißt „ein Set aus einem Archiv wird nicht **zurück**geschrieben"
// und wägt den Rundlauf über die **Herkunft** ab; eine Kopie neben die Ausgabe
// schreibt nichts zurück, und die Zusage aus
// #48 bleibt
// unberührt.
//
// **⌘S bleibt, wie es ist.** Dort gilt 0018 unverändert.
//
// # Über ein Verzeichnis und nicht daran vorbei
//
// **Der Setschreiber bleibt der Setschreiber.** Er legt ein Verzeichnis an,
// samt allem, was `service.Beiwerk` mitbringt — Bilder, Symbole, die
// eingesetzten Dateien —, und erst danach wird gepackt. Wer stattdessen einen
// zweiten Weg neben ihm bauteter, hätte zwei Wege für dieselbe Frage, und
// einer davon liefe irgendwann anders (§2).
func (b beigaben) LegeSet(ctx context.Context, datei string) (string, error) {
	name, err := harmloserName(datei)
	if err != nil {
		return "", err
	}

	// **Wie groß das Archiv wird, weiß erst das Packen** — der Probelauf nennt
	// den Namen und lässt die Größe offen.
	if b.ablage != nil {
		b.ablage.LegeOhneGroesse(name)

		return name, nil
	}

	// **Neben dem Ziel und nicht in ihm**, damit im Ausgabeverzeichnis nur das
	// fertige Archiv liegt: eine Vorlage listet, was dort steht.
	zwischen, err := os.MkdirTemp(filepath.Dir(b.ziel), "mse-setkopie-")
	if err != nil {
		return "", fmt.Errorf("das Zwischenverzeichnis anlegen: %w", err)
	}

	defer func() { _ = os.RemoveAll(zwischen) }() //nolint:errcheck // Zwischenstand

	entfaltet := filepath.Join(zwischen, name)

	if _, schreibfehler := b.setschreibt.Schreibe(ctx, b.setID, entfaltet); schreibfehler != nil {
		return "", schreibfehler
	}

	dateien, err := ausVerzeichnis(entfaltet)
	if err != nil {
		return "", err
	}

	if packfehler := msepaket.ArchivAnlegen(
		filepath.Join(b.ziel, name), dateien); packfehler != nil {
		return "", packfehler
	}

	return name, nil
}

// ausVerzeichnis liest einen Baum als Abbildung aus Paketnamen auf Inhalt.
//
// **Mit Schrägstrichen**, auch unter Windows: so heißt eine Datei **im** Paket
// (`docs/PORTIERUNG.md`).
//
// # Gelesen wird durch eine Wurzel und nicht über den Pfad
//
// **`os.Root` bindet jeden Zugriff an das Verzeichnis** — ein Symlink darin,
// der nach draußen zeigt, führt nicht hinaus.
//
// Der erste Anlauf las mit `os.ReadFile(pfad)` im Rückruf von `WalkDir` und
// setzte ein `//nolint:gosec` daneben. **Das war eine Ausrede und keine
// Antwort**: gosec meldete G122 (CWE-367), und der Befund stimmt. Dass wir
// dieses Verzeichnis eben selbst geschrieben haben, macht den Zugriff nicht
// wurzelgebunden — es macht nur den Angriff unwahrscheinlich, und das ist
// etwas anderes.
func ausVerzeichnis(wurzel string) (map[string][]byte, error) {
	griff, err := os.OpenRoot(wurzel)
	if err != nil {
		return nil, fmt.Errorf("das geschriebene Set öffnen: %w", err)
	}

	defer func() { _ = griff.Close() }() //nolint:errcheck // nur gelesen

	dateien := map[string][]byte{}

	fehler := fs.WalkDir(griff.FS(), ".", func(
		name string, eintrag fs.DirEntry, weiter error,
	) error {
		if weiter != nil || eintrag.IsDir() {
			return weiter
		}

		inhalt, lesefehler := fs.ReadFile(griff.FS(), name)
		if lesefehler != nil {
			return lesefehler
		}

		// **`fs.WalkDir` gibt Namen schon mit Schrägstrichen** — die
		// Umrechnung von vorher entfällt damit.
		dateien[name] = inhalt

		return nil
	})
	if fehler != nil {
		return nil, fmt.Errorf("das geschriebene Set einlesen: %w", fehler)
	}

	return dateien, nil
}

// harmloserName nimmt einen Namen an und weist einen Pfad ab.
//
// **Abgewiesen, nicht bereinigt.** Dieselbe Regel wie bei
// [service.unterverzeichnis] und aus demselben Grund: wer aus `../../x` ein
// `....x` machte, schriebe eine Datei, die niemand gemeint hat — und eine
// Regel, die aus einer Eingabe etwas anderes macht, muss man jedes Mal
// nachrechnen.
//
// **Ein Unterverzeichnis gibt es nicht.** MSE lässt `directory:` nur bei
// `write_image_file` zu; die drei hier nennen eine Datei.
func harmloserName(datei string) (string, error) {
	name := strings.TrimSpace(datei)

	if name == "" || name == "." || name == ".." ||
		strings.ContainsAny(name, "/\\\x00") || name != filepath.Base(name) {
		return "", fmt.Errorf("%w: %q ist kein Dateiname", domain.ErrValidation, datei)
	}

	return name, nil
}

// bildnaht macht aus dem Bildausführer der Anwendung einen [service.Bildschreiber].
//
// **Der Unterschied ist ein Argument**: die Fassade kennt ein `verhalten`
// („überschreiben" oder „nummerieren"), das Skript nicht. Eine Vorlage, die
// zweimal denselben Namen schreibt, meint dieselbe Datei — MSE überschreibt
// wortlos (`wxFileOutputStream out(out_path)`), und alles andere ergäbe
// Dateinamen, die in der erzeugten Ausgabe nicht stehen.
type bildnaht struct {
	aus Bildausfuehrer

	// ablage ist gesetzt, wenn dies ein **Probelauf** ist (#2015).
	ablage *service.Probeablage
}

func (b bildnaht) Schreibe(
	ctx context.Context, karteID int64, verzeichnis, datei string,
	format service.Bildformat, zielbreite int,
) (*service.Bildschreibung, error) {
	// **Im Probelauf wird geprüft und nicht gezeichnet.** Das Ziel geht durch
	// dieselbe Prüfung wie im Lauf ([service.PruefeBildziel]); gerastert wird
	// nichts, und die Maße des Bildes bleiben offen — eine Vorlage liest sie
	// nicht, sie hängt nur den Namen in ihre Ausgabe.
	if b.ablage != nil {
		if err := service.PruefeBildziel(datei, format, zielbreite); err != nil {
			return nil, err
		}

		pfad := filepath.Join(verzeichnis, datei)
		b.ablage.LegeOhneGroesse(pfad)

		return &service.Bildschreibung{Pfad: pfad}, nil
	}

	// **Parität.** Eine Beigabe zur Vorlage ist kein Bild, das jemand für
	// einen Druck wählt — sie gehört zur Ausgabe und soll aussehen wie MSEs.
	return b.aus.Schreibe(ctx, karteID, verzeichnis, datei, format, zielbreite,
		"ueberschreiben", render.GueteMittel)
}
