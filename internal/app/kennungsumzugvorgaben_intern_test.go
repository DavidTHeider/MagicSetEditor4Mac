package app

// **Die Vorgaben beim Umzug der Kennung** — #2424: eine kleine Fassung von
// cfprefsd für [aufzeichnung], und der Lauf, in dem ein zweiter Umzug nach
// einem gescheiterten ersten die Werte der neuen Fassung behält.

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// leerePlist ist eine Domäne ohne Schlüssel.
const leerePlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict/>
</plist>
`

// vorgabe führt `defaults export` und `defaults import` gegen [aufzeichnung.inhalte].
func (a *aufzeichnung) vorgabe(befehl []string) error {
	if len(befehl) != 4 || befehl[0] != "defaults" {
		return nil
	}

	switch befehl[1] {
	case "export":
		inhalt, da := a.inhalte[befehl[2]]
		if !da {
			inhalt = leerePlist
		}

		return os.WriteFile(befehl[3], []byte(inhalt), 0o600)
	case "import":
		roh, fehler := os.ReadFile(befehl[3])
		if fehler != nil {
			return fehler
		}

		if a.inhalte == nil {
			a.inhalte = map[string]string{}
		}

		if a.domaenen == nil {
			a.domaenen = map[string]bool{}
		}

		a.inhalte[befehl[2]] = string(roh)
		a.domaenen[befehl[2]] = true
		a.domaenen[befehl[2]+" "+umzugsmarke] = strings.Contains(string(roh), "<key>"+umzugsmarke+"</key>")
	}

	return nil
}

// **Ein zweiter Umzug behält, was die neue Fassung inzwischen schrieb** —
// #2424, Befund 3 des Nachlaufs über Phase 3. Scheiterte der erste vor dem
// Import, stand die neue Domäne ohne Marke da; AppKit schrieb hinein, und der
// nächste Start importierte die alte **ganz** — `defaults import` ersetzt.
func TestEinZweiterUmzugBehaeltWasDieNeueFassungSchrieb(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	buch.domaenen = map[string]bool{altkennungen[0]: true}
	buch.inhalte = map[string]string{altkennungen[0]: plistMit("alt", "alt")}
	buch.scheitert = "defaults import " + Kennung + " "

	if err := umzug.vorgabenZiehen(t.Context()); err == nil {
		t.Fatal("der erste Umzug scheiterte nicht")
	}

	// **Die neue Fassung schreibt** — wie AppKit die Lage eines Fensters.
	buch.inhalte[Kennung] = plistMit("", "neu")
	buch.domaenen[Kennung] = true
	buch.scheitert = ""

	if err := umzug.vorgabenZiehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	neu := buch.inhalte[Kennung]
	for _, erwartet := range []string{"<key>fenster</key>\n\t<string>neu</string>", "<key>a</key>", "<key>" + umzugsmarke + "</key>"} {
		if !strings.Contains(neu, erwartet) {
			t.Errorf("die neue Domäne trägt %q nicht:\n%s", erwartet, neu)
		}
	}

	if strings.Contains(neu, "<string>alt</string>\n\t<key>fenster") || strings.Count(neu, "<key>fenster</key>") != 1 {
		t.Errorf("der Wert der neuen Fassung wurde ersetzt:\n%s", neu)
	}
}

// plistMit baut eine Domäne mit `a` (wenn nicht leer) und `fenster`.
func plistMit(a, fenster string) string {
	eintraege := ""
	if a != "" {
		eintraege += "\t<key>a</key>\n\t<string>" + a + "</string>\n"
	}

	eintraege += "\t<key>fenster</key>\n\t<string>" + fenster + "</string>\n"

	return strings.Replace(leerePlist, "<dict/>", "<dict>\n"+eintraege+"</dict>", 1)
}

// **Ein Zwischenspeicher, der sich nicht anlegen oder beschreiben lässt,
// meldet sich** — und importiert wird nichts.
func TestEinUnbeschreibbarerZwischenspeicherWirdGemeldet(t *testing.T) {
	t.Parallel()

	for name, versperren := range map[string]func(heim string){
		"der Ordner": func(heim string) {
			legeAn(t, filepath.Join(heim, "Library", "Caches", Kennung), "eine Datei an der Stelle des Ordners")
		},
		"die Sperre": func(heim string) {
			legeAn(t, filepath.Join(heim, "Library", "Caches", Kennung, "vorgaben.sperre", "x"), "ein Ordner an ihrer Stelle")
		},
	} {
		umzug, buch := umzugIn(t)
		buch.domaenen = map[string]bool{altkennungen[0]: true}
		versperren(umzug.heim)

		if err := umzug.vorgabenZiehen(t.Context()); err == nil {
			t.Errorf("%s: der Fehler ging verloren", name)
		}

		for _, befehl := range buch.nur("defaults") {
			if befehl[1] == "import" {
				t.Errorf("%s: importiert trotz des Fehlers: %v", name, befehl)
			}
		}
	}
}

// **Lässt sich die vorhandene neue Domäne nicht lesen, wird nichts importiert**
// — ein Import ersetzte sonst, was sie trägt.
func TestEineUnlesbareNeueDomaeneWirdGemeldet(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	buch.domaenen = map[string]bool{altkennungen[0]: true, Kennung: true}
	buch.scheitert = "defaults export " + Kennung + " "

	// **Es gibt sie** — ihre Datei steht da; ein Export, der scheitert, ist
	// dann kein „leer" (Befund 6).
	legeAn(t, filepath.Join(umzug.heim, "Library", "Preferences", Kennung+".plist"), "binär")

	if err := umzug.vorgabenZiehen(t.Context()); err == nil || !strings.Contains(err.Error(), "neuen Vorgaben") {
		t.Errorf("Fehler %v", err)
	}
}

// **Eine alte Domäne, die sich nicht lesen lässt, zieht nicht halb um.**
func TestEineUnlesbareAlteDomaeneWirdGemeldet(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	buch.domaenen = map[string]bool{altkennungen[0]: true}
	buch.inhalte = map[string]string{altkennungen[0]: "<plist><dict><key>a"}

	if err := umzug.vorgabenZiehen(t.Context()); err == nil || len(buch.inhalte) != 1 {
		t.Errorf("Fehler %v, Domänen %v", err, buch.inhalte)
	}
}

func TestDieVorgabenZiehenUeberDefaultsUm(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	buch.domaenen = map[string]bool{altkennungen[1]: true}

	if err := umzug.ziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	// **Über cfprefsd und nach dem Namen der Domäne**, nicht über eine Datei,
	// und **mit einer Marke**: der Gastlauf fand eine neue Domäne, die es gab
	// und die leer war — `defaults delete` lässt eine leere zurück.
	//
	// **Die Marke kommt mit dem Import und nicht danach** (Zweitblick über
	// Phase 3, Befund 9): ein eigenes `defaults write` konnte scheitern, und
	// der nächste Start importierte die alten Werte noch einmal über die, die
	// die neue Fassung inzwischen schrieb. **Gibt es die neue Domäne nicht —
	// und `defaults` sagt es —, wird sie nicht ausgeführt** (Nachlauf über
	// Phase 4, Befund 5), und alles liegt im Arbeitsordner dieses Laufs
	// (Zweitblick, Befund 5).
	var vorgaben [][]string

	for _, befehl := range buch.befehle {
		if befehl[0] == "plutil" || (befehl[0] == "defaults" && befehl[1] != "read") {
			vorgaben = append(vorgaben, befehl)
		}
	}

	lauf := filepath.Dir(vorgaben[0][3])
	erwartet := [][]string{
		{"defaults", "export", altkennungen[1], filepath.Join(lauf, "vorgaben-alt.plist")},
		{"plutil", "-convert", "xml1", filepath.Join(lauf, "vorgaben-alt.plist")},
		{"defaults", "import", Kennung, filepath.Join(lauf, "vorgaben.plist")},
	}

	if filepath.Dir(lauf) != vorgabenablage(umzug.heim) {
		t.Errorf("der Arbeitsordner %s liegt nicht im Zwischenspeicher", lauf)
	}

	if !slices.EqualFunc(vorgaben, erwartet, slices.Equal) {
		t.Errorf("Befehle %v,\nerwartet %v", vorgaben, erwartet)
	}
}
