// **Nutzer und Profile ändern und löschen** — Werkstück 5 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, damit `nutzerverwaltung.go` unter der Zeilengrenze
// aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, wie ein Nutzer
// **entsteht**, hier, wie er sich **ändert** und wieder verschwindet.

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// NutzerAendern prüft und schreibt fort.
func (v *Nutzerverwaltung) NutzerAendern(ctx context.Context, geaendert domain.Nutzer) (
	domain.Nutzer, error,
) {
	if err := geaendert.Validate(); err != nil {
		return domain.Nutzer{}, err
	}

	return v.ablage.NutzerAendern(ctx, geaendert)
}

// ProfilAendern prüft und schreibt fort.
func (v *Nutzerverwaltung) ProfilAendern(ctx context.Context, geaendert domain.Profil) (
	domain.Profil, error,
) {
	if err := geaendert.Validate(); err != nil {
		return domain.Profil{}, err
	}

	return v.ablage.ProfilAendern(ctx, geaendert)
}

// NutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus — **und samt
// dem, was neben dem Bestand an ihm hängt**.
//
// # Der Befund
//
// **Zweitblick über die Phasen 4.0 bis 4.13** (2026-09-07): hier stand ein
// einziger Ruf an die Ablage, und die löscht eine Datenbankzeile. Was
// **daneben** liegt, blieb liegen:
//
//   - das Nutzerbild, eine Datei unter `nutzer-<id>.png`,
//   - das TOTP-Geheimnis im Schlüsselbund des Rechners.
//
// Die Datenschutzhilfe sagte derweil, ein Nutzer werde „samt allem, was an ihm
// hängt" gelöscht. **Ein persönliches Foto und ein Geheimnis, die eine
// Löschung überleben, sind genau das, wogegen es diese Zusage gibt.**
//
// # Was bewusst bleibt
//
// **Die Profilwurzeln** — Sets, Decks, Sammlung, Turniere. Das ist eine
// Entscheidung und steht so in `storage.Nutzerablage.NutzerLoeschen`; die
// Oberfläche warnt ausdrücklich davor. Der Hilfetext sagt es jetzt auch.
//
// # Warum die Reihenfolge so ist
//
// **Erst die Zeile, dann das Beiwerk.** Scheitert das Löschen — falscher Name,
// letzter Nutzer —, ist nichts angefasst. Scheitert danach das Bild, ist der
// Nutzer fort und eine Datei bleibt: ärgerlich, aber kein halber Nutzer. Der
// umgekehrte Weg ließe ein Bild verschwinden, das noch jemandem gehört.
func (v *Nutzerverwaltung) NutzerLoeschen(
	ctx context.Context, nutzerID int64, name string,
	bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
) error {
	// **Die Profile werden **vorher** gelesen** — Nachlauf zum Zweitblick
	// über die Phasen 4.0 bis 4.13 (2026-09-07), Befund 2.
	//
	// Ihre Abzeichen liegen als `profil-<id>.png` neben dem Nutzerfoto, und
	// nach dem Löschen ist nicht mehr zu erfahren, welche Kennungen ihm
	// gehörten: die Kaskade räumt die Zeilen mit. Wer erst löscht und dann
	// fragt, bekommt eine leere Liste und lässt jedes Abzeichen liegen.
	//
	// **Ein Lesefehler hält das Löschen nicht auf.** Er wird gemeldet wie ein
	// liegengebliebenes Bild — der Mensch will gelöscht werden, und ein
	// Verzeichnis, das gerade klemmt, ist kein Grund, ihn zu behalten.
	profile, lesefehler := v.ablage.Profile(ctx, nutzerID)

	if fehler := v.ablage.NutzerLoeschen(ctx, nutzerID, name, bindung...); fehler != nil {
		return fehler
	}

	return v.beiwerkWegnehmen(ctx, nutzerID, profile, lesefehler)
}

// beiwerkWegnehmen räumt weg, was neben dem Bestand an einem Nutzer hing.
//
// **Ein Fehlschlag wird gemeldet und macht das Löschen nicht rückgängig** (§7,
// Regel 3): der Nutzer ist fort, und das bleibt er. Wer die Meldung liest,
// weiß, dass eine Datei nachzusehen ist.
//
// **Fünf Dinge** — Foto, Tischbild, Fensterbild, Abzeichen je Profil,
// TOTP-Geheimnis. Das Abzeichen kam im Nachlauf dazu (Befund 2); `profile`
// und `lesefehler` stammen aus dem Ruf **vor** dem Löschen.
//
// **Tisch- und Fensterbild kamen zuletzt** — Befund G4 des Zweitblicks über
// Phase 6.3 (#1666).
// Sie entstanden mit G3 von Phase 5.5 und mit UX-09, liegen unter `nutzer/`
// neben dem Foto, und **dieselbe Klasse wie beim Abzeichen** traf sie ein
// drittes Mal: hier stand nur das Foto. Dass eine vierte Art nicht still
// entsteht, hält `nutzerhintergruende_test.go`.
func (v *Nutzerverwaltung) beiwerkWegnehmen(
	ctx context.Context, nutzerID int64,
	profile []domain.Profil, lesefehler error,
) error {
	var offen []string

	if lesefehler != nil {
		offen = append(offen,
			"die Profile ließen sich vorher nicht lesen, ihre Abzeichen "+
				"bleiben liegen: "+lesefehler.Error())
	}

	if v.bilder != nil {
		for _, bild := range []struct {
			name      string
			wegnehmen func(int64) error
		}{
			{"das Bild", v.bilder.NutzerbildWegnehmen},
			{"das Tischbild", v.bilder.TischbildWegnehmen},
			{"das Fensterbild", v.bilder.FensterbildWegnehmen},
		} {
			if fehler := bild.wegnehmen(nutzerID); fehler != nil {
				offen = append(offen, bild.name+": "+fehler.Error())
			}
		}

		for _, profil := range profile {
			if fehler := v.bilder.ProfilbadgeWegnehmen(profil.ID); fehler != nil {
				offen = append(offen, fmt.Sprintf(
					"das Abzeichen von Profil %d: %v", profil.ID, fehler))
			}
		}
	}

	if v.geheimnisse != nil {
		if fehler := v.geheimnisse.Nimm(ctx, nutzerID); fehler != nil {
			offen = append(offen, "das Geheimnis des zweiten Faktors: "+fehler.Error())
		}
	}

	if len(offen) > 0 {
		return domain.MitGrund(domain.ErrValidation,
			errors.New(strings.Join(offen, "; ")),
			"der Nutzer ist gelöscht, aber nicht alles zu ihm ließ sich "+
				"wegräumen — sieh im Anwendungsordner nach")
	}

	return nil
}

// ProfilLoeschen nimmt ein Profil heraus.
//
// # Der Zeiger wandert mit
//
// **Wer das Profil löscht, in dem er gerade arbeitet**, stünde danach auf
// einem Profil, das es nicht mehr gibt — und
// [storage.Nutzerablage.Profilzeiger] fiele auf das **erste** zurück, ohne
// dass es jemand veranlasst hätte. Deshalb wird hier ausdrücklich umgestellt,
// **bevor** gelöscht wird: was der Zeiger sagt, ist dann immer wahr.
func (v *Nutzerverwaltung) ProfilLoeschen(
	ctx context.Context, profilID int64, name string,
	bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
) error {
	gezeigt, lesefehler := v.zeiger.Profilzeiger(ctx)
	if lesefehler != nil {
		return lesefehler
	}

	// **Der Ausweich wird gesucht, aber nicht gesetzt** — Befund 1 des
	// Zweitblicks über die Phasen 4.0 bis 4.13 (2026-09-07).
	//
	// Hier stand ein `ProfilzeigerSetzen` vor dem Löschen. Wies die Ablage das
	// Löschen danach ab — falscher Name, letztes Profil des Nutzers —, war der
	// Zeiger trotzdem bewegt: der Mensch stand in einem anderen Profil, ohne
	// etwas gelöscht zu haben.
	//
	// **Die Ablage setzt ihn jetzt in derselben Transaktion**, in der sie
	// löscht. Null heißt „der Zeiger steht nicht auf diesem Profil".
	var ausweich int64

	if gezeigt == profilID {
		gefunden, fehler := v.anderesProfil(ctx, profilID)
		if fehler != nil {
			return fehler
		}

		ausweich = gefunden
	}

	if fehler := v.ablage.ProfilLoeschen(
		ctx, profilID, name, ausweich, bindung...); fehler != nil {
		return fehler
	}

	// **Und das Abzeichen geht mit** — Nachlauf zum Zweitblick über die
	// Phasen 4.0 bis 4.13 (2026-09-07), Befund 2.
	//
	// Dieselbe Klasse wie beim Nutzerfoto und dieselbe Reihenfolge: erst die
	// Zeile, dann die Datei. Ein Abzeichen, das ein Löschen überlebt, taucht
	// unter der nächsten Profilkennung wieder auf, die zufällig dieselbe Zahl
	// bekommt.
	if v.bilder == nil {
		return nil
	}

	if fehler := v.bilder.ProfilbadgeWegnehmen(profilID); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"das Profil ist gelöscht, sein Abzeichen ließ sich nicht "+
				"wegräumen — sieh im Anwendungsordner nach")
	}

	return nil
}

// anderesProfil gibt irgendein anderes Profil.
//
// **Irgendeines und nicht das des Nutzers**: der Zeiger sagt, wo gearbeitet
// wird, und nach dem Löschen muss er irgendwo hinzeigen. Welches es ist, sieht
// der Mensch sofort — die Seite nennt es.
func (v *Nutzerverwaltung) anderesProfil(ctx context.Context, ausser int64) (int64, error) {
	zeilen, fehler := v.Uebersicht(ctx)
	if fehler != nil {
		return 0, fehler
	}

	for _, zeile := range zeilen {
		for _, profil := range zeile.Profile {
			if profil.ID != ausser {
				return profil.ID, nil
			}
		}
	}

	// **Es gibt kein anderes** — dann bleibt auch dieses (der Speicher weist
	// es ab), und diese Meldung sagt es früher und deutlicher.
	return 0, fmt.Errorf(
		"%w: es gibt kein anderes Profil, in das gewechselt werden könnte",
		domain.ErrConflict)
}
