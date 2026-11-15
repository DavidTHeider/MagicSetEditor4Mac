package app

// **Der Riegel am Transport** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Warum der Lauf im Paket steht
//
// **Weil der Riegel in einem Abschluss sitzt, den [netz.Klient] hält.** Ein
// Lauf von außen käme nur über eine echte Anfrage an ihn heran und prüfte
// dabei den Transport mit; was hier geprüft werden soll, ist die
// **Entscheidung**.
//
// **Und eine exportierte Testtür gibt es nicht**: sie sähe wie eine Bindung
// aus, und ein Wächter zählt Bindungen nach ihrer Aufschrift.

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der Riegel am Transport hält, was der Schalter sagt.**
//
// # Warum er neben dem Vorfilter steht
//
// **Weil ein Schalter sich ändert, nachdem die Naht gebaut ist.** Der
// Vorfilter sieht einmal nach; dieser Riegel bei **jeder** Anfrage — auch bei
// denen, die ein laufender Vorschlag Minuten später noch stellt.
func TestDerRiegelAmTransportHaeltDenSchalter(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	anwendung := KIAnwendungUeber(db).MitKIZuhause(t.TempDir())

	// **Ohne Schalter**: der Riegel weist ab, bevor eine Verbindung entsteht.
	if fehler := anwendung.pruefeKILaufzeit(
		riegelkontext(t), ""); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.KILaufzeit, domain.KILaufzeitAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	// **Mit Schalter und auf der Schleife**: er lässt durch.
	if fehler := anwendung.pruefeKILaufzeit(
		riegelkontext(t), "http://127.0.0.1:11434"); fehler != nil {
		t.Fatalf("der eingeschaltete Riegel wies ab: %v", fehler)
	}

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.RichtlinieNurLokal, domain.NurLokalAn); fehler != nil {
		t.Fatalf("»nur lokal« setzen: %v", fehler)
	}

	// **Unter »nur lokal« hält er eine fremde Adresse auf** — und zwar
	// **bevor** `/api/tags` hinausgeht: schon die Liste der dortigen Modelle
	// wäre eine Auskunft nach draußen.
	if fehler := anwendung.pruefeKILaufzeit(riegelkontext(t),
		"http://ki.example.org:11434"); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	// **Und die Schleife bleibt erlaubt** — die Gegenprobe dazu.
	if fehler := anwendung.pruefeKILaufzeit(
		riegelkontext(t), "http://127.0.0.1:11434"); fehler != nil {
		t.Fatalf("»nur lokal« wies die Schleife ab: %v", fehler)
	}
}

// **Ohne Bestand weist der Riegel ab statt durchzulassen.**
func TestDerRiegelWeistOhneBestandAb(t *testing.T) {
	t.Parallel()

	if fehler := NeueApp("").pruefeKILaufzeit(
		riegelkontext(t), ""); fehler == nil {
		t.Fatal("der Riegel ließ ohne Bestand durch")
	}
}

// bestandFuerDenRiegel öffnet einen frischen Bestand.
//
// **Eine eigene Vorrichtung**, weil `frischerBestand` im **äußeren** Testpaket
// steht: dieser Lauf braucht das Paket von innen.
func bestandFuerDenRiegel(t *testing.T) *sql.DB {
	t.Helper()

	db, fehler := storage.Oeffnen(riegelkontext(t),
		filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if fehler := storage.Migriere(riegelkontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	return db
}

// riegelkontext trägt das Profil, ohne das kein Aufruf etwas findet.
func riegelkontext(t *testing.T) context.Context {
	t.Helper()

	return storage.MitProfil(t.Context(), 1)
}

// **Der Riegel weist auch ab, wenn die Richtlinie nicht zu lesen ist.**
//
// **Nicht „im Zweifel durch"**: was über »nur lokal« nicht feststeht, ist
// kein Grund, hinauszugehen.
func TestDerRiegelWeistBeiUnlesbarerRichtlinieAb(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&halbstolpernde{gutBis: 1})

	if fehler := anwendung.pruefeKILaufzeit(
		riegelkontext(t), ""); fehler == nil {
		t.Fatal("der Riegel ließ bei unlesbarer Richtlinie durch")
	}
}

// **Und ein Katalogruf, dessen zweiter Lesezugriff scheitert, ebenso.**
//
// **Warum der zweite**: `KIModelle` liest die Einstellungen zweimal — einmal
// für die Ortsregel, einmal für Schalter und Adresse. Ein Doppel, das beim
// ersten scheitert, erreicht den zweiten Weg nie.
func TestEinKatalogMitHalbLesbarenEinstellungenMeldetDenFehler(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&halbstolpernde{gutBis: 1})

	if antwort := anwendung.KIModelle(riegelkontext(t)); antwort.Fehler == nil {
		t.Fatal("der Lesefehler kam nicht durch")
	}
}

// halbstolpernde antworten die ersten Male und scheitern danach.
type halbstolpernde struct {
	gutBis  int
	gelesen int
}

func (h *halbstolpernde) Stand(context.Context) (map[string]service.Wahl, error) {
	h.gelesen++

	if h.gelesen > h.gutBis {
		return nil, errors.New("die Einstellungen sind unlesbar")
	}

	return map[string]service.Wahl{}, nil
}

func (h *halbstolpernde) Alle(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}

func (h *halbstolpernde) Setzen(context.Context, string, string) error { return nil }

// **Die Adresse steht im Befund, wenn eine gesetzt ist.**
//
// **Und sonst die Vorgabe** — ein Befund, der die Adresse verschwiege, sagte
// nicht, wo nachzusehen ist.
func TestDerBefundNenntDieGesetzteAdresse(t *testing.T) {
	t.Parallel()

	if gesehen := angezeigteAdresse("http://ki.example.org"); gesehen !=
		"http://ki.example.org" {
		t.Errorf("die Adresse ist %q", gesehen)
	}

	if gesehen := angezeigteAdresse(""); gesehen == "" {
		t.Error("die leere Adresse blieb leer")
	}
}

// **Ohne Zuhause bleibt der Cloudstand ungewiss — und sagt warum.**
//
// **`HOME` leer** ist der einzige Weg, an dem `os.UserHomeDir` scheitert; der
// Lauf setzt es für sich und lässt den Rechner in Ruhe (§3).
func TestOhneZuhauseBleibtDerCloudstandUngewiss(t *testing.T) {
	t.Setenv("HOME", "")

	befund := NeueApp("").cloudstand()

	if befund.Stand != string(ollama.CloudUngewiss) {
		t.Errorf("der Stand ist %q", befund.Stand)
	}

	if befund.Beleg == "" {
		t.Error("der Befund nennt keinen Grund")
	}
}
