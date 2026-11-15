// **Wer wohin darf** — die Erlaubnisfragen der Nutzerverwaltung.
//
// # Warum sie neben `profilwechsel.go` stehen und nicht darin
//
// **Weil es zwei Fragen sind** (§2): dort steht, **wie** der Zeiger sich
// bewegt; hier, **ob** er das darf. Die Zeilengrenze war der Anlass —
// Werkstück 16 von Phase 4.12 hat die Datei über 400 Zeilen gebracht —, der
// Schnitt folgt der Frage.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zutritt lässt den Wechsel zu, wenn sich der Mensch als Eigner **ausweist**
// oder eine Freigabe **hat**.
//
// # Warum das ein Oder ist und kein Und
//
// **Weil ein Und jeden zweiten Nutzer aussperrte, und zwar für immer.** Diese
// Anwendung hat keinen Begriff einer Sitzung neben dem Profilzeiger: wer
// arbeitet, ist der Eigner des gezeigten Profils. Ein unbedingtes „du brauchst
// eine Freigabe" heißt dann, dass Bewegung nur entlang vergebener Kanten
// möglich ist — und die gibt es am Anfang nicht. Ein frisch angelegter Nutzer
// war unerreichbar, und wer sein Profil einmal verließ, kam nicht zurück.
//
// **Nachgemessen und nicht überlegt**: die erste Fassung machte am 2026-09-06
// **sieben** Läufe rot — vier Flächenläufe über den Profilwechsel und drei im
// Dienst —, und keiner davon an einem Fehler des Bestands.
//
// # Was dann noch bewacht ist
//
// **Ein Profil, dessen Eigner einen zweiten Faktor hat.** Dort ist die
// Freigabe der Weg hinein ohne dessen Code — und genau das soll sie sein: der
// Eigner lässt jemanden herein, ohne ihm seinen Faktor zu geben.
//
// **Und was nicht, gehört gesagt** (§10): in einem Bestand, in dem niemand
// einen zweiten Faktor hat, verhindert die Freigabe **nichts**. Das ist
// dieselbe Grenze, die der Faktor selbst hat, und Vermerk
// [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md)
// sagt schon dort, dass die Zusage über die **Datei** allein die
// Verschlüsselung leistet.
//
// # Warum der Faktorfehler die Antwort ist und nicht der Erlaubnisfehler
//
// **Weil der häufige Fall der heimkehrende Eigner ist.** „Frag nach dem Code"
// sagt ihm, was zu tun ist; „dir fehlt die Rolle" schickte ihn zu jemandem,
// der er selbst ist.
// # Und warum das exportiert ist
//
// **Weil die Deckübernahme dieselbe Frage stellt** — Werkstück 12 (#1351):
// wer aus einem Profil nehmen darf, ist genau, wer hineingehen darf. Eine
// zweite Prüfung daneben wäre eine zweite Antwort auf dieselbe Frage, und
// beide liefen irgendwann auseinander (§10).
//
// **Die Übernahme reicht einen leeren Nachweis** — sie fragt niemanden nach
// einem Code. Für ein Profil ohne Faktor heißt das „geht", für eines mit
// Faktor „nur mit Freigabe", und beides ist dasselbe wie beim Wechsel. Wer
// hier strenger prüfte als dort, hätte eine Sperre gebaut, die man mit einem
// Wechsel umgeht: hineingehen, ansehen, abtippen.
func (v *Nutzerverwaltung) Zutritt(
	ctx context.Context, ziel domain.Profil, nachweis string,
) error {
	faktorfehler := v.faktorPruefen(ctx, ziel.NutzerID, nachweis)
	if faktorfehler == nil {
		return nil
	}

	reicht, fehler := v.freigabeReicht(ctx, ziel)
	if fehler != nil {
		return fehler
	}

	if reicht {
		return nil
	}

	return faktorfehler
}

// WerArbeitet sagt, wer gerade arbeitet — Werkstück 17 von Phase 4.12.
//
// # Warum das eine eigene Frage ist
//
// **Weil sie sich vom gezeigten Profil gelöst hat.** Vorher war die Antwort
// „der Eigner des gezeigten Profils", und damit wurde jeder, der in ein
// fremdes Profil wechselte, zu dessen Eigner — samt dessen Freigaben. Der
// Befund und was er gekostet hätte, steht in
// #1400.
func (v *Nutzerverwaltung) WerArbeitet(ctx context.Context) (int64, error) {
	arbeitet, fehler := v.zeiger.Nutzerzeiger(ctx)
	if fehler != nil {
		return 0, fehler
	}

	// **Eine Null ist keine Identität, sondern Unwissen.**
	//
	// Der Zeiger im Bestand fällt der Reihe nach zurück — gesetzter Nutzer,
	// Eigner des gezeigten Profils, erstes Profil überhaupt — und gibt **0**,
	// wenn keiner davon greift. Das geschieht bei einem Bestand ohne Profile
	// und bei einem Zeiger, der ins Leere steht.
	//
	// **Ein `0` weiterzureichen wäre eine Erlaubnis aus Unwissen** (§8): die
	// Freigabenprüfung fragte dann nach der Rolle des Nutzers 0, bekäme keine,
	// und die Absage klänge nach „dir fehlt die Rolle" statt nach „hier stimmt
	// etwas mit dem Bestand nicht".
	//
	// **Angestrichen von `TestOhneGezeigtesProfilWirdAbgewiesen`**, der genau
	// diese Zusage seit Werkstück 11 hält — sie lag bis Werkstück 17 in
	// `Gezeigt`, und mit dem eigenen Zeiger wäre sie fast still verlorengegangen.
	if arbeitet == 0 {
		return 0, fmt.Errorf(
			"%w: der Bestand sagt nicht, wer arbeitet", domain.ErrNotFound)
	}

	return arbeitet, nil
}

// RolleImGezeigten sagt, welche Rolle der arbeitende Mensch im gezeigten
// Profil hat — Werkstück 16 von Phase 4.12.
//
// # Wofür es sie gibt
//
// **Daran hängt, ob der Bestand nur lesend geöffnet wird.** Ein `leser` darf
// seit diesem Werkstück hinein und soll nichts ändern; der Schutz sitzt im
// Treiber, und wer ihn nehmen muss, fragt hier.
//
// **Ablesbar ist sie erst seit Werkstück 17** (#1400):
// vorher galt, wer hineingewechselt war, als der Eigner — und ein Eigner ist
// [domain.RolleVerwalter]. Die Frage hätte sich selbst beantwortet.
func (v *Nutzerverwaltung) RolleImGezeigten(ctx context.Context) (
	domain.Rolle, error,
) {
	if v.freigaben == nil {
		return domain.RolleVerwalter, nil
	}

	werArbeitet, fehler := v.WerArbeitet(ctx)
	if fehler != nil {
		return "", fehler
	}

	gezeigt, lesefehler := v.Gezeigt(ctx)
	if lesefehler != nil {
		return "", lesefehler
	}

	return v.freigaben.Rolle(ctx, werArbeitet, gezeigt)
}

// freigabeReicht sagt, ob der arbeitende Mensch in dieses Profil darf.
//
// # Wer der arbeitende Mensch ist
//
// **Der Eigner des gezeigten Profils.** Diese Anwendung hat keinen Begriff
// einer Sitzung neben dem Profilzeiger, und einen zu erfinden hieße, zwei
// Wahrheiten darüber zu führen, wer gerade arbeitet
// ([0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)).
//
// # Warum ein `nil` durchlässt
//
// **Dieselbe Begründung wie bei [Nutzerverwaltung.faktorPruefen]**: eine
// Verwaltung ohne Naht bricht keine Zusage, sondern macht keine. Der
// Composition Root setzt sie in **beiden** Wurzeln; der Flächenlauf fährt
// gegen die gebauten Programme (§10).
//
// # Warum eine fehlende Rolle **kein** Fehler ist
//
// **Weil sie eine Antwort ist.** [domain.ErrErlaubnis] heißt hier „nein" und
// nicht „es ging schief"; ein Lesefehler des Bestands heißt das andere, und
// der darf nicht als „nein" durchgehen — sonst käme jemand nicht hinein, weil
// die Platte klemmt, und läse „dir fehlt die Rolle".
func (v *Nutzerverwaltung) freigabeReicht(
	ctx context.Context, ziel domain.Profil,
) (bool, error) {
	if v.freigaben == nil {
		return false, nil
	}

	// **Wer arbeitet, und nicht wem das gezeigte Profil gehört** — Werkstück
	// 17 von Phase 4.12. Bis dahin stand hier `Gezeigt(ctx).NutzerID`, und
	// damit prüfte jeder Wechsel gegen den Eigner des Profils, in dem der
	// Zeiger gerade steht. Wer einmal in ein fremdes gewechselt war, wurde als
	// dessen Eigner geprüft und erreichte, was **der** erreicht
	// (#1400).
	werArbeitet, fehler := v.WerArbeitet(ctx)
	if fehler != nil {
		return false, fehler
	}

	// **Unter den eigenen Profilen entscheidet der Faktor allein** — und der
	// Fall steht hier, weil die Verwaltung sonst den Bestand fragte, um eine
	// Antwort zu bekommen, die schon dasteht.
	if werArbeitet == ziel.NutzerID {
		return false, nil
	}

	// **`RolleLeser` reicht zum Hineingehen** — Werkstück 16 von Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// Hier stand `RolleMitarbeiter`, und [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)
	// nannte den Grund: ein Nur-Lese-Modus **innerhalb** eines Profils braucht
	// einen lückenlosen Schreibschutz, und den gibt es nur im Treiber. **Den
	// gibt es jetzt** (`storage.OeffnenNurLesend`), und damit fällt die
	// Beschränkung.
	//
	// **Was `leser` darf, entscheidet danach die Verbindung und nicht dieser
	// Ruf**: wer als `leser` hineingeht, bekommt einen Bestand, der nichts
	// annimmt.
	// **Über die Eintrittsrolle und nicht über die Rolle**: die Sichtbarkeit
	// `bestand` gibt `leser` für die **Übersicht** und nicht für den Inhalt.
	// Der Grund steht bei `Freigabenverwaltung.Eintrittsrolle`.
	eintritt, rollenfehler := v.freigaben.Eintrittsrolle(ctx, werArbeitet, ziel)
	if rollenfehler != nil {
		return false, rollenfehler
	}

	// **Ein Wahrheitswert und keine Fallunterscheidung über einen Fehler.**
	//
	// `domain.Freigabefehler(…)` umschließt immer [domain.ErrErlaubnis], und
	// ein Lesefehler aus dem Bestand kommt schon eine Zeile höher an, wo die
	// Rolle gelesen wird — ein dritter Fall „ein anderer Fehler als die
	// Erlaubnis" wäre unbetretbar.
	//
	// **Ein Zweig, den kein Aufruf betritt, ist eine Behauptung darüber, was
	// im Ernstfall geschieht** (§3) — und niemand kann sie prüfen.
	return eintritt.Reicht(domain.RolleLeser), nil
}
