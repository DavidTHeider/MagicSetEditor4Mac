package domain_test

// **Die Handelsrichtlinie ist eine Schranke** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was hier gemessen wird
//
// **Dass „Richtlinie" hier dasselbe heißt wie überall sonst in diesem
// Programm**: eine Schranke, die über allen Ebenen steht und sich von unten
// nicht aufheben lässt — was gilt, ist die **Schnittmenge**, der strengste
// Wert gewinnt (`O3`,
// #1427,
// Vermerk 0079).
//
// **Nicht eine zweite Gestalt daneben.** Eine Handelsrichtlinie, die sich
// anders faltet als `richtlinie.nurlokal`, wäre ein zweiter Mechanismus für
// dieselbe Frage — und der erste, der sie verwechselt, hebt eine Schranke auf,
// die nicht ihm gehört.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jede Handelsschranke ist eine Richtlinie im Sinn der Faltung.**
func TestJedeHandelsschrankeIstEineRichtlinie(t *testing.T) {
	t.Parallel()

	for _, schluessel := range []string{
		domain.RichtlinieHandel, domain.RichtlinieHandelMarkt,
		domain.RichtlinieHandelMenge, domain.RichtlinieHandelWert,
		domain.RichtlinieHandelFrist,
	} {
		if !domain.IstRichtlinie(schluessel) {
			t.Errorf("%q faltet nicht wie eine Schranke", schluessel)
		}

		ebene, kennt := domain.EbeneVon(schluessel)
		if !kennt || ebene != domain.EbeneNutzer {
			t.Errorf("%q steht auf der Ebene %q (bekannt: %v)",
				schluessel, ebene, kennt)
		}
	}
}

// **Der strengste Wert gewinnt — bei jedem der fünf Schlüssel.**
func TestDerStrengsteWertGewinnt(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		werte    []string
		erwartet string
	}{
		domain.RichtlinieHandel: {
			werte:    []string{domain.HandelSchrankeAn, domain.HandelSchrankeAus},
			erwartet: domain.HandelSchrankeAus,
		},
		domain.RichtlinieHandelMarkt: {
			werte:    []string{domain.HandelSchrankeAus, domain.HandelSchrankeAn},
			erwartet: domain.HandelSchrankeAus,
		},
		domain.RichtlinieHandelMenge: {werte: []string{"9", "4"}, erwartet: "4"},
		domain.RichtlinieHandelWert:  {werte: []string{"500", "9000"}, erwartet: "500"},
		domain.RichtlinieHandelFrist: {werte: []string{"60", "15"}, erwartet: "15"},
	}

	for schluessel, fall := range faelle {
		t.Run(schluessel, func(t *testing.T) {
			t.Parallel()

			if wert := domain.Verschaerft(
				schluessel, fall.werte...); wert != fall.erwartet {
				t.Errorf("gefaltet wurde %q statt %q", wert, fall.erwartet)
			}
		})
	}
}

// **Die Vorgabe handelt — aber nicht am Markt.**
//
// **Der Markt greift auf einen fremden Anbieter und auf Geld.** Die Phase
// nennt ihn ein „optionales externes Schreibpaket, kein stiller Checkout"
// (#1410);
// eine Vorgabe, die ihn einschlösse, wäre genau der stille Checkout.
func TestDieVorgabeHandeltAberNichtAmMarkt(t *testing.T) {
	t.Parallel()

	vorgabe := domain.Standardhandelsrichtlinie()

	for _, art := range []domain.Handelsart{
		domain.HandelTeilen, domain.HandelDeckleihe, domain.HandelTausch,
	} {
		angebot := einAngebot()
		angebot.Art = art

		if fehler := vorgabe.Erlaubt(angebot); fehler != nil {
			t.Errorf("%q wird abgewiesen: %v", art, fehler)
		}
	}

	markt := einAngebot()
	markt.Art = domain.HandelMarkt

	if fehler := vorgabe.Erlaubt(markt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("der Markt kommt durch: %v", fehler)
	}
}

// **Ist der Handel aus, geht keine Art mehr.**
func TestOhneHandelGehtKeineArt(t *testing.T) {
	t.Parallel()

	aus := domain.HandelsrichtlinieAus(map[string]string{
		domain.RichtlinieHandel: domain.HandelSchrankeAus,
	})

	for _, art := range domain.Handelsarten() {
		angebot := einAngebot()
		angebot.Art = art

		if fehler := aus.Erlaubt(angebot); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Errorf("%q kommt durch: %v", art, fehler)
		}
	}
}

// **Der Markt geht, wenn er ausdrücklich angeschaltet ist.**
func TestDerMarktGehtNurAusdruecklich(t *testing.T) {
	t.Parallel()

	an := domain.HandelsrichtlinieAus(map[string]string{
		domain.RichtlinieHandelMarkt: domain.HandelSchrankeAn,
	})

	markt := einAngebot()
	markt.Art = domain.HandelMarkt

	if fehler := an.Erlaubt(markt); fehler != nil {
		t.Errorf("der angeschaltete Markt wird abgewiesen: %v", fehler)
	}
}

// **Eine Mengengrenze gilt je Position.**
//
// **Je Position und nicht je Angebot**: wer zehn Zeilen mit je einem Stück
// tauscht, tauscht nicht dasselbe wie einer mit zehn Stück in einer Zeile.
func TestEineMengengrenzeGiltJePosition(t *testing.T) {
	t.Parallel()

	eng := domain.HandelsrichtlinieAus(map[string]string{
		domain.RichtlinieHandelMenge: "1",
	})

	if fehler := eng.Erlaubt(einAngebot()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("zwei Stück kommen durch: %v", fehler)
	}

	gerade := einAngebot()
	gerade.Positionen[0].Menge = 1

	if fehler := eng.Erlaubt(gerade); fehler != nil {
		t.Errorf("ein Stück wird abgewiesen: %v", fehler)
	}
}

// **Eine Wertgrenze gilt nur, wo die Bewertung geteilt ist — und über das
// ganze Angebot.**
//
// **Sonst begrenzte sie eine Zahl, über die niemand gesprochen hat**, und
// eine neue Marktansicht machte ein gültiges Angebot unzulässig. Derselbe
// Satz wie bei der Unterschrift.
func TestEineWertgrenzeGiltNurBeiGeteilterBewertung(t *testing.T) {
	t.Parallel()

	eng := domain.HandelsrichtlinieAus(map[string]string{
		domain.RichtlinieHandelWert: "1000",
	})

	still := einAngebot()
	still.Positionen[0].Bewertung = 5_000

	if fehler := eng.Erlaubt(still); fehler != nil {
		t.Errorf("eine stille Bewertung wird begrenzt: %v", fehler)
	}

	geteilt := still
	geteilt.BewertungGeteilt = true

	if fehler := eng.Erlaubt(geteilt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die geteilte Bewertung kommt durch: %v", fehler)
	}

	// **Zwei Zeilen zu sechshundert sind zwölfhundert** — die Grenze meint
	// die Summe und nicht die Zeile.
	zweimal := einAngebot()
	zweimal.BewertungGeteilt = true
	zweimal.Positionen[0].Bewertung = 600
	zweite := zweimal.Positionen[0]
	zweimal.Positionen = append(zweimal.Positionen, zweite)

	if fehler := eng.Erlaubt(zweimal); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("zweimal sechshundert kommen unter tausend durch: %v", fehler)
	}

	// **Und darunter kommt durch** — sonst prüfte der Lauf nur, dass
	// überhaupt etwas abgewiesen wird (§10).
	darunter := einAngebot()
	darunter.BewertungGeteilt = true
	darunter.Positionen[0].Bewertung = 900

	if fehler := eng.Erlaubt(darunter); fehler != nil {
		t.Errorf("neunhundert unter tausend werden abgewiesen: %v", fehler)
	}
}

// **Eine Fristgrenze begrenzt, wie lange eine Zusage festhalten darf.**
//
// **In Minuten**, weil eine Schranke sich falten muss und
// `kleinsteGrenze` Zahlen vergleicht — eine Dauer wie `1h30m` ließe sich
// nicht der Größe nach ordnen, ohne sie vorher zu deuten.
func TestEineFristgrenzeBegrenztDieZusage(t *testing.T) {
	t.Parallel()

	eng := domain.HandelsrichtlinieAus(map[string]string{
		domain.RichtlinieHandelFrist: "60",
	})

	if fehler := eng.ErlaubtFrist(2 * time.Hour); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("zwei Stunden kommen durch: %v", fehler)
	}

	if fehler := eng.ErlaubtFrist(time.Minute); fehler != nil {
		t.Errorf("eine Minute wird abgewiesen: %v", fehler)
	}

	// **Eine Grenze über die Dauer hinaus läuft nicht um** (J05, #2269):
	// 153.722.868.000 Minuten mal `time.Minute` wäre negativ, und jede
	// Zusage stieße an eine Grenze, die niemand gesetzt hat.
	weit := domain.HandelsrichtlinieAus(map[string]string{
		domain.RichtlinieHandelFrist: "153722868000",
	})
	if fehler := weit.ErlaubtFrist(1000 * time.Hour); fehler != nil {
		t.Errorf("eine übergroße Grenze weist ab: %v", fehler)
	}

	// **Ohne Grenze geht jede Frist** — leer heißt „keine Grenze", nicht
	// „null Minuten".
	if fehler := domain.Standardhandelsrichtlinie().ErlaubtFrist(
		1000 * time.Hour); fehler != nil {
		t.Errorf("ohne Grenze wird abgewiesen: %v", fehler)
	}
}

// **Was fehlt oder keinen Sinn ergibt, ist die Vorgabe.**
//
// **Eine Einstellung ist fremde Eingabe** (§8): sie kommt aus einer Datei,
// die ein Mensch ändern kann. Wer daraus eine Grenze von null läse, hätte
// eine Schranke, die alles verbietet, ohne dass jemand es wollte.
func TestEineUnlesbareSchrankeFaelltAufDieVorgabeZurueck(t *testing.T) {
	t.Parallel()

	gelesen := domain.HandelsrichtlinieAus(map[string]string{
		domain.RichtlinieHandel:      "vielleicht",
		domain.RichtlinieHandelMenge: "viele",
		domain.RichtlinieHandelWert:  "-3",
		domain.RichtlinieHandelFrist: "bald",
	})

	if gelesen != domain.Standardhandelsrichtlinie() {
		t.Errorf("gelesen wurde %+v statt der Vorgabe %+v",
			gelesen, domain.Standardhandelsrichtlinie())
	}
}

// **Die Richtlinie wiederholt die Prüfung des Angebots nicht.**
//
// **Ein Angebot, das keines ist, ist kein Fall für die Schranke** — sonst
// stünde an zwei Stellen dieselbe Ablehnung mit zwei verschiedenen Sätzen.
func TestDieRichtlinieWiederholtDieAngebotspruefungNicht(t *testing.T) {
	t.Parallel()

	kaputt := einAngebot()
	kaputt.Positionen = nil

	if fehler := domain.Standardhandelsrichtlinie().Erlaubt(
		kaputt); fehler != nil {
		t.Errorf("die Schranke urteilt über ein ungültiges Angebot: %v", fehler)
	}
}

// **Eine Grenze ist eine ganze Zahl ab eins — und leer heißt „keine".**
//
// **Über [domain.PruefeEinstellung] und nicht über die Funktion selbst**: das
// ist der Weg, den jede Fläche geht. Eine Regel, die nur im Paket gilt, ist
// keine Einstellung, sondern eine Absicht.
func TestEineHandelsgrenzeIstEineGanzeZahl(t *testing.T) {
	t.Parallel()

	for _, schluessel := range []string{
		domain.RichtlinieHandelMenge, domain.RichtlinieHandelWert,
		domain.RichtlinieHandelFrist,
	} {
		t.Run(schluessel, func(t *testing.T) {
			t.Parallel()

			for _, gut := range []string{"", "  ", "1", "4200"} {
				if fehler := domain.PruefeEinstellung(
					schluessel, gut); fehler != nil {
					t.Errorf("%q wird abgewiesen: %v", gut, fehler)
				}
			}

			for _, schlecht := range []string{"viele", "-3", "0", "1,5"} {
				if fehler := domain.PruefeEinstellung(
					schluessel, schlecht); !errors.Is(
					fehler, domain.ErrValidation) {
					t.Errorf("%q kommt durch: %v", schlecht, fehler)
				}
			}
		})
	}
}

// **Und die zwei Schalter nehmen nur „an" und „aus".**
func TestEineHandelsschrankeNimmtNurAnUndAus(t *testing.T) {
	t.Parallel()

	for _, schluessel := range []string{
		domain.RichtlinieHandel, domain.RichtlinieHandelMarkt,
	} {
		for _, gut := range []string{
			domain.HandelSchrankeAn, domain.HandelSchrankeAus,
		} {
			if fehler := domain.PruefeEinstellung(
				schluessel, gut); fehler != nil {
				t.Errorf("%s=%q wird abgewiesen: %v", schluessel, gut, fehler)
			}
		}

		if fehler := domain.PruefeEinstellung(
			schluessel, "vielleicht"); fehler == nil {
			t.Errorf("%s nimmt „vielleicht\"", schluessel)
		}
	}
}
