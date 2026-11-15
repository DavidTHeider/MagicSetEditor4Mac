package app

// **Einladen und annehmen an der Naht** — NET-01
// (#1705).
//
// # Was hier gemessen wird
//
// **Zwei Anwendungen mit je eigenem Bestand und eigener Identität** — der
// Austausch aus dem Artboard: Anna stellt aus, Bert nimmt an, und beide sehen
// denselben Vergleichscode.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineEinladungsseite gibt eine Anwendung mit Bestand, Profil und Identität.
func eineEinladungsseite(t *testing.T) *App {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPeerbund(&peerprobebund{})

	if antwort := anwendung.PeeridentitaetSicherstellen(
		riegelkontext(t)); antwort.Fehler != nil {
		t.Fatalf("die Peeridentität: %+v", antwort.Fehler)
	}

	return anwendung
}

func TestEinladungAnDerNaht(t *testing.T) {
	t.Parallel()

	anna, bert := eineEinladungsseite(t), eineEinladungsseite(t)
	ctx := riegelkontext(t)

	ausgestellt := anna.KontakteinladungErstellen(ctx, "Anna", "handel")
	if ausgestellt.Fehler != nil || !strings.HasPrefix(ausgestellt.Einladung.Code, domain.Kontakteinladungsmarke+":") ||
		ausgestellt.Einladung.Ablauf == "" || ausgestellt.Einladung.Fingerabdruck == "" {
		t.Fatalf("Erstellen: %+v", ausgestellt)
	}

	vorschau := bert.KontakteinladungPruefen(ctx, ausgestellt.Einladung.Code)
	if vorschau.Fehler != nil || vorschau.Vorschau.Name != "Anna" ||
		vorschau.Vorschau.Zweck != "handel" || vorschau.Vorschau.Lage != "unbekannt" ||
		vorschau.Vorschau.Adresse != "" {
		t.Fatalf("Pruefen: %+v", vorschau)
	}

	// **Ohne Vergleich angenommen**: der Kontakt steht da, und die Liste
	// zeigt den Code, den beide gleich sehen.
	liste := bert.KontakteinladungAnnehmen(ctx, ausgestellt.Einladung.Code, "Anna", "", false, "")
	if liste.Fehler != nil || len(liste.Kontakte) != 1 {
		t.Fatalf("Annehmen: %+v", liste)
	}

	kontakt := liste.Kontakte[0]
	if kontakt.Lage != "unbestaetigt" || kontakt.Vergleichscode != vorschau.Vorschau.Vergleichscode {
		t.Fatalf("Kontakt: %+v", kontakt)
	}

	// **Bestätigt mit dem Code** — derselbe Weg wie mit dem Abdruck.
	bestaetigt := bert.PeerkontaktBestaetigen(ctx, kontakt.Kennung, kontakt.Vergleichscode)
	if bestaetigt.Fehler != nil || bestaetigt.Kontakte[0].Lage != "bestaetigt" ||
		bestaetigt.Kontakte[0].Vergleichscode != "" {
		t.Fatalf("Bestaetigen: %+v", bestaetigt)
	}

	// **Und ein zweites Mal geht dieselbe Einladung nicht.**
	if wieder := bert.KontakteinladungAnnehmen(
		ctx, ausgestellt.Einladung.Code, "", "", false, ""); wieder.Fehler == nil {
		t.Fatal("dieselbe Einladung zweimal angenommen")
	}
}

func TestEinladungFehlerAnDerNaht(t *testing.T) {
	t.Parallel()

	ohne := PeerflaecheUeber(bestandFuerDenRiegel(t))
	ohne.profil.Store(1)
	ctx := riegelkontext(t)

	// **Ohne Identität keine Einladung** — und die Antwort sagt es.
	if antwort := ohne.KontakteinladungErstellen(ctx, "Anna", "handel"); antwort.Fehler == nil {
		t.Fatal("ausgestellt ohne Identität")
	}

	if antwort := ohne.KontakteinladungPruefen(ctx, "kein Code"); antwort.Fehler == nil {
		t.Fatal("geprüft ohne Code")
	}

	if antwort := ohne.KontakteinladungAnnehmen(ctx, "kein Code", "", "", false, ""); antwort.Fehler == nil ||
		antwort.Kontakte == nil {
		t.Fatalf("Annehmen: %+v", antwort)
	}
}

func TestEinladungTraegtDieAdresse(t *testing.T) {
	t.Parallel()

	anna, bert := eineEinladungsseite(t), eineEinladungsseite(t)
	ctx := riegelkontext(t)

	anna.fernlage.einladung = domain.NeueEinladung("fern.example:1")

	ausgestellt := anna.KontakteinladungErstellen(ctx, "Anna", "partie")
	if vorschau := bert.KontakteinladungPruefen(
		ctx, ausgestellt.Einladung.Code); vorschau.Vorschau.Adresse != "mse4mac:1:fern.example:1" {
		t.Fatalf("Adresse: %+v", vorschau)
	}

	anna.fernlage.einladung = ""
	anna.peerlage.horchen.erreichbar.Adressen = []string{"192.168.1.42:41234"}

	ausgestellt = anna.KontakteinladungErstellen(ctx, "Anna", "partie")
	if vorschau := bert.KontakteinladungPruefen(
		ctx, ausgestellt.Einladung.Code); vorschau.Vorschau.Adresse != "192.168.1.42:41234" {
		t.Fatalf("Adresse: %+v", vorschau)
	}
}

func TestVorschauNenntDenVorgaenger(t *testing.T) {
	t.Parallel()

	vorgaenger := &domain.Peerkontakt{SelbstgewaehlterName: "Anna"}

	if dto := alsVorschau(vorschauMit(vorgaenger)); dto.Vorgaenger != "Anna" {
		t.Fatalf("ohne eigenen Namen: %+v", dto)
	}

	vorgaenger.EigenerName = "Anna vom Stammtisch"
	if dto := alsVorschau(vorschauMit(vorgaenger)); dto.Vorgaenger != "Anna vom Stammtisch" {
		t.Fatalf("mit eigenem Namen: %+v", dto)
	}
}

// vorschauMit gibt eine Vorschau, die nur einen Vorgänger trägt.
func vorschauMit(vorgaenger *domain.Peerkontakt) service.Einladungsvorschau {
	return service.Einladungsvorschau{Vorgaenger: vorgaenger}
}

// **Ohne Bestand fragt kein Riegel** (#2000) — und es gibt dann auch nichts
// anzunehmen: ohne Bestand keine Identität, und die Annahme scheitert vorher.
func TestOhneBestandGibtEsKeinenSenderiegel(t *testing.T) {
	t.Parallel()

	if riegel := NeueApp("").senderiegel(); riegel != nil {
		t.Errorf("ohne Bestand ein Riegel: %T", riegel)
	}

	if mit := eineEinladungsseite(t).senderiegel(); mit == nil {
		t.Error("mit Bestand kein Riegel")
	}
}
