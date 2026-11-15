package domain_test

// **Der Wächter über die Ebenentabelle** — Werkstück O3 von Phase 5.2
// (#1427),
// nach [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md).
//
// # Was er hält
//
// **Jeder Schlüssel beider Reihen steht in genau einer Ebene**, und keine
// Zeile der Tabelle nennt einen Schlüssel, den es nicht gibt. Beide
// Richtungen, weil je eine allein die Hälfte offenließe: eine zu kurze
// Tabelle ist ein Schlüssel ohne Ebene, eine zu lange eine Ebene ohne
// Schlüssel — und die zweite bleibt beim Umbenennen unbemerkt.
//
// # Was er nicht hält
//
// **Ob die Ebene die richtige ist.** Dass `vorgabe.hochladepfad` zum
// **Rechner** gehört und nicht zum Profil, ist eine fachliche Entscheidung;
// sie steht im Vermerk und wird von einem Menschen gelesen. Der Wächter sieht
// nur, dass eine dasteht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// alleSchluessel gibt die Vereinigung beider Reihen.
//
// **Eine Vereinigung und keine Verkettung**: vier Schlüssel stehen in beiden —
// `sprache` ist dieselbe Frage auf zwei Ebenen (`uebersteuerung.go`), und eine
// Liste mit Dubletten machte aus „genau eine Ebene" eine Aussage über Zeilen
// statt über Schlüssel.
func alleSchluessel() []string {
	gesehen := map[string]bool{}
	alle := []string{}

	for _, reihe := range [][]string{
		domain.Einstellungsreihe(), domain.Nutzereinstellungsreihe(),
	} {
		for _, schluessel := range reihe {
			if gesehen[schluessel] {
				continue
			}

			gesehen[schluessel] = true

			alle = append(alle, schluessel)
		}
	}

	return alle
}

func TestJedeEinstellungStehtInGenauEinerEbene(t *testing.T) {
	t.Parallel()

	erlaubt := map[domain.Ebene]bool{}
	for _, ebene := range domain.Ebenen() {
		erlaubt[ebene] = true
	}

	for _, schluessel := range alleSchluessel() {
		ebene, bekannt := domain.EbeneVon(schluessel)
		if !bekannt {
			t.Errorf("%q steht in keiner Ebene — dann weiß der Austausch "+
				"nicht, ob der Wert mitreisen darf", schluessel)

			continue
		}

		if !erlaubt[ebene] {
			t.Errorf("%q steht in der Ebene %q, und die gibt es nicht",
				schluessel, ebene)
		}
	}
}

func TestDieEbenentabelleKenntKeinenFremdenSchluessel(t *testing.T) {
	t.Parallel()

	gibtEs := map[string]bool{}
	for _, schluessel := range alleSchluessel() {
		gibtEs[schluessel] = true
	}

	for _, schluessel := range domain.Ebenenschluessel() {
		if !gibtEs[schluessel] {
			t.Errorf("die Ebenentabelle führt %q — diesen Schlüssel gibt es "+
				"in keiner der beiden Reihen", schluessel)
		}
	}
}

func TestEinUnbekannterSchluesselHatKeineEbene(t *testing.T) {
	t.Parallel()

	if ebene, bekannt := domain.EbeneVon("gibtesnicht"); bekannt {
		t.Errorf("für einen erfundenen Schlüssel kam die Ebene %q heraus", ebene)
	}
}

func TestDieEbeneEinesSchluessels(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		schluessel string
		ebene      domain.Ebene
	}{
		// **Die vier Erscheinungseinstellungen gehören dem Menschen** — das
		// steht so in `uebersteuerung.go`, und das Profil übersteuert sie nur.
		{domain.EinstellungSprache, domain.EbeneNutzer},
		{domain.EinstellungSchriftart, domain.EbeneNutzer},

		// **Der Rechner**: ob etwas ins Netz geht, ist eine Eigenschaft
		// dieser Maschine — und ein Pfad ist eine Eigenschaft dieser Platte.
		{domain.EinstellungAktualisierung, domain.EbeneGeraet},
		{domain.VorgabeHochladepfad, domain.EbeneGeraet},
		// **Die Seitenleiste ist eine Fensterfrage** — Werkstück U14 von Phase 5.1
		// (#1490): sie reist nicht mit dem Profil.
		{domain.EinstellungSeitenleiste, domain.EbeneGeraet},
		{domain.EinstellungLeistenlage, domain.EbeneGeraet},
		{domain.AnbieterKennung(domain.Anbieter()[0]), domain.EbeneGeraet},

		// **Das Profil**: womit ein Datensatz dieses Profils anfängt.
		{domain.VorgabeZustand, domain.EbeneProfil},
		{domain.DienstEinstellung(domain.DienstScryfall), domain.EbeneProfil},

		// **Der Arbeitsplatz**: wie das Fenster steht.
		{domain.ArbeitsplatzNavigator, domain.EbeneArbeitsplatz},
		{domain.ArbeitsplatzInspektorbreite, domain.EbeneArbeitsplatz},
	}

	for _, fall := range faelle {
		t.Run(fall.schluessel, func(t *testing.T) {
			t.Parallel()

			ebene, bekannt := domain.EbeneVon(fall.schluessel)
			if !bekannt {
				t.Fatalf("%q steht in keiner Ebene", fall.schluessel)
			}

			if ebene != fall.ebene {
				t.Errorf("%q steht in %q, erwartet war %q",
					fall.schluessel, ebene, fall.ebene)
			}
		})
	}
}
