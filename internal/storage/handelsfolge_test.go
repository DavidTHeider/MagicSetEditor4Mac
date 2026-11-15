package storage_test

// **Die Folge einer Handelsnachricht** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass die Folge ab eins zählt und nie zurückgeht** — auch dann nicht, wenn
// der Ausgang leer ist. Sie ist die Sequenz je Absender und Vorgang; wer sie
// aus dem Ausgang läse, bekäme nach jeder Bestätigung wieder eine Eins, und
// die Gegenseite lieferte auf einen Verlauf nach, den es nicht gibt (`T5`,
// [domain.Nachzuliefern]).
//
// **Und dass zwei Vorgänge zwei Zähler haben.**
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieFolgeZaehltAbEinsUndGehtNieZurueck(t *testing.T) {
	ablage := storage.NeueFolgeablage(frischeDatenbank(t))
	ctx := probekontext(t)

	for erwartet := int64(1); erwartet <= 3; erwartet++ {
		folge, fehler := ablage.NaechsteFolge(ctx, "v-1")
		if fehler != nil {
			t.Fatalf("NaechsteFolge: %v", fehler)
		}

		if folge != erwartet {
			t.Fatalf("es kam %d statt %d", folge, erwartet)
		}
	}

	// **Ein zweiter Vorgang fängt wieder bei eins an** — die Folge gilt je
	// Vorgang, und ein gemeinsamer Zähler ließe die Gegenseite Lücken sehen,
	// wo keine sind.
	folge, fehler := ablage.NaechsteFolge(ctx, "v-2")
	if fehler != nil {
		t.Fatalf("NaechsteFolge für v-2: %v", fehler)
	}

	if folge != 1 {
		t.Errorf("der zweite Vorgang fing bei %d an", folge)
	}
}

// **Ohne Profil im Kontext zählt nichts hoch.**
//
// **Die Gegenprobe zum Lauf darüber**: dieselbe Ablage, derselbe Vorgang, nur
// ohne die Wurzel — und es endet vor dem Bestand.
func TestOhneProfilZaehltDieFolgeNicht(t *testing.T) {
	ablage := storage.NeueFolgeablage(frischeDatenbank(t))

	if _, fehler := ablage.NaechsteFolge(t.Context(), "v-1"); fehler == nil {
		t.Error("ohne Profil kam eine Folge")
	}
}

// **Ein geschlossener Bestand vergibt keine Folge.**
//
// **Die Gegenprobe zum Lauf darüber**: dieselbe Ablage, derselbe Vorgang, nur
// ohne Verbindung — und die Klage nennt, wobei es klemmte, statt eine Null
// zurückzugeben, die wie eine Folge aussähe.
func TestEinGeschlossenerBestandVergibtKeineFolge(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueFolgeablage(db)
	ctx := probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	folge, fehler := ablage.NaechsteFolge(ctx, "v-1")
	if fehler == nil {
		t.Fatalf("es kam die Folge %d", folge)
	}

	if folge != 0 {
		t.Errorf("die Absage gab die Folge %d", folge)
	}
}
