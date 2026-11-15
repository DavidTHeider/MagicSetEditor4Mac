// **Der Vorlagenzeichner über einem echten Bestand** — Werkstück 3 von
// Phase 3.3 (#1164).
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Probekarte kommt aus dem Set, das diesen Stil benutzt.**
func TestDieProbekarteKommtAusDemPassendenSet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	// **Zwei Sets, und nur eines nennt den Stil.** Ohne das zweite sagte
	// dieser Lauf auch dann ja, wenn die Naht das erstbeste Set nähme.
	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Fremd", Game: "magic", Style: "anderer"}, nil); err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Passend", Game: "magic", Style: "m15"}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	karteID, kartenfehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID)
	if kartenfehler != nil {
		t.Fatalf("Karte anlegen: %v", kartenfehler)
	}

	naht, nahtfehler := app.VorlagennahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	karte, err := naht.Probekarte(probekontext(t), "/pakete/magic-m15.mse-style", 0)
	if err != nil {
		t.Fatalf("Probekarte: %v", err)
	}

	if karte.ID != karteID {
		t.Errorf("die Karte ist %d — erwartet war %d", karte.ID, karteID)
	}
}

// **Ohne passendes Set gibt es keine Probekarte — und das ist „nicht
// gefunden", kein Fehlschlag.**
//
// Das Fenster zeigt daraufhin, was zu tun ist: ein Set öffnen, das diese
// Vorlage nennt (§6).
func TestOhnePassendesSetGibtEsKeineProbekarte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht, err := app.VorlagennahtUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("Naht: %v", err)
	}

	if _, fehler := naht.Probekarte(
		probekontext(t), "/pakete/magic-m15.mse-style", 0); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", fehler)
	}
}

// **Ein Set ohne Karten ist keine Quelle für eine Probekarte.**
//
// Es wird übersprungen, nicht abgebrochen: ein leeres Set und ein volles mit
// demselben Stil kommen in jeder gewachsenen Sammlung vor.
func TestEinLeeresSetWirdUebersprungen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Leer", Game: "magic", Style: "m15"}, nil); err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	naht, nahtfehler := app.VorlagennahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	if _, fehler := naht.Probekarte(
		probekontext(t), "/pakete/magic-m15.mse-style", 0); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", fehler)
	}
}

// **Beide Schreibweisen eines Stils zählen.**
//
// Ein Set schreibt `stylesheet: m15` oder `stylesheet: magic-m15`; welche,
// hängt daran, wer es geschrieben hat. `fenhl/magic-set-generator` schreibt
// die kurze (§11), MSE2 selbst kennt beide.
func TestBeideSchreibweisenEinesStilsZaehlen(t *testing.T) {
	for name, stil := range map[string]string{"kurz": "m15", "lang": "magic-m15"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())

			db := frischerBestand(t)

			setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
				&domain.Set{Name: "Probe", Game: "magic", Style: stil}, nil)
			if err != nil {
				t.Fatalf("Set anlegen: %v", err)
			}

			if _, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(
				probekontext(t), setID); fehler != nil {
				t.Fatalf("Karte anlegen: %v", fehler)
			}

			naht, nahtfehler := app.VorlagennahtUeber(db)
			if nahtfehler != nil {
				t.Fatalf("Naht: %v", nahtfehler)
			}

			if _, fehler := naht.Probekarte(
				probekontext(t), "/pakete/magic-m15.mse-style", 0); fehler != nil {
				t.Errorf("Probekarte: %v", fehler)
			}
		})
	}
}

// **Der Zeichner mit Ersatz gibt einen Planbauer und den Rasterer.**
func TestDerZeichnerMitErsatzGibtBeides(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht, err := app.VorlagennahtUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("Naht: %v", err)
	}

	planbauer, rasterer := naht.MitErsatz(service.Paketersatz{
		Paket: "/pakete/magic-m15.mse-style", Datei: "style", Inhalt: []byte("x"),
	})

	if planbauer == nil || rasterer == nil {
		t.Errorf("Planbauer %v, Rasterer %v", planbauer, rasterer)
	}
}

// **Ohne Zuhause kommt die Naht nicht zusammen.**
func TestOhneZuhauseKommtDieVorlagennahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := app.VorlagennahtUeber(frischerBestand(t)); err == nil {
		t.Fatal("kein Fehler ohne Zuhause")
	}
}

// **Ein Stilname ohne Bindestrich bleibt, wie er ist.**
//
// Kein bezogenes Paket heißt so — sie tragen alle `<spiel>-<stil>` —, aber ein
// selbst angelegtes kann es. Der Zweig steht deshalb da und wird hier betreten
// (§3).
func TestEinStilnameOhneBindestrichBleibt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "probe", Style: "schlicht"}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	if _, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(
		probekontext(t), setID); fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	naht, nahtfehler := app.VorlagennahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	if _, fehler := naht.Probekarte(probekontext(t), "/pakete/schlicht.mse-style", 0); fehler != nil {
		t.Errorf("Probekarte: %v", fehler)
	}
}

// **Ein Bestand, der nicht mehr aufgeht, meldet sich — an beiden Stellen.**
//
// Die Naht liest zweimal: die Setköpfe und danach die Karten. Beide Wege
// gehören betreten; ein geschlossener Bestand ist der Fall, in dem eine
// Sicherung gerade zurückgespielt wird.
func TestEinGeschlossenerBestandMeldetSichBeimZeichner(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "magic", Style: "m15"}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	if _, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(
		probekontext(t), setID); fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	naht, nahtfehler := app.VorlagennahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := naht.Probekarte(probekontext(t), "/pakete/magic-m15.mse-style", 0); fehler == nil {
		t.Error("der geschlossene Bestand kam durch")
	}
}

// leserohneKarten gibt einen Setkopf und scheitert beim Lesen seiner Karten.
type leserohneKarten struct{ fehler error }

func (l leserohneKarten) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 1, Game: "magic", Style: "m15"}}, nil
}

func (l leserohneKarten) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, l.fehler
}

func (l leserohneKarten) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return nil, l.fehler
}

// **Ein Fehler beim Lesen der Karten kommt durch — nicht als „nicht gefunden".**
//
// # Warum das ein eigener Weg ist
//
// Die Naht liest zweimal: die Setköpfe und danach die Karten des passenden
// Sets. Ein geschlossener Bestand trifft **beide** auf einmal, und der zweite
// bliebe damit unbetreten. Der Unterschied ist keine Förmlichkeit: „kein Set
// benutzt diese Vorlage" schickt jemanden Sets öffnen, „der Bestand ließ sich
// nicht lesen" nicht.
func TestEinLesefehlerBeiDenKartenKommtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht, err := app.VorlagennahtUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("Naht: %v", err)
	}

	kaputt := errors.New("der Bestand ist weg")

	_, fehler := naht.MitLeser(leserohneKarten{fehler: kaputt}).
		Probekarte(probekontext(t), "/pakete/magic-m15.mse-style", 0)

	if !errors.Is(fehler, kaputt) {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Und ein Set mit einem anderen Stil wird übersprungen.**
//
// Die Gegenprobe zur Suche: ohne sie sagte sie auch dann ja, wenn die Naht das
// erstbeste Set nähme.
func TestEinSetMitAnderemStilWirdUebersprungen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht, err := app.VorlagennahtUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("Naht: %v", err)
	}

	if _, fehler := naht.MitLeser(nurFremdeStile{}).
		Probekarte(probekontext(t), "/pakete/magic-m15.mse-style", 0); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", fehler)
	}
}

// nurFremdeStile führt ein Set, das einen anderen Stil nennt.
type nurFremdeStile struct{}

func (nurFremdeStile) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 1, Game: "magic", Style: "anderer"}}, nil
}

func (nurFremdeStile) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, errors.New("hierher kommt niemand")
}

func (nurFremdeStile) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return nil, errors.New("hierher kommt niemand")
}

// **Ein Stil, dessen Name auf den gesuchten endet, ist ein anderer Stil.**
//
// # Warum das ein eigener Lauf ist
//
// `counter-m15` endet auf `-m15`, und ein Vergleich über die Endung nahm es
// für `m15`. Die Vorschau zeigte dann eine Karte des **fremden** Stils, und
// der Ersatz greift auf ihr nicht: er zielt auf den bearbeiteten Paketpfad
// und wird exakt verglichen. Der Weg lief vollständig durch, das Bild trug
// die Änderung nicht — eine Vorschau, die nichts vorschaut.
//
// Gefunden vom Zweitblick durch Codex über Phase 3.3. **Gemessen**: unter den
// 376 Stilpaketen des Packs gibt es 26 solcher Paare, darunter
// `m15 ← counter-m15`, `mini ← booster-mini` und `duplex ← m15-duplex`.
//
// **Das fremde Set wird als letztes angelegt**, denn `Setkoepfe` gibt zuletzt
// Geöffnetes zuerst. Andersherum fände die Schleife das richtige Set zuerst,
// und der Lauf wäre auch mit dem Fehler grün.
func TestEinStilDerAufDenGesuchtenEndetIstEinAnderer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	kartenID := map[string]int64{}

	for _, stil := range []string{"m15", "counter-m15"} {
		setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
			&domain.Set{Name: stil, Game: "magic", Style: stil}, nil)
		if err != nil {
			t.Fatalf("Set anlegen: %v", err)
		}

		karte, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(
			probekontext(t), setID)
		if fehler != nil {
			t.Fatalf("Karte anlegen: %v", fehler)
		}

		kartenID[stil] = karte
	}

	naht, nahtfehler := app.VorlagennahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	karte, err := naht.Probekarte(probekontext(t), "/pakete/magic-m15.mse-style", 0)
	if err != nil {
		t.Fatalf("Probekarte: %v", err)
	}

	if karte.ID != kartenID["m15"] {
		t.Errorf("die Probekarte ist %d — sie kommt aus dem Set mit dem Stil"+
			" counter-m15 statt aus dem mit m15", karte.ID)
	}
}
