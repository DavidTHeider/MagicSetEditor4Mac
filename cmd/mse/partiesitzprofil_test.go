package main

// **Der Schalter `--sitzprofil=` bindet einen Sitz** — UXR-F21
// (#2041).
//
// **Hier steht, was die Form der Kommandozeile betrifft**: welche Schalter
// Sitzprofile sind, wohin sie gehören und was abgewiesen wird. Dass die
// Bindung an Dienst und MCP dasselbe tut und dass ein Profil, das es nicht
// gibt, überall ein Befund ist, hält `internal/flaechen/sitzprofil_test.go`.

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieSitzprofileWerdenVonDenSitzenGetrennt(t *testing.T) {
	t.Parallel()

	uebrig, profile := trenneSitzprofile([]string{
		"magic", "--sitzprofil=bert:3", "anna:Anna:1", "bert:Bert:2",
		"--sitzprofil=anna:4",
	})

	if want := []string{"magic", "anna:Anna:1", "bert:Bert:2"}; !slices.Equal(uebrig, want) {
		t.Errorf("übrig: %q, erwartet %q", uebrig, want)
	}

	// **In der Reihenfolge, in der sie dastanden.**
	if want := []string{"--sitzprofil=bert:3", "--sitzprofil=anna:4"}; !slices.Equal(profile, want) {
		t.Errorf("Profile: %q, erwartet %q", profile, want)
	}

	// **Ohne Schalter bleibt alles beim Alten** (§16, verträglich).
	uebrig, profile = trenneSitzprofile([]string{"magic", "anna:Anna:1"})
	if len(uebrig) != 2 || len(profile) != 0 {
		t.Errorf("ohne Schalter: %q und %q", uebrig, profile)
	}
}

func TestEinSitzprofilBindetDenGenanntenSitz(t *testing.T) {
	t.Parallel()

	wuensche := []app.Sitzwunsch{{Kennung: "anna"}, {Kennung: "bert"}}

	// **Dasselbe Profil an zwei Sitzen ist erlaubt**: wer gegen sich selbst
	// spielt, hat zwei.
	if fehler := sitzprofileBinden(wuensche,
		[]string{"--sitzprofil=bert:3", "--sitzprofil=anna:3"}); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if wuensche[0].ProfilID != "3" || wuensche[1].ProfilID != "3" {
		t.Errorf("gebunden: %+v", wuensche)
	}

	// **Ohne Schalter bindet nichts.**
	frei := []app.Sitzwunsch{{Kennung: "anna"}}
	if fehler := sitzprofileBinden(frei, nil); fehler != nil || frei[0].ProfilID != "" {
		t.Errorf("ohne Schalter: %v, %+v", fehler, frei)
	}
}

func TestEinSitzprofilNachDerFormWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		schalter []string
		nennt    string
	}{
		"ohne Doppelpunkt":            {[]string{"--sitzprofil=bert"}, "ist keines"},
		"ohne Sitz":                   {[]string{"--sitzprofil=:3"}, "ist keines"},
		"ohne Profil":                 {[]string{"--sitzprofil=bert:"}, "ist keines"},
		"ein Sitz, den es nicht gibt": {[]string{"--sitzprofil=carla:3"}, "kommt unter den Sitzen nicht vor"},
		"zweimal derselbe Sitz": {
			[]string{"--sitzprofil=bert:3", "--sitzprofil=bert:4"}, "ein zweites",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := sitzprofileBinden(
				[]app.Sitzwunsch{{Kennung: "anna"}, {Kennung: "bert"}}, fall.schalter)
			if fehler == nil || !strings.Contains(fehler.Error(), fall.nennt) {
				t.Errorf("Fehler %v, erwartet einer mit %q", fehler, fall.nennt)
			}
		})
	}
}

// **Auch am Befehl endet die Abweisung nach der Form mit dem Bedienungscode**
// — und zählt ein Sitzprofil nicht als Sitz.
func TestDieAbweisungNachDerFormEndetMitDemBedienungscode(t *testing.T) {
	t.Parallel()

	faelle := map[string][]string{
		"ein Sitz, den es nicht gibt": {
			"partie-anlegen", "magic", "anna:Anna:1", "bert:Bert:2",
			"--sitzprofil=carla:3",
		},
		"ein Schalter ersetzt keinen Sitz": {
			"partie-anlegen", "magic", "anna:Anna:1", "--sitzprofil=anna:3",
		},
		"kein Spiel": {"partie-anlegen"},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var aus, fehler bytes.Buffer

			code := lauf(probekontext(t), argumente, nil, &aus, &fehler,
				frischerBestand(t))
			if code != CodeBedienung {
				t.Errorf("Code %d (%s), erwartet %d", code, fehler.String(), CodeBedienung)
			}
		})
	}
}

// **Ein Befund steht in der Textausgabe** — wie bei den Wegen der Verwaltung.
func TestDieTextausgabeNenntDieBefunde(t *testing.T) {
	t.Parallel()

	var raus, rfehler bytes.Buffer

	an := &schreiber{ziel: &raus, fehler: &rfehler}

	antwort := app.PartieAntwort{
		Partie: app.PartieDTO{ID: "p1", Stand: "laeuft", Fassung: "1", Fuer: "anna"},
		Befunde: []string{
			`der Sitz "bert" ist an ein Profil gebunden, das es nicht mehr gibt (9)`,
		},
	}

	if code := partieAusgeben(an, antwort); code != CodeGut {
		t.Fatalf("Code %d (%s)", code, rfehler.String())
	}

	if !strings.Contains(raus.String(),
		"  Befund: der Sitz \"bert\" ist an ein Profil gebunden") {
		t.Errorf("kein Befund in der Ausgabe: %q", raus.String())
	}

	// **Ohne Befund steht keine Zeile da.**
	raus.Reset()

	if partieAusgeben(an, app.PartieAntwort{}) != CodeGut ||
		strings.Contains(raus.String(), "Befund") {
		t.Errorf("ohne Befund: %q", raus.String())
	}
}
