package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Bindungen des Paketbezugs.
//
// **Eine eigene Datei**, weil `bindungen_test.go` an der Zeilengrenze steht
// (§2) — und weil der Schnitt einer Frage folgt: hier steht, was beim Beziehen
// über die Naht geht.
func TestDieBezugsbindungenErreichenIhreAnwendung(t *testing.T) {
	anwendung := app.NeueApp("")
	bindungen := app.NeueBindungen(anwendung)

	t.Run("PaketHerkunft und PaketBeziehen", func(t *testing.T) {
		probe := &bezugsprobe{ergebnis: &service.Bezugsergebnis{
			Ort: "/pakete/M15", Stand: "3.1.0", Pruefsumme: "sha256:abc",
		}}
		anwendung.MitPaketbezieher(probe)

		// **Auch der Katalog** (#56): der Bezug vermerkt danach Herkunft und
		// Prüfsumme. Ohne diese Naht stürzt der Lauf ab, statt sie zu prüfen —
		// derselbe Fall wie beim Datenstand oben.
		katalog := &paketfuehrer{}
		anwendung.MitPaketen(katalog)

		// **Die Herkunft zuerst**: die Oberfläche zeigt sie und bezieht erst
		// danach. Ein Durchreicher, der sie überspränge, wäre hier grün.
		if woher := bindungen.PaketHerkunft(t.Context(), "https://github.com/x/y", "3.1.0"); woher.Stand == "" {
			t.Errorf("die Herkunft ist leer: %+v", woher)
		}

		antwort := bindungen.PaketBeziehen(t.Context(), "https://github.com/x/y", "3.1.0", "data", "M15")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Pruefsumme != "sha256:abc" || probe.name != "M15" {
			t.Errorf("%+v, Name %q", antwort, probe.name)
		}

		// **Angezeigt und vermerkt** — die zweite Hälfte der Zusage aus #56.
		// Ohne sie stünde die Summe in der Antwort und sonst nirgends; nach dem
		// nächsten Start wäre sie fort.
		if katalog.vermerkt.Pruefsumme != "sha256:abc" ||
			katalog.vermerkt.Herkunft != "https://github.com/x/y@3.1.0" ||
			katalog.vermerkt.Ort != "/pakete/M15" {
			t.Errorf("vermerkt: %+v", katalog.vermerkt)
		}
	})
}
