package app

// **Der Pfadvertrag eines offenen Servers** — OPS-02.AK-04
// (#1719).
//
// # Was hier geschieht und was in der Domäne
//
// **Die Regel steht in `domain.PruefeDienstpfad`**: das Zuhause ohne seine
// verborgenen Winkel, dazu weitere Ordner. Hier steht, **welche** weiteren —
// die vier eingestellten und der Zwischenablageordner dieses Prozesses, weil
// Werkzeuge über ihn Dateien reichen, und was `MSE_DIENSTORDNER` nennt — und
// die Frage an die Platte, die die Domäne nicht stellen kann: **wohin eine
// Verknüpfung zeigt.** Geprüft wird der aufgelöste Pfad.

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Dienstordnerumgebung nennt weitere Ordner, die ein offener Server anfassen
// darf — getrennt wie `PATH`. **Für Betrieb und Läufe**: wer `mse-rest` für
// eine Automatisierung auf `/srv/karten` ansetzt, sagt es hier, statt den
// Vertrag abzuschalten. Dieselbe Form wie [Bestandsumgebung].
const Dienstordnerumgebung = "MSE_DIENSTORDNER"

// Dienstpfadvertrag gibt den Vertrag für einen offenen Server über diesem
// Bestand — für `mse-rest` und den Werkzeugserver über HTTP.
func Dienstpfadvertrag(db *sql.DB) func(ctx context.Context, pfad string) error {
	verwalter := einstellungsverwalterUeber(db)
	feste := append([]string{os.TempDir()}, filepath.SplitList(os.Getenv(Dienstordnerumgebung))...)

	// **Und die eigenen Orte** — Befund 4 des Zweitblicks über Phase 6.4:
	// Pakete und Vorsicherungen liegen unter `~/Library/Application
	// Support`, und der Vertrag wies die Pfade ab, die der Dienst selbst
	// ausgibt. Was sich nicht bestimmen lässt, fehlt — enger, nie weiter.
	if bestand, fehler := BestandsortFuerFlaechen(); fehler == nil {
		feste = append(feste, filepath.Dir(bestand))
	}

	if pakete, fehler := Paketverzeichnis(); fehler == nil {
		feste = append(feste, pakete)
	}

	return dienstpfadvertragMit(verwalter.Alle, os.UserHomeDir, feste, domain.Laufwerke)
}

func dienstpfadvertragMit(
	einstellungen func(context.Context) (map[string]string, error),
	heim func() (string, error),
	feste []string,
	laufwerke string,
) func(ctx context.Context, pfad string) error {
	return func(ctx context.Context, pfad string) error {
		zuhause, fehler := heim()
		if fehler != nil {
			return domain.MitGrund(domain.ErrValidation, fehler, "das Zuhause bestimmen")
		}

		weitere := slices.Clone(feste)

		// **Ein Lesefehler nimmt nur die eingestellten Ordner weg** — das
		// Zuhause bleibt; enger wird es dadurch, nie weiter.
		if werte, lesefehler := einstellungen(ctx); lesefehler == nil {
			for _, schluessel := range domain.Ordnerschluessel() {
				// **Nur ein Ordner, den ein Server als seinen gelten lässt** —
				// Befund 2 des Zweitblicks über Phase 6.4: eingestellt am
				// Fenster oder früher, bevor es diese Regel gab. Gefragt wird
				// der aufgelöste, damit eine Verknüpfung nicht hineinführt.
				ordner := domain.OrdnerAus(werte[schluessel], zuhause)
				if domain.PruefeDienstordnerUnter(aufgeloest(ordner), aufgeloest(zuhause), aufgeloest(laufwerke)) == nil {
					weitere = append(weitere, ordner)
				}
			}
		}

		// **Gefragt wird, wo die Datei wirklich liegt** — der aufgelöste Pfad
		// gegen die aufgelösten Ordner. Der genannte allein sagte nichts: ein
		// `~/Karten`, das nach `~/.ssh` zeigt, sähe harmlos aus, und unter
		// macOS ist schon `/tmp` eine Verknüpfung auf `/private/tmp`.
		if !filepath.IsAbs(pfad) {
			return domain.PruefeDienstpfad(pfad, zuhause, weitere)
		}

		aufgeloeste := make([]string, 0, len(weitere))
		for _, ordner := range weitere {
			aufgeloeste = append(aufgeloeste, aufgeloest(ordner))
		}

		return domain.PruefeDienstpfad(aufgeloest(pfad), aufgeloest(zuhause), aufgeloeste)
	}
}

// aufgeloest folgt den Verknüpfungen des längsten Anfangs, den es gibt —
// was dahinter noch nicht existiert, hängt unverändert an.
//
// **Auch die Ordner selbst**: unter macOS ist `/var` eine Verknüpfung auf
// `/private/var`, und ein aufgelöster Pfad läge sonst neben seinem nicht
// aufgelösten Zwischenablageordner.
func aufgeloest(pfad string) string {
	if pfad == "" {
		return ""
	}

	rest := ""
	anfang := filepath.Clean(pfad)

	for {
		echt, fehler := filepath.EvalSymlinks(anfang)
		if fehler == nil {
			return filepath.Join(echt, rest)
		}

		oben := filepath.Dir(anfang)
		if !errors.Is(fehler, fs.ErrNotExist) || oben == anfang {
			return filepath.Clean(pfad)
		}

		rest = filepath.Join(filepath.Base(anfang), rest)
		anfang = oben
	}
}
