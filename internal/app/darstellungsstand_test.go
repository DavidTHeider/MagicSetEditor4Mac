package app_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der Stand des Fensters führt die Menüleiste** — UXR-F14
// (#2029).
//
// # Wohin der Stand gehört
//
// **Die Oberfläche hat ihn, Go nimmt ihn entgegen.** Ob die Seite einen
// Arbeitsplatz hat und welche Spalte steht, weiß nur das Fenster — es hängt am
// Platz und an der Seite (`bausteine/Arbeitsplatz.tsx`). Go führt keine
// zweite Wahrheit, sondern den **letzten Stand, der gemeldet wurde**, und baut
// daraus die Titel und die Aktivierung. Gestellt wird nichts: ein Eintrag
// meldet ein Ereignis (`fensternaht.go`), das Fenster schaltet, und der neue
// Stand kommt über denselben Weg zurück.
//
// Die Menüleiste ist nativ; was hier steht, sieht kein React. Dass macOS die
// Einträge wirklich graut und umbenennt, zeigt nur das Fenster — dafür gibt es
// das Drehbuch `scripts/huellenprobe/darstellung.mjs`.

// leistenfang gibt die Naht, die jede gesetzte Leiste festhält.
type leistenfang struct {
	zuletzt *app.Menue
	anzahl  int
}

func neuerFang(anwendung *app.App) *leistenfang {
	fang := &leistenfang{}

	anwendung.MitMenue(func(_ context.Context, leiste *app.Menue) {
		fang.zuletzt = leiste
		fang.anzahl++
	})

	return fang
}

func TestDerGemeldeteStandFuehrtDieMenueleiste(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)
	texte := app.MenuetexteFuer("de")

	if fang.anzahl != 0 {
		t.Fatalf("vor der ersten Meldung wurde %d-mal gesetzt", fang.anzahl)
	}

	if antwort := anwendung.DarstellungsstandMelden(teile(true)); antwort.Fehler != nil {
		t.Fatalf("DarstellungsstandMelden: %+v", antwort.Fehler)
	}

	if fang.anzahl != 1 {
		t.Fatalf("die Leiste wurde %d-mal gesetzt, erwartet einmal", fang.anzahl)
	}

	navigator := eintragMitKuerzel(fang.zuletzt, navigatorkuerzel())
	if navigator.Beschriftung != texte.NavigatorAusblenden || navigator.Gegraut {
		t.Errorf("Navigator: %q, gegraut=%v", navigator.Beschriftung, navigator.Gegraut)
	}

	// **Der Navigator klappt zu, der Inspector fällt weg.**
	zugeklappt := app.Darstellungsstand{
		Seitenleiste: app.Fensterteil{Vorhanden: true, Offen: true},
		Navigator:    app.Fensterteil{Vorhanden: true, Offen: false},
	}
	anwendung.DarstellungsstandMelden(zugeklappt)

	if fang.anzahl != 2 {
		t.Fatalf("die Leiste wurde %d-mal gesetzt, erwartet zweimal", fang.anzahl)
	}

	navigator = eintragMitKuerzel(fang.zuletzt, navigatorkuerzel())
	inspector := eintragMitKuerzel(fang.zuletzt, inspectorkuerzel())

	if navigator.Beschriftung != texte.NavigatorEinblenden || navigator.Gegraut {
		t.Errorf("Navigator: %q, gegraut=%v", navigator.Beschriftung, navigator.Gegraut)
	}

	if !inspector.Gegraut {
		t.Error("der Inspector ist nicht gegraut, obwohl die Seite keinen hat")
	}
}

// **Derselbe Stand setzt die Leiste nicht noch einmal.** Die Oberfläche meldet
// bei jedem Wechsel der Seite; wer jedes Mal die ganze Leiste tauschte,
// ließe macOS sie bei jedem Klick neu aufbauen.
func TestDerselbeStandSetztDieLeisteNichtNeu(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)

	anwendung.DarstellungsstandMelden(teile(true))
	anwendung.DarstellungsstandMelden(teile(true))
	anwendung.DarstellungsstandMelden(teile(true))

	if fang.anzahl != 1 {
		t.Errorf("die Leiste wurde %d-mal gesetzt, erwartet einmal", fang.anzahl)
	}

	// **Und ein anderer Stand tut es wieder** — sonst wäre die Prüfung darüber
	// auch bei einer Naht grün, die nie etwas setzt (§3).
	anwendung.DarstellungsstandMelden(teile(false))

	if fang.anzahl != 2 {
		t.Errorf("nach einem anderen Stand wurde %d-mal gesetzt, erwartet zweimal", fang.anzahl)
	}
}

// **Der Stand überlebt den Sprachwechsel.** Die Leiste entsteht dabei neu; ohne
// den gemerkten Stand stünde sie danach wieder gegraut da.
func TestDerStandUeberlebtDenSprachwechsel(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)

	anwendung.DarstellungsstandMelden(teile(true))

	if antwort := anwendung.SpracheSetzen("en"); antwort.Fehler != nil {
		t.Fatalf("SpracheSetzen: %+v", antwort.Fehler)
	}

	navigator := eintragMitKuerzel(fang.zuletzt, navigatorkuerzel())
	if navigator.Beschriftung != "Hide Navigator" || navigator.Gegraut {
		t.Errorf("Navigator: %q, gegraut=%v", navigator.Beschriftung, navigator.Gegraut)
	}

	// **Und umgekehrt folgt ein neuer Stand der gewählten Sprache.**
	anwendung.DarstellungsstandMelden(teile(false))

	navigator = eintragMitKuerzel(fang.zuletzt, navigatorkuerzel())
	if navigator.Beschriftung != "Show Navigator" {
		t.Errorf("Navigator: %q, erwartet »Show Navigator«", navigator.Beschriftung)
	}
}

// zuletztEintraege gibt, was die Ablage unter „Zuletzt benutzt" führt.
func zuletztEintraege(t *testing.T, leiste *app.Menue) []*app.Menueeintrag {
	t.Helper()

	texte := app.MenuetexteFuer("de")

	ablage := untermenueMitNamen(leiste, texte.Ablage)
	if ablage == nil {
		t.Fatal("die Leiste führt keine Ablage")
	}

	zuletzt := untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege}, texte.Zuletzt)
	if zuletzt == nil {
		return nil
	}

	return zuletzt.Untermenue.Eintraege
}

// **Ein Stand nimmt „Zuletzt benutzt" nicht mit.** Die Leiste wird beim Melden
// neu gebaut; die Liste der Sets stammt aus dem Bestand und ist vom Stand des
// Fensters unabhängig — sie darf nicht verschwinden, nur weil jemand eine Spalte
// zuklappt.
func TestEinStandBehaeltDieZuletztBenutzten(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSetbenutzung(&benutzungsprobe{})

	for _, name := range []string{"Erstes Set", "Zweites Set"} {
		if antwort := anwendung.SetAnlegen(t.Context(), name, "probe", "probe-schlicht"); antwort.Fehler != nil {
			t.Fatalf("anlegen: %+v", antwort.Fehler)
		}
	}

	fang := neuerFang(anwendung)

	if antwort := anwendung.SetBenutzt(t.Context(), "1"); antwort.Fehler != nil {
		t.Fatalf("SetBenutzt: %+v", antwort.Fehler)
	}

	davor := zuletztEintraege(t, fang.zuletzt)
	if len(davor) < 2 {
		t.Fatalf("Vorbedingung: nur %d Sets unter »Zuletzt benutzt«", len(davor))
	}

	anwendung.DarstellungsstandMelden(teile(true))

	if fang.anzahl != 2 {
		t.Fatalf("die Leiste wurde %d-mal gesetzt, erwartet zweimal", fang.anzahl)
	}

	danach := zuletztEintraege(t, fang.zuletzt)
	if len(danach) != len(davor) {
		t.Fatalf("danach %d Sets unter »Zuletzt benutzt«, davor %d", len(danach), len(davor))
	}

	for stelle := range davor {
		if davor[stelle].Beschriftung != danach[stelle].Beschriftung {
			t.Errorf("Stelle %d: %q, davor %q", stelle, danach[stelle].Beschriftung, davor[stelle].Beschriftung)
		}
	}
}

// **Gesperrt gibt es keinen Arbeitsplatz.** Hinter dem Sperrbildschirm steht
// weder Seitenleiste noch Navigator; die Leiste wird beim Sperren ohnehin neu
// gebaut (sie zeigt dann auch keine Sets mehr), und der gemerkte Stand darf sie
// nicht wieder öffnen. Entsperrt gilt wieder, was zuletzt gemeldet wurde — bis
// das Fenster neu meldet.
func TestGesperrtGibtEsKeinenArbeitsplatz(t *testing.T) {
	anwendung := faktoranwendung(t)
	ctx := t.Context()
	fang := neuerFang(anwendung)

	if antwort := anwendung.KennwortWechseln(ctx, "", probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	anwendung.DarstellungsstandMelden(teile(true))

	if navigator := eintragMitKuerzel(fang.zuletzt, navigatorkuerzel()); navigator.Gegraut {
		t.Fatal("Vorbedingung: der Navigator ist gegraut")
	}

	if antwort := anwendung.Sperren(ctx); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	for name, kuerzel := range map[string]*app.Kuerzel{
		"Seitenleiste": seitenleistenkuerzel(), "Navigator": navigatorkuerzel(), "Inspector": inspectorkuerzel(),
	} {
		if !eintragMitKuerzel(fang.zuletzt, kuerzel).Gegraut {
			t.Errorf("gesperrt ist %s nicht gegraut", name)
		}
	}

	// **Auch eine späte Meldung öffnet nichts**, solange gesperrt ist: sie
	// käme von einem Fenster, das es nicht mehr gibt.
	anwendung.DarstellungsstandMelden(teile(false))

	if !eintragMitKuerzel(fang.zuletzt, navigatorkuerzel()).Gegraut {
		t.Error("eine Meldung hat den Navigator gesperrt wieder geöffnet")
	}

	if antwort := anwendung.Entsperren(ctx, probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	texte := app.MenuetexteFuer("de")

	navigator := eintragMitKuerzel(fang.zuletzt, navigatorkuerzel())
	if navigator.Gegraut || navigator.Beschriftung != texte.NavigatorEinblenden {
		t.Errorf("entsperrt: Navigator %q, gegraut=%v — erwartet der letzte gemeldete Stand",
			navigator.Beschriftung, navigator.Gegraut)
	}
}

// **Die Bindung reicht durch** (§16) — und sie ist am Tor: gesperrt hält sie an
// wie jede andere, ohne dass sie in der Liste der Ausnahmen steht.
func TestDieBindungMeldetDenStand(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)

	if antwort := app.NeueBindungen(anwendung).DarstellungsstandMelden(t.Context(), teile(true)); antwort.Fehler != nil {
		t.Fatalf("DarstellungsstandMelden: %+v", antwort.Fehler)
	}

	if fang.anzahl != 1 {
		t.Errorf("die Leiste wurde %d-mal gesetzt, erwartet einmal", fang.anzahl)
	}
}

// **Ohne Bestand und ohne Sitzung geht es auch.** Der Fall ist erreichbar:
// scheitert der Start, gibt es weder das eine noch das andere, und die Leiste
// entsteht trotzdem (#962). Der Stand darf dort nichts voraussetzen.
func TestDerStandBrauchtWederBestandNochSitzung(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	fang := neuerFang(anwendung)

	if antwort := anwendung.DarstellungsstandMelden(teile(true)); antwort.Fehler != nil {
		t.Fatalf("DarstellungsstandMelden: %+v", antwort.Fehler)
	}

	if fang.anzahl != 1 || eintragMitKuerzel(fang.zuletzt, navigatorkuerzel()).Gegraut {
		t.Errorf("%d-mal gesetzt, Navigator gegraut=%v", fang.anzahl,
			eintragMitKuerzel(fang.zuletzt, navigatorkuerzel()).Gegraut)
	}
}
