// **Wie das Fenster aufgehen soll** — UX-14.AK-01
// (#1686).
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// errLesefehler steht für „die Lesung ging schief" — mehr braucht der Lauf
// unten nicht.
var errLesefehler = errors.New("planmäßig gescheitert")

// mitEinstellungen gibt eine Anwendung mit gesetzten Einstellungen.
func mitEinstellungen(werte map[string]string) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{werte: werte})

	return anwendung
}

// **Ohne offenen Bestand gilt die Vorgabe.** Genau dann gibt es keine
// Einstellung zu lesen — und genau dann öffnet das Fenster zur
// Wiederherstellung (#962). Ein Fenster, das dabei in einer merkwürdigen
// Größe aufgeht, wäre der zweite Schreck nach dem ersten.
func TestOhneBestandGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	stand := app.NeueApp("").Fensterstand(context.Background())

	if stand.Start != domain.FensterstartMaximiert {
		t.Errorf("der Start ist %q, erwartet %q", stand.Start, domain.FensterstartMaximiert)
	}
}

// **Was eingestellt ist, kommt heraus.**
func TestDerEingestellteStartKommtHeraus(t *testing.T) {
	t.Parallel()

	anwendung := mitEinstellungen(map[string]string{
		domain.EinstellungFensterstart: domain.FensterstartNormal,
	})

	if stand := anwendung.Fensterstand(context.Background()); stand.Start != domain.FensterstartNormal {
		t.Errorf("der Start ist %q, erwartet %q", stand.Start, domain.FensterstartNormal)
	}
}

// **„Wie zuletzt" bringt die zwei Maße mit** — sonst wäre der Wert eine
// Zusage ohne Rumpf.
func TestZuletztBringtDieMasseMit(t *testing.T) {
	t.Parallel()

	anwendung := mitEinstellungen(map[string]string{
		domain.EinstellungFensterstart: domain.FensterstartZuletzt,
	})
	ctx := context.Background()

	anwendung.FenstergroesseMerken(ctx, 1600, 900)

	stand := anwendung.Fensterstand(ctx)

	if stand.Breite != 1600 || stand.Hoehe != 900 {
		t.Errorf("die Maße sind %d × %d, erwartet 1600 × 900", stand.Breite, stand.Hoehe)
	}
}

// **„Normal" heißt die eingebaute Größe** — auch wenn eine gemerkt ist
// (`domain.FensterstartNormal`). Die Maße gehören zu „wie zuletzt"; bis zum
// 2026-09-26 kamen sie bei jedem Start mit, der nicht maximiert war, und
// „normal" war ein zweites „zuletzt" (Hüllenprobe).
func TestNormalNimmtDieEingebauteGroesse(t *testing.T) {
	t.Parallel()

	anwendung := mitEinstellungen(map[string]string{
		domain.EinstellungFensterstart: domain.FensterstartNormal,
	})
	ctx := context.Background()

	anwendung.FenstergroesseMerken(ctx, 1600, 900)

	if stand := anwendung.Fensterstand(ctx); stand.Breite != 0 || stand.Hoehe != 0 {
		t.Errorf("die Maße sind %d × %d, erwartet 0 × 0 — die eingebaute Größe", stand.Breite, stand.Hoehe)
	}
}

// **Ohne gemerkte Maße bleibt es bei null** — und `main.go` nimmt dann seine
// eingebaute Größe. Wer „wie zuletzt" wählt, bevor es ein Zuletzt gibt,
// bekommt kein Fenster von null Punkten.
func TestOhneGemerkteMasseBleibtEsBeiNull(t *testing.T) {
	t.Parallel()

	stand := mitEinstellungen(map[string]string{
		domain.EinstellungFensterstart: domain.FensterstartZuletzt,
	}).Fensterstand(context.Background())

	if stand.Breite != 0 || stand.Hoehe != 0 {
		t.Errorf("die Maße sind %d × %d, erwartet 0 × 0", stand.Breite, stand.Hoehe)
	}
}

// **Eine unsinnige Größe wird nicht gemerkt.** Ein Bildschirm, der zwischen
// zwei Sitzungen wechselt, kann eine Größe hinterlassen, die auf dem nächsten
// nicht mehr passt; ein Fenster von 20 Punkten wäre nicht mehr zu fassen.
func TestEineUnsinnigeGroesseWirdNichtGemerkt(t *testing.T) {
	t.Parallel()

	anwendung := mitEinstellungen(nil)
	ctx := context.Background()

	anwendung.FenstergroesseMerken(ctx, 20, 20)

	if stand := anwendung.Fensterstand(ctx); stand.Breite != 0 || stand.Hoehe != 0 {
		t.Errorf("die Maße sind %d × %d, erwartet 0 × 0 — 20 Punkte sind keine Fenstergröße",
			stand.Breite, stand.Hoehe)
	}
}

// **Beim Schließen wird gemerkt** — die Hülle reicht die Maße aus Wails'
// Fenster herein (`internal/fenster`, ENG-02); was hier geprüft wird, ist der
// Weg dahinter.
//
// **Gelesen mit „wie zuletzt"** — nur dort kommen die Maße heraus
// ([TestNormalNimmtDieEingebauteGroesse]).
func TestBeimSchliessenWirdGemerkt(t *testing.T) {
	anwendung := mitEinstellungen(map[string]string{
		domain.EinstellungFensterstart: domain.FensterstartZuletzt,
	})
	ctx := context.Background()

	anwendung.FenstergroesseMerken(ctx, 1440, 900)

	if stand := anwendung.Fensterstand(ctx); stand.Breite != 1440 || stand.Hoehe != 900 {
		t.Errorf("die Maße sind %d × %d, erwartet 1440 × 900", stand.Breite, stand.Hoehe)
	}
}

// **Ein Lesefehler gibt die Vorgabe.** Der Bestand kann offen sein und die
// Lesung trotzdem scheitern; dann ist ein maximiertes Fenster die bessere
// Antwort als keines.
func TestEinLesefehlerGibtDieVorgabe(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{fehler: errLesefehler})

	stand := anwendung.Fensterstand(context.Background())

	if stand.Start != domain.FensterstartMaximiert || !stand.Maximiert {
		t.Errorf("der Stand ist %+v, erwartet die Vorgabe", stand)
	}
}

// **Ohne Einstellungen geschieht beim Merken nichts.** Diesen Fall gibt es
// im Betrieb nicht — `main.go` baut die Anwendung vollständig —, aber der
// Rückfall steht da, und ein unbetretener Zweig ist entweder ein fehlender
// Test oder toter Code (§3).
func TestOhneEinstellungenWirdNichtsGemerkt(t *testing.T) {
	t.Parallel()

	app.NeueApp("").FenstergroesseMerken(context.Background(), 1600, 900)
}
