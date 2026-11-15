// **Die Rolle wechselt, während MCP und Dienst laufen** — SEC01
// (#2354),
// Befund A-15: die Rollenmatrix (`rollenmatrix_test.go`) fährt jede Lage in
// frisch gestarteten Prozessen; hier ändert der Eigentümer die Freigabe
// **zwischen zwei Rufen derselben laufenden Prozesse**.
//
// # Die Folge
//
// Bob gibt sein Profil frei, und ich arbeite darin über die Kommandozeile,
// den Werkzeugserver und den Dienst — je mit vier Wirkungen: lesen, im
// Bestand schreiben, eine Datei schreiben, etwas hinausschicken.
//
//  1. **Mitarbeiter**: alles geht.
//  2. **Leser**, ohne Neustart: lesen geht, sonst nichts.
//  3. **Verwalter**, ohne Neustart: alles geht wieder — das Nein davor hieß
//     nicht, dass der Weg nicht ging (§10).
//  4. **Widerrufen** (`freigabe-loeschen`), ohne Neustart: lesen geht — wie
//     ohne Rolle in der Matrix —, sonst nichts.
//
// **Gezählt wird an der Wirkung**: Decks im Profil, Dateien auf der Platte,
// Anfragen an der Gegenstelle. Null unberechtigte Wirkungen heißt: nach den
// Schritten 2 und 4 kommt nichts hinzu.
package flaechen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDieRolleWechseltWaehrendMCPUndDienstLaufen(t *testing.T) {
	t.Parallel()

	var gefragt atomic.Int64

	server := gefragteGegenstelle(t, &gefragt)
	umgebung := append(eigenerBestand(t),
		app.Scryfallumgebung+"="+server.URL, "HOME="+t.TempDir(),
		schluesselbund.Bundort+"="+probebund(t))

	meins := mussKommandozeile(t, umgebung, "profil", "--json")
	ich, meinProfil := eignerkennung(t, meins), profilkennung(t, meins)
	bob, bobsProfil := profilEinesFremden(t, mussKommandozeile(t, umgebung, "nutzer-neu", "Bob", "--json"))

	// **Bob gibt frei und ändert** — als er selbst, danach wieder ich, und
	// wieder in seinem Profil: der Wechsel des Nutzers stellt das gezeigte
	// Profil auf das eigene zurück, und die laufenden Prozesse folgen ihm.
	//
	// **Und vorher ins eigene Profil**: in einem fremden, in dem ich nur
	// lesen darf, ist auch der Wechsel des Nutzers ein Schreiben.
	alsBob := func(argumente ...string) {
		mussKommandozeile(t, umgebung, "profil", meinProfil)
		mussKommandozeile(t, umgebung, "nutzer-wechseln", bob)
		mussKommandozeile(t, umgebung, argumente...)
		mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
		mussKommandozeile(t, umgebung, "profil", bobsProfil)
	}

	mussKommandozeile(t, umgebung, "nutzer-wechseln", bob)
	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
	mussKommandozeile(t, umgebung, "freigabe-setzen", bobsProfil, ich, "mitarbeiter")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", bobsProfil)

	// **Einmal gestartet, und dann nicht wieder.**
	mcp := starteMCPProzess(t, umgebung)
	an := starteDienst(t, umgebung)
	flaechen := map[string]rollenflaeche{
		"Kommandozeile": kommandozeilenflaeche(t, umgebung),
		"MCP":           mcpflaeche(mcp),
		"REST":          restflaeche(t, an),
	}

	ziel := t.TempDir()
	decks := 0

	for schritt, lage := range []struct {
		name    string
		aendern []string
		darf    bool
	}{
		{"Mitarbeiter", nil, true},
		{"Leser", []string{"freigabe-setzen", bobsProfil, ich, "leser"}, false},
		{"Verwalter", []string{"freigabe-setzen", bobsProfil, ich, "verwalter"}, true},
		{"widerrufen", []string{"freigabe-loeschen", bobsProfil, ich}, false},
	} {
		if lage.aendern != nil {
			alsBob(lage.aendern...)
		}

		vorher := gefragt.Load()

		for name, flaeche := range flaechen {
			wo := lage.name + " · " + name

			if gelesen, gut := flaeche.sets(); !gut || !strings.Contains(gelesen, "Probe") {
				t.Errorf("%s: lesen ging nicht (%v): %s", wo, gut, gelesen)
			}

			if _, gut := flaeche.deck(wo + " Deck"); gut != lage.darf {
				t.Errorf("%s: ein Deck anlegen ging %v, erwartet %v", wo, gut, lage.darf)
			}

			datei := filepath.Join(ziel, strconv.Itoa(schritt)+"-"+name+".json")
			if _, gut := flaeche.ausfuhr(setID, datei); gut != lage.darf {
				t.Errorf("%s: die Ausfuhr ging %v, erwartet %v", wo, gut, lage.darf)
			}

			if _, fehlt := os.Stat(datei); (fehlt == nil) != lage.darf {
				t.Errorf("%s: die Datei liegt da: %v, erwartet %v", wo, fehlt == nil, lage.darf)
			}

			// **Hinausschicken wird an der Gegenstelle gezählt**, nicht am
			// Rückgabewert: der Riegel steht an der Netzgrenze (AUTORISIERUNG).
			//
			// **Je Lage und Fläche ein eigener Begriff** (Zweitblick über
			// Phase 5, Befund 4): mit „probe" überall antwortete ab der ersten
			// Lage der Zwischenspeicher, und die Netzgrenze wurde nie
			// betreten — ein fortbestehendes Senderecht wäre so nicht
			// aufgefallen. Erlaubt muss die Anfrage **an der Gegenstelle**
			// ankommen; verboten darf nichts hinaus (unten).
			davor := gefragt.Load()

			if _, gut := flaeche.suche(fmt.Sprintf("probe %d %s", schritt, name)); gut != lage.darf {
				t.Errorf("%s: die Scryfallsuche ging %v, erwartet %v", wo, gut, lage.darf)
			}

			if lage.darf && gefragt.Load() == davor {
				t.Errorf("%s: die erlaubte Suche kam an der Gegenstelle nicht an", wo)
			}

			// **Was der Zwischenspeicher schon hat, bekommt auch ein Leser**
			// — öffentliche Daten, keine Anfrage. Der Begriff der ersten Lage
			// liegt dort.
			if lage.name == "Leser" {
				zuvor := gefragt.Load()

				if _, gut := flaeche.suche("probe 0 " + name); !gut || gefragt.Load() != zuvor {
					t.Errorf("%s: der Zwischenspeicher antwortete nicht ohne Anfrage (%v)", wo, gut)
				}
			}
		}

		// **An der Wirkung**: Decks im Profil und Anfragen an der Gegenstelle.
		if lage.darf {
			decks += len(flaechen)
		}

		if gezaehlt := deckeDerLage(t, umgebung); gezaehlt != decks {
			t.Errorf("%s: %d Decks im Profil, erwartet %d", lage.name, gezaehlt, decks)
		}

		if !lage.darf && gefragt.Load() != vorher {
			t.Errorf("%s: %d Anfragen gingen hinaus", lage.name, gefragt.Load()-vorher)
		}
	}
}
