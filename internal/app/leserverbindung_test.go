//go:build darwin

// **Wann die zweite Verbindung entsteht und wann nicht** — Werkstück 16 von
// Phase 4.12
// (#1351).
//
// **Eigene Datei neben `nurlesen_test.go`.** Der Schnitt folgt der Frage: dort
// steht, **was ein Leser kann**, hier, **wann er einer wird** — und das sind
// die Wege, auf denen es schiefgeht.

package app_test

import (
	"errors"
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine unlesbare Rolle hält die Anwendung nicht auf.**
//
// **Über einen Wechsel und nicht über den Start**: `Starten` verdrahtet zuerst
// und setzt dabei die echte Nutzerverwaltung ein — ein Doppel davor wäre
// überschrieben, bevor gefragt wird. Der erste Anlauf tat genau das und war
// grün, ohne den Zweig zu betreten (§10).
func TestEineUnlesbareRolleHaeltDenWechselNichtAuf(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	anwendung.MitProfilwechsel(&wechseldoppel{
		gezeigt:      domain.Profil{ID: 1, NutzerID: 1, Name: "Meins"},
		rollenfehler: errors.New("die Rolle klemmt"),
	})

	if gewechselt := anwendung.ProfilWechseln(
		t.Context(), "1", ""); gewechselt.Fehler != nil {
		t.Fatalf("der Wechsel scheiterte: %+v", gewechselt.Fehler)
	}

	// **Und danach wird geschrieben** — das ist der Preis, und er gehört
	// gesagt: wer die Rolle nicht lesen kann, bekommt die schreibende
	// Verbindung. Eine Anwendung, die deswegen klemmte, tauschte einen kleinen
	// Mangel gegen einen großen (§8).
	if gesetzt := anwendung.EinstellungSetzen(
		t.Context(), "sprache", "en"); gesetzt.Fehler != nil {
		t.Errorf("nach dem Lesefehler wurde nicht geschrieben: %+v", gesetzt.Fehler)
	}
}

// **Eine zweite Verbindung wird nicht zweimal geöffnet.**
//
// **Sonst hinge die erste offen** und hielte den Bestand — ein zweiter Wechsel
// innerhalb desselben Leserzustands darf sie nicht ersetzen.
func TestDieZweiteVerbindungWirdNichtZweimalGeoeffnet(t *testing.T) {
	anwendung := alsLeserHinein(t)

	// **Das Doppel sagt weiterhin `leser`** — der Wechsel selbst geht
	// hier nicht in ein anderes Profil, sondern prüft, was beim zweiten
	// Richten geschieht.
	anwendung.MitProfilwechsel(&wechseldoppel{
		gezeigt: domain.Profil{ID: 1, NutzerID: 1, Name: "Meins"},
		rolle:   domain.RolleLeser,
	})

	if gewechselt := anwendung.ProfilWechseln(
		t.Context(), "1", ""); gewechselt.Fehler != nil {
		t.Fatalf("der zweite Wechsel scheiterte: %+v", gewechselt.Fehler)
	}

	// **Und die alte Verbindung steht noch** — sie nimmt nichts an.
	gesetzt := anwendung.EinstellungSetzen(t.Context(), "sprache", "en")

	if gesetzt.Fehler == nil || gesetzt.Fehler.Code != app.CodeNurLesend {
		t.Errorf("nach dem zweiten Richten wurde geschrieben: %+v", gesetzt.Fehler)
	}
}

// **Und eine Verbindung, die sich nicht öffnen lässt, hält nichts auf.**
//
// # Wie der Fehlschlag herbeigeführt wird
//
// **Die Bestandsdatei wird durch ein Verzeichnis ersetzt.** Die offene,
// schreibende Verbindung hält ihre Dateikennung und arbeitet weiter; ein
// **neues** Öffnen derselben Adresse scheitert. Das ist der Ernstfall in
// klein: die Datei ist fort, während der Prozess läuft.
func TestEineNichtOeffenbareLeserverbindungHaeltNichtsAuf(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := os.Remove(bestand); fehler != nil {
		t.Fatalf("Bestand entfernen: %v", fehler)
	}

	if fehler := os.Mkdir(bestand, 0o700); fehler != nil {
		t.Fatalf("Verzeichnis anlegen: %v", fehler)
	}

	anwendung.MitProfilwechsel(&wechseldoppel{
		gezeigt: domain.Profil{ID: 1, NutzerID: 1, Name: "Meins"},
		rolle:   domain.RolleLeser,
	})

	if gewechselt := anwendung.ProfilWechseln(
		t.Context(), "1", ""); gewechselt.Fehler != nil {
		t.Fatalf("der Wechsel scheiterte: %+v", gewechselt.Fehler)
	}

	// **Die schreibende Verbindung ist geblieben** — der Zweig hat sie nicht
	// gegen eine halbe getauscht.
	if gesetzt := anwendung.EinstellungSetzen(
		t.Context(), "sprache", "en"); gesetzt.Fehler != nil {
		t.Errorf("nach dem Öffnungsfehler wurde nicht geschrieben: %+v", gesetzt.Fehler)
	}
}

// **Vor dem Start gibt es keinen Arbeitskontext** — und die Bindung sagt
// das, statt einen leeren Menschen zu melden.
//
// **Der Fall ist echt und nicht gestellt**: das Fenster kann sie rufen, bevor
// alles steht. Bis Werkstück O1 von Phase 5.2 hieß die Bindung `Leseart` und
// antwortete hier „nicht nur lesend" — für einen Streifen reicht das, für
// ein Fenster, das den Menschen zeigt, nicht.
func TestOhneAutorisierungGibtEsKeinenArbeitskontext(t *testing.T) {
	t.Parallel()

	antwort := app.NeueBindungen(app.NeueApp("")).Arbeitskontext(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("eine Anwendung vor dem Start meldete einen Arbeitskontext")
	}

	if antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Code %q, erwartet war %q", antwort.Fehler.Code, app.CodeNichtGefunden)
	}
}

// **Und ein Fehler beim Fragen kommt an der Bindung an** (§8).
func TestDerArbeitskontextMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)
	anwendung := app.NeueApp("")
	anwendung.MitAutorisierung(service.NeueAutorisierung(
		stolperzeiger{fehler: errors.New("der Zeiger klemmt")},
		service.NeueFreigabenverwaltung(
			storage.NeueFreigabenablage(db), storage.NeueNutzerablage(db))))

	if antwort := anwendung.Arbeitskontext(probekontext(t)); antwort.Fehler == nil {
		t.Error("der Lesefehler ging verloren")
	}
}
