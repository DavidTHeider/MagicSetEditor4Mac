package app

// **Ein Vorschlag an einem fertigen Werkzeugsatz** — herausgelöst aus
// `kivorschlag.go` für AI-05
// (#2042).
//
// # Was AI-05 hier ändert
//
// **Der Lauf sagt, ob ein anderer Anbieter es versuchen darf.** Für eine
// ausdrückliche Wahl ist das gleichgültig; für den Vorrang
// (`kivorrangnaht.go`) ist es die Frage, ob der nächste in der Reihe drankommt.
// Versuchen darf er es, wenn **das Modell** nicht geliefert hat — es war nicht
// bereit, hat abgelehnt, ist abgebrochen, gab kein JSON oder nichts, was das
// Schema durchließ. Nicht, wenn **die Anwendung** etwas verweigert —
// Budget, Rolle, Schema der Karte: das sagte jeder andere Anbieter genauso.
//
// **Und eine Ablehnung ist ein Ausgang**, kein Fehler (AI-05.AK-06): der
// Vorschlag kommt ohne Felder zurück, nennt sie und bietet die Rückfälle an.

import (
	"context"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KIAusgangAbgelehnt heißt: das Modell hat die Antwort verweigert.
const KIAusgangAbgelehnt = "abgelehnt"

// errKIAbgelehnt markiert eine Ablehnung durch die Schutzregeln eines
// Modells — sie ist ein Ausgang und kein Fehler.
var errKIAbgelehnt = errors.New("das Modell hat die Antwort abgelehnt")

// kifrage ist, was ein Vorschlag fragt.
type kifrage struct {
	profil, karte            int64
	anbieter, alias, auftrag string

	// nur sind die Felder, nach denen gefragt wird — leer heißt: alle.
	nur []string
}

// vorschlagMit holt einen Vorschlag mit einem **fertigen** Werkzeugsatz.
// **Das Profil kommt als Wert und nicht aus dem Kontext**: der Aufrufer hat
// es gelesen, und ein zweites Lesen hier wäre ein Zweig, den nichts betritt —
// wer bis hierher kommt, hat eines (§10).
func (a *App) vorschlagMit(ctx context.Context, satz kiwerkzeugsatz, frage kifrage) KIVorschlagAntwort {
	antwort, _ := a.vorschlagVersuch(ctx, satz, frage)

	return antwort
}

// vorschlagVersuch ist [App.vorschlagMit] samt der Auskunft, ob ein anderer
// Anbieter es versuchen darf.
func (a *App) vorschlagVersuch(
	ctx context.Context, satz kiwerkzeugsatz, frage kifrage,
) (KIVorschlagAntwort, bool) {
	werk, quelle := satz.werk, satz.quelle

	angelegt, fehler := werk.Anlegen(ctx, service.KIWunsch{
		ProfilID: frage.profil, Faehigkeit: domain.KIStrukturierteGenerierung,
		Anbieter: frage.anbieter, Alias: frage.alias, Zielkarte: frage.karte,
	})
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}, false
	}

	// **Der Posten wird geschlossen, gleich wie es ausgeht** (#1542): der
	// Auftrag ist gebucht, sobald [service.KIAuftragswerk.Anlegen]
	// zurückkommt — und eine Reservierung, die offen bleibt, bindet Budget
	// für immer.
	defer a.kiPostenSchliessen(ctx, angelegt.Kostenposten)

	schema, fehler := kischemaquelle{app: a}.Schema(ctx, frage.karte)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}, false
	}

	// **Gefragt wird nach den genannten Feldern** — geprüft wird die Antwort
	// weiter gegen das ganze Schema, im Werk.
	gefragt, fehler := schema.Nur(frage.nur)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}, false
	}

	// **Wonach ausdrücklich gefragt wird, ist Pflicht** — gemessen am echten
	// Gerätemodell (Vermerk 0166): die Kartenanweisung erlaubt, Felder
	// wegzulassen, und ohne `required` kam auf eine Frage nach dem Namen
	// mehrmals `{}`. Ohne `nur` bleibt alles frei.
	form := alsJSONSchema(gefragt)
	if len(frage.nur) > 0 {
		form["required"] = frage.nur
	}

	roh, fehler := quelle.Vorschlagen(ctx, KIKartenanweisung, frage.alias, frage.auftrag, form)
	if errors.Is(fehler, errKIAbgelehnt) {
		return KIVorschlagAntwort{KIVorschlagDTO: KIVorschlagDTO{
			Anbieter: frage.anbieter, Modell: frage.alias, Ausgang: KIAusgangAbgelehnt,
			Felder: []KIFeldwertDTO{}, Befunde: []string{fehler.Error()},
		}}, true
	}

	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}, true
	}

	felder, fehler := feldwerteAus(roh)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}, true
	}

	befunde, fehler := werk.Vorschlagen(ctx, domain.KIVorschlag{
		Auftrag: angelegt.Kennung, Felder: felder,

		// **Modell und Aufgabe reisen mit** (A11, #1450): aus ihnen entsteht
		// bei der Übernahme die Herkunft, und danach weiß sie niemand mehr.
		Modell: roh.Modell, Aufgabe: frage.auftrag,
	})
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}, false
	}

	gut := ohneBeanstandete(felder, befunde)

	return KIVorschlagAntwort{KIVorschlagDTO: KIVorschlagDTO{
		Kennung: angelegt.Kennung, Anbieter: frage.anbieter, Modell: roh.Modell,
		Felder:  a.mitBisherigem(ctx, frage.karte, zuFeldwerten(gut)),
		Befunde: alsSaetze(befunde),
		Ablauf:  angelegt.Ablauf.Format(time.RFC3339),
	}}, len(gut) == 0
}
