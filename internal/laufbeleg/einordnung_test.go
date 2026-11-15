package laufbeleg_test

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// **Ein Beleg sagt, ob er zum heutigen Stand passt** (AK-04) — **über den
// Inhalt dessen, was er misst, und nicht über den Commit**. Hier wird jeder
// PR per Squash gemergt: danach liegt der gemessene Commit in keiner
// Geschichte mehr, und nach der Frist von `git gc` gibt es ihn nicht mehr.
// Derselbe Inhalt an einem anderen Commit ist derselbe Stand.
//
// **Und der Inhalt ist mehr als der Index**: ein Arbeitsbaum, der dort
// abweicht, wo der Lauf misst, und eine Pflichteingabe mit anderem Inhalt
// sind ein anderer Stand — angestrichen vom Zweitblick über Phase 6.0.
func TestEinBelegWirdEingeordnet(t *testing.T) {
	t.Parallel()

	beleg := laufbeleg.Belegen(einKopf(), []laufbeleg.Eingabe{eineEingabe()}, eineAuswertung(), nil)
	jetzt := beleg.Ende.Add(24 * time.Hour)
	heute := laufbeleg.Heute{Inhalt: "sha256:inhalt", Eingaben: []laufbeleg.Eingabe{eineEingabe()}}

	schmutzig := beleg
	schmutzig.BaumSauber = false

	ohneInhalt := beleg
	ohneInhalt.Inhalt = ""

	kurz := beleg
	kurz.Commit = "abc"

	neu := heute
	neu.Inhalt = "sha256:neu"

	// **Die erste Zeile kommt ohne führenden Leerraum** — der Werkzeugkasten
	// schneidet die Ausgabe von `git` zu, und aus » M« wird »M«.
	offen := heute
	offen.Offen = []string{"M internal/app/k.go", " M internal/app/l.go", "?? internal/app/neu.go"}

	faelle := map[string]struct {
		beleg      laufbeleg.Beleg
		heute      laufbeleg.Heute
		frist      time.Duration
		einordnung laufbeleg.Einordnung
		grund      string
	}{
		"derselbe Inhalt":                      {beleg, heute, 0, laufbeleg.Aktuell, "gemessen an 7a1c915 am 2026-09-22"},
		"ein anderer Inhalt":                   {beleg, neu, 0, laufbeleg.AndererStand, "seit 7a1c915 hat sich geändert, was der Lauf misst"},
		"an einem schmutzigen Baum":            {schmutzig, heute, 0, laufbeleg.AndererStand, "ungesicherten Änderungen"},
		"ein Beleg ohne Inhaltshash":           {ohneInhalt, heute, 0, laufbeleg.AndererStand, "ohne Inhaltshash"},
		"über der Frist":                       {beleg, heute, time.Hour, laufbeleg.Ueberfaellig, "gemessen am 2026-09-22, älter als die Frist von 1h0m0s"},
		"ein kurzer Commit bleibt, wie er ist": {kurz, heute, 0, laufbeleg.Aktuell, "gemessen an abc am 2026-09-22"},
		"ein Arbeitsbaum mit Änderungen, wo der Lauf misst": {
			beleg, offen, 0, laufbeleg.AndererStand,
			"der Arbeitsbaum weicht ab, wo der Lauf misst: internal/app/k.go, internal/app/l.go, internal/app/neu.go",
		},
		"viele Änderungen, fünf beim Namen": {
			beleg,
			laufbeleg.Heute{Inhalt: "sha256:inhalt", Offen: []string{
				"M a.go", "M b.go", "M c.go", "M d.go", "M e.go", "M f.go", "?? g.go",
			}},
			0, laufbeleg.AndererStand,
			"wo der Lauf misst: a.go, b.go, c.go, d.go, e.go und 2 weitere",
		},
		"eine Eingabe mit anderem Inhalt": {
			beleg, mitEingabe(heute, func(e *laufbeleg.Eingabe) { e.Hash = "sha256:cd" }), 0,
			laufbeleg.AndererStand, "die Eingabe MSE_KANON hat heute einen anderen Inhalt",
		},
		"eine Eingabe, die heute fehlt": {
			beleg, mitEingabe(heute, func(e *laufbeleg.Eingabe) { e.Hash, e.Fehlt = "", "nicht gesetzt" }), 0,
			laufbeleg.AndererStand, "die Eingabe MSE_KANON ist heute nicht zu prüfen: nicht gesetzt",
		},
		"eine Eingabe, die heute niemand erhebt": {
			beleg,
			laufbeleg.Heute{Inhalt: "sha256:inhalt"},
			0,
			laufbeleg.AndererStand, "die Eingabe MSE_KANON ist heute nicht zu prüfen: nicht erhoben",
		},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			einordnung, grund := laufbeleg.Ordne(einer.beleg, einer.heute, jetzt, einer.frist)
			if einordnung != einer.einordnung || !strings.Contains(grund, einer.grund) {
				t.Errorf("%s, %q — erwartet %s und %q", einordnung, grund, einer.einordnung, einer.grund)
			}
		})
	}
}

// **Was beim Messen fehlte und heute noch fehlt, ist derselbe Stand** — ein
// Beleg, der BLOCKED sagt, bleibt BLOCKED und wird nicht zusätzlich »anders«.
// **Liegt die Eingabe heute vor, ist es ein anderer**: dann lohnt die Messung.
func TestEineFehlendeEingabeBleibtDerselbeStand(t *testing.T) {
	t.Parallel()

	fehlend := eineEingabe()
	fehlend.Hash, fehlend.Fehlt = "", "leer"

	beleg := laufbeleg.Belegen(einKopf(), []laufbeleg.Eingabe{fehlend}, eineAuswertung(), nil)
	jetzt := beleg.Ende.Add(time.Hour)

	if einordnung, grund := laufbeleg.Ordne(beleg, laufbeleg.Heute{
		Inhalt: "sha256:inhalt", Eingaben: []laufbeleg.Eingabe{fehlend},
	}, jetzt, 0); einordnung != laufbeleg.Aktuell {
		t.Errorf("weiterhin fehlend: %s, %q", einordnung, grund)
	}

	if einordnung, grund := laufbeleg.Ordne(beleg, laufbeleg.Heute{
		Inhalt: "sha256:inhalt", Eingaben: []laufbeleg.Eingabe{eineEingabe()},
	}, jetzt, 0); einordnung != laufbeleg.AndererStand ||
		!strings.Contains(grund, "die Eingabe MSE_KANON hat heute einen anderen Inhalt") {
		t.Errorf("heute vorhanden: %s, %q", einordnung, grund)
	}
}

// **Wer die Eingaben nicht rechnet, bekommt kein »aktuell«** — die
// Reichweite am Ende von `make ci` kennt sie nicht und soll zwölf Gigabyte
// nicht bei jedem Lauf lesen. Sie sagt dann, was sie nicht geprüft hat.
// **Was sie geprüft hat, zählt trotzdem**: ein anderer Inhalt bleibt ein
// anderer Stand, und eine überschrittene Frist bleibt überfällig.
func TestOhneEingabenHeisstEsNichtAktuell(t *testing.T) {
	t.Parallel()

	beleg := laufbeleg.Belegen(einKopf(), []laufbeleg.Eingabe{eineEingabe()}, eineAuswertung(), nil)
	jetzt := beleg.Ende.Add(time.Hour)

	faelle := map[string]struct {
		heute      laufbeleg.Heute
		frist      time.Duration
		einordnung laufbeleg.Einordnung
		grund      string
	}{
		"sonst derselbe Stand": {
			laufbeleg.Heute{Inhalt: "sha256:inhalt", OhneEingaben: true},
			0,
			laufbeleg.OhneEingaben, "gemessen an 7a1c915 am 2026-09-22 — ob die Eingaben dieselben sind, sagt make belege",
		},
		"ein anderer Inhalt": {
			laufbeleg.Heute{Inhalt: "sha256:neu", OhneEingaben: true},
			0,
			laufbeleg.AndererStand, "hat sich geändert",
		},
		"über der Frist": {
			laufbeleg.Heute{Inhalt: "sha256:inhalt", OhneEingaben: true},
			time.Minute,
			laufbeleg.Ueberfaellig, "älter als die Frist",
		},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			einordnung, grund := laufbeleg.Ordne(beleg, einer.heute, jetzt, einer.frist)
			if einordnung != einer.einordnung || !strings.Contains(grund, einer.grund) {
				t.Errorf("%s, %q — erwartet %s und %q", einordnung, grund, einer.einordnung, einer.grund)
			}
		})
	}
}

// mitEingabe gibt den heutigen Stand mit einer veränderten ersten Eingabe.
func mitEingabe(heute laufbeleg.Heute, aendern func(*laufbeleg.Eingabe)) laufbeleg.Heute {
	eingabe := heute.Eingaben[0]
	aendern(&eingabe)
	heute.Eingaben = []laufbeleg.Eingabe{eingabe}

	return heute
}

// **Womit gemessen wurde, gehört zum Stand** — Nachlauf zum Zweitblick über
// Phase 6.0: eine andere Go-Fassung, eine andere Plattform oder ein anderer
// Schalter (`MSE_STICHPROBE=alle` misst andere Karten) hieß bis dahin
// »aktuell«. **Ohne Eingaben** bleiben die Schalter ungeprüft, die Fassung
// nicht — sie liegt auch der Pipeline vor.
func TestWerkzeugUndSchalterGehoerenZumStand(t *testing.T) {
	t.Parallel()

	kopf := einKopf()
	kopf.Schalter = map[string]string{"MSE_STICHPROBE": "alle"}

	beleg := laufbeleg.Belegen(kopf, []laufbeleg.Eingabe{eineEingabe()}, eineAuswertung(), nil)
	jetzt := beleg.Ende.Add(time.Hour)

	heute := func(aendern func(*laufbeleg.Heute)) laufbeleg.Heute {
		stand := laufbeleg.Heute{
			Inhalt: "sha256:inhalt", Eingaben: []laufbeleg.Eingabe{eineEingabe()},
			Werkzeuge: map[string]string{"go": "go1.27.1", "plattform": "darwin/arm64"},
			Schalter:  map[string]string{"MSE_STICHPROBE": "alle"},
		}
		aendern(&stand)

		return stand
	}

	faelle := map[string]struct {
		heute      laufbeleg.Heute
		einordnung laufbeleg.Einordnung
		grund      string
	}{
		"derselbe Stand": {heute(func(*laufbeleg.Heute) {}), laufbeleg.Aktuell, "gemessen an 7a1c915"},
		"eine andere Go-Fassung": {
			heute(func(h *laufbeleg.Heute) { h.Werkzeuge["go"] = "go1.27.2" }),
			laufbeleg.AndererStand, "gemessen mit go go1.27.1, heute go1.27.2",
		},
		"eine andere Plattform": {
			heute(func(h *laufbeleg.Heute) { h.Werkzeuge["plattform"] = "linux/amd64" }),
			laufbeleg.AndererStand, "gemessen mit plattform darwin/arm64, heute linux/amd64",
		},
		"ein Schalter, der heute fehlt": {
			heute(func(h *laufbeleg.Heute) { h.Schalter["MSE_STICHPROBE"] = "" }),
			laufbeleg.AndererStand, "gemessen mit MSE_STICHPROBE=alle, heute ohne",
		},
		"ein Schalter mit anderem Wert": {
			heute(func(h *laufbeleg.Heute) { h.Schalter["MSE_STICHPROBE"] = "10" }),
			laufbeleg.AndererStand, "gemessen mit MSE_STICHPROBE=alle, heute MSE_STICHPROBE=10",
		},
		"ein Schalter, den der Beleg nicht kennt": {
			heute(func(h *laufbeleg.Heute) { h.Schalter["MSE_NEU"] = "10" }),
			laufbeleg.AndererStand, "gemessen ohne MSE_NEU, heute MSE_NEU=10",
		},
		"ohne Eingaben bleibt der Schalter ungeprüft": {
			heute(func(h *laufbeleg.Heute) { h.Schalter, h.Eingaben, h.OhneEingaben = nil, nil, true }),
			laufbeleg.OhneEingaben, "ob die Eingaben dieselben sind",
		},
		"ohne Eingaben zählt die Fassung trotzdem": {
			heute(func(h *laufbeleg.Heute) {
				h.Schalter, h.Eingaben, h.OhneEingaben = nil, nil, true
				h.Werkzeuge["go"] = "go1.28.0"
			}),
			laufbeleg.AndererStand, "heute go1.28.0",
		},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			einordnung, grund := laufbeleg.Ordne(beleg, einer.heute, jetzt, 0)
			if einordnung != einer.einordnung || !strings.Contains(grund, einer.grund) {
				t.Errorf("%s, %q — erwartet %s und %q", einordnung, grund, einer.einordnung, einer.grund)
			}
		})
	}
}
