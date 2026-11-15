package app

// **Ein Set als Umschlag weitergeben und einen annehmen** — Werkstück N4 von
// Phase 5.4 (#1457).
//
// # Warum das in `app` steht und nicht in `service`
//
// **Weil es zwei vorhandene Wege verschnürt.** Ein Set schreibt
// [service.Setschreiber], ein Set liest [App.SetOeffnen] — beide gibt es, und
// beide sind geprüft. Was hier entsteht, ist die Verdrahtung dazwischen und
// die Entscheidung, welcher Umschlag erzeugt wird: der Composition Root (§2).
//
// # Die drei Schritte, und warum es drei sind
//
//  1. **Schnüren** gibt einen Umschlag heraus.
//  2. **Ansehen** liest ihn und sagt, was drin ist — **ohne** etwas in den
//     Bestand zu schreiben.
//  3. **Übernehmen** schreibt ihn hinein.
//
// **Zwei und drei sind getrennt, und das ist die Zusage aus #1457**: „Import
// zeigt Objekte/Abhängigkeiten/Konflikte und benötigt fachliche
// Übernahmefreigabe." Wer beides zusammenlegte, hätte ein Ansehen, das
// schreibt — und dann wäre die Freigabe eine Frage nach etwas, das schon
// geschehen ist.

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// UmschlagSchnueren legt ein Set als portablen Umschlag ab.
//
// **Über eine Zwischendatei**, weil [service.Setschreiber] an einen Pfad
// schreibt und nicht in einen Puffer. Sie liegt in einem eigenen Verzeichnis
// und geht mit ihm; ein Umweg über das Zielverzeichnis ließe beim Abbruch
// eine halbe Datei neben dem Umschlag liegen.
// **Und seit `T7` auch eine Karte, eine Vorlage, ein Deck und eine
// freigegebene Sammlungssicht**
// (#1472) —
// die fünf Arten, die `N4` kennt. **Eine leere Art heißt Set**: bis heute nahm
// jede Fläche eine Setkennung und sonst nichts, und wer weiter nur eine
// schickt, bekommt weiter ein Set (0033, Regel 6).
func (a *App) UmschlagSchnueren(
	ctx context.Context, art, kennung, ziel string,
) UmschlagAntwort {
	ctx = a.imProfil(ctx)

	manifest, fehler := a.umschlagSchnueren(ctx, art, kennung, ziel)
	if fehler != nil {
		return UmschlagAntwort{
			Objekte: []UmschlagobjektDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return UmschlagAntwort{Objekte: objekteAus(manifest)}
}

func (a *App) umschlagSchnueren(
	ctx context.Context, rohart, kennung, ziel string,
) (austausch.Manifest, error) {
	art, fehler := umschlagart(rohart)
	if fehler != nil {
		return austausch.Manifest{}, fehler
	}

	// **Der Ausweis steht vor dem Packen.** Eine halbe Datei am Ziel wäre
	// schlimmer als gar keine — und das gilt für jede Art.
	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return austausch.Manifest{}, fehler
	}

	if art != austausch.ArtSet {
		inhalte, packfehler := a.inhalteFuer(ctx, art, kennung)
		if packfehler != nil {
			return austausch.Manifest{}, packfehler
		}

		return austausch.Schreibe(ziel, inhalte, ich)
	}

	setkennung, fehler := alsKennung(kennung, "Setkennung")
	if fehler != nil {
		return austausch.Manifest{}, fehler
	}

	unterwegs, fehler := os.MkdirTemp("", "mse-umschlag-")
	if fehler != nil {
		return austausch.Manifest{}, domain.MitGrund(domain.ErrValidation,
			fehler, "ein Verzeichnis für den Umschlag anlegen")
	}

	defer func() { _ = os.RemoveAll(unterwegs) }() //nolint:errcheck // Wegwerfort

	zwischen := filepath.Join(unterwegs, Setordner)

	speicherung, fehler := a.speicherer.Schreibe(ctx, setkennung, zwischen)
	if speicherung == nil {
		return austausch.Manifest{}, fehler
	}

	inhalte, lesefehler := setEntfalten(speicherung.Pfad)
	if lesefehler != nil {
		return austausch.Manifest{}, lesefehler
	}

	return austausch.Schreibe(ziel, inhalte, ich)
}

// Setordner ist der Name, unter dem ein Set im Umschlag liegt.
const Setordner = "set.mse-set"

// setEntfalten legt jede Datei des Setordners als eigenes Objekt hinein.
//
// # Warum entfaltet und nicht als ein Stück
//
// **Weil ein `.mse-set` ein Verzeichnis ist** — das hat ein Lauf gezeigt, und
// mein erster Anlauf las es mit `os.ReadFile`.
//
// **Und weil ein Archiv im Archiv die Grenzen aushebelte.** Der Umschlag
// prüft Zahl und Größe seiner Objekte, **bevor** er auspackt; ein einzelnes
// Objekt, das selbst ein ZIP wäre, käme als ein paar Kilobyte durch die
// Prüfung und entfaltete sich danach. Genau das ist eine Archivbombe.
func setEntfalten(ordner string) ([]austausch.Inhalt, error) {
	// **Gelesen wird durch eine [os.Root] und nicht über den Pfad.**
	//
	// Zwischen dem `WalkDir`, das eine Datei als gewöhnlich meldet, und dem
	// Lesen liegt ein Augenblick; wer in ihm einen Symlink hinlegt, ließe uns
	// eine Datei von anderswo einpacken. Dass der Ordner uns gehört, macht
	// das unwahrscheinlich und nicht unmöglich — und `os.Root` schließt es
	// aus, statt es zu beteuern. Angestrichen von `gosec` (G122).
	wurzel, fehler := os.OpenRoot(ordner)
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"das geschriebene Set öffnen")
	}

	defer func() { _ = wurzel.Close() }() //nolint:errcheck // nur gelesen

	var inhalte []austausch.Inhalt

	// **Gelaufen wird über das Dateisystem der Wurzel** und nicht über
	// Pfade daneben: dann sind die Namen schon relativ, und es gibt kein
	// zweites Paar aus Zusammensetzen und wieder Auseinandernehmen, das
	// auseinanderlaufen könnte.
	dateien := wurzel.FS()

	fehler = fs.WalkDir(dateien, ".",
		func(rel string, eintrag fs.DirEntry, weg error) error {
			if weg != nil {
				return weg
			}

			// **Nur gewöhnliche Dateien.** Ein Symlink im Setordner zeigte
			// nach dem Auspacken irgendwohin — und #1457 verbietet sie
			// ausdrücklich.
			if !eintrag.Type().IsRegular() {
				return nil
			}

			bytes, lesefehler := fs.ReadFile(dateien, rel)
			if lesefehler != nil {
				return lesefehler
			}

			inhalte = append(inhalte, austausch.Inhalt{
				Art:         austausch.ArtSet,
				Name:        path.Join(Setordner, rel),
				Bezeichnung: rel,
				Bytes:       bytes,
			})

			return nil
		})
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"das geschriebene Set lesen")
	}

	return inhalte, nil
}

// UmschlagAnsehen liest einen Umschlag und sagt, was darin steht.
//
// **Ohne etwas in den Bestand zu schreiben.** Was hier zurückkommt, ist eine
// Auskunft und keine Übernahme.
func (a *App) UmschlagAnsehen(ctx context.Context, pfad string) UmschlagAntwort {
	ctx = a.imProfil(ctx)

	sendung, fehler := a.umschlagLesen(pfad)
	if fehler != nil {
		return UmschlagAntwort{
			Objekte: []UmschlagobjektDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	defer func() { _ = os.RemoveAll(sendung.ort) }() //nolint:errcheck // Wegwerfort

	return UmschlagAntwort{
		Objekte:       objekteAus(sendung.Manifest),
		Urheber:       string(sendung.Urheber),
		Fingerabdruck: fingerabdruckVon(sendung.Urheber),
	}
}

// UmschlagUebernehmen schreibt einen angesehenen Umschlag in den Bestand.
//
// **Und legt neue lokale Kennungen an** — #1457 verlangt es: „Lokale IDs neu
// zuordnen". Das geschieht von selbst, weil der Weg derselbe ist wie beim
// Öffnen einer Setdatei: ein Set, das hereinkommt, ist ein neues Set.
func (a *App) UmschlagUebernehmen(
	ctx context.Context, pfad string,
) OeffnenAntwort {
	ctx = a.imProfil(ctx)

	// **Eine leere Liste und kein `nil`** — dieselbe Zusage wie in
	// [App.SetOeffnen], und sie gilt auch auf dem Fehlerweg. Ein `null` im
	// JSON macht jede Schleife im Fenster rot; angestrichen von
	// `leerelistenlaufzeit_test.go`, der genau danach fragt.
	leer := OeffnenAntwort{Befunde: []string{}, Angelegt: []string{}}

	sendung, fehler := a.umschlagLesen(pfad)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	defer func() { _ = os.RemoveAll(sendung.ort) }() //nolint:errcheck // Wegwerfort

	// **Der Setordner ist wieder ein `.mse-set`** — entfaltet hineingelegt,
	// entfaltet ausgepackt, und was dasteht, ist dasselbe Verzeichnis.
	for _, objekt := range sendung.Manifest.Objekte {
		if objekt.Art != austausch.ArtSet {
			continue
		}

		// **Der innere Weg und nicht [App.SetOeffnen]** — sonst vergäbe
		// [App.imProfil] hier eine **zweite** Vorgangskennung, und die
		// Meldung nennte eine andere als das Protokoll.
		return a.setOeffnenIm(ctx,
			filepath.Join(sendung.Verzeichnis, Setordner))
	}

	leer.Fehler = Uebersetze(ctx, fmt.Errorf(
		"%w: dieser Umschlag enthält kein Set", domain.ErrValidation))

	return leer
}

// geleseneSendung ist eine Sendung samt ihrem Wegwerfort.
type geleseneSendung struct {
	austausch.Sendung

	ort string
}

// umschlagLesen packt einen Umschlag in einen Wegwerfort aus.
//
// **In einen eigenen und nicht in den Bestand** — das ist die Quarantäne aus
// #1457: was hier liegt, ist geprüft und trotzdem noch nicht angenommen.
func (a *App) umschlagLesen(pfad string) (geleseneSendung, error) {
	ort, fehler := os.MkdirTemp("", "mse-sendung-")
	if fehler != nil {
		return geleseneSendung{}, domain.MitGrund(domain.ErrValidation,
			fehler, "ein Verzeichnis für die Sendung anlegen")
	}

	sendung, lesefehler := austausch.Lies(pfad, ort)
	if lesefehler != nil {
		_ = os.RemoveAll(ort) //nolint:errcheck // Wegwerfort

		return geleseneSendung{}, lesefehler
	}

	return geleseneSendung{Sendung: sendung, ort: ort}, nil
}
