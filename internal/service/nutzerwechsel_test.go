// **Wer arbeitet, ändert sich nicht durch einen Profilwechsel** — Werkstück 17
// von Phase 4.12 (#1400),
// nach [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).
//
// # Die Abnahme dieses Werkstücks
//
// **Freigaben lassen sich nicht mehr aneinanderhängen.** Wer auf B eine
// Freigabe hat, erreicht damit **nicht**, was B erreicht — geprüft wird gegen
// die eigenen Freigaben, gleich wo der Zeiger steht.
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dreiNutzer baut die Lage aus #1400:
// Erster darf zu Zweitem, Zweiter darf zu Drittem, Erster darf **nicht** zu
// Drittem.
func dreiNutzer(t *testing.T) (*service.Nutzerverwaltung, *nutzerdoppel) {
	t.Helper()

	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{
			{ID: 1, Name: "Erster"}, {ID: 2, Name: "Zweiter"}, {ID: 3, Name: "Dritter"},
		},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Meins"}},
			2: {{ID: 2, NutzerID: 2, Name: "Zweites"}},
			3: {{ID: 3, NutzerID: 3, Name: "Drittes"}},
		},
		zeiger: 1,
	}

	doppel := neuesFreigabendoppel()
	doppel.abgelegt = []domain.Profilfreigabe{
		{ProfilID: 2, NutzerID: 1, Rolle: domain.RolleMitarbeiter},
		{ProfilID: 3, NutzerID: 2, Rolle: domain.RolleMitarbeiter},
	}

	// **Mit verschlossenem zweiten Faktor**, und das ist keine Beigabe: die
	// Regel aus Werkstück 11 lautet „ausweisen **oder** freigegeben sein". Wo
	// niemand einen Faktor führt, kommt jeder überall hinein — dann prüfte
	// dieser Lauf die Freigaben gar nicht, sondern die Abwesenheit von
	// Schutz.
	//
	// **Gefunden hat es der erste Anlauf**: ohne diese Zeile glückte Schritt 2
	// auch mit der Korrektur, und zwar aus dem richtigen Grund.
	verwaltung := service.NeueNutzerverwaltung(ablage).
		MitFaktor(verschlossen()).
		MitFreigaben(service.NeueFreigabenverwaltung(doppel, ablage))

	return verwaltung, ablage
}

// **Die Kette bricht.**
//
// **Gemessen am 2026-09-06, bevor es diesen Lauf gab**: derselbe Ablauf
// glückte, und der Zeiger stand danach auf dem dritten Profil. Geprüft wurde
// als *Zweiter*, obwohl der Erste arbeitete.
func TestEineFreigabeReichtNichtWeiterAlsSieVergebenIst(t *testing.T) {
	verwaltung, ablage := dreiNutzer(t)

	// Schritt 1: der Erste geht in das Profil des Zweiten. Dafür hat er eine
	// Freigabe.
	if _, err := verwaltung.Wechseln(t.Context(), 2, ""); err != nil {
		t.Fatalf("Schritt 1 war erlaubt und scheiterte: %v", err)
	}

	if ablage.zeiger != 2 {
		t.Fatalf("der Zeiger steht auf %d, erwartet war 2", ablage.zeiger)
	}

	// **Und er ist dabei nicht zum Zweiten geworden.**
	werArbeitet, err := verwaltung.WerArbeitet(t.Context())
	if err != nil {
		t.Fatalf("WerArbeitet: %v", err)
	}

	if werArbeitet != 1 {
		t.Errorf("es arbeitet %d, erwartet war der Erste", werArbeitet)
	}

	// Schritt 2: von dort aus in das Profil des Dritten. Der **Erste** hat
	// darauf keine Freigabe — und er ist weiterhin der Erste.
	_, fehler := verwaltung.Wechseln(t.Context(), 3, "")

	// **Die Absage ist die des zweiten Faktors und nicht die der Erlaubnis**,
	// und das ist richtig: die Regel aus Werkstück 11 lautet „ausweisen
	// **oder** freigegeben sein". Der Erste ist auf dem dritten Profil nicht
	// freigegeben, also bleibt ihm der Ausweis — und den hat er hier nicht.
	//
	// **Vorher glückte derselbe Ruf ohne alles**, weil geprüft wurde, was der
	// *Zweite* darf.
	if !errors.Is(fehler, domain.ErrZweiterFaktor) {
		t.Fatalf("Fehler %v, erwartet war eine Absage", fehler)
	}

	// **Am Zeiger und nicht am Rückgabewert** (§10): eine Abnahme ist
	// beobachtbares Verhalten.
	if ablage.zeiger != 2 {
		t.Errorf("der Zeiger steht auf %d — der Wechsel fand statt", ablage.zeiger)
	}

	// **Und die Identität hat sich dabei nicht verschoben.**
	nachher, err := verwaltung.WerArbeitet(t.Context())
	if err != nil {
		t.Fatalf("WerArbeitet: %v", err)
	}

	if nachher != 1 {
		t.Errorf("es arbeitet %d, erwartet war weiterhin der Erste", nachher)
	}
}

// **Und der Weg zurück ins eigene Profil bleibt offen.**
//
// **Ohne diesen Lauf wäre die Sperre oben auch dann grün, wenn sie **jeden**
// Wechsel abwiese** — und dann säße der Erste in einem fremden Profil fest.
func TestWerInEinemFremdenProfilStehtKommtInSeinesZurueck(t *testing.T) {
	verwaltung, ablage := dreiNutzer(t)

	if _, err := verwaltung.Wechseln(t.Context(), 2, ""); err != nil {
		t.Fatalf("hinein: %v", err)
	}

	// **Zurück mit dem eigenen Ausweis.** Das eigene Profil steht unter dem
	// zweiten Faktor wie jedes andere; entscheidend ist, dass der Weg **offen
	// ist** — ohne diesen Lauf wäre die Sperre oben auch dann grün, wenn sie
	// jeden Wechsel abwiese, und der Erste säße im fremden Profil fest.
	verwaltung.MitFaktor(&pruefdoppel{})

	if _, err := verwaltung.Wechseln(t.Context(), 1, "der Ausweis"); err != nil {
		t.Fatalf("zurück ins eigene Profil: %v", err)
	}

	if ablage.zeiger != 1 {
		t.Errorf("der Zeiger steht auf %d, erwartet war 1", ablage.zeiger)
	}
}

// **Ein anderer Nutzer wird man ausdrücklich — und mit dessen Ausweis.**
//
// # Warum ein eigener Ruf und nicht ein Wechsel ohne Freigabe
//
// **Weil eine Geste nicht zwei Bedeutungen haben soll.** Ein Profilwechsel
// heißt „ich sehe da hinein"; ein Nutzerwechsel heißt „ich bin jetzt jemand
// anderes". Wer beides an einer Geste hängte, machte aus einer abgelehnten
// Freigabe stillschweigend einen Identitätswechsel — und niemand sähe, wann er
// geschah.
func TestEinAndererNutzerWirdManNurMitDessenAusweis(t *testing.T) {
	verwaltung, ablage := dreiNutzer(t)

	// **Ohne Ausweis nicht.** Der Faktor des Ziels ist verschlossen.
	if fehler := verwaltung.NutzerWechseln(t.Context(), 2, ""); fehler == nil {
		t.Fatal("der Nutzerwechsel glückte ohne Ausweis")
	}

	if ablage.nutzerzeiger != 0 {
		t.Errorf("der Nutzerzeiger steht auf %d — der Wechsel fand statt",
			ablage.nutzerzeiger)
	}

	// **Und der Faktor, der gefragt wurde, ist der des Ziels** — nicht der
	// eigene. Wer den eigenen fragte, ließe jeden zu jedem werden, der seinen
	// eigenen Code kennt.
	tor := &pruefdoppel{}
	verwaltung.MitFaktor(tor)

	if fehler := verwaltung.NutzerWechseln(t.Context(), 2, "der Ausweis"); fehler != nil {
		t.Fatalf("Nutzerwechsel: %v", fehler)
	}

	if tor.nutzerID != 2 {
		t.Errorf("gefragt wurde nach dem Faktor von %d, erwartet war 2", tor.nutzerID)
	}

	werArbeitet, err := verwaltung.WerArbeitet(t.Context())
	if err != nil {
		t.Fatalf("WerArbeitet: %v", err)
	}

	if werArbeitet != 2 {
		t.Errorf("es arbeitet %d, erwartet war der Zweite", werArbeitet)
	}

	// **Und der Zeiger steht in einem Profil des neuen Menschen.** Wer als
	// Zweiter im Profil des Ersten stünde, sähe Daten, für die er keine
	// Freigabe hat.
	if ablage.zeiger != 2 {
		t.Errorf("der Zeiger steht auf %d, erwartet war ein Profil des Zweiten",
			ablage.zeiger)
	}
}

// **Einen Nutzer, den es nicht gibt, wird niemand.**
func TestEinNutzerwechselAufNichtsWirdAbgewiesen(t *testing.T) {
	verwaltung, _ := dreiNutzer(t)
	verwaltung.MitFaktor(&pruefdoppel{})

	if fehler := verwaltung.NutzerWechseln(t.Context(), 99, "x"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war „nicht gefunden\"", fehler)
	}
}

// **Und wer sich selbst wählt, wird nicht gefragt.**
//
// **Ein Ausweis für das, was ohnehin gilt, wäre eine Hürde ohne Zweck** — und
// eine, die jemanden aussperrte, dessen Faktor gerade klemmt.
func TestWerSichSelbstWaehltWirdNichtGefragt(t *testing.T) {
	verwaltung, _ := dreiNutzer(t)

	tor := &pruefdoppel{fehler: domain.FehlenderFaktor(domain.FaktorTOTP)}
	verwaltung.MitFaktor(tor)

	if fehler := verwaltung.NutzerWechseln(t.Context(), 1, ""); fehler != nil {
		t.Fatalf("der Wechsel auf sich selbst wurde abgewiesen: %v", fehler)
	}

	if tor.nutzerID != 0 {
		t.Errorf("gefragt wurde nach dem Faktor von %d — es sollte niemand gefragt werden",
			tor.nutzerID)
	}
}

// **Das Profil einer Anmeldung wird genannt, nicht gesetzt** (Nachlauf über
// ID-02, Befund 2): die Zeiger schreibt `Sitzung.Anmelden` zusammen mit dem
// Zustand. Ohne zweite Prüfung — die strenge hat die Anmeldung verlangt.
func TestDasAnmeldeprofilIstDasErsteUndSchreibtNichts(t *testing.T) {
	verwaltung, ablage := dreiNutzer(t)
	vorher, vorherProfil := ablage.nutzerzeiger, ablage.zeiger

	profil, fehler := verwaltung.Anmeldeprofil(t.Context(), 2)
	if fehler != nil || profil != 2 {
		t.Fatalf("Profil %d: %v", profil, fehler)
	}

	if ablage.nutzerzeiger != vorher || ablage.zeiger != vorherProfil {
		t.Errorf("geschrieben: Nutzer %d, Profil %d", ablage.nutzerzeiger, ablage.zeiger)
	}

	if _, fehler := verwaltung.Anmeldeprofil(t.Context(), 99); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("unbekannt: %v", fehler)
	}
}
