package service_test

// **Eine Einladung genügt** — NET-01b
// (#1954).
//
// # Was hier gemessen wird
//
// **Beide Seiten an einer Probeleitung**: Anna stellt aus und beantwortet,
// Bert nimmt an und meldet sich. Die echte Leitung fährt der Lauf in
// `internal/app`; hier geht es um das, was jede Seite entscheidet.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func (a *einladungsablage) Ausgestellt(
	_ context.Context, nonce string, _ domain.Zweck, ablauf, _ time.Time,
) error {
	if a.ausstellfehler != nil {
		return a.ausstellfehler
	}

	a.ausgestellt[nonce] = ablauf

	return nil
}

func (a *einladungsablage) Einloesen(
	_ context.Context, nonce string, kontakt domain.Peerkontakt, jetzt time.Time,
) error {
	ablauf, da := a.ausgestellt[nonce]

	switch {
	case a.einloesefehler != nil:
		return a.einloesefehler
	case !da:
		return domain.ErrEinladungFremd
	case a.eingeloest[nonce]:
		return domain.ErrEinladungVerbraucht
	case !jetzt.Before(ablauf):
		return domain.ErrEinladungAbgelaufen
	}

	a.eingeloest[nonce] = true
	a.kontakte[kontakt.Kennung] = kontakt

	return nil
}

// probeleitung ist eine Leitung, die aufschreibt und auf Zuruf antwortet.
type probeleitung struct {
	stand     service.Sitzungsstand
	gesendet  [][]byte
	antworten [][]byte

	sendefehler, leerefehler, empfangefehler error
	geschlossen                              bool
}

func (l *probeleitung) Stand() service.Sitzungsstand { return l.stand }

func (l *probeleitung) Sende(_ context.Context, nachricht []byte, _ service.Zustellung) error {
	l.gesendet = append(l.gesendet, nachricht)

	return l.sendefehler
}

func (l *probeleitung) Leere(context.Context) error { return l.leerefehler }

func (l *probeleitung) Empfange(context.Context) ([]byte, error) {
	if l.empfangefehler != nil || len(l.antworten) == 0 {
		return nil, errors.Join(l.empfangefehler, errors.New("die Leitung ist zu"))
	}

	naechste := l.antworten[0]
	l.antworten = l.antworten[1:]

	return naechste, nil
}

func (l *probeleitung) Schliesse() error {
	l.geschlossen = true

	return nil
}

// annaStelltAus gibt Annas Werk samt Ablage und einen einlösbaren Code.
func annaStelltAus(t *testing.T) (
	anna peer.Identitaet, werk *service.Einladungswerk, ablage *einladungsablage,
	code string, nonce string,
) {
	t.Helper()

	anna = mussIdentitaet(t)
	ablage = neueEinladungsablage()
	werk = service.NeuesEinladungswerk(festerAusweis{ich: anna}, festeRotation{}, ablage,
		func() time.Time { return einladungsuhr }).
		MitGrenze(service.NeueEinloesegrenze(func() time.Time { return einladungsuhr }))

	code = annasCode(t, werk)

	einladung, _, fehler := domain.ZerlegeKontakteinladung(code)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return anna, werk, ablage, code, einladung.Nonce
}

func mussIdentitaet(t *testing.T) peer.Identitaet {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return ich
}

func kodiert(t *testing.T, kodiere func() ([]byte, error)) []byte {
	t.Helper()

	roh, fehler := kodiere()
	if fehler != nil {
		t.Fatal(fehler)
	}

	return roh
}

// **Mit Adresse ist sie einlösbar, und der Einladende merkt sie sich.**
func TestEineAdresseMachtDieEinladungEinloesbar(t *testing.T) {
	t.Parallel()

	_, werk, ablage, code, nonce := annaStelltAus(t)

	if einladung := zerlegt(t, code); !einladung.Einloesbar {
		t.Errorf("mit Adresse nicht einlösbar: %+v", einladung)
	}

	if ablauf := ablage.ausgestellt[nonce]; !ablauf.Equal(einladungsuhr.Add(domain.KontakteinladungGueltig)) {
		t.Errorf("vermerkt: %v", ablage.ausgestellt)
	}

	// **Ohne Adresse gibt es keinen Ort** — und nichts zu merken.
	ohne, fehler := werk.Erstellen(t.Context(), service.Einladungsbitte{Name: "Anna", Zweck: domain.ZweckHandel})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if einladung := zerlegt(t, ohne.Code); einladung.Einloesbar || len(ablage.ausgestellt) != 1 {
		t.Errorf("ohne Adresse: %+v, vermerkt %d", einladung, len(ablage.ausgestellt))
	}
}

// **Was sich nicht vermerken lässt, wird nicht ausgegeben** — eine Einladung,
// die der Einladende nicht kennt, hieße drüben „fremd".
func TestOhneVermerkKeineEinladung(t *testing.T) {
	t.Parallel()

	ablage := neueEinladungsablage()
	ablage.ausstellfehler = errors.New("der Bestand klemmt")

	werk := service.NeuesEinladungswerk(festerAusweis{ich: mussIdentitaet(t)}, festeRotation{}, ablage,
		func() time.Time { return einladungsuhr })

	if ausgestellt, fehler := werk.Erstellen(t.Context(), service.Einladungsbitte{
		Name: "Anna", Zweck: domain.ZweckHandel, Adresse: "192.0.2.1:1",
	}); fehler == nil || ausgestellt.Code != "" {
		t.Errorf("ausgegeben trotz klemmendem Bestand: %+v, %v", ausgestellt, fehler)
	}
}

func TestEinloesenLegtDenGegenueberUnbestaetigtAn(t *testing.T) {
	t.Parallel()

	_, werk, ablage, _, nonce := annaStelltAus(t)
	bert := mussIdentitaet(t)

	if fehler := werk.Einloesen(t.Context(), bert.Kennung(),
		domain.Einloesung{Nonce: nonce, Adresse: "192.0.2.8:41234", Name: "  Bert "}); fehler != nil {
		t.Fatalf("Einloesen: %v", fehler)
	}

	kontakt := ablage.kontakte[string(bert.Kennung())]
	if kontakt.Stand != domain.KontaktUnbestaetigt || kontakt.Fingerabdruck != bert.Fingerabdruck() ||
		kontakt.SelbstgewaehlterName != "Bert" || kontakt.Adresse != "192.0.2.8:41234" ||
		len(kontakt.Rechte) != 0 {
		t.Errorf("Kontakt: %+v", kontakt)
	}
}

// **Die Grenze greift vor dem Bestand** — sonst wäre sie eine Zählung
// dessen, was schon geschehen ist.
func TestDieGrenzeGreiftVorDemBestand(t *testing.T) {
	t.Parallel()

	_, werk, ablage, _, nonce := annaStelltAus(t)
	bert := mussIdentitaet(t)

	for range service.EinloesungenJeAbsender {
		if fehler := werk.Einloesen(t.Context(), bert.Kennung(),
			domain.Einloesung{Nonce: eineFremdeNonce()}); !errors.Is(fehler, domain.ErrEinladungFremd) {
			t.Fatalf("geraten: %v", fehler)
		}
	}

	if fehler := werk.Einloesen(t.Context(), bert.Kennung(),
		domain.Einloesung{Nonce: nonce}); !errors.Is(fehler, domain.ErrEinloesenGedrosselt) {
		t.Errorf("über der Grenze: %v", fehler)
	}

	if len(ablage.kontakte) != 0 {
		t.Errorf("angelegt: %+v", ablage.kontakte)
	}
}

// **Ohne Grenze wird nicht eingelöst** — lieber nein als ungezählt.
func TestOhneGrenzeWirdNichtEingeloest(t *testing.T) {
	t.Parallel()

	werk := service.NeuesEinladungswerk(festerAusweis{ich: mussIdentitaet(t)}, festeRotation{},
		neueEinladungsablage(), func() time.Time { return einladungsuhr })

	if fehler := werk.Einloesen(t.Context(), mussIdentitaet(t).Kennung(),
		domain.Einloesung{Nonce: eineFremdeNonce()}); fehler == nil {
		t.Error("eingelöst ohne Grenze")
	}
}

func eineFremdeNonce() string { return "AAAAAAAAAAAAAAAAAAAAAA" }

// **Ein Name, der hinausgeht, wird vor der Annahme geprüft** — und nicht
// erst drüben still verworfen.
func TestEinUnlesbarerMeldenameNimmtNichtsAn(t *testing.T) {
	t.Parallel()

	_, _, _, code, _ := annaStelltAus(t)
	ablage := neueEinladungsablage()
	werk := service.NeuesEinladungswerk(festerAusweis{ich: mussIdentitaet(t)}, festeRotation{}, ablage,
		func() time.Time { return einladungsuhr })

	if _, fehler := werk.Annehmen(t.Context(), service.Einladungsannahme{
		Code: code, MeinName: "Be\u202ert",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("unlesbar: %v", fehler)
	}

	if len(ablage.kontakte) != 0 {
		t.Errorf("angenommen: %+v", ablage.kontakte)
	}
}

// zerlegt liest einen Code — ein krummer bricht den Lauf ab.
func zerlegt(t *testing.T, code string) domain.Kontakteinladung {
	t.Helper()

	einladung, _, fehler := domain.ZerlegeKontakteinladung(code)
	if fehler != nil {
		t.Fatalf("ZerlegeKontakteinladung: %v", fehler)
	}

	return einladung
}

// klemmendesNachlesen nimmt auf und klemmt beim Lesen danach.
type klemmendesNachlesen struct{ *einladungsablage }

func (k klemmendesNachlesen) Lesen(ctx context.Context, kennung string) (domain.Peerkontakt, error) {
	if len(k.nonces) > 0 {
		return domain.Peerkontakt{}, errors.New("die Platte stottert")
	}

	return k.einladungsablage.Lesen(ctx, kennung)
}

// **Geschrieben ist geschrieben** (§7, Regel 3; Zweitblick über Phase 6.7,
// Z9): klemmt das Nachlesen nach einer geglückten Aufnahme, ist die Annahme
// trotzdem geschehen — und sagt es.
func TestEinKlemmendesNachlesenMachtDieAnnahmeNichtZumFehler(t *testing.T) {
	t.Parallel()

	_, _, _, code, _ := annaStelltAus(t)
	ablage := neueEinladungsablage()
	werk := service.NeuesEinladungswerk(festerAusweis{ich: mussIdentitaet(t)}, festeRotation{},
		klemmendesNachlesen{ablage}, func() time.Time { return einladungsuhr })

	kontakt, fehler := werk.Annehmen(t.Context(), service.Einladungsannahme{Code: code})
	if fehler != nil || kontakt.SelbstgewaehlterName != "Anna" || len(ablage.kontakte) != 1 {
		t.Errorf("Annehmen: %+v, %v", kontakt, fehler)
	}
}
