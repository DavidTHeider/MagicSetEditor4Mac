package app_test

// **Ein Modell liest ein Kartenbild** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Was hier geprüft wird
//
// **Dass ein Anbieter ohne Sicht gar nicht erst gefragt wird** — die Meldung
// nennt ihn und sagt, woran es liegt.
//
// **Dass aus der Antwort Beobachtungen werden** — und dass ausdrücklich
// dasteht, wozu das Modell **nichts** gesagt hat.
//
// **Und dass nichts geschrieben wird.**

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einProbebildAufPlatte schreibt ein winziges PNG und gibt seinen **Pfad**.
//
// **Nicht `einProbebild`**: den Namen trägt schon der Bildweg aus A12, und er
// gibt die **Bytes**. Zwei Bedeutungen unter einem Namen sind eine Falle für
// den nächsten Leser — dieselbe, die A13.3 schon dreimal gefunden hat.
//
// **Ein wirklich kodiertes PNG und keine Signatur**: der Weg liest den Kopf
// mit [image.DecodeConfig], und acht Bytes Signatur kämen dort nicht durch.
// Der erste Anlauf schrieb genau die acht — und prüfte damit die Abweisung
// statt der Erkennung, sobald der Schnupperer wich.
func einProbebildAufPlatte(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "karte.png")

	var kodiert bytes.Buffer
	if fehler := png.Encode(&kodiert,
		image.NewRGBA(image.Rect(0, 0, 1, 1))); fehler != nil {
		t.Fatalf("das Probebild kodieren: %v", fehler)
	}

	if fehler := os.WriteFile(pfad, kodiert.Bytes(), 0o600); fehler != nil {
		t.Fatalf("das Probebild schreiben: %v", fehler)
	}

	return pfad
}

// bilddeuterprobe gibt eine feste Deutung — und hält fest, was hereinkam.
type bilddeuterprobe struct {
	gesehen domain.Bilderkennung
	deutung service.Bilddeutung
	fehler  error
}

func (b *bilddeuterprobe) Deuten(
	_ context.Context, erkennung domain.Bilderkennung,
) (service.Bilddeutung, error) {
	b.gesehen = erkennung

	return b.deutung, b.fehler
}

// einesichtwelt ist die KI-Welt mit einem Modell, das **Bilder liest**.
//
// **Ein eigener Bauer und kein Schalter in [einekiwelt]**: das gewöhnliche
// Probemodell darf ausdrücklich **nicht** sehen — sonst wäre der Lauf über
// „dieser Anbieter liest keine Bilder" nicht zu stellen.
func einesichtwelt(t *testing.T) kiwelt {
	t.Helper()

	welt := einekiwelt(t)
	welt.laufzeit.modelle = append(welt.laufzeit.modelle, einSeher("sieht:8b"))

	return welt
}

// **Ohne Bilderkenner scheitert der Weg, bevor er liest.**
func TestOhneBilderkennerScheitertDieErkennung(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{Pfad: "egal.png", Anbieter: "anthropic"})

	if antwort.Fehler == nil {
		t.Fatal("ohne Bilderkenner kam kein Fehler")
	}

	// **Und die Listen sind Listen** — „nichts gelesen" und „die Frage ging
	// schief" sind zwei Auskünfte.
	if antwort.Beobachtungen == nil || antwort.Kandidaten == nil {
		t.Errorf("eine Liste ist nil: %+v", antwort)
	}
}

// **Eine Datei, die kein Bild ist, wird abgewiesen — am Inhalt.**
//
// **Und nicht an der Endung**: eine Datei, die `.png` heißt und Text enthält,
// bekäme sonst eine Kopfzeile, die der Anbieter zu Recht abweist — und die
// Meldung spräche dann von etwas anderem.
func TestEineDateiDieKeinBildIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "karte.png")
	if fehler := os.WriteFile(pfad, []byte("kein Bild"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	welt := einekiwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{Pfad: pfad, Anbieter: "ollama", Alias: "gemma4:8b"})

	if antwort.Fehler == nil {
		t.Fatal("die Textdatei kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "kein lesbares Bild") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}

	// **Die Datei steht in der Meldung, der Grund nicht** (§8, 0075):
	// „image: unknown format" ist eine Auskunft übers Protokoll.
	if !strings.Contains(antwort.Fehler.Meldung, "karte.png") ||
		strings.Contains(antwort.Fehler.Meldung, "unknown format") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Und eine Datei, die wie ein PNG anfängt und keins ist, ebenso.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil die Klasse dahinter eine andere ist.** Der Lauf darüber prüft eine
// Textdatei — die scheitert an jedem Prüfverfahren. Diese hier trägt die
// **richtige Signatur** und wäre einem Schnupperer, der nur die ersten Bytes
// vergleicht, ein Bild: sie ginge bis zum Anbieter und dort auf die Rechnung.
func TestEinePngSignaturAlleinIstKeinBild(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "karte.png")
	roh := append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a},
		[]byte("und dahinter Unsinn")...)

	if fehler := os.WriteFile(pfad, roh, 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	welt := einekiwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{Pfad: pfad, Anbieter: "ollama", Alias: "gemma4:8b"})

	if antwort.Fehler == nil {
		t.Fatal("die halbe Datei kam durch")
	}
}

// **Ohne Pfad ebenso.**
func TestOhnePfadWirdNichtsGelesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	if antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{Anbieter: "ollama", Alias: "gemma4:8b"},
	); antwort.Fehler == nil {
		t.Fatal("ohne Pfad wurde gelesen")
	}
}

// **Eine Datei, die es nicht gibt, meldet sich als solche.**
func TestEineFehlendeDateiMeldetSich(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	if antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad:     filepath.Join(t.TempDir(), "gibtesnicht.png"),
			Anbieter: "ollama", Alias: "gemma4:8b",
		}); antwort.Fehler == nil {
		t.Fatal("die fehlende Datei kam durch")
	}
}

// **Ein Anbieter, der keine Bilder liest, wird gar nicht erst gefragt.**
//
// **Die Meldung nennt ihn** — damit niemand rät, welcher von vieren gemeint
// ist — und sie sagt, was stattdessen nötig wäre.
func TestEinAnbieterOhneSichtWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama", Alias: "gemma4:8b",
		})

	if antwort.Fehler == nil {
		t.Fatal("der Anbieter ohne Sicht kam durch")
	}

	// **Die Meldung nennt Modell und Fähigkeit** — und nicht den Anbieter.
	// Die Fähigkeitsprüfung des Werks greift **vor** der Anbieterprüfung, und
	// das ist die genauere Auskunft: nicht »ollama sieht nicht«, sondern
	// »dieses Modell hat es nicht belegt«.
	for _, teil := range []string{"gemma4:8b", "vision"} {
		if !strings.Contains(antwort.Fehler.Meldung, teil) {
			t.Errorf("die Meldung nennt %q nicht: %q", teil,
				antwort.Fehler.Meldung)
		}
	}
}

// **Eine gelesene Karte wird zu Beobachtungen — und das Bild geht mit.**
//
// # Was dieser Lauf hält
//
// **Dass das Bild wirklich hinausgeht**: eine Antwort käme auch ohne eines,
// und die Zusage „das Modell sieht die Karte" wäre dann unbelegt.
//
// **Dass ausdrücklich dasteht, wozu nichts gesagt wurde.** Eine kürzere Liste
// sieht aus wie ein Versehen; „zur Sammlernummer steht nichts" ist eine
// Antwort.
func TestEineGeleseneKarteWirdZuBeobachtungen(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	deuter := &bilddeuterprobe{deutung: service.Bilddeutung{
		Kandidaten: []service.Bildkandidat{},
	}}

	welt.anwendung.MitBilderkenner(deuter)
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"beobachtungen":[` +
			`{"art":"name","wert":"Lightning Bolt","lesbar":true},` +
			`{"art":"nummer","wert":"","lesbar":false}]}`,
		Modell: "claude-probe",
	}

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
			Alias: "sieht:8b",
		})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(welt.laufzeit.gesehene) != 1 ||
		len(welt.laufzeit.gesehene[0].Bild) == 0 {
		t.Fatalf("das Bild ging nicht mit: %+v", welt.laufzeit.gesehene)
	}

	if welt.laufzeit.gesehene[0].Medientyp != "image/png" {
		t.Errorf("der Medientyp ist %q", welt.laufzeit.gesehene[0].Medientyp)
	}

	if len(deuter.gesehen.Beobachtungen) != 2 {
		t.Errorf("die Beobachtungen sind %+v", deuter.gesehen.Beobachtungen)
	}

	// **Vier der sechs Arten kamen nicht vor** — und stehen deshalb da.
	if len(deuter.gesehen.NichtGelesen) != 4 {
		t.Errorf("nicht gelesen wurde %v", deuter.gesehen.NichtGelesen)
	}

	if antwort.Modell != "claude-probe" {
		t.Errorf("das Modell ist %q", antwort.Modell)
	}
}

// **Eine abgeschnittene Antwort wird verworfen.**
//
// **Weil eine halbe Leseliste aussieht wie eine ganze** — und dann fehlte
// genau die Stelle, die den Druck festgelegt hätte.
func TestEineAbgeschnitteneLeselisteWirdVerworfen(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"beobachtungen":[{"art":"name"`, Abgeschnitten: true,
	}

	if antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
			Alias: "sieht:8b",
		}); antwort.Fehler == nil {
		t.Fatal("die abgeschnittene Antwort kam durch")
	}
}

// **Und jeder weitere Fehlerweg reicht durch.**
func TestJederFehlerwegDerErkennungReichtDurch(t *testing.T) {
	for name, kaputt := range map[string]func(*bilddeuterprobe, *laufzeitdoppel){
		"kein JSON": func(_ *bilddeuterprobe, l *laufzeitdoppel) {
			l.antwort = app.KIRohantwort{Inhalt: "kaputt"}
		},
		"das Modell klemmt": func(_ *bilddeuterprobe, l *laufzeitdoppel) {
			l.redefehler = fmt.Errorf("%w: keine Antwort", domain.ErrValidation)
		},
		"der Deuter klemmt": func(d *bilddeuterprobe, _ *laufzeitdoppel) {
			d.fehler = fmt.Errorf("%w: die Suche klemmt", domain.ErrValidation)
		},
	} {
		t.Run(name, func(t *testing.T) {
			welt := einesichtwelt(t)
			deuter := &bilddeuterprobe{}

			welt.anwendung.MitBilderkenner(deuter)
			welt.laufzeit.antwort = app.KIRohantwort{
				Inhalt: `{"beobachtungen":[{"art":"name","wert":"x","lesbar":true}]}`,
			}

			kaputt(deuter, welt.laufzeit)

			antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
				app.KIErkennungsbezug{
					Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
					Alias: "sieht:8b",
				})
			if antwort.Fehler == nil {
				t.Fatal("der Fehler kam nicht durch")
			}

			// **Und die Listen bleiben Listen.**
			if antwort.Beobachtungen == nil || antwort.Kandidaten == nil ||
				antwort.Befunde == nil || antwort.NichtGelesen == nil {
				t.Errorf("eine Liste ist nil: %+v", antwort)
			}
		})
	}
}
