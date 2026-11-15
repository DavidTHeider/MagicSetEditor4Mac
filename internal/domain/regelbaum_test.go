package domain_test

// **Die Regeln als Baum** — PLAY-02.AK-02
// (#1710).

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Bereich steht in genau einer Gruppe** — keiner fehlt, keiner
// doppelt.
//
// **Das ist die Zusage, für die es den Baum gibt** (§10): eine Gruppierung,
// die einen Bereich verlöre, zeigte ihn nirgends — und niemand sähe es, weil
// zwölf Zeilen wie dreizehn aussehen.
func TestJederRegelbereichStehtInGenauEinerGruppe(t *testing.T) {
	t.Parallel()

	gesehen := map[domain.Regelbereich]int{}

	for _, gruppe := range domain.Regelbaum() {
		for _, bereich := range gruppe.Bereiche {
			gesehen[bereich]++
		}
	}

	for _, bereich := range domain.Regelbereiche() {
		if gesehen[bereich] != 1 {
			t.Errorf("%q steht %dmal im Baum", bereich, gesehen[bereich])
		}
	}

	if len(gesehen) != len(domain.Regelbereiche()) {
		t.Errorf("der Baum führt %d Bereiche, es gibt %d",
			len(gesehen), len(domain.Regelbereiche()))
	}
}

// **Jede Abhängigkeit nennt zwei bekannte Bereiche und eine Quelle.**
//
// **Eine Abhängigkeit ohne Regelnummer wäre eine Meinung** (§11): sie sähe
// aus wie ein Beleg, und niemand schlüge sie nach.
func TestJedeAbhaengigkeitTraegtIhreQuelle(t *testing.T) {
	t.Parallel()

	for _, abhaengig := range domain.Regelabhaengigkeiten() {
		switch {
		case !domain.IstRegelbereich(abhaengig.Bereich):
			t.Errorf("%q ist kein Bereich", abhaengig.Bereich)
		case !domain.IstRegelbereich(abhaengig.Braucht):
			t.Errorf("%q ist kein Bereich", abhaengig.Braucht)
		case abhaengig.Bereich == abhaengig.Braucht:
			t.Errorf("%q hängt an sich selbst", abhaengig.Bereich)
		case len(abhaengig.Quelle.Nummern) == 0:
			t.Errorf("%q → %q ohne Regelnummer",
				abhaengig.Bereich, abhaengig.Braucht)
		case abhaengig.Quelle.Fassung == "":
			t.Errorf("%q → %q ohne Fassung",
				abhaengig.Bereich, abhaengig.Braucht)
		}
	}
}

// **Eine Abwahl, an der noch etwas hängt, wird erklärt.**
func TestEineAbwahlMitAnhangWirdErklaert(t *testing.T) {
	t.Parallel()

	wahl := domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichPrioritaet},
	}

	widersprueche := wahl.Widersprueche()
	if len(widersprueche) == 0 {
		t.Fatal("die abgewählte Priorität blieb ohne Widerspruch")
	}

	haengt := make([]domain.Regelbereich, 0, len(widersprueche))

	for _, widerspruch := range widersprueche {
		if widerspruch.Bereich != domain.BereichPrioritaet {
			t.Errorf("der Widerspruch nennt %q", widerspruch.Bereich)
		}

		haengt = append(haengt, widerspruch.Haengt)
	}

	// **Stapel und zustandsbasierte Aktionen hängen daran** — CR 608.1 und
	// CR 704.3.
	for _, erwartet := range []domain.Regelbereich{
		domain.BereichStapel, domain.BereichZustandsbasiert,
	} {
		if !slices.Contains(haengt, erwartet) {
			t.Errorf("%q fehlt unter %v", erwartet, haengt)
		}
	}
}

// **Wer beides abwählt, bekommt keinen Widerspruch** — das ist der Ausweg,
// den die Erklärung anbietet.
func TestWerBeidesAbwaehltBekommtKeinenWiderspruch(t *testing.T) {
	t.Parallel()

	wahl := domain.Regelwahl{Abgewaehlt: []domain.Regelbereich{
		domain.BereichPrioritaet, domain.BereichStapel,
		domain.BereichZustandsbasiert,
	}}

	if widersprueche := wahl.Widersprueche(); len(widersprueche) != 0 {
		t.Errorf("es blieb ein Widerspruch: %+v", widersprueche)
	}
}

// **Eine leere Wahl widerspricht nichts** — und gibt eine leere Liste.
func TestEineLeereWahlWidersprichtNichts(t *testing.T) {
	t.Parallel()

	widersprueche := domain.Regelwahl{}.Widersprueche()
	if widersprueche == nil {
		t.Fatal("die Liste ist nil statt leer")
	}

	if len(widersprueche) != 0 {
		t.Errorf("es kam %+v", widersprueche)
	}
}

// **Ein Name, den es nicht gibt, wird abgewiesen** — und zwar mit ihm im
// Satz (§8).
func TestEinUnbekannterRegelbereichWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	fehler := domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{"zonnen"},
	}.Pruefe()

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "zonnen") {
		t.Errorf("die Meldung nennt den Tippfehler nicht: %v", fehler)
	}
}

// **Eine gültige Wahl geht durch.**
func TestEineGueltigeWahlGehtDurch(t *testing.T) {
	t.Parallel()

	wahl := domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichCommander},
	}

	if fehler := wahl.Pruefe(); fehler != nil {
		t.Fatalf("die gültige Wahl wurde abgewiesen: %v", fehler)
	}
}

// **Abgewählt heißt manuell — und die Zusage geht mit weg.**
//
// **Nicht `unbekannt`**: der Unterschied ist, dass jemand entschieden hat,
// es selbst zu machen.
func TestEineAbwahlMachtAusAutomatischManuell(t *testing.T) {
	t.Parallel()

	profil := domain.Grundprofil("magic")

	vorher, gibtEs := profil.Matrix.Eintrag(domain.BereichZonen)
	if !gibtEs || vorher.Grad != domain.UnterstuetzungAutomatisch {
		t.Fatalf("die Vorlage führt die Zonen als %q", vorher.Grad)
	}

	wahl := domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichZonen},
	}

	wirksam := wahl.Wirksam(profil.Matrix)

	nachher, gibtEs := wirksam.Eintrag(domain.BereichZonen)

	switch {
	case !gibtEs:
		t.Fatal("der abgewählte Bereich fiel ganz heraus")
	case nachher.Grad != domain.UnterstuetzungManuell:
		t.Errorf("der Grad ist %q", nachher.Grad)
	case len(nachher.Quelle.Nummern) != 0:
		t.Errorf("die Quelle blieb stehen: %+v", nachher.Quelle)
	case len(nachher.Grenzen) != 0:
		t.Errorf("die Grenzen blieben stehen: %v", nachher.Grenzen)
	}

	// **Und die Vorlage bleibt, wie sie war** — sonst änderte eine Runde die
	// Regeln jeder anderen mit.
	if profil.Matrix.Grad(domain.BereichZonen) !=
		domain.UnterstuetzungAutomatisch {
		t.Error("die Vorlage wurde mitgeändert")
	}

	// **Was bleibt, ist unangetastet** — eine Abwahl trifft ihren Bereich
	// und keinen zweiten (§10).
	if wirksam.Grad(domain.BereichMarken) !=
		profil.Matrix.Grad(domain.BereichMarken) {
		t.Error("ein anderer Bereich hat sich mitverändert")
	}

	// **Und die gewandelte Matrix bleibt prüfbar** — eine Quelle neben
	// „manuell" wiese `Pruefe` ab.
	if fehler := wirksam.Pruefe(); fehler != nil {
		t.Errorf("die gewandelte Matrix ist ungültig: %v", fehler)
	}
}

// **Ein Regelwerk mit einem Bereich, den es nicht gibt, wird abgewiesen** —
// PLAY-02.AK-03
// (#1710).
//
// **Hier und nicht erst im Aufbau**: die Abwahl ist Teil des Regelwerks, und
// wer sie erst eine Schicht höher prüfte, hätte eine Partie, die mit einem
// Tippfehler im Bestand steht.
func TestEinRegelwerkMitUnbekannterAbwahlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk := domain.Regelwerk{
		Kennung:    "sandkasten",
		Fassung:    1,
		Strenge:    domain.StrengeSandkasten,
		Abgewaehlt: []domain.Regelbereich{"zonnen"},
	}

	fehler := werk.Pruefe()
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "zonnen") {
		t.Errorf("die Meldung nennt den Tippfehler nicht: %v", fehler)
	}
}
