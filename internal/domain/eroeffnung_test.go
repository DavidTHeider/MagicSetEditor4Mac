package domain_test

// **Der Beginn einer Partie** — Mischen, Starthand und Mulligan, CR 103.3
// und CR 103.5 (Fassung 2026-09-25),
// #1839.
//
// # Wogegen hier geprüft wird
//
// **Gegen den Tisch, der am 2026-09-24 mit dem Set „Green" gesehen wurde**:
// jede Bibliothek lag nach dem Anfangen in der Reihenfolge der Deckliste, und
// niemand hatte eine Hand. Wer spielen wollte, mischte von Hand und schob
// sieben Karten einzeln hinüber — und einen Mulligan gab es nicht.
//
// **Die Mischung wird eingesetzt**, wie beim Mischen einer Zone: `mischung[i]
// = j` heißt, an Stelle i+1 liegt danach, was vorher an Stelle j+1 lag.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitBibliotheken gibt eine Partie mit `sitze` Sitzen und je `karten`
// Karten in der Bibliothek — oben liegt `…/b01`.
func mitBibliotheken(sitze, karten int) domain.Partie {
	partie := einePartie()

	partie.Sitze = nil
	partie.Objekte = nil

	for sitz := 1; sitz <= sitze; sitz++ {
		kennung := fmt.Sprintf("sitz-%d", sitz)

		partie.Sitze = append(partie.Sitze, domain.Sitz{
			Kennung: kennung, Name: kennung, Deck: einDeckabzug(), Leben: 20,
			Eroeffnung: domain.EroeffnungAusstehend,
		})

		for stelle := 1; stelle <= karten; stelle++ {
			partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
				Kennung:     fmt.Sprintf("%s/b%02d", kennung, stelle),
				Art:         domain.ObjektKarte,
				Beherrscher: kennung, Eigner: kennung,
				Zone: domain.ZoneBibliothek, Karte: "k1", Ordnung: stelle,
			})
		}
	}

	return partie
}

// umgekehrt ist die Mischung, die eine Zone umdreht.
func umgekehrt(anzahl int) []int {
	reihe := make([]int, anzahl)

	for stelle := range reihe {
		reihe[stelle] = anzahl - 1 - stelle
	}

	return reihe
}

// inZone gibt die Kennungen eines Sitzes in einer Zone, von oben.
func inZone(partie domain.Partie, wer string, zone domain.Zone) []string {
	darin := []domain.Spielobjekt{}

	for _, objekt := range partie.Objekte {
		if objekt.Eigner == wer && objekt.Zone == zone {
			darin = append(darin, objekt)
		}
	}

	for aussen := range darin {
		for innen := aussen + 1; innen < len(darin); innen++ {
			if darin[innen].Ordnung < darin[aussen].Ordnung {
				darin[aussen], darin[innen] = darin[innen], darin[aussen]
			}
		}
	}

	kennungen := make([]string, len(darin))
	for stelle, objekt := range darin {
		kennungen[stelle] = objekt.Kennung
	}

	return kennungen
}

// sitzVon gibt den Sitz zu einer Kennung.
func sitzVon(t *testing.T, partie domain.Partie, wer string) domain.Sitz {
	t.Helper()

	sitz := partie.SitzVon(wer)
	if sitz == nil {
		t.Fatalf("kein Sitz %q", wer)
	}

	return *sitz
}

// **CR 103.3, dann CR 103.5**: erst gemischt, dann sieben gezogen.
func TestDieStarthandMischtUndZiehtSieben(t *testing.T) {
	t.Parallel()

	vorher := mitBibliotheken(2, 10)

	// **Wer vorher hineingesehen hat, schuldet ein Mischen** — und die
	// Starthand mischt.
	vorher.MerkeMischpflicht("sitz-1", domain.ZoneBibliothek)

	danach, fehler := vorher.ZieheStarthand("sitz-1", umgekehrt(10))
	if fehler != nil {
		t.Fatalf("Starthand: %v", fehler)
	}

	if danach.SchuldetMischen("sitz-1", domain.ZoneBibliothek) {
		t.Error("die Mischpflicht steht nach der Starthand noch")
	}

	if !vorher.SchuldetMischen("sitz-1", domain.ZoneBibliothek) {
		t.Error("die Mischpflicht der Eingabe ist verschwunden")
	}

	hand := inZone(danach, "sitz-1", domain.ZoneHand)
	if len(hand) != domain.Starthand {
		t.Fatalf("%d Karten auf der Hand, erwartet %d", len(hand), domain.Starthand)
	}

	// **Umgedreht und dann von oben gezogen**: oben lag danach b10.
	bibliothek := inZone(danach, "sitz-1", domain.ZoneBibliothek)
	if fmt.Sprint(bibliothek) != "[sitz-1/b03 sitz-1/b02 sitz-1/b01]" {
		t.Errorf("Bibliothek danach: %v", bibliothek)
	}

	if sitzVon(t, danach, "sitz-1").Eroeffnung != domain.EroeffnungOffen {
		t.Errorf("Stand danach: %q", sitzVon(t, danach, "sitz-1").Eroeffnung)
	}

	// **Der andere Sitz bleibt, wie er war** — und die Eingabe auch.
	if len(inZone(danach, "sitz-2", domain.ZoneHand)) != 0 ||
		sitzVon(t, danach, "sitz-2").Eroeffnung != domain.EroeffnungAusstehend {
		t.Error("der andere Sitz hat etwas vom Beginn abbekommen")
	}

	if len(inZone(vorher, "sitz-1", domain.ZoneHand)) != 0 {
		t.Error("die Eingabe wurde verändert")
	}
}

// **Eine zu kleine Bibliothek gibt, was sie hat — und der Rest ist ein
// Griff ins Leere.** CR 103.5 lässt die Starthand *ziehen*, eine Karte nach
// der anderen (CR 121.2); wer mehr ziehen muss, als daliegt, zieht den Rest
// und verliert, sobald jemand Priorität bekäme (CR 104.3c, CR 704.5b, alle
// Fassung 2026-09-25).
//
// **Hier stand bis zum Nachlesen das Gegenteil** — „kein Ziehen aus dem
// Leeren, sondern eine kurze Starthand", erschlossen aus dem Übungsdeck mit
// fünf Karten. Die Regel sagt es anders, und sie gilt.
func TestEineKleineBibliothekZiehtInsLeere(t *testing.T) {
	t.Parallel()

	danach, fehler := mitBibliotheken(2, 3).ZieheStarthand("sitz-1", umgekehrt(3))
	if fehler != nil {
		t.Fatalf("Starthand: %v", fehler)
	}

	if len(inZone(danach, "sitz-1", domain.ZoneHand)) != 3 {
		t.Error("nicht die ganze kleine Bibliothek gezogen")
	}

	if !sitzVon(t, danach, "sitz-1").Leergezogen {
		t.Error("der Griff ins Leere wurde nicht vermerkt (CR 104.3c)")
	}
}

// **Mehr, als auf der Hand liegt, geht nicht unter die Bibliothek** —
// CR 103.5 sagt dazu nichts, CR 101.3 (Fassung 2026-09-25) schon: was sich
// nicht ausführen lässt, entfällt. Ohne diese Grenze säße ein Sitz mit
// kleiner Bibliothek fest: er schuldete zwei Karten und hielte eine.
func TestEsWirdNichtMehrUntergelegtAlsAufDerHandLiegt(t *testing.T) {
	t.Parallel()

	partie, _ := mitBibliotheken(2, 1).ZieheStarthand("sitz-1", umgekehrt(1))         //nolint:errcheck // der Lauf darüber hält es
	partie, _ = partie.NimmMulligan("sitz-1", umgekehrt(1))                           //nolint:errcheck // dito
	partie, _ = partie.Unterlege("sitz-1", inZone(partie, "sitz-1", domain.ZoneHand)) //nolint:errcheck // dito

	danach, fehler := partie.NimmMulligan("sitz-1", umgekehrt(1))
	if fehler != nil {
		t.Fatalf("zweiter Mulligan: %v", fehler)
	}

	if geschuldet := sitzVon(t, danach, "sitz-1").Unterlegen; geschuldet != 1 {
		t.Errorf("geschuldet %d Karten bei einer auf der Hand", geschuldet)
	}
}

// **Ein Mulligan mischt die Hand zurück, zieht neu und verlangt, eine Karte
// unterzulegen** — im Spiel zu zweit zählt schon der erste (CR 103.5).
func TestEinMulliganZiehtNeuUndVerlangtUnterlegen(t *testing.T) {
	t.Parallel()

	partie, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // dito

	danach, fehler := partie.NimmMulligan("sitz-1", umgekehrt(10))
	if fehler != nil {
		t.Fatalf("Mulligan: %v", fehler)
	}

	if len(inZone(danach, "sitz-1", domain.ZoneHand)) != domain.Starthand {
		t.Error("nach dem Mulligan keine volle Hand")
	}

	sitz := sitzVon(t, danach, "sitz-1")
	if sitz.Mulligans != 1 || sitz.Unterlegen != 1 {
		t.Errorf("Mulligans %d, unterzulegen %d — erwartet 1 und 1",
			sitz.Mulligans, sitz.Unterlegen)
	}

	// **Solange unterzulegen ist, gibt es weder Behalten noch Mulligan.**
	if _, fehler := danach.Behalte("sitz-1"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Behalten vor dem Unterlegen: %v", fehler)
	}

	if _, fehler := danach.NimmMulligan("sitz-1", umgekehrt(10)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Mulligan vor dem Unterlegen: %v", fehler)
	}
}

// **Unterlegen legt die genannten Karten unter die Bibliothek**, in der
// genannten Reihenfolge — CR 103.5: „in beliebiger Reihenfolge" heißt, der
// Spieler bestimmt sie.
func TestUnterlegenLegtDieGenanntenUnterDieBibliothek(t *testing.T) {
	t.Parallel()

	partie, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // dito
	partie, _ = partie.NimmMulligan("sitz-1", umgekehrt(10))                    //nolint:errcheck // dito

	karte := inZone(partie, "sitz-1", domain.ZoneHand)[0]

	danach, fehler := partie.Unterlege("sitz-1", []string{karte})
	if fehler != nil {
		t.Fatalf("unterlegen: %v", fehler)
	}

	bibliothek := inZone(danach, "sitz-1", domain.ZoneBibliothek)
	if bibliothek[len(bibliothek)-1] != karte {
		t.Errorf("unten liegt %q, erwartet %q", bibliothek[len(bibliothek)-1], karte)
	}

	if len(inZone(danach, "sitz-1", domain.ZoneHand)) != domain.Starthand-1 {
		t.Error("die Hand ist nicht um eine Karte kleiner")
	}

	if sitzVon(t, danach, "sitz-1").Unterlegen != 0 {
		t.Error("nach dem Unterlegen ist noch etwas geschuldet")
	}
}

// **Behalten beendet die Mulligans** — CR 103.5: danach nimmt dieser
// Spieler keinen mehr.
func TestBehaltenBeendetDieMulligans(t *testing.T) {
	t.Parallel()

	partie, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // dito

	danach, fehler := partie.Behalte("sitz-1")
	if fehler != nil {
		t.Fatalf("behalten: %v", fehler)
	}

	if sitzVon(t, danach, "sitz-1").Eroeffnung != domain.EroeffnungBehalten {
		t.Errorf("Stand danach: %q", sitzVon(t, danach, "sitz-1").Eroeffnung)
	}

	if _, fehler := danach.NimmMulligan("sitz-1", umgekehrt(10)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Mulligan nach dem Behalten: %v", fehler)
	}

	if _, fehler := danach.Behalte("sitz-1"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("zweimal behalten: %v", fehler)
	}

	// **Ohne Starthand gibt es nichts zu behalten und keinen Mulligan.**
	ohne := mitBibliotheken(2, 10)
	if _, fehler := ohne.Behalte("sitz-1"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("behalten ohne Starthand: %v", fehler)
	}

	if _, fehler := ohne.NimmMulligan("sitz-1", umgekehrt(10)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Mulligan ohne Starthand: %v", fehler)
	}
}

// **Im Mehrspielerspiel ist der erste Mulligan frei** — CR 103.5c: er zählt
// weder für die Karten, die unter die Bibliothek gehen, noch für die Zahl
// der Mulligans.
func TestImMehrspielerIstDerErsteMulliganFrei(t *testing.T) {
	t.Parallel()

	partie, _ := mitBibliotheken(3, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // dito

	einmal, fehler := partie.NimmMulligan("sitz-1", umgekehrt(10))
	if fehler != nil {
		t.Fatalf("erster Mulligan: %v", fehler)
	}

	if sitzVon(t, einmal, "sitz-1").Unterlegen != 0 {
		t.Errorf("der erste Mulligan zu dritt verlangt %d Karten",
			sitzVon(t, einmal, "sitz-1").Unterlegen)
	}

	zweimal, fehler := einmal.NimmMulligan("sitz-1", umgekehrt(10))
	if fehler != nil {
		t.Fatalf("zweiter Mulligan: %v", fehler)
	}

	if sitzVon(t, zweimal, "sitz-1").Unterlegen != 1 {
		t.Errorf("der zweite Mulligan zu dritt verlangt %d Karten, erwartet 1",
			sitzVon(t, zweimal, "sitz-1").Unterlegen)
	}
}

// **Welche Schritte offenstehen, sagt der Kern** — und zwar dieselben, die
// er danach annimmt. Die Handlungsspalte fragt ihn, statt die Bedingungen
// nachzubauen (§16: eine Auskunft für alle Flächen).
func TestDieOffenenSchritteFolgenDemStand(t *testing.T) {
	t.Parallel()

	ausstehend := mitBibliotheken(2, 10)
	offen, _ := ausstehend.ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // der Lauf darüber hält es
	geschuldet, _ := offen.NimmMulligan("sitz-1", umgekehrt(10))   //nolint:errcheck // dito
	behalten, _ := offen.Behalte("sitz-1")                         //nolint:errcheck // dito

	erschoepft := offen
	erschoepft.Sitze = append([]domain.Sitz(nil), offen.Sitze...)
	erschoepft.Sitze[0].Mulligans = domain.Starthand

	for _, fall := range []struct {
		name   string
		partie domain.Partie
		wer    string
		soll   domain.Eroeffnungsschritte
	}{
		{"vor der Starthand", ausstehend, "sitz-1", domain.Eroeffnungsschritte{Starthand: true}},
		{"mit offener Hand", offen, "sitz-1", domain.Eroeffnungsschritte{Mulligan: true, Behalten: true}},
		{"mit Schuld", geschuldet, "sitz-1", domain.Eroeffnungsschritte{Unterlegen: true}},
		{"behalten", behalten, "sitz-1", domain.Eroeffnungsschritte{}},
		{"ohne Mulligan übrig", erschoepft, "sitz-1", domain.Eroeffnungsschritte{Behalten: true}},
		{"unbekannter Sitz", offen, "fremd", domain.Eroeffnungsschritte{}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ist := fall.partie.Eroeffnungsschritte(fall.wer); ist != fall.soll {
				t.Errorf("offen %+v, erwartet %+v", ist, fall.soll)
			}
		})
	}
}

// **Eine Partie von vorher beginnt nicht noch einmal.** Ihre Sitze tragen
// keinen Stand, und eine Starthand mitten im fünften Zug mischte die
// Bibliothek und legte sieben Karten dazu.
func TestEinePartieVonVorherBeginntNicht(t *testing.T) {
	t.Parallel()

	partie := mitBibliotheken(2, 10)
	for stelle := range partie.Sitze {
		partie.Sitze[stelle].Eroeffnung = ""
	}

	if _, fehler := partie.ZieheStarthand("sitz-1", umgekehrt(10)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Starthand in einer Partie von vorher: %v", fehler)
	}

	if schritte := partie.Eroeffnungsschritte("sitz-1"); schritte != (domain.Eroeffnungsschritte{}) {
		t.Errorf("angeboten: %+v", schritte)
	}
}
