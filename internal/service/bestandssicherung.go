// **Die Sicherung trägt auch die Bilder** —
// #1162,
// Befund Z6-01 des Audits für 0.3.0 (Entwicklungsarchiv).
//
// # Der Befund
//
// Bis 0.2.0 schrieb „Bestand sichern" genau eine `.db`. Die eingesetzten
// Kartenbilder liegen daneben unter `beiwerk/`, und
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md) sagt
// **selbst**: ein Set ist ohne dieses Verzeichnis unvollständig.
//
// `docs/ADMINISTRATION.md` versprach „eine einzelne Datei, aus der heraus die
// Anwendung wieder aufsetzen kann".
//
// **Auf demselben Rechner fällt es nie auf** — das alte `beiwerk/` liegt ja
// noch da. Sichtbar wird es erst nach einem Rechnerwechsel, also genau dann,
// wenn die Sicherung das Einzige ist, was bleibt.
//
// # Warum das hier steht und nicht in `internal/sicherung`
//
// Weil hier die **Orte** bekannt sind: `internal/app` ist der einzige Platz,
// der Datenbankpfad und Bilderverzeichnis zugleich kennt (§2). Das Blatt
// daneben weiß, wie ein Archiv aussieht — nicht, was hineingehört.

package service

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Bestandsname ist der Name der Datenbank **im Archiv**.
const Bestandsname = "bestand.db"

// Beiwerkordner ist der Ordner der eingesetzten Bilder **im Archiv**.
const Beiwerkordner = "beiwerk"

// Sichern schreibt Bestand **und** Beiwerk als eine Sicherung — mit
// Kennwort, wenn der [Sicherungswunsch] im Kontext eines trägt — und
// vermerkt sie (OPS01, #2347).
func (b *Bestandsweg) Sichern(ctx context.Context, ziel string) error {
	return b.sichern(ctx, ziel, true)
}

// Probesichern sichert und prüft wie [Bestandsweg.Sichern] und **vermerkt
// nichts**: die Probe vor einer Aktualisierung ([SicherungsprobeUeber])
// löscht ihre Datei gleich danach, und ein Vermerk nennte eine Sicherung,
// die es nicht gibt (OPS01, #2347).
func (b *Bestandsweg) Probesichern(ctx context.Context, ziel string) error {
	return b.sichern(ctx, ziel, false)
}

func (b *Bestandsweg) sichern(ctx context.Context, ziel string, vermerkt bool) error {
	wunsch := sicherungswunschAus(ctx)
	if fehler := b.SicherungswunschPruefen(wunsch); fehler != nil {
		return fehler
	}

	// **Die Datenbank kommt in eine Nachbardatei**, nicht ins Archiv
	// unmittelbar: `VACUUM INTO` schreibt eine Datei, keinen Strom.
	//
	// **Mit Kennwort neben den Bestand und nicht neben das Ziel** (OPS-06):
	// dort läge sie im Klartext auf dem Medium, das außer Haus geht, und ein
	// USB-Stick behält eine gelöschte Datei lesbar.
	zwischen := filepath.Dir(ziel)
	if wunsch.Kennwort != "" {
		zwischen = filepath.Dir(b.orte.Datenbank)
	}

	nachbar, err := os.MkdirTemp(zwischen, Sicherungsvorsatz)
	if err != nil {
		return fmt.Errorf("den Zwischenplatz anlegen: %w", err)
	}

	defer func() { _ = os.RemoveAll(nachbar) }() //nolint:errcheck // nach dem Packen ohne Wirkung

	bestand := filepath.Join(nachbar, Bestandsname)
	if fehler := b.sicherer.Sichern(ctx, bestand); fehler != nil {
		return fehler
	}

	beiwerk, err := b.beiwerkteile()
	if err != nil {
		return err
	}

	pakete, verweise, err := b.paketteile(ctx)
	if err != nil {
		return err
	}

	teile := make([]sicherung.Teil, 0, 1+len(beiwerk)+len(pakete))
	teile = append(teile, sicherung.Teil{Name: Bestandsname, Quelle: bestand})
	teile = append(teile, pakete...)

	angaben := sicherung.Angaben{Fassung: b.fassung, Erstellt: b.uhr().UTC(), Pakete: verweise}

	// **Erst geprüft, dann vermerkt** — OPS01 (#2347), der Grund steht in
	// `sicherungsbuch.go`.
	abdruck, fehler := sicherung.PackenGeprueft(ziel, append(teile, beiwerk...), angaben, wunsch.Kennwort)
	if fehler != nil || !vermerkt {
		return fehler
	}

	// **Erst aufräumen, dann vermerken** — sonst meldet der Stand „gesichert“,
	// während neben der Datei noch der Zwischenplatz liegt (gefunden in der
	// Pipeline am 2026-10-16). Das `defer` oben bleibt für die Fehlerwege.
	_ = os.RemoveAll(nachbar) //nolint:errcheck // ein Rest fällt der Restepflege zu

	return b.vermerken(ctx, ziel, abdruck, wunsch)
}

// SicherungswunschPruefen urteilt vor dem ersten Schritt über den Wunsch —
// das Fenster fragt, bevor es einen Dialog öffnet.
//
// **Ein verschlüsselter Bestand gibt keinen Klartext her, ohne dass es jemand
// ausdrücklich will** (OPS-06.AK-02). Bis dahin stand in der Dokumentation,
// die Sicherung eines verschlüsselten Bestands sei Klartext — und nirgends
// sonst: die Oberfläche sagte es nicht.
//
// **Eine Fläche, die die Verschlüsselung nicht kennt, gilt als
// unverschlüsselt** — das sind die Läufe dieses Pakets; jede ausgelieferte
// Fläche setzt die Chiffre (`internal/app`).
func (b *Bestandsweg) SicherungswunschPruefen(wunsch Sicherungswunsch) error {
	if wunsch.Kennwort != "" {
		return domain.PruefeSicherungskennwort(wunsch.Kennwort)
	}

	if verschluesselt, _ := b.Verschluesselungsstand(); verschluesselt && !wunsch.Klartext { //nolint:errcheck // siehe oben
		return domain.NewFieldError("kennwort", domain.SchluesselSicherungImKlartext,
			"Der Bestand ist verschlüsselt — eine Sicherung ohne Kennwort läge im Klartext. "+
				"Setze ein Kennwort oder wähle ausdrücklich „im Klartext sichern“.")
	}

	return nil
}

// beiwerkteile sammelt die eingesetzten Bilder.
//
// **Ein fehlendes Verzeichnis ist kein Fehler**: es entsteht beim ersten Bild
// (`service.Bildablage.Ablegen`), und wer nie eines eingesetzt hat, hat keines.
func (b *Bestandsweg) beiwerkteile() ([]sicherung.Teil, error) {
	wurzel, err := b.orte.Bilder()
	if err != nil {
		return nil, fmt.Errorf("das Bilderverzeichnis finden: %w", err)
	}

	teile, fehler := teileUnter(wurzel, Beiwerkordner)
	if fehler != nil {
		return nil, fmt.Errorf("die eingesetzten Bilder sammeln: %w", fehler)
	}

	return teile, nil
}

// teileUnter sammelt jede Datei unter `wurzel` als Teil unter `vorsatz` —
// und `wurzel` selbst, wenn sie eine Datei ist (ein gepacktes Paket).
func teileUnter(wurzel, vorsatz string) ([]sicherung.Teil, error) {
	var teile []sicherung.Teil

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, unterwegs error) error {
		if unterwegs != nil {
			if os.IsNotExist(unterwegs) && pfad == wurzel {
				return fs.SkipAll
			}

			return unterwegs
		}

		if eintrag.IsDir() {
			return nil
		}

		// **Abgeschnitten und nicht ausgerechnet.** `filepath.Rel` gäbe hier
		// dasselbe und dazu einen Fehlerzweig, den `WalkDir` nie herstellen
		// kann — beide Pfade kommen von ihm und teilen die Wurzel. Ein Zweig,
		// den niemand betreten kann, ist eine Behauptung und keine Prüfung.
		name := vorsatz
		if darunter, ist := strings.CutPrefix(pfad, wurzel+string(filepath.Separator)); ist {
			name += "/" + filepath.ToSlash(darunter)
		}

		teile = append(teile, sicherung.Teil{Name: name, Quelle: pfad})

		return nil
	})

	return teile, fehler
}
