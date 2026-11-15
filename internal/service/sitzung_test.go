package service_test

// **Die lokale Sitzung** — ID-02.AK-01/03/04
// (#1701).

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sitzungsdoppel hält Nutzer und merkt sich, was geschrieben wurde.
type sitzungsdoppel struct {
	nutzer   []domain.Nutzer
	gesetzt  []string
	codes    map[string]bool
	sensorDa bool
	gefragt  error

	// festgehalten ist, was der Bestand über die Sitzung weiß;
	// festhaltefehler trifft nur das Festhalten.
	festgehalten    domain.Sitzungszustand
	festhaltefehler error
	sensor          error
	fehler          error
	gerufen         []string

	// faktorfehler ist die Antwort der strengen Faktorprüfung beim Anmelden;
	// codesetzfehler trifft nur das Geschäft aus Code und Kennwort.
	faktorfehler   error
	codesetzfehler error
	zeiger         [2]int64
}

func neuesSitzungsdoppel(t *testing.T) *sitzungsdoppel {
	t.Helper()

	gehuetet, fehler := kennwort.Huete("richtig-und-lang")
	if fehler != nil {
		t.Fatal(fehler)
	}

	return &sitzungsdoppel{
		nutzer: []domain.Nutzer{
			{ID: 1, Name: "Ada", Kennwort: gehuetet},
			{ID: 2, Name: "Ich"},
		},
		codes: map[string]bool{"zettel-1": true},
	}
}

func (d *sitzungsdoppel) Nutzer(context.Context) ([]domain.Nutzer, error) { return d.nutzer, d.fehler }

func (d *sitzungsdoppel) KennwortSetzen(_ context.Context, nutzerID int64, gehuetet string) error {
	d.gerufen = append(d.gerufen, "KennwortSetzen")

	for stelle := range d.nutzer {
		if d.nutzer[stelle].ID == nutzerID {
			d.nutzer[stelle].Kennwort = gehuetet
		}
	}

	d.gesetzt = append(d.gesetzt, gehuetet)

	return d.fehler
}

func (d *sitzungsdoppel) CodeFinden(_ context.Context, _ int64, code string) (int64, error) {
	d.gerufen = append(d.gerufen, "CodeFinden")

	if !d.codes[code] {
		return 0, domain.NewFieldError("faktor", domain.SchluesselCodeFalsch, "Dieser Code stimmt nicht.")
	}

	return 1, nil
}

// KennwortMitCodeSetzen verbraucht Code 1 und setzt das Kennwort — **beides
// oder keines**, wie das Geschäft im Bestand.
func (d *sitzungsdoppel) KennwortMitCodeSetzen(_ context.Context, nutzerID, _ int64, gehuetet string, danach domain.Sitzungszustand) error {
	d.gerufen = append(d.gerufen, "KennwortMitCodeSetzen")

	if d.codesetzfehler != nil {
		return d.codesetzfehler
	}

	delete(d.codes, "zettel-1")
	d.festgehalten = danach

	for stelle := range d.nutzer {
		if d.nutzer[stelle].ID == nutzerID {
			d.nutzer[stelle].Kennwort = gehuetet
		}
	}

	return nil
}

func (d *sitzungsdoppel) Pruefe(_ context.Context, _ int64, nachweis string) error {
	d.gerufen = append(d.gerufen, "Pruefe:"+nachweis)

	return d.faktorfehler
}

func (d *sitzungsdoppel) SitzungLesen(context.Context) (domain.Sitzungszustand, error) {
	if d.festgehalten == "" {
		return domain.SitzungAngemeldet, d.fehler
	}

	return d.festgehalten, d.fehler
}

func (d *sitzungsdoppel) SitzungFesthalten(_ context.Context, zustand domain.Sitzungszustand) error {
	if d.festhaltefehler != nil {
		return d.festhaltefehler
	}

	d.festgehalten = zustand

	return nil
}

func (d *sitzungsdoppel) SensorMoeglich(context.Context, int64) (bool, error) {
	return d.sensorDa, d.gefragt
}

func (d *sitzungsdoppel) SensorPruefen(context.Context, int64) error {
	d.gerufen = append(d.gerufen, "SensorPruefen")

	return d.sensor
}

func neueSitzung(t *testing.T) (*service.Sitzung, *sitzungsdoppel) {
	t.Helper()

	doppel := neuesSitzungsdoppel(t)

	return begonnen(t, doppel, 1), doppel
}

func begonnen(t *testing.T, doppel *sitzungsdoppel, nutzerID int64) *service.Sitzung {
	t.Helper()

	sitzung := service.NeueSitzung(doppel, doppel)
	if fehler := sitzung.Beginnen(t.Context(), nutzerID); fehler != nil {
		t.Fatal(fehler)
	}

	return sitzung
}

func TestSperrenUndEntsperrenMitKennwort(t *testing.T) {
	t.Parallel()

	sitzung, _ := neueSitzung(t)

	if zustand, nutzer := sitzung.Stand(); zustand != domain.SitzungAngemeldet || nutzer != 1 {
		t.Fatalf("Beginn: %s %d", zustand, nutzer)
	}

	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Sperren(t.Context()); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("zweimal gesperrt: %v", fehler)
	}

	// **Ein falsches Kennwort lässt gesperrt.**
	if fehler := sitzung.Entsperren(t.Context(), "falsch"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("falsch: %v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungGesperrt {
		t.Errorf("nach dem Fehlversuch: %s", zustand)
	}

	if fehler := sitzung.Entsperren(t.Context(), "richtig-und-lang"); fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAngemeldet {
		t.Errorf("nach dem Entsperren: %s", zustand)
	}

	// **Entsperren, was nicht gesperrt ist, ist ein Konflikt.**
	if fehler := sitzung.Entsperren(t.Context(), "richtig-und-lang"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("zweimal: %v", fehler)
	}
}

// **Ohne Kennwort und ohne Sensor ließe sich eine Sperre nicht lösen** —
// deshalb wird gar nicht erst gesperrt.
func TestOhneKennwortWirdNichtGesperrt(t *testing.T) {
	t.Parallel()

	doppel := neuesSitzungsdoppel(t)
	sitzung := begonnen(t, doppel, 2)

	if fehler := sitzung.Sperren(t.Context()); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Kennwort: %v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAngemeldet {
		t.Errorf("gesperrt, obwohl sich nichts lösen ließe: %s", zustand)
	}

	// **Mit eingerichteter lokaler Entsperrung genügt der Sensor.**
	doppel.sensorDa = true
	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Errorf("mit Sensor: %v", fehler)
	}
}

func TestEntsperrenUeberDenSensor(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)

	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	doppel.sensor = errors.New("abgebrochen")
	if fehler := sitzung.EntsperrenUeberDenSensor(t.Context()); fehler == nil {
		t.Error("ein abgebrochener Sensor entsperrte")
	}

	doppel.sensor = nil
	if fehler := sitzung.EntsperrenUeberDenSensor(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAngemeldet {
		t.Errorf("%s", zustand)
	}

	if fehler := sitzung.EntsperrenUeberDenSensor(t.Context()); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("nicht gesperrt: %v", fehler)
	}
}

func TestAbmeldenUndAnmelden(t *testing.T) {
	t.Parallel()

	sitzung, _ := neueSitzung(t)

	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, nutzer := sitzung.Stand(); zustand != domain.SitzungAbgemeldet || nutzer != 0 {
		t.Errorf("abgemeldet: %s %d", zustand, nutzer)
	}

	if fehler := sitzung.Abmelden(t.Context()); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("zweimal: %v", fehler)
	}

	if fehler := sitzung.Anmelden(t.Context(), 2, "", "", festesZiel); fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, nutzer := sitzung.Stand(); zustand != domain.SitzungAngemeldet || nutzer != 2 {
		t.Errorf("angemeldet: %s %d", zustand, nutzer)
	}
}

func TestEinKennwortWechseltNurMitDemAlten(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)

	if fehler := sitzung.KennwortWechseln(t.Context(), "falsch", "ganz-neu-und-lang"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("falsch: %v", fehler)
	}

	if fehler := sitzung.KennwortWechseln(t.Context(), "richtig-und-lang", ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("leer: %v", fehler)
	}

	if fehler := sitzung.KennwortWechseln(t.Context(), "richtig-und-lang", "ganz-neu-und-lang"); fehler != nil {
		t.Fatal(fehler)
	}

	if len(doppel.gesetzt) != 1 || kennwort.Stimmt(doppel.gesetzt[0], "ganz-neu-und-lang") != nil {
		t.Errorf("gesetzt: %v", doppel.gesetzt)
	}

	// **Gesperrt wechselt kein Kennwort** — es ginge an der Sperre vorbei.
	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.KennwortWechseln(t.Context(), "ganz-neu-und-lang", "noch-eins-und-lang"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("gesperrt: %v", fehler)
	}
}

// **Wer noch keines hat, setzt eines ohne altes.**
func TestEinErstesKennwortBrauchtKeinAltes(t *testing.T) {
	t.Parallel()

	doppel := neuesSitzungsdoppel(t)
	sitzung := begonnen(t, doppel, 2)

	if fehler := sitzung.KennwortWechseln(t.Context(), "", "erstes-und-lang"); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Errorf("mit Kennwort sperrt es: %v", fehler)
	}
}

// **Die Wiederherstellung nimmt einen Code vom Zettel und setzt nur das
// Kennwort** (AK-03/04): keine Rolle, keine Freigabe, kein Peer.
func TestDieWiederherstellungSetztNurDasKennwort(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)

	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "falsch", "wieder-da-und-lang"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("falscher Code: %v", fehler)
	}

	// **Gesperrt stellt nur der Gesperrte wieder her.**
	if fehler := sitzung.Wiederherstellen(t.Context(), 2, "zettel-1", "wieder-da-und-lang"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("fremder Nutzer: %v", fehler)
	}

	doppel.gerufen = nil
	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "wieder-da-und-lang"); fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(doppel.gerufen, []string{"CodeFinden", "KennwortMitCodeSetzen"}) {
		t.Errorf("gerufen: %v", doppel.gerufen)
	}

	if zustand, nutzer := sitzung.Stand(); zustand != domain.SitzungAngemeldet || nutzer != 1 {
		t.Errorf("%s %d", zustand, nutzer)
	}

	// **Ein Code gilt einmal.**
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "noch-einmal-lang"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("zweimal derselbe Code: %v", fehler)
	}

	// **Das alte Kennwort gilt nicht mehr, das neue schon.**
	if fehler := sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "", festesZiel); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("das alte Kennwort: %v", fehler)
	}

	// **Angemeldet gibt es nichts wiederherzustellen.**
	if fehler := sitzung.Anmelden(t.Context(), 1, "wieder-da-und-lang", "", festesZiel); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "x-und-lang"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("angemeldet: %v", fehler)
	}
}

func TestFehlerDerAblageKommenAn(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)
	doppel.fehler = errors.New("die Platte klemmt")

	if fehler := sitzung.Sperren(t.Context()); fehler == nil {
		t.Error("Sperren ohne Nutzerliste")
	}

	if fehler := sitzung.KennwortWechseln(t.Context(), "richtig-und-lang", "neu-und-lang"); fehler == nil {
		t.Error("Wechsel ohne Nutzerliste")
	}

	// **Ein Nutzer, den es nicht mehr gibt, sperrt nicht.**
	doppel.fehler = nil
	doppel.nutzer = nil

	if fehler := sitzung.Sperren(t.Context()); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("verschwundener Nutzer: %v", fehler)
	}
}
