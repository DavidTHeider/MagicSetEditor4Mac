package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Fassade übersetzt Fehler in eine für das Frontend stabile Struktur
// (AGENTS.md §8). Stabil heißt: Code und Schlüssel ändern sich nicht, wenn
// jemand einen Satz umformuliert.
func TestUebersetzeOrdnetJedeFehlerartEin(t *testing.T) {
	faelle := []struct {
		name       string
		fehler     error
		code       app.Fehlercode
		schluessel domain.Schluessel
		feld       string
	}{
		{
			name:       "nicht gefunden",
			fehler:     fmt.Errorf("Set laden: %w", domain.ErrNotFound),
			code:       app.CodeNichtGefunden,
			schluessel: domain.SchluesselNichtGefunden,
		},
		{
			// **Nur ein überholter Stand heißt `konflikt`** (ENG-05.AK-03,
			// F08) — ein Konflikt ohne eigene Art nennt seinen Grund.
			name:       "Konflikt",
			fehler:     fmt.Errorf("Karte anlegen: %w", domain.ErrConflict),
			code:       app.CodeKonflikt,
			schluessel: domain.SchluesselKonfliktgrund,
		},
		{
			name:       "überholter Stand",
			fehler:     fmt.Errorf("Karte ändern: %w", domain.ErrUeberholt),
			code:       app.CodeKonflikt,
			schluessel: domain.SchluesselKonflikt,
		},
		{
			// **Ein eigener Schlüssel unter demselben Code.** Die
			// Vorlagenvorschau macht daran einen **leeren Zustand** fest;
			// bekäme er `nichtGefunden`, hinge derselbe Satz an jedem
			// fehlenden Ding. Angestrichen vom Nachlauf des Zweitblicks über
			// Phase 3.3.
			name: "kein Set mit diesem Stil",
			fehler: fmt.Errorf("Probekarte: %w",
				domain.ErrKeinSetMitStil),
			code:       app.CodeNichtGefunden,
			schluessel: domain.SchluesselKeinSetMitStil,
		},
		{
			// **Ein Gegenüber, das nicht horcht** (#2491): der Satz zu
			// `nichtGefunden` riet, die Liste neu zu laden. Der Code bleibt
			// — Kommandozeile, MCP und REST prüfen auf ihn.
			name: "ein Gegenüber, das nicht horcht",
			fehler: fmt.Errorf("an k-anna zustellen: %w", domain.MitGrund(
				domain.ErrPeerNichtErreichbar, errors.New("connection refused"),
				"die Gegenstelle ist nicht erreichbar")),
			code:       app.CodeNichtGefunden,
			schluessel: domain.SchluesselPeerNichtErreichbar,
		},
		{
			// **Und eines ohne Adresse** (#2491, gefunden mit `p10handel09`).
			name: "ein Gegenüber ohne Adresse",
			fehler: fmt.Errorf("%w (8749-8908) — ohne sie findet dieses "+
				"Programm es nur im selben Netz", domain.ErrPeerOhneAdresse),
			code:       app.CodeNichtGefunden,
			schluessel: domain.SchluesselPeerOhneAdresse,
		},
		{
			// **Vor den Fachfehlern** — er umschließt oft deren Meldung, und
			// hinter ihnen einsortiert träfe er sie nie
			// (#886).
			name:       "belegter Bestand",
			fehler:     fmt.Errorf("Transaktion beginnen: %w", domain.ErrBelegt),
			code:       app.CodeBelegt,
			schluessel: domain.SchluesselBelegt,
		},
		{
			// **Und die Probe auf die Reihenfolge.** Ein belegter Bestand kann
			// den Fehler des Vorgangs mitbringen, den er antraf; wer dann
			// `nichtGefunden` meldete, schickte die Oberfläche zum Nachladen —
			// in dieselbe Sperre.
			name: "belegt gewinnt gegen den mitgebrachten Fachfehler",
			fehler: fmt.Errorf("die Karte lesen: %w: %w",
				domain.ErrBelegt, domain.ErrNotFound),
			code:       app.CodeBelegt,
			schluessel: domain.SchluesselBelegt,
		},
		{
			// **Der Nur-Lese-Bestand** — Werkstück 16 von Phase 4.12
			// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
			//
			// **Warum das mehr ist als eine Zuordnung**: ohne ihn käme
			// „attempt to write a readonly database" durch — ein roher
			// Treibertext, den §8 verbietet, und einer, der nicht sagt, dass
			// es an einer **Freigabe** liegt und nicht an der Datei.
			name: "in diesem Profil darf nur gelesen werden",
			fehler: fmt.Errorf("die Karte schreiben: %w: %w",
				domain.ErrNurLesend, errSchreibsperre),
			code:       app.CodeNurLesend,
			schluessel: domain.SchluesselNurLesend,
		},
		{
			// **Der fehlende Schlüssel steht weit vorn und gewinnt gegen den
			// Wortlaut, in den er eingewickelt ist** — Werkstück 14 von Phase
			// 4.12.
			//
			// **Warum das mehr ist als eine Zuordnung**: ohne ihn käme
			// `SQLITE_NOTADB` durch, und das heißt in dieser Anwendung „die
			// Datei ist kein Bestand". Wer das über einer Datei liest, die
			// nicht aufgeht, ersetzt sie — und macht den Verlust endgültig.
			name: "verschlüsselter Bestand ohne Schlüssel",
			fehler: fmt.Errorf("Datenbank %q öffnen: %w", "bestand.db",
				domain.ErrSchluesselFehlt),
			code:       app.CodeSchluesselFehlt,
			schluessel: domain.SchluesselSchluesselFehlt,
		},
		{
			// **Die drei Ausgänge eines Umschlags** (N4, #1457) — jeder mit
			// eigenem Rat, und deshalb drei Fälle statt eines.
			name:       "Umschlag aus anderer Fassung",
			fehler:     fmt.Errorf("lesen: %w", domain.ErrUmschlagFassung),
			code:       app.CodeValidierung,
			schluessel: domain.SchluesselUmschlagFassung,
		},
		{
			name:       "Unterschrift des Umschlags trägt nicht",
			fehler:     fmt.Errorf("lesen: %w", domain.ErrUmschlagUnterschrift),
			code:       app.CodeValidierung,
			schluessel: domain.SchluesselUmschlagUnterschrift,
		},
		{
			name:       "Inhalt passt nicht zum Manifest",
			fehler:     fmt.Errorf("lesen: %w", domain.ErrUmschlagInhalt),
			code:       app.CodeValidierung,
			schluessel: domain.SchluesselUmschlagInhalt,
		},
		{
			// **Eine abgelaufene Anmeldung ist kein Abbruch** (BK-07d,
			// #2118): niemand hat aufgehört — der Rückruf kam nicht, und
			// der Rat ist ein anderer.
			name:       "Anmeldung abgelaufen",
			fehler:     fmt.Errorf("anmelden: %w", domain.ErrAnmeldungAbgelaufen),
			code:       app.CodeAbgebrochen,
			schluessel: domain.SchluesselAnmeldungAbgelaufen,
		},
		{
			// **Ein offener Fernhorchposten ist ein Konflikt mit eigenem
			// Rat** (N3, #1456): nichts zu ändern, sondern erst zu beenden.
			name:       "Fernhorchen läuft schon",
			fehler:     fmt.Errorf("starten: %w", domain.ErrFernhorchenLaeuft),
			code:       app.CodeKonflikt,
			schluessel: domain.SchluesselFernhorchenLaeuft,
		},
		{
			// **Eine Einladung aus einer neueren Fassung ist kein
			// Tippfehler** (N3, #1456). Der allgemeine Satz zu
			// `validierung` riete, die Eingabe zu ändern — hier ist die
			// Eingabe in Ordnung und die Gegenseite neuer.
			name: "Einladung aus einer neueren Fassung",
			fehler: fmt.Errorf("wählen: %w",
				domain.ErrEinladungZuNeu),
			code:       app.CodeValidierung,
			schluessel: domain.SchluesselEinladungZuNeu,
		},
		{
			// **Hier stand bis zum 2026-09-17 `unbekannt`**, und damit zeigte
			// das Fenster „Es ist ein unerwarteter Fehler aufgetreten. Was
			// genau, steht im Protokoll." — an einer Stelle, an der das
			// Protokoll nach eigener Entscheidung leer bleibt
			// (`TestErwarteteFehlerStehenNichtImProtokoll`). Gesehen an der
			// laufenden Anwendung, #1601.
			name:       "Validierung ohne Feldbezug",
			fehler:     fmt.Errorf("Eingabe: %w", domain.ErrValidation),
			code:       app.CodeValidierung,
			schluessel: domain.SchluesselValidierungsgrund,
		},
		{
			name:       "Feldbefund",
			fehler:     domain.NewFieldError("name", domain.SchluesselPflichtfeld, "%s darf nicht leer sein", "Name"),
			code:       app.CodeValidierung,
			schluessel: domain.SchluesselPflichtfeld,
			feld:       "name",
		},
		{
			// **Ein Abbruch ist kein unerwarteter Fehler** — Nachlauf über
			// Phase A: ein abgebrochenes Zurückspielen meldete „Es ist ein
			// unerwarteter Fehler aufgetreten".
			name:       "ein Abbruch",
			fehler:     fmt.Errorf("zurückspielen: %w", context.Canceled),
			code:       app.CodeAbgebrochen,
			schluessel: domain.SchluesselVorgangAbgebrochen,
		},
		{
			name:       "eine abgelaufene Frist",
			fehler:     fmt.Errorf("prüfen: %w", context.DeadlineExceeded),
			code:       app.CodeAbgebrochen,
			schluessel: domain.SchluesselVorgangAbgebrochen,
		},
		{
			name:       "etwas Unbekanntes",
			fehler:     errors.New("SQL logic error: no such table: cards"),
			code:       app.CodeUnbekannt,
			schluessel: domain.SchluesselUnbekannt,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			uebersetzt := app.Uebersetze(t.Context(), fall.fehler)

			if uebersetzt == nil {
				t.Fatal("Uebersetze gab nichts zurück")
			}
			if uebersetzt.Code != fall.code {
				t.Errorf("Code %q, erwartet war %q", uebersetzt.Code, fall.code)
			}
			if uebersetzt.Schluessel != fall.schluessel {
				t.Errorf("Schlüssel %q, erwartet war %q", uebersetzt.Schluessel, fall.schluessel)
			}
			if uebersetzt.Feld != fall.feld {
				t.Errorf("Feld %q, erwartet war %q", uebersetzt.Feld, fall.feld)
			}
		})
	}
}

// Rohe SQLite-Meldungen erreichen die Oberfläche nie (AGENTS.md §8).
//
// Das ist der eine Fall, in dem die Fassade den Satz **wegwirft** statt ihn
// durchzureichen: „SQL logic error: no such table" sagt Nutzenden nichts und
// verrät zugleich den inneren Aufbau.
func TestUebersetzeReichtEinenUnbekanntenSatzNichtDurch(t *testing.T) {
	roh := "SQL logic error: no such table: cards"

	uebersetzt := app.Uebersetze(t.Context(), errors.New(roh))

	if uebersetzt.Meldung == roh {
		t.Errorf("die rohe Meldung %q erreicht die Oberfläche", roh)
	}
	if uebersetzt.Meldung == "" {
		t.Error("die Meldung ist leer — der Rückfall fehlt")
	}
}

// Ein Feldbefund behält seinen Satz: er ist für Nutzende geschrieben, und er
// ist der Rückfall, wenn der Katalog den Schlüssel nicht kennt.
func TestUebersetzeBehaeltDenSatzEinesFeldbefunds(t *testing.T) {
	uebersetzt := app.Uebersetze(t.Context(),
		domain.NewFieldError("name", domain.SchluesselPflichtfeld, "%s darf nicht leer sein", "Name"))

	if uebersetzt.Meldung != "Name darf nicht leer sein" {
		t.Errorf("Meldung %q, erwartet war %q", uebersetzt.Meldung, "Name darf nicht leer sein")
	}
}

func TestUebersetzeGibtFuerKeinenFehlerNichts(t *testing.T) {
	if uebersetzt := app.Uebersetze(t.Context(), nil); uebersetzt != nil {
		t.Errorf("für nil kam %+v zurück, erwartet war nichts", uebersetzt)
	}
}

// Ein Feldbefund, der am Konstruktor vorbei gebaut wurde.
//
// `domain.FieldError` ist ein exportierter Verbund; `NewFieldError` ist der
// gute, aber nicht der einzige Weg. Der Nachlauf des Zweitblicks hat gezeigt,
// dass damit die Zusage offenblieb — geschlossen war nur der Konstruktor. Die
// Fassade ist die Grenze, an der kein Weg vorbeiführt.
func TestUebersetzeNormalisiertEinenUmgangenenSchluessel(t *testing.T) {
	roh := &domain.FieldError{
		Field:   "name",
		Key:     "gibtsNicht",
		Message: "etwas stimmt nicht",
	}

	uebersetzt := app.Uebersetze(t.Context(), roh)

	if uebersetzt.Schluessel != domain.SchluesselUnbekannt {
		t.Errorf("Schlüssel %q, erwartet war %q", uebersetzt.Schluessel, domain.SchluesselUnbekannt)
	}
	if uebersetzt.Meldung != "etwas stimmt nicht" {
		t.Errorf("die Meldung ging verloren: %q", uebersetzt.Meldung)
	}
	if uebersetzt.Feld != "name" {
		t.Errorf("der Feldbezug ging verloren: %q", uebersetzt.Feld)
	}
}

// Ein Feldbefund mit Konfliktschlüssel bekommt den Konfliktcode.
//
// Der Code steuert das Verhalten im Frontend: ein Konflikt braucht ein
// Nachladen, eine Validierung nicht. Ihn am Umstand festzumachen, dass es ein
// Feldbefund ist, verwischte beides.
func TestUebersetzeGibtEinemFeldkonfliktDenKonfliktcode(t *testing.T) {
	uebersetzt := app.Uebersetze(t.Context(), domain.NewFieldError("name", domain.SchluesselKonflikt,
		"Das Feld %q kommt zweimal vor.", "name"))

	if uebersetzt.Code != app.CodeKonflikt {
		t.Errorf("Code %q, erwartet war %q", uebersetzt.Code, app.CodeKonflikt)
	}
	if uebersetzt.Feld != "name" {
		t.Errorf("der Feldbezug ging verloren: %q", uebersetzt.Feld)
	}
}

// errSchreibsperre steht für den rohen Treibertext, den §8 nicht durchlässt.
var errSchreibsperre = errors.New("sqlite3: attempt to write a readonly database")
