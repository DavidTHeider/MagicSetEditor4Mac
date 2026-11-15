package app_test

// **Nutzer und Profile ändern und löschen am Fenster** — Werkstück 5 von
// Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Nutzer lässt sich ändern und kommt mit seinen Profilen zurück.**
func TestEinNutzerLaesstSichAmFensterAendern(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	antwort := anwendung.NutzerAendern(t.Context(), app.NutzerangabenDTO{
		ID: "1", Name: "Umbenannt", Bild: "ich.png", Fassung: "1",
	})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Nutzer == nil || antwort.Nutzer.Name != "Umbenannt" {
		t.Fatalf("Nutzer %+v", antwort.Nutzer)
	}

	// **Mit Profilen** — eine leere Liste behauptete, er habe keine.
	if len(antwort.Nutzer.Profile) == 0 {
		t.Error("die Profile fehlen")
	}
}

// **Ein Profil lässt sich ändern — und behält seinen Nutzer.**
func TestEinProfilLaesstSichAmFensterAendern(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	antwort := anwendung.ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: "1", Name: "Turnier", Fassung: "1",
	})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Profil.NutzerID != "1" {
		t.Errorf("Nutzer %q — das Profil hat ihn gewechselt", antwort.Profil.NutzerID)
	}
}

// **Ein Profil, das es nicht gibt, ist „nicht gefunden".**
func TestEinFehlendesProfilAmFensterIstNichtGefunden(t *testing.T) {
	antwort := angelaufeneAnwendung(t).ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: "4711", Name: "X", Fassung: "1",
	})

	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", antwort.Fehler)
	}
}

// **Der letzte Nutzer und das letzte Profil bleiben.**
func TestDerLetzteNutzerUndDasLetzteProfilBleibenAmFenster(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	if antwort := anwendung.NutzerLoeschen(t.Context(), "1", "Ich"); antwort.Fehler == nil {
		t.Error("der letzte Nutzer ließ sich löschen")
	}

	if antwort := anwendung.ProfilLoeschen(
		t.Context(), "1", "Mein Profil"); antwort.Fehler == nil {
		t.Error("das letzte Profil ließ sich löschen")
	}
}

// **Ein zweites Profil lässt sich löschen — und der Zeiger wandert mit.**
//
// # Warum das hier geprüft wird und nicht nur im Fachkern
//
// **Weil das Fenster einen eigenen Stand hat** (`App.profil`). Der Fachkern
// stellt den Zeiger im Bestand um; zieht das Fenster nicht nach, arbeitet es
// weiter in einem Profil, das es nicht mehr gibt — und das fiele erst beim
// nächsten Start auf.
func TestEinZweitesProfilLaesstSichLoeschenUndDerZeigerWandert(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	angelegt := anwendung.ProfilAnlegen(t.Context(), "1", "Turnier")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if gewechselt := anwendung.ProfilWechseln(
		t.Context(), angelegt.Profil.ID, ""); gewechselt.Fehler != nil {
		t.Fatalf("%+v", gewechselt.Fehler)
	}

	if geloescht := anwendung.ProfilLoeschen(
		t.Context(), angelegt.Profil.ID, "Turnier"); geloescht.Fehler != nil {
		t.Fatalf("%+v", geloescht.Fehler)
	}

	// **Und das Fenster steht danach im anderen Profil.**
	gezeigt := anwendung.GezeigtesProfil(t.Context())
	if gezeigt.Fehler != nil {
		t.Fatalf("%+v", gezeigt.Fehler)
	}

	if gezeigt.Profil.ID == angelegt.Profil.ID {
		t.Errorf("der Zeiger steht auf dem gelöschten Profil: %+v", gezeigt.Profil)
	}

	// **Und eine Abfrage glückt** — das ist die Probe darauf, dass `a.profil`
	// mitgewandert ist und nicht auf eine tote Kennung zeigt.
	if sets := anwendung.Sets(t.Context()); sets.Fehler != nil {
		t.Errorf("die Sets nach dem Löschen: %+v", sets.Fehler)
	}
}

// **Eine unbrauchbare Kennung wird an allen vier Wegen abgewiesen.**
func TestEineUnbrauchbareKennungWirdInDerPflegeAbgewiesen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	if antwort := anwendung.NutzerAendern(t.Context(), app.NutzerangabenDTO{
		ID: "keine Zahl", Name: "X", Fassung: "1",
	}); antwort.Fehler == nil {
		t.Error("NutzerAendern kam durch")
	}

	// **Und eine unbrauchbare Fassung ebenso** — Werkstück 8 (#1351).
	if antwort := anwendung.NutzerAendern(t.Context(), app.NutzerangabenDTO{
		ID: "1", Name: "X", Fassung: "keine Zahl",
	}); antwort.Fehler == nil {
		t.Error("die unbrauchbare Fassung kam durch")
	}

	if antwort := anwendung.ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: "keine Zahl", Name: "X", Fassung: "1",
	}); antwort.Fehler == nil {
		t.Error("ProfilAendern kam durch")
	}

	if antwort := anwendung.ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: "1", Name: "X", Fassung: "keine Zahl",
	}); antwort.Fehler == nil {
		t.Error("die unbrauchbare Profilfassung kam durch")
	}

	if antwort := anwendung.NutzerLoeschen(
		t.Context(), "keine Zahl", "X"); antwort.Fehler == nil {
		t.Error("NutzerLoeschen kam durch")
	}

	if antwort := anwendung.ProfilLoeschen(
		t.Context(), "keine Zahl", "X"); antwort.Fehler == nil {
		t.Error("ProfilLoeschen kam durch")
	}
}

// **Und jeder Fehler der Naht reicht durch.**
func TestJederPflegefehlerReichtAmFensterDurch(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitNutzern(&nutzerfuehrer{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.NutzerAendern(t.Context(), app.NutzerangabenDTO{
		ID: "1", Name: "X", Bild: "", Fassung: "1",
	}); antwort.Fehler == nil {
		t.Error("NutzerAendern kam durch")
	}

	if antwort := anwendung.ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: "1", Name: "X", Fassung: "1",
	}); antwort.Fehler == nil {
		t.Error("ProfilAendern kam durch")
	}

	if antwort := anwendung.NutzerLoeschen(t.Context(), "1", "X"); antwort.Fehler == nil {
		t.Error("NutzerLoeschen kam durch")
	}

	if antwort := anwendung.ProfilLoeschen(t.Context(), "1", "X"); antwort.Fehler == nil {
		t.Error("ProfilLoeschen kam durch")
	}
}

// **Ein Profil zu ändern, das es nicht gibt, meldet den Lesefehler durch.**
func TestEineUnlesbareUebersichtBeimProfilaendernMeldetSich(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitNutzern(&nutzerfuehrer{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: "1", Name: "X", Fassung: "1",
	}); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Ein Nutzer, den es nicht gibt, lässt sich nicht löschen.**
func TestEinFehlenderNutzerLaesstSichAmFensterNichtLoeschen(t *testing.T) {
	antwort := angelaufeneAnwendung(t).NutzerLoeschen(t.Context(), "4711", "Wer")

	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", antwort.Fehler)
	}
}

// **Ein Profilname, den es schon gibt, meldet sich am Fenster.**
//
// **Der Weg ist ein anderer als beim Lesefehler darüber**: die Übersicht
// glückt, der Eigner steht — und erst das Schreiben scheitert.
func TestEinBelegterProfilnameMeldetSichAmFenster(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	angelegt := anwendung.ProfilAnlegen(t.Context(), "1", "Turnier")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	antwort := anwendung.ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: angelegt.Profil.ID, Name: "Mein Profil", Fassung: "1",
	})
	if antwort.Fehler == nil {
		t.Fatal("der belegte Name kam durch")
	}

	if antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("Code %q", antwort.Fehler.Code)
	}
}

// **Ein Nutzer lässt sich löschen — und die Antwort trägt keinen Fehler.**
func TestEinNutzerLaesstSichAmFensterLoeschen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	angelegt := anwendung.NutzerAnlegen(t.Context(), "Zweite", "")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if geloescht := anwendung.NutzerLoeschen(
		t.Context(), angelegt.Nutzer.ID, "Zweite"); geloescht.Fehler != nil {
		t.Fatalf("%+v", geloescht.Fehler)
	}

	// **Und er ist weg.**
	uebrig := anwendung.Nutzer(t.Context())
	if uebrig.Fehler != nil || len(uebrig.Nutzer) != 1 {
		t.Errorf("%+v / %+v", uebrig.Nutzer, uebrig.Fehler)
	}
}
