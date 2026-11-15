// **Der Vorlagenbereich am Fenster** — Werkstück 2 von Phase 3.3
// (#1164).
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// werkstattprobe antwortet fest — oder scheitert.
type werkstattprobe struct {
	dateien []string
	beiwerk []string
	inhalt  string
	stand   string
	fehler  error

	// geschrieben hält fest, was ankam.
	paket, datei, wert, erwartet string

	// texte sind die Dateien samt Inhalt — der Stoff des Symbolindex (U7).
	texte []service.Paketdateitext

	// quelle und neu halten den Dateibaum fest (FMT-02).
	quelle, neu string
}

func (v *werkstattprobe) Texte(_ context.Context, paket string) ([]service.Paketdateitext, error) {
	v.paket = paket

	return v.texte, v.fehler
}

func (v *werkstattprobe) DateienUndBeiwerk(_ context.Context, paket string) (dateien, beiwerk []string, fehler error) {
	v.paket = paket

	return v.dateien, v.beiwerk, v.fehler
}

func (v *werkstattprobe) Anlegen(_ context.Context, paket, datei, inhalt string) (string, error) {
	v.paket, v.datei, v.wert = paket, datei, inhalt

	return v.stand, v.fehler
}

func (v *werkstattprobe) Einfuegen(_ context.Context, paket, quelle, datei string) error {
	v.paket, v.quelle, v.datei = paket, quelle, datei

	return v.fehler
}

func (v *werkstattprobe) Umbenennen(_ context.Context, paket, alt, neu, erwartet string) error {
	v.paket, v.datei, v.neu, v.erwartet = paket, alt, neu, erwartet

	return v.fehler
}

func (v *werkstattprobe) Entfernen(_ context.Context, paket, datei, erwartet string) error {
	v.paket, v.datei, v.erwartet = paket, datei, erwartet

	return v.fehler
}

func (v *werkstattprobe) Lesen(_ context.Context, paket, datei string) (inhalt, stand string, fehler error) {
	v.paket, v.datei = paket, datei

	return v.inhalt, v.stand, v.fehler
}

func (v *werkstattprobe) Schreiben(
	_ context.Context, paket, datei, inhalt, erwartet string,
) (string, error) {
	v.paket, v.datei, v.wert, v.erwartet = paket, datei, inhalt, erwartet

	return v.stand, v.fehler
}

// **Die Dateien eines Pakets kommen ans Fenster.**
func TestDieVorlagendateienKommenAnsFenster(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{dateien: []string{"style", "card_fields"}})

	antwort := anwendung.Paketdateien(t.Context(), "probe.mse-style")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Dateien) != 2 {
		t.Errorf("es kamen %v", antwort.Dateien)
	}
}

// **Eine leere Liste und kein `nil`** — dieselbe Zusage wie überall sonst.
func TestEinPaketOhneVorlagenGibtEineLeereListe(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{dateien: nil})

	if antwort := anwendung.Paketdateien(t.Context(), "probe.mse-style"); antwort.Dateien == nil {
		t.Error("die Dateien sind nil statt leer")
	}
}

// **Gelesen kommen Inhalt und Stand.**
func TestEineVorlageKommtMitIhremStand(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{inhalt: "mse version: 2.0.0\n", stand: "abc"})

	antwort := anwendung.PaketdateiLesen(t.Context(), "probe.mse-style", "style")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Inhalt != "mse version: 2.0.0\n" || antwort.Stand != "abc" {
		t.Errorf("Antwort %+v", antwort)
	}
}

// **Geschrieben wird mit dem Stand, den das Fenster hält — und es bekommt den
// neuen zurück.**
func TestEineVorlageWirdMitStandGeschrieben(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &werkstattprobe{stand: "neu"}
	anwendung.MitPaketdateien(probe)

	antwort := anwendung.PaketdateiSchreiben(t.Context(), "probe.mse-style", "style", "x\n", "alt")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.erwartet != "alt" || probe.wert != "x\n" {
		t.Errorf("weitergereicht: %+v", probe)
	}

	if antwort.Stand != "neu" {
		t.Errorf("der neue Stand ist %q", antwort.Stand)
	}
}

// **Ein Konflikt kommt als Konflikt an**, nicht als unbekannter Fehler: das
// Fenster muss ihn von einem Tippfehler unterscheiden können.
func TestEinVorlagenkonfliktKommtAlsKonfliktAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{fehler: domain.ErrConflict})

	antwort := anwendung.PaketdateiSchreiben(t.Context(), "probe.mse-style", "style", "x", "alt")
	if antwort.Fehler == nil {
		t.Fatal("der Konflikt kam als Erfolg durch")
	}

	if antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("der Code ist %q", antwort.Fehler.Code)
	}
}

// **Ohne eingerichteten Bereich sagt jede der drei Bindungen es.**
func TestOhneVorlagenbereichSagenAlleDreiEsAn(t *testing.T) {
	anwendung := app.NeueApp("")

	if antwort := anwendung.Paketdateien(t.Context(), "p"); antwort.Fehler == nil {
		t.Error("Vorlagendateien kam durch")
	}

	if antwort := anwendung.PaketdateiLesen(t.Context(), "p", "style"); antwort.Fehler == nil {
		t.Error("VorlageLesen kam durch")
	}

	if antwort := anwendung.PaketdateiSchreiben(t.Context(), "p", "style", "x", ""); antwort.Fehler == nil {
		t.Error("VorlageSchreiben kam durch")
	}
}

// **Und die Fehler der Naht kommen bei allen dreien an.**
func TestDieFehlerDerVorlagennahtKommenAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{fehler: errors.New("keine Platte")})

	if antwort := anwendung.Paketdateien(t.Context(), "p"); antwort.Fehler == nil {
		t.Error("Vorlagendateien kam durch")
	} else if antwort.Dateien == nil {
		t.Error("die Dateien sind nil statt leer")
	}

	if antwort := anwendung.PaketdateiLesen(t.Context(), "p", "style"); antwort.Fehler == nil {
		t.Error("VorlageLesen kam durch")
	}

	if antwort := anwendung.PaketdateiSchreiben(t.Context(), "p", "style", "x", ""); antwort.Fehler == nil {
		t.Error("VorlageSchreiben kam durch")
	}
}

// **Die Bindungen am Fenster gehen denselben Weg.**
func TestDieVorlagenbindungenGehenDenselbenWeg(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{
		dateien: []string{"style"}, inhalt: "a", stand: "s",
	})

	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.Paketdateien(t.Context(), "p"); len(antwort.Dateien) != 1 {
		t.Errorf("Vorlagendateien: %+v", antwort)
	}

	if antwort := bindungen.PaketdateiLesen(t.Context(), "p", "style"); antwort.Inhalt != "a" {
		t.Errorf("VorlageLesen: %+v", antwort)
	}

	if antwort := bindungen.PaketdateiSchreiben(t.Context(), "p", "style", "b", "s"); antwort.Stand != "s" {
		t.Errorf("VorlageSchreiben: %+v", antwort)
	}
}
