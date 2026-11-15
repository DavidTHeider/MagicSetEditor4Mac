package service_test

// **Was hinausginge** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Der Satz, gegen den hier geprüft wird**: gezeigt wird der Umfang des
// **zusammengestellten** Kontexts — nicht der der Anfrage. Zwischen beidem
// liegt die Auswahl.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einEigenertreffer ist ein Text aus diesem Bestand.
func einEigenertreffer(kennung, text string, rang int) domain.Retrievaltreffer {
	return domain.Retrievaltreffer{
		Rang: rang,
		Dokument: domain.Retrievaldokument{
			Kennung: kennung, Quelle: domain.QuelleEigenerText,
			Titel: "Meine Notiz", Text: text, Sprache: "de",
			Quellrevision: "1", Abgerufen: dieRechercheuhr(),
			ProfilID: 1, Sichtbarkeit: domain.SichtbarProfil,
		},
	}
}

// eineFreigabe stellt ein Paket zusammen und beschreibt es.
func eineFreigabe(
	t *testing.T, treffer []domain.Retrievaltreffer, ort domain.KIOrt,
) service.Kontextfreigabe {
	t.Helper()

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand: einRegelstand(), treffer: treffer,
	}).Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	return service.FreigabeFuer(paket, ort, "")
}

// **Die Freigabe nennt Anbieter, Umfang und Kategorien.**
func TestDieFreigabeNenntAnbieterUmfangUndKategorien(t *testing.T) {
	t.Parallel()

	freigabe := eineFreigabe(t, []domain.Retrievaltreffer{
		einRegeltreffer("702.9b", "Trampelschaden ändert die Zuweisung.", 1),
	}, domain.KIOrtLokal)

	if freigabe.Zielanbieter != "ollama" || freigabe.Belege != 1 {
		t.Fatalf("die Freigabe ist %+v", freigabe)
	}

	if freigabe.Tokens <= 0 || freigabe.Zeichen <= 0 {
		t.Errorf("der Umfang ist %d Tokens, %d Zeichen",
			freigabe.Tokens, freigabe.Zeichen)
	}

	// **Die Aufgabe zählt mit** — sie geht mit hinaus.
	if len(freigabe.Kategorien) != 2 {
		t.Fatalf("die Kategorien sind %v", freigabe.Kategorien)
	}

	if freigabe.Privat {
		t.Error("eine Regelstelle galt als eigener Inhalt")
	}

	if freigabe.Quellen[domain.QuelleRegelwerk] != 1 {
		t.Errorf("die Quellen sind %+v", freigabe.Quellen)
	}
}

// **Eigener Inhalt steht als eigene Zeile da.**
//
// **Und nicht nur in der Kategorienliste**: das ist die Frage, bei der jemand
// zweimal hinsieht, und sie soll nicht in einer Aufzählung stehen, die man
// überfliegt.
func TestEigenerInhaltStehtAlsEigeneZeileDa(t *testing.T) {
	t.Parallel()

	freigabe := eineFreigabe(t, []domain.Retrievaltreffer{
		einRegeltreffer("702.9b", "Eine Regel.", 1),
		einEigenertreffer("meins:0", "Eine eigene Notiz zu meinem Set.", 2),
	}, domain.KIOrtCloud)

	if !freigabe.Privat {
		t.Fatal("der eigene Inhalt fiel nicht auf")
	}

	if !strings.Contains(freigabe.Zusammenfassung(), "eigene Inhalte") {
		t.Errorf("die Zusammenfassung ist %q", freigabe.Zusammenfassung())
	}

	// **Und in der Cloud steht nicht, dass hier gerechnet wird.**
	if strings.Contains(freigabe.Zusammenfassung(), "diesem Rechner") {
		t.Errorf("die Zusammenfassung ist %q", freigabe.Zusammenfassung())
	}
}

// **Wo lokal gerechnet wird, steht es dabei.**
func TestWoLokalGerechnetWirdStehtEsDabei(t *testing.T) {
	t.Parallel()

	freigabe := eineFreigabe(t, []domain.Retrievaltreffer{
		einRegeltreffer("702.9b", "Eine Regel.", 1),
	}, domain.KIOrtLokal)

	satz := freigabe.Zusammenfassung()

	if !strings.Contains(satz, "diesem Rechner") ||
		!strings.Contains(satz, "1 Beleg ") {
		t.Fatalf("die Zusammenfassung ist %q", satz)
	}
}

// **Jede Quelle bekommt ihre Kategorie.**
//
// **Und die eigene Karte ist eine eigene** — sie steht im Bestand und gehört
// einem Profil, auch wenn ihr Text wie ein Oracle-Text aussieht.
func TestJedeQuelleBekommtIhreKategorie(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		quelle domain.Retrievalquelle
		privat bool
	}{
		"Oracle":       {domain.QuelleOracle, false},
		"Ruling":       {domain.QuelleRuling, false},
		"Regelwerk":    {domain.QuelleRegelwerk, false},
		"eigene Karte": {domain.QuelleEigeneKarte, true},
		"eigener Text": {domain.QuelleEigenerText, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			freigabe := service.FreigabeFuer(service.Kontextpaket{
				Manifest: domain.Kontextmanifest{
					Zielanbieter: "ollama", Zielmodell: "m",
					Belege: []domain.Beleg{{Kennung: "B1", Quelle: fall.quelle}},
				},
			}, domain.KIOrtLokal, "")

			if freigabe.Privat != fall.privat {
				t.Fatalf("%s gilt als privat=%v", name, freigabe.Privat)
			}
		})
	}
}

// **Eine Quelle, die niemand kennt, gilt als eigen.**
//
// **Und nicht als öffentlich**: eine neue Quelle stillschweigend
// hinauszulassen ist die Gestalt, in der ein Datenabfluss entsteht, den
// niemand entschieden hat.
func TestEineUnbekannteQuelleGiltAlsEigen(t *testing.T) {
	t.Parallel()

	fremd := einRegeltreffer("x", "Ein Text.", 1)
	fremd.Dokument.Quelle = domain.Retrievalquelle("etwas-neues")

	// **Über den Dienst geht das nicht** — der Filter ließe es nicht durch.
	// Geprüft wird deshalb die Zuordnung selbst, über ein Paket von Hand.
	freigabe := service.FreigabeFuer(service.Kontextpaket{
		Manifest: domain.Kontextmanifest{
			Zielanbieter: "ollama", Zielmodell: "m",
			Belege: []domain.Beleg{{
				Kennung: "B1", Quelle: domain.Retrievalquelle("etwas-neues"),
			}},
		},
	}, domain.KIOrtCloud, "")

	if !freigabe.Privat {
		t.Fatal("eine unbekannte Quelle galt als öffentlich")
	}
}
