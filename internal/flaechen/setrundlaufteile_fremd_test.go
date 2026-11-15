//go:build fremdpakete

package flaechen_test

// **Die Teile des Setdatei-Rundlaufs** — F05
// (#2247):
// Umgebung, eigene Saat, Vergleiche und Bericht. Der Lauf steht in
// `setrundlauf_fremd_test.go`, die Maße in `setrundlaufmass_fremd_test.go`.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// rundlaufumgebung baut eine Umgebung mit eigenem Zuhause, in dessen
// Paketordner eine **Kopie** des Packs liegt.
//
// **Geklont, nicht verlinkt**: `cp -c` legt auf APFS eine Kopie an, die
// keinen Platz kostet, bis jemand schreibt — und schreibt jemand, trifft es
// die Kopie. Eine Verknüpfung führte jeden Schreibfehler ins echte Pack (§3).
func rundlaufumgebung(t *testing.T, pack string) []string {
	t.Helper()

	umgebung := eigenerBestand(t)
	zuhause := ""

	for _, eintrag := range umgebung {
		if wert, ist := strings.CutPrefix(eintrag, "HOME="); ist {
			zuhause = wert
		}
	}

	ort := filepath.Join(zuhause, "Library", "Application Support", "MagicSetEditor4Mac", "pakete")
	if fehler := os.MkdirAll(ort, 0o700); fehler != nil {
		t.Fatalf("Paketordner: %v", fehler)
	}

	// #nosec G204 G702 -- `cp` mit zwei Pfaden: dem Pack aus der Umgebung des
	// Laufs und einem Ordner unter `t.TempDir()`.
	if aus, fehler := exec.CommandContext(t.Context(), "cp", "-cR", pack, filepath.Join(ort, "fremd")).CombinedOutput(); fehler != nil {
		t.Fatalf("das Pack klonen: %v (%s)", fehler, aus)
	}

	return umgebung
}

// mitEigenemBestand gibt derselben Umgebung einen frischen Bestand — bei
// doppeltem Schlüssel gilt der letzte.
//
// **Je Richtung einer**: ein zweites `oeffnen` derselben Quelle im selben
// Bestand gäbe die offene Kennung zurück, ohne die Datei zu lesen
// ([istWiederLesbaresSet]).
func mitEigenemBestand(umgebung []string, arbeit, name string) []string {
	return append(slices.Clone(umgebung), "MSE_BESTAND="+filepath.Join(arbeit, name+".db"))
}

// feldliste nennt jedes Feld, nach dem das Orakel gefragt wird: die
// Kartenfelder des Spiels und jeden Schlüssel, den eine Karte trägt.
func feldliste(spiel map[string]bool, werte []map[string]string) []string {
	felder := make(map[string]bool)

	for feld, bild := range spiel {
		if !bild {
			felder[feld] = true
		}
	}

	for _, karte := range werte {
		for feld := range karte {
			if bild, bekannt := spiel[feld]; !bekannt || !bild {
				felder[feld] = true
			}
		}
	}

	liste := make([]string, 0, len(felder))
	for feld := range felder {
		liste = append(liste, feld)
	}

	slices.Sort(liste)

	return liste
}

// vergleiche hält zwei Auszüge des Orakels gegeneinander — bis auf eine
// gewollte Änderung.
func vergleiche(vorher, nachher []map[string]string, gewollt func(karte int, feld string) (string, bool)) []string {
	if len(vorher) != len(nachher) {
		return []string{fmt.Sprintf("%d Karten statt %d", len(nachher), len(vorher))}
	}

	var abweichungen []string

	for i, karte := range vorher {
		for _, feld := range sortierteSchluessel(karte) {
			soll := karte[feld]
			if wert, ist := gewollt(i, feld); ist {
				soll = wert
			}

			if ist := nachher[i][feld]; ist != soll {
				abweichungen = append(abweichungen, fmt.Sprintf("Karte %d, %s: %s", i, feld, unterschied(soll, ist)))
			}
		}
	}

	return abweichungen
}

// vergleicheWerte hält, was diese Anwendung aus **MSE2s Abschrift** liest,
// gegen das, was MSE2 selbst aus ihr liest — und nennt daneben, wo MSE2 den
// gespeicherten Wert gegenüber dem Original umgeschrieben hat.
//
// # Warum nicht die eigene Lesung des Originals
//
// **MSE2 speichert den Wert nach dem Feldskript** (`TextValue` rechnet
// `script:` bei jeder Änderung und legt das Ergebnis ab): aus „„…“" wird
// „„…”", und ein unbekanntes Wort bekommt seine Unterringelung. Diese
// Anwendung speichert, was in der Datei steht, und rechnet das Skript beim
// Zeichnen. Eine Abweichung zwischen der eigenen Lesung von Original und
// Abschrift ist deshalb oft MSE2s Umschreibung — sie steht unter `fremd`.
// **Gegen diese Anwendung zählt**, wenn sie dieselbe Datei anders liest als
// MSE2.
//
// **Die Rechtschreibmarken werden beiderseits herausgenommen**: diese
// Anwendung liest sie mit Absicht nicht
// (`internal/service/import_rechtschreibmarken_test.go`).
func vergleicheWerte(vorher, nachher, orakelNachher []map[string]string) (eigen, fremd []string) {
	if len(vorher) != len(nachher) || len(nachher) != len(orakelNachher) {
		return []string{fmt.Sprintf("%d Karten, MSE2 liest %d, im Original %d", len(nachher), len(orakelNachher), len(vorher))}, nil
	}

	for i, karte := range nachher {
		for _, feld := range sortierteSchluessel(karte) {
			if karte[feld] != vorher[i][feld] {
				fremd = append(fremd, fmt.Sprintf("Karte %d, %s: %s", i, feld, unterschied(vorher[i][feld], karte[feld])))
			}

			mse2, gelesen := orakelNachher[i][strings.ReplaceAll(feld, " ", "_")]
			if !gelesen || mse2 == unlesbar {
				continue
			}

			if ohneMarken(mse2) != ohneMarken(karte[feld]) {
				eigen = append(eigen, fmt.Sprintf("Karte %d, %s: %s", i, feld, unterschied(ohneMarken(mse2), ohneMarken(karte[feld]))))
			}
		}
	}

	return eigen, fremd
}

var rechtschreibmarke = regexp.MustCompile(`</?error-spelling(?::[^<>]*)?>`)

func ohneMarken(text string) string { return rechtschreibmarke.ReplaceAllString(text, "") }

// unterschied zeigt zwei Texte ab der ersten Stelle, an der sie auseinandergehen.
func unterschied(soll, ist string) string {
	a, b := []rune(soll), []rune(ist)
	stelle := 0

	for stelle < len(a) && stelle < len(b) && a[stelle] == b[stelle] {
		stelle++
	}

	ab := max(stelle-10, 0)

	return fmt.Sprintf("ab Zeichen %d %q → %q", stelle, string(a[ab:min(len(a), stelle+30)]), string(b[ab:min(len(b), stelle+30)]))
}

func sortierteSchluessel(werte map[string]string) []string {
	schluessel := make([]string, 0, len(werte))
	for name := range werte {
		schluessel = append(schluessel, name)
	}

	slices.Sort(schluessel)

	return schluessel
}

// pruefeRundlauf macht rot, was **diese** Anwendung verliert.
func pruefeRundlauf(t *testing.T, mass rundlaufmass) {
	t.Helper()

	melde := func(was string, liste []string) {
		if len(liste) > 0 {
			t.Errorf("%s — %d: %s", was, len(liste), strings.Join(liste[:min(len(liste), 12)], " · "))
		}
	}

	melde("MSE2 liest die Abschrift dieser Anwendung anders als das Original", mass.abweichungEins)
	melde("diese Anwendung liest MSE2s Abschrift anders als das Original", mass.abweichungZwei)
	melde("beim Schreiben verlorene Schlüssel", mass.verlorenEins)
	melde("beim Schreiben verlorene Dateien", mass.dateienFehlenEins)
	melde("beim Schreiben veränderte Dateien", mass.dateienAndersEins)
	melde("MSE2 liest die setweite Stiloption der Abschrift nicht", mass.stilEins)
	melde("MSE2 warnt über die Abschrift, was es über das Original nicht tat", mass.warnungenEins)

	t.Logf("MSE2 selbst: %d Schlüssel und %d Dateien nicht mitgeschrieben, %d Werte geändert",
		len(mass.verlorenZweiMSE2), len(mass.dateienFehlenZweiMSE2), len(mass.fremdeAbweichungZweiMSE))
}

// berichtSchreiben gibt die Zahlen als Tafel aus — und legt sie ab, wenn
// `MSE_SETRUNDLAUF_BERICHT` einen Pfad nennt.
func berichtSchreiben(t *testing.T, bericht []rundlaufmass) {
	t.Helper()

	var tafel strings.Builder

	tafel.WriteString("| Set | Karten | Werte | davon Unicode | gerechnet (MSE2) | Dateien | Stiloption | 1: Abweichungen | 1: Schlüssel weg | 1: Dateien weg/anders | 1: neue Warnungen | 2: Abweichungen | MSE2: Schlüssel weg | MSE2: Dateien weg | MSE2: Werte geändert |\n")
	tafel.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")

	for _, m := range bericht {
		fmt.Fprintf(&tafel, "| %s | %d | %d | %d | %d | %d | %s | %d | %d | %d/%d | %d | %d | %d | %d | %d |\n",
			m.name, m.karten, m.werte, m.unicode, m.gerechnet, m.dateien, stilspalte(m),
			len(m.abweichungEins), len(m.verlorenEins), len(m.dateienFehlenEins), len(m.dateienAndersEins),
			len(m.warnungenEins), len(m.abweichungZwei), len(m.verlorenZweiMSE2), len(m.dateienFehlenZweiMSE2),
			len(m.fremdeAbweichungZweiMSE))
	}

	for _, m := range bericht {
		if len(m.verlorenZweiMSE2)+len(m.fremdeAbweichungZweiMSE) > 0 {
			fmt.Fprintf(&tafel, "\n**%s, was MSE2 nicht mitschreibt**: %s\n", m.name,
				strings.Join(slices.Concat(m.verlorenZweiMSE2[:min(len(m.verlorenZweiMSE2), 8)],
					m.fremdeAbweichungZweiMSE[:min(len(m.fremdeAbweichungZweiMSE), 8)]), " · "))
		}
	}

	t.Log("\n" + tafel.String())

	if pfad := os.Getenv("MSE_SETRUNDLAUF_BERICHT"); pfad != "" {
		// #nosec G703 -- der Pfad kommt aus der Umgebung dessen, der misst.
		if fehler := os.WriteFile(pfad, []byte(tafel.String()), 0o600); fehler != nil {
			t.Errorf("den Bericht ablegen: %v", fehler)
		}
	}
}

// stilspalte nennt die gesetzte Stiloption und ob MSE2 sie las.
func stilspalte(m rundlaufmass) string {
	switch {
	case m.stiloption == "":
		return "keine im Stil"
	case len(m.stilEins) > 0:
		return "`" + m.stiloption + "` nicht gelesen"
	default:
		return "`" + m.stiloption + "` gelesen"
	}
}
