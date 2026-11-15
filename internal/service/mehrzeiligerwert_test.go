package service_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein mehrzeiliger Feldwert übersteht das Speichern** —
// #704.
//
// # Der Befund
//
// Gefunden beim Rundlauf über das mitgelieferte `Vanguard`: von 18 Feldwerten
// der ersten Karte kamen nach „sichern, wieder öffnen" **zwei** zurück, und der
// mehrzeilige stand danach **doppelt** da.
//
// Zwei Hälften, und beide zählen:
//
//  1. `msepaket.Schreibe` setzte `key: <wert mit Umbrüchen>` in **eine** Zeile.
//     Die zweite Zeile stand ohne Einrückung und beendete den `card:`-Block.
//  2. `werteSetzen` verglich den Wert des Bestands mit `kind.Wert` — und der
//     ist bei einem mehrzeiligen Wert **leer**, denn der Wert steht in den
//     Kindern. Jeder Lauf hielt ihn deshalb für geändert, schrieb die Zeile neu
//     und ließ die alten Kinder stehen: aus einem Wert wurden zwei.
type mehrzeilenleser struct {
	leserDouble

	ort    string
	quelle string

	// notiz ist der Wert, den der Bestand für `notes` führt — leer heißt: der
	// aus der Datei.
	notiz string
}

func (m *mehrzeilenleser) Herkunft(context.Context, int64) (ort, datei string, fehler error) {
	return m.ort, "set", nil
}

func (m *mehrzeilenleser) Quelle(context.Context, int64) (*domain.Quelle, error) {
	return &domain.Quelle{Datei: "set", Inhalt: []byte(m.quelle), Herkunft: m.ort}, nil
}

func (m *mehrzeilenleser) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, nil
}

func (m *mehrzeilenleser) Setwerte(context.Context, int64) ([]domain.Value, error) {
	return nil, nil
}

func (m *mehrzeilenleser) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return []*domain.CardRow{{
		ID: 1, Number: 0, Values: []domain.Value{
			{Field: "name", Text: "Feuerball"},
			{Field: "notes", Text: m.notizOder("Erste Zeile\n\nDritte Zeile")},
		},
	}}, nil
}

func (m *mehrzeilenleser) notizOder(vorgabe string) string {
	if m.notiz != "" {
		return m.notiz
	}

	return vorgabe
}

func TestEinMehrzeiligerWertUeberstehtDasSpeichern(t *testing.T) {
	t.Parallel()

	herkunft := filepath.Join(t.TempDir(), "Probe.mse-set")
	if err := os.MkdirAll(herkunft, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	quelle := "mse version: 2.0.0\ncard:\n\tname: Feuerball\n" +
		"\tnotes:\n\t\tErste Zeile\n\t\t\n\t\tDritte Zeile\n"

	schreibe(t, filepath.Join(herkunft, "set"), quelle)

	ziel := filepath.Join(t.TempDir(), "Ziel.mse-set")

	schreiber := service.NeuerSetschreiber(
		&mehrzeilenleser{ort: herkunft, quelle: quelle}, &ortsschreiber{})
	if _, err := schreiber.Schreibe(t.Context(), 1, ziel); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	roh, err := os.ReadFile(filepath.Join(ziel, "set")) //nolint:gosec // eigenes Temp
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Unverändert heißt unverändert.** Der Bestand führt denselben Wert;
	// eine Datei, die sich beim Speichern ändert, ohne dass jemand etwas
	// geändert hat, bricht die Byte-Zusage (#48).
	if string(roh) != quelle {
		t.Errorf("geschrieben\n%q\nerwartet\n%q", string(roh), quelle)
	}

	if zahl := strings.Count(string(roh), "Erste Zeile"); zahl != 1 {
		t.Errorf("»Erste Zeile« steht %d-mal in der Datei", zahl)
	}
}

// **Und ein geänderter mehrzeiliger Wert ersetzt den alten, statt sich
// danebenzustellen.**
//
// Der Block unter der Feldzeile **war** der alte Wert; ihn stehen zu lassen
// hieße, beide zu behalten — und beim nächsten Öffnen stünde der Wert doppelt
// da (#704).
func TestEinGeaenderterMehrzeiligerWertErsetztDenAlten(t *testing.T) {
	t.Parallel()

	herkunft := filepath.Join(t.TempDir(), "Probe.mse-set")
	if err := os.MkdirAll(herkunft, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	quelle := "mse version: 2.0.0\ncard:\n\tname: Feuerball\n" +
		"\tnotes:\n\t\tAlte Zeile\n\t\t\n\t\tAlter Rest\n"

	schreibe(t, filepath.Join(herkunft, "set"), quelle)

	ziel := filepath.Join(t.TempDir(), "Ziel.mse-set")

	leser := &mehrzeilenleser{ort: herkunft, quelle: quelle}
	leser.notiz = "Neue Zeile\n\nNeuer Rest"

	schreiber := service.NeuerSetschreiber(leser, &ortsschreiber{})
	if _, err := schreiber.Schreibe(t.Context(), 1, ziel); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	roh, err := os.ReadFile(filepath.Join(ziel, "set")) //nolint:gosec // eigenes Temp
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	erwartet := "mse version: 2.0.0\ncard:\n\tname: Feuerball\n" +
		"\tnotes:\n\t\tNeue Zeile\n\t\t\n\t\tNeuer Rest\n"

	if string(roh) != erwartet {
		t.Errorf("geschrieben\n%q\nerwartet\n%q", string(roh), erwartet)
	}

	if strings.Contains(string(roh), "Alte Zeile") {
		t.Error("der alte Wert steht noch in der Datei")
	}
}
