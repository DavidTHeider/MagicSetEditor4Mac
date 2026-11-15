package storage

// **Die Stolperstellen des Nutzerspeichers** — Werkstück 1 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, damit `stolperabfragen_test.go` unter der
// Zeilengrenze aus §2 bleibt — dieselbe Form wie `stolperturnier_test.go`.
//
// # Was hier geprüft wird
//
// **Die Wege, die kein echter Treiber hergibt**: eine Abfrage, die scheitert;
// eine Zeile, die sich nicht lesen lässt; ein Abbruch **nach** der Schleife.
// Der letzte ist der, den man sonst nie sieht — `io.EOF` beendet still, und
// nur ein anderer Fehler landet hinter `zeilen.Err()`.

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
)

// Die Stellen, an denen der Nutzerweg stolpern kann.
const (
	beimNutzerscan stolperstelle = iota + 1200
	beimNutzerende
	beimProfilscan
	beimProfilende
	beimNutzerkennung
	beimProfilkennung
	beimNutzerbeginn
	beimNutzercommit

	// **Die Konten** — Werkstück 4 (#1351).
	beimKontoscan
	beimKontoende

	// **Die Pflege** — Werkstück 5 desselben Issues. Beide Zählungen stehen
	// **in** einer Transaktion und entscheiden über ein Nein; ob die Zählung
	// selbst scheitern kann, gibt kein echter Treiber her.
	beimNutzerzaehlen
	beimProfilzaehlen

	// **Die Nachfrage nach einem Schreiben, das null Zeilen traf** —
	// Werkstück 8 (#1351).
	beimProfilnachschlagen

	// **Die Profilliste der Umfangsbindung** — Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2. Sie läuft **in** der Löschtransaktion, und
	// ihre beiden Fehlerwege gibt eine echte Datenbank nicht her: eine Zeile,
	// die sich nicht lesen lässt, und ein Abbruch **nach** der Schleife.
	beimUmfangsprofilscan
	beimUmfangsprofilende

	// **Die Abfrage selbst** — und nicht das Lesen ihrer Zeilen. `beimFragen`
	// träfe hier die Namensprüfung davor, und der Lauf endete, bevor die
	// Bindung je liefe (§3).
	beimUmfangsprofilabfrage

	// **Die Zugänge eines Menschen an fremden Profilen** (#2287) — gezählt
	// **vor** der Bindung, im selben Geschäft.
	beimFremdzugang

	// **Der Aufzeichner** — Hinweis 12 desselben Nachlaufs. Er stolpert
	// nirgends; er schreibt mit (`aufzeichner_test.go`).
	beimAufzeichnen

	// **Die Karte von MTGJSON auf die Drucke** — Werkstück A3 von Phase 5.3
	// (#1442). Eine Zeile, die sich nicht lesen lässt, gibt eine echte
	// Datenbank nicht her: beide Spalten sind `TEXT NOT NULL`.
	beimKrosswalkscan
)

func init() {
	sql.Register("stolper-nutzer-scan", stolpertreiber{stelle: beimNutzerscan})
	sql.Register("stolper-nutzer-ende", stolpertreiber{stelle: beimNutzerende})
	sql.Register("stolper-profil-scan", stolpertreiber{stelle: beimProfilscan})
	sql.Register("stolper-profil-ende", stolpertreiber{stelle: beimProfilende})
	sql.Register("stolper-nutzer-kennung", stolpertreiber{stelle: beimNutzerkennung})
	sql.Register("stolper-profil-kennung", stolpertreiber{stelle: beimProfilkennung})
	sql.Register("stolper-nutzer-beginn", stolpertreiber{stelle: beimNutzerbeginn})
	sql.Register("stolper-nutzer-commit", stolpertreiber{stelle: beimNutzercommit})
	sql.Register("stolper-konto-scan", stolpertreiber{stelle: beimKontoscan})
	sql.Register("stolper-konto-ende", stolpertreiber{stelle: beimKontoende})
	sql.Register("stolper-nutzer-zaehlen", stolpertreiber{stelle: beimNutzerzaehlen})
	sql.Register("stolper-profil-zaehlen", stolpertreiber{stelle: beimProfilzaehlen})
	sql.Register("stolper-profil-nachschlagen",
		stolpertreiber{stelle: beimProfilnachschlagen})
	sql.Register("stolper-umfangsprofil-scan",
		stolpertreiber{stelle: beimUmfangsprofilscan})
	sql.Register("stolper-umfangsprofil-ende",
		stolpertreiber{stelle: beimUmfangsprofilende})
	sql.Register("stolper-umfangsprofil-abfrage",
		stolpertreiber{stelle: beimUmfangsprofilabfrage})
	sql.Register("stolper-fremdzugang", stolpertreiber{stelle: beimFremdzugang})
	sql.Register("stolper-krosswalk-scan",
		stolpertreiber{stelle: beimKrosswalkscan})
}

// nutzerabfrage beantwortet die Abfragen des Nutzerwegs — oder stolpert.
func nutzerabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	// **Seit `RETURNING` ist das Einfügen eine Abfrage** (B01, #2220). Bis
	// dahin beantwortete `nutzerExec` es, und die Stellen „das zweite
	// Einfügen" und „das Festschreiben" scheiterten danach schon hier — grün,
	// aber an der falschen Anweisung; ihre Zweige blieben unbetreten, bis die
	// Abdeckung es zeigte (C05, #2230).
	if strings.Contains(geglaettet, "RETURNING") &&
		(strings.HasPrefix(geglaettet, "INSERT INTO nutzer") ||
			strings.HasPrefix(geglaettet, "INSERT INTO profile")) {
		return einfuegeantwort(stelle, strings.HasPrefix(geglaettet, "INSERT INTO nutzer"), zurueckgegeben(geglaettet))
	}

	switch {
	case strings.HasPrefix(geglaettet, "SELECT id, name, bild, email"):
		switch stelle {
		case beimNutzerscan:
			// **Siebzehn Spalten seit Werkstück 8** (#1351): zu den acht aus
			// Werkstück 6 kamen fünf Namensteile, drei freiwillige Texte und
			// die Fassung. Eine Zeile mit weniger scheiterte an der
			// **Spaltenzahl** statt am Scan — und dann prüfte der Lauf etwas
			// anderes als seinen Fall.
			return &krummeNutzerzeile{}, true, nil
		case beimNutzerende:
			return &abbrechendeNutzerzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 17}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT id, nutzer_id, name, fassung, subslogan"):
		switch stelle {
		case beimProfilscan:
			// **Acht Spalten seit Werkstück 9** (#1351): zu den vier aus
			// Werkstück 8 kamen Subslogan, zwei Beschreibungen und das
			// Abzeichen. Eine Zeile mit weniger scheiterte an der
			// **Spaltenzahl** statt am Scan — und dann prüfte der Lauf etwas
			// anderes als seinen Fall.
			return &krummeZeile{spalten: nutzerspaltennamen(profilspaltenzahl)}, true, nil
		case beimProfilende:
			return &abbrechendeProfilzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: profilspaltenzahl}, true, nil

	// **Die Frage, ob es den Nutzer gibt** — eine Zahl.
	case strings.HasPrefix(geglaettet, "SELECT count(*) FROM nutzer WHERE id"):
		if stelle == beimFragen {
			return nil, true, errGestolpert
		}

		return &einezahl{}, true, nil

	case strings.HasPrefix(geglaettet,
		"SELECT id, nutzer_id, anbieter, subjekt, anzeigename"):
		switch stelle {
		case beimKontoscan:
			return &krummeFuenferzeile{}, true, nil
		case beimKontoende:
			return &abbrechendeFuenferzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 5}, true, nil

	// **Die Namensprüfung vor dem Löschen** — Werkstück 5 (#1351).
	//
	// **Sie muss glücken**, damit die Wege dahinter überhaupt erreicht werden:
	// scheitert schon sie, sind das Zählen und das Löschen unbetreten (§3).
	case strings.HasPrefix(geglaettet, "SELECT name FROM nutzer WHERE id"):
		if stelle == beimFragen {
			return nil, true, errGestolpert
		}

		return &namenszeile{name: "Ich"}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT name FROM profile WHERE id"):
		if stelle == beimFragen {
			return nil, true, errGestolpert
		}

		return &namenszeile{name: "Mein Profil"}, true, nil

	// **Der Mensch zum Profil und die beiden Einstellungstafeln** — der
	// Aufzeichner braucht Antworten, sonst endete der Weg vor dem, was er
	// aufzeichnen soll (Hinweis 12 des Nachlaufs zum Zweitblick über Phase
	// 5.2).
	//
	// **Und nur für ihn.** Diese drei Abfragen beantwortet sonst ein anderer
	// Zweig, der seine Stolperstellen kennt; hier zu antworten nähme ihm jeden
	// Fehlerweg — nachgemessen am 2026-09-11, als neun Läufe grün wurden, die
	// einen Fehler erwarteten.
	// **Die Karte von MTGJSON auf die Drucke** (A3, #1442).
	case strings.HasPrefix(geglaettet, "SELECT fremd.wert, eigen.wert"):
		switch stelle {
		case beimKrosswalkscan:
			// **Zwei NULL und keine krumme Zahl**: `Scan` auf zwei
			// Textspalten nimmt eine Zahl klaglos an und formatiert sie.
			// Nachgemessen am 2026-09-11 — der erste Anlauf war grün.
			return &nullzweierzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 2}, true, nil

	case stelle == beimAufzeichnen &&
		strings.HasPrefix(geglaettet, "SELECT nutzer_id FROM profile WHERE id"):
		return &zahlenzeile{zahl: 1}, true, nil

	case stelle == beimAufzeichnen &&
		(strings.HasPrefix(geglaettet, "SELECT schluessel, wert FROM settings") ||
			strings.HasPrefix(geglaettet,
				"SELECT schluessel, wert FROM nutzereinstellungen")):
		return &leerezeile{spalten: 2}, true, nil

	// **Die Profilliste der Umfangsbindung** — Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2.
	//
	// **Leer als Vorgabe, und das genügt hier**: geprüft werden die beiden
	// Wege, auf denen das Lesen selbst scheitert. Was die Bindung mit einer
	// gefüllten Liste tut, misst `umfangsbindung_test.go` gegen SQLite.
	// **Die fremden Zugänge** (#2287) — sie stehen vor der Profilliste, und
	// scheiterten sie mit jeder Stelle, erreichte kein Lauf die Wege dahinter.
	case strings.HasPrefix(geglaettet, "SELECT count(*) FROM profilfreigaben f JOIN profile p"):
		if stelle == beimFremdzugang {
			return nil, true, errGestolpert
		}

		return &einezahl{}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT id, name FROM profile WHERE"):
		switch stelle {
		case beimUmfangsprofilscan:
			return &krummeZweierzeile{}, true, nil
		case beimUmfangsprofilende:
			return &abbrechendeZweierzeile{}, true, nil
		case beimFragen, beimUmfangsprofilabfrage:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 2}, true, nil

	// **Das Zählen der Nutzer und der Profile** — Werkstück 5 (#1351).
	case strings.HasPrefix(geglaettet, "SELECT count(*) FROM nutzer\n"),
		geglaettet == "SELECT count(*) FROM nutzer":
		if stelle == beimNutzerzaehlen {
			return nil, true, errGestolpert
		}

		// **Zwei und nicht eine.** Bei eins greift „der letzte bleibt", und
		// das Löschen dahinter wäre unbetreten (§3).
		return &zahlenzeile{zahl: 2}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT count(*) FROM profile\n"):
		if stelle == beimProfilzaehlen {
			return nil, true, errGestolpert
		}

		return &zahlenzeile{zahl: 2}, true, nil

	// **Die Frage, ob es das Profil gibt** — dieselbe Gestalt, und sie muss
	// bei `beimFragen` **glücken**, damit das Schreiben des Zeigers dahinter
	// überhaupt läuft: scheitert schon sie, bleibt der zweite Fehlerweg
	// unbetreten (§3).
	//
	// **Seit Werkstück 8 fragt sie noch jemand** (#1351): `ProfilAendern`
	// unterscheidet damit „gibt es nicht" von „hat sich geändert". Für
	// **diesen** Weg gibt es eine eigene Stelle — eine gemeinsame nähme dem
	// Zeiger seinen Lauf.
	// **Und seit Werkstück 15 ein Dritter**: `Vorgabenablage.Vorgaben` fragt
	// damit, ob es das Profil überhaupt gibt (§8). Auch der bekommt eine
	// eigene Stelle — aus demselben Grund wie die Zeile darüber.
	case strings.HasPrefix(geglaettet, "SELECT count(*) FROM profile WHERE id"):
		if stelle == beimProfilnachschlagen ||
			stelle == beimVorgabennachschlagen {
			return nil, true, errGestolpert
		}

		return &einezahl{}, true, nil
	}

	return nil, false, nil
}

// zurueckgegeben nennt die Spalten hinter `RETURNING` — **aus der Anweisung
// gelesen**: hier standen fest `id, fassung`, und das Profil gibt seit B01
// auch seine Sichtbarkeit zurück. Jedes Einfügen eines Profils scheiterte
// danach an der Spaltenzahl, auch dort, wo es glücken sollte (Zweitblick
// über Phase B, Befund 3).
func zurueckgegeben(anweisung string) []string {
	_, hinten, _ := strings.Cut(anweisung, "RETURNING")

	spalten := strings.Split(hinten, ",")
	for i, spalte := range spalten {
		spalten[i] = strings.TrimSpace(spalte)
	}

	return spalten
}

// einfuegeantwort beantwortet ein Einfügen mit `RETURNING …` — dieselben
// Stellen wie `nutzerExec` für die Form ohne.
func einfuegeantwort(stelle stolperstelle, nutzer bool, spalten []string) (driver.Rows, bool, error) {
	switch {
	case stelle == beimNutzerkennung && nutzer, stelle == beimProfilkennung && !nutzer:
		return &einfuegezeile{spalten: spalten, krumm: true}, true, nil
	case stelle == beimSchreiben:
		return nil, true, errGestolpert
	case stelle == beimZweitenSchreiben:
		zweiterSchreibzaehler++
		if zweiterSchreibzaehler >= 2 {
			return nil, true, errGestolpert
		}
	}

	return &einfuegezeile{spalten: spalten}, true, nil
}

// einfuegezeile ist die Antwort eines Einfügens: je Spalte ein Wert — die
// Sichtbarkeit als Text, alles andere als Zahl —, oder, krumm, lauter NULL,
// an denen der Scan in eine Zahl scheitert.
type einfuegezeile struct {
	spalten        []string
	krumm, gegeben bool
}

func (r *einfuegezeile) Columns() []string { return r.spalten }
func (r *einfuegezeile) Close() error      { return nil }

func (r *einfuegezeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	for i, spalte := range r.spalten {
		switch {
		case r.krumm:
			werte[i] = nil
		case spalte == "sichtbarkeit":
			werte[i] = "privat"
		default:
			werte[i] = int64(1)
		}
	}

	return nil
}
