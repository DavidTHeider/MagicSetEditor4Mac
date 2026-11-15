package service_test

// **Wer in ein fremdes Profil wechseln darf** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Das ist die Abnahme dieses Werkstücks**: nicht, dass sich eine Freigabe
// speichern lässt, sondern dass sie den Wechsel entscheidet.
//
// # Warum in jedem Fall ein zweiter Faktor steht
//
// **Weil die Freigabe der Ersatz für ihn ist und keine zweite Hürde davor.**
// Ohne Faktor kommt jeder überall hinein — das ist die Lage seit Werkstück 7
// und ändert sich hier nicht (`TestOhneZweitenFaktorSperrtKeineFreigabe`).
// Erst wo einer steht, entscheidet die Freigabe etwas, und nur dort ist ein
// Lauf darüber eine Aussage.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// verschlossen ist ein Faktorprüfer, der jeden Nachweis abweist.
func verschlossen() *pruefdoppel {
	return &pruefdoppel{fehler: domain.FehlenderFaktor(domain.FaktorTOTP)}
}

// zweiNutzer gibt eine Verwaltung mit **zwei** Nutzern, je einem Profil.
//
// **Der Zeiger steht auf Profil 1** — der arbeitende Mensch ist damit Nutzer 1,
// und Profil 2 gehört einem anderen.
//
// **Und beide Nutzer haben einen zweiten Faktor**, den niemand vorzeigen kann:
// so entscheidet allein die Freigabe.
func zweiNutzer(t *testing.T, freigaben []domain.Profilfreigabe,
	sichtbar string,
) (*service.Nutzerverwaltung, *nutzerdoppel, *freigabendoppel) {
	t.Helper()

	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}, {ID: 2, Name: "Zweiter"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Meins"}},
			2: {{ID: 2, NutzerID: 2, Name: "Fremd", Sichtbarkeit: sichtbar}},
		},
		zeiger: 1,
	}

	doppel := neuesFreigabendoppel()
	doppel.abgelegt = freigaben

	verwaltung := service.NeueNutzerverwaltung(ablage).
		MitFaktor(verschlossen()).
		MitFreigaben(service.NeueFreigabenverwaltung(doppel, ablage))

	return verwaltung, ablage, doppel
}

// **Ohne Freigabe kommt niemand in ein fremdes Profil.**
//
// **Am Zeiger und nicht am Rückgabewert** (§10): eine Abnahme ist
// beobachtbares Verhalten.
func TestOhneFreigabeKeinFremdesProfil(t *testing.T) {
	t.Parallel()

	verwaltung, ablage, _ := zweiNutzer(t, nil, "")

	if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler == nil {
		t.Fatal("der Wechsel kam durch")
	}

	if ablage.zeiger != 1 {
		t.Errorf("der Zeiger wanderte auf %d", ablage.zeiger)
	}
}

// **Ein Leser kommt hinein — seit Werkstück 16.**
//
// # Hier stand das Gegenteil, und das war richtig
//
// **„Ein Leser kommt auch nicht hinein" war die Aussage von Vermerk
// [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)**,
// und der Grund stand daneben: ein Nur-Lese-Modus **innerhalb** eines Profils
// braucht einen lückenlosen Schreibschutz, und den gab es nicht. Eine Rolle,
// die hineinlässt und aufs Nichtschreiben vertraut, wäre eine Zusage an 116
// Bindungen gewesen — **am 2026-09-06 gezählt** — sie bricht beim ersten vergessenen Ruf still (§10).
//
// **Den Schutz gibt es jetzt**, im Treiber: `storage.OeffnenNurLesend` nimmt
// `PRAGMA query_only`, und das gilt für die **Verbindung**. Damit fällt die
// Beschränkung, und [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)
// löst 0070 in diesem Punkt ab.
//
// **Was er darin darf, entscheidet die Verbindung** — der Lauf dafür steht in
// `nurlesen_test.go` und in `internal/storage`.
func TestEinLeserKommtHinein(t *testing.T) {
	t.Parallel()

	verwaltung, ablage, _ := zweiNutzer(t, []domain.Profilfreigabe{
		{ProfilID: 2, NutzerID: 1, Rolle: domain.RolleLeser},
	}, "")

	if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler != nil {
		t.Fatalf("ein Leser kam nicht hinein: %v", fehler)
	}

	if ablage.zeiger != 2 {
		t.Errorf("der Zeiger steht auf %d, erwartet war 2", ablage.zeiger)
	}
}

// **Und die Sichtbarkeit `bestand` genügt weiterhin nicht.**
//
// # Warum das so bleibt, obwohl `leser` jetzt hineinkommt
//
// **Weil die Sichtbarkeit etwas anderes verspricht.** An
// [domain.SichtbarImBestand] steht seit Werkstück 11: „der Name steht in der
// Übersicht, der Inhalt nicht." Wer die Einstellung gesetzt hat, hat das
// gelesen.
//
// **Bis Werkstück 16 machte es keinen Unterschied** — `leser` kam ohnehin
// nicht hinein, und beide Wege endeten am selben Nein. Seither nicht mehr, und
// eine Sichtbarkeit, die plötzlich den Inhalt öffnet, wäre eine
// stillschweigende Änderung bestehender Semantik (§15, Regel 4).
//
// **Wer den Inhalt öffnen will, vergibt eine Freigabe** — eine Handlung, die
// in einer Liste steht, die man ansehen kann.
func TestSichtbarImBestandGenuegtNicht(t *testing.T) {
	t.Parallel()

	verwaltung, ablage, _ := zweiNutzer(t, nil, string(domain.SichtbarImBestand))

	if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler == nil {
		t.Fatal("die Sichtbarkeit allein ließ hinein")
	}

	if ablage.zeiger != 1 {
		t.Errorf("der Zeiger wanderte auf %d", ablage.zeiger)
	}
}

// **Ein Mitarbeiter kommt hinein — ohne den Code des Eigners.**
//
// **Das ist der Lauf, um den es geht.** Der Faktor des Eigners weist jeden
// Nachweis ab; wer eine Freigabe hat, kommt trotzdem herein. Genau das soll
// eine Freigabe sein: der Eigner lässt jemanden ein, ohne ihm seinen Faktor
// zu geben.
//
// **Ohne diesen Fall stünden die drei darüber auch dann grün da, wenn jeder
// Wechsel abgewiesen würde** (§10).
func TestEinMitarbeiterKommtHinein(t *testing.T) {
	t.Parallel()

	verwaltung, ablage, _ := zweiNutzer(t, []domain.Profilfreigabe{
		{ProfilID: 2, NutzerID: 1, Rolle: domain.RolleMitarbeiter},
	}, "")

	profil, fehler := verwaltung.Wechseln(t.Context(), 2, "")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if profil.ID != 2 || ablage.zeiger != 2 {
		t.Errorf("gewechselt wurde auf %d, der Zeiger steht auf %d",
			profil.ID, ablage.zeiger)
	}
}

// **Und in ein eigenes Profil kommt man ohne jede Freigabe.**
func TestDasEigeneProfilBrauchtKeineFreigabe(t *testing.T) {
	t.Parallel()

	verwaltung, _, doppel := zweiNutzer(t, nil, "")

	// **Mit einem Faktor, den man vorzeigen kann** — der eigene Weg führt
	// über ihn und nicht über eine Rolle.
	if _, fehler := verwaltung.MitFaktor(&pruefdoppel{}).Wechseln(
		t.Context(), 1, ""); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	// **Und der Bestand wurde dafür nicht gefragt** — die Antwort steht schon
	// am Profil.
	if doppel.gefragt != 0 {
		t.Errorf("der Bestand wurde %dmal gefragt", doppel.gefragt)
	}
}

// **Auch der eigene Weg fragt keine Freigabe, wenn der Faktor abweist.**
//
// **Der Fall gehört dazu, weil er den Kurzschluss betritt**: wer in sein
// eigenes Profil will und den Code nicht kennt, bleibt draußen — und die
// Verwaltung fragt den Bestand gar nicht erst nach einer Rolle, die ihm
// nichts nützte.
func TestDasEigeneProfilFragtAuchBeimNeinKeineFreigabe(t *testing.T) {
	t.Parallel()

	verwaltung, _, doppel := zweiNutzer(t, nil, "")

	if _, fehler := verwaltung.Wechseln(
		t.Context(), 1, ""); !errors.Is(fehler, domain.ErrZweiterFaktor) {
		t.Fatalf("Fehler = %v, erwartet ErrZweiterFaktor", fehler)
	}

	if doppel.gefragt != 0 {
		t.Errorf("der Bestand wurde %dmal gefragt", doppel.gefragt)
	}
}

// **Ohne Freigabennaht wird nicht gefragt.**
//
// **Eine Verwaltung ohne sie bricht keine Zusage, sondern macht keine** —
// dieselbe Begründung wie beim zweiten Faktor. Ohne diesen Fall bliebe der
// Zweig unbetreten (§3).
//
// **Der Faktor weist ab**, damit der Weg überhaupt bis zur fehlenden Naht
// läuft: ein glückender Faktor endete davor.
func TestOhneFreigabennahtWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueNutzerverwaltung(&nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}, {ID: 2, Name: "Zweiter"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Meins"}},
			2: {{ID: 2, NutzerID: 2, Name: "Fremd"}},
		},
		zeiger: 1,
	}).MitFaktor(verschlossen())

	// **Es bleibt beim Faktorfehler** — ohne Naht kommt kein zweiter dazu.
	if _, fehler := verwaltung.Wechseln(
		t.Context(), 2, ""); !errors.Is(fehler, domain.ErrZweiterFaktor) {
		t.Errorf("Fehler = %v, erwartet ErrZweiterFaktor", fehler)
	}
}

// **Ein Bestand, der das gezeigte Profil nicht hergibt, reißt ab.**
//
// **Und lässt nicht durch**: wer nicht sagen kann, wer arbeitet, kann auch
// nicht sagen, ob er darf — ein `nil` hier wäre eine Erlaubnis aus Unwissen
// (§8).
func TestOhneGezeigtesProfilWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	// **Der Zeiger steht auf einem Profil, das die Übersicht nicht führt.**
	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 2, NutzerID: 1, Name: "Fremd"}},
		},
		zeiger: 4711,
	}

	verwaltung := service.NeueNutzerverwaltung(ablage).
		MitFaktor(verschlossen()).
		MitFreigaben(service.NeueFreigabenverwaltung(
			neuesFreigabendoppel(), ablage))

	// **Und zwar mit dem Lesefehler und nicht mit dem Faktorfehler**: was hier
	// schiefging, ist der Bestand und nicht der Nachweis.
	_, fehler := verwaltung.Wechseln(t.Context(), 2, "")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler = %v, erwartet ErrNotFound", fehler)
	}
}

// **Ein stolpernder Freigabenbestand reißt ab und lässt nicht durch.**
//
// **Ein Lesefehler ist kein Nein** (§8): käme er als „dir fehlt die Rolle"
// zurück, schickte er jemanden zu einem Verwalter, während in Wahrheit die
// Platte klemmt.
func TestEinStolpernderBestandReisstDenWechselAb(t *testing.T) {
	t.Parallel()

	verwaltung, _, doppel := zweiNutzer(t, nil, "")
	doppel.kaputt = errors.New("die Platte klemmt")

	_, fehler := verwaltung.Wechseln(t.Context(), 2, "")
	if !errors.Is(fehler, doppel.kaputt) {
		t.Errorf("Fehler = %v, erwartet den Bestandsfehler", fehler)
	}
}

// **Ohne zweiten Faktor sperrt keine Freigabe.**
//
// # Warum dieser Lauf der wichtigste dieser Datei ist
//
// **Weil die erste Fassung von Werkstück 11 jeden zweiten Nutzer aussperrte.**
// Sie fragte die Freigabe **unbedingt**, und wer arbeitet, ist der Eigner des
// gezeigten Profils — also kam niemand je in das Profil eines Nutzers, der
// gerade nicht gezeigt wird, und ein frisch angelegter Nutzer war von Geburt
// an unerreichbar. Angestrichen hat es der Flächenlauf: **sieben** Läufe
// wurden rot, und keiner davon an einem Fehler des Bestands.
//
// **Ein Profil ohne Schloss wird nicht bewacht** — dieselbe Zusage, die der
// zweite Faktor seit Werkstück 7 macht: wer keinen hat, wird nicht gefragt.
// Die Freigabe ist der **Ersatz** für diesen Nachweis, nicht eine zweite
// Hürde davor
// ([0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)).
func TestOhneZweitenFaktorSperrtKeineFreigabe(t *testing.T) {
	t.Parallel()

	verwaltung, ablage, _ := zweiNutzer(t, nil, "")

	profil, fehler := verwaltung.MitFaktor(nil).Wechseln(t.Context(), 2, "")
	if fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}

	if profil.ID != 2 || ablage.zeiger != 2 {
		t.Errorf("gewechselt wurde auf %d, der Zeiger steht auf %d",
			profil.ID, ablage.zeiger)
	}
}

// **Und wer sich als Eigner ausweist, braucht keine Freigabe.**
//
// **Das ist der Rückweg in das eigene Profil**: Nutzer 1 steht in seinem und
// will in ein fremdes, hat dort keine Rolle — und kommt hinein, weil er den
// Code kennt. Für den Eigner des Ziels ist das der Heimweg.
func TestWerSichAusweistBrauchtKeineFreigabe(t *testing.T) {
	t.Parallel()

	verwaltung, _, doppel := zweiNutzer(t, nil, "")

	if _, fehler := verwaltung.MitFaktor(&pruefdoppel{}).Wechseln(
		t.Context(), 2, "123456"); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}

	// **Und die Freigaben wurden dafür nicht gelesen** — der Faktor entschied.
	if doppel.gefragt != 0 {
		t.Errorf("der Bestand wurde %dmal gefragt", doppel.gefragt)
	}
}
