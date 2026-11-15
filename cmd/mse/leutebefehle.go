// **Der Verteiler für die Befehle über Menschen** — Nutzer, Profile, Konten,
// zweite Faktoren und Zugriffstoken.
//
// **Herausgelöst aus `lauf.go` am 2026-09-05**, als `token-loeschen` die Datei
// über die Zeilengrenze aus §2 hob (Werkstück 10 von Phase 4.12,
// #1351).
//
// **Der Schnitt folgt der Frage** und nicht der Zeilenzahl: `werkstattbefehle`
// und `fremdbefehle` stehen aus demselben Grund eigens da. Was hier steht,
// beantwortet **wer arbeitet und wie er hereinkommt** — dieselbe Frage wie
// `internal/app/anwendungleute.go`.

package main

import "context"

// leuteschluessel markiert einen Kontext als Leute-Befehl.
type leuteschluessel struct{}

// alsLeutebefehl markiert den Kontext.
//
// # Warum das über den Kontext geht und nicht über ein Argument
//
// **Weil `mitBestand` an dreißig Stellen gerufen wird** und die Antwort auf
// „darf dieser Befehl schreiben?" an **einer** Stelle steht: hier. Ein
// Argument durch dreißig Funktionen zu fädeln hieße, dieselbe Entscheidung
// dreißigmal zu wiederholen — und beim einunddreißigsten Befehl zu vergessen.
//
// **Ein Schreiber und ein Leser**, beide in dieser Datei und in
// `bestandsgriff.go`. Das ist die ganze Reichweite.
func alsLeutebefehl(ctx context.Context) context.Context {
	return context.WithValue(ctx, leuteschluessel{}, true)
}

// istLeutebefehl sagt, ob der Kontext markiert ist.
func istLeutebefehl(ctx context.Context) bool {
	markiert, steht := ctx.Value(leuteschluessel{}).(bool)

	return steht && markiert
}

// leutebefehle verteilt, was über Menschen geht.
//
// # Warum sie an der schreibenden Verbindung arbeiten
//
// **Weil man aus einem nur lesenden Profil wieder hinausfinden muss** —
// Werkstück 16 von Phase 4.12
// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)),
// und das Hinausfinden schreibt den Profilzeiger. **Dieselbe benannte
// Ausnahme wie am Fenster**, wo `verdrahten` die Leute an `App.db` hängt und
// alles andere an die Verbindung, die es bekommt.
//
// **Und sie ist eng**: Nutzer, Profile, Konten, Faktoren, Token, Freigaben —
// alles, was **nicht** zum Inhalt eines Profils gehört. Sets, Karten, Decks,
// Sammlung, Turniere und Einstellungen liegen darin, und für die gilt der
// Schutz.
func leutebefehle(
	ctx context.Context, befehl string, rest []string, bestandsort string,
	an *schreiber,
) (code int, getroffen bool) {
	ctx = alsLeutebefehl(ctx)

	switch befehl {
	case "nutzer":
		return nutzerbefehl(ctx, bestandsort, an, rest), true

	case "nutzer-aendern":
		return nutzeraendernbefehl(ctx, bestandsort, an, rest), true

	case "nutzereinstellung":
		return nutzereinstellungsbefehl(ctx, bestandsort, an, rest), true

	case "personendaten":
		return personendatenbefehl(ctx, bestandsort, an, rest), true

	case "profil-badge":
		return profilbadgebefehl(ctx, bestandsort, an, rest), true

	case "nutzer-bild":
		return nutzerbildbefehl(ctx, bestandsort, an, rest), true

	case "nutzer-loeschen":
		return nutzerloeschenbefehl(ctx, bestandsort, an, rest), true

	case "profil-aendern":
		return profilaendernbefehl(ctx, bestandsort, an, rest), true

	case "profil-loeschen":
		return profilloeschenbefehl(ctx, bestandsort, an, rest), true

	case "token":
		return tokenbefehl(ctx, bestandsort, an, rest), true

	case "token-anlegen":
		return tokenanlegenbefehl(ctx, bestandsort, an, rest), true

	case "token-aendern":
		return tokenaendernbefehl(ctx, bestandsort, an, rest), true

	case "token-rotieren":
		return tokenrotierenbefehl(ctx, bestandsort, an, rest), true

	case "token-alle-widerrufen":
		return tokenallewiderrufenbefehl(ctx, bestandsort, an, rest), true

	case "token-ablaufwahlen":
		return tokenablaufwahlenbefehl(ctx, bestandsort, an, rest), true

	case "token-loeschen":
		return tokenloeschenbefehl(ctx, bestandsort, an, rest), true

	case "vorgaben":
		return vorgabenbefehl(ctx, bestandsort, an, rest), true

	case "vorgabe-setzen":
		return vorgabesetzenbefehl(ctx, bestandsort, an, rest), true

	case "freigaben":
		return freigabenbefehl(ctx, bestandsort, an, rest), true

	case "freigabe-setzen":
		return freigabesetzenbefehl(ctx, bestandsort, an, rest), true

	case "freigabe-loeschen":
		return freigabeloeschenbefehl(ctx, bestandsort, an, rest), true

	case "anbieter":
		return anbieterbefehl(ctx, bestandsort, an, rest), true

	case "anmelden":
		return anmeldebefehl(ctx, bestandsort, an, rest), true

	case "konten":
		return kontenbefehl(ctx, bestandsort, an, rest), true

	case "konto-neu":
		return kontoneubefehl(ctx, bestandsort, an, rest), true

	case "konto-anmelden":
		return kontoanmeldebefehl(ctx, bestandsort, an, rest), true

	case "konto-anonym":
		return kontoanonymbefehl(ctx, bestandsort, an, rest), true

	case "sitzung":
		return sitzungsbefehl(ctx, bestandsort, an, rest), true

	case "kennwort-wechseln":
		return kennwortwechselbefehl(ctx, bestandsort, an, rest), true

	case "kennwort-wiederherstellen":
		return kennwortwiederherstellbefehl(ctx, bestandsort, an, rest), true

	case "faktoren":
		return faktorenbefehl(ctx, bestandsort, an, rest), true

	case "faktor-neu":
		return faktorneubefehl(ctx, bestandsort, an, rest), true

	case "faktor-bestaetigen":
		return faktorbestaetigenbefehl(ctx, bestandsort, an, rest), true

	case "faktor-loeschen":
		return faktorloeschenbefehl(ctx, bestandsort, an, rest), true

	case "profil":
		return profilbefehl(ctx, bestandsort, an, rest), true

	case "nutzer-neu":
		return nutzerneubefehl(ctx, bestandsort, an, rest), true

	case "deck-uebernehmen":
		return deckuebernahmebefehl(ctx, bestandsort, an, rest), true

	case "decks-uebernehmbar":
		return deckwahlbefehl(ctx, bestandsort, an, rest), true

	case "profil-neu":
		return profilneubefehl(ctx, bestandsort, an, rest), true
	}

	return CodeGut, false
}
