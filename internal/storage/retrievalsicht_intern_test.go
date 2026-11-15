package storage

// **Die zweite Sichtprüfung** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum der Lauf im Paket steht
//
// **Weil die Prüfung von außen nicht zu erreichen ist.** Sie ist die zweite
// von zweien: die `WHERE`-Zeile hält die Kandidatenmenge klein, und diese
// hier hält, dass keine Zeile durchkommt, die eine geänderte Abfrage doch
// einsammelte. Genau deshalb betritt sie im Betrieb niemand — und genau
// deshalb wäre sie ohne diesen Lauf eine Zusage, die nie geprüft wurde (§10).
//
// **Das Issue verlangt beide**: „ACL bereits in Kandidatenauswahl und erneut
// vor Ergebnis-/Snippet-/Kontextausgabe prüfen."

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Wer was sehen darf.**
func TestDieZweiteSichtpruefungHaeltDieProfile(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		dokument domain.Retrievaldokument
		profil   int64
		erwartet bool
	}{
		"öffentlich für jeden": {
			domain.Retrievaldokument{Sichtbarkeit: domain.SichtbarOeffentlich}, 7, true,
		},
		"eigenes für den Eigentümer": {
			domain.Retrievaldokument{
				Sichtbarkeit: domain.SichtbarProfil, ProfilID: 7,
			}, 7, true,
		},
		"eigenes für einen anderen": {
			domain.Retrievaldokument{
				Sichtbarkeit: domain.SichtbarProfil, ProfilID: 7,
			}, 8, false,
		},
		"eigenes ohne Profil im Kontext": {
			domain.Retrievaldokument{
				Sichtbarkeit: domain.SichtbarProfil, ProfilID: 7,
			}, 0, false,
		},
		// **Ein drittes Wort ist keine Auskunft über Sichtbarkeit.** Die
		// Abfrage ließe diese Zeile durch — sie fragt „öffentlich oder mein
		// Profil", und das Profil stimmt. Hier fällt sie.
		"mit unbekannter Sichtbarkeit": {
			domain.Retrievaldokument{Sichtbarkeit: "geheim", ProfilID: 7}, 7, false,
		},
		"ohne Sichtbarkeit": {
			domain.Retrievaldokument{ProfilID: 7}, 7, false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if darf := sichtbar(fall.dokument, fall.profil); darf != fall.erwartet {
				t.Fatalf("sichtbar gab %v, erwartet %v", darf, fall.erwartet)
			}
		})
	}
}

// **Ein halber Wert am Ende eines Vektors wird weggelassen.**
//
// **Er wäre eine Zahl, die niemand geschrieben hat.** Der Fall entsteht an
// einem Bestand, den jemand von Hand bearbeitet hat (§8).
func TestEinHalberWertAmEndeFaelltWeg(t *testing.T) {
	t.Parallel()

	vektor := alsVektor(append(alsBlob([]float32{1, 2}), 0, 0))

	if len(vektor) != 2 || vektor[0] != 1 || vektor[1] != 2 {
		t.Fatalf("der Vektor ist %v", vektor)
	}
}

// **Und ein Vektor kommt heraus, wie er hineinging.**
func TestEinVektorKommtHerausWieErHineinging(t *testing.T) {
	t.Parallel()

	hinein := []float32{0.5, -0.25, 0}

	heraus := alsVektor(alsBlob(hinein))
	if len(heraus) != len(hinein) {
		t.Fatalf("heraus kamen %d Zahlen statt %d", len(heraus), len(hinein))
	}

	for i := range hinein {
		if heraus[i] != hinein[i] {
			t.Fatalf("Stelle %d ist %v statt %v", i, heraus[i], hinein[i])
		}
	}
}
