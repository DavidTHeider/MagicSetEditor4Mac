package app_test

// **Die Bindungen der Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Geprüft wird die Übersetzung** — was `app` tut und sonst nichts: die
// Anfrage in die Gestalt des Dienstes, der Treffer ins DTO, der Grund einer
// leeren Antwort ans Fenster. Ob die Recherche richtig sucht, steht in
// `internal/service`.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitRecherche baut eine Anwendung über einem Doppel.
func mitRecherche(doppel *recherchedoppel) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitRecherche(doppel)

	return anwendung
}

// einProbetreffer ist ein vollständig gefüllter.
func einProbetreffer() domain.Recherchetreffer {
	return domain.Recherchetreffer{
		Rang: 1, Kennung: "oracle:abc:0", Quelle: domain.QuelleOracle,
		Kennungssystem: domain.SystemScryfall, Fachkennung: "abc",
		Semantik: domain.SemantikKonzept, EigeneKarteID: 0,
		Titel: "Lightning Bolt", Ausschnitt: "3 damage",
		Text:    "Lightning Bolt deals 3 damage to any target.",
		Sprache: "en", Regelnummer: "", Urheber: "",
		Abschnitt: 0, Von: 0, Bis: 42,
		Bezugsadresse: "https://scryfall.com/card/lea/161",
		Quellrevision: "2026-09-01", Quelldatum: "2026-09-01",
		Abgerufen: "2026-09-02", Weitere: []string{"oracle:abc:1"},
	}
}

// **Ein Treffer kommt vollständig ans Fenster — samt seiner Herkunft.**
//
// **Drei Angaben zur Herkunft und nicht eine**: das Abrufdatum ist kein
// Gültigkeitsdatum, und die Revision ist keins von beiden.
func TestEinRecherchetrefferKommtVollstaendigAnsFenster(t *testing.T) {
	t.Parallel()

	antwort := mitRecherche(&recherchedoppel{
		ergebnis: domain.Rechercheergebnis{
			Treffer: []domain.Recherchetreffer{einProbetreffer()},
			Modus:   "hybrid", Modell: "all-minilm", Modellrevision: "aaaa",
			Suchkonfiguration: service.Rangfusionsname,
			Korpusstaende: []domain.Korpusstand{{
				Quelle: domain.QuelleOracle, Revision: "2026-09-01",
				Datum: "2026-09-01",
			}},
			Fortsetzung: "weiter",
		},
	}).Recherchieren(t.Context(), string(domain.FallKarten), "bolt", 5,
		app.RecherchefilterDTO{}, "")

	if antwort.Fehler != nil {
		t.Fatalf("Recherchieren: %+v", antwort.Fehler)
	}

	if len(antwort.Treffer) != 1 {
		t.Fatalf("die Treffer sind %+v", antwort.Treffer)
	}

	treffer := antwort.Treffer[0]
	if treffer.Kennung != "oracle:abc:0" || treffer.Rang != 1 ||
		treffer.Kennungssystem != string(domain.SystemScryfall) ||
		treffer.Fachkennung != "abc" || treffer.Bis != 42 ||
		treffer.Quellrevision != "2026-09-01" ||
		treffer.Quelldatum != "2026-09-01" ||
		treffer.Abgerufen != "2026-09-02" ||
		treffer.Bezugsadresse == "" || len(treffer.Weitere) != 1 {
		t.Errorf("der Treffer ist %+v", treffer)
	}

	// **Ein Treffer ohne weitere Abschnitte bekommt eine leere Liste** und
	// keinen Nullwert — sonst käme er im Fenster als „gibt es nicht" an.
	ohne := alsRecherchetrefferDTOProbe(t)
	if ohne.Weitere == nil {
		t.Errorf("die Liste ist null: %+v", ohne)
	}

	if antwort.Modus != "hybrid" || antwort.Modell != "all-minilm" ||
		antwort.Suchkonfiguration != service.Rangfusionsname ||
		antwort.Fortsetzung != "weiter" ||
		len(antwort.Korpusstaende) != 1 {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// alsRecherchetrefferDTOProbe gibt den Treffer eines Ergebnisses ohne
// gebündelte Abschnitte.
func alsRecherchetrefferDTOProbe(t *testing.T) app.RecherchetrefferDTO {
	t.Helper()

	nackt := einProbetreffer()
	nackt.Weitere = nil

	antwort := mitRecherche(&recherchedoppel{
		ergebnis: domain.Rechercheergebnis{
			Treffer: []domain.Recherchetreffer{nackt},
		},
	}).Recherchieren(t.Context(), string(domain.FallKarten), "bolt", 5,
		app.RecherchefilterDTO{}, "")

	if len(antwort.Treffer) != 1 {
		t.Fatalf("die Treffer sind %+v", antwort.Treffer)
	}

	return antwort.Treffer[0]
}

// **Der Grund einer leeren Antwort kommt mit.**
//
// **Sonst stünde das Fenster vor „nichts gefunden"** und wüsste nicht, ob es
// nichts gibt oder ob kein Index da ist — und nur eines davon lässt sich
// beheben.
func TestDerGrundEinerLeerenAntwortKommtMit(t *testing.T) {
	t.Parallel()

	antwort := mitRecherche(&recherchedoppel{
		ergebnis: domain.Rechercheergebnis{
			Modus: "ohne-index", Leere: domain.LeereOhneIndex,
			Grund:            "es ist kein Suchindex gebaut",
			Einschraenkungen: []string{"ohne Vektoren gesucht"},
		},
	}).Recherchieren(t.Context(), string(domain.FallRegeln), "trample", 5,
		app.RecherchefilterDTO{}, "")

	if antwort.Leere != string(domain.LeereOhneIndex) ||
		antwort.Grund == "" || len(antwort.Einschraenkungen) != 1 {
		t.Fatalf("die Antwort ist %+v", antwort)
	}
}

// **Der Filter kommt vollständig an — und wird nicht zur Frage.**
func TestDerRecherchefilterKommtVollstaendigAn(t *testing.T) {
	t.Parallel()

	doppel := &recherchedoppel{}

	mitRecherche(doppel).Recherchieren(t.Context(), string(domain.FallRulings),
		"bolt", 7, app.RecherchefilterDTO{
			Sprache: "de", Korpusstand: "2026-08-01",
		}, "zeiger")

	if doppel.gesehen.Fall != domain.FallRulings ||
		doppel.gesehen.Frage != "bolt" || doppel.gesehen.Grenze != 7 ||
		doppel.gesehen.Sprache != "de" ||
		doppel.gesehen.Korpusstand != "2026-08-01" ||
		doppel.fortsetzung != "zeiger" {
		t.Fatalf("die Anfrage ist %+v (Zeiger %q)", doppel.gesehen,
			doppel.fortsetzung)
	}
}

// **Die Ähnlichkeitsfrage trägt Karte und Stand.**
//
// **Beides, nicht nur die Karte**: eine Karte, die sich seit der Anfrage
// geändert hat, ist eine andere Frage.
func TestDieAehnlichkeitsfrageTraegtKarteUndStand(t *testing.T) {
	t.Parallel()

	doppel := &recherchedoppel{}

	mitRecherche(doppel).AehnlicheKarten(t.Context(), 42, "abcdef", 5,
		app.RecherchefilterDTO{Sprache: "en"}, "")

	if doppel.gesehen.Fall != domain.FallAehnlich ||
		doppel.gesehen.AusgangskarteID != 42 ||
		doppel.gesehen.Ausgangsstand != "abcdef" ||
		doppel.gesehen.Sprache != "en" {
		t.Fatalf("die Anfrage ist %+v", doppel.gesehen)
	}
}

// **Ohne Bestand antwortet die Recherche mit einem Fehler und leeren Listen.**
func TestOhneBestandGibtDieRechercheLeereListen(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Recherchieren(t.Context(),
		string(domain.FallKarten), "bolt", 5, app.RecherchefilterDTO{}, "")

	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam kein Fehler")
	}

	if antwort.Treffer == nil || antwort.Korpusstaende == nil ||
		antwort.Einschraenkungen == nil {
		t.Fatalf("eine Liste ist null: %+v", antwort)
	}
}

// **Ein Fehler des Dienstes kommt übersetzt heraus.**
func TestEinRecherchefehlerKommtUebersetztHeraus(t *testing.T) {
	t.Parallel()

	antwort := mitRecherche(&recherchedoppel{
		suchfehler: domain.MitGrund(domain.ErrConflict, errors.New("innen"),
			"die Ausgangskarte hat sich geändert"),
	}).AehnlicheKarten(t.Context(), 1, "alt", 5, app.RecherchefilterDTO{}, "")

	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeKonflikt {
		t.Fatalf("der Fehler ist %+v", antwort.Fehler)
	}

	// **Der Grund bleibt drinnen** (§8): der Satz ist für Menschen.
	if antwort.Fehler.Meldung == "innen" {
		t.Errorf("der innere Grund steht in der Meldung: %+v", antwort.Fehler)
	}
}

// einProbepaket ist ein vollständiges Kontextpaket.
func einProbepaket() service.Kontextpaket {
	return service.Kontextpaket{
		Manifest: domain.Kontextmanifest{
			PaketID: "p1", Aufgabe: "schreib mir einen Regeltext",
			Zielrevision: "abc", Zielanbieter: "anthropic",
			Zielmodell: "claude", Zaehlweise: service.Zaehlweise,
			Kontextgrenze: 8000, Ausgabereserve: 1000, Verbraucht: 120,
			Kontextpolicy: domain.Kontextpolicyfassung,
			Korpusstaende: []domain.Korpusstand{{
				Quelle: domain.QuelleRegelwerk, Revision: "2026-09-01",
			}},
			Belege: []domain.Beleg{{
				Kennung: "B1", Quellkennung: "regeln:702.19:0",
				Quelle: domain.QuelleRegelwerk,
				// **Eine Regelstelle hat kein Kennungssystem** — sie ist
				// keine Karte, und ihre Nummer ist ihre Identität.
				Fachkennung: "702.19",
				Titel:       "Trample", Ausschnitt: "702.19a Trample is a static …",
				Inhaltshash: "deadbeef", Sprache: "en", Regelnummer: "702.19",
				Bezugsadresse: "https://magic.wizards.com/rules",
				Quellrevision: "2026-09-01", Quelldatum: "2026-09-01",
				Abgerufen: "2026-09-02",
			}},
			Laeuftab: time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC),
		},
		Kontext:  "[B1] 702.19a Trample is a static …",
		Ergebnis: domain.Rechercheergebnis{Modus: "hybrid"},
	}
}

// **Das Paket kommt vollständig ans Fenster — und der Kontext nicht.**
//
// **Das ist die Zusage des Issues**: „Kurzmanifest für A11 von
// schutzbedürftigem Textsnapshot trennen". Was zurückkommt, sind Manifest und
// Freigabe.
func TestDasKontextpaketKommtOhneSeinenKontextAnsFenster(t *testing.T) {
	t.Parallel()

	antwort := mitRecherche(&recherchedoppel{paket: einProbepaket()}).
		KontextZusammenstellen(t.Context(), "schreib mir einen Regeltext",
			"abc", string(domain.FallRegeln), "trample", 5,
			"anthropic", "claude", 8000, 1000, 0, nil)

	if antwort.Fehler != nil {
		t.Fatalf("KontextZusammenstellen: %+v", antwort.Fehler)
	}

	if antwort.PaketID != "p1" || antwort.Zaehlweise != service.Zaehlweise ||
		antwort.Verbraucht != 120 ||
		antwort.Kontextpolicy != domain.Kontextpolicyfassung ||
		antwort.Laeuftab != "2026-09-12T10:00:00Z" {
		t.Fatalf("die Antwort ist %+v", antwort)
	}

	if len(antwort.Belege) != 1 || antwort.Belege[0].Kennung != "B1" ||
		antwort.Belege[0].Ausschnitt == "" ||
		antwort.Belege[0].Inhaltshash != "deadbeef" {
		t.Errorf("die Belege sind %+v", antwort.Belege)
	}

	// **Der Kontext selbst hat keine Stelle im DTO** — wäre eine da, ginge
	// der Textschnappschuss mit jeder Antwort hinaus.
	if antwort.Ergebnis.Modus != "hybrid" {
		t.Errorf("das Ergebnis fehlt: %+v", antwort.Ergebnis)
	}
}

// **Die Freigabe sagt den Ort — und ein fremder Anbieter rechnet nie lokal.**
func TestEinFremderAnbieterRechnetNichtLokal(t *testing.T) {
	t.Parallel()

	antwort := mitRecherche(&recherchedoppel{paket: einProbepaket()}).
		KontextZusammenstellen(t.Context(), "a", "abc",
			string(domain.FallRegeln), "trample", 5,
			"anthropic", "claude", 8000, 1000, 0, nil)

	if antwort.Freigabe.Ort != string(domain.KIOrtCloud) ||
		antwort.Freigabe.Zielanbieter != "anthropic" ||
		antwort.Freigabe.Belege != 1 || antwort.Freigabe.Satz == "" {
		t.Fatalf("die Freigabe ist %+v", antwort.Freigabe)
	}
}

// **Der Wunsch kommt vollständig am Dienst an.**
func TestDerKontextwunschKommtVollstaendigAn(t *testing.T) {
	t.Parallel()

	doppel := &recherchedoppel{paket: einProbepaket()}

	mitRecherche(doppel).KontextZusammenstellen(t.Context(), "aufgabe", "rev",
		string(domain.FallRulings), "frage", 9, "ollama", "gemma3", 4096,
		512, 6, nil)

	wunsch := doppel.gesehenWunsch
	if wunsch.Aufgabe != "aufgabe" || wunsch.Zielrevision != "rev" ||
		wunsch.Anfrage.Fall != domain.FallRulings ||
		wunsch.Anfrage.Frage != "frage" || wunsch.Anfrage.Grenze != 9 ||
		wunsch.Zielanbieter != "ollama" || wunsch.Zielmodell != "gemma3" ||
		wunsch.Kontextgrenze != 4096 || wunsch.Ausgabereserve != 512 ||
		wunsch.HoechstensBelege != 6 {
		t.Fatalf("der Wunsch ist %+v", wunsch)
	}
}

// **Ohne Bestand gibt das Kontextpaket leere Listen und keinen Nullwert.**
func TestOhneBestandGibtDasKontextpaketLeereListen(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KontextZusammenstellen(t.Context(), "a", "b",
		string(domain.FallKarten), "c", 5, "ollama", "m", 4096, 512, 0, nil)

	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam kein Fehler")
	}

	if antwort.Belege == nil || antwort.Korpusstaende == nil ||
		antwort.Freigabe.Kategorien == nil ||
		antwort.Ergebnis.Treffer == nil {
		t.Fatalf("eine Liste ist null: %+v", antwort)
	}
}

// **Ein Fehler beim Zusammenstellen kommt übersetzt heraus.**
func TestEinKontextfehlerKommtUebersetztHeraus(t *testing.T) {
	t.Parallel()

	antwort := mitRecherche(&recherchedoppel{
		paketfehler: domain.MitGrund(domain.ErrValidation, nil,
			"ohne Aufgabe gibt es kein Paket"),
	}).KontextZusammenstellen(t.Context(), "", "b", string(domain.FallKarten),
		"c", 5, "ollama", "m", 4096, 512, 0, nil)

	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeValidierung {
		t.Fatalf("der Fehler ist %+v", antwort.Fehler)
	}
}
