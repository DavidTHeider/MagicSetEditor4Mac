package service_test

// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// gesehen holt die Revision aus der Vorschau — so, wie eine Bedienung es tut.
//
// **Seit Befund 9 des Nachlaufs** zum Zweitblick über Phase 5.2 ist sie für
// jede löschende Art Pflicht: wer nicht hingesehen hat, löscht nicht. Ein Test,
// der `""` reicht, prüfte einen Weg, den es nicht mehr gibt.
func gesehen(
	t *testing.T, dienst *service.Ruecksetzung,
	art domain.Ruecksetzungsart, ziel service.Ruecksetzungsziel,
) string {
	t.Helper()

	vorschau, fehler := dienst.Vorschau(t.Context(), art, ziel)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	return vorschau.Revision
}

func ruecksetzung(doppel *ruecksetzungsdoppel) *service.Ruecksetzung {
	return service.NeueRuecksetzung(doppel, doppel, wache(), doppel.neuanfang)
}

// **Die Vorschau ändert nichts** — und nennt bei der Layoutrücksetzung genau
// die vier Schlüssel.
func TestDieLayoutvorschauNenntGenauIhreSchluessel(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungLayout, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if len(vorschau.Betroffen) != 1 ||
		vorschau.Betroffen[0].Anzahl != len(domain.Layouteinstellungen()) {
		t.Errorf("die Vorschau ist %+v", vorschau.Betroffen)
	}

	// **Und sie empfiehlt keine Sicherung** — eine Layoutrücksetzung nimmt
	// keine Daten weg.
	if vorschau.SicherungEmpfohlen {
		t.Error("die Layoutrücksetzung empfiehlt eine Sicherung")
	}

	if doppel.zurueck != 0 || doppel.geloescht != "" {
		t.Error("die Vorschau hat etwas angefasst")
	}
}

// **Eine offene Rückkehr sperrt die Profillöschung** (Vermerk 0088).
func TestEineOffeneRueckkehrSperrtDieProfilloeschung(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		offen: 3, profile: []domain.Profil{{ID: 3, Name: "Meins"}},
	}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 3, Name: "Meins"})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if vorschau.Bereit() {
		t.Fatal("die Vorschau gilt als bereit")
	}

	if vorschau.Hindernisse[0].Art != domain.RuecksetzungOffeneRueckkehr {
		t.Errorf("das Hindernis ist %+v", vorschau.Hindernisse[0])
	}

	// **Und die Ausführung weist ab, statt zu löschen.**
	if _, ausfuehrfehler := ruecksetzung(doppel).Zuruecksetzen(
		t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 3, Name: "Meins"}, "", ""); ausfuehrfehler == nil {
		t.Fatal("das gesperrte Profil wurde gelöscht")
	}

	if doppel.geloescht != "" {
		t.Errorf("gelöscht wurde %q", doppel.geloescht)
	}
}

// **Der letzte Mensch lässt sich nicht löschen** — und die Vorschau sagt es
// vorher.
func TestDerLetzteMenschLaesstSichNichtLoeschen(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{nutzer: 1, profile: []domain.Profil{{ID: 3, Name: "Meins"}}}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungNutzer,
		service.Ruecksetzungsziel{Name: "David"})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if vorschau.Bereit() {
		t.Fatal("die Vorschau gilt als bereit")
	}

	if vorschau.Hindernisse[0].Art != domain.RuecksetzungLetzterNutzer {
		t.Errorf("das Hindernis ist %+v", vorschau.Hindernisse[0])
	}
}

// **Ein zweiter Mensch macht den Weg frei** — die Gegenprobe zum Hindernis.
func TestEinZweiterMenschMachtDenWegFrei(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{nutzer: 2, profile: []domain.Profil{{ID: 3, Name: "Meins"}}}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungNutzer,
		service.Ruecksetzungsziel{Name: "David"})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if !vorschau.Bereit() {
		t.Errorf("die Vorschau ist nicht bereit: %+v", vorschau.Hindernisse)
	}

	// **Und sie empfiehlt eine Sicherung** — hier gehen Daten fort.
	if !vorschau.SicherungEmpfohlen {
		t.Error("die Nutzerlöschung empfiehlt keine Sicherung")
	}

	dienst := ruecksetzung(doppel)
	ziel := service.Ruecksetzungsziel{Name: "David"}

	if _, ausfuehrfehler := dienst.Zuruecksetzen(
		t.Context(), domain.RuecksetzungNutzer, ziel, "",
		gesehen(t, dienst, domain.RuecksetzungNutzer, ziel)); ausfuehrfehler != nil {
		t.Fatalf("löschen: %v", ausfuehrfehler)
	}

	if doppel.geloescht != "nutzer:David" {
		t.Errorf("gelöscht wurde %q — der Name ist die Bindung", doppel.geloescht)
	}
}

// **Die Umfänge mehrerer Profile werden zusammengezählt.**
func TestDieUmfaengeMehrererProfileWerdenZusammengezaehlt(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		nutzer: 2,
		profile: []domain.Profil{
			{ID: 1, Name: "Meins"}, {ID: 2, Name: "Deins"},
		},
		umfang: []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins"}},
		},
	}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungNutzer,
		service.Ruecksetzungsziel{Name: "David"})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	nachWas := map[string]domain.Betroffenes{}
	for _, gruppe := range vorschau.Betroffen {
		nachWas[gruppe.Was] = gruppe
	}

	if nachWas[domain.BetroffenSets].Anzahl != 6 {
		t.Errorf("es sind %d Sets, erwartet waren 6", nachWas[domain.BetroffenSets].Anzahl)
	}

	if nachWas[domain.BetroffenProfile].Anzahl != 2 ||
		len(nachWas[domain.BetroffenProfile].Namen) != 2 {
		t.Errorf("die Profile sind %+v", nachWas[domain.BetroffenProfile])
	}
}

// **Die ganze Anwendung legt den Bestand beiseite** — und nennt, wohin.
func TestDieGanzeAnwendungLegtDenBestandBeiseite(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{profile: []domain.Profil{{ID: 3, Name: "Meins"}}}

	dienst := ruecksetzung(doppel)

	bilanz, fehler := dienst.Zuruecksetzen(
		t.Context(), domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{}, "nachweis",
		gesehen(t, dienst, domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{}))
	if fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	if doppel.neuanfaenge != 1 {
		t.Errorf("es gab %d Neuanfänge", doppel.neuanfaenge)
	}

	if bilanz.Beiseite == "" {
		t.Error("die Bilanz nennt den Ort des alten Bestands nicht")
	}
}

// **Eine erfundene Art wird abgewiesen** — an beiden Wegen.
func TestEineErfundeneRuecksetzungsartWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}

	if _, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), "alles", service.Ruecksetzungsziel{}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die Vorschau gab %v", fehler)
	}

	if _, fehler := ruecksetzung(doppel).Zuruecksetzen(
		t.Context(), "alles", service.Ruecksetzungsziel{}, "", ""); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("das Zurücksetzen gab %v", fehler)
	}
}

// **Ein Fehler beim Nachsehen kommt durch** — statt einer leeren Vorschau,
// die wie „nichts betroffen" aussähe.
func TestEinFehlerBeimNachsehenDerRuecksetzungKommtDurch(t *testing.T) {
	t.Parallel()

	for name, art := range map[string]domain.Ruecksetzungsart{
		"die Profileinstellungen": domain.RuecksetzungProfileinstellungen,
		"ein Profil":              domain.RuecksetzungProfil,
		"ein Nutzer":              domain.RuecksetzungNutzer,
		"die ganze Anwendung":     domain.RuecksetzungAnwendung,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := &ruecksetzungsdoppel{fehler: errors.New("die Datenbank ist zu")}

			if _, fehler := ruecksetzung(doppel).Vorschau(
				t.Context(), art, service.Ruecksetzungsziel{},
			); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Und ein Fehler beim Ausführen ebenso.**
func TestEinFehlerBeimAusfuehrenDerRuecksetzungKommtDurch(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{nutzer: 2}

	// **Erst die Vorschau glücken lassen, dann das Ausführen scheitern** —
	// sonst prüfte der Lauf die Vorschau und nicht den Weg dahinter.
	weg := ruecksetzung(doppel)

	if _, fehler := weg.Zuruecksetzen(t.Context(), domain.RuecksetzungLayout,
		service.Ruecksetzungsziel{}, "", ""); fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	doppel.fehler = errors.New("die Platte ist voll")

	if _, fehler := weg.Zuruecksetzen(t.Context(), domain.RuecksetzungLayout,
		service.Ruecksetzungsziel{}, "", ""); fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}

// **Mehr Namen als die Grenze werden gekürzt** — eine Vorschau soll nicht die
// halbe Tabelle tragen.
func TestZuVieleNamenWerdenGekuerzt(t *testing.T) {
	t.Parallel()

	profile := make([]domain.Profil, 0, service.Ruecksetzungsbeispiele+3)
	for nummer := range service.Ruecksetzungsbeispiele + 3 {
		profile = append(profile,
			domain.Profil{ID: int64(nummer + 1), Name: string(rune('A' + nummer))})
	}

	doppel := &ruecksetzungsdoppel{
		nutzer:  2,
		profile: profile,
		umfang: []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 1, Namen: []string{"Eins", "Zwei"}},
		},
	}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungNutzer,
		service.Ruecksetzungsziel{Name: "David"})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	for _, gruppe := range vorschau.Betroffen {
		if len(gruppe.Namen) > service.Ruecksetzungsbeispiele {
			t.Errorf("%s nennt %d Namen", gruppe.Was, len(gruppe.Namen))
		}
	}
}

// **Jedes Hindernis trägt seinen Satz** — ein Fehler, der eine Rücksetzung
// anhält, ist ein Fehler und kein Zustand (§8).
func TestJedesHindernisTraegtSeinenSatz(t *testing.T) {
	t.Parallel()

	for name, doppel := range map[string]*ruecksetzungsdoppel{
		"eine offene Rückkehr": {
			offen: 1, nutzer: 2,
			profile: []domain.Profil{{ID: 3, Name: "Meins"}},
		},
		"der letzte Mensch": {
			nutzer:  1,
			profile: []domain.Profil{{ID: 3, Name: "Meins"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := ruecksetzung(doppel).Zuruecksetzen(
				t.Context(), domain.RuecksetzungNutzer,
				service.Ruecksetzungsziel{Name: "David"}, "", "")
			if fehler == nil {
				t.Fatal("das Hindernis kam durch")
			}

			// **Der Satz ist für Menschen** — eine Marke allein wäre keine
			// Antwort.
			if len(fehler.Error()) < 40 {
				t.Errorf("der Satz ist %q", fehler)
			}
		})
	}
}
