//go:build fremdpakete

// Die weitesten Karten neben ihrem Referenzbild
// (#523).
//
// **Ansehen, nicht rechnen.** #498 sagt es selbst: die drei größten Funde des
// 2026-08-18 kamen aus dem Vergleich zweier Bilder, nicht aus einer Zahl — die
// gedrehte Aftermath-Hälfte (#486), die doppelte Saga (#482) und die fehlenden
// Stufenkosten (#494).
//
// Dieser Lauf **rechnet nichts ab**. Er legt Bilder hin, damit ein Mensch sie
// ansieht: unseres, MSE2s, und die Differenz.
package app

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// WeitesteJeSatz ist die Zahl der Karten, die je Satz abgelegt werden.
const WeitesteJeSatz = 3

// abstandsbilder ist der Ort für die Ablage — **neben** den Referenzdaten und
// nicht im Baum (§11).
func abstandsbilder() string {
	// **`MSE_ABSTANDSBILDER` geht vor `HOME`** (#1170).
	//
	// Hier ist der Preis ein anderer als bei den Lesewegen: dies ist ein
	// **Schreibort**. Unter dem Testzuhause landeten die weitesten Karten in
	// einem Verzeichnis unter `/tmp`, das `TestMain` nach dem Lauf wegräumt —
	// abgelegt wurde also, und ansehen konnte es niemand.
	if ort := os.Getenv("MSE_ABSTANDSBILDER"); ort != "" {
		return ort
	}

	return filepath.Join(os.Getenv("HOME"), "Library", "Application Support",
		"MagicSetEditor4Mac", "abstandsbilder")
}

// TestDieWeitestenKartenAblegen legt je Satz die weitesten Karten ab.
func TestDieWeitestenKartenAblegen(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	// **Erst räumen, dann ablegen**
	// (#537).
	//
	// Der Dateiname trägt den gerundeten Abstand; ändert sich die Auswahl oder
	// die Zahl, bleibt der alte Dreiklang liegen. Nach zwei Läufen lagen 48
	// PNGs dort, während die Roadmap zwölf Karten führte — und welche davon
	// den Messstand bilden, war nicht mehr zu erkennen. Angestrichen vom
	// Zweitblick über Phase I.
	if err := os.RemoveAll(abstandsbilder()); err != nil {
		t.Fatalf("Ablage räumen: %v", err)
	}

	if err := os.MkdirAll(abstandsbilder(), 0o750); err != nil {
		t.Fatalf("Ablage anlegen: %v", err)
	}

	for _, satz := range saetzeMitReferenz(t) {
		if _, err := os.Stat(mse2ausgabe(satz)); err != nil {
			continue
		}

		legeWeitesteAb(t, anwendung, satz)
	}

	t.Logf("abgelegt unter %s", abstandsbilder())
}

// kartenabstand ist eine Karte samt ihrem Abstand.
type kartenabstand struct {
	karte   KarteDTO
	name    string
	vorlage string
	abstand float64
}

// legeWeitesteAb misst ein Set je Vorlage und legt die weitesten Karten ab.
func legeWeitesteAb(t *testing.T, anwendung *App, satz string) {
	t.Helper()

	geoeffnet := anwendung.SetOeffnen(t.Context(), filepath.Join(kanonverzeichnis(),
		satz+".mse-set"))
	if geoeffnet.Fehler != nil {
		t.Errorf("%s öffnen: %+v", satz, geoeffnet.Fehler)

		return
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Errorf("%s: Karten: %+v", satz, karten.Fehler)

		return
	}

	desSets := gewaehlterStil(t, anwendung, geoeffnet.SetID)

	verluste := ausfaelle{}

	// **Genannte Karten kommen hin, gleich wie weit sie sind — und vor der
	// Stichprobe.**
	//
	// Ein Befund über eine bestimmte Karte
	// (#529)
	// war bisher gar nicht nachzusehen: [nachVorlage] **zieht eine
	// Stichprobe**, und die beiden Karten aus dem Befund waren nicht darin.
	// Wer `WeitesteJeSatz` hochsetzte, sah sie trotzdem nicht.
	//
	//	MSE_KARTEN="Kumano Faces Kakkazan,Jugan Defends the Temple" …
	//
	// **Und ihr Abstand wird gemeldet**, denn genau danach fragt so ein
	// Befund: hat sich etwas bewegt oder nicht?
	for i := range karten.Karten {
		if !genannt(kartenname(karten.Karten[i])) {
			continue
		}

		vorlage := vorlageDerKarte(karten.Karten[i], desSets)

		eintrag, ok := abstandEiner(t, anwendung, satz, karten.Karten[i], vorlage, verluste)
		if !ok {
			t.Logf("  %-10s %-30s — kein Vergleich möglich", satz,
				kartenname(karten.Karten[i]))

			continue
		}

		t.Logf("  %-10s %-30s %-28s %7.3f  (genannt)", satz, eintrag.name,
			eintrag.vorlage, eintrag.abstand)

		schreibeDreiklang(t, anwendung, satz, eintrag)
	}

	var alle []kartenabstand

	for vorlage, gewaehlte := range nachVorlage(karten.Karten, desSets) {
		for i := range gewaehlte {
			eintrag, ok := abstandEiner(t, anwendung, satz, gewaehlte[i], vorlage, verluste)
			if ok {
				alle = append(alle, eintrag)
			}
		}
	}

	// **Was nicht gemessen werden konnte, steht dabei** (#616) — sonst spricht
	// die Rangliste über eine Menge, die niemand kennt.
	verluste.melde(t, len(alle))

	sort.Slice(alle, func(i, j int) bool { return alle[i].abstand > alle[j].abstand })

	for i := range alle {
		if i >= WeitesteJeSatz {
			break
		}

		t.Logf("  %-10s %-30s %-28s %7.3f", satz, alle[i].name, alle[i].vorlage,
			alle[i].abstand)

		schreibeDreiklang(t, anwendung, satz, alle[i])
	}
}

// abstandEiner misst eine Karte gegen ihr Referenzbild.
//
// **Jeder Ausfall wird gezählt** (#616). Ohne das verschwand eine Karte, die
// sich nicht messen ließ, wortlos aus der Rangliste der weitesten — und gerade
// die weiteste könnte diejenige sein, die gar nicht mehr rastert. Eine
// Verschlechterung im Renderpfad nähme damit Karten aus der Messmenge und
// **verbesserte** die Zahl.
func abstandEiner(
	t *testing.T, anwendung *App, satz string, karte KarteDTO, vorlage string,
	verluste ausfaelle,
) (kartenabstand, bool) {
	t.Helper()

	name := kartenname(karte)
	if name == "" {
		verluste.nimm("die Karte hat keinen Namen")

		return kartenabstand{}, false
	}

	referenz, err := ladeBild(filepath.Join(referenzort(t, satz), referenzdatei(name)))
	if err != nil {
		verluste.nimm("kein Referenzbild", name)

		return kartenabstand{}, false
	}

	unsere, ok := unsereKarte(t, anwendung, karte, name)
	if !ok {
		verluste.nimm("die Vorschau scheiterte", name)

		return kartenabstand{}, false
	}

	ganz, err := render.Ganzabstand(unsere, referenz)
	if err != nil {
		verluste.nimm("der Abstand ließ sich nicht rechnen", name)

		return kartenabstand{}, false
	}

	return kartenabstand{
		karte: karte, name: name, vorlage: vorlage, abstand: ganz.Abstand(),
	}, true
}

// unsereKarte rastert eine Karte über die Naht der Anwendung.
func unsereKarte(
	t *testing.T, anwendung *App, karte KarteDTO, name string,
) (image.Image, bool) {
	t.Helper()

	vorschau := anwendung.Vorschau(t.Context(), karte.ID, 1)
	if vorschau.Fehler != nil {
		t.Errorf("%s: %+v", name, vorschau.Fehler)

		return nil, false
	}

	return bildAus(t, anwendung, vorschau.Vorschau.Bild), true
}

// schreibeDreiklang legt unseres, MSE2s und die Differenz nebeneinander ab.
func schreibeDreiklang(t *testing.T, anwendung *App, satz string, eintrag kartenabstand) {
	t.Helper()

	referenz, err := ladeBild(filepath.Join(referenzort(t, satz),
		referenzdatei(eintrag.name)))
	if err != nil {
		return
	}

	unsere, ok := unsereKarte(t, anwendung, eintrag.karte, eintrag.name)
	if !ok {
		return
	}

	stamm := filepath.Join(abstandsbilder(), fmt.Sprintf("%s-%s-%.0f", satz,
		saeuberlich(eintrag.name), eintrag.abstand))

	schreibeBild(t, stamm+"-unser.png", unsere)
	schreibeBild(t, stamm+"-mse2.png", referenz)
	schreibeBild(t, stamm+"-differenz.png", differenzbild(unsere, referenz))
}

// saeuberlich macht aus einem Kartennamen einen Dateinamen.
func saeuberlich(name string) string {
	sauber := make([]rune, 0, len(name))

	for _, zeichen := range name {
		switch {
		case zeichen >= 'a' && zeichen <= 'z', zeichen >= 'A' && zeichen <= 'Z',
			zeichen >= '0' && zeichen <= '9':
			sauber = append(sauber, zeichen)

		default:
			sauber = append(sauber, '-')
		}
	}

	return string(sauber)
}

// differenzbild malt, **wo** sich zwei Bilder unterscheiden.
//
// **Der Betrag je Kanal, verstärkt.** Ein Unterschied von 10 auf 255 ist am
// Bildschirm nicht zu sehen; hier wird er sichtbar gemacht, damit die Stelle
// auffällt. Die **Stärke** liest man an der Zahl im Dateinamen, nicht am Bild.
func differenzbild(unser, referenz image.Image) image.Image {
	rahmen := unser.Bounds()
	bild := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	uMin, rMin := unser.Bounds().Min, referenz.Bounds().Min

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			a, b, c, _ := unser.At(uMin.X+x, uMin.Y+y).RGBA()
			d, e, f, _ := referenz.At(rMin.X+x, rMin.Y+y).RGBA()

			bild.SetNRGBA(x, y, color.NRGBA{
				R: verstaerkt(int(a>>8) - int(d>>8)),
				G: verstaerkt(int(b>>8) - int(e>>8)),
				B: verstaerkt(int(c>>8) - int(f>>8)),
				A: 255,
			})
		}
	}

	return bild
}

// Verstaerkung ist der Faktor, mit dem ein Unterschied sichtbar gemacht wird.
const Verstaerkung = 4

func verstaerkt(unterschied int) uint8 {
	if unterschied < 0 {
		unterschied = -unterschied
	}

	return uint8(min(255, unterschied*Verstaerkung)) //nolint:gosec // geklemmt
}

// schreibeBild legt ein PNG ab.
func schreibeBild(t *testing.T, pfad string, bild image.Image) {
	t.Helper()

	datei, err := os.Create(pfad) //nolint:gosec // ein Pfad aus der Ablage
	if err != nil {
		t.Errorf("%s: %v", pfad, err)

		return
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // geschrieben ist geschrieben

	if err := png.Encode(datei, bild); err != nil {
		t.Errorf("%s: %v", pfad, err)
	}
}

// genannt sagt, ob eine Karte über `MSE_KARTEN` ausdrücklich verlangt ist.
//
// **Nach Namen und ohne Rücksicht auf Groß- und Kleinschreibung**: der Name
// kommt von Hand aus einem Issue, und ein Fehlschlag daran wäre eine stille
// Leerausgabe.
func genannt(name string) bool {
	liste := os.Getenv("MSE_KARTEN")
	if liste == "" {
		return false
	}

	for _, eintrag := range strings.Split(liste, ",") {
		if strings.EqualFold(strings.TrimSpace(eintrag), name) {
			return true
		}
	}

	return false
}
