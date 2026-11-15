package austausch_test

// **Die Grenzen eines Umschlags** — Werkstück N4 von Phase 5.4
// (#1457).
//
// **Am Manifest und nicht am Inhalt**: eine Archivbombe behauptet im Manifest,
// klein zu sein, und ist es gepackt auch. Was sie groß macht, steht in den
// Größenangaben — und die werden geprüft, **bevor** etwas ausgepackt wird.

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// eineSumme ist eine gültige SHA-256 in ihrer Textform.
const eineSumme = "0000000000000000000000000000000000000000000000000000000000000000"

// mitManifest baut einen Umschlag um ein Manifest herum und liest ihn.
func mitManifest(
	t *testing.T, ich peer.Identitaet, manifest austausch.Manifest,
) error {
	t.Helper()

	roh, fehler := json.Marshal(manifest)
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	eintraege := map[string][]byte{
		"austausch.json": roh,
		"austausch.sig":  ich.Signiere(roh),
	}

	for _, objekt := range manifest.Objekte {
		if objekt.Name != "" {
			eintraege[objekt.Name] = []byte("x")
		}
	}

	_, lesefehler := austausch.Lies(baueUmschlag(t, eintraege), t.TempDir())

	return lesefehler
}

// einManifest gibt eines mit den genannten Objekten.
func einManifest(
	ich peer.Identitaet, objekte ...austausch.Objekt,
) austausch.Manifest {
	return austausch.Manifest{
		Fassung: 1, Urheber: string(ich.Kennung()), Objekte: objekte,
	}
}

func TestDieGrenzenEinesUmschlags(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	viele := make([]austausch.Objekt, 0, austausch.HoechstensObjekte+1)
	for i := range austausch.HoechstensObjekte + 1 {
		viele = append(viele, austausch.Objekt{
			Art: austausch.ArtKarte, Name: "k" + itoa(i) + ".json",
			Summe: eineSumme, Groesse: 1,
		})
	}

	for name, manifest := range map[string]austausch.Manifest{
		"ohne Objekte": einManifest(ich),

		"zu viele Objekte": einManifest(ich, viele...),

		// **Die Archivbombe**: gepackt ein Byte je Objekt, ausgepackt
		// angeblich mehr, als der Umschlag tragen darf.
		//
		// **Viele Objekte und nicht eines** — der erste Anlauf nahm ein
		// einzelnes von 2 GB, und das scheiterte schon an
		// [austausch.HoechstensObjekt]. Der Lauf war grün und maß die
		// falsche Grenze (§10).
		"zu groß insgesamt": einManifest(ich, bombe()...),

		"ein Objekt zu groß": einManifest(ich, austausch.Objekt{
			Art: austausch.ArtKarte, Name: "a.json",
			Summe: eineSumme, Groesse: austausch.HoechstensObjekt + 1,
		}),

		"negative Größe": einManifest(ich, austausch.Objekt{
			Art: austausch.ArtKarte, Name: "a.json",
			Summe: eineSumme, Groesse: -1,
		}),

		"ohne Prüfsumme": einManifest(ich, austausch.Objekt{
			Art: austausch.ArtKarte, Name: "a.json", Groesse: 1,
		}),

		"unbekannte Art": einManifest(ich, austausch.Objekt{
			Art: "zauberspruch", Name: "a.json", Summe: eineSumme, Groesse: 1,
		}),

		"derselbe Name zweimal": einManifest(ich,
			austausch.Objekt{
				Art: austausch.ArtKarte, Name: "a.json",
				Summe: eineSumme, Groesse: 1,
			},
			austausch.Objekt{
				Art: austausch.ArtSet, Name: "a.json",
				Summe: eineSumme, Groesse: 1,
			}),

		// **Ein Verweis nach draußen wäre eine Frage nach dem Bestand des
		// Empfängers** — #1457 verbietet sie: „keine beliebige
		// Besitz-/Existenzabfrage".
		"braucht etwas von draußen": einManifest(ich, austausch.Objekt{
			Art: austausch.ArtVorlage, Name: "a.json",
			Summe: eineSumme, Groesse: 1,
			Braucht: []string{"gibtesnicht.json"},
		}),

		"Name führt heraus": einManifest(ich, austausch.Objekt{
			Art: austausch.ArtKarte, Name: "../draussen.json",
			Summe: eineSumme, Groesse: 1,
		}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := mitManifest(t, ich, manifest); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("kam durch: %v", fehler)
			}
		})
	}
}

// bombe baut so viele Objekte, dass sie zusammen die Gesamtgrenze reißen —
// jedes einzelne aber unter der Einzelgrenze bleibt.
func bombe() []austausch.Objekt {
	je := austausch.HoechstensObjekt
	zahl := int(austausch.HoechstensGesamt/je) + 1

	objekte := make([]austausch.Objekt, 0, zahl)
	for i := range zahl {
		objekte = append(objekte, austausch.Objekt{
			Art: austausch.ArtKarte, Name: "b" + itoa(i) + ".json",
			Summe: eineSumme, Groesse: je,
		})
	}

	return objekte
}

// itoa ohne strconv — der Lauf braucht nur kleine Zahlen.
func itoa(zahl int) string {
	if zahl == 0 {
		return "0"
	}

	var ziffern []byte

	for zahl > 0 {
		ziffern = append([]byte{byte('0' + zahl%10)}, ziffern...)
		zahl /= 10
	}

	return string(ziffern)
}
