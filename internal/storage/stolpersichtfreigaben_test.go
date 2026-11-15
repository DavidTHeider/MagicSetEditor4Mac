package storage

// **Die Fehlerwege der Sammlungsfreigaben** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// **Nicht `stolperfreigaben_test.go`** — den Namen trägt seit Phase 4.12 die
// Freigabe eines **Profils** an einen zweiten Nutzer dieses Rechners. Zwei
// Freigaben unter einem Dateinamen wären die Verwechslung, die man später
// nicht mehr findet.
//
// # Warum es dafür Haken braucht
//
// **Weil eine Freigabe zwei Tabellen anfasst.** Die Zeile und ihre gewählten
// Drucke stehen oder fallen zusammen; die Wege dazwischen — das Räumen der
// alten Auswahl, das Schreiben der neuen, die krumme Frist beim Lesen — gibt
// kein echter Treiber her. Ohne sie stünden sie als ungeprüfte Zusagen da
// (§10).

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen die Sammlungsfreigaben stolpern können.
const (
	beimSichtenscan stolperstelle = iota + 2400
	beimSichtenende
	beimSichtenfrist
	beimErteiltenfrist
	beimAuswahlscan
	beimAuswahlende
	beimSichtenbeginn
	beimSichtenkennung
	beimAuswahlabfrage
)

func init() {
	sql.Register("stolper-sichten-scan", stolpertreiber{stelle: beimSichtenscan})
	sql.Register("stolper-sichten-ende", stolpertreiber{stelle: beimSichtenende})
	sql.Register("stolper-sichten-frist", stolpertreiber{stelle: beimSichtenfrist})
	sql.Register("stolper-erteilte-frist",
		stolpertreiber{stelle: beimErteiltenfrist})
	sql.Register("stolper-auswahl-scan", stolpertreiber{stelle: beimAuswahlscan})
	sql.Register("stolper-auswahl-ende", stolpertreiber{stelle: beimAuswahlende})
	sql.Register("stolper-sichten-beginn",
		stolpertreiber{stelle: beimSichtenbeginn})
	sql.Register("stolper-sichten-kennung",
		stolpertreiber{stelle: beimSichtenkennung})
	sql.Register("stolper-auswahl-abfrage",
		stolpertreiber{stelle: beimAuswahlabfrage})
}

// sichtfreigabeabfrage beantwortet die zwei Abfragen der Freigabeablage.
func sichtfreigabeabfrage(
	stelle stolperstelle, abfrage string,
) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "SELECT f.id, f.empfaenger"):
		// **`beimFragen` geht durch, nicht hier hinein.** Die allgemeine
		// Stelle lässt die Abfrage selbst scheitern; wer sie hier abfinge,
		// gäbe eine Zeile zurück — und der Zweig, der einen Abfragefehler
		// meldet, bliebe unbetreten.
		if stelle == beimFragen {
			return nil, false, nil
		}

		return sichtfreigabenzeile(stelle), true, nil

	case strings.HasPrefix(geglaettet, "INSERT INTO sammlungsfreigaben"):
		// **Das Einfügen gibt seine Kennung über `RETURNING`.** Eine Zeile,
		// die sich nicht lesen lässt, ist der Weg, auf dem die Auswahl an
		// eine Kennung geschrieben würde, die es nicht gibt.
		if stelle == beimSichtenkennung {
			return &krummeZeile{spalten: []string{"id"}}, true, nil
		}

		return &werteZeile{
			spalten: []string{"id"}, werte: []driver.Value{int64(1)},
		}, true, nil

	case strings.HasPrefix(geglaettet,
		"SELECT printing FROM sammlungsfreigabeauswahl"):
		// **Die zweite Abfrage scheitert für sich.** Die allgemeine Stelle
		// ließe schon die erste scheitern — dann bliebe dieser Weg
		// unbetreten, denn er kommt erst danach.
		if stelle == beimAuswahlabfrage {
			return nil, true, errGestolpert
		}

		spalten := []string{"printing"}

		switch stelle {
		case beimAuswahlscan:
			return &krummeZeile{spalten: spalten}, true, nil

		case beimAuswahlende:
			return &abbrechendeZeile{spalten: spalten}, true, nil
		}

		return &leerezeile{spalten: len(spalten)}, true, nil
	}

	return nil, false, nil
}

// sichtfreigabenzeile gibt die Zeile einer Freigabe — oder einen Stolperstein.
//
// **Eine Auswahlfreigabe und keine über die Sammlung**: nur bei ihr liest die
// Ablage die gewählten Drucke nach, und ohne das bliebe der zweite Weg
// unbetreten.
//
// **Sie gibt immer eine Zeile.** Der Fall „kenne ich nicht" steht beim
// Aufrufer: dort geht `beimFragen` durch, damit die Abfrage selbst scheitert.
func sichtfreigabenzeile(stelle stolperstelle) driver.Rows {
	spalten := []string{
		"id", "empfaenger", "umfang", "zweck", "behalten", "ablauf", "erteilt",
	}

	switch stelle {
	case beimSichtenscan:
		return &krummeZeile{spalten: spalten}

	case beimSichtenende:
		return &abbrechendeZeile{spalten: spalten}

	// **Eine Frist, die keine ist** — sie kommt aus der Datei und nicht aus
	// diesem Programm. Wer sie für den Nullwert hielte, machte aus einer
	// kaputten Zeile eine **unbefristete** Freigabe (§8).
	case beimSichtenfrist:
		return &werteZeile{spalten: spalten, werte: []driver.Value{
			int64(1), "peer-1", "auswahl", "handel", int64(0),
			"vorgestern", "2026-09-18T12:00:00Z",
		}}

	case beimErteiltenfrist:
		return &werteZeile{spalten: spalten, werte: []driver.Value{
			int64(1), "peer-1", "auswahl", "handel", int64(0),
			"", "irgendwann",
		}}
	}

	return &werteZeile{spalten: spalten, werte: []driver.Value{
		int64(1), "peer-1", "auswahl", "handel", int64(0), "",
		"2026-09-18T12:00:00Z",
	}}
}

// **Jeder Weg des Lesens reicht seinen Fehler durch.**
func TestJederWegDerSichtfreigabenReichtSeinenFehlerDurch(t *testing.T) {
	for _, treiber := range []string{
		"stolper-fragen", "stolper-sichten-scan", "stolper-sichten-ende",
		"stolper-sichten-frist", "stolper-erteilte-frist",
		"stolper-auswahl-scan", "stolper-auswahl-ende",
		"stolper-auswahl-abfrage",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeueFreigabeablage(db).Fuer(
				probekontext(t), "peer-1"); fehler == nil {
				t.Error("das Lesen kam ohne Fehler zurück")
			}
		})
	}
}

// **Und jeder Weg des Erteilens ebenso.**
func TestJederWegDesErteilensReichtSeinenFehlerDurch(t *testing.T) {
	for _, treiber := range []string{
		"stolper-sichten-beginn", "stolper-sichten-kennung",
		"stolper-schreiben", "stolper-zweites-schreiben", "stolper-commit",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if fehler := NeueFreigabeablage(db).Erteilen(probekontext(t),
				domain.Sammlungsfreigabe{
					Empfaenger: "peer-1", Umfang: domain.UmfangAuswahl,
					Zweck: domain.ZweckHandel, Auswahl: []string{"abc"},
				}); fehler == nil {
				t.Error("das Erteilen kam ohne Fehler zurück")
			}
		})
	}
}

// **Und das Widerrufen.**
func TestDasWiderrufenReichtSeinenFehlerDurch(t *testing.T) {
	for _, treiber := range []string{"stolper-schreiben", "stolper-zaehlen"} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if fehler := NeueFreigabeablage(db).Widerrufen(probekontext(t),
				"peer-1", domain.ZweckHandel, domain.UmfangSammlung,
			); fehler == nil {
				t.Error("das Widerrufen kam ohne Fehler zurück")
			}
		})
	}
}

// **Und ohne Profil geht das Widerrufen gar nicht.**
//
// **Jede Wurzelabfrage klemmt auf ein Profil** (0066): ohne den Zweig
// widerriefe dieser Weg über alle Profile hinweg.
func TestDasWiderrufenBrauchtEinProfil(t *testing.T) {
	db, err := sql.Open("stolper-lesen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if fehler := NeueFreigabeablage(db).Widerrufen(t.Context(), "peer-1",
		domain.ZweckHandel, domain.UmfangSammlung); fehler == nil {
		t.Error("das Widerrufen kam ohne Profil durch")
	}
}
