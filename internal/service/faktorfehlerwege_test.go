package service_test

// **Die Fehlerwege des zweiten Faktors** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Sie stehen eigens hier**, weil sie eine eigene Verhaltensklasse sind: was
// geschieht, wenn nicht der Nutzer sich vertippt, sondern etwas darunter
// schiefgeht (§2).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/totp"
)

// **Ein kaputter Bestand reißt an jedem Weg ab.**
//
// **Jeder einzeln, weil jeder seinen eigenen `if fehler != nil` hat** — ein
// vergessener machte aus einem Fehlschlag ein stilles Ja, und das ist bei
// einem zweiten Faktor der schlimmste Ausgang: er ließe jemanden herein.
func TestEinKaputterBestandReisstAnJedemFaktorwegAb(t *testing.T) {
	t.Parallel()

	wege := map[string]func(*service.Zweiterfaktor) error{
		"beim Prüfen": func(z *service.Zweiterfaktor) error {
			return z.Pruefe(t.Context(), 1, "")
		},
		"beim Fragen, was verlangt wird": func(z *service.Zweiterfaktor) error {
			_, fehler := z.Verlangt(t.Context(), 1)

			return fehler
		},
		"beim Lesen der Faktoren": func(z *service.Zweiterfaktor) error {
			_, fehler := z.Faktoren(t.Context(), 1)

			return fehler
		},
		"beim Einrichten": func(z *service.Zweiterfaktor) error {
			_, fehler := z.TOTPEinrichten(t.Context(), 1)

			return fehler
		},
		"beim Bestätigen": func(z *service.Zweiterfaktor) error {
			return z.TOTPBestaetigen(t.Context(), 1, "123456")
		},
		"beim Einrichten der Biometrie": func(z *service.Zweiterfaktor) error {
			return z.BiometrieEinrichten(t.Context(), 1)
		},
		"beim Entfernen": func(z *service.Zweiterfaktor) error {
			return z.Entfernen(t.Context(), 1, domain.FaktorTOTP)
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := weg(zweiterfaktor(&faktordoppel{kaputt: errStummerBestand}))
			if !errors.Is(fehler, errStummerBestand) {
				t.Fatalf("durchgelassen wurde: %v", fehler)
			}
		})
	}
}

// **Ein Lesefehler beim Geheimnis ist nicht „kein Geheimnis".**
//
// # Warum der Unterschied hier zählt
//
// **Weil „kein Geheimnis" der Anlass für den Zettel ist** — genau dann, wenn
// der Schlüsselbund eines anderen Macs es nicht kennt. Ein **Lesefehler**
// dagegen heißt, dass etwas kaputt ist; ihn als „kein Geheimnis" zu lesen
// schickte jemanden zu seinen Wiederherstellungscodes, obwohl seine App
// funktioniert.
func TestEinLesefehlerIstNichtKeinGeheimnis(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	// **Erst Codes anlegen, dann den Bund kaputt machen.** Sonst scheiterte
	// schon das Lesen der Faktoren, und der Weg dahinter bliebe unbetreten.
	if _, fehler := zweiterfaktor(ablage).TOTPEinrichten(t.Context(), 1); fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	ablage.geheimnisfehler = errStummerBestand

	fehler := zweiterfaktor(ablage).Pruefe(t.Context(), 1, "123456")
	if !errors.Is(fehler, errStummerBestand) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein krummes Geheimnis im Schlüsselbund ist ein Fehler und keine
// Abweisung.**
func TestEinKrummesGeheimnisIstEinFehler(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren:  []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
		geheimnis: "kein base32 !!",
	}

	fehler := zweiterfaktor(ablage).TOTPBestaetigen(t.Context(), 1, "123456")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Code, den ein anderer gerade verbraucht hat, gilt nicht.**
//
// **Der Weg, den zwei gleichzeitige Anmeldungen nehmen** (§7, Regel 4): der
// Bestand entscheidet, und der zweite trifft null Zeilen.
func TestEinSchonVerbrauchterCodeGiltNicht(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	ablage.verbrauchfehler = errStummerBestand

	if fehler := dienst.Pruefe(t.Context(), 1, einrichtung.Codes[0]); !errors.Is(
		fehler, errStummerBestand) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Nutzer ohne Faktor verlangt keine Art.**
func TestEinNutzerOhneFaktorVerlangtKeineArt(t *testing.T) {
	t.Parallel()

	art, fehler := zweiterfaktor(&faktordoppel{}).Verlangt(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("fragen: %v", fehler)
	}

	if art != "" {
		t.Errorf("verlangt wird %q", art)
	}
}

// **Und wer beide hat, bekommt die Biometrie angeboten.**
//
// **Weil sie ohne Tippen geht** — und wer sie nicht bedienen kann, kommt
// trotzdem über den Code aus der App oder den Zettel herein.
func TestWerBeideHatBekommtDieBiometrieAngeboten(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{faktoren: []domain.Faktor{
		{NutzerID: 1, Art: domain.FaktorTOTP},
		{NutzerID: 1, Art: domain.FaktorBiometrie},
	}}

	art, fehler := zweiterfaktor(ablage).Verlangt(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("fragen: %v", fehler)
	}

	if art != domain.FaktorBiometrie {
		t.Errorf("verlangt wird %q", art)
	}
}

// **Die Faktoren eines Nutzers kommen an.**
func TestDieFaktorenEinesNutzersKommenAn(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	gelesen, fehler := zweiterfaktor(ablage).Faktoren(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 1 {
		t.Errorf("gelesen wurde %+v", gelesen)
	}
}

// **Ein Sensor, der nein sagt, trägt nichts ein.**
func TestEinSensorDerNeinSagtTraegtNichtsEin(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{}
	sensor := &sensordoppel{vorhanden: true, fehler: errStummerBestand}

	if fehler := service.NeuerZweiterfaktor(ablage, ablage, sensor).
		BiometrieEinrichten(t.Context(), 1); !errors.Is(fehler, errStummerBestand) {
		t.Fatalf("gemeldet wurde: %v", fehler)
	}

	if len(ablage.faktoren) != 0 {
		t.Errorf("trotzdem eingetragen: %+v", ablage.faktoren)
	}
}

// **Ein Nutzer mit TOTP, dessen Geheimnis fehlt, kommt über den Zettel.**
//
// **Das ist der Fall, für den es die Codes gibt**: der Bestand liegt auf einem
// neuen Mac, und dessen Schlüsselbund kennt das Geheimnis nicht.
func TestOhneGeheimnisBleibtDerZettel(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	// **Das Geheimnis ist weg, die Codes sind es nicht.**
	ablage.geheimnis = ""

	if fehler := dienst.Pruefe(t.Context(), 1, einrichtung.Codes[0]); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}
}

// **Und ein Einmalkennwort im Fenster gilt weiter.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie prüfte der Lauf nur, dass
// der Zettel geht, und der Weg über die App bliebe unbelegt.
func TestDerCodeAusDerAppGiltNebenDemZettel(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("einrichten: %v", fehler)
	}

	geheim := geheimnisAus(t, einrichtung.Geheimnis)

	if fehler := dienst.Pruefe(t.Context(), 1,
		totp.Code(totp.Zaehler{Geheimnis: geheim}, probezeit)); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}
}

// **Die Uhr ist eine Naht und keine Konstante.**
//
// **Sonst prüfte jeder Lauf den Augenblick, in dem er läuft** — und ein Code
// aus Anhang B der RFC ginge nie durch.
func TestDieUhrIstEineNaht(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren:  []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
		geheimnis: totp.AlsText([]byte("12345678901234567890")),
	}

	spaeter := probezeit.Add(10 * time.Minute)

	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{}).
		MitUhr(func() time.Time { return spaeter })

	if fehler := dienst.Pruefe(t.Context(), 1, totp.Code(totp.Zaehler{
		Geheimnis: []byte("12345678901234567890"),
	}, spaeter)); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}
}

// **Ein Lesefehler beim Zettel reißt ab.**
func TestEinLesefehlerBeimZettelReisstAb(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren:   []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
		lesefehler: errStummerBestand,
	}

	if fehler := zweiterfaktor(ablage).Pruefe(t.Context(), 1, "abcd-efgh-ijkl"); !errors.Is(
		fehler, errStummerBestand) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Und ein Fehler beim Ablegen der Codes lässt kein halbes Konto zurück.**
//
// **Der Faktor entsteht ohnehin erst nach dem Beweis** — hier wird geprüft,
// dass auch das Einrichten abbricht und nicht mit Codes weitermacht, die
// niemand aufbewahrt hat (§8).
func TestEinFehlerBeimAblegenDerCodesReisstAb(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{setzfehler: errStummerBestand}

	_, fehler := zweiterfaktor(ablage).TOTPEinrichten(t.Context(), 1)
	if !errors.Is(fehler, errStummerBestand) {
		t.Fatalf("gemeldet wurde: %v", fehler)
	}

	if len(ablage.faktoren) != 0 {
		t.Errorf("trotzdem eingetragen: %+v", ablage.faktoren)
	}
}
