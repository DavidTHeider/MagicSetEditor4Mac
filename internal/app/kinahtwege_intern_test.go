package app

// **Die Fehlerwege beim Bauen der Naht** — Werkstück A7 von Phase 5.3
// (#1446),
// abgetrennt mit A8
// (#1447),
// weil `kifehlerwege_intern_test.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die Wege, die ein **Ruf** nehmen kann; hier die, an denen die
// **Verdrahtung** scheitert — ein fehlendes Zuhause und eine Richtlinie, die
// erst beim Bauen des Werks unlesbar wird.

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Zuhause lässt sich der Schreibweg nicht bauen — und das bleibt
// stehen.**
//
// **`HOME` leer** ist der einzige Weg, an dem `Paketverzeichnis` scheitert;
// der Lauf setzt es für sich und lässt den Rechner in Ruhe (§3).
//
// **Der Fehler wird gemerkt und nicht verschluckt**: sonst käme beim ersten
// Vorschlag „es ist kein Bestand offen" statt dessen, was wirklich fehlt (§8).
func TestOhneZuhauseBleibtDerNahtfehlerStehen(t *testing.T) {
	t.Setenv("HOME", "")

	anwendung := KIAnwendungUeber(bestandFuerDenRiegel(t))

	if anwendung.kinahtfehler == nil {
		t.Fatal("der Nahtfehler wurde verschluckt")
	}

	if _, _, fehler := anwendung.kiwerkzeug(riegelkontext(t), ollamaanbieter, kideckung{}); fehler == nil {
		t.Fatal("das Werkzeug kam trotz Nahtfehler")
	}
}

// einModellFuerDenBauplan gibt ein gültiges Modell.
func einModellFuerDenBauplan() []domain.KIModell {
	return []domain.KIModell{{
		Alias: "gemma4:8b", Anbieter: "ollama", Laufzeit: "ollama 0.12.3",
		Ort: domain.KIOrtLokal,
		Geprueft: []domain.KIGeprueft{{
			Faehigkeit: domain.KITextgenerierung, Datum: "2026-09-11",
			Beleg: "Ollama meldet completion",
		}},
	}}
}

// **Und eine Richtlinie, die beim Bauen des Werks scheitert, kommt durch.**
//
// **Der dritte Lesezugriff**: `frageAnbieter` liest Schalter und Adresse,
// `werkzeugAus` danach die Ortsregel. Scheitert sie dort, gibt es kein Werk —
// und nicht eines, das „nur lokal" nicht kennt.
func TestEineSpaeteRichtlinieKommtAlsFehlerDurch(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&eingeschalteteDannStolpernde{})
	anwendung.MitKILaufzeit(func(string) KILaufzeitquelle {
		return festeLaufzeit{modelle: einModellFuerDenBauplan()}
	})

	staende := []anbieterstand{{
		name: ollamaanbieter, einrichtung: KIEinrichtungBereit,
		modelle: einModellFuerDenBauplan(),
	}}

	if _, _, werkfehler := anwendung.werkzeugAus(
		riegelkontext(t), staende, ollamaanbieter, kideckung{}); werkfehler == nil {
		t.Fatal("die unlesbare Richtlinie kam nicht durch")
	}
}

// **Die beiden Durchreichen der Feldherkunft reichen durch** — A11, #1450.
//
// # Warum sie geprüft werden, obwohl sie nichts tun
//
// **Weil „nichts tun" hier die Zusage ist.** Beide sind Nähte: `feldnaht` für
// das Fenster, `mcpnaht` für die Flächen ohne. Eine Entscheidung in ihnen
// wäre eine, die kein Lauf sieht — und ein vertauschtes Argument fiele erst
// auf, wenn eine Herkunft unter der falschen Karte stünde.
func TestDieDurchreichenDerFeldherkunftReichenDurch(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	anwendung := KIAnwendungUeber(db).MitMarktbund(erfundenerBund{})

	karteID := eineKarteImBestand(t, db)

	herkunft := []domain.KIFeldherkunft{{
		Feld: "name", Anbieter: "ollama", Modell: "gemma4:8b-q4",
		Aufgabe: "x", Auftrag: "ki-1", Zeitpunkt: time.Now().UTC(),
		Bezug: domain.KIInhaltsbezug([]string{"Blitzschlag"}),
	}}

	werte := []domain.Value{{Field: "name", Text: "Blitzschlag"}}

	// **Die Naht des Fensters** — sie kennt kein Set und prüft keines.
	if _, _, fehler := anwendung.schreiber.FelderVonAussenMitHerkunft(
		riegelkontext(t), 1, karteID, werte, "ki", herkunft, nil, ""); fehler != nil {
		t.Fatalf("die Fensternaht: %v", fehler)
	}

	gelesen := anwendung.KIFeldherkunft(riegelkontext(t), karteID)
	if gelesen.Fehler != nil || len(gelesen.Herkunft) != 1 {
		t.Fatalf("gelesen wurde %+v", gelesen)
	}

	// **Und derselbe Weg mit Bedingung** (Befund 2 des Zweitblicks über
	// Phase 5.3). Beide Nähte reichen nur durch; dass sie es **richtig** tun,
	// sieht man erst, wenn die Bedingung greift — deshalb steht hier ein
	// Stand, den die Karte nicht hat.
	if _, _, fehler := anwendung.schreiber.FelderVonAussenMitHerkunft(
		riegelkontext(t), 1, karteID, werte, "ki", herkunft, nil,
		"ein-stand-von-gestern"); fehler == nil {
		t.Fatal("die Fensternaht schrieb gegen einen fremden Stand")
	}
}

// eineKarteImBestand legt ein Set mit einer Karte an und gibt deren Kennung.
func eineKarteImBestand(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(riegelkontext(t),
		&domain.Set{
			Name: "Probeset", Game: "magic",
			Cards: []*domain.Card{{
				Values: []domain.Value{{Field: "name", Text: "Alt"}},
			}},
		}, nil)
	if fehler != nil {
		t.Fatalf("das Set anlegen: %v", fehler)
	}

	karten, fehler := storage.NeuerSetLeser(db).Karten(riegelkontext(t), setID)
	if fehler != nil {
		t.Fatalf("die Karten lesen: %v", fehler)
	}

	return karten[0].ID
}
