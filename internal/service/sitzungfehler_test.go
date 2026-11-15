package service_test

// **Die Fehlerwege der lokalen Sitzung** — ID-02
// (#1701).

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Neustart nimmt den festgehaltenen Zustand auf** — gesperrt bleibt
// gesperrt, abgemeldet abgemeldet.
func TestEinNeustartBehaeltDieSperre(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)
	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if doppel.festgehalten != domain.SitzungGesperrt {
		t.Fatalf("festgehalten: %q", doppel.festgehalten)
	}

	neu := begonnen(t, doppel, 1)
	if zustand, nutzer := neu.Stand(); zustand != domain.SitzungGesperrt || nutzer != 1 {
		t.Errorf("nach dem Neustart: %s %d", zustand, nutzer)
	}

	if fehler := neu.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, nutzer := begonnen(t, doppel, 1).Stand(); zustand != domain.SitzungAbgemeldet || nutzer != 0 {
		t.Errorf("abgemeldet nach dem Neustart: %s %d", zustand, nutzer)
	}

	if fehler := neu.Anmelden(t.Context(), 1, "richtig-und-lang", "", festesZiel); fehler != nil || doppel.festgehalten != domain.SitzungAngemeldet {
		t.Errorf("angemeldet: %q %v", doppel.festgehalten, fehler)
	}
}

// **Was sich nicht festhalten lässt, geschieht nicht** — sonst stünde der
// Bestand auf angemeldet, während das Fenster gesperrt zeigt.
func TestOhneFesthaltenKeinUebergang(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)
	doppel.festhaltefehler = errors.New("die Platte klemmt")

	if fehler := sitzung.Sperren(t.Context()); !errors.Is(fehler, doppel.festhaltefehler) {
		t.Errorf("%v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAngemeldet {
		t.Errorf("%s", zustand)
	}

	lesefehler := neuesSitzungsdoppel(t)
	lesefehler.fehler = errors.New("die Platte klemmt")

	if fehler := service.NeueSitzung(lesefehler, lesefehler).Beginnen(t.Context(), 1); fehler == nil {
		t.Error("ohne festgehaltenen Zustand begonnen")
	}
}

func TestEineFehlendeAuskunftDesSensorsSperrtNicht(t *testing.T) {
	t.Parallel()

	doppel := neuesSitzungsdoppel(t)
	doppel.gefragt = errors.New("die Faktorliste klemmt")
	sitzung := begonnen(t, doppel, 2)

	if fehler := sitzung.Sperren(t.Context()); !errors.Is(fehler, doppel.gefragt) {
		t.Errorf("%v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAngemeldet {
		t.Errorf("%s", zustand)
	}
}

// **Leer ist falsch, und eine kaputte Form bleibt ein Fehler** — beides lässt
// gesperrt.
func TestEntsperrenBleibtBeiJedemFehlerGesperrt(t *testing.T) {
	t.Parallel()

	doppel := neuesSitzungsdoppel(t)
	doppel.sensorDa = true
	ohne := begonnen(t, doppel, 2)

	if fehler := ohne.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ohne.Entsperren(t.Context(), "irgendein-kennwort"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Kennwort: %v", fehler)
	}

	sitzung, doppel := neueSitzung(t)
	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	doppel.nutzer[0].Kennwort = "$argon2id$kaputt"
	if fehler := sitzung.Entsperren(t.Context(), "richtig-und-lang"); fehler == nil {
		t.Error("eine kaputte Form entsperrte")
	}

	doppel.fehler = errors.New("die Platte klemmt")
	if fehler := sitzung.Entsperren(t.Context(), "richtig-und-lang"); !errors.Is(fehler, doppel.fehler) {
		t.Errorf("ohne Nutzerliste: %v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungGesperrt {
		t.Errorf("%s", zustand)
	}
}

// **Ein untaugliches Kennwort kostet keinen Code**, und ein Schreibfehler
// lässt die Sitzung, wo sie war.
func TestEineGescheiterteWiederherstellungVerbrauchtNichts(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "kurz"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("zu kurz: %v", fehler)
	}

	if !doppel.codes["zettel-1"] {
		t.Error("ein zu kurzes Kennwort hat den Code verbraucht")
	}

	// **Code und Kennwort in einem Geschäft** (Zweitblick über ID-02,
	// Befund 5): scheitert es, ist der Code nicht verbraucht und das
	// Kennwort das alte.
	doppel.codesetzfehler = errors.New("die Platte klemmt")
	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "wieder-da-und-lang"); !errors.Is(fehler, doppel.codesetzfehler) {
		t.Errorf("Schreibfehler: %v", fehler)
	}

	if !doppel.codes["zettel-1"] || kennwort.Stimmt(doppel.nutzer[0].Kennwort, "richtig-und-lang") != nil {
		t.Error("ein gescheitertes Geschäft hat Code oder Kennwort angefasst")
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAbgemeldet {
		t.Errorf("nach dem Schreibfehler: %s", zustand)
	}
}

// **Anmelden prüft das Kennwort, dann den Wechsel** — und bleibt abgemeldet,
// wenn eines von beiden scheitert.
func TestAnmeldenPrueftKennwortUndWechsel(t *testing.T) {
	t.Parallel()

	sitzung, _ := neueSitzung(t)
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	gewechselt := 0
	wechsel := zaehlendesZiel(&gewechselt)

	// **Angemeldet meldet sich niemand noch einmal an.**
	if fehler := begonnen(t, neuesSitzungsdoppel(t), 1).Anmelden(t.Context(), 1, "richtig-und-lang", "", wechsel); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("zweimal angemeldet: %v", fehler)
	}

	if fehler := sitzung.Anmelden(t.Context(), 1, "falsch", "", wechsel); !errors.Is(fehler, domain.ErrValidation) || gewechselt != 0 {
		t.Errorf("falsches Kennwort: %v, gewechselt %d", fehler, gewechselt)
	}

	scheitert := errors.New("das Profil fehlt")
	ohneProfil := service.Anmeldeziel{Profil: func(context.Context, int64) (int64, error) { return 0, scheitert }}

	if fehler := sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "", ohneProfil); !errors.Is(fehler, scheitert) {
		t.Errorf("Wechsel: %v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAbgemeldet {
		t.Errorf("nach gescheitertem Wechsel: %s", zustand)
	}

	if fehler := sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "", wechsel); fehler != nil || gewechselt != 1 {
		t.Errorf("%v, gewechselt %d", fehler, gewechselt)
	}

	// **Ein Nutzer, den es nicht gibt, meldet sich nicht an.**
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Anmelden(t.Context(), 99, "", "", festesZiel); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("unbekannt: %v", fehler)
	}
}

// **Wer den Nutzer wechselt, nimmt die Sitzung mit** — sonst sperrte
// `Sperren` danach den, der vorher arbeitete.
func TestEinNutzerwechselStelltDieSitzungUm(t *testing.T) {
	t.Parallel()

	sitzung, _ := neueSitzung(t)
	sitzung.Umgestellt(2)

	if _, nutzer := sitzung.Stand(); nutzer != 2 {
		t.Errorf("nach dem Wechsel: %d", nutzer)
	}

	// **Gesperrt wechselt niemand** — die Umstellung geht dort ins Leere.
	sitzung.Umgestellt(1)
	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	sitzung.Umgestellt(2)
	if _, nutzer := sitzung.Stand(); nutzer != 1 {
		t.Errorf("gesperrt umgestellt: %d", nutzer)
	}
}

func TestDieAuskunftNenntKennwortUndSensor(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)

	if auskunft, fehler := sitzung.Auskunft(t.Context()); fehler != nil || !auskunft.Kennwort || auskunft.Sensor {
		t.Errorf("%+v %v", auskunft, fehler)
	}

	doppel.sensorDa = true
	if auskunft, fehler := begonnen(t, doppel, 2).Auskunft(t.Context()); fehler != nil || auskunft.Kennwort || !auskunft.Sensor {
		t.Errorf("ohne Kennwort, mit Sensor: %+v %v", auskunft, fehler)
	}

	doppel.gefragt = errors.New("die Faktorliste klemmt")
	if _, fehler := sitzung.Auskunft(t.Context()); !errors.Is(fehler, doppel.gefragt) {
		t.Errorf("Sensorfehler: %v", fehler)
	}

	doppel.nutzer = nil
	if _, fehler := sitzung.Auskunft(t.Context()); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("verschwunden: %v", fehler)
	}
}

// **Der Stand antwortet auch mitten in einem Übergang** — der Nutzerwechsel
// beim Anmelden fragt über die Profilnaht nach ihm, und ein Schloss, das
// `Anmelden` hält, klemmte ihn für immer (gefunden am 2026-09-27 an
// `LokalAnmelden`). Er sagt dabei den Stand **vor** dem Übergang.
func TestDerStandAntwortetMittenImAnmelden(t *testing.T) {
	t.Parallel()

	sitzung, _ := neueSitzung(t)
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	erreicht := make(chan struct{}, 1)
	gesehen := make(chan domain.Sitzungszustand, 1)
	fertig := make(chan error, 1)

	go func() {
		fertig <- sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "", service.Anmeldeziel{
			Profil: festesZiel.Profil,
			Nachziehen: func(context.Context, int64) {
				erreicht <- struct{}{}

				zustand, _ := sitzung.Stand()
				gesehen <- zustand
			},
		})
	}()

	// **Die Frist gilt dem Schloss, nicht dem Kennwort.** Bis zum Nachziehen
	// liegt die Prüfung mit Argon2id, und die brauchte unter `-race` bei
	// voller Maschine 3,5 Sekunden (gemessen am 2026-09-27) — die Frist lief
	// ab, bevor jemand das Schloss auch nur hielt.
	select {
	case <-erreicht:
	case <-time.After(time.Minute):
		t.Fatal("das Anmelden kam nicht bis zum Nachziehen")
	}

	select {
	case zustand := <-gesehen:
		if zustand != domain.SitzungAbgemeldet {
			t.Errorf("mitten im Anmelden: %s", zustand)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("der Stand klemmt hinter dem Schloss des Anmeldens")
	}

	if fehler := <-fertig; fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, nutzer := sitzung.Stand(); zustand != domain.SitzungAngemeldet || nutzer != 1 {
		t.Errorf("danach: %s %d", zustand, nutzer)
	}
}

// **Jede Abweisung der Sitzung trägt ihren eigenen Schlüssel** — mit
// `validierung` zeigte die Oberfläche nur „Die Eingabe ist nicht gültig" und
// nicht, **was** nicht stimmt (§6).
func TestDieAbweisungenDerSitzungSagenWasNichtStimmt(t *testing.T) {
	t.Parallel()

	schluessel := func(fehler error) domain.Schluessel {
		var feld *domain.FieldError
		if !errors.As(fehler, &feld) {
			return ""
		}

		return feld.Key
	}

	ohne := begonnen(t, neuesSitzungsdoppel(t), 2)
	sitzung, _ := neueSitzung(t)

	for name, fall := range map[string]struct {
		fehler error
		soll   domain.Schluessel
	}{
		"falsches Kennwort": {sitzung.KennwortWechseln(t.Context(), "falsch", "ganz-neu-und-lang"), domain.SchluesselKennwortFalsch},
		"zu kurz":           {sitzung.KennwortWechseln(t.Context(), "richtig-und-lang", "kurz"), domain.SchluesselKennwortLaenge},
		"kein Rückweg":      {ohne.Sperren(t.Context()), domain.SchluesselSperreOhneRueckweg},
	} {
		if got := schluessel(fall.fehler); got != fall.soll {
			t.Errorf("%s: %q (%v)", name, got, fall.fehler)
		}
	}
}

// **Anmelden prüft den zweiten Faktor streng** (Zweitblick über ID-02,
// Befund 2) — auch für den, der schon arbeitete, und unabhängig davon, ob
// jemand ihn beim Profilwechsel abgeschaltet hat.
func TestAnmeldenPrueftDenZweitenFaktorStreng(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	umgestellt := 0
	umstellen := zaehlendesZiel(&umgestellt)

	doppel.faktorfehler = domain.FehlenderFaktor(domain.FaktorTOTP)
	if fehler := sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "", umstellen); !errors.Is(fehler, doppel.faktorfehler) || umgestellt != 0 {
		t.Errorf("ohne Faktor: %v, umgestellt %d", fehler, umgestellt)
	}

	doppel.faktorfehler = nil
	doppel.gerufen = nil

	if fehler := sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "123456", umstellen); fehler != nil || umgestellt != 1 {
		t.Fatalf("%v, umgestellt %d", fehler, umgestellt)
	}

	if !slices.Contains(doppel.gerufen, "Pruefe:123456") {
		t.Errorf("gerufen: %v", doppel.gerufen)
	}
}

// **Nach dem Abmelden setzt die Wiederherstellung nur das Kennwort** —
// angemeldet wird danach (Zweitblick über ID-02, Befund 3).
func TestWiederherstellenNachDemAbmeldenBleibtAbgemeldet(t *testing.T) {
	t.Parallel()

	sitzung, _ := neueSitzung(t)
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "wieder-da-und-lang"); fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, nutzer := sitzung.Stand(); zustand != domain.SitzungAbgemeldet || nutzer != 0 {
		t.Errorf("nach der Wiederherstellung: %s %d", zustand, nutzer)
	}

	if fehler := sitzung.Anmelden(t.Context(), 1, "wieder-da-und-lang", "", festesZiel); fehler != nil {
		t.Errorf("mit dem neuen Kennwort: %v", fehler)
	}
}
