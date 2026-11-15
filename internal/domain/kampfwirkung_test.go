package domain_test

// **Was ein Schlüsselwort am Schaden ändert** — Werkstück M14 von Phase 5.5
// (#1616).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, wer angreifen und blocken darf** — CR 508 und CR 509.
// Hier steht, **was der Schaden tut**: Trampelschaden verteilt ihn anders
// (CR 702.19b), Lebensverknüpfung gibt Leben (CR 702.15b),
// Todesberührung macht ihn tödlich (CR 702.2b) und Unzerstörbarkeit hält
// dagegen (CR 702.12b).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Trampelschaden geht am Blocker vorbei** — CR 702.19b.
//
// **Der Riese ist 3/3, der Bär 2/2**: zwei tödliche auf den Bären, eines
// darüber hinaus auf den Sitz.
func TestTrampelschadenGehtAmBlockerVorbei(t *testing.T) {
	t.Parallel()

	danach := nachDemKampf(t, map[string]string{"grosser": "Trample"})

	if leben := lebenVon(t, danach, "sitz-2"); leben != 19 {
		t.Errorf("sitz-2 hat %d Leben", leben)
	}

	// **Der Bär stirbt trotzdem** — zwei sind seine Widerstandskraft
	// (CR 704.5g). Trampelschaden nimmt ihm nichts weg, er gibt nur den
	// Rest weiter.
	if zone := objektVon(t, danach, "sitz-2/baer/0").Zone; zone !=
		domain.ZoneFriedhof {
		t.Errorf("der Bär liegt in %q", zone)
	}
}

// **Und ohne Trampelschaden bleibt alles beim Blocker** — die Gegenprobe.
func TestOhneTrampelschadenBleibtAllesBeimBlocker(t *testing.T) {
	t.Parallel()

	danach := nachDemKampf(t, nil)

	if leben := lebenVon(t, danach, "sitz-2"); leben != 20 {
		t.Errorf("sitz-2 hat %d Leben", leben)
	}

	if zone := objektVon(t, danach, "sitz-2/baer/0").Zone; zone !=
		domain.ZoneFriedhof {
		t.Errorf("der Bär liegt in %q", zone)
	}
}

// **Trampelschaden ohne Rest gibt nichts weiter** — CR 702.19b.
//
// **Der Bär ist 2/2 und trampelt gegen den Riesen 3/3**: alle zwei sind für
// ihn nötig, und nichts bleibt übrig. Die Gegenprobe zur Rechnung darüber —
// ohne sie sagte der Lauf nur, dass **etwas** durchgeht.
func TestTrampelschadenOhneRestGibtNichtsWeiter(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, map[string]string{"kleiner": "Trample"})

	// **Ein trampelnder Bär bei sitz-1, ein Riese bei sitz-2** — zwei
	// Stärke gegen drei Widerstandskraft.
	partie.Objekte = append(append([]domain.Spielobjekt{}, partie.Objekte...),
		domain.Spielobjekt{
			Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "kleiner",
		},
		domain.Spielobjekt{
			Kennung: "sitz-2/riese/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Karte: "grosser",
		})

	partie, fehler := partie.AngreiferErklaeren("sitz-1", []domain.Angriff{
		{Kreatur: "sitz-1/baer/0", Ziel: "sitz-2"},
	})
	if fehler != nil {
		t.Fatalf("angreifen: %v", fehler)
	}

	partie = bisZumSchritt(t, partie, domain.SchrittBlocker)

	partie, fehler = partie.BlockerErklaeren("sitz-2", []domain.Block{
		{Kreatur: "sitz-2/riese/0", Angreifer: "sitz-1/baer/0"},
	})
	if fehler != nil {
		t.Fatalf("blocken: %v", fehler)
	}

	danach := bisZumSchritt(t, partie, domain.SchrittKampfende)

	if leben := lebenVon(t, danach, "sitz-2"); leben != 20 {
		t.Errorf("sitz-2 hat %d Leben", leben)
	}

	// **Und der Riese hat alle zwei** — nichts ist vorbeigegangen.
	if schaden := objektVon(t, danach, "sitz-2/riese/0").Schaden; schaden != 2 {
		t.Errorf("der Riese hat %d Schaden", schaden)
	}
}

// **Lebensverknüpfung gibt Leben** — CR 702.15b.
func TestLebensverknuepfungGibtLeben(t *testing.T) {
	t.Parallel()

	danach := nachDemKampf(t, map[string]string{"grosser": "Lifelink"})

	if leben := lebenVon(t, danach, "sitz-1"); leben != 23 {
		t.Errorf("sitz-1 hat %d Leben", leben)
	}
}

// **Todesberührung tötet mit jedem Schaden** — CR 702.2b mit CR 704.5h.
//
// **Der Bär ist 2/2 und teilt dem Riesen 2 zu** — zu wenig für seine 3.
// Mit Todesberührung genügt es.
func TestTodesberuehrungToetetMitJedemSchaden(t *testing.T) {
	t.Parallel()

	danach := nachDemKampf(t, map[string]string{"kleiner": "Deathtouch"})

	if zone := objektVon(t, danach, "sitz-1/riese/0").Zone; zone !=
		domain.ZoneFriedhof {
		t.Errorf("der Riese liegt in %q", zone)
	}
}

// **Unzerstörbares stirbt nicht an Schaden** — CR 702.12b mit CR 704.5g.
func TestUnzerstoerbaresStirbtNichtAnSchaden(t *testing.T) {
	t.Parallel()

	// **Der Bär ist 2/2 und bekommt 3 vom Riesen** — genug, und trotzdem
	// bleibt er liegen.
	danach := nachDemKampf(t, map[string]string{"kleiner": "Indestructible"})

	if zone := objektVon(t, danach, "sitz-2/baer/0").Zone; zone !=
		domain.ZoneTisch {
		t.Errorf("der unzerstörbare Bär liegt in %q", zone)
	}

	// **Und auch nicht an Todesberührung** — CR 702.12b nennt beides.
	mitBeidem := nachDemKampf(t, map[string]string{
		"kleiner": "Indestructible", "grosser": "Deathtouch",
	})

	if zone := objektVon(t, mitBeidem, "sitz-2/baer/0").Zone; zone !=
		domain.ZoneTisch {
		t.Errorf("der unzerstörbare Bär liegt in %q", zone)
	}
}

// **Und der Befehl zerstört ihn auch nicht** — CR 702.12b, CR 701.8a.
func TestEinUnzerstoerbaresPermanentWirdNichtZerstoert(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, map[string]string{
		"kleiner": "Indestructible",
	})

	_, fehler := partie.Zerstoeren("sitz-2/baer/0")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es wurde zerstört: %v", fehler)
	}

	if !contains(fehler.Error(), "702.12b") {
		t.Errorf("die Meldung nennt die Regel nicht: %v", fehler)
	}

	// **Geopfert wird es sehr wohl** — CR 701.21a ist kein Zerstören, und
	// CR 702.12b spricht nur vom Zerstören.
	if _, fehler := partie.Opfern("sitz-2", "sitz-2/baer/0"); fehler != nil {
		t.Errorf("opfern ging nicht: %v", fehler)
	}
}

// nachDemKampf lässt den Riesen angreifen, den Bären blocken und den Schaden
// austeilen.
func nachDemKampf(t *testing.T, texte map[string]string) domain.Partie {
	t.Helper()

	partie := imBlockerschritt(t, texte)

	partie, fehler := partie.BlockerErklaeren("sitz-2", []domain.Block{
		{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"},
	})
	if fehler != nil {
		t.Fatalf("blocken: %v", fehler)
	}

	return bisZumSchritt(t, partie, domain.SchrittKampfende)
}

// lebenVon gibt den Lebensstand eines Sitzes.
func lebenVon(t *testing.T, partie domain.Partie, wer string) int {
	t.Helper()

	sitz := partie.SitzVon(wer)
	if sitz == nil {
		t.Fatalf("den Sitz %q gibt es nicht", wer)
	}

	return sitz.Leben
}

// **Eine nicht umgesetzte Fähigkeit ist ein Befund am Tisch** — die dritte
// Abnahme von `M14`
// (#1616).
//
// # Warum das ein eigener Lauf ist
//
// **Weil ein stiller Ausfall wie ein Erfolg aussieht.** Eine Kreatur mit
// Flugfähigkeit **und** Schutzschild liegt auf dem Tisch, das eine gilt und
// das andere nicht — und ohne diesen Satz sähe das niemand.
func TestEineOffeneFaehigkeitIstEinBefundAmTisch(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, map[string]string{
		"grosser": "Flying, hexproof",
		"kleiner": "Shroud",
	})

	befunde := partie.Sicht("sitz-1").Befunde

	if len(befunde) != 2 {
		t.Fatalf("%d Befunde: %v", len(befunde), befunde)
	}

	for _, wort := range []string{"hexproof", "shroud"} {
		gefunden := false

		for _, befund := range befunde {
			if contains(befund, wort) {
				gefunden = true
			}
		}

		if !gefunden {
			t.Errorf("%q steht in keinem Befund: %v", wort, befunde)
		}
	}

	// **Und jeder nennt die Regel** (§11).
	for _, befund := range befunde {
		if !contains(befund, "CR 702") {
			t.Errorf("der Befund nennt die Regel nicht: %q", befund)
		}
	}
}

// **Und was durchgesetzt wird, ist keiner** — die Gegenprobe (§10).
//
// **Ohne sie sagte der Lauf darüber nur, dass *etwas* gemeldet wird** — nicht,
// dass die zwölf schweigen.
func TestEineUmgesetzteFaehigkeitIstKeinBefund(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, map[string]string{
		"grosser": "Flying, trample",
		"kleiner": "Deathtouch (Any amount of damage this deals to a " +
			"creature is enough to destroy it.)",
	})

	if befunde := partie.Sicht("sitz-1").Befunde; len(befunde) != 0 {
		t.Errorf("%d Befunde: %v", len(befunde), befunde)
	}
}

// **Und eine Karte auf der Hand meldet nichts** — statische Fähigkeiten
// gelten auf dem Schlachtfeld (CR 604.3).
func TestEineKarteAufDerHandMeldetNichts(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, map[string]string{"grosser": "Hexproof"})

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Kennung == "sitz-1/riese/0" {
			objekte[stelle].Zone = domain.ZoneHand
		}
	}

	partie.Objekte = objekte

	if befunde := partie.Sicht("sitz-1").Befunde; len(befunde) != 0 {
		t.Errorf("%d Befunde: %v", len(befunde), befunde)
	}
}
