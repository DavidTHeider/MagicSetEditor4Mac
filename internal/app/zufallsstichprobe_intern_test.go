//go:build fremdpakete

// **Die Zufallsstichprobe gegen MSE2** — QA-02
// (#1734).
//
// # Zwei Kohorten, zwei Aussagen
//
// **Die feste Kohorte** (`mse2kanon_intern_test.go`) misst in jedem Lauf
// dieselben, systematisch verteilten Karten — sie sagt, ob sich die Zahl
// **bewegt** hat. **Diese hier** misst eine einmal zufällig gezogene
// Stichprobe aus `docs/belege/kanonstichprobe.json` — sie trägt, was eine
// systematische Auswahl nicht trägt: eine **Entdeckungszusage** und ein
// **Vertrauensband**, beides unter der Annahme, die ihre Ziehung erfüllt.
// Berichtet werden beide getrennt (AK-04), unter `mse2.*` und `zufall.*`.
//
// # Was sie zusagt
//
// **Je Satz**: liefert kein gezogener Satz einen Befund der Art
// `Portierung`, dann liefern mit mindestens 99,9 Prozent Sicherheit weniger
// als ein Prozent **seiner** Karten einen — exakt hypergeometrisch gerechnet
// (AK-02). **Über alle zehn zugleich** nur mit dem Produkt: rund 99,0 Prozent
// (AK-03). Findet sie Befunde, gilt die Schranke nicht, und der Lauf ist rot.
//
// # Was sie nicht zusagt
//
// **Dass eine Karte richtig ist.** Eine Karte kann still falsch sein und
// keinen Befund liefern — dafür misst der Abstand daneben, und der ist keine
// Ja-Nein-Frage.
package app

import (
	"crypto/rand"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// planort ist der Ort des Plans im Baum.
var planort = filepath.Join("..", "..", "docs", "belege", "kanonstichprobe.json")

// **Die Vorrichtung, die den Plan einmal zieht** — nur mit
// `MSE_STICHPROBENPLAN=<Ziel>`. Ein neuer Plan ist eine **neue Kohorte**:
// wer ihn zieht, ersetzt den alten mit Absicht und nicht nebenbei.
func TestZieheDenStichprobenplan(t *testing.T) {
	ziel := os.Getenv("MSE_STICHPROBENPLAN")
	if ziel == "" {
		t.Skip("eine Vorrichtung: zieht den Plan nur mit MSE_STICHPROBENPLAN")
	}

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	anwendung := gestartetMit(t, func() string { return wurzel })

	schichten := make([]planschicht, 0, len(kanonsaetze))

	for _, satz := range kanonsaetze {
		karten := kartenDesSatzes(t, anwendung, satz)
		schichten = append(schichten, planschicht{
			Satz: satz, Karten: len(karten), Kartenhash: kartenhash(namenDer(karten)),
		})
	}

	// **Die Saat kommt aus dem Zufall des Systems** und steht danach offen im
	// Plan — wer ihn nachrechnen will, braucht nichts sonst.
	var roh [16]byte
	if _, fehler := rand.Read(roh[:]); fehler != nil {
		t.Fatal(fehler)
	}

	plan := ziehePlan([2]uint64{binary.BigEndian.Uint64(roh[:8]), binary.BigEndian.Uint64(roh[8:])},
		schichten)

	datei, fehler := os.Create(ziel) //nolint:gosec // das Ziel nennt, wer die Vorrichtung startet
	if fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := plan.schreiben(datei); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := datei.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	t.Logf("gezogen: %d Sätze, gemeinsame Sicherheit %.4f", len(plan.Schichten), plan.Gemeinsam)
}

// **Die Messung über den Plan** — Befunde und Abstand in einem Durchgang.
func TestDieZufallsstichprobe(t *testing.T) {
	datei, fehler := os.Open(planort) //nolint:gosec // ein fester Ort im Baum
	if fehler != nil {
		t.Fatalf("ohne Plan keine Zufallsstichprobe — MSE_STICHPROBENPLAN=… zieht ihn: %v", fehler)
	}

	plan, fehler := liesPlan(datei)
	_ = datei.Close() //nolint:errcheck // nur gelesen

	if fehler != nil {
		t.Fatal(fehler)
	}

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	anwendung := gestartetMit(t, func() string { return wurzel })
	gruende := kanongruende{}
	stand := neuerMessstand()

	schichten := make([]schicht, 0, len(plan.Schichten))
	gezogen, sicherste := 0, 1.0

	for _, geplant := range plan.Schichten {
		abstaende := messeSchicht(t, anwendung, geplant, gruende, stand)
		schichten = append(schichten, schicht{name: geplant.Satz, ganz: geplant.Karten, werte: abstaende})
		gezogen += len(geplant.Stichprobe)
		sicherste = min(sicherste, geplant.Findewahrscheinlichkeit)
	}

	nachArt := zaehleNachArt(gruende)

	t.Logf("Zufallsstichprobe: %d Karten aus %d Sätzen, Saat %x-%x",
		gezogen, len(plan.Schichten), plan.Saat[0], plan.Saat[1])
	t.Logf("  je Satz mindestens %.4f, gemeinsam %.4f Sicherheit", sicherste, plan.Gemeinsam)
	t.Logf("kennzahl zufall.karten=%d", gezogen)
	t.Logf("kennzahl zufall.sicherheit.je=%.4f", sicherste)
	t.Logf("kennzahl zufall.sicherheit.gemeinsam=%.4f", plan.Gemeinsam)

	for _, art := range []render.Befundart{render.ArtPortierung, render.ArtVorlage, render.ArtUmgebung} {
		t.Logf("  %-12s %4d", art, nachArt[art])
		t.Logf("kennzahl zufall.%s=%d", kleingeschrieben(art), nachArt[art])
	}

	meldeKanonAls(t, "zufall", zufallsKohorte, schichten)

	// **Findet sie Befunde, gilt die Schranke nicht** — dann steht die Zusage
	// „Portierung = 0" zur Prüfung, und nicht die Rechnung.
	if nachArt[render.ArtPortierung] > 0 {
		t.Errorf("%d Befunde der Art Portierung in der Zufallsstichprobe — die "+
			"Entdeckungszusage setzt null voraus", nachArt[render.ArtPortierung])
	}
}
