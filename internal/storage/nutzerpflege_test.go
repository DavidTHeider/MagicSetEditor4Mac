package storage_test

// **Nutzer und Profile ändern und löschen** — Werkstück 5 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Nutzer lässt sich umbenennen und bekommt ein Bild.**
func TestEinNutzerLaesstSichAendern(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	geaendert, fehler := ablage.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 1, Name: "Neuer Name", Bild: "ich.png", Fassung: 1})
	if fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if geaendert.Name != "Neuer Name" {
		t.Errorf("Name %q", geaendert.Name)
	}

	gelesen, lesefehler := ablage.Nutzer(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen[0].Name != "Neuer Name" || gelesen[0].Bild != "ich.png" {
		t.Errorf("gelesen %+v", gelesen[0])
	}
}

// **Ein Nutzer, den es nicht gibt, ist „nicht gefunden".**
func TestEinFehlenderNutzerLaesstSichNichtAendern(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	_, fehler := ablage.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 4711, Name: "Wer"})

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// **Ein Name, den es schon gibt, ist ein Konflikt.**
func TestEinBelegterNameIstEinKonflikt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweiter, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{Name: "Zweite"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	erste, lesefehler := ablage.Nutzer(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	_, aenderfehler := ablage.NutzerAendern(t.Context(),
		domain.Nutzer{ID: zweiter.ID, Name: erste[0].Name})

	if !errors.Is(aenderfehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", aenderfehler)
	}
}

// **Ein unbrauchbarer Name erreicht den Bestand nicht.**
func TestEinUnbrauchbarerNameErreichtDenBestandNicht(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 1, Name: " "}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Ein Profil lässt sich umbenennen.**
func TestEinProfilLaesstSichAendern(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	geaendert, fehler := ablage.ProfilAendern(t.Context(),
		domain.Profil{ID: 1, NutzerID: 1, Name: "Turnier", Fassung: 1})
	if fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if geaendert.Name != "Turnier" {
		t.Errorf("Name %q", geaendert.Name)
	}
}

// **Ein Profil, das es nicht gibt, ist „nicht gefunden".**
func TestEinFehlendesProfilLaesstSichNichtAendern(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	_, fehler := ablage.ProfilAendern(t.Context(),
		domain.Profil{ID: 4711, NutzerID: 1, Name: "Turnier"})

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// **Ein Profil lässt sich löschen — wenn der Nutzer noch eines behält.**
func TestEinProfilLaesstSichLoeschen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweites, fehler := ablage.ProfilAnlegen(t.Context(),
		domain.Profil{NutzerID: 1, Name: "Turnier"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if loeschfehler := ablage.ProfilLoeschen(
		t.Context(), zweites.ID, "Turnier", 0); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	profile, lesefehler := ablage.Profile(t.Context(), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(profile) != 1 {
		t.Errorf("%d Profile: %+v", len(profile), profile)
	}
}

// **Das letzte Profil eines Nutzers bleibt.**
//
// # Warum das eine Regel und keine Bequemlichkeit ist
//
// **Ein Nutzer ohne Arbeitsumgebung könnte sich anmelden und nichts tun.**
// Genau deshalb legt der Speicher beim Anlegen eines Nutzers eines mit an
// (Werkstück 1) — und deshalb darf das letzte nicht wegfallen.
func TestDasLetzteProfilEinesNutzersBleibt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	fehler := ablage.ProfilLoeschen(t.Context(), 1, storage.Erstprofil, 0)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", fehler)
	}
}

// **Der Name bestätigt** — dieselbe Form wie beim Löschen eines Sets.
func TestEinFalscherNameLoeschtKeinProfil(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweites, fehler := ablage.ProfilAnlegen(t.Context(),
		domain.Profil{NutzerID: 1, Name: "Turnier"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if loeschfehler := ablage.ProfilLoeschen(
		t.Context(), zweites.ID, "etwas anderes", 0); !errors.Is(
		loeschfehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", loeschfehler)
	}
}

// **Ein Nutzer lässt sich löschen — wenn ein anderer bleibt.**
func TestEinNutzerLaesstSichLoeschen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweiter, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{Name: "Zweite"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if loeschfehler := ablage.NutzerLoeschen(
		t.Context(), zweiter.ID, "Zweite"); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	gelesen, lesefehler := ablage.Nutzer(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen) != 1 {
		t.Errorf("%d Nutzer: %+v", len(gelesen), gelesen)
	}

	// **Und seine Profile gehen mit** — `ON DELETE CASCADE` (Wanderung 0034).
	profile, profilfehler := ablage.Profile(t.Context(), zweiter.ID)
	if profilfehler != nil {
		t.Fatalf("die Profile lesen: %v", profilfehler)
	}

	if len(profile) != 0 {
		t.Errorf("%d Profile blieben: %+v", len(profile), profile)
	}
}

// **Der letzte Nutzer bleibt.**
//
// **Sonst stünde die Anwendung ohne Profil da**, und jede Abfrage gäbe eine
// leere Liste — der Bestand sähe aus, als wäre er weg.
func TestDerLetzteNutzerBleibt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	fehler := ablage.NutzerLoeschen(t.Context(), 1, storage.Erstnutzer)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", fehler)
	}
}

// **Und ein falscher Name löscht keinen Nutzer.**
func TestEinFalscherNameLoeschtKeinenNutzer(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweiter, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{Name: "Zweite"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if loeschfehler := ablage.NutzerLoeschen(
		t.Context(), zweiter.ID, "etwas anderes"); !errors.Is(
		loeschfehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", loeschfehler)
	}
}

// **Ein unbrauchbares Profil erreicht den Bestand nicht.**
func TestEinUnbrauchbaresProfilErreichtDenBestandNicht(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.ProfilAendern(t.Context(),
		domain.Profil{ID: 1, NutzerID: 1, Name: " "}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Ein Profilname, den derselbe Nutzer schon führt, ist ein Konflikt.**
func TestEinBelegterProfilnameIstEinKonflikt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweites, fehler := ablage.ProfilAnlegen(t.Context(),
		domain.Profil{NutzerID: 1, Name: "Turnier"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	_, aenderfehler := ablage.ProfilAendern(t.Context(),
		domain.Profil{ID: zweites.ID, NutzerID: 1, Name: storage.Erstprofil})

	if !errors.Is(aenderfehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", aenderfehler)
	}
}

// **Und was es nicht gibt, lässt sich nicht löschen.**
func TestWasEsNichtGibtLaesstSichNichtLoeschen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if fehler := ablage.NutzerLoeschen(t.Context(), 4711, "Wer"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Nutzer: %v, erwartet war ErrNotFound", fehler)
	}

	if fehler := ablage.ProfilLoeschen(t.Context(), 4711, "Wer", 0); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Profil: %v, erwartet war ErrNotFound", fehler)
	}
}

// **Eine überholte Fassung wird abgewiesen — und ändert nichts.**
//
// # Was das schließt
//
// **Die offene Stelle aus Werkstück 5**: `nutzer` und `profile` trugen keine
// Fassung, und wer gleichzeitig an zwei Flächen einen Namen änderte, verlor
// eine der beiden Änderungen — ohne dass jemand es merkte (§7, Regel 2 über
// die Flächen hinweg).
func TestEineUeberholteFassungWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	// **Die erste Änderung glückt und hebt die Fassung.**
	erste, fehler := ablage.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 1, Name: "Erste", Fassung: 1})
	if fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if erste.Fassung != 2 {
		t.Errorf("die Fassung steht auf %d", erste.Fassung)
	}

	// **Die zweite kommt mit der alten und wird abgewiesen.**
	_, zweite := ablage.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 1, Name: "Zweite", Fassung: 1})
	if !errors.Is(zweite, domain.ErrConflict) {
		t.Fatalf("angenommen wurde: %v", zweite)
	}

	// **Und der Name ist der der ersten.**
	gelesen, lesefehler := ablage.Nutzer(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen[0].Name != "Erste" {
		t.Errorf("der Name ist %q", gelesen[0].Name)
	}
}

// **Und dasselbe beim Profil.**
func TestEineUeberholteProfilfassungWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.ProfilAendern(t.Context(),
		domain.Profil{ID: 1, NutzerID: 1, Name: "Erstes", Fassung: 1}); fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	_, zweite := ablage.ProfilAendern(t.Context(),
		domain.Profil{ID: 1, NutzerID: 1, Name: "Zweites", Fassung: 1})
	if !errors.Is(zweite, domain.ErrConflict) {
		t.Errorf("angenommen wurde: %v", zweite)
	}
}

// **„Gibt es nicht" und „hat sich geändert" sind zwei Antworten.**
//
// **Weil die Oberfläche dafür etwas anderes tut** (§8): die eine heißt „du
// hast den falschen erwischt", die andere „lies neu und versuch es noch
// einmal".
func TestGibtEsNichtUndHatSichGeaendertSindZwei(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	_, fehler := ablage.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 4711, Name: "Wer", Fassung: 1})
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}

	_, profilfehler := ablage.ProfilAendern(t.Context(),
		domain.Profil{ID: 4711, NutzerID: 1, Name: "Wer", Fassung: 1})
	if !errors.Is(profilfehler, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", profilfehler)
	}
}
