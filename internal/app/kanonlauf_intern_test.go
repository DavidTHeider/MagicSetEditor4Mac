//go:build fremdpakete

// Der **vierte Messlauf**: die kanonischen MSE2-Sets
// (#403).
//
// # Was er misst, und was die drei anderen nicht messen
//
// | Lauf | misst | fehlt ihm |
// | --- | --- | --- |
// | `service.TestDerPlanbauerUeberFremdpakete` | 190 Stile, eine erfundene Probekarte je Stil | echte Kartendaten |
// | `app.TestDemokartenZeichnen` | sieben Karten, von uns geschrieben | fremde Kartendaten |
// | `app.TestDerAbstandZuMSE2` | Abstand zu 27 Referenzbildern | deren Vorlage ist unbekannt |
//
// Hier stehen **MSE2s eigene Sets**: bekanntes Stylesheet, bekannte
// Stiloptionen, echte Kartendaten. Was der Lauf **nicht** kann, ist der
// Vergleich mit MSE2s Bild — dafür bräuchte es dessen Ausgabe, und die liegt
// nicht vor (§10).
//
// # Der volle Lauf, und warum er nicht der gewöhnliche ist
//
// **Am 2026-08-21 einmal über alle zehn Sets und alle 25.148 Karten gefahren**
// (`MSE_STICHPROBE=alle`), Laufzeit **42 Minuten**:
//
//	10 Sets, 25148 Karten gezeichnet → 3 Befunde
//	  Portierung      0
//	  Vorlage         3
//	  Umgebung        0
//
// Die drei sind **ein** Fall, dreimal: die Symbolschrift `magic-mana-small`
// führt kein Symbol für `\n`, und es wird übersprungen — zuletzt an
// `Comet, Stellar Pup`. Das ist eine Eigenschaft der Vorlage; MSE tut dort
// dasselbe.
//
// **Gegenüber dem 2026-08-20 sind es 43 → 3.** Die 40, die weggefallen sind,
// waren ein Pfad, den eine Vorlage ausrechnet und den es nicht gibt.
//
// **Gewöhnlich läuft er über eine Stichprobe**, und das ist keine Bequemlichkeit:
// 42 Minuten in jedem Lauf hätte niemand, und die Stichprobe hat ihre eigene,
// ausgerechnete Aussage ([stichprobengroesse], `stichprobe_test.go`). Was sie
// **nicht** kann, hat #610
// gezeigt: 680 Karten zeigten null, wo 25.148 dreiundvierzig zeigten. Wer die
// Zusage `Portierung = 0` über **alle** Karten braucht, fährt den vollen Lauf —
// und schreibt sein Ergebnis hierher.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// kanonverzeichnis ist der Ort der kanonischen Sets.
//
// **Nicht im Repositorium** (§11): sechs der zehn wiegen je über 227 MB.
//
// # `MSE_KANON` geht vor `HOME`
//
// **Weil `TestMain` dem ganzen Lauf ein eigenes Zuhause gibt** — ein leeres
// Verzeichnis unter `/tmp`, damit kein Test den Stand dieses Rechners sieht
// (`echteszuhause_test.go`, seit dem 2026-08-26). Wer den Kanon aus `HOME`
// baute, suchte ihn seither dort und fände nichts.
//
// **Was das gekostet hat**: `TestDerAbstandZuMSE2AmKanon` und
// `TestDerAbstandZuMSE2` sprangen ab, und ein Sprung sieht aus wie ein
// grüner Lauf. Beide stehen auf der Freigabeliste in `docs/RELEASE.md` —
// gefahren wurden sie zwei Freigaben lang nicht. Gefunden beim Vorbereiten
// von 0.3.0.
//
// **Dasselbe Muster wie in [kanonsatzpfad]**, und dort steht es seit jenem
// Tag. Es war nur nie hierher gezogen worden.
func kanonverzeichnis() string {
	if ort := os.Getenv("MSE_KANON"); ort != "" {
		return ort
	}

	return filepath.Join(os.Getenv("HOME"), "Library", "Application Support",
		"MagicSetEditor4Mac", "kanonische-sets")
}

// KartenJeSet ist die Stichprobe, die dieser Lauf je Set zeichnet — der
// **Rückfall**, wenn niemand [stichprobengroesse] gefragt hat.
//
// **Die Zahl steht in der Ausgabe**, damit die Reichweite mitläuft; woher sie
// kommt und was sie belegt, steht in `stichprobe_test.go`.
const KartenJeSet = 25

// **Die kanonischen Sets zeichnen — und was dabei herauskommt.**
//
// Er misst über die **laufende Anwendung**: öffnen, Karten holen, Vorschau.
// Nicht am Planbauer vorbei — eine Vorrichtung, die ärmer ist als der Betrieb,
// prüft den falschen Weg (§3).
func TestDieKanonischenSetsZeichnen(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := os.ReadDir(kanonverzeichnis())
	if err != nil {
		t.Skip("die kanonischen Sets liegen nicht vor")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	gruende := kanongruende{}

	var gesamtkarten, gesamtbefunde, gesetzteSets int

	// **Die Stichprobe hat eine Aussage, nicht nur eine Zahl** — der Grund
	// steht in `stichprobe_test.go`
	// (#499).
	jeSet := stichprobengroesse(t, zaehleSets(eintraege))

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".mse-set") {
			continue
		}

		karten, befunde := zeichneKanonset(t, anwendung, eintrag.Name(), gruende, jeSet)

		gesetzteSets++
		gesamtkarten += karten
		gesamtbefunde += befunde
	}

	t.Logf("%d Sets, %d Karten gezeichnet (Stichprobe %d je Set) → %d Befunde",
		gesetzteSets, gesamtkarten, jeSet, gesamtbefunde)

	nachArt := zaehleNachArt(gruende)
	for _, art := range []render.Befundart{
		render.ArtPortierung, render.ArtVorlage, render.ArtUmgebung,
	} {
		t.Logf("  %-12s %4d", art, nachArt[art])

		// **Und als Kennzahl für den Beleg** — QA-01 (#1733): die Zeile
		// darüber ist für Menschen, diese für `internal/laufbeleg`.
		t.Logf("kennzahl kanon.%s=%d", strings.ToLower(string(art)), nachArt[art])
	}

	nenneSpeicher(t, anwendung)

	t.Logf("kennzahl kanon.sets=%d", gesetzteSets)
	t.Logf("kennzahl kanon.karten=%d", gesamtkarten)
	t.Logf("kennzahl kanon.befunde=%d", gesamtbefunde)

	// **Was die Zahl belegt, steht daneben.** Eine Stichprobe ohne ihre
	// Aussage ist eine Zahl, die niemand einordnen kann (§10).
	nenneReichweite(t, gesamtkarten, gesamtbefunde)

	nenneGruende(t, gruende)

	if gesetzteSets == 0 {
		t.Fatal("kein einziges Set gezeichnet")
	}

	// **Portierung steht auf null**, und das ist die Zusage: was dieses
	// Programm nicht kann, ist gezählt und soll keines sein.
	if nachArt[render.ArtPortierung] > 0 {
		t.Errorf("%d Befunde der Art Portierung — erwartet keine",
			nachArt[render.ArtPortierung])
	}

	// **Der Deckel und seine Geschichte stehen in `kanondeckel_intern_test.go`**
	// (§2: diese Datei ginge sonst über 400 Zeilen). Dort steht auch, was er
	// **nicht** fängt.
	pruefeRate(t, "Vorlage", nachArt[render.ArtVorlage], gesamtkarten, RateVorlage)
	pruefeRate(t, "Umgebung", nachArt[render.ArtUmgebung], gesamtkarten, RateUmgebung)
}

// zeichneKanonset öffnet ein Set und zeichnet seine Stichprobe.
func zeichneKanonset(
	t *testing.T, anwendung *App, name string, gruende kanongruende, jeSet int,
) (karten, befunde int) {
	t.Helper()

	geoeffnet := anwendung.SetOeffnen(t.Context(), filepath.Join(kanonverzeichnis(), name))
	if geoeffnet.Fehler != nil {
		t.Errorf("%s öffnen: %+v", name, geoeffnet.Fehler)

		return 0, 0
	}

	liste := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if liste.Fehler != nil {
		t.Errorf("%s: Karten: %+v", name, liste.Fehler)

		return 0, 0
	}

	// **Geschichtet nach Stilangaben, nicht jede n-te**
	// (#610).
	// Warum, steht in `schichten_test.go`.
	schichten := stilschichten(t, filepath.Join(kanonverzeichnis(), name))
	stichprobe := geschichteteStichprobe(liste.Karten, schichten, jeSet)

	for i := range stichprobe {
		vorschau := anwendung.Vorschau(t.Context(), stichprobe[i].ID, 1)
		if vorschau.Fehler != nil {
			t.Errorf("%s, Karte %s: %+v", name, stichprobe[i].ID, vorschau.Fehler)

			continue
		}

		karten++

		for _, meldung := range vorschau.Vorschau.Befunde {
			befunde++

			gruende.zaehle(kanonklasse(meldung), satzname(name),
				kartenname(stichprobe[i]))
		}
	}

	t.Logf("  %-24s %4d von %5d Karten aus %3d Schichten → %4d Befunde",
		name, karten, len(liste.Karten), zaehleSchichten(schichten), befunde)

	return karten, befunde
}

// kanonklasse fasst eine Meldung auf ihre Klasse zusammen.
//
// # Die Klasse ist der Grund, nicht der Ort
//
// Eine Meldung trägt vorn ihren Ort — `text`, `Ebene 15`, `arrow 7 →
// senkrecht → kante` — und dahinter den Grund. Bis
// #427 schnitt
// diese Funktion am **ersten** `": "` und behielt damit genau den Ort. Die
// Ausgabe des Laufs las sich als „30 × text, 25 × type, 24 × card color", und
// was an diesen Feldern falsch ist, stand nirgends; die Meldungen mussten über
// einen eigens geschriebenen Abgriff geholt werden.
//
// # Und die Einzelheit muss trotzdem weg
//
// Ein Feldname oder ein Pfad im **Grund** macht aus einer Klasse hundert.
// Deshalb wird nach dem Ort noch einmal gekürzt — an der Stelle, an der die
// Einzelheit anfängt. Was zusammenfallen und was auseinanderbleiben muss,
// steht in `kanonklasse_intern_test.go`; ohne diese Gegenprobe wäre eine
// Funktion, die alles auf „Befund" abbildet, ebenso grün.
func kanonklasse(meldung string) string {
	meldung = ohneOrt(meldung)

	// **Eine umschlossene Meldung wird auf ihren Kern gebracht — so oft, wie
	// sie umschlossen ist.** Eine fehlende Maske nennt das Bild, das Bild
	// nennt die Datei, und die Datei nennt den Grund; ein einziger Schnitt
	// bliebe bei „das Bild … ist nicht lesbar" stehen.
	for {
		_, kern, ist := strings.Cut(meldung, " ist nicht lesbar: ")
		if !ist {
			break
		}

		meldung = ohneOrt(kern)
	}

	// **Ein Pfad ganz vorn ist ein Ort, kein Grund.** Der innerste Kern lautet
	// „/a.mse-style/bild.png: das Bild ist in keinem Paket zu finden"; ohne
	// diese Zeile bliebe der Pfad stehen, und jede Datei wäre eine Klasse.
	//
	// **Am Anfang und nicht irgendwo**: „in der Schrift X fehlen Zeichen: ǵ"
	// beginnt mit einem Satz, und dessen Rest ist die Einzelheit, nicht der
	// Grund.
	if strings.HasPrefix(meldung, "/") {
		if _, kern, ist := strings.Cut(meldung, ": "); ist {
			meldung = kern
		}
	}

	// **Und die Einzelheit im Kern weicht einem Auslassungszeichen.** Der
	// Nachsatz dahinter fällt mit — „gezeichnet wird der bloße Text" steht an
	// jeder Meldung dieser Klasse.
	for _, muster := range klassenmuster {
		if vor, nach, ist := strings.Cut(meldung, muster.beginn); ist {
			if _, _, hat := strings.Cut(nach, muster.ende); hat {
				return vor + muster.beginn + "…" + muster.ende
			}
		}
	}

	if stelle := strings.Index(meldung, ": "); stelle > 0 {
		meldung = meldung[:stelle]
	}

	if len(meldung) > 90 {
		meldung = meldung[:90]
	}

	return meldung
}

// klassenmuster ersetzt die Einzelheit **innerhalb** eines Grundes.
//
// **Nur dort, wo das Abschneiden am `": "` die Klasse zerschlüge.** „das Bild
// /a/b.png ist nicht lesbar" endete sonst bei „das Bild /a/b.png ist nicht
// lesbar" — je Datei eine Klasse.
var klassenmuster = []struct{ beginn, ende string }{
	{beginn: "die Auszeichnungen ", ende: " sind entfernt"},
}

// ohneOrt streift den Ort vor dem Grund ab.
//
// **Am ersten `": "`, aber nur, wenn davor keiner steht.** Ein Ort ist kurz
// und nennt ein Feld oder eine Ebene; ein Grund enthält Satzzeichen, Zahlen
// und Pfade. Die Grenze ist die **Länge** — ein Ort über sechzig Zeichen kommt
// im Pack nicht vor, ein Grund darunter schon.
func ohneOrt(meldung string) string {
	vor, nach, ist := strings.Cut(meldung, ": ")
	if !ist || len(vor) > 60 {
		return meldung
	}

	// **Die Zeilenangabe des Zerteilers zählt zum Ort.** „Zeile 798, Spalte 35"
	// steht zwischen Ort und Grund und ist die Einzelheit selbst.
	for _, weiter := range []string{
		"die Angabe ließ sich nicht ausrechnen: ",
		"der Vorgabewert ließ sich nicht ausrechnen: ", "ausrechnen: ",
	} {
		if _, rest, hat := strings.Cut(nach, weiter); hat {
			nach = rest

			break
		}
	}

	if rest, hat := strings.CutPrefix(nach, "Zeile "); hat {
		if _, dahinter, gefunden := strings.Cut(rest, ": "); gefunden {
			return dahinter
		}
	}

	return nach
}
