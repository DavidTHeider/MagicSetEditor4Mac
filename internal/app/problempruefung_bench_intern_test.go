//go:build fremdpakete

package app

// **Was eine Problemprüfung über ein ganzes Set kostet** — PERF02
// (#2352).
//
// Der Weg ist der echte: [service.Problempruefung] mit [kartenpruefer] über
// Planbauer und Rasterer der Anwendung, mit [problemarbeiter] Arbeitern und
// den Schriften des Packs. **Die Schwelle ist die Grenze aus dem Issue** —
// eine Prüfung von 3.788 Karten in höchstens 60 Sekunden —, gemessen an
// `Black`, dem größten der kanonischen Sets, auf Mac17,3 (Apple M5, vier
// Leistungs- und sechs Effizienzkerne, 32 GiB) am 2026-10-15:
//
//	zwei Arbeiter, volle Größe                  131 s
//	fünf Arbeiter, volle Größe                   87 s
//	fünf Arbeiter, halbe Größe                   61 s
//	… und die Schriftsuche unter dem Leseschloss 58 s
//	… und 512 statt 64 Bilder im Speicher        51 s
//
// **Gefunden hat das ein Profil und nicht eine Vermutung**: die Arbeiter
// warteten ein Drittel ihrer Zeit auf das Schloss der Schriftsuche, und der
// Bildspeicher räumte bei 64 Einträgen, während das Set 255 brauchte.

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schwellePruefung ist die Grenze aus PERF02 für ein Set mit 3.788 Karten.
const schwellePruefung = 60 * time.Second

func BenchmarkWegProblempruefung(b *testing.B) {
	anwendung := messanwendung(b)
	setID := eingefahren(b, anwendung, "Black.mse-set")

	// **Im Profil** — ohne fiele der Planbauer still auf den Behelfsplan.
	ctx := anwendung.imProfil(context.Background())

	zeilen := anwendung.Kartenzeilen(ctx, setID)
	if zeilen.Fehler != nil {
		b.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	karten := make([]*domain.CardRow, 0, len(zeilen.Zeilen))

	for _, zeile := range zeilen.Zeilen {
		kennung, befund := kennungAus(zeile.ID)
		if befund != nil {
			b.Fatalf("Kennung %q: %+v", zeile.ID, befund)
		}

		karte, err := anwendung.kartenleser.KarteLesen(ctx, kennung)
		if err != nil {
			b.Fatal(err)
		}

		karten = append(karten, karte)
	}

	pruefung := service.NeueProblempruefung(
		func(context.Context, int64) ([]*domain.CardRow, error) { return karten, nil },
		kartenpruefer(anwendung.planbauer, anwendung.rasterer),
		func(context.Context, int64) (string, error) { return "black", nil },
		time.Now,
	).MitArbeitern(problemarbeiter())

	for b.Loop() {
		stand, err := pruefung.Pruefen(ctx, 1, nil)
		if err != nil || !stand.Vollstaendig || stand.Gesamt != kartenGross {
			b.Fatalf("Stand: %d von %d, vollständig %v, %v", stand.Geprueft, stand.Gesamt, stand.Vollstaendig, err)
		}

		b.Logf("kennzahl problempruefung.karten_mit_problem=%d", len(stand.Karten))

		// **Was eine Karte hat**, damit eine Zahl nahe der Gesamtzahl nicht
		// unbesehen bleibt — sie spräche für eine gemeinsame Ursache.
		if len(stand.Karten) > 0 {
			b.Logf("die erste Karte mit Problem: %v", stand.Karten[0].Befunde)
		}
	}

	// **Die Zwischenspeicher danach** — ein Speicher, der sich während der
	// Prüfung immer wieder leert, kostet jede Karte einen Weg auf die Platte.
	for _, quelle := range anwendung.speicherquellen {
		b.Logf("Zwischenspeicher %+v", quelle())
	}

	haelt(b, "ein Set mit 3.788 Karten prüfen", schwellePruefung)
}
