package main

// **Der Bestand der Kommandozeile** — abgetrennt von `lauf.go`, als die Datei
// mit der Profilklemme (#1351) über die Zeilengrenze aus §2 kam.
//
// **Nach der Frage geschnitten**: nebenan steht, **welcher** Befehl gemeint
// ist; hier steht, **womit** er arbeitet — Bestand auf, Wanderung, Profil,
// Bestand zu.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitBestand öffnet den Bestand, migriert ihn und schließt ihn wieder.
//
// **Wie die beiden anderen Roots**, und mit derselben Sicherung davor.
func mitBestand(
	ctx context.Context, ort string, an *schreiber,
	arbeit func(context.Context, *sql.DB) int,
) int {
	return mitBestandWie(ctx, ort, an, false, arbeit)
}

// mitBestandWie ist [mitBestand], und `ueberNeuerem` lässt einen Befehl auch
// über einem Bestand einer **neueren** Fassung arbeiten — für Befehle, die
// ihn ersetzen oder nur eine Sicherung ansehen (M02, #2346): sie scheiterten
// an der Wanderung, die diese Fassung nicht kennt. **Nur dann** fällt die
// Wanderung weg; ein leerer oder älterer Bestand wandert wie immer — sonst
// fragten Profil und Rolle in einem Schema, das es noch nicht gibt, und der
// neue Rechner mit der alten Sicherung scheiterte an der ersten Abfrage
// (Zweitblick über Phase 2, Befund 7).
func mitBestandWie(
	ctx context.Context, ort string, an *schreiber, ueberNeuerem bool,
	arbeit func(context.Context, *sql.DB) int,
) int {
	// **Das Verzeichnis anlegen, nicht voraussetzen.** Wer die Kommandozeile
	// benutzt, ohne das Fenster je geöffnet zu haben, hat noch keines.
	if fehler := os.MkdirAll(filepath.Dir(ort), 0o700); fehler != nil {
		an.klageUeber(fmt.Errorf("Bestandsverzeichnis anlegen: %w", fehler))

		return CodeFachlich
	}

	// **Und wenn er nicht aufgeht, steht daneben, wie es weitergeht** —
	// Werkstück 5 von Phase 4.13 (#1352). Dieselbe Auskunft wie auf der
	// Wiederherstellungsseite des Fensters, nur als Zeile.
	db, err := storage.Oeffnen(ctx, ort)
	if err != nil {
		an.klageUeber(app.MitWegZurueck(ctx, ort, err))

		return CodeFachlich
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // beim Beenden ist nichts mehr zu retten

	// **Und gesichert wird auch hier** — der dritte Root, dieselbe Zusage
	// (Zweitblick über Phase S).
	neuer := ueberNeuerem && ausNeuererFassung(ctx, db)

	var wanderung error
	if !neuer {
		wanderung = storage.Migriere(ctx, db, app.SicherungVorMigration(db, filepath.Dir(ort), app.Fassung))
	}

	if wanderung != nil {
		an.klageUeber(app.MitWegZurueck(ctx, ort, wanderung))

		return CodeFachlich
	}

	// **Über einem neueren Bestand ohne Profil und Rolle**: sein Schema kennt
	// diese Fassung nicht, also auch nicht seine Profile. Wer hier tippt, hält
	// die Dateien ohnehin in der Hand; eine Rolle schützte nichts, was er
	// nicht auch im Finder täte. Gefragt wird wieder am zurückgespielten,
	// gewanderten Bestand.
	if neuer {
		return arbeit(app.MitDienstmodus(ctx, domain.ModusKommandozeile), db)
	}

	// **Und das Profil hängt an den Wurzelkontext** — jede Wurzelabfrage
	// klemmt darauf (#1351). Ohne die Zeile liefe jeder Befehl auf denselben
	// Fehler, und keiner sagte, woran es liegt.
	ctx, profilfehler := profilkontext(ctx, db)
	if profilfehler != nil {
		an.klageUeber(profilfehler)

		return CodeFachlich
	}

	// **Und die Fläche nennt sich** — Werkstück O1 von Phase 5.2 (#1425):
	// ein Nachweis wird nur am Fenster ausgestellt, und der Akteur eines
	// Befehls trägt, dass er von hier kam.
	ctx = app.MitDienstmodus(ctx, domain.ModusKommandozeile)

	// **Und wer hier nur lesen darf, bekommt einen Bestand, der nichts
	// annimmt** — Werkstück 16 von Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// # Warum an dieser Stelle
	//
	// **Weil hier die einzige Verbindung entsteht, die ein Befehl je sieht.**
	// Ein Befehl öffnet, arbeitet und endet; er wechselt das Profil höchstens
	// selbst, und dann ist der nächste Ruf ein neuer Prozess.
	//
	// **Nach der Migration**, denn `Migriere` beginnt unbedingt eine
	// Transaktion — und die scheitert nur lesend (gemessen am 2026-09-06).
	// **Leute-Befehle bleiben an der schreibenden Verbindung** — der Grund
	// steht bei `alsLeutebefehl`.
	if istLeutebefehl(ctx) {
		return arbeit(ctx, db)
	}

	nurLesend, lesefehler := app.NurLesendGefragt(ctx, db)
	if lesefehler != nil {
		an.klageUeber(lesefehler)

		return CodeFachlich
	}

	if !nurLesend {
		return arbeit(ctx, db)
	}

	zweite, oeffnungsfehler := storage.OeffnenNurLesend(ctx, ort)
	if oeffnungsfehler != nil {
		an.klageUeber(oeffnungsfehler)

		return CodeFachlich
	}

	defer func() { _ = zweite.Close() }() //nolint:errcheck // beim Beenden ist nichts mehr zu retten

	// **Die erste bleibt offen** — sie schreibt den Profilzeiger, wenn dieser
	// Befehl ein Wechsel ist. Dieselbe Ausnahme wie am Fenster, und dort steht
	// sie begründet (`internal/app/verdrahtung.go`).
	return arbeit(ctx, zweite)
}

// ausNeuererFassung sagt, ob der Bestand eine Wanderung führt, die diese
// Fassung nicht kennt.
func ausNeuererFassung(ctx context.Context, db *sql.DB) bool {
	var neuer *domain.NeuereFassung

	return errors.As(storage.PruefeWanderungsstand(ctx, db), &neuer)
}

// feldbefehl prüft die Argumente von `feld` und führt ihn aus.
//
// **Eine eigene Funktion**, weil vier Argumente vier Wachen brauchen und
// `lauf` sonst zur Sammelstelle würde (§2).
func feldbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 4 {
		an.klage("feld braucht Setkennung, Kartenkennung, Feldname und Wert.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[1])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return feldSetzen(ctx, db, an, setID, karteID, rest[2], rest[3])
	})
}

// wunschschluessel trägt `--profil` durch den Kontext.
type wunschschluessel struct{}

// mitWunschprofil merkt sich, was `--profil` verlangt hat.
//
// **Im Kontext und nicht in jeder Signatur** — derselbe Grund wie beim Profil
// selbst (`storage.MitProfil`): der Kontext fließt ohnehin durch jeden Befehl,
// und ein zweiter Weg wäre derselbe Wert an zwei Stellen.
func mitWunschprofil(ctx context.Context, profil string) context.Context {
	if profil == "" {
		return ctx
	}

	return context.WithValue(ctx, wunschschluessel{}, profil)
}

// profilkontext stattet den Kontext mit dem Profil aus, das gilt.
//
// **`--profil` schlägt den Zeiger, und es verstellt ihn nicht.** Das Issue
// verlangt beides: der gewählte Stand überlebt den Aufruf, **und** ein Skript
// darf sich davon unabhängig machen. Wer beim Wählen auch schriebe, hätte den
// zweiten Satz nicht.
func profilkontext(ctx context.Context, db *sql.DB) (context.Context, error) {
	wunsch, genannt := ctx.Value(wunschschluessel{}).(string)
	if !genannt {
		return storage.MitGezeigtemProfil(ctx, db)
	}

	kennung, fehler := strconv.ParseInt(wunsch, 10, 64)
	if fehler != nil {
		return ctx, fmt.Errorf(
			"%w: --profil braucht eine Kennung aus »mse nutzer«, nicht %q",
			domain.ErrValidation, wunsch)
	}

	// **Geprüft und nicht geglaubt.** Eine Kennung, die es nicht gibt, machte
	// jede Abfrage leer — der Bestand sähe aus, als wäre er weg (§8).
	//
	// **Über die Nutzerliste und nicht über eine eigene Fähigkeit**: sie führt
	// jedes Profil, sie klemmt selbst auf keines (Nutzer gehören keinem
	// Profil), und eine Bindung „gibt es dieses Profil" wäre eine fünfte
	// Fläche für eine Frage, die niemand sonst stellt.
	if !profilVorhanden(ctx, db, wunsch) {
		return ctx, fmt.Errorf(
			"%w: das Profil %s führt der Bestand nicht", domain.ErrNotFound, wunsch)
	}

	return storage.MitProfil(ctx, kennung), nil
}

// profilVorhanden sagt, ob es ein Profil dieser Kennung gibt.
func profilVorhanden(ctx context.Context, db *sql.DB, profilID string) bool {
	antwort := app.NutzerflaecheUeber(db).Nutzer(ctx)
	if antwort.Fehler != nil {
		return false
	}

	for _, einer := range antwort.Nutzer {
		for _, eines := range einer.Profile {
			if eines.ID == profilID {
				return true
			}
		}
	}

	return false
}
