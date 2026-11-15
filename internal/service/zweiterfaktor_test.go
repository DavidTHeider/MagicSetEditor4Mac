package service_test

// **Der zweite Faktor** — Werkstück 7 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/totp"
)

// **Ein TOTP-Faktor gilt erst, wenn ein Code aus ihm stimmt.**
//
// # Warum zwei Schritte und nicht einer
//
// **Weil sonst jemand sich selbst aussperrt.** Wer den Faktor beim Anzeigen
// des Musters einträgt und den Scan vergisst, verlangt danach einen Code, den
// niemand erzeugen kann. Der Faktor entsteht deshalb **nach** dem Beweis.
func TestEinTOTPFaktorGiltErstNachDemBeweis(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{}
	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	// **Noch kein Faktor** — nur ein Geheimnis, das darauf wartet.
	if len(ablage.faktoren) != 0 {
		t.Fatalf("schon eingetragen: %+v", ablage.faktoren)
	}

	if !strings.HasPrefix(einrichtung.Adresse, "otpauth://totp/") {
		t.Errorf("die Adresse ist %q", einrichtung.Adresse)
	}

	if len(einrichtung.Codes) != domain.Wiederherstellungscodes {
		t.Errorf("es kamen %d Codes", len(einrichtung.Codes))
	}

	geheim := geheimnisAus(t, einrichtung.Geheimnis)

	if fehler := dienst.TOTPBestaetigen(t.Context(), 1,
		totp.Code(totp.Zaehler{Geheimnis: geheim}, probezeit)); fehler != nil {
		t.Fatalf("bestätigen: %v", fehler)
	}

	if len(ablage.faktoren) != 1 || ablage.faktoren[0].Art != domain.FaktorTOTP {
		t.Errorf("eingetragen wurde %+v", ablage.faktoren)
	}
}

// **Ein falscher Code trägt nichts ein.**
func TestEinFalscherCodeTraegtNichtsEin(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{}
	dienst := zweiterfaktor(ablage)

	if _, fehler := dienst.TOTPEinrichten(t.Context(), 1); fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	if fehler := dienst.TOTPBestaetigen(t.Context(), 1, "000000"); fehler == nil {
		t.Fatal("angenommen")
	}

	if len(ablage.faktoren) != 0 {
		t.Errorf("trotzdem eingetragen: %+v", ablage.faktoren)
	}
}

// **Ohne Geheimnis lässt sich nichts bestätigen.**
//
// **Der Weg, den niemand geht und der trotzdem zu bedenken ist**: wer die
// Anwendung zwischen den beiden Schritten schließt, kommt hier an.
func TestOhneGeheimnisLaesstSichNichtsBestaetigen(t *testing.T) {
	t.Parallel()

	fehler := zweiterfaktor(&faktordoppel{}).TOTPBestaetigen(t.Context(), 1, "123456")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Nutzer ohne Faktor verlangt keinen.**
//
// **Das ist die Abnahme aus dem Issue**: „ein bestehender Aufruf ohne
// Anmeldung bricht nicht."
func TestEinNutzerOhneFaktorVerlangtKeinen(t *testing.T) {
	t.Parallel()

	if fehler := zweiterfaktor(&faktordoppel{}).Pruefe(t.Context(), 1, ""); fehler != nil {
		t.Errorf("verlangt wurde: %v", fehler)
	}
}

// **Ein Nutzer mit TOTP verlangt einen Code — und sagt, welche Art.**
func TestEinNutzerMitTOTPVerlangtEinenCode(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	fehler := zweiterfaktor(ablage).Pruefe(t.Context(), 1, "")
	if !errors.Is(fehler, domain.ErrZweiterFaktor) {
		t.Fatalf("durchgelassen wurde: %v", fehler)
	}

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Werte["art"] != string(domain.FaktorTOTP) {
		t.Errorf("gemeldet wurde %+v", fehler)
	}
}

// **Ein Code aus der App kommt durch, ein anderer nicht.**
func TestEinCodeAusDerAppKommtDurch(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren:  []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
		geheimnis: totp.AlsText([]byte("12345678901234567890")),
	}

	dienst := zweiterfaktor(ablage)
	richtig := totp.Code(totp.Zaehler{
		Geheimnis: []byte("12345678901234567890"),
	}, probezeit)

	if fehler := dienst.Pruefe(t.Context(), 1, richtig); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}

	if fehler := dienst.Pruefe(t.Context(), 1, "000000"); fehler == nil {
		t.Error("angenommen")
	}
}

// **Ein Wiederherstellungscode kommt durch — und nur einmal.**
//
// # Wozu er da ist
//
// **Weil das Geheimnis im Schlüsselbund liegt.** Wer seinen Bestand auf einem
// neuen Mac zurückspielt, hat den Schlüsselbund des alten nicht — und käme
// ohne diesen Weg nicht mehr an sein eigenes Profil.
func TestEinWiederherstellungscodeKommtEinmalDurch(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	erster := einrichtung.Codes[0]

	if fehler := dienst.Pruefe(t.Context(), 1, erster); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}

	if fehler := dienst.Pruefe(t.Context(), 1, erster); fehler == nil {
		t.Error("zweimal angenommen")
	}
}

// **Biometrie fragt den Sensor und nicht den Aufrufer.**
//
// # Warum der Dienst ihn selbst fragt
//
// **Weil eine Fläche, die „ja, war Touch ID" sagt, nichts nachweist.** Der
// Sensor hängt an einer Naht, und der Composition Root setzt sie — dieselbe
// Richtung wie bei jeder anderen Plattformfrage (§2).
func TestBiometrieFragtDenSensor(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorBiometrie}},
	}

	sensor := &sensordoppel{vorhanden: true}
	dienst := service.NeuerZweiterfaktor(ablage, ablage, sensor)

	if fehler := dienst.Pruefe(t.Context(), 1, ""); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}

	if sensor.gefragt == "" {
		t.Error("der Sensor wurde nicht gefragt")
	}

	sensor.fehler = errors.New("der Finger passte nicht")

	if fehler := dienst.Pruefe(t.Context(), 1, ""); fehler == nil {
		t.Error("angenommen, obwohl der Sensor nein sagt")
	}
}

// **Fehlt der Sensor, bleibt der Code aus der App.**
//
// **§16, Regel 5, sinngemäß**: fällt ein Weg aus, fällt ein Weg aus und nicht
// die Anmeldung. Wer Touch ID eingerichtet hat und den Bestand auf einem Mac
// ohne Sensor öffnet, kommt über seinen zweiten Faktor herein — oder über
// einen Wiederherstellungscode.
func TestFehltDerSensorBleibtDerCode(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren:  []domain.Faktor{{NutzerID: 1, Art: domain.FaktorBiometrie}},
		geheimnis: totp.AlsText([]byte("12345678901234567890")),
	}

	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{})

	fehler := dienst.Pruefe(t.Context(), 1, "")
	if !errors.Is(fehler, domain.ErrZweiterFaktor) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Faktor lässt sich entfernen — und nimmt sein Geheimnis mit.**
func TestEinFaktorNimmtSeinGeheimnisMit(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren:  []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
		geheimnis: "GEZDGNBVGY3TQOJQ",
	}

	if fehler := zweiterfaktor(ablage).Entfernen(
		t.Context(), 1, domain.FaktorTOTP); fehler != nil {
		t.Fatalf("entfernen: %v", fehler)
	}

	if len(ablage.faktoren) != 0 {
		t.Errorf("noch da: %+v", ablage.faktoren)
	}

	if ablage.geheimnis != "" {
		t.Errorf("das Geheimnis blieb: %q", ablage.geheimnis)
	}
}

// **Und Biometrie hat keines, das mitginge.**
func TestBiometrieHatKeinGeheimnis(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren:  []domain.Faktor{{NutzerID: 1, Art: domain.FaktorBiometrie}},
		geheimnis: "GEZDGNBVGY3TQOJQ",
	}

	if fehler := zweiterfaktor(ablage).Entfernen(
		t.Context(), 1, domain.FaktorBiometrie); fehler != nil {
		t.Fatalf("entfernen: %v", fehler)
	}

	// **Das TOTP-Geheimnis bleibt**, denn es gehört dem anderen Faktor.
	if ablage.geheimnis == "" {
		t.Error("das Geheimnis des anderen Faktors ging mit")
	}
}

// **Biometrie lässt sich nur einrichten, wo es sie gibt.**
func TestBiometrieLaesstSichNurEinrichtenWoEsSieGibt(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{}

	fehler := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{}).
		BiometrieEinrichten(t.Context(), 1)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}

	if len(ablage.faktoren) != 0 {
		t.Errorf("trotzdem eingetragen: %+v", ablage.faktoren)
	}
}

// **Und wo es sie gibt, verlangt sie einmal den Finger.**
//
// **Sonst trüge jemand einen Faktor ein, den er nicht bedienen kann** —
// dieselbe Überlegung wie beim zweistufigen TOTP.
func TestBiometrieVerlangtEinmalDenFinger(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{}
	sensor := &sensordoppel{vorhanden: true}

	if fehler := service.NeuerZweiterfaktor(ablage, ablage, sensor).
		BiometrieEinrichten(t.Context(), 1); fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	if sensor.gefragt == "" {
		t.Error("der Sensor wurde nicht gefragt")
	}

	if len(ablage.faktoren) != 1 {
		t.Errorf("eingetragen wurde %+v", ablage.faktoren)
	}
}

// probezeit ist der Zeitpunkt, an dem diese Läufe rechnen.
//
// **Ein fester und nicht `time.Now`**: ein Code, der zur Laufzeit entsteht und
// zur Laufzeit geprüft wird, ginge auch dann durch, wenn beide Seiten dieselbe
// falsche Rechnung machten. Der Wert ist einer aus Anhang B der RFC.
var probezeit = time.Unix(1111111111, 0)

// zweiterfaktor baut den Dienst über einem Doppel — mit vorhandenem Sensor
// und stehender Uhr.
func zweiterfaktor(ablage *faktordoppel) *service.Zweiterfaktor {
	return service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{vorhanden: true}).
		MitUhr(func() time.Time { return probezeit })
}

// geheimnisAus liest ein base32-Geheimnis und bricht ab, wenn es krumm ist.
func geheimnisAus(t *testing.T, text string) []byte {
	t.Helper()

	geheim, fehler := totp.GeheimnisAus(text)
	if fehler != nil {
		t.Fatalf("das Geheimnis lesen: %v", fehler)
	}

	return geheim
}
