package app_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Handlungsliste des gewählten Objekts steht auch in der Menüleiste** —
// GESTE (#2358).
//
// Die Oberfläche meldet Titel, Objekt, Epoche und die Einträge, wie sie im
// Kontextmenü und im Überlauf stehen; Go baut daraus ein Menü und meldet einen
// Klick mit derselben Epoche zurück. **Was der Klick tut, entscheidet das
// Fenster** — und ob die Epoche noch gilt, auch (`objekthandlungsraum.tsx`).

func karteMeldung(epoche int) app.Objekthandlungen {
	return app.Objekthandlungen{
		Titel:   "Karte",
		Art:     "karte",
		Kennung: "7",
		Epoche:  epoche,
		Eintraege: []app.Objekthandlung{
			{Schluessel: "duplizieren", Name: "Duplizieren"},
			{Schluessel: "loeschen", Name: "Löschen …", Gesperrt: true},
		},
	}
}

// titel gibt die Beschriftungen der obersten Ebene.
func titel(leiste *app.Menue) []string {
	namen := make([]string, 0, len(leiste.Eintraege))
	for _, eintrag := range leiste.Eintraege {
		namen = append(namen, eintrag.Beschriftung)
	}

	return namen
}

func TestDasGemeldeteObjektStehtAlsMenueHinterSet(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)
	gemeldet := [][]any{}
	anwendung.MitMeldung(func(_ context.Context, ereignis string, werte ...any) {
		gemeldet = append(gemeldet, append([]any{ereignis}, werte...))
	})
	texte := app.MenuetexteFuer("de")

	if antwort := anwendung.ObjekthandlungenMelden(karteMeldung(3)); antwort.Fehler != nil {
		t.Fatalf("ObjekthandlungenMelden: %+v", antwort.Fehler)
	}

	namen := titel(fang.zuletzt)
	set := slices.Index(namen, texte.Set)
	if set < 0 || set+1 >= len(namen) || namen[set+1] != "Karte" || namen[set+2] != texte.Fenster {
		t.Fatalf("die Leiste steht %v — erwartet »Karte« zwischen %q und %q", namen, texte.Set, texte.Fenster)
	}

	karte := fang.zuletzt.Eintraege[set+1].Untermenue.Eintraege
	if len(karte) != 2 || karte[0].Beschriftung != "Duplizieren" || karte[0].Gegraut ||
		karte[1].Beschriftung != "Löschen …" || !karte[1].Gegraut {
		t.Fatalf("das Menü trägt %+v", karte)
	}
	// **Ein Eintrag ohne Kürzel trägt keines** — welche es gibt, sagt die
	// Oberfläche (AP-7, #2364).
	if karte[0].Kuerzel != nil {
		t.Errorf("ein Eintrag trägt ein Kürzel: %+v", karte[0].Kuerzel)
	}

	karte[0].Handlung()
	want := []any{app.EreignisObjekthandlung, app.Objektausloesung{Art: "karte", Kennung: "7", Epoche: 3, Schluessel: "duplizieren"}}
	if len(gemeldet) != 1 || len(gemeldet[0]) != 2 || gemeldet[0][0] != want[0] || gemeldet[0][1] != want[1] {
		t.Fatalf("gemeldet: %+v, erwartet %+v", gemeldet, want)
	}
}

func TestEinLeeresObjektNimmtDasMenueWeg(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)
	anwendung.ObjekthandlungenMelden(karteMeldung(1))
	anwendung.ObjekthandlungenMelden(app.Objekthandlungen{})

	if slices.Contains(titel(fang.zuletzt), "Karte") {
		t.Fatalf("das Menü steht noch: %v", titel(fang.zuletzt))
	}
}

func TestDieselbeMeldungBautKeineNeueLeiste(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)
	anwendung.ObjekthandlungenMelden(karteMeldung(1))
	anwendung.ObjekthandlungenMelden(karteMeldung(1))

	if fang.anzahl != 1 {
		t.Fatalf("gesetzt %d-mal, erwartet einmal", fang.anzahl)
	}

	// **Eine neue Epoche ist ein neuer Stand** — der Klick muss sie tragen.
	anwendung.ObjekthandlungenMelden(karteMeldung(2))
	if fang.anzahl != 2 {
		t.Fatalf("gesetzt %d-mal nach neuer Epoche, erwartet zweimal", fang.anzahl)
	}
}

func TestEineUnbrauchbareMeldungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	viele := karteMeldung(1)
	for stelle := range 41 {
		viele.Eintraege = append(viele.Eintraege, app.Objekthandlung{Schluessel: "x" + strings.Repeat("y", stelle), Name: "X"})
	}
	ohneSchluessel := karteMeldung(1)
	ohneSchluessel.Eintraege[0].Schluessel = ""
	ohneTitel := karteMeldung(1)
	ohneTitel.Titel = " "
	doppelt := karteMeldung(1)
	doppelt.Eintraege[1].Schluessel = "duplizieren"
	unbekannt := karteMeldung(1)
	unbekannt.Eintraege[0].Kuerzel = "befehl+q"
	zweimal := karteMeldung(1)
	zweimal.Eintraege[0].Kuerzel = "befehl+d"
	zweimal.Eintraege[1].Kuerzel = "befehl+d"

	for name, meldung := range map[string]app.Objekthandlungen{
		"mehr als vierzig Einträge":     viele,
		"ein Eintrag ohne Schluessel":   ohneSchluessel,
		"Einträge ohne Titel":           ohneTitel,
		"ein Schlüssel zweimal":         doppelt,
		"ein Kürzel, das es nicht gibt": unbekannt,
		"ein Kürzel zweimal":            zweimal,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := app.NeueApp("")
			fang := neuerFang(anwendung)
			antwort := anwendung.ObjekthandlungenMelden(meldung)

			if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeValidierung {
				t.Fatalf("Antwort %+v, erwartet eine Abweisung", antwort.Fehler)
			}
			if fang.anzahl != 0 {
				t.Errorf("die Leiste wurde trotzdem gesetzt")
			}
		})
	}
}

// **Gesperrt gibt es kein Objektmenü** — hinter dem Sperrbildschirm ist nichts
// gewählt, und eine späte Meldung bringt es nicht zurück. Entsperrt gilt
// wieder die letzte Meldung, wie beim Darstellungsstand.
func TestGesperrtGibtEsKeinObjektmenue(t *testing.T) {
	anwendung := faktoranwendung(t)
	ctx := t.Context()
	fang := neuerFang(anwendung)

	if antwort := anwendung.KennwortWechseln(ctx, "", probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}
	anwendung.ObjekthandlungenMelden(karteMeldung(1))
	if !slices.Contains(titel(fang.zuletzt), "Karte") {
		t.Fatal("Vorbedingung: das Menü steht")
	}

	if antwort := anwendung.Sperren(ctx); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}
	anwendung.ObjekthandlungenMelden(karteMeldung(2))
	if slices.Contains(titel(fang.zuletzt), "Karte") {
		t.Fatalf("gesperrt steht das Menü: %v", titel(fang.zuletzt))
	}

	if antwort := anwendung.Entsperren(ctx, probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}
	if !slices.Contains(titel(fang.zuletzt), "Karte") {
		t.Fatalf("entsperrt fehlt das Menü: %v", titel(fang.zuletzt))
	}
}

// **Die Bindung reicht durch** (§16) — mit Tor, wie DarstellungsstandMelden.
func TestDieBindungMeldetDieObjekthandlungen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)

	if antwort := app.NeueBindungen(anwendung).ObjekthandlungenMelden(t.Context(), karteMeldung(1)); antwort.Fehler != nil {
		t.Fatalf("ObjekthandlungenMelden: %+v", antwort.Fehler)
	}
	if fang.anzahl != 1 {
		t.Errorf("die Leiste wurde %d-mal gesetzt, erwartet einmal", fang.anzahl)
	}
}

// **Die Kürzel der Kartenhandlungen stehen an der Leiste** — AP-7
// (#2364),
// Entwurf `Kontext09`: ⌘↩ legt an, ⌘D verdoppelt, ⌘⌫ löscht (mit
// Rückfrage), ⇞/⇟ blättern — die beiden **ohne** ⌘. Ob ein Kürzel im
// Textfeld dem Text gehört, misst `ap7kuerzel` am Fenster.
func TestDieKuerzelDerHandlungenStehenAnDerLeiste(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fang := neuerFang(anwendung)
	texte := app.MenuetexteFuer("de")
	meldung := karteMeldung(2)
	meldung.Eintraege = []app.Objekthandlung{
		{Schluessel: "anlegen", Name: "Karte anlegen", Kuerzel: "befehl+return"},
		{Schluessel: "duplizieren", Name: "Duplizieren", Kuerzel: "befehl+d"},
		{Schluessel: "loeschen", Name: "Löschen …", Kuerzel: "befehl+backspace"},
		{Schluessel: "vorige", Name: "Vorige Karte", Kuerzel: "bild auf"},
		{Schluessel: "naechste", Name: "Nächste Karte", Kuerzel: "bild ab"},
	}

	if antwort := anwendung.ObjekthandlungenMelden(meldung); antwort.Fehler != nil {
		t.Fatalf("ObjekthandlungenMelden: %+v", antwort.Fehler)
	}

	namen := titel(fang.zuletzt)
	karte := fang.zuletzt.Eintraege[slices.Index(namen, texte.Set)+1].Untermenue.Eintraege
	erwartet := []app.Kuerzel{
		{Taste: "return"},
		{Taste: "d"},
		{Taste: "backspace"},
		{Taste: "page up", OhneBefehl: true},
		{Taste: "page down", OhneBefehl: true},
	}
	for stelle, kuerzel := range erwartet {
		if karte[stelle].Kuerzel == nil || *karte[stelle].Kuerzel != kuerzel {
			t.Errorf("%s: %+v, erwartet %+v", karte[stelle].Beschriftung, karte[stelle].Kuerzel, kuerzel)
		}
	}
}
