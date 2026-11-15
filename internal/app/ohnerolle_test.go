//go:build darwin

// **Ohne Rolle wird in einem fremden Profil nichts geschrieben** — ID-03
// (#1702),
// der native Befund F01 aus der Sichtprüfung zu 0.6.0.
//
// # Der Befund
//
// **In 0.5.0 ließ sich in einem fremden Profil ohne jede Freigabe ein Deck
// anlegen**, während das Banner „nur lesen" sagte — dazu Setangaben, ein
// Kartenname, Profiltext, Lagerort, ein Turnier, ein Set, Ersetzen,
// Schlüsselwörter, Symbol und Stil. Abgewiesen wurden nur die Wege, die eine
// Datei schreiben: sie fragen den Akteur (0078).
//
// # Die Ursache — zwei Antworten auf dieselbe Frage
//
// **Hineingekommen** war der Mensch über `Zutritt`: ein Eigner ohne zweiten
// Faktor lässt jeden herein (`service/wechselerlaubnis.go` sagt es in seinem
// Kopf). Wer so hineinkommt, **bleibt er selbst** und hat dort keine Rolle —
// `Freigabenverwaltung.Rolle` gibt dann `""` und keinen Fehler.
//
// **Das Banner** fragt `Akteur.Darf(WirkungSchreiben)` — ohne Rolle: nein,
// also „nur lesen". **Die Verbindung** fragte, ob die Rolle genau `leser` ist
// — ohne Rolle: auch nein, also blieb sie **schreibend**.
//
// # Was er zeigt
//
// Jeder Schreibweg aus dem Befund, der ohne Paket auskommt, wird abgewiesen —
// und **an der Wirkung** geprüft (AK-04): keine neue Zeile, keine geänderte,
// nicht nur ein Fehlertext. Und im eigenen Profil geht Schreiben weiter
// (S02.NAT-03).

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// fremdesProfilOhneRolle ist, was der Lauf braucht, um die Wirkung zu prüfen.
type fremdesProfilOhneRolle struct {
	anwendung *app.App
	eigenes   string
	fremdes   *app.ProfilDTO
	set       app.OeffnenAntwort
}

// alsFremderOhneRolle geht in das Profil eines zweiten Nutzers, der keinen
// zweiten Faktor hat, ein Set besitzt und niemandem eine Freigabe gegeben hat.
func alsFremderOhneRolle(t *testing.T) fremdesProfilOhneRolle {
	t.Helper()

	anwendung, _ := gestarteteAnwendung(t)
	ctx := t.Context()

	ich := anwendung.WerArbeitet(ctx)
	eigenes := anwendung.GezeigtesProfil(ctx)

	if ich.Fehler != nil || eigenes.Fehler != nil {
		t.Fatalf("wer und wo: %+v %+v", ich.Fehler, eigenes.Fehler)
	}

	bob := anwendung.NutzerAnlegen(ctx, "Audit Bob", "")
	if bob.Fehler != nil {
		t.Fatalf("anlegen: %+v", bob.Fehler)
	}

	// **Bob legt selbst ein Set an** — damit es im fremden Profil etwas gibt,
	// das sich ändern ließe.
	if wechsel := anwendung.NutzerWechseln(ctx, bob.Nutzer.ID, ""); wechsel.Fehler != nil {
		t.Fatalf("zu Bob: %+v", wechsel.Fehler)
	}

	set := anwendung.SetOeffnen(ctx, probesetVerzeichnis(t))
	if set.Fehler != nil {
		t.Fatalf("Bobs Set: %+v", set.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(ctx, ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	// **Der Weg aus dem Befund**: in Bobs Profil, ohne Nachweis und ohne
	// Freigabe — Bob hat keinen Faktor, also lässt `Zutritt` herein.
	if hinein := anwendung.ProfilWechseln(ctx, bob.Nutzer.Profile[0].ID, ""); hinein.Fehler != nil {
		t.Fatalf("hinein: %+v", hinein.Fehler)
	}

	return fremdesProfilOhneRolle{
		anwendung: anwendung, eigenes: eigenes.Profil.ID,
		fremdes: profilAusUebersicht(t, anwendung, bob.Nutzer.Profile[0].ID), set: set,
	}
}

func TestOhneRolleWirdImFremdenProfilNichtsGeschrieben(t *testing.T) {
	lage := alsFremderOhneRolle(t)
	anwendung, ctx := lage.anwendung, t.Context()

	// **Das Banner sagt es schon** — und die Verbindung muss dasselbe sagen.
	kontext := anwendung.Arbeitskontext(ctx)
	if kontext.Fehler != nil || !kontext.NurLesend || kontext.ProfilID != lage.fremdes.ID {
		t.Fatalf("Arbeitskontext %+v, erwartet: nur lesend im fremden Profil", kontext)
	}

	zeilen := anwendung.Kartenzeilen(ctx, lage.set.SetID)
	if zeilen.Fehler != nil || len(zeilen.Zeilen) != 1 {
		t.Fatalf("Bobs Karten: %+v", zeilen)
	}

	faelle := map[string]*app.Fehler{
		"Deck anlegen":    anwendung.DeckAnlegen(ctx, "Audit Unberechtigt", "magic", "").Fehler,
		"Set anlegen":     anwendung.SetAnlegen(ctx, "Audit Unberechtigt", "magic", "magic-m15").Fehler,
		"Turnier anlegen": anwendung.TurnierAnlegen(ctx, "Audit Unberechtigt", "magic", "", false, "").Fehler,
		"Profiltext": anwendung.ProfilAendern(ctx, app.ProfilangabenDTO{
			ID: lage.fremdes.ID, Name: lage.fremdes.Name, Kurzebeschreibung: "Unberechtigt",
			Fassung: lage.fremdes.Fassung,
		}).Fehler,
		"Kartenname": anwendung.FeldSetzen(probekontext(t), lage.set.SetID, zeilen.Zeilen[0].ID,
			"name", "Unberechtigt").Fehler,
		"Ersetzen": anwendung.ErsetzungAusfuehren(ctx, lage.set.SetID,
			app.ErsetzungsfrageDTO{Text: "Feuer", Ersatz: "Wasser"}).Fehler,
		"Schlüsselwort": anwendung.SchluesselwortAnlegen(ctx, lage.set.SetID).Fehler,
	}

	for name, fehler := range faelle {
		if fehler == nil {
			t.Errorf("%s: ohne Rolle geschrieben", name)

			continue
		}

		if fehler.Code != app.CodeNurLesend {
			t.Errorf("%s: Code %q, erwartet %q — %s", name, fehler.Code, app.CodeNurLesend, fehler.Meldung)
		}
	}

	pruefeNichtsGeaendert(t, lage)

	// **Und im eigenen Profil geht Schreiben weiter** (S02.NAT-03) — sonst
	// bewiese der Lauf eine Anwendung, die gar nichts mehr schreibt.
	if zurueck := anwendung.ProfilWechseln(ctx, lage.eigenes, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	if eigenesDeck := anwendung.DeckAnlegen(ctx, "Eigenes", "magic", ""); eigenesDeck.Fehler != nil {
		t.Errorf("im eigenen Profil: %+v", eigenesDeck.Fehler)
	}
}

// pruefeNichtsGeaendert sagt zu, dass im fremden Profil nichts hinzukam und
// nichts sich änderte — **an der Wirkung** (AK-04).
func pruefeNichtsGeaendert(t *testing.T, lage fremdesProfilOhneRolle) {
	t.Helper()

	anwendung, ctx := lage.anwendung, t.Context()

	if decks := anwendung.Decks(ctx, ""); decks.Fehler != nil || len(decks.Decks) != 0 {
		t.Errorf("Decks im fremden Profil: %+v", decks)
	}

	if sets := anwendung.Sets(ctx); sets.Fehler != nil || len(sets.Sets) != 1 {
		t.Errorf("Sets im fremden Profil, erwartet nur Bobs eines: %+v", sets)
	}

	if turniere := anwendung.Turniere(ctx, ""); turniere.Fehler != nil || len(turniere.Turniere) != 0 {
		t.Errorf("Turniere im fremden Profil: %+v", turniere)
	}

	if profil := profilAusUebersicht(t, anwendung, lage.fremdes.ID); profil.Kurzebeschreibung != lage.fremdes.Kurzebeschreibung ||
		profil.Fassung != lage.fremdes.Fassung {
		t.Errorf("Bobs Profil geändert: %+v, vorher %+v", profil, lage.fremdes)
	}

	if zeilen := anwendung.Kartenzeilen(ctx, lage.set.SetID); zeilen.Fehler != nil ||
		len(zeilen.Zeilen) != 1 || zeilen.Zeilen[0].Name != "Feuerball" {
		t.Errorf("Bobs Karte: %+v", zeilen)
	}
}
