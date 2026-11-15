// **Was der Stolpertreiber auf eine Abfrage antwortet.**
//
// Eine zweite Datei, weil `stolpertreiber_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **wo** gestolpert wird,
// hier, **was** zurückkommt, solange er es nicht tut.
package storage

import (
	"database/sql/driver"
	"strings"
)

func (s stolperstmt) Query([]driver.Value) (driver.Rows, error) {
	// **Die Buchführung unter der Sperre ist leer** (Vermerk 0189): die Folge
	// liest sie seither innerhalb ihrer Transaktion, und ein Lauf über das
	// gescheiterte Commit soll an diesem scheitern und nicht am Lesen.
	if s.stelle == beimCommit &&
		strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT name, pruefsumme FROM migrationen") {
		return &stolperrows{leer: true}, nil
	}

	// **Die Profilklemme vor allem anderen** (#1815): sie steht vor jeder
	// Ablage, die eine Kennung annimmt, und antwortet jedem Lauf außer ihrem
	// eigenen mit „gehört dir".
	if zeilen, erkannt, fehler := besitzabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Der Handelsabschluss danach** (T6, #1471). Er liest
	// `reservierungen`, `sammlung` und `drucke` — alles Tabellen, die auch
	// andere Stellen lesen. Deshalb steht er vorn **und** fragt zuerst, ob
	// die Stelle überhaupt seine ist: sonst nähme er fremden Läufen ihre
	// Antworten ab (§10).
	if zeilen, erkannt, fehler := abschlussabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Der Arbeitsverlauf zuerst** (#1422) — seine Stellen kennt sonst niemand.
	if zeilen, erkannt := arbeitsverlaufabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Dann der Bereichswechsel** (Befund 7 des Zweitblicks über Phase 5.1).
	if zeilen, erkannt, fehler := bereichabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Dann die Druckablage** (#1441) — ihre Stellen kennt sonst niemand.
	if zeilen, erkannt := druckabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und die Standprüfung** (Befund 2 des Zweitblicks über Phase 5.3) —
	// aus demselben Grund in einer eigenen Datei.
	if zeilen, erkannt, fehler := standabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und die Modellproben** (A9, #1448) — aus demselben Grund in einer
	// eigenen Datei.
	if zeilen, erkannt := probenabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und die Kontaktablage** (N1, #1454) — eigene Datei, §2.
	if zeilen, erkannt := kontaktabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und der Partieweg** (G1, #1458) — aus demselben Grund eigene Datei.
	if zeilen, erkannt := partieabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und Freigabe und Leihe** (G5 und G6, #1462 und #1463) — dito.
	if zeilen, erkannt := leihabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und die Zusagen des Handels** (T2, #1467) — dito.
	if zeilen, erkannt, fehler := reservierungsabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und die freigegebenen Sichten** (T3, #1468) — dito. Sie gibt einen
	// Fehler mit zurück: die **zweite** Abfrage muss für sich scheitern
	// können, und die allgemeine Stelle ließe schon die erste fallen.
	if zeilen, erkannt, fehler := sichtfreigabeabfrage(
		s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und der Ausgang des Handels** (T5, #1470) — dito.
	if zeilen, erkannt := ausgangsabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und der Briefkasten des Handels** (T5, #1470) — dito.
	if zeilen, erkannt := briefkastenabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Der Turnierweg zuerst** — er steht in einer eigenen Datei, damit
	// diese hier unter der Zeilengrenze aus §2 bleibt. Wer dort nichts
	// erkennt, gibt `false` zurück und lässt alles Folgende unverändert.
	if zeilen, erkannt, fehler := turnierabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und der Nutzerweg** — aus demselben Grund in einer eigenen Datei
	// (#1351).
	if zeilen, erkannt, fehler := nutzerabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und der Faktorweg** — aus demselben Grund in einer eigenen Datei
	// (#1351,
	// Werkstück 7).
	if zeilen, erkannt, fehler := faktorabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und der Tokenweg** — aus demselben Grund in einer eigenen Datei
	// (#1351,
	// Werkstück 10).
	if zeilen, erkannt, fehler := tokenabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und die Profilvorgaben** — Werkstück 15 desselben Issues.
	if zeilen, erkannt, fehler := vorgabenabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und der Umfang einer Rücksetzung** — Werkstück O4 von Phase 5.2 (#1428).
	if zeilen, erkannt := umfangabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und das Kostenjournal** — Werkstück O12 von Phase 5.2 (#1439).
	if zeilen, erkannt, fehler := kostenabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und die Datenpflege** — Werkstück O8 von Phase 5.2 (#1432).
	if zeilen, erkannt := pflegeabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, nil
	}

	// **Und das Wiederherstellungsbuch** — Werkstück O7 von Phase 5.2 (#1431).
	if zeilen, erkannt, fehler := rueckkehrabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und die Freigaben** — Werkstück 11 desselben Issues.
	if zeilen, erkannt, fehler := freigabenabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Und die Buchführung über die Fassung** — aus demselben Grund in einer
	// eigenen Datei.
	if zeilen, erkannt, fehler := fassungsabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// **Die Abfragen des Paketkatalogs** — in einer eigenen Datei, damit diese
	// hier unter der Zeilengrenze aus §2 bleibt. Wer dort nichts erkennt, gibt
	// `false` zurück und lässt alles Folgende unverändert.
	if zeilen, erkannt, fehler := katalogabfrage(s.stelle, s.abfrage); erkannt {
		return zeilen, fehler
	}

	// Die Vorprüfung der Massenänderung fragt vorhandene Kennungen ab — seit
	// #1815 über ihr Set auf das Profil geklemmt, und deshalb mit Tafelnamen.
	if geglaettet := strings.TrimSpace(s.abfrage); strings.HasPrefix(geglaettet, "SELECT id FROM cards") ||
		strings.HasPrefix(geglaettet, "SELECT cards.id FROM cards") {
		switch s.stelle {
		case beimMassenvorwert, beimMassenverlauf, beimErwartungcommit:
			return &kennungszeilen{}, nil
		}
	}

	// **Der Verlauf der Feldänderung fragt Wert und Stelle ab — zwei Spalten**
	// (#160). Der Präfix muss beide nennen, sonst fällt die Abfrage durch und
	// bekommt die Zeile der Vorprüfung: drei Spalten statt zwei, und der Lauf
	// scheitert am Scan statt an der Stelle, um die es geht.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT wert, reihenfolge FROM card_values") {
		if s.stelle == beimVerlaufslesen {
			return nil, errGestolpert
		}

		switch s.stelle {
		case beimVerlaufseintrag, beimVerlaufkuerzen, beimZweitenSchreiben, beimFeldcommit:
			return &nachwertzeile{}, nil
		}
	}

	// **Die Setänderung fragt Wert und Stelle ab — zwei Spalten**
	// (#567).
	//
	// Dieselbe Falle wie darüber, andere Tabelle: ohne diese Zeilen bekäme sie
	// die achtspaltige Verlaufszeile und scheiterte am Scan statt an der
	// Stelle, um die es geht.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT wert, reihenfolge FROM set_values") {
		switch s.stelle {
		case beimSetkuerzen, beimSetcommit:
			return &nachwertzeile{}, nil
		}
	}

	// **Und ihre Vorprüfung drei Spalten** — ob es das Set gibt, der alte Wert,
	// die alte Stelle. `vorwertzeile` beantwortet genau das.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT (SELECT 1 FROM sets") {
		switch s.stelle {
		case beimSetkuerzen, beimSetcommit:
			return &vorwertzeile{}, nil
		}
	}

	// **Das Kartenlöschen fragt die Stelle der Karte — eine Spalte** (#621).
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT reihenfolge FROM cards") {
		switch s.stelle {
		case beimKarteschreiben, beimKartecommit, beimKarteverlauf,
			beimKartefeldscan, beimKartefeldende, beimKartefeldabfrage,
			beimKartesetid, beimKarteeinreihen, beimKartefeldeinsetzen,
			beimKartekuerzen:
			return &kartenstellezeile{}, nil
		}
	}

	// **Das Umordnen fragt beide Karten in einer Abfrage — drei Spalten**
	// (#628).
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT id, set_id, reihenfolge FROM cards") {
		switch s.stelle {
		case beimStellenscan:
			return &krummeStellenzeile{}, nil

		case beimStellenende:
			return &stolperndesStellenende{}, nil

		case beimStellenverlauf, beimStellenparken, beimStellensetzen, beimKartecommit:
			return &stellenzeilen{}, nil
		}
	}

	// **Das Kopieren fragt zusätzlich das Set der Karte — eine Spalte** (#625).
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT set_id FROM cards") {
		if s.stelle == beimKartesetid {
			return nil, errGestolpert
		}

		switch s.stelle {
		case beimKarteeinreihen, beimKartefeldeinsetzen, beimKarteverlauf,
			beimKartekuerzen, beimKartecommit:
			return &kartenstellezeile{}, nil
		}
	}

	// **Und legt die Kopie mit `RETURNING id, reihenfolge` an — zwei Spalten.**
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO cards") {
		if s.stelle == beimKarteeinreihen {
			return nil, errGestolpert
		}

		switch s.stelle {
		case beimKartefeldeinsetzen, beimKarteverlauf, beimKartekuerzen, beimKartecommit:
			return &kartenanlegezeile{}, nil
		}
	}

	// **Und ihre Felder — drei Spalten.**
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT feld, wert, reihenfolge FROM card_values") {
		switch s.stelle {
		case beimKarteschreiben, beimKartecommit, beimKarteverlauf,
			beimKartesetid, beimKarteeinreihen, beimKartefeldeinsetzen,
			beimKartekuerzen:
			return &kartenfeldzeilen{}, nil

		// **Die Abfrage selbst scheitert** — nicht erst das Lesen. Die beiden
		// Zweige tragen dieselbe Meldung und sind nur so zu unterscheiden.
		case beimKartefeldabfrage:
			return nil, errGestolpert

		case beimKartefeldscan:
			return &krummesKartenfeld{}, nil

		case beimKartefeldende:
			return &stolperndesKartenende{}, nil
		}
	}

	// **Die Schlüsselwortpflege legt mit `RETURNING id, reihenfolge` an — zwei
	// Spalten** (#1077).
	//
	// Ohne diese Zeilen bekäme sie die achtspaltige Verlaufszeile aus dem
	// Schalter unten und scheiterte am Scan statt am Bestätigen, um das es
	// geht — derselbe Fehler wie beim Paketkatalog darunter.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO set_keywords") {
		if s.stelle == beimWortcommit {
			return &kartenanlegezeile{}, nil
		}
	}

	// **Das Setlöschen fragt den Namen ab — eine Spalte** (#1218).
	//
	// Ohne diese Zeilen bekäme es die achtspaltige Verlaufszeile aus dem
	// Schalter unten und scheiterte am Scan statt an der Stelle, um die es
	// geht — derselbe Fehler wie beim Paketkatalog darunter.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT name FROM sets") {
		if s.stelle == beimSetnamenLesen {
			return nil, errGestolpert
		}

		return &einspaltigeZeile{}, nil
	}

	// **Der Paketkatalog schreibt mit `RETURNING id` — eine Spalte** (#54).
	//
	// Ohne diese Zeilen bekäme er die achtspaltige Verlaufszeile aus dem
	// Schalter unten und scheiterte am Scan statt an der Stelle, um die es
	// geht. Genau dieser Fehler ist in Phase 6 zweimal aufgetreten: „wer eine
	// Abfrage um eine Spalte erweitert, zieht die Zeilengestalten mit".
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO packages") {
		if s.stelle == beimFragen {
			return nil, errGestolpert
		}

		return &kennungszeilen{}, nil
	}

	// Die Massenänderung fragt weiterhin nur den Wert ab — eine Spalte.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT wert FROM card_values") {
		if s.stelle == beimVerlaufslesen || s.stelle == beimMassenvorwert {
			return nil, errGestolpert
		}

		switch s.stelle {
		// **Auch `beimFeldcommit`**, seit die Feldänderung ihren Verlauf in
		// derselben Transaktion schreibt (#52): ohne diese Zeile scheitert der
		// Lauf schon am Scan und erreicht den Commit nie.
		case beimVerlaufseintrag, beimVerlaufkuerzen, beimZweitenSchreiben,
			beimFeldcommit, beimMassenverlauf:
			return &einspaltigeZeile{}, nil
		}
	}

	// **Die Herkunft des vorigen Werts — eine Spalte** (#1036). Ohne diese
	// Zeilen bekäme sie die dreispaltige Vorwertzeile aus dem Schalter unten
	// und scheiterte am Scan statt an der Stelle, um die es geht.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "SELECT herkunft FROM wert_herkunft") {
		if s.stelle == beimHerkunftLesen {
			return nil, errGestolpert
		}

		return &einspaltigeZeile{}, nil
	}

	switch s.stelle {
	case beimLesen:
		return &stolperrows{}, nil

	case beimUmwandeln:
		return &krummezeile{}, nil

	case beimFragen:
		return nil, errGestolpert

	case beimCommit, beimSchreiben:
		return &kennungszeilen{}, nil

	case beimFeldschreiben, beimFeldcommit, beimZurueckrollen:
		return &vorwertzeile{}, nil

	case beimVerlaufschreiben, beimVerlaufcommit, beimZweitenSchreiben, beimVerlaufzeiger:
		return &verlaufszeile{}, nil

	case beimVerlaufseintrag, beimVerlaufkuerzen, beimVerlaufslesen,
		beimHerkunftSchreiben, beimHerkunftLoeschen, beimHerkunftLesen:
		return &vorwertzeile{}, nil
	}

	return &stolperrows{leer: true}, nil
}
