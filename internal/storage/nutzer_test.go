package storage_test

// **Nutzer und Profile im Bestand** — Werkstück 1 von Phase 4.12
// (#1351),
// nach Vermerk 0066.
//
// # Die Zusage, die zuerst zählt
//
// **Niemand ist danach ausgesperrt.** Eine Bestandsdatei, die es heute gibt,
// bekommt einen ersten Nutzer mit einem ersten Profil — und wer nie einen
// zweiten anlegt, merkt von dieser Phase nichts (Vermerk 0066,
// Entscheidung 4).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein frischer Bestand hat einen Nutzer mit einem Profil.**
//
// Nicht null: eine Anwendung, die beim ersten Start eine Anmeldemaske zeigt,
// wo vorher ein Fenster aufging, hat eine Phase eingeführt und eine Fähigkeit
// genommen.
func TestEinFrischerBestandHatEinenNutzer(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	nutzer, err := ablage.Nutzer(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(nutzer) != 1 {
		t.Fatalf("%d Nutzer, erwartet war einer: %+v", len(nutzer), nutzer)
	}

	if nutzer[0].Name == "" {
		t.Error("der erste Nutzer hat keinen Namen")
	}

	profile, lesefehler := ablage.Profile(probekontext(t), nutzer[0].ID)
	if lesefehler != nil {
		t.Fatalf("die Profile lesen: %v", lesefehler)
	}

	if len(profile) != 1 {
		t.Fatalf("%d Profile, erwartet war eines: %+v", len(profile), profile)
	}
}

// **Ein zweiter Nutzer lässt sich anlegen und wieder lesen.**
func TestEinZweiterNutzerLaesstSichAnlegen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.ID == 0 {
		t.Fatal("der neue Nutzer hat keine Kennung")
	}

	nutzer, lesefehler := ablage.Nutzer(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(nutzer) != 2 {
		t.Errorf("%d Nutzer, erwartet waren zwei", len(nutzer))
	}
}

// **Ein Nutzer entsteht mit einem Profil, nicht ohne.**
//
// Ein Nutzer ohne Profil hätte keine Arbeitsumgebung — er könnte sich anmelden
// und nichts tun. Die Ablage legt es mit an, damit dieser Zustand gar nicht
// erst entsteht.
func TestEinNeuerNutzerBekommtEinProfil(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	profile, lesefehler := ablage.Profile(probekontext(t), angelegt.ID)
	if lesefehler != nil {
		t.Fatalf("die Profile lesen: %v", lesefehler)
	}

	if len(profile) != 1 {
		t.Errorf("%d Profile, erwartet war eines", len(profile))
	}
}

// **Zwei Nutzer desselben Namens sind ein Konflikt.**
//
// Nicht aus Ordnungsliebe: der Name ist das Einzige, woran man sie in der
// Profilwahl unterscheidet. Zwei gleiche wären zwei Zeilen, zwischen denen
// niemand wählen kann.
func TestZweiNutzerDesselbenNamensSindEinKonflikt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"}); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	_, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler ist %v, erwartet ErrConflict", err)
	}
}

// **Ein krummer Nutzer kommt auch dann nicht in den Bestand, wenn niemand
// gefragt hat.**
//
// # Warum das hier steht und nicht nur im Dienst
//
// **Geprüft wird im Dienst** (§2) — das ist die erste Reihe, und dort kommt
// der Feldfehler her. Dies ist die zweite: ein `CHECK` in der Wanderung. Er
// greift, wo niemand durch den Dienst geht — eine frühere Fassung, ein Import,
// eine Hand am Bestand.
//
// **Und er ist keine zweite Wahrheit**, sondern dieselbe: „nicht leer, keine
// Ränder", buchstäblich das, was `domain.Nutzer.Validate` sagt.
func TestEinKrummerNutzerKommtNichtInDenBestand(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	for _, name := range []string{"  ", " David ", ""} {
		t.Run(name, func(t *testing.T) {
			if _, err := ablage.NutzerAnlegen(
				probekontext(t), domain.Nutzer{Name: name}); err == nil {
				t.Errorf("%q kam durch", name)
			}
		})
	}
}

// **Ein zweites Profil lässt sich anlegen.**
func TestEinZweitesProfilLaesstSichAnlegen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	nutzer, err := ablage.Nutzer(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if _, err := ablage.ProfilAnlegen(probekontext(t), domain.Profil{
		NutzerID: nutzer[0].ID, Name: "Turnier",
	}); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	profile, lesefehler := ablage.Profile(probekontext(t), nutzer[0].ID)
	if lesefehler != nil {
		t.Fatalf("die Profile lesen: %v", lesefehler)
	}

	if len(profile) != 2 {
		t.Errorf("%d Profile, erwartet waren zwei", len(profile))
	}
}

// **Ein Profil für einen Nutzer, den es nicht gibt, ist ein Nein.**
func TestEinProfilOhneNutzerWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, err := ablage.ProfilAnlegen(probekontext(t), domain.Profil{
		NutzerID: 999, Name: "Waise",
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Der Zeiger folgt dem Bestand, und ein Wechsel wird gemerkt** (#1351).
func TestDerProfilzeigerLaesstSichSetzenUndLesen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	// **Ohne Eintrag zeigt er auf das erste** — die Wanderung legt genau eines
	// an, und ohne Rückfall stünde eine frische Anwendung ohne Profil da.
	erst, fehler := ablage.Profilzeiger(t.Context())
	if fehler != nil {
		t.Fatalf("den Zeiger lesen: %v", fehler)
	}

	if erst != 1 {
		t.Fatalf("der Zeiger steht auf %d, erwartet war das erste Profil", erst)
	}

	nutzer, anlagefehler := ablage.NutzerAnlegen(t.Context(),
		domain.Nutzer{Name: "Zweite"})
	if anlagefehler != nil {
		t.Fatalf("den zweiten Nutzer anlegen: %v", anlagefehler)
	}

	profile, lesefehler := ablage.Profile(t.Context(), nutzer.ID)
	if lesefehler != nil {
		t.Fatalf("die Profile lesen: %v", lesefehler)
	}

	if setzfehler := ablage.ProfilzeigerSetzen(t.Context(), profile[0].ID); setzfehler != nil {
		t.Fatalf("den Zeiger setzen: %v", setzfehler)
	}

	jetzt, nachfehler := ablage.Profilzeiger(t.Context())
	if nachfehler != nil {
		t.Fatalf("den Zeiger erneut lesen: %v", nachfehler)
	}

	if jetzt != profile[0].ID {
		t.Errorf("der Zeiger steht auf %d, gesetzt war %d", jetzt, profile[0].ID)
	}
}

// **Ein Zeiger auf ein Profil, das es nicht gibt, wird abgewiesen.**
//
// Sonst zeigte die Anwendung auf nichts, und **jede** Abfrage gäbe eine leere
// Liste — der Bestand sähe aus, als wäre er weg.
func TestEinZeigerAufEinFehlendesProfilWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	fehler := ablage.ProfilzeigerSetzen(t.Context(), 4711)
	if fehler == nil {
		t.Fatal("das fehlende Profil kam durch")
	}

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// **Ein geschlossener Bestand meldet sich an beiden Zeigerwegen.**
func TestDieZeigerwegeMeldenEinenGeschlossenenBestand(t *testing.T) {
	ablage := storage.NeueNutzerablage(kaputterBestand(t))

	if _, fehler := ablage.Profilzeiger(t.Context()); fehler == nil {
		t.Error("das Lesen kam durch")
	}

	if fehler := ablage.ProfilzeigerSetzen(t.Context(), 1); fehler == nil {
		t.Error("das Setzen kam durch")
	}
}

// **[storage.MitGezeigtemProfil] stattet einen Wurzelkontext aus** — der Weg,
// den Kommandozeile, MCP und REST gehen (#1351).
func TestDerWurzelkontextBekommtDasGezeigteProfil(t *testing.T) {
	db := frischeDatenbank(t)

	ausgestattet, fehler := storage.MitGezeigtemProfil(t.Context(), db)
	if fehler != nil {
		t.Fatalf("ausstatten: %v", fehler)
	}

	profilID, lesefehler := storage.ProfilAus(ausgestattet)
	if lesefehler != nil {
		t.Fatalf("das Profil lesen: %v", lesefehler)
	}

	if profilID != 1 {
		t.Errorf("Profil %d, erwartet war das erste", profilID)
	}

	// **Und ein Bestand ohne Zeiger reicht den Fehler durch**, statt einen
	// Kontext ohne Profil zurückzugeben, der erst später auffiele.
	if _, zufehler := storage.MitGezeigtemProfil(
		t.Context(), kaputterBestand(t)); zufehler == nil {
		t.Error("der geschlossene Bestand kam durch")
	}
}
