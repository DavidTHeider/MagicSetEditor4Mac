package service

// **Ein gemerkter Stil gilt nur, solange der Bestand dasselbe Set beschreibt**
// — #2005.
//
// Bis dahin bauten die Flächen ohne Fenster ihren Stilspeicher je Ruf neu,
// und ein Stilwechsel aus einem anderen Prozess — dem Fenster, der
// Kommandozeile — galt beim nächsten Ruf von selbst. Seit sie ihr Zeichenwerk
// behalten, gilt das nur noch, wenn der Speicher den Kopf des Sets nachsieht.
// Der Wechsel im **selben** Prozess räumt über [Stilspeicher.Vergiss]; diese
// Tests prüfen den Weg ohne ihn.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wandelbarerKopf führt ein Set, dessen Stil und Spiel ein Test umstellt.
type wandelbarerKopf struct {
	sperre      sync.Mutex
	stil, spiel string
}

func (k *wandelbarerKopf) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	k.sperre.Lock()
	defer k.sperre.Unlock()

	return []domain.SetHead{{ID: 7, Style: k.stil, Game: k.spiel}}, nil
}

func (*wandelbarerKopf) Setwerte(context.Context, int64) ([]domain.Value, error) { return nil, nil }

func (k *wandelbarerKopf) setze(stil, spiel string) {
	k.sperre.Lock()
	defer k.sperre.Unlock()

	k.stil, k.spiel = stil, spiel
}

// kopfplatz legt Spiel und zwei Stile der Saat in eine eigene Wurzel.
func kopfplatz(t *testing.T, kopf *wandelbarerKopf) (wurzel string, speicher *Stilspeicher) {
	t.Helper()

	vorher := stempelpause
	stempelpause = 0

	t.Cleanup(func() { stempelpause = vorher })

	wurzel = t.TempDir()
	for _, name := range []string{"probe.mse-game", "probe-schlicht.mse-style", "probe-bild.mse-style"} {
		if err := os.CopyFS(filepath.Join(wurzel, name), os.DirFS(filepath.Join("..", "..", "saat", name))); err != nil {
			t.Fatal(err)
		}
	}

	return wurzel, NeuerStilspeicher(kopf, wurzel)
}

func TestEinStilwechselImBestandGiltOhneVergessen(t *testing.T) {
	kopf := &wandelbarerKopf{stil: "probe-schlicht", spiel: "probe"}
	_, speicher := kopfplatz(t, kopf)

	if vorher, _, _ := speicher.StilFuerSet(t.Context(), 7); vorher.Stil.Paket != "probe-schlicht.mse-style" {
		t.Fatalf("vorher %q", vorher.Stil.Paket)
	}

	kopf.setze("probe-bild", "probe")

	if nachher, _, _ := speicher.StilFuerSet(t.Context(), 7); nachher.Stil.Paket != "probe-bild.mse-style" {
		t.Errorf("nachher %q — der Speicher hielt den Stil, den der Bestand nicht mehr nennt", nachher.Stil.Paket)
	}
}

// **Auch ein gemerkter Fehlschlag**: ein Set ohne Stil, das einen bekommt.
func TestEinSetBekommtSeinenStilOhneNeustart(t *testing.T) {
	kopf := &wandelbarerKopf{spiel: "probe"}
	_, speicher := kopfplatz(t, kopf)

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); da {
		t.Fatal("ein Set ohne Stilangabe bekam einen Stil")
	}

	kopf.setze("probe-schlicht", "probe")

	if _, befunde, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Errorf("kein Stil nach dem Setzen — Befunde %v", befunde)
	}
}

// **Der Stil einer Karte hängt am Spiel des Sets** — nach Namen gemerkt, und
// derselbe Name meint unter einem anderen Spiel ein anderes Paket.
func TestEinSpielwechselGiltFuerDenKartenstil(t *testing.T) {
	kopf := &wandelbarerKopf{stil: "probe-schlicht", spiel: "probe"}
	_, speicher := kopfplatz(t, kopf)

	karte := &domain.CardRow{SetID: 7, Values: []domain.Value{{Field: Kartenstilfeld, Text: "probe-schlicht"}}}

	vorher, _, da := speicher.StilFuerKarte(t.Context(), karte)
	if !da || vorher.Spielskript == "" {
		t.Fatalf("vorher da=%v, Spielskript %q", da, vorher.Spielskript)
	}

	kopf.setze("probe-schlicht", "anderes")

	if nachher, _, _ := speicher.StilFuerKarte(t.Context(), karte); nachher.Spielskript != "" {
		t.Error("das Skript des alten Spiels kam mit — gemerkt war nach dem Namen allein")
	}
}

// **Ein geändertes Paket meldet sich** — einmal, und nicht beim ersten Lesen.
//
// Die Bildquelle und der Ebenenspeicher daneben halten **entschlüsselte**
// Bilder; ohne die Meldung zeigte eine Karte den neuen Stil mit dem alten
// Rahmen.
func TestEinGeaendertesPaketMeldetSich(t *testing.T) {
	kopf := &wandelbarerKopf{stil: "probe-bild", spiel: "probe"}
	wurzel, speicher := kopfplatz(t, kopf)

	gemeldet := 0
	speicher.MitPaketwechsel(func() { gemeldet++ })

	for range 2 {
		if _, _, da := speicher.StilFuerSet(t.Context(), 7); !da {
			t.Fatal("der Stil fehlt")
		}
	}

	if gemeldet != 0 {
		t.Fatalf("%d Meldungen ohne Änderung", gemeldet)
	}

	rahmen := filepath.Join(wurzel, "probe-bild.mse-style", "rahmen.png")
	if err := os.WriteFile(rahmen, []byte("kein Bild mehr"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Fatal("der Stil fehlt nach der Änderung")
	}

	if gemeldet != 1 {
		t.Errorf("%d Meldungen nach einer Änderung, erwartet eine", gemeldet)
	}

	// **Derselbe Weg für den Stil einer Karte.**
	karte := &domain.CardRow{SetID: 7, Values: []domain.Value{{Field: Kartenstilfeld, Text: "probe-bild"}}}
	if _, _, da := speicher.StilFuerKarte(t.Context(), karte); !da {
		t.Fatal("der Kartenstil fehlt")
	}

	if err := os.WriteFile(rahmen, []byte("wieder ein anderes"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, da := speicher.StilFuerKarte(t.Context(), karte); !da {
		t.Fatal("der Kartenstil fehlt nach der Änderung")
	}

	if gemeldet != 2 {
		t.Errorf("%d Meldungen nach der Änderung am Kartenstil, erwartet zwei", gemeldet)
	}
}

// **Ein Set, das der Bestand nicht führt, wird nicht gemerkt** — auch nicht
// für die Dauer einer Pause.
//
// Ein Zeichenwerk, das ein Prozess behält, sieht Rufe aus mehreren Profilen:
// `mse-rest` bedient jeden Principal über denselben Bestand. Ein Ruf aus einem
// Profil, das das Set nicht führt, merkte sich „gibt es nicht" — und der
// nächste Ruf aus dem richtigen Profil bekäme den Behelfsplan.
func TestEinFehlendesSetWirdNichtGemerkt(t *testing.T) {
	kopf := &fehlenderKopf{}
	_, speicher := kopfplatz(t, &wandelbarerKopf{})
	speicher.koepfe = kopf

	stempelpause = time.Hour

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); da {
		t.Fatal("ein Set, das der Bestand nicht führt, bekam einen Stil")
	}

	kopf.da = true

	if _, befunde, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Errorf("das Set fehlt weiter — gemerkt war das Fehlen: %v", befunde)
	}
}

// fehlenderKopf führt das Set 7 erst, wenn `da` gesetzt ist.
type fehlenderKopf struct{ da bool }

func (k *fehlenderKopf) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	if !k.da {
		return nil, nil
	}

	return []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}}, nil
}

func (*fehlenderKopf) Setwerte(context.Context, int64) ([]domain.Value, error) { return nil, nil }

// zaehlenderKopf zählt, wie oft nach den Köpfen gefragt wird.
type zaehlenderKopf struct {
	wandelbarerKopf
	gefragt int
}

func (k *zaehlenderKopf) Setkoepfe(ctx context.Context) ([]domain.SetHead, error) {
	k.gefragt++

	return k.wandelbarerKopf.Setkoepfe(ctx)
}

// **Der Kopf je Ruf, das Paket einmal** — das Paket verschwindet nach dem
// ersten Ruf, und der Stil kommt trotzdem: aus dem Gedächtnis.
func TestDerKopfJeRufDasPaketEinmal(t *testing.T) {
	kopf := &zaehlenderKopf{wandelbarerKopf: wandelbarerKopf{stil: "probe-schlicht", spiel: "probe"}}
	wurzel, speicher := kopfplatz(t, &kopf.wandelbarerKopf)
	speicher.koepfe = kopf

	stempelpause = time.Hour

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Fatal("der Stil fehlt")
	}

	if err := os.RemoveAll(filepath.Join(wurzel, "probe-schlicht.mse-style")); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if _, befunde, da := speicher.StilFuerSet(t.Context(), 7); !da {
			t.Fatalf("der Stil wurde neu gesucht: %v", befunde)
		}
	}

	if kopf.gefragt != 3 {
		t.Errorf("%d Abfragen der Köpfe, erwartet je Ruf eine", kopf.gefragt)
	}
}

// **Und ein Fehlschlag ebenso**: das Paket entsteht nach dem ersten Ruf, und
// solange die Pause gilt, bleibt es beim gemerkten Fehlen.
func TestAuchEinFehlschlagBleibtOhnePaketGemerkt(t *testing.T) {
	kopf := &wandelbarerKopf{stil: "gibtesnicht", spiel: "probe"}
	wurzel, speicher := kopfplatz(t, kopf)

	stempelpause = time.Hour

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); da {
		t.Fatal("ein Stil, den es nicht gibt")
	}

	if err := os.CopyFS(filepath.Join(wurzel, "gibtesnicht.mse-style"),
		os.DirFS(filepath.Join(wurzel, "probe-schlicht.mse-style"))); err != nil {
		t.Fatal(err)
	}

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); da {
		t.Error("der Fehlschlag wurde nicht gemerkt — das Paket wurde neu gesucht")
	}
}

// stolpernderKopf scheitert beim Lesen, sobald `fehler` gesetzt ist.
type stolpernderKopf struct {
	wandelbarerKopf
	fehler error
}

func (k *stolpernderKopf) Setkoepfe(ctx context.Context) ([]domain.SetHead, error) {
	if k.fehler != nil {
		return nil, k.fehler
	}

	return k.wandelbarerKopf.Setkoepfe(ctx)
}

// **Ein Lesefehler lässt das Gemerkte gelten, solange sein Stempel gilt** —
// er sagt nichts über den Stil. Ist auch der Stempel verfallen, bleibt nur der
// Befund: was neu zu lesen wäre, hinge an einem Kopf, den es gerade nicht gibt.
func TestEinLesefehlerLaesstDasGemerkteGelten(t *testing.T) {
	kopf := &stolpernderKopf{wandelbarerKopf: wandelbarerKopf{stil: "probe-schlicht", spiel: "probe"}}
	wurzel, speicher := kopfplatz(t, &kopf.wandelbarerKopf)
	speicher.koepfe = kopf

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Fatal("der Stil fehlt")
	}

	kopf.fehler = errors.New("Bestand weg")

	if _, befunde, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Fatalf("ein Lesefehler warf das Gemerkte weg: %v", befunde)
	}

	zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"), "card width: 375", "card width: 200")

	_, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if da || len(befunde) == 0 || !strings.Contains(befunde[0].Meldung, "Bestand weg") {
		t.Errorf("da=%v Befunde=%v — erwartet der Lesefehler", da, befunde)
	}
}
