package storage

// **Was angekommen ist** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Warum das Ergebnis mitgespeichert wird
//
// **Weil eine neu gerechnete Antwort nicht dieselbe Antwort ist.** Sie ist
// eine zweite, die zufällig gleich aussieht — und sobald sich etwas am Stand
// geändert hat, sieht sie nicht mehr gleich aus. Das Issue verlangt
// ausdrücklich: „zweite Zustellung gibt dieselbe Antwort auch nach Neustart".
//
// # Warum ein zweites Vermerken abgewiesen wird
//
// **Weil es den ersten Eintrag verlöre.** Ein `INSERT OR REPLACE` gäbe beim
// dritten Mal eine dritte Antwort — und niemand erführe, dass es zwei gab.
// Wer wissen will, ob schon etwas dasteht, schlägt nach; wer trotzdem
// schreibt, bekommt einen Konflikt.

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Briefkastenablage führt, was schon behandelt ist.
type Briefkastenablage struct{ db *sql.DB }

// NeueBriefkastenablage baut sie über einer Datenbank.
func NeueBriefkastenablage(db *sql.DB) *Briefkastenablage {
	return &Briefkastenablage{db: db}
}

// Vermerken hält eine behandelte Nachricht samt ihrer Antwort fest.
//
// **Die Uhr entsteht hier und nicht im Aufrufer**: sie soll die eigene sein
// und nicht die des Absenders, und dessen steht in keiner Unterschrift.
func (b *Briefkastenablage) Vermerken(
	ctx context.Context, vorgang string, eintrag domain.Inboxeintrag,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	_, fehler = b.db.ExecContext(ctx,
		`INSERT INTO handelsbriefkasten
		   (profil_id, schluessel, abdruck, ergebnis, vorgang)
		 VALUES (?, ?, ?, ?, ?)`,
		profilID, eintrag.Schluessel, eintrag.Nutzlastabdruck,
		eintrag.Ergebnis, vorgang)

	return uebersetze(fehler, "eine behandelte Nachricht vermerken")
}

// Zustellen schreibt den Vermerk **und** seine Wirkung in einem Zug.
//
// # Warum beides in einer Transaktion steht
//
// **Weil sonst eines von beidem allein dastünde.** Ein Vermerk ohne Wirkung
// hieße „schon behandelt", obwohl nichts geschehen ist — die Nachricht käme
// nie wieder, und der Tausch bliebe stehen. Eine Wirkung ohne Vermerk hieße,
// dass die nächste Zustellung sie **noch einmal** ausführt. Das Issue
// verlangt beides in derselben lokalen Transaktion, und hier ist sie.
//
// # Und warum der zweite Zug am Schlüssel scheitert
//
// **Weil das der Wettlauf ist.** Zwischen Nachschlagen und Schreiben kann
// eine zweite Zustellung liegen; der Primärschlüssel des Briefkastens fängt
// sie, und der Aufrufer liest danach nach, was dasteht — statt zu raten.
//
// # Und warum `vorher` dabei ist
//
// **Weil der Dienst gegen einen Stand entschieden hat** (Befund Z1 des
// Zweitblicks über Phase 6.8, #1669): geschrieben wird nur, wenn er noch
// gilt, sonst [domain.ErrUeberholt] — und der Vermerk fällt mit.
func (b *Briefkastenablage) Zustellen(
	ctx context.Context, eintrag domain.Inboxeintrag,
	vorher, angebot domain.Handelsangebot, wirkung domain.Zusagenwirkungsart,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	zug, fehler := b.db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "eine Zustellung beginnen")
	}

	defer func() { _ = zug.Rollback() }() //nolint:errcheck // nach dem Commit ohne Wirkung

	if _, fehler := zug.ExecContext(ctx,
		`INSERT INTO handelsbriefkasten
		   (profil_id, schluessel, abdruck, ergebnis, vorgang)
		 VALUES (?, ?, ?, ?, ?)`,
		profilID, eintrag.Schluessel, eintrag.Nutzlastabdruck,
		eintrag.Ergebnis, angebot.Vorgang); fehler != nil {
		return uebersetze(fehler, "eine Zustellung vermerken")
	}

	if fehler := revisionFortschreiben(ctx, zug, profilID, vorher, angebot); fehler != nil {
		return fehler
	}

	// **Und die Wirkung an den Zusagen — im selben Zug** (Befund 3 des
	// Nachlaufs zum Zweitblick über Phase 5.6). Geschah sie vorher für
	// sich, blieb nach einem gescheiterten Schreiben eine Zusage als
	// „übergeben" stehen, während der Vorgang noch auf `vorbereitet` stand
	// — und `Freigeben` gab sie nie wieder frei.
	if fehler := zusagenwirkungInTx(
		ctx, zug, profilID, angebot.Vorgang, wirkung); fehler != nil {
		return fehler
	}

	return uebersetze(zug.Commit(), "eine Zustellung abschließen")
}

// Nachschlagen gibt den Eintrag zu einem Schlüssel — **oder nichts**.
//
// **`nil` und kein `ErrNotFound`**: „noch nie gesehen" ist der gewöhnliche
// Fall einer Zustellung und kein Fehler. Ein Fehler zwänge jeden Aufrufer,
// ihn wegzuprüfen, und irgendwann prüfte einer ihn falsch weg.
func (b *Briefkastenablage) Nachschlagen(
	ctx context.Context, schluessel string,
) (*domain.Inboxeintrag, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	var (
		eintrag   domain.Inboxeintrag
		behandelt string
	)

	eintrag.Schluessel = schluessel

	fehler = b.db.QueryRowContext(ctx,
		`SELECT abdruck, ergebnis, behandelt
		   FROM handelsbriefkasten
		  WHERE profil_id = ? AND schluessel = ?`,
		profilID, schluessel).
		Scan(&eintrag.Nutzlastabdruck, &eintrag.Ergebnis, &behandelt)

	if errors.Is(fehler, sql.ErrNoRows) {
		return nil, nil //nolint:nilnil // „nichts" ist hier kein Fehler
	}

	if fehler != nil {
		return nil, uebersetze(fehler, "eine Nachricht nachschlagen")
	}

	eintrag.Behandelt, fehler = time.Parse(time.DateTime, behandelt)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Uhr eines Eintrags lesen")
	}

	return &eintrag, nil
}
