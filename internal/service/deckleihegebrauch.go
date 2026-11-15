package service

// **Eine Leihe einlösen, zurückgeben und auslaufen lassen** — Werkstück G6 von
// Phase 5.5 (#1463).
//
// # Warum das neben dem Ausstellen steht und nicht darin
//
// **Weil es die andere Seite ist.** `deckleihe.go` beantwortet „wie entsteht
// eine Leihe und wie prüfe ich sie"; hier steht, was ein **Entleiher** damit
// tut — und das ist die Seite, an der die Zusagen des Issues hängen: kein
// Start ohne Prüfung, kein zweiter Start aus derselben Leihe, kein heimliches
// Wegnehmen aus einer laufenden Partie.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Vorlegen prüft eine Leihe und gibt den Abzug her, mit dem gestartet wird.
//
// # Die Reihenfolge ist die Zusage
//
// **Erst der eigene Stand, dann das Dokument, dann der Fall.** Ein Widerruf,
// den dieser Rechner kennt, schlägt jede Unterschrift; was danach kommt, ist
// die Prüfung, die auch ein fremder Rechner machen könnte.
//
// **Und der Verbrauch wird eingetragen, bevor der Abzug herausgeht.** Wer ihn
// danach einträge, hätte zwischen Ausgabe und Eintrag ein Fenster, in dem
// dieselbe Leihe zwei Partien trägt.
func (s *Leihstelle) Vorlegen(
	ctx context.Context, kennung string, bindung domain.Leihbindung,
	recht domain.Leihrecht,
) (domain.Partiedeck, error) {
	stand, lesefehler := s.ablage.Leihe(ctx, kennung)
	if lesefehler != nil {
		return domain.Partiedeck{}, lesefehler
	}

	// **Eine eigene Leihe legt niemand vor.** Wer sie ausgestellt hat, hat
	// das Deck; er braucht keine Erlaubnis dafür, und eine, die er sich
	// selbst gäbe, prüfte nichts.
	if stand.Verliehen {
		return domain.Partiedeck{}, fmt.Errorf(
			"%w: diese Leihe hat dieses Profil ausgestellt — wer verleiht, "+
				"spielt mit seinem eigenen Deck", domain.ErrConflict)
	}

	if stand.Widerrufen {
		return domain.Partiedeck{}, fmt.Errorf(
			"%w: diese Leihe ist zurückgenommen", domain.ErrConflict)
	}

	if fehler := s.pruefeVerbrauch(stand, bindung); fehler != nil {
		return domain.Partiedeck{}, fehler
	}

	if fehler := PruefeLeihe(stand.Leihe); fehler != nil {
		return domain.Partiedeck{}, fehler
	}

	ich, fehler := s.ausweis.Ausweis(ctx)
	if fehler != nil {
		return domain.Partiedeck{}, fehler
	}

	if fehler := stand.Leihe.Gilt(s.uhr(), string(ich.Kennung()),
		stand.Abzug.Abdruck, bindung, recht); fehler != nil {
		return domain.Partiedeck{}, fehler
	}

	if stand.Verbraucht == "" {
		stand.Verbraucht = bindung.Woran

		if fehler := s.ablage.LeiheSchreiben(ctx, stand); fehler != nil {
			return domain.Partiedeck{}, fehler
		}
	}

	// **Seine Kartenkennungen sind Zeilen des Verleihers** (P9, #2377,
	// Vermerk 0198) — hier nachgeschlagen fänden sie andere Karten.
	abzug := stand.Abzug
	abzug.Quelle = stand.Leihe.Verleiher

	return abzug, nil
}

// pruefeVerbrauch hält eine eingelöste Leihe auf ihrer Partie fest.
//
// **Derselbe Start noch einmal geht.** Ein Neustart nach einem Absturz ist
// kein zweiter Start — und eine Leihe, die daran scheiterte, machte aus einer
// Störung einen Verlust.
func (s *Leihstelle) pruefeVerbrauch(
	stand domain.Leihstand, bindung domain.Leihbindung,
) error {
	if stand.Verbraucht == "" || stand.Verbraucht == bindung.Woran {
		return nil
	}

	return fmt.Errorf(
		"%w: diese Leihe ist an %q eingelöst — eine zweite Partie bekommt "+
			"sie nicht", domain.ErrConflict, stand.Verbraucht)
}

// Zurueckgeben trägt ein, dass diese Leihe nicht mehr gilt.
//
// # Warum der Entleiher das selbst tut
//
// **Weil der Widerruf über die Leitung kommt und nicht über den Bestand.** Der
// Verleiher nimmt sie bei sich zurück; damit sie hier wirkt, muss die Nachricht
// ankommen — und **diese** Naht trägt sie ein. Wer sie ausließe, hätte einen
// Widerruf, der nur beim Verleiher steht.
//
// **Und wer selbst zurückgibt, tut dasselbe.** „Ich brauche es nicht mehr" und
// „du bekommst es nicht mehr" enden am selben Ort; sie zu trennen hieße, zwei
// Wege zu demselben Stand zu führen.
func (s *Leihstelle) Zurueckgeben(ctx context.Context, kennung string) error {
	stand, fehler := s.ablage.Leihe(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	stand.Widerrufen = true

	return s.ablage.LeiheSchreiben(ctx, stand)
}

// Fortdauer sagt, ob eine **laufende** Partie mit dieser Leihe weiterläuft.
//
// # Die Zusage, für die es diesen Weg gibt
//
// **#1463**: „Ablauf/Widerruf verhindert neue Starts und betrifft laufende
// Nutzung nur nach vorab vereinbartem Modus, **niemals heimliches Entfernen
// von Spielobjekten**."
//
// **Deshalb gibt es einen Grund und nicht nur ein Ja.** Was hier
// zurückkommt, steht im Fenster: eine Partie, die endet, endet mit einem Satz,
// den beide lesen können — und eine, die weiterläuft, sagt, dass die Leihe
// darunter abgelaufen ist.
func (s *Leihstelle) Fortdauer(
	ctx context.Context, kennung, woran string,
) (bleibt bool, grund string, fehler error) {
	stand, fehler := s.ablage.Leihe(ctx, kennung)
	if fehler != nil {
		return false, "", fehler
	}

	if stand.Verbraucht != woran {
		return false, "", fmt.Errorf(
			"%w: diese Leihe trägt %q und nicht %q",
			domain.ErrValidation, stand.Verbraucht, woran)
	}

	abgelaufen := !s.uhr().Before(stand.Leihe.Bis)

	if !stand.Widerrufen && !abgelaufen {
		return true, "die Leihe gilt weiter", nil
	}

	if stand.Leihe.Modus == domain.LaufZuEnde {
		return true, "die Leihe ist zu Ende, diese Partie läuft aus", nil
	}

	return false, "die Leihe ist zu Ende, und der vereinbarte Modus ist " +
		"der Abbruch", nil
}
