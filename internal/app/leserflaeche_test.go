//go:build darwin

// **Der Leserwechsel an den Flächen ohne Fenster** — Werkstück 16 von Phase
// 4.12 (#1351),
// nach [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md).
//
// **Dienst und MCP-Server halten eine Verbindung über ihre ganze Laufzeit.**
// Sie mitten im Betrieb gegen eine nur lesende zu tauschen hieße, jede Naht
// neu zu bauen — deshalb weisen sie den Wechsel ab und sagen, wo er geht.

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// leserflaeche gibt eine Fläche, auf der der arbeitende Mensch im fremden
// Profil `leser` ist — und die Kennung dieses Profils.
//
// **Über die Bindungen und nicht über den Bestand**: der Weg dorthin gehört zu
// dem, was hier geprüft wird.
func leserflaeche(t *testing.T) (flaeche *app.Nutzerflaeche, fremdesProfil string) {
	t.Helper()

	flaeche = app.NutzerflaecheUeber(frischerBestand(t))

	angelegt := flaeche.NutzerAnlegen(probekontext(t), "Zweiter", "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	fremdesProfil = angelegt.Nutzer.Profile[0].ID

	// **Die Freigabe vergibt der Eigner** — dafür wird man er.
	if wechsel := flaeche.NutzerWechseln(
		probekontext(t), angelegt.Nutzer.ID, ""); wechsel.Fehler != nil {
		t.Fatalf("Nutzer wechseln: %+v", wechsel.Fehler)
	}

	if gesetzt := flaeche.FreigabeSetzen(probekontext(t), app.FreigabeangabenDTO{
		ProfilID: fremdesProfil, NutzerID: "1", Rolle: string(domain.RolleLeser),
	}); gesetzt.Fehler != nil {
		t.Fatalf("Freigabe setzen: %+v", gesetzt.Fehler)
	}

	if zurueck := flaeche.NutzerWechseln(
		probekontext(t), "1", ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	return flaeche, fremdesProfil
}

// **Der Wechsel in ein Leserprofil wird abgewiesen — und der Weg genannt.**
func TestOhneLeserwechselWirdDerLeserAbgewiesen(t *testing.T) {
	t.Parallel()

	flaeche, fremdesProfil := leserflaeche(t)

	// **Erst die Gegenprobe**: ohne die Absage geht derselbe Wechsel durch.
	// Ohne sie sagte der Lauf darunter nur, dass irgendetwas abweist (§10).
	if hinein := flaeche.ProfilWechseln(
		probekontext(t), fremdesProfil, ""); hinein.Fehler != nil {
		t.Fatalf("als Leser hinein: %+v", hinein.Fehler)
	}

	if zurueck := flaeche.ProfilWechseln(
		probekontext(t), "1", ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	abgewiesen := flaeche.OhneLeserwechsel().ProfilWechseln(
		probekontext(t), fremdesProfil, "")

	if abgewiesen.Fehler == nil {
		t.Fatal("die Fläche ließ den Leser hinein")
	}

	if abgewiesen.Fehler.Code != app.CodeErlaubnis {
		t.Errorf("Code %q, erwartet war %q", abgewiesen.Fehler.Code, app.CodeErlaubnis)
	}
}

// **Und das eigene Profil bleibt offen** — verwaltet wird weiter.
func TestOhneLeserwechselBleibtDasEigeneProfilOffen(t *testing.T) {
	t.Parallel()

	flaeche, _ := leserflaeche(t)

	if hinein := flaeche.OhneLeserwechsel().ProfilWechseln(
		probekontext(t), "1", ""); hinein.Fehler != nil {
		t.Errorf("das eigene Profil wurde verweigert: %+v", hinein.Fehler)
	}
}

// **Ein Profil, das es nicht gibt, wird als solches gemeldet** (§8).
func TestOhneLeserwechselMeldetEinUnbekanntesZiel(t *testing.T) {
	t.Parallel()

	flaeche, _ := leserflaeche(t)

	antwort := flaeche.OhneLeserwechsel().ProfilWechseln(probekontext(t), "999", "")
	if antwort.Fehler == nil {
		t.Fatal("ein Profil ohne Zeile im Bestand wurde angenommen")
	}

	if antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Code %q, erwartet war %q", antwort.Fehler.Code, app.CodeNichtGefunden)
	}
}

// **Und ein stolpernder Bestand reißt die Frage ab, statt sie zu erraten.**
//
// **Der Bestand wird geschlossen**, und damit klemmt schon die Frage, wer
// arbeitet — der erste der drei Wege in dieser Prüfung.
func TestOhneLeserwechselMeldetEinenStolperndenBestand(t *testing.T) {
	t.Parallel()

	bestand := frischerBestand(t)
	flaeche := app.NutzerflaecheUeber(bestand).OhneLeserwechsel()

	if fehler := bestand.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if antwort := flaeche.ProfilWechseln(
		probekontext(t), "1", ""); antwort.Fehler == nil {
		t.Error("ein geschlossener Bestand gab eine Antwort")
	}
}

// **Und eine unlesbare Freigabentabelle ebenso.**
//
// **Die Tabelle wird entfernt und nicht die Verbindung geschlossen**: sonst
// klemmte schon die Frage, wer arbeitet, und der Zweig darunter bliebe
// unbetreten (§10).
func TestOhneLeserwechselMeldetUnlesbareFreigaben(t *testing.T) {
	t.Parallel()

	bestand := frischerBestand(t)
	flaeche := app.NutzerflaecheUeber(bestand)

	angelegt := flaeche.NutzerAnlegen(probekontext(t), "Zweiter", "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	fremdesProfil := angelegt.Nutzer.Profile[0].ID

	if _, fehler := bestand.ExecContext(
		probekontext(t), "DROP TABLE profilfreigaben"); fehler != nil {
		t.Fatalf("Tabelle entfernen: %v", fehler)
	}

	if antwort := flaeche.OhneLeserwechsel().ProfilWechseln(
		probekontext(t), fremdesProfil, ""); antwort.Fehler == nil {
		t.Error("eine fehlende Freigabentabelle gab eine Antwort")
	}
}

// **Und die Fläche sagt den Arbeitskontext** — dieselbe Frage wie am
// Fenster (§16).
func TestDieFlaecheSagtDenArbeitskontext(t *testing.T) {
	t.Parallel()

	bestand := frischerBestand(t)

	antwort := app.NutzerflaecheUeber(bestand).Arbeitskontext(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("Arbeitskontext: %+v", antwort.Fehler)
	}

	if antwort.NurLesend || antwort.Rolle != "verwalter" || antwort.NutzerID != antwort.EignerID {
		t.Errorf("im eigenen Profil: %+v — erwartet der Verwalter, nicht nur lesend", antwort)
	}

	// **Und dieselbe Antwort über den Weg, den die Wurzeln nehmen** — er baut
	// die Fläche selbst, damit keine zweite Verdrahtung daneben entsteht.
	nurLesend, fehler := app.NurLesendGefragt(probekontext(t), bestand)
	if fehler != nil {
		t.Fatalf("nur lesend gefragt: %v", fehler)
	}

	if nurLesend {
		t.Error("die Wurzel meldete das eigene Profil als nur lesend")
	}
}
