package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Marken an Karten** —
// #1208.

// markenprobe merkt sich, was gesetzt und entfernt wurde.
type markenprobe struct {
	gesetzt  []string
	entfernt []string
	anKarte  []string
	fehler   error
}

func (m *markenprobe) Setzen(_ context.Context, _ int64, marke string) error {
	m.gesetzt = append(m.gesetzt, marke)

	return m.fehler
}

func (m *markenprobe) Entfernen(_ context.Context, _ int64, marke string) error {
	m.entfernt = append(m.entfernt, marke)

	return m.fehler
}

func (m *markenprobe) AnKarte(context.Context, int64) ([]string, error) {
	return m.anKarte, m.fehler
}

func (m *markenprobe) ImSet(context.Context, int64) ([]string, error) {
	return m.anKarte, m.fehler
}

func (m *markenprobe) JeKarte(context.Context, int64) (map[int64][]string, error) {
	// **Karte 7 für die Markenpflege, Karte 1 für die Kartenliste** — beide
	// Vorrichtungen benutzen dieses Doppel, und zwei wären zwei Wahrheiten.
	return map[int64][]string{7: m.anKarte, 1: m.anKarte}, m.fehler
}

// markenkarte gibt eine Karte, die zu Set 1 gehört.
type markenkarte struct{ setID int64 }

func (m markenkarte) KarteLesen(_ context.Context, karteID int64) (*domain.CardRow, error) {
	if karteID != 7 {
		return nil, domain.ErrNotFound
	}

	return &domain.CardRow{ID: 7, SetID: m.setID}, nil
}

// markenkoepfe kennt genau Set 1.
//
// **Damit ein unbekanntes Set abgewiesen wird** — für `storage` ist es
// dasselbe wie ein leeres (#1282,
// Befund 8).
type markenkoepfe struct{ fehler error }

func (m markenkoepfe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 1, Name: "Erstes", Game: "probe"}}, m.fehler
}

func (m markenkoepfe) Setwerte(
	context.Context, int64,
) ([]domain.Value, error) {
	return nil, m.fehler
}

// markenpflege baut die Pflege über den drei Doppeln.
func markenpflege(ablage *markenprobe) *service.Markenpflege {
	return service.NeueMarkenpflege(ablage, markenkarte{setID: 1}, markenkoepfe{})
}

// **Die Normalform entsteht hier und nicht an einer Fläche** (§16).
func TestEineMarkeGehtInNormalformInDenBestand(t *testing.T) {
	ablage := &markenprobe{}

	pflege := markenpflege(ablage)

	if err := pflege.Setzen(t.Context(), 1, 7, "  Noch Zu Tun "); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if len(ablage.gesetzt) != 1 || ablage.gesetzt[0] != "noch zu tun" {
		t.Errorf("gesetzt: %v", ablage.gesetzt)
	}
}

// **Und beim Entfernen genauso** — sonst träfe „Fertig" die Marke „fertig"
// nicht, die dasteht.
func TestEineMarkeWirdInNormalformEntfernt(t *testing.T) {
	ablage := &markenprobe{}

	if err := markenpflege(ablage).
		Entfernen(t.Context(), 1, 7, "FERTIG"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if len(ablage.entfernt) != 1 || ablage.entfernt[0] != "fertig" {
		t.Errorf("entfernt: %v", ablage.entfernt)
	}
}

// **Eine Karte aus einem anderen Set wird abgewiesen.**
//
// Die Fremdschlüssel halten fest, dass es die Karte gibt; dass sie zu **diesem**
// Set gehört, hält keiner.
func TestEineFremdeKarteWirdNichtMarkiert(t *testing.T) {
	ablage := &markenprobe{}

	err := service.NeueMarkenpflege(ablage, markenkarte{setID: 2}, markenkoepfe{}).
		Setzen(t.Context(), 1, 7, "fertig")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Setzen = %v, erwartet ErrNotFound", err)
	}

	if len(ablage.gesetzt) != 0 {
		t.Errorf("trotzdem gesetzt: %v", ablage.gesetzt)
	}
}

// **Und die Karte wird vor der Marke geprüft.**
//
// Wer die Marke zuerst prüfte, sagte bei einer vertippten Kartennummer „die
// Marke ist zu lang" — eine Meldung über das falsche Feld.
func TestDieKarteWirdVorDerMarkeGeprueft(t *testing.T) {
	err := service.NeueMarkenpflege(&markenprobe{},
		markenkarte{setID: 2}, markenkoepfe{}).
		Setzen(t.Context(), 1, 7, "")

	var feld *domain.FieldError
	if errors.As(err, &feld) {
		t.Fatalf("die Meldung nennt das Feld %q, gemeint war die Karte", feld.Field)
	}
}

// **Eine leere Marke wird abgewiesen, und zwar mit Feldbezug.**
func TestEineLeereMarkeWirdAbgewiesen(t *testing.T) {
	err := service.NeueMarkenpflege(&markenprobe{},
		markenkarte{setID: 1}, markenkoepfe{}).
		Setzen(t.Context(), 1, 7, "   ")

	var feld *domain.FieldError
	if !errors.As(err, &feld) || feld.Field != "marke" {
		t.Fatalf("Setzen = %v, erwartet einen Feldbefund `marke`", err)
	}
}

// **Lesen prüft die Zugehörigkeit ebenso** — sonst zählte ein Ruf die Marken
// einer fremden Karte auf.
func TestDasLesenPrueftDieZugehoerigkeit(t *testing.T) {
	pflege := service.NeueMarkenpflege(&markenprobe{anKarte: []string{"fertig"}},
		markenkarte{setID: 2}, markenkoepfe{})

	if _, err := pflege.AnKarte(t.Context(), 1, 7); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("AnKarte = %v, erwartet ErrNotFound", err)
	}
}

// **Die Aufzählung je Set und je Karte geht ohne Kartenprüfung** — es gibt
// keine einzelne Karte, die zu prüfen wäre.
func TestDieAufzaehlungenGehenUeberDasSet(t *testing.T) {
	pflege := service.NeueMarkenpflege(&markenprobe{anKarte: []string{"fertig"}},
		markenkarte{setID: 1}, markenkoepfe{})

	imSet, err := pflege.ImSet(t.Context(), 1)
	if err != nil || len(imSet) != 1 {
		t.Fatalf("ImSet = %v, %v", imSet, err)
	}

	jeKarte, err := pflege.JeKarte(t.Context(), 1)
	if err != nil || len(jeKarte[7]) != 1 {
		t.Fatalf("JeKarte = %v, %v", jeKarte, err)
	}
}

// **Und die Marken einer eigenen Karte kommen durch.**
func TestDieMarkenDerEigenenKarteKommenDurch(t *testing.T) {
	pflege := service.NeueMarkenpflege(&markenprobe{anKarte: []string{"fertig"}},
		markenkarte{setID: 1}, markenkoepfe{})

	marken, err := pflege.AnKarte(t.Context(), 1, 7)
	if err != nil {
		t.Fatalf("AnKarte: %v", err)
	}

	if len(marken) != 1 || marken[0] != "fertig" {
		t.Errorf("Marken: %v", marken)
	}
}

// **Die Marken stehen an der Zeile der Kartenliste** (#1208).
//
// Sie werden **gezeigt** und es wird nach ihnen **gefiltert**; beides braucht
// sie zu jeder Zeile, und eine Abfrage je Zeile wären bei 3.788 Karten 3.788
// Abfragen.
func TestDieKartenlisteTraegtDieMarken(t *testing.T) {
	bau := listenbauMitMarken(t, &markenprobe{anKarte: []string{"fertig"}})

	aufbau, err := bau.Aufbau(t.Context(), 1)
	if err != nil {
		t.Fatalf("Aufbau: %v", err)
	}

	if len(aufbau.Zeilen) == 0 {
		t.Fatal("keine Zeilen")
	}

	gefunden := false

	for _, zeile := range aufbau.Zeilen {
		if len(zeile.Marken) == 1 && zeile.Marken[0] == "fertig" {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("keine Zeile trägt ihre Marke: %+v", aufbau.Zeilen)
	}
}

// **Ein Lesefehler hält die Liste nicht auf** (§8).
//
// Die Karten sind der Grund, warum jemand hier ist; eine Zeile ohne Marken ist
// eine Zeile.
func TestEinMarkenfehlerLaesstDieListeStehen(t *testing.T) {
	bau := listenbauMitMarken(t, &markenprobe{fehler: errListenmarken})

	aufbau, err := bau.Aufbau(t.Context(), 1)
	if err != nil {
		t.Fatalf("Aufbau: %v", err)
	}

	if len(aufbau.Zeilen) == 0 {
		t.Fatal("keine Zeilen")
	}

	for _, zeile := range aufbau.Zeilen {
		if len(zeile.Marken) != 0 {
			t.Errorf("eine Zeile trägt Marken, obwohl das Lesen scheiterte: %+v", zeile)
		}
	}
}

var errListenmarken = errors.New("planmäßig gescheitert")

// listenbauMitMarken gibt einen Listenbau über der Vorrichtung von
// `kartenliste_test.go`, mit dieser Markenquelle.
func listenbauMitMarken(t *testing.T, marken service.Markenjekarte) *service.Kartenlistenbau {
	t.Helper()

	return kartenliste(dreiFelder()).MitMarken(marken)
}

// **Ein Lesefehler hält jeden der vier Wege an** (§8).
//
// Die Karte muss zum Set gehören; ist der Bestand nicht zu lesen, ist die
// Frage **unbeantwortet** — und das ist etwas anderes als „gehört nicht dazu".
func TestEinLesefehlerHaeltDieMarkenpflegeAn(t *testing.T) {
	pflege := service.NeueMarkenpflege(&markenprobe{},
		kaputteKarte{}, markenkoepfe{})

	for name, ruf := range map[string]func() error{
		"Setzen":    func() error { return pflege.Setzen(t.Context(), 1, 7, "a") },
		"Entfernen": func() error { return pflege.Entfernen(t.Context(), 1, 7, "a") },
		"AnKarte": func() error {
			_, err := pflege.AnKarte(t.Context(), 1, 7)

			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if !errors.Is(ruf(), errKartenlesen) {
				t.Error("der Lesefehler ging verloren")
			}
		})
	}
}

// kaputteKarte kann keine Karte lesen.
type kaputteKarte struct{}

func (kaputteKarte) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return nil, errKartenlesen
}

var errKartenlesen = errors.New("planmäßig gescheitert")

// **Ein unbekanntes Set wird abgewiesen, nicht als leer beantwortet** —
// #1282,
// Befund 8.
//
// # Warum das an einer Fläche nicht reicht
//
// `storage` gibt für ein unbekanntes Set eine **leere Liste** und keinen
// Fehler — der JOIN findet einfach keine Zeile. Im Fenster ist das richtig:
// dort kann niemand ein Set wählen, das es nicht gibt.
//
// **An den drei anderen Flächen kann er es.** `mse marken 4711`,
// `lesen.marken` und `GET /sets/4711/marken` sind Tippfehler, und „keine
// Marke" wäre darauf eine irreführende Antwort: ein Skript schlösse daraus,
// das Set sei leer.
//
// Der Kartenbefehl schützt dieselbe Storage-Eigenschaft seit #1253 — dort
// aber **im Befehl**, und deshalb hatten die anderen drei Flächen sie nicht.
// Hier ist sie im Dienst, und damit an allen vier (§16).
func TestEinUnbekanntesSetWirdBeiDenMarkenAbgewiesen(t *testing.T) {
	pflege := markenpflege(&markenprobe{anKarte: []string{"fertig"}})

	if _, err := pflege.ImSet(t.Context(), 4711); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ImSet(4711) = %v, erwartet ErrNotFound", err)
	}

	if _, err := pflege.JeKarte(t.Context(), 4711); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("JeKarte(4711) = %v, erwartet ErrNotFound", err)
	}

	// **Und das bekannte Set kommt weiter durch.** Ohne diese Hälfte wäre der
	// Lauf auch grün, wenn die Marken gar nicht mehr zu lesen wären.
	if _, err := pflege.ImSet(t.Context(), 1); err != nil {
		t.Errorf("ImSet(1) = %v", err)
	}
}

// **Und ein Fehler beim Lesen der Sets kommt durch, statt zu verschwinden.**
func TestEinFehlerBeimSetlesenErreichtDieMarken(t *testing.T) {
	kaputt := errors.New("der Bestand antwortet nicht")

	pflege := service.NeueMarkenpflege(&markenprobe{},
		markenkarte{setID: 1}, markenkoepfe{fehler: kaputt})

	if _, err := pflege.ImSet(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("ImSet = %v, erwartet %v", err, kaputt)
	}
}
