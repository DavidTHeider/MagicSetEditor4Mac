package service

// **Die Vorlagenpakete einer Sicherung** — OPS-06.AK-01
// (#1723).
//
// # Welche mitreisen und welche nicht
//
// **Ein Paket mit Prüfsumme ist bezogen** — es liegt an einer Adresse, von
// der es sich mit derselben Prüfsumme wieder holen lässt, und die bezogenen
// wiegen zusammen über ein Gigabyte. Es steht deshalb nur als **Verweis** im
// Manifest: Name, Herkunft, Prüfsumme.
//
// **Eines ohne ist nirgends sonst** — selbst angelegt, von Hand
// hineingelegt, bearbeitet. Es reist **mit**, unter `pakete/`. Bis OPS-06
// sagte `ADMINISTRATION.md`, die Pakete seien „unverändert wiederbeziehbar";
// für ein eigenes stimmte das nie.
//
// **Und ein bezogenes, das hier jemand bearbeitet hat, auch** — FMT-02.AK-04
// (#1691).
// Wiederbeziehbar ist nur die bezogene Fassung; käme es als Verweis zurück,
// wäre die Bearbeitung fort. Ob es bearbeitet ist, sagt sein Abdruck
// (DOM-01) — gerechnet hier und nicht dem Katalog geglaubt, der ihn ohne
// Rechnung gar nicht kennt.
//
// **Und nur, was im verwalteten Ordner liegt.** Der Katalog kennt auch Pakete
// aus einer MSE2-Installation; die gehören nicht zum Bestand, und eine
// Sicherung, die sie mitnähme, trüge fremde Dateien auf einen anderen Rechner.
//
// # Zurückgespielt wird dazu, nicht darüber
//
// **Was am Ziel schon liegt, bleibt** — auch ein älteres. Ein Paket, das hier
// inzwischen jemand bearbeitet hat, soll ein Zurückspielen nicht still
// ersetzen; der Bestand wird getauscht, die Vorlagen werden ergänzt.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Paketordner ist der Ordner der eigenen Pakete **im Archiv**.
const Paketordner = "pakete"

// Paketquelle nennt die Pakete des Katalogs.
type Paketquelle interface {
	Pakete(ctx context.Context) ([]domain.Paket, error)
}

// MitPaketen nimmt die Pakete in Sicherung und Zurückspielen auf — `ordner`
// ist das verwaltete Paketverzeichnis. **Ohne diesen Ruf bleibt es beim
// Bestand und den Bildern**, wie bis OPS-06.
func (b *Bestandsweg) MitPaketen(quelle Paketquelle, ordner func() (string, error)) {
	b.pakete, b.paketordner = quelle, ordner
}

// paketteile sammelt die eigenen Pakete und die Verweise auf alle.
func (b *Bestandsweg) paketteile(ctx context.Context) ([]sicherung.Teil, []sicherung.Paketverweis, error) {
	if b.pakete == nil {
		return nil, nil, nil
	}

	ordner, fehler := b.paketordner()
	if fehler != nil {
		return nil, nil, fmt.Errorf("das Paketverzeichnis finden: %w", fehler)
	}

	pakete, fehler := b.pakete.Pakete(ctx)
	if fehler != nil {
		return nil, nil, fmt.Errorf("die Pakete lesen: %w", fehler)
	}

	pakete = mitAenderung(pakete)

	var (
		teile    []sicherung.Teil
		verweise []sicherung.Paketverweis
	)

	for _, paket := range pakete {
		name, verwaltet := strings.CutPrefix(filepath.Clean(paket.Pfad), filepath.Clean(ordner)+string(filepath.Separator))
		if !verwaltet || strings.ContainsRune(name, filepath.Separator) {
			continue
		}

		verweis := sicherung.Paketverweis{
			Name: name, Herkunft: paket.Herkunft, Pruefsumme: paket.Pruefsumme,
			Enthalten: paket.Pruefsumme == "" || paket.Geaendert,
		}

		// **Ein Verweis wird aufgelöst** (C04, #2229): `WalkDir` folgt
		// keinem, und ein Paket, das als Verweis im Ordner lag, wurde zu
		// einem Teil, dessen Quelle ein Verzeichnis ist — die Sicherung brach
		// ab. Zeigt er nach außen, gilt er wie ein Paket von dort: er reist
		// nicht mit.
		wurzel, imOrdner := aufgeloestImOrdner(paket.Pfad, ordner)
		if !imOrdner {
			continue
		}

		if verweis.Enthalten {
			eigene, sammelfehler := teileUnter(wurzel, Paketordner+"/"+name)
			if sammelfehler != nil {
				return nil, nil, fmt.Errorf("das Paket %s sammeln: %w", name, sammelfehler)
			}

			teile = append(teile, eigene...)
		}

		verweise = append(verweise, verweis)
	}

	return teile, verweise, nil
}

// aufgeloestImOrdner gibt den echten Ort eines Pakets und ob er im
// verwalteten Ordner liegt.
//
// **Was sich nicht auflösen lässt, reist nicht mit** — ein Paket, das seit
// dem Lesen des Katalogs verschwand, ein Verweis ins Leere. Bis C04 stand ein
// verschwundenes als „enthalten" im Manifest, ohne dass ein Teil davon im
// Archiv lag. Genannt wird es im Protokoll.
func aufgeloestImOrdner(pfad, ordner string) (string, bool) {
	echt, fehler := filepath.EvalSymlinks(pfad)
	if fehler != nil {
		slog.Warn("ein Paket reist nicht mit: es lässt sich nicht auflösen", "datei", filepath.Base(pfad), "grund", fehler)

		return pfad, false
	}

	// **Liegt ein auflösbares Paket darin, ist auch der Ordner auflösbar** —
	// der Pfad stand schon textuell unter ihm.
	echterOrdner, _ := filepath.EvalSymlinks(ordner) //nolint:errcheck // siehe darüber

	return echt, strings.HasPrefix(echt, echterOrdner+string(filepath.Separator))
}

// paketlage teilt die Verweise einer Sicherung: welche mitkommen und welche
// hier fehlen — bezogen und nicht im Paketverzeichnis.
func (b *Bestandsweg) paketlage(verweise []sicherung.Paketverweis) (enthalten, fehlen []string) {
	enthalten, fehlen = []string{}, []string{}

	ordner := ""
	if b.paketordner != nil {
		ordner, _ = b.paketordner() //nolint:errcheck // ohne Ordner fehlt jedes bezogene
	}

	for _, verweis := range verweise {
		if verweis.Enthalten {
			enthalten = append(enthalten, verweis.Name)

			continue
		}

		if _, da := os.Stat(filepath.Join(ordner, verweis.Name)); ordner == "" || da != nil {
			fehlen = append(fehlen, verweis.Name)
		}
	}

	return enthalten, fehlen
}

// paketeEinsetzen legt die eigenen Pakete aus dem Zwischenplatz dazu.
//
// **Bestenfalls-Arbeit** (§7, Regel 3): der Bestand ist schon getauscht, und
// ein Paket, das sich nicht legen lässt, macht das nicht rückgängig. Es steht
// im Protokoll, und die Vorschau hatte es genannt.
func (b *Bestandsweg) paketeEinsetzen(ctx context.Context, platz string) {
	if platz == "" || b.paketordner == nil {
		return
	}

	ordner, fehler := b.paketordner()

	// **Der Zwischenplatz ist eben ausgepackt** — ein Lesefehler hieße, dass
	// es ihn nicht gibt, und dann gibt es auch nichts einzusetzen.
	eintraege, _ := os.ReadDir(platz) //nolint:errcheck // siehe oben

	for _, eintrag := range eintraege {
		ziel := filepath.Join(ordner, eintrag.Name())
		if _, da := os.Lstat(ziel); fehler != nil || da == nil {
			continue
		}

		fehler = errors.Join(os.MkdirAll(ordner, 0o750),
			os.Rename(filepath.Join(platz, eintrag.Name()), ziel))
	}

	if fehler != nil {
		slog.Default().ErrorContext(ctx, "die eigenen Pakete ließen sich nicht einsetzen",
			"grund", fehler.Error())
	}
}
