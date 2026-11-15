// Die Gegenproben zur Regeltabelle.
//
// **Eine eigene Datei**, weil `regeln_test.go` sonst über die 400 Zeilen aus §2
// liefe. Der Schnitt folgt einer Frage: dort steht die **Regel**, hier steht,
// dass sie **wirkt** — und dass sie nicht alles abweist.
//
// Getrennt am 2026-08-22.
package schichten

import (
	"strings"
	"testing"
)

// TestJedeZeileDerRegelWeistIhrenVerstossAb führt die Gegenprobe zu jeder Zeile.
//
// Das ist der Beleg, dass die Regel **wirkt** und nicht nur dasteht. Ohne ihn
// wäre `TestKeinPaketVerletztDieSchichtenregel` grün, solange niemand einen
// Verstoß baut — also auch dann, wenn `pruefe` gar nichts prüfte.
func TestJedeZeileDerRegelWeistIhrenVerstossAb(t *testing.T) {
	faelle := []struct {
		name     string
		paket    string
		datei    string
		importe  []string
		erwartet string
	}{
		{
			name:     "domain darf nichts aus dem Projekt",
			paket:    "domain",
			importe:  []string{modulpfad + "/internal/storage"},
			erwartet: "nicht erlaubt; zulässig sind: nichts aus dem Projekt",
		},
		{
			name:     "domain darf kein database/sql",
			paket:    "domain",
			importe:  []string{"database/sql"},
			erwartet: "in dieser Schicht verboten (AGENTS.md §2)",
		},
		{
			name:     "msescript darf auch domain nicht",
			paket:    "msescript",
			importe:  []string{modulpfad + "/internal/domain"},
			erwartet: "nicht erlaubt; zulässig sind: nichts aus dem Projekt",
		},
		{
			name:     "msepaket darf nur domain und dateisperre",
			paket:    "msepaket",
			importe:  []string{modulpfad + "/internal/service"},
			erwartet: "nicht erlaubt; zulässig sind: domain, dateisperre",
		},
		{
			name:     "render darf nur domain",
			paket:    "render",
			importe:  []string{modulpfad + "/internal/msescript"},
			erwartet: "nicht erlaubt; zulässig sind: domain",
		},
		{
			name:     "storage darf nur domain, dateisperre und schluesselbund",
			paket:    "storage",
			importe:  []string{modulpfad + "/internal/service"},
			erwartet: "nicht erlaubt; zulässig sind: domain, dateisperre, schluesselbund",
		},
		{
			// **Und `dateisperre` gar nichts** (#1163). Es ist das einzige
			// Paket ohne jede Kante — auch `domain` kennt es nicht.
			name:     "dateisperre darf nichts",
			paket:    "dateisperre",
			importe:  []string{modulpfad + "/internal/domain"},
			erwartet: "nicht erlaubt; zulässig sind: nichts aus dem Projekt",
		},
		{
			name:     "service darf app nicht",
			paket:    "service",
			importe:  []string{modulpfad + "/internal/app"},
			erwartet: "nicht erlaubt; zulässig sind: bezug, dateisperre, domain, msepaket, msescript, mtgjson, pdf, render, sammelliste, scryfall, sicherung, svg, totp, kennwort, vcard, peer",
		},
		{
			// **Die Ausnahme für Tests hält Produktivcode nicht** (#1175).
			//
			// Ohne diese Zeile ließe sich [nurInTests] auf jede Datei
			// ausdehnen, und der ganze Umzug des Rückspielwegs wäre wieder
			// rückgängig zu machen, ohne dass etwas rot würde.
			name:     "app darf sicherung nur aus Tests",
			paket:    "app",
			datei:    "sicherungsnaht.go",
			importe:  []string{modulpfad + "/internal/sicherung"},
			erwartet: "nicht erlaubt; zulässig sind: aktualisierung, bezug, domain, msepaket, msescript, mtgjson, netz, pdf, render, schriften, scryfall, service, storage, svg, vorschauweg, mtgapi, mtggraphql, mtgjsonpreise, scrydex, cardtrader, ollama, anthropic, openai, xai, applefm, schluesselbund, leitung, fernleitung, austausch, entdeckung, anmeldung, biometrie, peer, dateisperre, prozessspeicher",
		},
		{
			name:     "app darf storage nur in app.go erzeugen",
			paket:    "app",
			datei:    "menue.go",
			importe:  []string{modulpfad + "/internal/storage"},
			erwartet: "nur in app.go und start.go und mcpnaht.go und kartenstand.go und scryfallnaht.go und skriptnaht.go und wortlistennaht.go und schluesselwortnaht.go und werkzeugnaehte.go und packnaht.go und bestandsnaehte.go und einstellungsverwalter.go und paketnaht.go und stilnaht.go und setangabennaht.go und vorlagenzeichnernaht.go und sammlungsnaht.go und sammlungfuerflaechen.go und preisbezugnaht.go und marktbindung.go und anbieterbindung.go und kinaht.go und kianwendung.go und kivorschlag.go und kianbieter.go und kimodellprobe.go und kifeldherkunftnaht.go und kibildnaht.go und kibildeinsetzen.go und diagnosenaht.go und gesundheitsnaht.go und vertraeglichkeitsnaht.go und scryfallstart.go und setloeschnaht.go und scryfallverdrahtung.go und scryfallnaehte.go und ablagenahtflaechen.go und ersetzungnahtflaechen.go und listennahtflaechen.go und bestandssicherungnaht.go und markennaht.go und rueckkehrnaht.go und pflegenaht.go und pflegeflaechen.go und ruecksetzungsnaht.go und kostennaht.go und mtgjsonnaht.go und bilddateinaht.go und vorlagennaht.go und listeneinfuhrnaht.go und sammlungssichtennaht.go und massennaht.go und setwache.go und kartenskriptnaht.go und objektmarkennaht.go und auszaehlnaht.go und autoersetzungspflege.go und decknahtflaechen.go und deckstart.go und turnierflaeche.go und nutzernaht.go und profilnaht.go und startmigration.go und anmeldestart.go und tokenwache.go und leutestart.go und vorgabennaht.go und chiffrenaht.go und verdrahtung.go und nurlesen.go und retrievalnaht.go und verdrahtungzeichnen.go und recherchenaht.go und peernaht.go und kontaktnaht.go und freigabenaht.go und tauschnaht.go und handelspostnaht.go und handelsverbuchungsnaht.go und kiuebernahmenaht.go und kigruppe.go und umschlagflaeche.go und netzflaeche.go und partieflaeche.go und marktlinkenaht.go und deckeinfuhrnaht.go und kontoschaunaht.go und handelsleitungsnaht.go und symbolbibliotheksflaeche.go und zusatznaht.go erlaubt (Composition Root, AGENTS.md §2)",
		},
		{
			name:     "auch ein gleichnamiges app.go im Unterordner zählt nicht",
			paket:    "app",
			datei:    "unterpaket/app.go",
			importe:  []string{modulpfad + "/internal/storage"},
			erwartet: "nur in app.go und start.go und mcpnaht.go und kartenstand.go und scryfallnaht.go und skriptnaht.go und wortlistennaht.go und schluesselwortnaht.go und werkzeugnaehte.go und packnaht.go und bestandsnaehte.go und einstellungsverwalter.go und paketnaht.go und stilnaht.go und setangabennaht.go und vorlagenzeichnernaht.go und sammlungsnaht.go und sammlungfuerflaechen.go und preisbezugnaht.go und marktbindung.go und anbieterbindung.go und kinaht.go und kianwendung.go und kivorschlag.go und kianbieter.go und kimodellprobe.go und kifeldherkunftnaht.go und kibildnaht.go und kibildeinsetzen.go und diagnosenaht.go und gesundheitsnaht.go und vertraeglichkeitsnaht.go und scryfallstart.go und setloeschnaht.go und scryfallverdrahtung.go und scryfallnaehte.go und ablagenahtflaechen.go und ersetzungnahtflaechen.go und listennahtflaechen.go und bestandssicherungnaht.go und markennaht.go und rueckkehrnaht.go und pflegenaht.go und pflegeflaechen.go und ruecksetzungsnaht.go und kostennaht.go und mtgjsonnaht.go und bilddateinaht.go und vorlagennaht.go und listeneinfuhrnaht.go und sammlungssichtennaht.go und massennaht.go und setwache.go und kartenskriptnaht.go und objektmarkennaht.go und auszaehlnaht.go und autoersetzungspflege.go und decknahtflaechen.go und deckstart.go und turnierflaeche.go und nutzernaht.go und profilnaht.go und startmigration.go und anmeldestart.go und tokenwache.go und leutestart.go und vorgabennaht.go und chiffrenaht.go und verdrahtung.go und nurlesen.go und retrievalnaht.go und verdrahtungzeichnen.go und recherchenaht.go und peernaht.go und kontaktnaht.go und freigabenaht.go und tauschnaht.go und handelspostnaht.go und handelsverbuchungsnaht.go und kiuebernahmenaht.go und kigruppe.go und umschlagflaeche.go und netzflaeche.go und partieflaeche.go und marktlinkenaht.go und deckeinfuhrnaht.go und kontoschaunaht.go und handelsleitungsnaht.go und symbolbibliotheksflaeche.go und zusatznaht.go erlaubt (Composition Root, AGENTS.md §2)",
		},
		{
			name:     "ein unbekanntes Paket wird abgewiesen",
			paket:    "irgendwas",
			importe:  []string{},
			erwartet: "das Paket steht weder in erlaubt noch in keineSchicht (AGENTS.md §2)",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			datei := fall.datei
			if datei == "" {
				datei = "irgendwo.go"
			}

			einfuhren := make([]Einfuhr, 0, len(fall.importe))
			for _, pfad := range fall.importe {
				einfuhren = append(einfuhren, Einfuhr{Datei: datei, Pfad: pfad})
			}

			befunde := pruefe(map[string][]Einfuhr{fall.paket: einfuhren})

			if len(befunde) != 1 {
				t.Fatalf("erwartet war genau ein Befund, gefunden: %v", befunde)
			}
			if befunde[0].Grund != fall.erwartet {
				t.Errorf("Grund %q, erwartet war %q", befunde[0].Grund, fall.erwartet)
			}
		})
	}
}

// TestErlaubteImporteErgebenKeinenBefund ist die Gegenprobe zur Gegenprobe.
//
// Eine Prüfung, die *alles* abweist, bestünde die Tabelle oben ebenfalls.
func TestErlaubteImporteErgebenKeinenBefund(t *testing.T) {
	erlaubteFaelle := map[string][]Einfuhr{
		"app": {
			{Datei: "app.go", Pfad: modulpfad + "/internal/storage"},
			{Datei: "app.go", Pfad: modulpfad + "/internal/service"},
			{Datei: "fehler.go", Pfad: modulpfad + "/internal/domain"},
			{Datei: "app.go", Pfad: "database/sql"},
		},
		"service": {
			{Datei: "import.go", Pfad: modulpfad + "/internal/msescript"},
			{Datei: "import.go", Pfad: modulpfad + "/internal/render"},
		},
		"storage": {
			{Datei: "datenbank.go", Pfad: modulpfad + "/internal/domain"},
			{Datei: "datenbank.go", Pfad: "database/sql"},
		},
		"domain":    {{Datei: "fehler.go", Pfad: "errors"}},
		"msescript": {{Datei: "lexer.go", Pfad: "strings"}},
		// ausgenommen, siehe keineSchicht
		"werkzeuge": {{Datei: "pinnung_test.go", Pfad: modulpfad + "/internal/app"}},
	}

	if befunde := pruefe(erlaubteFaelle); len(befunde) != 0 {
		t.Errorf("erwartet war kein Befund, gefunden: %v", befunde)
	}
}

// TestJedeAusnahmeHatEinenGrund hält fest, dass keine Ausnahme stillschweigend
// entsteht.
func TestJedeAusnahmeHatEinenGrund(t *testing.T) {
	for paket, grund := range keineSchicht {
		t.Run(paket, func(t *testing.T) {
			if strings.TrimSpace(grund) == "" {
				t.Errorf("%q ist von der Schichtenregel ausgenommen, aber ohne Grund", paket)
			}
		})
	}
}
