package service_test

// **Was eine Wahl zurücknimmt** — Werkstück O4 von Phase 5.2
// (#1428).
//
// **Eine eigene Datei**, weil `ruecksetzung_test.go` über die Zeilengrenze aus
// §2 ging. Der Schnitt folgt der Frage: dort stehen die Arten, die **löschen**,
// hier die drei, die nur eine **Wahl** zurücknehmen — und die Zusage dahinter
// ist eine andere: zurückgesetzt heißt gelöscht, und damit geerbt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Layoutrücksetzung fasst ihre vier an und keinen mehr.**
func TestDieLayoutruecksetzungFasstNurIhreSchluesselAn(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}

	if _, fehler := ruecksetzung(doppel).Zuruecksetzen(
		t.Context(), domain.RuecksetzungLayout,
		service.Ruecksetzungsziel{}, "", ""); fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	if doppel.zurueck != len(domain.Layouteinstellungen()) {
		t.Errorf("angefasst wurden %d Schlüssel", doppel.zurueck)
	}

	// **Und die Profilrücksetzung nimmt alles** — `-1` ist die Marke des
	// Doppels für „ohne Schlüssel".
	if _, fehler := ruecksetzung(doppel).Zuruecksetzen(
		t.Context(), domain.RuecksetzungProfileinstellungen,
		service.Ruecksetzungsziel{ProfilID: 3}, "", ""); fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	if doppel.zurueck != -1 {
		t.Errorf("die Profilrücksetzung nannte %d Schlüssel", doppel.zurueck)
	}
}

// **Die Nutzereinstellungen gehen ihren eigenen Weg.**
func TestDieNutzereinstellungenGehenIhrenEigenenWeg(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungNutzereinstellungen,
		service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if len(vorschau.Betroffen) != 1 ||
		vorschau.Betroffen[0].Was != domain.BetroffenEinstellungen {
		t.Errorf("die Vorschau ist %+v", vorschau.Betroffen)
	}

	bilanz, ausfuehrfehler := ruecksetzung(doppel).Zuruecksetzen(
		t.Context(), domain.RuecksetzungNutzereinstellungen,
		service.Ruecksetzungsziel{}, "", "")
	if ausfuehrfehler != nil {
		t.Fatalf("zurücksetzen: %v", ausfuehrfehler)
	}

	if bilanz.Zurueckgesetzt != 7 {
		t.Errorf("zurückgesetzt wurden %d", bilanz.Zurueckgesetzt)
	}
}

// **Die Profileinstellungen nennen, was dasteht.**
func TestDieProfileinstellungenNennenWasDasteht(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{umfang: []domain.Betroffenes{
		{Was: domain.BetroffenSets, Anzahl: 3},
		{Was: domain.BetroffenEinstellungen, Anzahl: 2, Namen: []string{"sprache"}},
	}}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungProfileinstellungen,
		service.Ruecksetzungsziel{ProfilID: 3})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	// **Nur die Einstellungen** — die Sets gehören nicht dazu.
	if len(vorschau.Betroffen) != 1 || vorschau.Betroffen[0].Anzahl != 2 {
		t.Errorf("die Vorschau ist %+v", vorschau.Betroffen)
	}
}

// **Ein Profil ohne Wahl nennt keine Gruppe.**
func TestEinProfilOhneWahlNenntKeineGruppe(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{umfang: []domain.Betroffenes{
		{Was: domain.BetroffenSets, Anzahl: 3},
	}}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungProfileinstellungen,
		service.Ruecksetzungsziel{ProfilID: 3})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if len(vorschau.Betroffen) != 0 {
		t.Errorf("die Vorschau ist %+v", vorschau.Betroffen)
	}
}

// **Die Vorschau nennt das Profil, das gelöscht würde — beim Namen.**
//
// **In der laufenden Anwendung nachgesehen** (2026-09-10): hier stand
// „Profile: 1" ohne den Namen, und genau ihn soll jemand gleich darunter
// abtippen.
func TestDieProfilvorschauNenntDenNamen(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		profile: []domain.Profil{{ID: 2, Name: "Deins"}, {ID: 3, Name: "Meins"}},
	}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(
		t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 3})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if !vorschau.NameNoetig {
		t.Error("die Profillöschung verlangt keinen Namen")
	}

	for _, gruppe := range vorschau.Betroffen {
		if gruppe.Was != domain.BetroffenProfile {
			continue
		}

		if gruppe.Anzahl != 1 || len(gruppe.Namen) != 1 || gruppe.Namen[0] != "Meins" {
			t.Errorf("die Gruppe der Profile ist %+v", gruppe)
		}

		return
	}

	t.Errorf("die Vorschau nennt keine Profile: %+v", vorschau.Betroffen)
}

// **Was jemand gesehen hat, wird zurückgesetzt — und sonst nichts.**
//
// Befund 12 des Zweitblicks über Phase 5.2: die Ausführung rechnete eine
// **neue** Vorschau und löschte auf deren Stand. Zwischen dem Hinsehen und dem
// Klicken kann ein Set dazukommen — und es ging ohne Zustimmung mit.
func TestEineUeberholteVorschauSetztNichtsZurueck(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		umfang: []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins"}},
		},
		nutzer: 2,
		profile: []domain.Profil{
			{ID: 1, Name: "Meins", NutzerID: 1},
			{ID: 2, Name: "Zweites", NutzerID: 1},
		},
	}

	dienst := ruecksetzung(doppel)

	vorschau, fehler := dienst.Vorschau(t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 1, Name: "Meins"})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if vorschau.Revision == "" {
		t.Fatal("die Vorschau trägt keine Revision")
	}

	// **Zwischen Hinsehen und Klicken kommt ein Set dazu.**
	doppel.umfang = []domain.Betroffenes{
		{Was: domain.BetroffenSets, Anzahl: 4, Namen: []string{"Eins"}},
	}

	_, ausfuehrfehler := dienst.Zuruecksetzen(t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 1, Name: "Meins"}, "", vorschau.Revision)

	if !errors.Is(ausfuehrfehler, domain.ErrConflict) {
		t.Fatalf("abgewiesen wurde mit %v", ausfuehrfehler)
	}

	if doppel.geloescht != "" {
		t.Errorf("gelöscht wurde trotzdem: %q", doppel.geloescht)
	}

	// **Die Gegenprobe**: mit der Revision des heutigen Standes geht es.
	heute, fehler := dienst.Vorschau(t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 1, Name: "Meins"})
	if fehler != nil {
		t.Fatalf("die zweite Vorschau: %v", fehler)
	}

	if _, fehler := dienst.Zuruecksetzen(t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 1, Name: "Meins"}, "", heute.Revision,
	); fehler != nil {
		t.Fatalf("mit der heutigen Revision: %v", fehler)
	}
}

// **Und was erst im Schreibgeschäft dazukommt, hält es dort auf.**
//
// Befund 9 des Nachlaufs zum Zweitblick über Phase 5.2: der Vergleich lag
// **vor** den Schreibaufrufen. Zwischen dem Zählen und dem Löschen liegt eine
// Lücke; wer sie trifft, verliert etwas, das in keiner Vorschau stand.
//
// **Der Lauf darüber prüft die Lücke nicht** — dort ändert sich der Stand vor
// dem Ruf, und schon der Vergleich davor weist ab. Hier stimmt der Stand beim
// Nachsehen und ist beim Löschen ein anderer: nur die Bindung im Geschäft
// sieht das.
func TestWasErstImGeschaeftDazukommtWirdNichtGeloescht(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		umfang: []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins"}},
		},
		nutzer:  2,
		profile: []domain.Profil{{ID: 1, Name: "Meins", NutzerID: 1}},

		// **Im Geschäft steht ein Set mehr.**
		beimLoeschen: []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 4, Namen: []string{"Eins"}},
		},
	}

	dienst := ruecksetzung(doppel)
	ziel := service.Ruecksetzungsziel{ProfilID: 1, Name: "Meins"}

	vorschau, fehler := dienst.Vorschau(t.Context(), domain.RuecksetzungProfil, ziel)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	_, ausfuehrfehler := dienst.Zuruecksetzen(t.Context(),
		domain.RuecksetzungProfil, ziel, "", vorschau.Revision)

	if !errors.Is(ausfuehrfehler, domain.ErrConflict) {
		t.Fatalf("abgewiesen wurde mit %v", ausfuehrfehler)
	}

	if doppel.geloescht != "" {
		t.Errorf("gelöscht wurde trotzdem: %q", doppel.geloescht)
	}

	// **Die Gegenprobe**: steht im Geschäft derselbe Stand, geht es durch.
	doppel.beimLoeschen = doppel.umfang

	if _, fehler := dienst.Zuruecksetzen(t.Context(),
		domain.RuecksetzungProfil, ziel, "", vorschau.Revision,
	); fehler != nil {
		t.Fatalf("bei gleichem Stand: %v", fehler)
	}
}

// **Ein Mensch wird ebenso gebunden gelöscht.**
//
// **Nicht dieselbe Zeile mit anderem Argument**: beim Nutzer zählt das
// Geschäft **alle** seine Profile, und der Umfang entsteht aus mehreren
// Stücken. Ein Lauf nur über das Profil ließe offen, ob das Zusammenzählen
// in der Bindung dasselbe ergibt wie in der Vorschau.
func TestAuchDieNutzerloeschungHaengtAmGesehenenStand(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		umfang: []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins"}},
		},
		nutzer: 2,
		profile: []domain.Profil{
			{ID: 1, Name: "Meins", NutzerID: 1},
			{ID: 2, Name: "Zweites", NutzerID: 1},
		},
	}

	dienst := ruecksetzung(doppel)
	ziel := service.Ruecksetzungsziel{Name: "David"}

	vorschau, fehler := dienst.Vorschau(t.Context(), domain.RuecksetzungNutzer, ziel)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	// **Zuerst die Gegenprobe**: unverändert geht es durch — sonst sagte das
	// Nein darunter nur, dass die beiden Wege verschieden rechnen.
	if _, fehler := dienst.Zuruecksetzen(t.Context(),
		domain.RuecksetzungNutzer, ziel, "", vorschau.Revision,
	); fehler != nil {
		t.Fatalf("bei gleichem Stand: %v", fehler)
	}

	doppel.geloescht = ""
	doppel.beimLoeschen = []domain.Betroffenes{
		{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins", "Neu"}},
	}

	if _, fehler := dienst.Zuruecksetzen(t.Context(),
		domain.RuecksetzungNutzer, ziel, "", vorschau.Revision,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("abgewiesen wurde mit %v", fehler)
	}

	if doppel.geloescht != "" {
		t.Errorf("gelöscht wurde trotzdem: %q", doppel.geloescht)
	}
}

// **Ohne Revision wird nichts gelöscht.**
//
// Befund 9 des Nachlaufs zum Zweitblick über Phase 5.2: die Bindung an die
// Vorschau war durch eine **leere** Revision vollständig zu umgehen — und
// damit war sie keine. Für die drei Arten, die nur eine Wahl zurücknehmen,
// bleibt sie wahlfrei; dort verschwindet nichts, was jemand nicht gesehen hat.
func TestOhneRevisionWirdNichtsGeloescht(t *testing.T) {
	t.Parallel()

	loescht := map[domain.Ruecksetzungsart]bool{
		domain.RuecksetzungProfil:              true,
		domain.RuecksetzungNutzer:              true,
		domain.RuecksetzungAnwendung:           true,
		domain.RuecksetzungLayout:              false,
		domain.RuecksetzungProfileinstellungen: false,
		domain.RuecksetzungNutzereinstellungen: false,
	}

	for art, gesperrt := range loescht {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			doppel := &ruecksetzungsdoppel{
				nutzer:  2,
				profile: []domain.Profil{{ID: 1, Name: "Meins", NutzerID: 1}},
			}

			_, fehler := ruecksetzung(doppel).Zuruecksetzen(t.Context(), art,
				service.Ruecksetzungsziel{ProfilID: 1, Name: "Meins"}, "nachweis", "")

			if gesperrt {
				if !errors.Is(fehler, domain.ErrValidation) {
					t.Fatalf("abgewiesen wurde mit %v", fehler)
				}

				if doppel.geloescht != "" || doppel.neuanfaenge != 0 {
					t.Errorf("getan wurde trotzdem: %q, %d Neuanfänge",
						doppel.geloescht, doppel.neuanfaenge)
				}

				return
			}

			// **Die Gegenprobe steht in derselben Tafel**: ohne sie sagte der
			// Lauf nur, dass irgendetwas abweist — nicht, dass die Grenze
			// zwischen Löschen und Wählen verläuft.
			if fehler != nil {
				t.Errorf("eine Wahl zurückzunehmen verlangt eine Revision: %v", fehler)
			}
		})
	}
}
