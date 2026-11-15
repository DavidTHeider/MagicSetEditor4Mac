package app

// **Die Sicherung vor der Migration** —
// #939.
//
// # Warum sie sein muss
//
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// sagt zu: **vorwärts immer, rückwärts nie.** Der zweite Satz ist nur
// vertretbar, wenn es einen Weg zurück gibt — und der führt über die
// Sicherung. Ohne sie kommt niemand, der auf eine ältere Fassung zurückgeht,
// je wieder an seinen Bestand.
//
// **Die Transaktion aus #938
// hilft dabei nicht.** Sie verhindert den halben Stand; der Bestand, um den es
// hier geht, ist **vollständig** migriert und wird von der alten Fassung
// abgewiesen.
//
// # Wann sie läuft
//
// Nur, wenn eine Migration aussteht — `storage.Migriere` ruft die Vorbereitung
// dann und sonst nie. Ein Start, der nur öffnet, schreibt keine Kopie.

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sicherungsname ist der Dateiname der Sicherung vor der Migration.
//
// **Er nennt die Fassung, auf die migriert wird**, nicht die, von der es
// kommt: die alte steht nirgends im Bestand, und wer zurückwill, sucht „der
// Stand, bevor 0.2.0 daran war".
func Sicherungsname(fassung string) string {
	return "bestand-vor-" + fassung + ".db"
}

// Sicherungspruefung sagt, ob eine vorhandene Datei als Sicherung taugt.
//
// **Eine Naht und kein Aufruf in `internal/storage`**: `internal/app` darf es
// außerhalb der Composition Roots nicht kennen (§2). Erfüllt wird sie von
// `storage.PruefeSicherung`, verdrahtet in `app.go` und `mcpnaht.go`.
type Sicherungspruefung func(ctx context.Context, pfad string) error

// sicherungVorMigration gibt die Vorbereitung, die `storage.Migriere` ruft.
//
// # Warum eine vorhandene Sicherung kein Fehler ist — **wenn sie eine ist**
//
// `storage.Sichern` weist ein vorhandenes Ziel ab — zu Recht: eine Sicherung
// überschreibt nichts. Hier heißt „liegt schon da" aber etwas anderes als
// sonst: ein früherer Lauf hat sie angelegt, die Migration danach ist
// gescheitert, und weil sie **eine Transaktion** ist, steht der Bestand
// unverändert da.
//
// Sie zu überschreiben wäre der Fehler: die zweite Kopie käme vom selben
// Stand, und wenn nicht, hätte man die gute gerade verloren.
//
// **Bis zum Zweitblick über Phase S galt jeder Namenskonflikt als Erfolg.**
// Der Kommentar hier behauptete, die vorhandene sei „genau die richtige" — und
// maß es nicht. Eine leere, beschädigte oder fremde Datei unter diesem Namen
// ließ die Migration ohne Sicherung laufen, und das ist genau der Fall, für
// den es sie gibt. Jetzt wird sie gelesen.
//
// # Und ob es der Stand von **jetzt** ist
//
// **Der Fall, in dem eine vorhandene Sicherung nicht der Stand von jetzt ist**
// (#967):
//
//  1. Upgrade, die Migration bricht ab. Die Sicherung liegt da und trägt den
//     Stand von vorher.
//  2. Wer weiterarbeiten will, startet die **alte** Fassung — sie öffnet den
//     Bestand, denn die Migration ist zurückgerollt — und ändert Karten.
//  3. Zweiter Upgradeversuch.
//
// Die vorhandene Sicherung enthält die Arbeit aus Schritt 2 nicht. Scheitert
// die Migration erneut und jemand spielt zurück, ist sie fort — und niemand
// hat etwas falsch gemacht.
//
// **Verglichen werden deshalb die Marken** ([Bestandsmarke]). Sind sie gleich,
// ist die vorhandene Sicherung genau die richtige und wird genommen; sind sie
// verschieden, entsteht eine **zweite** unter einem Namen, der die Marke
// trägt. Überschrieben wird nie: eine Sicherung überschreibt nichts.
//
// **Der Preis ist eine Kette von Ständen, die niemand aufräumt.** Er ist
// bewusst gewählt — die Alternative heißt, Arbeit zu verlieren, und das ist
// kein Preis, sondern ein Schaden.
func sicherungVorMigration(
	sichern Sicherer, pruefen Sicherungspruefung,
	eigene, fremde Bestandsmarke, ort, fassung string,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		ziel := filepath.Join(ort, Sicherungsname(fassung))

		fehler := sichern(ctx, ziel)
		if fehler == nil {
			return nil
		}

		if errors.Is(fehler, domain.ErrConflict) {
			return vorhandeneSicherung(ctx, sichern, pruefen, eigene, fremde, ziel)
		}

		return fmt.Errorf("den Bestand vor der Migration sichern: %w", fehler)
	}
}

// vorhandeneSicherung entscheidet über eine Datei, die schon dasteht.
func vorhandeneSicherung(
	ctx context.Context, sichern Sicherer, pruefen Sicherungspruefung,
	eigene, fremde Bestandsmarke, ziel string,
) error {
	if untauglich := pruefen(ctx, ziel); untauglich != nil {
		return fmt.Errorf(
			"%s liegt schon da, taugt aber nicht als Sicherung — lege sie beiseite "+
				"und starte erneut: %w", ziel, untauglich)
	}

	heute, heutefehler := eigene(ctx, "")
	if heutefehler != nil {
		return fmt.Errorf("den Stand des Bestands bestimmen: %w", heutefehler)
	}

	gesichert, gesichertfehler := fremde(ctx, ziel)
	if gesichertfehler != nil {
		return fmt.Errorf("den Stand von %s bestimmen: %w", ziel, gesichertfehler)
	}

	if gesichert == heute {
		return nil
	}

	// **Eine zweite Kopie unter einem Namen, der den Stand trägt.**
	//
	// Ohne die Marke im Namen hieße jeder zweite Versuch gleich, und der
	// nächste Konflikt wäre wieder eine Wiederverwendung — derselbe Fehler
	// eine Ebene tiefer.
	zweites := filepath.Join(filepath.Dir(ziel),
		strings.TrimSuffix(filepath.Base(ziel), ".db")+"-"+heute+".db")

	if fehler := sichern(ctx, zweites); fehler != nil {
		return fmt.Errorf(
			"%s stammt von einem anderen Stand, und die zweite Sicherung nach %s "+
				"scheiterte: %w", ziel, zweites, fehler)
	}

	return nil
}

// Bestandsmarke sagt, auf welchem Stand ein Bestand ist.
//
// **Eine Naht und kein Aufruf in `internal/storage`** (§2), wie
// [Sicherungspruefung]. Erfüllt wird sie von `storage.Bestandsmarke`,
// verdrahtet in `mcpnaht.go`.
//
// **Der Pfad ist leer für den eigenen Bestand** — er ist schon offen, und ihn
// ein zweites Mal über seinen Pfad zu öffnen wäre eine zweite Verbindung auf
// dieselbe Datei.
type Bestandsmarke func(ctx context.Context, pfad string) (string, error)

// Sicherer schreibt den Bestand nach `ziel`.
//
// **Eine Naht und kein `*sql.DB`**: der Test dieser Datei prüft die
// Entscheidung — sichern, weitermachen, abbrechen —, nicht `VACUUM INTO`. Das
// prüft `internal/storage`.
//
// **Und der Rückgabetyp ist ein nackter Rückruf**, kein `storage.Vorbereitung`:
// diese Datei ist kein Composition Root und darf `storage` nicht kennen (§2).
// Gos strukturelle Typisierung setzt beide gleich, und `app.go` reicht ihn
// weiter.
type Sicherer func(ctx context.Context, ziel string) error
