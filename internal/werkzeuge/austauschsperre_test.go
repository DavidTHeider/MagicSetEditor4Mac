package werkzeuge_test

// **Was nie in einen Austausch darf** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum es diesen Wächter gibt
//
// **Weil die Sperre sonst an einer Aufzählung hinge, die niemand nachzählt.**
// Was heute nicht mitreist, sind drei Schlüssel und zwei Familien; wer morgen
// einen vierten aufnimmt, trifft die Entscheidung an einer Stelle
// (`einstellungsebene.go`) — und dieser Lauf sagt ihm, wenn sein Name nach
// einem Geheimnis, einem Pfad oder einer lokalen Kennung klingt und er ihn
// trotzdem reisen lässt.
//
// # Was er hält
//
// Für **jeden** Schlüssel beider Reihen, dessen Name eines der Sperrwörter
// enthält: er reist weder aus dem Profil noch vom Menschen mit. Und
// umgekehrt: was mitreist, steht auf einer Ebene, die mitreisen darf.
//
// # Was er nicht hält
//
// **Ob ein Schlüssel ohne verdächtigen Namen ein Geheimnis trägt.** Das kann
// er nicht sehen; dafür steht die Regel in `einstellungsebene.go`, und dafür
// liest ein Mensch die Ebenentabelle. **Und er sieht keinen Wert** — dass eine
// gebaute Ausfuhr wirklich keinen Pfad enthält, belegt
// `internal/service/einstellungsaustausch_test.go` und der Lauf über alle drei
// Flächen in `internal/flaechen/einstellungsaustausch_test.go`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sperrwoerter sind die Namensteile, die einen Schlüssel am Reisen hindern.
//
// **Eine Aufzählung von Wörtern und keine von Schlüsseln**: eine Liste von
// Schlüsseln wäre beim nächsten neuen wieder vollständig zu machen, und
// niemand sähe, dass sie es nicht ist.
var sperrwoerter = []string{
	"pfad", "kennung", "token", "geheim", "kennwort", "passwort",
	"secret", "schluesselbund", "totp",
}

// alleEinstellungsschluessel gibt die Vereinigung beider Reihen.
func alleEinstellungsschluessel() []string {
	gesehen := map[string]bool{}
	alle := []string{}

	for _, reihe := range [][]string{
		domain.Einstellungsreihe(), domain.Nutzereinstellungsreihe(),
	} {
		for _, schluessel := range reihe {
			if !gesehen[schluessel] {
				gesehen[schluessel] = true

				alle = append(alle, schluessel)
			}
		}
	}

	return alle
}

func TestKeinVerdaechtigerSchluesselReistMit(t *testing.T) {
	t.Parallel()

	for _, schluessel := range alleEinstellungsschluessel() {
		if !verdaechtig(schluessel) {
			continue
		}

		if _, reist := domain.AbschnittImProfil(schluessel); reist {
			t.Errorf("%q reist aus dem Profil mit — sein Name nennt etwas, "+
				"das auf diesem Rechner bleibt", schluessel)
		}

		if domain.ReistBeimNutzer(schluessel) {
			t.Errorf("%q reist vom Menschen mit — sein Name nennt etwas, "+
				"das auf diesem Rechner bleibt", schluessel)
		}
	}
}

// verdaechtig sagt, ob ein Name eines der Sperrwörter enthält.
func verdaechtig(schluessel string) bool {
	klein := strings.ToLower(schluessel)

	for _, wort := range sperrwoerter {
		if strings.Contains(klein, wort) {
			return true
		}
	}

	return false
}

// **Die Gegenprobe: es gibt überhaupt einen verdächtigen Schlüssel.**
//
// **Ohne sie hielte der Lauf darüber nichts.** Eine Schleife, die über null
// Fälle läuft, ist grün — und sähe genauso aus wie eine, die greift. Heute
// sind es der Hochladepfad und vier Anbieterkennungen.
func TestEsGibtVerdaechtigeSchluessel(t *testing.T) {
	t.Parallel()

	gezaehlt := 0

	for _, schluessel := range alleEinstellungsschluessel() {
		if verdaechtig(schluessel) {
			gezaehlt++
		}
	}

	if gezaehlt < 2 {
		t.Errorf("es sind %d — der Wächter daneben liefe über nichts", gezaehlt)
	}
}

// **Und was mitreist, steht auf einer Ebene, die mitreisen darf.**
func TestWasMitreistStehtAufEinerReisendenEbene(t *testing.T) {
	t.Parallel()

	for _, schluessel := range domain.Einstellungsreihe() {
		abschnitt, reist := domain.AbschnittImProfil(schluessel)
		if !reist {
			continue
		}

		if abschnitt == domain.EbeneGeraet {
			t.Errorf("%q reist im Abschnitt »geraet«", schluessel)
		}

		ebene, _ := domain.EbeneVon(schluessel)
		if ebene == domain.EbeneGeraet {
			t.Errorf("%q gehört dem Rechner und reist trotzdem", schluessel)
		}
	}

	for _, schluessel := range domain.Nutzereinstellungsreihe() {
		if !domain.ReistBeimNutzer(schluessel) {
			continue
		}

		if ebene, _ := domain.EbeneVon(schluessel); ebene != domain.EbeneNutzer {
			t.Errorf("%q reist vom Menschen, gehört aber der Ebene %q",
				schluessel, ebene)
		}
	}
}
