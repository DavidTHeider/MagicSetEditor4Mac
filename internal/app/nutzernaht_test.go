// **Nutzer und Profile über den Flächenweg** — Werkstück 1 von Phase 4.12
// (#1351).
//
// **Über einen echten Bestand**, denn genau das ist die Zusage dieser Naht:
// dass die drei Flächen ohne Fenster dieselbe Anwendung bekommen (§16).
package app_test

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der frische Bestand führt einen Nutzer mit einem Profil.**
func TestDerNutzerflaechenwegFuehrtDenErstenNutzer(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	antwort := flaeche.Nutzer(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Nutzer) != 1 || len(antwort.Nutzer[0].Profile) != 1 {
		t.Fatalf("%+v", antwort.Nutzer)
	}
}

// **Und er legt an — Nutzer wie Profil.**
func TestDerNutzerflaechenwegLegtAn(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	angelegt := flaeche.NutzerAnlegen(probekontext(t), "Zweite", "")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	profil := flaeche.ProfilAnlegen(probekontext(t), angelegt.Nutzer.ID, "Turnier")
	if profil.Fehler != nil {
		t.Fatalf("%+v", profil.Fehler)
	}

	if profil.Profil.Name != "Turnier" {
		t.Errorf("%+v", profil.Profil)
	}
}

// **Und er wechselt** — Werkstück 3 (#1351).
func TestDerNutzerflaechenwegWechseltDasProfil(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	vorher := flaeche.GezeigtesProfil(probekontext(t))
	if vorher.Fehler != nil {
		t.Fatalf("%+v", vorher.Fehler)
	}

	// **Ein zweites Profil DESSELBEN Nutzers** — Werkstück 11 (#1351): unter
	// den eigenen zu wechseln braucht keine Freigabe, in ein fremdes schon.
	// Dieser Fall meint den Weg über die Fläche und nicht die Erlaubnis.
	angelegt := flaeche.ProfilAnlegen(probekontext(t),
		vorher.Profil.NutzerID, "Turnier")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	zweites := angelegt.Profil.ID

	gewechselt := flaeche.ProfilWechseln(probekontext(t), zweites, "")
	if gewechselt.Fehler != nil {
		t.Fatalf("%+v", gewechselt.Fehler)
	}

	if gewechselt.Profil.ID != zweites {
		t.Errorf("gewechselt auf %+v, erwartet war %s", gewechselt.Profil, zweites)
	}

	// **Und der Zeiger steht danach dort** — die Fläche ohne Fenster hat
	// keinen eigenen Stand, sie liest ihn jedes Mal.
	nachher := flaeche.GezeigtesProfil(probekontext(t))
	if nachher.Fehler != nil || nachher.Profil.ID != zweites {
		t.Errorf("%+v / %+v", nachher.Profil, nachher.Fehler)
	}

	if vorher.Profil.ID == zweites {
		t.Error("das erste Profil war schon das zweite — der Lauf sagt nichts")
	}
}

// **Trägt der Ruf ein Profil, nennt die Antwort dieses — nicht den Zeiger.**
//
// # Warum das die Zusage einlöst, die `lesen.profil` gibt
//
// **„Jede andere Antwort dieses Servers … gilt für DIESES Profil und kein
// anderes."** So steht es in der Werkzeugbeschreibung, und mit einem Token
// stimmte es nicht: `app.Tokenwache` hängt das Profil des Tokens an den
// Kontext, jede Abfrage des Bestands liest von dort — nur diese Antwort kam
// vom Zeiger. Ein Agent bekam den Namen eines Profils, in dem seine eigenen
// Sets nicht stehen (A14, #1453).
func TestDerNutzerflaechenwegNenntDasProfilDesRufs(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	zeiger := flaeche.GezeigtesProfil(probekontext(t))
	if zeiger.Fehler != nil {
		t.Fatalf("%+v", zeiger.Fehler)
	}

	angelegt := flaeche.ProfilAnlegen(probekontext(t),
		zeiger.Profil.NutzerID, "Des Tokens")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	zweites, fehler := strconv.ParseInt(angelegt.Profil.ID, 10, 64)
	if fehler != nil {
		t.Fatalf("die Kennung lesen: %v", fehler)
	}

	// **Der Zeiger bleibt, wo er ist** — gewechselt wird nicht, das Profil
	// steht nur am Ruf. Genau so kommt es aus einem Token.
	imRuf := flaeche.GezeigtesProfil(
		storage.MitProfil(probekontext(t), zweites))
	if imRuf.Fehler != nil {
		t.Fatalf("%+v", imRuf.Fehler)
	}

	if imRuf.Profil.ID != angelegt.Profil.ID {
		t.Errorf("die Antwort nennt %s statt %s", imRuf.Profil.ID, angelegt.Profil.ID)
	}

	if nachher := flaeche.GezeigtesProfil(probekontext(t)); nachher.Profil.ID != zeiger.Profil.ID {
		t.Errorf("der Zeiger wanderte auf %s", nachher.Profil.ID)
	}
}

// **Ein geschlossener Bestand meldet sich an beiden Wegen.**
func TestDieProfilwegeDerFlaecheMeldenEinenGeschlossenenBestand(t *testing.T) {
	db := offeneProbedatenbank(t)
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	flaeche := app.NutzerflaecheUeber(db)

	if antwort := flaeche.GezeigtesProfil(probekontext(t)); antwort.Fehler == nil {
		t.Error("das Lesen kam durch")
	}

	if antwort := flaeche.ProfilWechseln(probekontext(t), "1", ""); antwort.Fehler == nil {
		t.Error("der Wechsel kam durch")
	}
}

// **Und die Fläche ohne Fenster führt die Konten** — Werkstück 4 (#1351).
//
// **Ohne den Anmeldefluss**: der öffnet einen Browser und wartet auf einen
// Menschen. Was hier zählt, ist die **Naht** — dass die drei Flächen ohne
// Fenster dieselben Antworten geben wie das Fenster.
func TestDerNutzerflaechenwegFuehrtDieKonten(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	anbieter := flaeche.Anbieter(probekontext(t))
	if anbieter.Fehler != nil || len(anbieter.Anbieter) != 4 {
		t.Fatalf("%+v / %+v", anbieter.Anbieter, anbieter.Fehler)
	}

	konten := flaeche.Anbieterkonten(probekontext(t), "1")
	if konten.Fehler != nil {
		t.Fatalf("%+v", konten.Fehler)
	}

	if len(konten.Konten) != 0 {
		t.Errorf("Konten %+v", konten.Konten)
	}

	// **Ein Konto, das es nicht gibt, zu lösen ist ein Nein.**
	if geloest := flaeche.KontoLoesen(
		probekontext(t), "1", "github"); geloest.Fehler == nil {
		t.Error("das fehlende Konto kam durch")
	}
}

// **Und der Anmeldeweg der Fläche meldet, was fehlt.**
//
// **Ohne Kennung geht er nicht los** — und damit auch kein Browser auf. Das
// ist die Zusage, die ein Lauf hier prüfen kann, ohne einen Menschen zu
// brauchen.
func TestDerNutzerflaechenwegMeldetDieFehlendeKennung(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	antwort := flaeche.Anmelden(probekontext(t), "github")
	if antwort.Fehler == nil {
		t.Fatal("die Anmeldung ohne Kennung kam durch")
	}

	if antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("Code %q", antwort.Fehler.Code)
	}
}

// **Und die Fläche ohne Fenster löscht ein Profil, ohne abzustürzen.**
//
// # Warum dieser Lauf eigens dasteht
//
// **Weil `ProfilLoeschen` danach den Profilzeiger nachlädt** — und die drei
// Flächen ohne Fenster bauen ihre Anwendung über `NeueApp("")`: ihr `db` ist
// **nil**, ihr Profil steht im Wurzelkontext. Ohne die Wache in `profilLaden`
// stürzt jede von ihnen ab.
//
// **Gefunden hat es der Flächenlauf**, nicht ein Lauf dieses Pakets — die
// Bindungen am Fenster haben immer einen Bestand. Zwei grüne Seiten einer
// Grenze sind kein Nachweis für die Grenze; dieser Lauf ist die Grenze.
func TestDerNutzerflaechenwegLoeschtOhneAbsturz(t *testing.T) {
	db := offeneProbedatenbank(t)
	flaeche := app.NutzerflaecheUeber(db)

	angelegt := flaeche.ProfilAnlegen(probekontext(t), "1", "Turnier")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if geloescht := flaeche.ProfilLoeschen(
		probekontext(t), angelegt.Profil.ID, "Turnier"); geloescht.Fehler != nil {
		t.Fatalf("%+v", geloescht.Fehler)
	}
}

// **Und die Fläche ohne Fenster ändert und löscht.**
func TestDerNutzerflaechenwegAendertUndLoescht(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	geaendert := flaeche.NutzerAendern(probekontext(t), app.NutzerangabenDTO{
		ID: "1", Name: "Umbenannt", Bild: "", Fassung: "1",
	})
	if geaendert.Fehler != nil {
		t.Fatalf("%+v", geaendert.Fehler)
	}

	if geaendert.Nutzer.Name != "Umbenannt" {
		t.Errorf("Name %q", geaendert.Nutzer.Name)
	}

	profil := flaeche.ProfilAendern(probekontext(t), app.ProfilangabenDTO{
		ID: "1", Name: "Legacy", Fassung: "1",
	})
	if profil.Fehler != nil || profil.Profil.Name != "Legacy" {
		t.Fatalf("%+v / %+v", profil.Profil, profil.Fehler)
	}

	// **Der letzte Nutzer bleibt** — auch hier.
	if geloescht := flaeche.NutzerLoeschen(
		probekontext(t), "1", "Umbenannt"); geloescht.Fehler == nil {
		t.Error("der letzte Nutzer ließ sich löschen")
	}
}

// **Und er legt Konten an — eigene Kennung wie anonym.**
//
// **Über einen echten Bestand**, denn das ist die Zusage dieser Naht: die drei
// Flächen ohne Fenster bekommen dieselbe Anwendung (§16). Ein Doppel prüfte
// die Weiterleitung und nicht den Weg.
func TestDerNutzerflaechenwegLegtKontenAn(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	eingetragen := flaeche.KontoEintragen(probekontext(t),
		"wer@example.org", "einlangesgeheimnis", "")
	if eingetragen.Fehler != nil {
		t.Fatalf("%+v", eingetragen.Fehler)
	}

	herein := flaeche.KontoAnmelden(probekontext(t),
		"wer@example.org", "einlangesgeheimnis", "")
	if herein.Fehler != nil {
		t.Fatalf("%+v", herein.Fehler)
	}

	if herein.Nutzer.ID != eingetragen.Nutzer.ID {
		t.Errorf("herein kam %q, eingetragen war %q",
			herein.Nutzer.ID, eingetragen.Nutzer.ID)
	}

	anonym := flaeche.KontoAnonym(probekontext(t))
	if anonym.Fehler != nil {
		t.Fatalf("%+v", anonym.Fehler)
	}

	if !anonym.Nutzer.Anonym {
		t.Errorf("der Vermerk fehlt: %+v", anonym.Nutzer)
	}
}
