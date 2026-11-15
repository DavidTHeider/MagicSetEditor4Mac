// **Was „ein Zug" bedeutet, geprüft an beiden Enden** — Abdeckungsbefund vom
// 2026-09-07.
//
// Befund 1 des Zweitblicks über die Phasen 4.0 bis 4.13 hat zwei Übergänge
// in je eine Transaktion gezogen: [storage.Nutzerablage.ZeigerSetzen] stellt
// beide Zeiger zusammen, und [storage.Nutzerablage.ProfilLoeschen] bewegt den
// Ausweichzeiger in derselben Transaktion, in der es löscht.
//
// **Geprüft war davon nur das Nein.** Der Test daneben zeigt, dass ein
// abgewiesener Wechsel nichts zurücklässt; dass ein geglückter **beide**
// Zeiger bewegt, stand nirgends — und die Abdeckung hat es angestrichen: der
// ganze Erfolgsweg von `ZeigerSetzen` war unbetreten, und `ausweich` war in
// jedem Löschtest 0.
//
// **Beides zusammen ist die Zusage**: entweder ganz oder gar nicht. Wer nur
// die eine Hälfte prüft, hat die Zusage nicht geprüft, sondern eine ihrer
// beiden Aussagen (`AGENTS.md` §10).
package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein geglückter Wechsel bewegt beide Zeiger.**
func TestEinGeglueckterWechselBewegtBeideZeiger(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweiter, anlegefehler := ablage.NutzerAnlegen(probekontext(t),
		domain.Nutzer{Name: "Zweiter"})
	if anlegefehler != nil {
		t.Fatalf("Nutzer anlegen: %v", anlegefehler)
	}

	// **Sein Profil ist beim Anlegen mitentstanden** (Werkstück 1).
	profile, lesefehler := ablage.Profile(probekontext(t), zweiter.ID)
	if lesefehler != nil {
		t.Fatalf("Profile lesen: %v", lesefehler)
	}

	if len(profile) != 1 {
		t.Fatalf("%d Profile statt einem", len(profile))
	}

	if fehler := ablage.ZeigerSetzen(
		probekontext(t), zweiter.ID, profile[0].ID); fehler != nil {
		t.Fatalf("Zeiger setzen: %v", fehler)
	}

	arbeitet, nutzerfehler := ablage.Nutzerzeiger(probekontext(t))
	if nutzerfehler != nil {
		t.Fatalf("Nutzerzeiger lesen: %v", nutzerfehler)
	}

	if arbeitet != zweiter.ID {
		t.Errorf("es arbeitet %d statt %d", arbeitet, zweiter.ID)
	}

	gezeigt, profilfehler := ablage.Profilzeiger(probekontext(t))
	if profilfehler != nil {
		t.Fatalf("Profilzeiger lesen: %v", profilfehler)
	}

	if gezeigt != profile[0].ID {
		t.Errorf("gezeigt ist %d statt %d", gezeigt, profile[0].ID)
	}
}

// **Und ein Nutzer, den es nicht gibt, bewegt keinen von beiden.**
//
// **Die andere Hälfte der Schleife.** Der Test nebenan gibt ein Profil an,
// das der Bestand nicht führt; hier ist es der Nutzer — und er wird
// **zuerst** gesucht. Ohne diesen Fall bliebe offen, ob die Prüfung vor dem
// Schreiben steht oder erst hinterher aufräumt.
func TestEinUnbekannterNutzerBewegtKeinenZeiger(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDateidatenbank(t))

	vorherNutzer, nutzerfehler := ablage.Nutzerzeiger(probekontext(t))
	if nutzerfehler != nil {
		t.Fatalf("Nutzerzeiger lesen: %v", nutzerfehler)
	}

	vorherProfil, profilfehler := ablage.Profilzeiger(probekontext(t))
	if profilfehler != nil {
		t.Fatalf("Profilzeiger lesen: %v", profilfehler)
	}

	if fehler := ablage.ZeigerSetzen(
		probekontext(t), 999_999, vorherProfil); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
	}

	nachherNutzer, nachlesefehler := ablage.Nutzerzeiger(probekontext(t))
	if nachlesefehler != nil {
		t.Fatalf("Nutzerzeiger lesen: %v", nachlesefehler)
	}

	nachherProfil, nachprofilfehler := ablage.Profilzeiger(probekontext(t))
	if nachprofilfehler != nil {
		t.Fatalf("Profilzeiger lesen: %v", nachprofilfehler)
	}

	if nachherNutzer != vorherNutzer || nachherProfil != vorherProfil {
		t.Errorf("die Zeiger stehen auf %d/%d statt auf %d/%d",
			nachherNutzer, nachherProfil, vorherNutzer, vorherProfil)
	}
}

// **Ein geglücktes Löschen nimmt den Zeiger mit.**
func TestEinGeloeschtesProfilNimmtDenZeigerMit(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweites, anlegefehler := ablage.ProfilAnlegen(probekontext(t),
		domain.Profil{NutzerID: 1, Name: "Turnier"})
	if anlegefehler != nil {
		t.Fatalf("Profil anlegen: %v", anlegefehler)
	}

	if fehler := ablage.ProfilzeigerSetzen(
		probekontext(t), zweites.ID); fehler != nil {
		t.Fatalf("Zeiger setzen: %v", fehler)
	}

	// **Das Erstprofil ist der Ausweich** — es bleibt stehen.
	if fehler := ablage.ProfilLoeschen(
		probekontext(t), zweites.ID, "Turnier", 1); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	gezeigt, lesefehler := ablage.Profilzeiger(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Zeiger lesen: %v", lesefehler)
	}

	if gezeigt != 1 {
		t.Errorf("gezeigt ist %d statt des Ausweichs 1", gezeigt)
	}
}

// **Und ein abgewiesenes Löschen bewegt ihn nicht.**
//
// # Was dieser Test belegt — und was nicht
//
// **Vorher stellte der Dienst den Zeiger um, bevor er löschte.** Ein Löschen,
// das an der Namensprüfung scheiterte, ließ den Menschen in einem anderen
// Profil zurück, ohne dass er etwas gelöscht hätte.
//
// **Dieser Test hätte den alten Stand nicht gefunden**, und das gehört
// dazugesagt: dort nahm `ProfilLoeschen` gar keinen Ausweich entgegen — er
// hätte nicht einmal übersetzt. Die Gegenprobe gegen die alte Reihenfolge
// steht im Rumpf von 39187a6c und lief am Dienst.
//
// **Was er belegt, ist die Zusage von heute**: die Ablage bewegt den Zeiger
// nur, wenn sie auch löscht. Das ist die Stelle, an der sie eingelöst wird —
// und ein Wächter dort ist mehr wert als einer am Aufrufer (§10).
func TestEinAbgewiesenesLoeschenBewegtDenZeigerNicht(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zweites, anlegefehler := ablage.ProfilAnlegen(probekontext(t),
		domain.Profil{NutzerID: 1, Name: "Turnier"})
	if anlegefehler != nil {
		t.Fatalf("Profil anlegen: %v", anlegefehler)
	}

	if fehler := ablage.ProfilzeigerSetzen(
		probekontext(t), zweites.ID); fehler != nil {
		t.Fatalf("Zeiger setzen: %v", fehler)
	}

	if fehler := ablage.ProfilLoeschen(
		probekontext(t), zweites.ID, "ein falscher Name", 1); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war ErrValidation", fehler)
	}

	gezeigt, lesefehler := ablage.Profilzeiger(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Zeiger lesen: %v", lesefehler)
	}

	if gezeigt != zweites.ID {
		t.Errorf("gezeigt ist %d statt %d — das abgewiesene Löschen hat den "+
			"Menschen aus seinem Profil geholt", gezeigt, zweites.ID)
	}
}
