package service_test

// **Das Kontextpaket** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Wogegen hier geprüft wird
//
// **Dass der Dienst die Belege vergibt** und nicht das Modell. Was ein Modell
// nennen darf, steht im Manifest; alles andere ist ein Befund.
//
// **Und dass das Budget hält.** Ein Paket, das das Fenster sprengt, wird
// abgeschnitten — vom Modell, still, an einer Stelle, die niemand kennt.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einKontextwunsch ist ein vollständiger Auftrag.
func einKontextwunsch() service.Kontextwunsch {
	return service.Kontextwunsch{
		Aufgabe:      "Erkläre, wie Trampelschaden zugewiesen wird",
		Zielrevision: "rev-1",
		Anfrage: domain.Rechercheanfrage{
			Fall: domain.FallRegeln, Frage: "trampelschaden", Grenze: 10,
		},
		Zielanbieter: "ollama", Zielmodell: "gemma3:270m",
		Laufzeit:      "ollama 0.33.3",
		Kontextgrenze: 4096, Ausgabereserve: 1024,
	}
}

// einKontextdienst baut ihn über einer Recherchequelle.
func einKontextdienst(quelle *recherchequellendoppel) *service.Kontextdienst {
	return service.NeuerKontextdienst(
		einRecherchedienst(quelle, nil), profildoppel{}).
		MitUhr(dieRechercheuhr).
		MitKennung(func() string { return "kp-probe" })
}

// einRegeltreffer baut einen Treffer des Regelwerks.
func einRegeltreffer(nummer, text string, rang int) domain.Retrievaltreffer {
	return domain.Retrievaltreffer{
		Rang: rang, Ausschnitt: "…" + text[:min(20, len(text))] + "…",
		Dokument: domain.Retrievaldokument{
			Kennung: "regel:" + nummer, Quelle: domain.QuelleRegelwerk,
			Titel: "Regel " + nummer, Text: text, Regelnummer: nummer,
			Sprache: "de", Quellrevision: "2026-07-26",
			Quelldatum: "2026-07-26", Abgerufen: dieRechercheuhr(),
		},
	}
}

// einRegelstand ist ein Index, der das Regelwerk führt.
func einRegelstand() service.Indexstand {
	stand := einAktiverStand()
	stand.Manifest.Korpusstaende = []domain.Korpusstand{{
		Quelle: domain.QuelleRegelwerk, Revision: "2026-07-26",
		Datum: "2026-07-26",
	}}

	return stand
}

// **Der Dienst vergibt die Belege, und das Manifest führt sie.**
func TestDerDienstVergibtDieBelege(t *testing.T) {
	t.Parallel()

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand: einRegelstand(),
		treffer: []domain.Retrievaltreffer{
			einRegeltreffer("702.9b", "Trampelschaden ändert, wie überschüssiger Kampfschaden zugewiesen wird.", 1),
			einRegeltreffer("509.1a", "Der verteidigende Spieler wählt, ob geblockt wird.", 2),
		},
	}).Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	if len(paket.Manifest.Belege) != 2 {
		t.Fatalf("die Belege sind %+v", paket.Manifest.Belege)
	}

	// **Die Kennungen zählen ab eins.**
	if paket.Manifest.Belege[0].Kennung != "B1" ||
		paket.Manifest.Belege[1].Kennung != "B2" {
		t.Fatalf("die Kennungen sind %v", paket.Manifest.Kennungen())
	}

	// **Und der Kontext trägt sie vor dem Ausschnitt** — sonst kann ein
	// Modell sie nicht nennen.
	if !strings.Contains(paket.Kontext, "[B1] ") ||
		!strings.Contains(paket.Kontext, "[B2] ") {
		t.Fatalf("der Kontext ist %q", paket.Kontext)
	}

	// **Mit der Herkunft dahinter** — damit das Modell sie nicht erfinden
	// muss.
	if !strings.Contains(paket.Kontext, "702.9b") ||
		!strings.Contains(paket.Kontext, "2026-07-26") {
		t.Errorf("der Kontext nennt die Herkunft nicht: %q", paket.Kontext)
	}
}

// **Das Manifest ist vollständig — und nennt seine Zählweise.**
//
// **Ohne sie wäre jede Zahl darin bedeutungslos**: was ein Modell als Token
// zählt, zählt das nächste anders.
func TestDasManifestNenntSeineZaehlweise(t *testing.T) {
	t.Parallel()

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand:   einRegelstand(),
		treffer: []domain.Retrievaltreffer{einRegeltreffer("702.9b", "Ein Text.", 1)},
	}).Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	manifest := paket.Manifest

	if manifest.Zaehlweise != service.Zaehlweise ||
		manifest.Kontextpolicy != domain.Kontextpolicyfassung {
		t.Fatalf("das Manifest ist %+v", manifest)
	}

	if manifest.PaketID != "kp-probe" || manifest.ProfilID != 1 {
		t.Errorf("Kennung %q, Profil %d", manifest.PaketID, manifest.ProfilID)
	}

	// **Und es läuft ab.**
	if !manifest.Laeuftab.Equal(dieRechercheuhr().Add(domain.Kontextfrist)) {
		t.Errorf("es läuft %v ab", manifest.Laeuftab)
	}

	// **Der Korpusstand steht darin** — er sagt, gegen welchen Regelstand
	// das Modell geantwortet hat.
	if len(manifest.Korpusstaende) != 1 {
		t.Errorf("die Stände sind %+v", manifest.Korpusstaende)
	}
}

// **Was nicht ins Budget passt, wird gekürzt — sichtbar.**
func TestWasNichtInsBudgetPasstWirdSichtbarGekuerzt(t *testing.T) {
	t.Parallel()

	lang := strings.Repeat("Dies ist ein sehr langer Regeltext. ", 200)

	wunsch := einKontextwunsch()
	wunsch.Kontextgrenze = 400
	wunsch.Ausgabereserve = 100

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand:   einRegelstand(),
		treffer: []domain.Retrievaltreffer{einRegeltreffer("702.9b", lang, 1)},
	}).Zusammenstellen(t.Context(), wunsch)
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	if len(paket.Manifest.Belege) != 1 {
		t.Fatalf("die Belege sind %d", len(paket.Manifest.Belege))
	}

	if !strings.HasSuffix(paket.Manifest.Belege[0].Ausschnitt,
		service.Kuerzungszeichen) {
		t.Fatalf("der Ausschnitt endet mit %q",
			paket.Manifest.Belege[0].Ausschnitt[len(paket.Manifest.Belege[0].Ausschnitt)-10:])
	}

	// **Und das Budget hält.**
	if paket.Manifest.Verbraucht > paket.Manifest.Verfuegbar() {
		t.Errorf("verbraucht sind %d von %d", paket.Manifest.Verbraucht,
			paket.Manifest.Verfuegbar())
	}
}

// **Wo kein Platz für einen ganzen Satz bleibt, kommt ein Beleg weniger.**
//
// **Und kein unbrauchbarer**: ein Ausschnitt von zwanzig Zeichen trägt keinen
// Satz, und ein Beleg ohne Satz belegt nichts.
func TestOhnePlatzFuerEinenSatzKommtEinBelegWeniger(t *testing.T) {
	t.Parallel()

	wunsch := einKontextwunsch()
	wunsch.Kontextgrenze = 130
	wunsch.Ausgabereserve = 100

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand: einRegelstand(),
		treffer: []domain.Retrievaltreffer{
			einRegeltreffer("702.9b", strings.Repeat("Lang. ", 30), 1),
			einRegeltreffer("509.1a", strings.Repeat("Auch lang. ", 30), 2),
		},
	}).Zusammenstellen(t.Context(), wunsch)
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	if len(paket.Manifest.Belege) != 1 {
		t.Fatalf("die Belege sind %d", len(paket.Manifest.Belege))
	}
}

// **Die Zahl der Belege lässt sich begrenzen.**
func TestDieZahlDerBelegeLaesstSichBegrenzen(t *testing.T) {
	t.Parallel()

	wunsch := einKontextwunsch()
	wunsch.HoechstensBelege = 1

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand: einRegelstand(),
		treffer: []domain.Retrievaltreffer{
			einRegeltreffer("702.9b", "Ein Text.", 1),
			einRegeltreffer("509.1a", "Noch einer.", 2),
		},
	}).Zusammenstellen(t.Context(), wunsch)
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	if len(paket.Manifest.Belege) != 1 {
		t.Fatalf("die Belege sind %d", len(paket.Manifest.Belege))
	}
}

// **Ein Auftrag ohne Aufgabe oder Zielmodell wird abgewiesen.**
func TestEinAuftragOhneAufgabeOderZielWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, aendern := range map[string]func(*service.Kontextwunsch){
		"ohne Aufgabe":  func(w *service.Kontextwunsch) { w.Aufgabe = "" },
		"ohne Modell":   func(w *service.Kontextwunsch) { w.Zielmodell = "" },
		"ohne Anbieter": func(w *service.Kontextwunsch) { w.Zielanbieter = "" },
		"ohne Fenster":  func(w *service.Kontextwunsch) { w.Kontextgrenze = 0 },
		"Reserve zu groß": func(w *service.Kontextwunsch) {
			w.Ausgabereserve = w.Kontextgrenze
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wunsch := einKontextwunsch()
			aendern(&wunsch)

			if _, fehler := einKontextdienst(&recherchequellendoppel{
				stand: einRegelstand(),
			}).Zusammenstellen(t.Context(), wunsch); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **Ein Fehler der Recherche kommt als Fehler zurück.**
func TestEinFehlerDerRechercheKommtAlsFehlerZurueck(t *testing.T) {
	t.Parallel()

	if _, fehler := einKontextdienst(&recherchequellendoppel{
		stand:  einRegelstand(),
		fehler: domain.MitGrund(domain.ErrConflict, nil, "der Bestand ist zu"),
	}).Zusammenstellen(t.Context(), einKontextwunsch()); fehler == nil {
		t.Fatal("der Recherchefehler kam als Paket zurück")
	}
}

// **Ein unlesbares Profil hält das Paket auf.**
func TestEinUnlesbaresProfilHaeltDasPaketAuf(t *testing.T) {
	t.Parallel()

	dienst := service.NeuerKontextdienst(
		einRecherchedienst(&recherchequellendoppel{stand: einRegelstand()}, nil),
		profildoppel{fehler: domain.MitGrund(domain.ErrValidation, nil,
			"kein Profil im Kontext")})

	if _, fehler := dienst.Zusammenstellen(
		t.Context(), einKontextwunsch()); fehler == nil {
		t.Fatal("das Paket kam ohne Profil durch")
	}
}

// **Ein Paket ohne Treffer ist eines ohne Belege — und kein Fehler.**
//
// **Das Issue trennt es**: „Ohne ausreichende Quellen bleibt Regelanalyse
// »nicht hinreichend belegt« **oder wird abgewiesen**." Das Abweisen ist die
// Sache des Aufrufers; hier steht, was da ist.
func TestEinPaketOhneTrefferIstEinesOhneBelege(t *testing.T) {
	t.Parallel()

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand: einRegelstand(),
	}).Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	if len(paket.Manifest.Belege) != 0 || paket.Kontext != "" {
		t.Fatalf("das Paket ist %+v", paket.Manifest.Belege)
	}

	// **Und das Ergebnis sagt, warum** — die Auskunft geht nicht verloren.
	if paket.Ergebnis.Leere != domain.LeereNichtsGefunden {
		t.Errorf("die Leere ist %q", paket.Ergebnis.Leere)
	}
}

// **Die gewürfelte Kennung ist jedes Mal eine andere.**
func TestDieGewuerfelteKennungIstJedesMalEineAndere(t *testing.T) {
	t.Parallel()

	dienst := service.NeuerKontextdienst(
		einRecherchedienst(&recherchequellendoppel{stand: einRegelstand()}, nil),
		profildoppel{}).MitUhr(dieRechercheuhr)

	erstes, fehler := dienst.Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	zweites, fehler := dienst.Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	if erstes.Manifest.PaketID == zweites.Manifest.PaketID {
		t.Fatalf("beide heißen %q", erstes.Manifest.PaketID)
	}

	if !strings.HasPrefix(erstes.Manifest.PaketID, "kp-") {
		t.Errorf("die Kennung ist %q", erstes.Manifest.PaketID)
	}
}
