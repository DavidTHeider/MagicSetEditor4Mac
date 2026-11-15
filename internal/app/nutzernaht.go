// **Nutzer und Profile für die drei Flächen ohne Fenster** — Werkstück 1 von
// Phase 4.12
// (#1351).
//
// **Dieselbe Form wie `decknahtflaechen.go`** und aus demselben Grund (§16):
// alle vier Flächen benutzen dieselbe Anwendung. Ein zweiter Aufbau wäre eine
// zweite Fachlogik.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Nutzerflaeche ist die Anwendung selbst, für eine Fläche ohne Fenster.
type Nutzerflaeche struct{ app *App }

// NutzerflaecheUeber baut sie über einem Bestand.
func NutzerflaecheUeber(db *sql.DB) *Nutzerflaeche {
	return nutzerflaeche(db, nil, nil)
}

// NutzerflaecheOhneSensor baut sie mit einer eingesetzten Biometrie.
//
// # Wofür es das gibt
//
// **Damit ein Lauf nicht den Sensor dieses Rechners anfasst.**
// `BiometrieEinrichten` ruft `Bestaetige`, und das **zeigt einen Dialog und
// wartet auf einen Finger** — `sensornaht.go` sagt das in seinem Kopf. Am
// 2026-09-06 ist mitten in einem Pipelinelauf ein Touch-ID-Fenster aufgegangen
// und hat auf den Auftraggeber gewartet, weil ein Flächenlauf diesen Weg fuhr.
//
// # Warum ein eigener Bauweg und kein Ruf danach
//
// **Weil der zweite Faktor den Sensor beim Bauen einfängt.**
// `anmeldungVerdrahten` reicht ihn an `service.NeuerZweiterfaktor`; wer ihn
// danach setzte, änderte ein Feld, das niemand mehr liest — der erste Versuch
// tat genau das und war grün, ohne etwas zu bewirken (§10: eine Gegenprobe,
// die nichts ändert, ist keine).
//
// **Im Betrieb ruft ihn niemand.** `NutzerflaecheUeber` bleibt der Weg, und
// ohne Einsatz nimmt sie den Sensor des Rechners wie bisher.
func NutzerflaecheOhneSensor(
	db *sql.DB, geber service.Sensor, auskunft func(context.Context) biometrie.Auskunft,
) *Nutzerflaeche {
	return nutzerflaeche(db, geber, auskunft)
}

// nutzerflaeche ist der gemeinsame Bauweg.
func nutzerflaeche(
	db *sql.DB, geber service.Sensor, auskunft func(context.Context) biometrie.Auskunft,
) *Nutzerflaeche {
	anwendung := NeueApp("")

	// **Vor allem anderen**, denn `anmeldungVerdrahten` reicht ihn weiter.
	if geber != nil {
		anwendung.MitSensor(geber, auskunft)
	}
	verwaltung := service.NeueNutzerverwaltung(storage.NeueNutzerablage(db))
	anwendung.MitNutzern(verwaltung)
	anwendung.MitProfilwechsel(verwaltung)
	anwendung.MitProfilkenner(storage.NeueNutzerablage(db))

	// **Erst die Anmeldung, dann der Faktor an der Verwaltung.** Sie baut den
	// Prüfer, und der Wechsel braucht ihn — die Reihenfolge ist keine
	// Bequemlichkeit, sondern die Abhängigkeit (Werkstück 7, #1351).
	anmeldungVerdrahten(anwendung, db)
	verwaltung.MitFaktor(anwendung.faktor)

	// **Und dasselbe Beiwerk wie am Fenster** — Befund 2 des Zweitblicks über
	// die Phasen 4.0 bis 4.13 (2026-09-07).
	//
	// **Gefunden hat es der Nahtwächter**, nicht ich: `MitBeiwerk` stand nur in
	// `leutestart.go`, und `mse nutzer loeschen` ließ Bild und Geheimnis
	// weiter liegen — derselbe Befund an einer anderen Fläche. Genau dafür
	// gibt es die Zusage „jede Wurzel verdrahtet dieselben Nähte".
	verwaltung.MitBeiwerk(beiwerkNaehte(anwendung, db))
	anwendung.MitNutzereinstellungen(
		service.NeueNutzereinstellungen(storage.NeueNutzerablage(db)))
	anwendung.MitPersonenauskunft(
		service.NeuePersonenauskunft(storage.NeueNutzerablage(db)))
	anwendung.MitToken(
		service.NeueTokenverwaltung(storage.NeueZugriffstokenablage(db)))

	// **Und wer in welches fremde Profil darf** — Werkstück 11 desselben
	// Issues, dieselbe Überlegung wie in `leutestart.go`.
	freigaben := service.NeueFreigabenverwaltung(
		storage.NeueFreigabenablage(db), storage.NeueNutzerablage(db))

	anwendung.MitFreigaben(freigaben)
	verwaltung.MitFreigaben(freigaben)
	anwendung.MitAutorisierung(autorisierungUeber(db, freigaben))

	// **Und die Deckübernahme aus einem fremden Profil** — Werkstück 12
	// desselben Issues. **Sie bekommt dieselbe Freigabenverwaltung**: sie
	// verlangt dieselbe Rolle wie der Wechsel, und eine zweite wäre ein
	// zweiter Stand.
	anwendung.MitDeckuebernahme(service.NeueDeckuebernahme(
		storage.NeueDeckentnahme(db), storage.NeueDeckablage(db),
		verwaltung, freigaben))

	// **Und die Vorgaben** — Werkstück 15 desselben Issues; dieselbe
	// Überlegung wie in `start.go`.
	if verwaltung, fehler := vorgabenverwaltungUeber(db); fehler == nil {
		anwendung.MitVorgaben(verwaltung)
	}

	return &Nutzerflaeche{app: anwendung}
}

// Nutzer gibt alle Nutzer samt ihren Profilen.
func (f *Nutzerflaeche) Nutzer(ctx context.Context) NutzerAntwort {
	return f.app.Nutzer(ctx)
}

// NutzerAnlegen legt einen Nutzer samt seinem ersten Profil an.
func (f *Nutzerflaeche) NutzerAnlegen(ctx context.Context, name, bild string) EinNutzerAntwort {
	return f.app.NutzerAnlegen(ctx, name, bild)
}

// NutzerbildSetzen legt eine Bilddatei als Profilbild ab.
//
// **Der Pfad und kein Dialog** — Werkstück 8 (#1351). Ohne Fenster gibt es
// keinen; was hereinkommt, ist der Ort einer Datei auf dem Rechner, auf dem
// diese Fläche läuft.
func (f *Nutzerflaeche) NutzerbildSetzen(
	ctx context.Context, nutzerID, fassung, pfad string,
) EinNutzerAntwort {
	return f.app.NutzerbildSetzen(ctx, nutzerID, fassung, pfad)
}

// NutzerbildEntfernen nimmt das Profilbild eines Nutzers wieder weg.
func (f *Nutzerflaeche) NutzerbildEntfernen(
	ctx context.Context, nutzerID, fassung string,
) EinNutzerAntwort {
	return f.app.NutzerbildEntfernen(ctx, nutzerID, fassung)
}

// ProfilbadgeSetzen legt eine Bilddatei als Abzeichen eines Profils ab.
//
// **Der Pfad und kein Dialog** — dieselbe Form wie bei
// [Nutzerflaeche.NutzerbildSetzen]: ohne Fenster gibt es keinen.
func (f *Nutzerflaeche) ProfilbadgeSetzen(
	ctx context.Context, profilID, fassung, pfad string,
) EinProfilAntwort {
	return f.app.ProfilbadgeSetzen(ctx, profilID, fassung, pfad)
}

// ProfilbadgeEntfernen nimmt das Abzeichen eines Profils wieder weg.
func (f *Nutzerflaeche) ProfilbadgeEntfernen(
	ctx context.Context, profilID, fassung string,
) EinProfilAntwort {
	return f.app.ProfilbadgeEntfernen(ctx, profilID, fassung)
}

// Nutzereinstellungen gibt die Einstellungen eines Menschen.
func (f *Nutzerflaeche) Nutzereinstellungen(
	ctx context.Context, nutzerID string,
) NutzereinstellungenAntwort {
	return f.app.Nutzereinstellungen(ctx, nutzerID)
}

// NutzereinstellungSetzen schreibt eine fort und gibt alle zurück.
func (f *Nutzerflaeche) NutzereinstellungSetzen(
	ctx context.Context, nutzerID, schluessel, wert string,
) NutzereinstellungenAntwort {
	return f.app.NutzereinstellungSetzen(ctx, nutzerID, schluessel, wert)
}

// Personendaten gibt heraus, was über einen Menschen gespeichert ist.
//
// **Ohne Dialog und ohne Datei.** Wer den Weg ohne Fenster nimmt, bekommt die
// Auskunft selbst und entscheidet, was er damit tut — an der Kommandozeile
// reicht ein `>`.
func (f *Nutzerflaeche) Personendaten(
	ctx context.Context, nutzerID string,
) PersonendatenAntwort {
	return f.app.Personendaten(ctx, nutzerID)
}

// ProfilAnlegen legt ein weiteres Profil an.
func (f *Nutzerflaeche) ProfilAnlegen(ctx context.Context, nutzerID, name string) EinProfilAntwort {
	return f.app.ProfilAnlegen(ctx, nutzerID, name)
}

// ProfilWechseln stellt das Profil um, in dem gearbeitet wird.
//
// # Ein Zeiger für alle vier Flächen
//
// **Er liegt im Bestand**, nicht im Prozess: das Issue verlangt, dass „der
// gewählte Nutzer und das gewählte Profil den Aufruf überleben" — sonst wäre
// jeder Befehl an der Kommandozeile eine eigene Anmeldung.
//
// **Was das heißt, gehört gesagt**: wer an der Kommandozeile wechselt,
// wechselt auch für das Fenster. Das ist bei **einem** Rechner mit **einer**
// arbeitenden Person richtig; wer für einen einzelnen Aufruf etwas anderes
// will, nimmt `--profil` (Werkstück 3).
func (f *Nutzerflaeche) ProfilWechseln(
	ctx context.Context, profilID, nachweis string,
) EinProfilAntwort {
	return f.app.ProfilWechseln(ctx, profilID, nachweis)
}

// GezeigtesProfil sagt, in welchem Profil gearbeitet wird.
func (f *Nutzerflaeche) GezeigtesProfil(ctx context.Context) EinProfilAntwort {
	return f.app.GezeigtesProfil(ctx)
}

// Anbieter nennt die Anbieter, über die man sich anmelden kann.
func (f *Nutzerflaeche) Anbieter(ctx context.Context) AnbieterAntwort {
	return f.app.Anbieter(ctx)
}

// Anmelden führt den Anmeldefluss über einen Anbieter.
//
// **Auch ohne Fenster** — und das ist der Punkt: das Issue verlangt
// `mse anmelden`, und der Anbieterfluss „öffnet einen Browser und nimmt die
// Antwort auf einem Rückrufziel an". Der Rückruf geht auf die
// Schleifenadresse, und die braucht kein Fenster.
func (f *Nutzerflaeche) Anmelden(ctx context.Context, anbieter string) EinNutzerAntwort {
	return f.app.Anmelden(ctx, anbieter)
}

// Anbieterkonten gibt die Konten eines Nutzers.
func (f *Nutzerflaeche) Anbieterkonten(ctx context.Context, nutzerID string) KontenAntwort {
	return f.app.Anbieterkonten(ctx, nutzerID)
}

// KontoLoesen nimmt eine fremde Kennung von einem Nutzer.
func (f *Nutzerflaeche) KontoLoesen(ctx context.Context, nutzerID, anbieter string) SchreibAntwort {
	return f.app.KontoLoesen(ctx, nutzerID, anbieter)
}

// NutzerAendern schreibt die Angaben eines Nutzers fort.
func (f *Nutzerflaeche) NutzerAendern(
	ctx context.Context, angaben NutzerangabenDTO,
) EinNutzerAntwort {
	return f.app.NutzerAendern(ctx, angaben)
}

// ProfilAendern schreibt den Namen eines Profils fort.
func (f *Nutzerflaeche) ProfilAendern(
	ctx context.Context, angaben ProfilangabenDTO,
) EinProfilAntwort {
	return f.app.ProfilAendern(ctx, angaben)
}

// NutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus.
func (f *Nutzerflaeche) NutzerLoeschen(
	ctx context.Context, nutzerID, name string,
) SchreibAntwort {
	return f.app.NutzerLoeschen(ctx, nutzerID, name)
}

// ProfilLoeschen nimmt ein Profil heraus.
func (f *Nutzerflaeche) ProfilLoeschen(
	ctx context.Context, profilID, name string,
) SchreibAntwort {
	return f.app.ProfilLoeschen(ctx, profilID, name)
}

// KontoEintragen legt ein Konto unter einer Adresse oder einer Rufnummer an.
func (f *Nutzerflaeche) KontoEintragen(
	ctx context.Context, kennung, kennwort, name string,
) EinNutzerAntwort {
	return f.app.KontoEintragen(ctx, kennung, kennwort, name)
}

// KontoAnmelden lässt einen Nutzer unter seiner Kennung herein.
func (f *Nutzerflaeche) KontoAnmelden(
	ctx context.Context, kennung, kennwort, nachweis string,
) EinNutzerAntwort {
	return f.app.KontoAnmelden(ctx, kennung, kennwort, nachweis)
}

// KontoAnonym legt ein Konto ohne jede Angabe an.
func (f *Nutzerflaeche) KontoAnonym(ctx context.Context) EinNutzerAntwort {
	return f.app.KontoAnonym(ctx)
}

// Faktoren gibt die zweiten Faktoren eines Nutzers samt der Rechnerauskunft.
func (f *Nutzerflaeche) Faktoren(ctx context.Context, nutzerID string) FaktorenAntwort {
	return f.app.Faktoren(ctx, nutzerID)
}

// FaktorEinrichten zieht ein TOTP-Geheimnis.
func (f *Nutzerflaeche) FaktorEinrichten(
	ctx context.Context, nutzerID string,
) Einrichtungsantwort {
	return f.app.FaktorEinrichten(ctx, nutzerID)
}

// FaktorBestaetigen trägt den TOTP-Faktor ein.
func (f *Nutzerflaeche) FaktorBestaetigen(
	ctx context.Context, nutzerID, code string,
) SchreibAntwort {
	return f.app.FaktorBestaetigen(ctx, nutzerID, code)
}

// BiometrieEinrichten trägt Touch ID oder Face ID als zweiten Faktor ein.
func (f *Nutzerflaeche) BiometrieEinrichten(
	ctx context.Context, nutzerID string,
) SchreibAntwort {
	return f.app.BiometrieEinrichten(ctx, nutzerID)
}

// FaktorEntfernen nimmt einen zweiten Faktor heraus.
func (f *Nutzerflaeche) FaktorEntfernen(
	ctx context.Context, nutzerID, art string,
) SchreibAntwort {
	return f.app.FaktorEntfernen(ctx, nutzerID, art)
}

// **Die Zugriffstoken technischer Profile** — Werkstück 10 (#1351).
//
// **Verwaltet an jeder Fläche, geprüft nur am REST-Zugang**: das Anlegen und
// Zurückziehen ist eine Verwaltungsaufgabe wie jede andere und gehört überall
// hin (§16). Wer ein Token **vorzeigt**, tut das nur dort, wo jemand auf einem
// Port lauscht.

// Token gibt die Zugriffstoken eines Profils.
func (f *Nutzerflaeche) Token(ctx context.Context, profilID string) TokenAntwort {
	return f.app.Token(ctx, profilID)
}

// TokenAnlegen legt eines an und gibt seinen Klartext **einmal**.
func (f *Nutzerflaeche) TokenAnlegen(
	ctx context.Context, angaben TokenangabenDTO,
) EinTokenAntwort {
	return f.app.TokenAnlegen(ctx, angaben)
}

// TokenAendern schreibt Bezeichnung, Erlaubnis und Ablauf fort.
func (f *Nutzerflaeche) TokenAendern(
	ctx context.Context, angaben TokenangabenDTO,
) EinTokenAntwort {
	return f.app.TokenAendern(ctx, angaben)
}

// TokenLoeschen nimmt eines heraus.
func (f *Nutzerflaeche) TokenLoeschen(ctx context.Context, tokenID string) SchreibAntwort {
	return f.app.TokenLoeschen(ctx, tokenID)
}

// **Die Vorgaben eines Profils, die auf Zeilen zeigen** — Werkstück 15
// (#1351).

// Profilvorgaben gibt sie samt ihrem Angebot.
func (f *Nutzerflaeche) Profilvorgaben(ctx context.Context, profilID string) VorgabenAntwort {
	return f.app.Profilvorgaben(ctx, profilID)
}

// ProfilvorgabeSetzen ersetzt die Werte **einer** Art.
func (f *Nutzerflaeche) ProfilvorgabeSetzen(
	ctx context.Context, angaben VorgabeangabenDTO,
) VorgabenAntwort {
	return f.app.ProfilvorgabeSetzen(ctx, angaben)
}

// **Wer in welches fremde Profil darf** — Werkstück 11 (#1351).

// Freigaben gibt die Freigaben eines Profils samt der eigenen Rolle.
func (f *Nutzerflaeche) Freigaben(ctx context.Context, profilID string) FreigabenAntwort {
	return f.app.Freigaben(ctx, profilID)
}

// FreigabeSetzen legt eine Freigabe an oder schreibt ihre Rolle fort.
func (f *Nutzerflaeche) FreigabeSetzen(
	ctx context.Context, angaben FreigabeangabenDTO,
) FreigabenAntwort {
	return f.app.FreigabeSetzen(ctx, angaben)
}

// FreigabeLoeschen zieht eine Freigabe zurück.
func (f *Nutzerflaeche) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID string,
) FreigabenAntwort {
	return f.app.FreigabeLoeschen(ctx, profilID, nutzerID)
}

// **Ein Deck aus einem fremden Profil** — Werkstück 12 (#1351).

// DeckUebernehmen holt es in das eigene Profil.
func (f *Nutzerflaeche) DeckUebernehmen(
	ctx context.Context, quellProfil, quellDeck string,
) Deckuebernahmeantwort {
	return f.app.DeckUebernehmen(ctx, quellProfil, quellDeck)
}
