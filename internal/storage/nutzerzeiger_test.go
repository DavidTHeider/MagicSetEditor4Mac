// **Wer arbeitet, steht im Bestand** — Werkstück 17 von Phase 4.12
// (#1400),
// nach [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).
package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Eintrag gilt der Eigner des gezeigten Profils.**
//
// # Warum ein Rückfall und kein Zwang
//
// **Weil es Bestände gibt, die den Eintrag nicht haben** — jeder, der vor
// diesem Werkstück entstand. Ein Zeiger, der auf nichts steht, ließe die
// Anwendung ohne Identität dastehen; der Rückfall ist genau das Verhalten von
// vorher, und damit ändert sich für einen alten Bestand beim ersten Start
// nichts.
func TestOhneEintragGiltDerEignerDesGezeigtenProfils(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	nutzer, err := ablage.NutzerAnlegen(probekontext(t),
		domain.Nutzer{Name: "Wer"})
	if err != nil {
		t.Fatalf("Nutzer anlegen: %v", err)
	}

	profil, err := ablage.ProfilAnlegen(probekontext(t),
		domain.Profil{NutzerID: nutzer.ID, Name: "Meins"})
	if err != nil {
		t.Fatalf("Profil anlegen: %v", err)
	}

	if fehler := ablage.ProfilzeigerSetzen(probekontext(t), profil.ID); fehler != nil {
		t.Fatalf("Zeiger setzen: %v", fehler)
	}

	werArbeitet, err := ablage.Nutzerzeiger(probekontext(t))
	if err != nil {
		t.Fatalf("Nutzerzeiger: %v", err)
	}

	if werArbeitet != nutzer.ID {
		t.Errorf("es arbeitet %d, erwartet war %d", werArbeitet, nutzer.ID)
	}
}

// **Und ist er gesetzt, gilt er — auch wenn der Zeiger auf einem fremden
// Profil steht.**
//
// **Das ist der ganze Punkt dieses Werkstücks.** Vorher war „wer arbeitet"
// aus dem gezeigten Profil abgeleitet; wer hineinwechselte, wurde dessen
// Eigner. Gemessen am 2026-09-06, und der Weg stand in
// #1400.
func TestEinGesetzterNutzerzeigerUeberlebtDenProfilwechsel(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	erster, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Erster"})
	if err != nil {
		t.Fatalf("Nutzer anlegen: %v", err)
	}

	zweiter, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweiter"})
	if err != nil {
		t.Fatalf("Nutzer anlegen: %v", err)
	}

	fremd, err := ablage.ProfilAnlegen(probekontext(t),
		domain.Profil{NutzerID: zweiter.ID, Name: "Fremd"})
	if err != nil {
		t.Fatalf("Profil anlegen: %v", err)
	}

	if fehler := ablage.NutzerzeigerSetzen(probekontext(t), erster.ID); fehler != nil {
		t.Fatalf("Nutzerzeiger setzen: %v", fehler)
	}

	if fehler := ablage.ProfilzeigerSetzen(probekontext(t), fremd.ID); fehler != nil {
		t.Fatalf("Profilzeiger setzen: %v", fehler)
	}

	werArbeitet, err := ablage.Nutzerzeiger(probekontext(t))
	if err != nil {
		t.Fatalf("Nutzerzeiger: %v", err)
	}

	if werArbeitet != erster.ID {
		t.Errorf("es arbeitet %d, erwartet war der Erste (%d)", werArbeitet, erster.ID)
	}
}

// **Ein Zeiger auf einen Nutzer, den es nicht mehr gibt, fällt zurück.**
//
// **Derselbe Grund wie beim Profilzeiger**: die Kaskade räumt beim Löschen
// eines Nutzers seine Profile, und ein Zeiger ins Leere ließe jede Abfrage
// ohne Identität dastehen.
func TestEinZeigerAufEinenGeloeschtenNutzerFaelltZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	bleibt, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Bleibt"})
	if err != nil {
		t.Fatalf("Nutzer anlegen: %v", err)
	}

	geht, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Geht"})
	if err != nil {
		t.Fatalf("Nutzer anlegen: %v", err)
	}

	profil, err := ablage.ProfilAnlegen(probekontext(t),
		domain.Profil{NutzerID: bleibt.ID, Name: "Meins"})
	if err != nil {
		t.Fatalf("Profil anlegen: %v", err)
	}

	if fehler := ablage.ProfilzeigerSetzen(probekontext(t), profil.ID); fehler != nil {
		t.Fatalf("Profilzeiger setzen: %v", fehler)
	}

	if fehler := ablage.NutzerzeigerSetzen(probekontext(t), geht.ID); fehler != nil {
		t.Fatalf("Nutzerzeiger setzen: %v", fehler)
	}

	if fehler := ablage.NutzerLoeschen(probekontext(t), geht.ID, geht.Name); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	werArbeitet, err := ablage.Nutzerzeiger(probekontext(t))
	if err != nil {
		t.Fatalf("Nutzerzeiger: %v", err)
	}

	if werArbeitet != bleibt.ID {
		t.Errorf("es arbeitet %d, erwartet war der Eigner des gezeigten Profils (%d)",
			werArbeitet, bleibt.ID)
	}
}

// **Und einen Nutzer, den es nicht gibt, nimmt der Zeiger nicht an.**
func TestEinNutzerzeigerAufNichtsWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)

	fehler := storage.NeueNutzerablage(db).NutzerzeigerSetzen(probekontext(t), 999)

	if fehler == nil {
		t.Error("der Zeiger nahm eine Kennung an, die es nicht gibt")
	}
}

// **Die Fehlerwege des Zeigers** — je einer, und beide betreten.
//
// **Ein Lesefehler ist kein „niemand arbeitet"** (§8): käme er als `0` zurück,
// hielte der Dienst darüber das für Unwissen und meldete etwas über Rollen,
// während in Wahrheit die Platte klemmt.
func TestDieFehlerwegeDesNutzerzeigers(t *testing.T) {
	// **Eine geschlossene Datenbank** — der Helfer nebenan liegt im Paket
	// selbst und ist von hier aus nicht zu sehen; drei Zeilen sind billiger
	// als eine zweite Datei.
	zu, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := zu.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	ablage := storage.NeueNutzerablage(zu)

	if _, fehler := ablage.Nutzerzeiger(probekontext(t)); fehler == nil {
		t.Error("das Lesen glückte an einer geschlossenen Datenbank")
	}

	if fehler := ablage.NutzerzeigerSetzen(probekontext(t), 1); fehler == nil {
		t.Error("das Setzen glückte an einer geschlossenen Datenbank")
	}
}

// **Und das Schreiben scheitert auch, wenn den Nutzer nur der Bestand nicht
// mehr nimmt.**
//
// **Der Zweig hinter der Daseinsprüfung**: der Nutzer ist da, und das
// Schreiben geht trotzdem schief. Ohne ihn bliebe die Zeile ungeprüft, die
// den Fehler übersetzt.
func TestEinScheiterndesSchreibenDesZeigersWirdGemeldet(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	nutzer, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Wer"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Die Tafel weg, der Nutzer bleibt.** Damit läuft die Daseinsprüfung
	// durch und das Schreiben nicht.
	if _, fehler := db.ExecContext(probekontext(t),
		`DROP TABLE bestand`); fehler != nil {
		t.Fatalf("Tafel entfernen: %v", fehler)
	}

	if fehler := ablage.NutzerzeigerSetzen(
		probekontext(t), nutzer.ID); fehler == nil {
		t.Error("das Setzen glückte ohne die Tafel")
	}
}

// **Ein Nutzerwechsel lässt keinen halben Stand zurück.**
//
// # Der Befund
//
// **Zweitblick über die Phasen 4.0 bis 4.13** (2026-09-07): `NutzerWechseln`
// setzte erst den Nutzer-, dann den Profilzeiger, in zwei Anweisungen ohne
// gemeinsame Transaktion. Scheiterte die zweite, war die erste durch —
// **es arbeitete ein anderer Mensch, während das Profil des alten gezeigt
// wurde**, und der Ruf meldete dazu einen Fehler.
//
// # Wie der Fall hergestellt wird
//
// **Über ein Profil, das es nicht gibt.** Die Prüfung liegt in derselben
// Transaktion wie das Schreiben; wer nicht da ist, lässt nichts zurück. Ohne
// Transaktion stünde der Nutzerzeiger danach auf dem neuen Menschen.
func TestEinGescheiterterWechselLaesstKeinenHalbenStand(t *testing.T) {
	t.Parallel()

	db := frischeDateidatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	vorher, lesefehler := ablage.Nutzerzeiger(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("den Zeiger lesen: %v", lesefehler)
	}

	// **Ein zweiter Nutzer, damit es überhaupt etwas umzustellen gibt.**
	zweiter, anlegefehler := ablage.NutzerAnlegen(probekontext(t),
		domain.Nutzer{Name: "Zweiter"})
	if anlegefehler != nil {
		t.Fatalf("anlegen: %v", anlegefehler)
	}

	// **Und ein Profil, das der Bestand nicht führt.**
	if fehler := ablage.ZeigerSetzen(
		probekontext(t), zweiter.ID, 999_999); fehler == nil {
		t.Fatal("der Wechsel glückte mit einem Profil, das es nicht gibt")
	}

	nachher, nachfehler := ablage.Nutzerzeiger(probekontext(t))
	if nachfehler != nil {
		t.Fatalf("den Zeiger lesen: %v", nachfehler)
	}

	if nachher != vorher {
		t.Errorf("der Nutzerzeiger steht auf %d statt auf %d — der Wechsel "+
			"ist halb durchgegangen", nachher, vorher)
	}
}
