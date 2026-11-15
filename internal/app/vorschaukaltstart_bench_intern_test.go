//go:build fremdpakete

// **Wo die Zeit der *ersten* Vorschau hingeht** —
// #1270.
//
// # Warum eine zweite Messung
//
// `vorschauzerlegung_bench_intern_test.go` misst den **eingeschwungenen** Weg:
// Stilpaket gelesen, Schriften gefunden, Bilder da. Das ist, was jeder
// Tastendruck kostet, und seit dem Bildweg sind es im Fenster **72–75 ms**.
//
// **Die erste Vorschau eines Stils kostet ein Vielfaches.** Im Fenster
// gemessen am 2026-08-28, `Media and Collaboration Promos`: 4.715 ms für die
// erste Karte, und 2.506 ms noch einmal, als eine spätere Karte einen
// **anderen** Stil nannte. Zwischendrin 72–78 ms.
//
// Eine Zahl über den warmen Weg sagt darüber nichts (§10). Diese Messung
// trennt den Kaltstart in seine Teile, damit die nächste Optimierung an der
// Stelle ansetzt, die sie kostet — und nicht an der, die man vermutet.
package app

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// schwelleKaltstart gilt für die beiden **kalten** Hälften zusammen.
//
// **Gemessen am 2026-08-28** an `Vanguard.mse-set` mit `magic-m15-altered`:
// Kartenplan kalt 683 ms, erstes Zeichnen 678 ms — zusammen 1,36 s. Warm sind
// dieselben Schritte 29 ms und 13 ms.
//
// **Dass es zwei Hälften sind, ist die Auskunft dieser Messung.** Die erste
// war zu erwarten: ein `.mse-style` wird ausgepackt und seine Skripte werden
// zerteilt. Die zweite nicht — das erste **Zeichnen** kostet das Fünfzigfache
// jedes weiteren, weil dabei Schriften gesucht und Rahmenbilder entschlüsselt
// werden. Wer nur den Planbau beschleunigte, hätte die Hälfte gewonnen.
//
// **Neu gemessen am 2026-10-15 mit den Schriften des Packs** (#2416, fünf
// Läufe): 754–802 ms zusammen. Bis dahin fehlte der Messung `Beleren Bold`.
//
// **2,5 s lassen keine Verdopplung durch** (Faktor 1,8). Weiter als bei den
// warmen Wegen, weil ein einzelner Durchlauf keine Streuung mittelt.
const schwelleKaltstart = 2500 * time.Millisecond

// BenchmarkWegVorschaukaltstart zerlegt die erste Vorschau eines Stils.
//
// **Kein `b.Loop`**: ein Kaltstart geschieht genau einmal. Gemessen wird mit
// der Uhr, und die Zahlen stehen im Protokoll — eine Schwelle steht hier
// **nicht**, weil ein Durchlauf keine Streuung kennt und ein Wächter darauf
// nur flackerte.
func BenchmarkWegVorschaukaltstart(b *testing.B) {
	anwendung := messanwendung(b)
	setID := eingefahren(b, anwendung, "Vanguard.mse-set")

	zeilen := anwendung.Kartenzeilen(context.Background(), setID)
	if zeilen.Fehler != nil {
		b.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	kennung, befund := kennungAus(zeilen.Zeilen[0].ID)
	if befund != nil {
		b.Fatalf("Kennung: %+v", befund)
	}

	// **Im Profil der Anwendung, wie die Vorschau** — ohne las der Planbauer
	// den Setkopf nicht und maß still den Behelfsplan (QA-03, #1656).
	profil := anwendung.imProfil(context.Background())

	karte, lesefehler := anwendung.kartenleser.KarteLesen(profil, kennung)
	if lesefehler != nil {
		b.Fatalf("KarteLesen: %v", lesefehler)
	}

	// **Der Kartenplan kalt.** Das ist der Schritt, der das Stilpaket öffnet,
	// die Skripte zerteilt und die Schriften sucht — alles, was ein
	// Zwischenspeicher danach behält.
	kaltzeit := misszeit(b, func() {
		//nolint:errcheck // gemessen wird die Zeit, nicht der Fehler
		_, _, _ = anwendung.planbauer.Kartenplan(profil, karte)
	})

	warmzeit := misszeit(b, func() {
		//nolint:errcheck // dito
		_, _, _ = anwendung.planbauer.Kartenplan(profil, karte)
	})

	plan, befunde, planfehler := anwendung.planbauer.Kartenplan(profil, karte)
	if planfehler != nil || len(befunde) > 0 {
		b.Fatalf("Kartenplan: %v %v — gemessen würde der Behelfsplan", planfehler, befunde)
	}

	rasterzeit := misszeit(b, func() {
		//nolint:errcheck // dito
		_, _, _ = anwendung.rasterer.Zeichne(
			context.Background(), plan, 1, render.GueteMittel)
	})

	var erste VorschauAntwort

	erstezeit := misszeit(b, func() {
		erste = anwendung.Vorschau(context.Background(), zeilen.Zeilen[0].ID, 1)
	})

	// **Erst nach der Messung geprüft** (#2416) — vorher gefragt, wäre der
	// Kaltstart warm.
	if erste.Fehler != nil || erste.Vorschau == nil {
		b.Fatalf("Vorschau: %+v", erste.Fehler)
	}

	keineFehlendeSchrift(b, erste.Vorschau.Befunde)

	zweitezeit := misszeit(b, func() {
		_ = anwendung.Vorschau(context.Background(), zeilen.Zeilen[0].ID, 1)
	})

	// **Und der Kaltstart hält seine Schwelle.**
	//
	// Über `kaltzeit + rasterzeit`, denn das sind die beiden Hälften, die
	// jemand **wartet**: der Rest ist warm und hat seine eigene Schwelle in
	// `vorschauzerlegung_bench_intern_test.go`.
	haeltZeit(b, "der Kaltstart eines Stils", kaltzeit+rasterzeit, schwelleKaltstart)

	b.Logf("Kartenplan kalt   %v", kaltzeit.Round(time.Millisecond))
	b.Logf("Kartenplan warm   %v", warmzeit.Round(time.Millisecond))
	b.Logf("Zeichne           %v", rasterzeit.Round(time.Millisecond))
	b.Logf("Vorschau erste    %v", erstezeit.Round(time.Millisecond))
	b.Logf("Vorschau zweite   %v", zweitezeit.Round(time.Millisecond))
}

// misszeit gibt, wie lange ein Aufruf gebraucht hat.
func misszeit(b *testing.B, ruf func()) time.Duration {
	b.Helper()

	angefangen := time.Now()
	ruf()

	return time.Since(angefangen)
}
