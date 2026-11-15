package domain_test

// **Wie eine Partie endet** — Werkstück M17 der Regelmaschine
// (#1618).
//
// # Wogegen hier geprüft wird
//
//   - CR 104.1 — eine Partie endet, wenn jemand gewinnt, alle unentschieden
//     sind oder sie neu anfängt.
//   - CR 104.2a — wer übrig bleibt, gewinnt: **sofort**, und es überstimmt
//     alles, was das verhindern wollte.
//   - CR 104.3a — aufgeben geht **jederzeit**, und wer aufgibt, verlässt die
//     Partie sofort.
//   - CR 104.4a — verlieren alle Verbliebenen gleichzeitig, ist es
//     unentschieden.
//   - CR 800.4a — wer geht, nimmt seine Objekte mit.
//
// **Und warum es dieses Werkstück zuerst gibt**: `M5` setzt alle 25
// zustandsbasierten Aktionen um, und drei davon sagen „verliert die Partie".
// Ohne einen Empfänger wäre das ein Satz ohne Wirkung.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineEndpartie gibt zwei Sitze mit Objekten und laufendem Zug.
func eineEndpartie() domain.Partie {
	partie := einePartie()
	partie.ID = "p1"
	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "sitz-1/k1/0", Art: domain.ObjektKarte, Karte: "wald",
			Beherrscher: "sitz-1", Eigner: "sitz-1", Zone: domain.ZoneTisch,
		},
		{
			Kennung: "sitz-2/k1/0", Art: domain.ObjektKarte, Karte: "wald",
			Beherrscher: "sitz-2", Eigner: "sitz-2", Zone: domain.ZoneTisch,
		},
	}

	partie.Stand = domain.StandLaeuft

	return partie.MitStartleben().ZugBeginnen("sitz-1")
}

// **Wer aufgibt, ist draußen — und der andere hat gewonnen.**
func TestWerAufgibtVerliertUndDerAndereGewinnt(t *testing.T) {
	t.Parallel()

	danach := eineEndpartie().Scheidetaus("sitz-2", domain.GrundAufgabe)

	switch {
	case danach.Ausgang.Sieger != "sitz-1":
		t.Errorf("Sieger ist %q", danach.Ausgang.Sieger)
	case danach.Ausgang.Unentschieden:
		t.Error("die Partie gilt als unentschieden")
	case danach.Stand != domain.StandBeendet:
		t.Errorf("der Stand ist %q", danach.Stand)
	case len(danach.Ausgang.Ausgeschieden) != 1:
		t.Fatalf("ausgeschieden sind %+v", danach.Ausgang.Ausgeschieden)
	case danach.Ausgang.Ausgeschieden[0].Grund != domain.GrundAufgabe:
		t.Errorf("der Grund ist %q", danach.Ausgang.Ausgeschieden[0].Grund)
	}
}

// **Wer geht, nimmt seine Objekte mit** — CR 800.4a.
func TestWerGehtNimmtSeineObjekteMit(t *testing.T) {
	t.Parallel()

	danach := eineEndpartie().Scheidetaus("sitz-2", domain.GrundAufgabe)

	for _, objekt := range danach.Objekte {
		if objekt.Eigner == "sitz-2" {
			t.Errorf("%q liegt noch da", objekt.Kennung)
		}
	}

	// **Und die des anderen bleiben liegen** — ohne diese Zeile wäre der Lauf
	// auch grün, wenn das Ausscheiden den Tisch leerräumte.
	if len(danach.Objekte) != 1 {
		t.Errorf("es liegen %d Objekte", len(danach.Objekte))
	}
}

// **Der Sitz bleibt am Tisch stehen, auch wenn er draußen ist.**
//
// **Weil die Partie ihn erklären muss.** Der Verlauf nennt ihn, die
// Sichtwahl kennt ihn, und ein Ergebnis ohne Namen wäre keines. Was ihn
// draußen sein lässt, ist der Ausgang und nicht sein Verschwinden.
func TestEinAusgeschiedenerSitzBleibtAmTisch(t *testing.T) {
	t.Parallel()

	danach := eineEndpartie().Scheidetaus("sitz-2", domain.GrundAufgabe)

	if len(danach.Sitze) != 2 {
		t.Errorf("es sitzen %d am Tisch", len(danach.Sitze))
	}

	if danach.NochImSpiel() != 1 {
		t.Errorf("%d sind noch im Spiel", danach.NochImSpiel())
	}
}

// **Scheiden alle zugleich aus, ist es unentschieden** — CR 104.4a.
func TestScheidenAlleAusIstEsUnentschieden(t *testing.T) {
	t.Parallel()

	// **Gleichzeitig und nicht nacheinander** — CR 704.4: zustandsbasierte
	// Aktionen laufen alle auf einmal. Nacheinander hätte der zweite schon
	// gewonnen, bevor er selbst verliert (CR 104.2a).
	danach := eineEndpartie().Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-1", Grund: domain.GrundLeben},
		{Sitz: "sitz-2", Grund: domain.GrundLeben},
	})

	switch {
	case danach.Ausgang.Sieger != "":
		t.Errorf("Sieger ist %q", danach.Ausgang.Sieger)
	case !danach.Ausgang.Unentschieden:
		t.Error("die Partie gilt nicht als unentschieden")
	case danach.Stand != domain.StandBeendet:
		t.Errorf("der Stand ist %q", danach.Stand)
	}
}

// **Ein zweites Ausscheiden desselben Sitzes ändert nichts.**
//
// **Weil eine zustandsbasierte Aktion sich wiederholt** (CR 704.3): sie läuft,
// bis keine mehr anliegt, und ein Sitz, der schon draußen ist, darf nicht
// zweimal in der Liste stehen.
func TestEinZweitesAusscheidenAendertNichts(t *testing.T) {
	t.Parallel()

	einmal := eineEndpartie().Scheidetaus("sitz-2", domain.GrundAufgabe)
	zweimal := einmal.Scheidetaus("sitz-2", domain.GrundLeben)

	switch {
	case len(zweimal.Ausgang.Ausgeschieden) != 1:
		t.Errorf("ausgeschieden sind %+v", zweimal.Ausgang.Ausgeschieden)
	case zweimal.Ausgang.Ausgeschieden[0].Grund != domain.GrundAufgabe:
		t.Errorf("der Grund ist %q", zweimal.Ausgang.Ausgeschieden[0].Grund)
	case zweimal.Ausgang.Sieger != "sitz-1":
		t.Errorf("Sieger ist %q", zweimal.Ausgang.Sieger)
	}
}

// **Ein Sitz, den es nicht gibt, scheidet nicht aus.**
func TestEinUnbekannterSitzScheidetNichtAus(t *testing.T) {
	t.Parallel()

	danach := eineEndpartie().Scheidetaus("niemand", domain.GrundAufgabe)

	switch {
	case len(danach.Ausgang.Ausgeschieden) != 0:
		t.Errorf("ausgeschieden sind %+v", danach.Ausgang.Ausgeschieden)
	case danach.Stand != domain.StandLaeuft:
		t.Errorf("der Stand ist %q", danach.Stand)
	}
}

// **Bei einem Sitz endet die Partie nicht von selbst.**
//
// **CR 104.2a verlangt, dass die Gegner gegangen sind** — wer allein anfängt
// (Solo Goldfish, #1460), hat keine, und ein Aufbau ist kein Spiel
// (CR 100.1). Eine Partie, die sofort mit einem Sieger endete, wäre für die
// Kurvenprobe unbrauchbar.
func TestEinAlleinigerSitzGewinntNichtVonSelbst(t *testing.T) {
	t.Parallel()

	allein := eineEndpartie()
	allein.Sitze = allein.Sitze[:1]

	danach := allein.Scheidetaus("niemand", domain.GrundAufgabe)

	switch {
	case danach.Ausgang.Sieger != "":
		t.Errorf("Sieger ist %q", danach.Ausgang.Sieger)
	case danach.Stand != domain.StandLaeuft:
		t.Errorf("der Stand ist %q", danach.Stand)
	}
}

// **Und der Ausgang steht in jeder Sicht.**
//
// **Weil ein Ergebnis öffentlich ist.** Wer verloren hat, erfährt es; wer
// zusieht, sieht es. Ein Ausgang, den nur der Wirt kennt, wäre ein Ergebnis
// ohne Zeugen.
func TestDerAusgangStehtInJederSicht(t *testing.T) {
	t.Parallel()

	danach := eineEndpartie().Scheidetaus("sitz-2", domain.GrundAufgabe)

	for _, fuer := range []string{"sitz-1", "sitz-2", "zuschauer"} {
		sicht := danach.Sicht(fuer)

		if sicht.Ausgang.Sieger != "sitz-1" {
			t.Errorf("für %q ist der Sieger %q", fuer, sicht.Ausgang.Sieger)
		}

		if len(sicht.Ausgang.Ausgeschieden) != 1 {
			t.Errorf("für %q sind %d ausgeschieden",
				fuer, len(sicht.Ausgang.Ausgeschieden))
		}
	}
}

// **Bei drei Sitzen endet die Partie erst, wenn zwei gegangen sind.**
//
// **Das ist der `default`-Zweig der Ausgangsprüfung** — und er ist kein
// Formfehler: CR 104.2a verlangt, dass **alle** Gegner gegangen sind. Wer
// nach dem ersten schon einen Sieger ausriefe, beendete jede
// Mehrspielerpartie beim ersten Ausscheiden.
func TestBeiDreiSitzenEndetSieErstMitDemZweiten(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Sitze = append(partie.Sitze, domain.Sitz{
		Kennung: "sitz-3", Name: "Cem", Deck: einDeckabzug(),
	})

	nachEinem := partie.Scheidetaus("sitz-3", domain.GrundAufgabe)

	switch {
	case nachEinem.Ausgang.Sieger != "":
		t.Errorf("Sieger ist %q", nachEinem.Ausgang.Sieger)
	case nachEinem.Stand != domain.StandLaeuft:
		t.Errorf("der Stand ist %q", nachEinem.Stand)
	case nachEinem.NochImSpiel() != 2:
		t.Errorf("%d sind noch im Spiel", nachEinem.NochImSpiel())
	}

	nachZweien := nachEinem.Scheidetaus("sitz-2", domain.GrundLeben)

	if nachZweien.Ausgang.Sieger != "sitz-1" {
		t.Errorf("Sieger ist %q", nachZweien.Ausgang.Sieger)
	}
}

// **Und eine beendete Partie endet nicht noch einmal** — CR 104.1.
func TestEineBeendetePartieEndetNichtNochEinmal(t *testing.T) {
	t.Parallel()

	beendet := eineEndpartie().Scheidetaus("sitz-2", domain.GrundAufgabe)

	danach := beendet.Scheidetaus("sitz-1", domain.GrundLeben)

	switch {
	case danach.Ausgang.Sieger != "sitz-1":
		t.Errorf("Sieger ist %q", danach.Ausgang.Sieger)
	case len(danach.Ausgang.Ausgeschieden) != 1:
		t.Errorf("ausgeschieden sind %+v", danach.Ausgang.Ausgeschieden)
	}
}
