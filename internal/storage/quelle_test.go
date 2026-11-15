package storage_test

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Quelldatei kommt Byte für Byte zurück.**
//
// # Woher diese Zusage kommt
//
// #48: ein
// unverändert durchgereichtes Set ist nach dem Speichern dasselbe. Aus
// `card_values` allein ist das nicht zu halten — Einrückung, Abstände,
// Kommentare und Leerzeilen stehen dort nicht
// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
//
// **Als BLOB und nicht als TEXT**: eine `.mse-set` kann in einer Kodierung
// vorliegen, die niemand mehr benutzt (§8). Was hineinkam, muss unverändert
// wieder heraus — auch eine BOM, auch gemischte Zeilenenden.
func TestDieQuelldateiKommtByteFuerByteZurueck(t *testing.T) {
	db := frischeDatenbank(t)

	// Mit BOM, gemischten Zeilenenden und ohne letzten Umbruch: genau das, was
	// eine Rekonstruktion verlöre.
	quelle := []byte("\ufeffmse version: 2.0.0\r\ngame:  probe\n\ncard:\n\tname: Feuerball")

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Origin: "/pfad/probe.mse-set",
		Quelle: &domain.Quelle{Datei: "set", Inhalt: quelle},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	gelesen, err := storage.NeuerSetLeser(db).Quelle(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen == nil {
		t.Fatal("keine Quelle")
	}

	if gelesen.Datei != "set" {
		t.Errorf("Datei: %q", gelesen.Datei)
	}

	if !bytes.Equal(gelesen.Inhalt, quelle) {
		t.Errorf("Inhalt:\n%q\nerwartet\n%q", gelesen.Inhalt, quelle)
	}
}

// **Ein Set ohne Quelle ist kein Fehler.**
//
// Ein Set, das in der Anwendung entsteht, hat keine — und die Byte-Zusage gilt
// dort nicht, weil es nichts gibt, womit zu vergleichen wäre.
func TestEinSetOhneQuelleIstKeinFehler(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Neu", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	quelle, err := storage.NeuerSetLeser(db).Quelle(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if quelle != nil {
		t.Errorf("Quelle: %+v, erwartet keine", quelle)
	}
}

// Eine Quelle, die sich nicht lesen lässt, wird gemeldet.
func TestEineUnlesbareQuelleWirdGemeldet(t *testing.T) {
	db, err := sql.Open("stolper-fragen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := storage.NeuerSetLeser(db).Quelle(probekontext(t), 1); fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Die Quelle nennt auch, wo sie herkommt.**
//
// # Warum das an der Quelle hängt und nicht am Setkopf
//
// Beim Speichern muss das Beiwerk mit — Bilder, Symbole —, und das liegt im
// Herkunftspaket, nicht in der Datenbank
// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
// Wer den Ort über die Kopfliste sucht, liest **alle** Sets, um einen Pfad zu
// finden, und hat zwei Wege, die auseinanderlaufen können. Er steht in
// derselben Zeile wie die Quelle, also kommt er mit ihr.
func TestDieQuelleNenntDenHerkunftsort(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Origin: "/pfad/probe.mse-set",
		Quelle: &domain.Quelle{Datei: "set", Inhalt: []byte("game: probe\n")},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	gelesen, err := storage.NeuerSetLeser(db).Quelle(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Herkunft != "/pfad/probe.mse-set" {
		t.Errorf("Herkunft: %q, erwartet %q", gelesen.Herkunft, "/pfad/probe.mse-set")
	}
}

// **Der Arbeitsort lässt sich setzen** — die Naht aus #226.
//
// Nach „Sichern unter" führt der Bestand den neuen Ort; das folgende ⌘S geht
// dorthin und nicht mehr gegen das alte Archiv.
func TestDerArbeitsortLaesstSichSetzen(t *testing.T) {
	db := frischeDatenbank(t)
	leser := storage.NeuerSetLeser(db)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Origin: "/alt/probe.mse-set",
		Quelle: &domain.Quelle{Datei: "set", Inhalt: []byte("mse version: 2.0.0\n")},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := leser.HerkunftSetzen(probekontext(t), setID, "/neu/probe.mse-set"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	quelle, fehler := leser.Quelle(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if quelle.Herkunft != "/neu/probe.mse-set" {
		t.Errorf("Herkunft: %q", quelle.Herkunft)
	}
}

// **Ein Set, das es nicht gibt, ist ein `ErrNotFound`.**
//
// Ohne dieses Urteil ginge ein Ortswechsel an einem gelöschten Set als geglückt
// durch — und der Bestand nennte einen Ort, den niemand trägt.
func TestEinArbeitsortAmUnbekanntenSetIstNichtGefunden(t *testing.T) {
	err := storage.NeuerSetLeser(frischeDatenbank(t)).
		HerkunftSetzen(probekontext(t), 4711, "/neu")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("kein ErrNotFound: %v", err)
	}
}
