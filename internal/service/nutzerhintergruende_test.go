// **Tisch- und Fensterbild gehen mit dem Menschen** — Befund G4 des
// Zweitblicks über Phase 6.3, verlagert nach Phase 6.5
// (#1666).
//
// # Der Befund
//
// **Dieselbe Klasse wie beim Abzeichen, zum dritten Mal.** Das Tischbild kam
// mit Werkstück G3 von Phase 5.5, das Fensterbild mit UX-09; beide liegen
// unter `nutzer/` neben dem Foto, beide gehören einem Menschen — und
// `NutzerLoeschen` kannte nur Foto und Abzeichen. Wer einen Menschen löschte,
// ließ seine beiden Hintergründe auf der Platte.
//
// # Warum über die echte Ablage
//
// **Weil die Zusage auf der Platte gilt** und nicht an einer Schnittstelle:
// ein Doppel belegte, dass gerufen wird, nicht dass danach nichts mehr
// daliegt. Und weil der Wächter darunter die Ablage selbst fragt, welche
// Bilder sie kennt.
package service_test

import (
	"image"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// personenbilder sind die Arten, unter denen die Ablage ein Bild **eines
// Menschen** oder seiner Profile ablegt — gegen [service.Bildablage]
// gezählt im Wächter darunter.
//
// **`Ablegen` steht nicht darin**: es legt ein Kartenbild unter `beiwerk/<setID>/`
// ab, und das gehört einem Set.
var personenbilder = []string{
	"FensterbildAblegen", "NutzerbildAblegen", "ProfilbadgeAblegen", "TischbildAblegen",
}

// **Eine neue Bildart erzwingt die Entscheidung**, ob das Löschen eines
// Menschen sie mitnimmt — sonst entstünde sie still, wie Tisch- und
// Fensterbild.
//
// **Was er nicht hält**: ob die neue Art im Lauf darunter auch abgelegt wird.
// Das ist die Zeile, die er erzwingt — wer sie in `personenbilder` einträgt,
// steht vor dem Lauf, der sie braucht.
func TestJedeBildartDerAblageIstEntschieden(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[*service.Bildablage]()

	var gefunden []string

	for i := range typ.NumMethod() {
		if name := typ.Method(i).Name; strings.HasSuffix(name, "Ablegen") && name != "Ablegen" {
			gefunden = append(gefunden, name)
		}
	}

	slices.Sort(gefunden)

	if !slices.Equal(gefunden, personenbilder) {
		t.Errorf("die Ablage kennt %v, entschieden sind %v — gehört die neue "+
			"Art einem Menschen, nimmt `NutzerLoeschen` sie mit", gefunden, personenbilder)
	}
}

// **Nach dem Löschen liegt von ihm nichts mehr** — Foto, Abzeichen,
// Tischbild, Fensterbild. **Und von dem anderen alles**: wer nach dem Muster
// räumte, könnte `tischbild-30.png` mit wegnehmen.
func TestEinGeloeschterNutzerNimmtTischUndFensterbildMit(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)
	bild := image.NewNRGBA(image.Rect(0, 0, 4, 4))

	for _, nutzer := range []int64{3, 30} {
		legeAlleBilder(t, ablage, nutzer, bild)
	}

	doppel := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 3, Name: "Wer"}, {ID: 30, Name: "Andere"}},
		profile: map[int64][]domain.Profil{3: {{ID: 3, NutzerID: 3}}},
	}

	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(ablage, &geheimnisdoppel{})

	if fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	uebrig := dateienUnter(t, ablage.Nutzerverzeichnis())

	for _, name := range uebrig {
		if strings.Contains(name, "-3.") || strings.Contains(name, "-3-") {
			t.Errorf("%s liegt nach dem Löschen noch da", name)
		}
	}

	// **Je Art ein Bild des anderen** — sein Abzeichen liegt unter Profil 30,
	// und das gehört nicht dem Gelöschten.
	if len(uebrig) != len(personenbilder) {
		t.Errorf("übrig sind %v — erwartet waren die %d Bilder des anderen",
			uebrig, len(personenbilder))
	}
}

// **Und was sich nicht wegnehmen lässt, wird genannt** — je Bild, damit wer
// nachsieht, weiß wonach (§8). Der Mensch ist trotzdem fort (§7, Regel 3).
func TestEinLiegengebliebenerHintergrundWirdGenannt(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	// **Ein Verzeichnis mit Inhalt unter dem Namen des Bildes** — `os.Remove`
	// weist es ab, und zwar auf jeder Plattform.
	for _, name := range []string{"tischbild-3.png", "fensterbild-3.png"} {
		klemmt := filepath.Join(ablage.Nutzerverzeichnis(), name)
		if err := os.MkdirAll(klemmt, 0o750); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if err := os.WriteFile(filepath.Join(klemmt, "drin"), []byte("x"), 0o600); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	doppel := &nutzerdoppel{nutzer: []domain.Nutzer{{ID: 3, Name: "Wer"}}}
	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(ablage, &geheimnisdoppel{})

	fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer")
	if fehler == nil {
		t.Fatal("die liegengebliebenen Hintergründe wurden verschwiegen")
	}

	if doppel.geloescht != 3 {
		t.Error("das Löschen wurde zurückgenommen")
	}

	grund := domain.GrundVon(fehler)
	for _, teil := range []string{"Tischbild", "Fensterbild"} {
		if grund == nil || !strings.Contains(grund.Error(), teil) {
			t.Errorf("der Grund lautet %v und nennt %s nicht", grund, teil)
		}
	}
}

// legeAlleBilder legt unter jeder Art aus [personenbilder] ein Bild ab — das
// Abzeichen unter dem Profil mit derselben Kennung.
func legeAlleBilder(t *testing.T, ablage *service.Bildablage, kennung int64, bild image.Image) {
	t.Helper()

	for _, ablegen := range []func(int64, image.Image) (string, error){
		ablage.FensterbildAblegen, ablage.NutzerbildAblegen,
		ablage.ProfilbadgeAblegen, ablage.TischbildAblegen,
	} {
		if _, fehler := ablegen(kennung, bild); fehler != nil {
			t.Fatalf("ablegen: %v", fehler)
		}
	}
}

// dateienUnter gibt die Namen im Verzeichnis.
func dateienUnter(t *testing.T, ort string) []string {
	t.Helper()

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	namen := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		namen = append(namen, eintrag.Name())
	}

	return namen
}
