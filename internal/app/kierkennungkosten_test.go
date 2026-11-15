package app_test

// **Was eine Kartenerkennung kostet** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum eine eigene Datei
//
// **Weil sie eine andere Frage beantwortet** — und weil
// `kierkennungfenster_test.go` mit ihr an die Zeilengrenze aus §2 kam. Dort
// steht, was **nur das Fenster** tut: nach der Datei fragen. Hier steht, was
// vor jedem Auftrag steht, der anderswo rechnet: die Freigabe eines Menschen
// ([0078](../../docs/entscheidungen/0078-ein-nachweis-nur-am-fenster.md)).
//
// **Und ein Bild ist dabei die größte Ausgabe**: es kostet bei jedem
// Anbieter ein Vielfaches eines Satzes.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einCloudseher ist ein Cloudmodell, das **Bilder liest** — A13.3 (#1452).
func einCloudseher() domain.KIModell {
	modell := einSeher(derCloudalias)
	modell.Anbieter = derCloudanbieter
	modell.Ort = domain.KIOrtCloud

	return modell
}

// **Ohne Freigabe liest ein bezahltes Modell kein Foto.**
//
// # Warum das eigens dasteht
//
// **Weil ein Bild teurer ist als ein Satz.** Die Kostenwache verlangt vor
// jedem Auftrag, der anderswo rechnet, die ausdrückliche Freigabe eines
// Menschen ([0078](../../docs/entscheidungen/0078-ein-nachweis-nur-am-fenster.md));
// eine Sichtanfrage ist keine Ausnahme davon, und an ihr fällt am meisten an.
//
// **Und der Beleg steht daneben**: es ging nichts hinaus. Ein Fehler *nach*
// der Anfrage wäre eine Rechnung mit einer Fehlermeldung.
func TestOhneFreigabeLiestEinBezahltesModellKeinFoto(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)
	welt.fremd.modelle = append(welt.fremd.modelle, einCloudseher())
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: derCloudanbieter,
			Alias: derCloudalias,
		})

	if antwort.Fehler == nil {
		t.Fatal("der bezahlte Auftrag ging ohne Freigabe hinaus")
	}

	if len(welt.fremd.gesehene) != 0 {
		t.Errorf("es ging trotzdem ein Bild hinaus: %d", len(welt.fremd.gesehene))
	}
}

// **Und eine Freigabe, die es nicht gibt, trägt ihn auch nicht.**
//
// **Ein anderer Zweig als der darüber**: dort fehlt der Nachweis, hier wird
// einer **genannt** — und die Wache sieht in der Ablage nach. Eine erfundene
// Kennung dürfte nicht wie eine gültige wirken.
func TestEineErfundeneFreigabeTraegtKeineErkennung(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)
	welt.fremd.modelle = append(welt.fremd.modelle, einCloudseher())
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: derCloudanbieter,
			Alias: derCloudalias, Nachweis: "gibtesnicht",
		})

	if antwort.Fehler == nil {
		t.Fatal("die erfundene Freigabe kam durch")
	}

	if len(welt.fremd.gesehene) != 0 {
		t.Errorf("es ging trotzdem ein Bild hinaus: %d", len(welt.fremd.gesehene))
	}
}
