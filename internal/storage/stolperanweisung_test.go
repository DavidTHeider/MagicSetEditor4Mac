// **Die Anweisung des Stolpertreibers** — wie eine einzelne Abfrage stolpert.
//
// Aus `stolperwerk_test.go` herausgelöst, als es über 400 Zeilen ging (§2).
// Der Schnitt folgt der Frage: dort **Treiber, Verbindung und Zug**, hier die
// **Anweisung** und was sie beantwortet.
package storage

import (
	"database/sql/driver"
	"strings"
)

// stolperstmt merkt sich seine Abfrage.
//
// **Weil eine Antwort zur Frage passen muss.** Der Verlauf liest den neuen Wert
// mit **einer** Spalte, die Feldänderung ihre Vorprüfung mit **zwei**. Ein
// Treiber, der immer dasselbe liefert, scheitert am Scan statt an der Stelle,
// die geprüft werden soll.
type stolperstmt struct {
	stelle  stolperstelle
	abfrage string
}

func (s stolperstmt) Close() error  { return nil }
func (s stolperstmt) NumInput() int { return -1 }

func (s stolperstmt) Exec([]driver.Value) (driver.Result, error) {
	// **Die Druckablage zuerst** (#1441) — ihre Stellen kennt sonst niemand.
	if erkannt, fehler := bereichschreiben(s.stelle, s.abfrage); erkannt {
		return nil, fehler
	}

	if ergebnis, erkannt := druckschreiben(s.stelle, s.abfrage); erkannt {
		return ergebnis, nil
	}

	// **Der Partieweg** (G1, #1458) — eigene Datei, §2.
	if ergebnis, erkannt, fehler := partieschreiben(s.stelle, s.abfrage); erkannt {
		return ergebnis, fehler
	}

	// **Und Freigabe und Leihe** (G5 und G6, #1462 und #1463) — dito.
	if ergebnis, erkannt, fehler := leihschreiben(s.stelle, s.abfrage); erkannt {
		return ergebnis, fehler
	}

	// **Und der Handelsabschluss** (T6, #1471) — dito. **Vor allem
	// Übrigen**, weil er Anweisungen beantwortet, deren Vorsatz auch
	// anderswo vorkommt; er nimmt sie nur an seinen eigenen Stellen an.
	if ergebnis, erkannt, fehler := abschlussschreiben(s.stelle, s.abfrage); erkannt {
		return ergebnis, fehler
	}

	// **Die Datenpflege** — Werkstück O8 von Phase 5.2 (#1432).
	if ergebnis, erkannt, fehler := s.pflegeExec(); erkannt {
		return ergebnis, fehler
	}

	// **Das Wiederherstellungsbuch** — Werkstück O7 von Phase 5.2 (#1431).
	if ergebnis, erkannt, fehler := s.rueckkehrExec(); erkannt {
		return ergebnis, fehler
	}

	// **Das Fortschreiben der Fassung zuerst** — es steht in einer eigenen
	// Datei, damit diese hier unter der Zeilengrenze aus §2 bleibt.
	if ergebnis, erkannt, fehler := fassungsschreiben(s.stelle, s.abfrage); erkannt {
		return ergebnis, fehler
	}

	// **Die Kürzung der Setänderung, an ihrer Anweisung erkannt.** Ein Zähler
	// wäre hier die falsche Wahl: er hinge an der Zahl der Schreibvorgänge,
	// und die ändert sich, sobald jemand einen dazwischen einfügt.
	if s.stelle == beimSetkuerzen &&
		strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM history") {
		return nil, errGestolpert
	}

	// **Das Löschen eines Sets** (#1218) — dieselbe Frage, eine andere
	// Anweisung.
	if s.stelle == beimSetloeschen &&
		strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM sets") {
		return nil, errGestolpert
	}

	// **Die Rundennummer wird nach dem Einfügen gelesen** — und `LastInsertId`
	// scheitert beim echten Treiber nie. Ohne diese Stelle bliebe der Zweig
	// dahinter unbetreten.
	if s.stelle == beimRundenkennung &&
		strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO turnierrunden") {
		return kennungsfehler{}, nil
	}

	// **Abmelden und Löschen zählen ebenso.** Alle machen aus null Zeilen ein
	// `ErrNotFound` oder einen Konflikt; ob das Zählen selbst scheitern kann,
	// gibt kein echter Treiber her. Welche Anweisungen dazugehören, steht in
	// [zaehlendeAnweisung] — die Aufzählung wuchs mit jedem Werkstück, und
	// hier ginge die Datei mit ihr über die Zeilengrenze aus §2.
	if s.stelle == beimZaehlen && zaehlendeAnweisung(s.abfrage) {
		return zaehlfehler{}, nil
	}

	// **Die Turniereinfügungen brauchen eine brauchbare Kennung.**
	//
	// `driver.RowsAffected` hat keine — sein `LastInsertId` gibt einen
	// Fehler. Ohne diese Zeilen scheiterte jeder Turnierweg dort, und alles
	// dahinter (das Festschreiben, die weiteren Paarungen) bliebe ungeprüft.
	//
	// **Die allgemeinen Stellen behalten ihren Vorrang** — sonst legte diese
	// Zeile `stolper-schreiben` still, und der Fehlerweg jedes Einfügens
	// bliebe unbetreten.
	if ergebnis, meiner := s.turnierExec(); meiner {
		return ergebnis, nil
	}

	// **Und der Tokenweg** (#1351, Werkstück 10) — beide antworten dort, wo
	// auch ihre Stolperstellen stehen (§2, Zeilengrenze).
	if ergebnis, meiner := s.tokenExec(); meiner {
		return ergebnis, nil
	}

	// **Der Vorgabenweg gibt seinen Fehler selbst** — anders als die beiden
	// darüber: sein Räumen steht **vor** der Schleife, und `stolper-schreiben`
	// träfe sonst nur das Eintragen darin (Werkstück 15, #1351).
	if ergebnis, meiner, fehler := s.vorgabenExec(); meiner {
		return ergebnis, fehler
	}

	if ergebnis, meiner, fehler := s.freigabenExec(); meiner {
		return ergebnis, fehler
	}

	// **Die Nutzerwege ebenso** (#1351). Ein Nutzer entsteht samt seinem
	// ersten Profil; ohne diese Zeilen scheiterte das Einfügen schon an der
	// Kennung, und das zweite Einfügen bliebe unbetreten.
	if ergebnis, meiner, fehler := s.nutzerExec(); meiner {
		return ergebnis, fehler
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO turnier") &&
		s.stelle != beimSchreiben && s.stelle != beimRundenkennung {
		return kennung(1), nil
	}

	// **Der Stilwechsel zählt die betroffene Zeile** (#41).
	//
	// `StilSetzen` macht aus null betroffenen Zeilen ein `ErrNotFound` — mit
	// dem Vorgabewert `RowsAffected(0)` käme jeder Lauf dort zum Stehen, und
	// die Wege dahinter (Räumen, Schreiben, Commit) blieben ungeprüft. Der
	// Zähler wird trotzdem geführt: die Stelle `beimZweitenSchreiben` meint
	// dann das Räumen.
	// **Und der Ortswechsel** (#226): dieselbe Frage, eine andere Spalte.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE sets SET herkunft") {
		if s.stelle == beimSchreiben {
			return nil, errGestolpert
		}

		if s.stelle == beimZaehlen {
			return zaehlfehler{}, nil
		}

		return driver.RowsAffected(1), nil
	}

	// **Und der Vermerk »zuletzt benutzt« zählt ebenso** (#891): dieselbe
	// Frage, eine andere Spalte.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE sets SET geoeffnet") {
		if s.stelle == beimSchreiben {
			return nil, errGestolpert
		}

		if s.stelle == beimZaehlen {
			return zaehlfehler{}, nil
		}

		return driver.RowsAffected(1), nil
	}

	// **Der Zettel zählt seine Zeile** (Nachlauf über ID-02): sonst endete
	// `KennwortMitCodeSetzen` am Code, und das scheiternde Kennwort bliebe ungeprüft.
	if s.stelle == beimZweitenSchreiben && strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE wiederherstellungscodes") {
		zweiterSchreibzaehler++

		return driver.RowsAffected(1), nil
	}

	// **Und die Schlüsselwortpflege zählt ebenso** (#1062): `getroffen` macht
	// aus null betroffenen Zeilen ein `ErrNotFound`, und ob die Zählung selbst
	// scheitern kann, gibt kein echter Treiber her.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE set_keywords") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	// **Und die Nutzerpflege zählt ebenso** (#1351, Werkstück 5): aus null
	// betroffenen Zeilen wird dort „gibt es nicht", und ist die **Zählung**
	// gescheitert, ist die Frage unbeantwortet.
	if (strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE nutzer") ||
		strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE profile")) &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	// **Und das Löschen von Nutzer und Profil**: der Weg dorthin führt durch
	// eine Transaktion, und ein `DELETE`, das klagt, ist der Zweig, den kein
	// echter Treiber hergibt.
	if (strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM nutzer") ||
		strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM profile")) &&
		s.stelle == beimSchreiben {
		return nil, errGestolpert
	}

	// **Und das Lösen eines Anbieterkontos** (#1351): aus null betroffenen
	// Zeilen wird dort „der Nutzer hat kein Konto bei diesem Anbieter" —
	// und wer das liest, hört auf zu suchen. Ist die **Zählung** gescheitert,
	// ist die Frage unbeantwortet, und das ist etwas anderes.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM anbieterkonten") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	// **Und das Lösen eines Verweises** (#1167): aus null betroffenen Zeilen
	// wird dort „es gab keinen", und diese Auskunft entscheidet, ob eine
	// Fläche Erfolg meldet, wo nichts geschah.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM karten_verweis") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	// **Und das Abnehmen einer Marke** (#1208): dieselbe Frage, dieselbe
	// Auskunft — „die Karte trug sie gar nicht" darf nicht dabei herauskommen,
	// wenn schon das Zählen scheiterte.
	// **Und die Autoersetzung zählt ebenso** (#1341): dieselbe Frage, eine
	// andere Tabelle — „die Regel stand nicht da" gegen „das Zählen ging
	// schief".
	if strings.HasPrefix(strings.TrimSpace(s.abfrage),
		"DELETE FROM spiel_autoersetzungen") && s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM karten_marken") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	// **Und die Decks** (#1293): dieselbe Frage, zwei weitere Tabellen — „das
	// Deck stand nicht da" gegen „das Zählen ging schief".
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM decks") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM deckplaetze") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	// **Und die Änderung eines Decks** (#1293, Werkstück 3): sie zählt, um
	// „gibt es nicht" von „überholt" zu unterscheiden. Zählt der Treiber
	// nicht, ist es **keines von beidem** — und beides zu melden wäre falsch.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE decks") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	// **Und die Formate** (#1293, Werkstück 4): dieselbe Frage, dieselben
	// zwei Antworten — an der Änderung und am Löschen.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE formate") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM formate") &&
		s.stelle == beimZaehlen {
		return zaehlfehler{}, nil
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE sets SET spiel") {
		if s.stelle == beimSchreiben {
			return nil, errGestolpert
		}

		if s.stelle == beimZaehlen {
			return zaehlfehler{}, nil
		}

		zweiterSchreibzaehler++

		return driver.RowsAffected(1), nil
	}

	// **Das Löschen einer Karte schreibt dreimal**: den Verlaufseintrag, die
	// Kürzung, und dann die Karte weg. `beimKarteschreiben` lässt die dritte
	// scheitern, `beimKartecommit` keine — die stolpert erst beim Commit.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM cards") &&
		s.stelle == beimKarteschreiben {
		return nil, errGestolpert
	}

	// **Und `beimKarteverlauf` lässt den Verlaufseintrag scheitern** — den
	// Fall, in dem die Karte verschwände, ohne dass ein Weg zurück entstünde.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO history") &&
		s.stelle == beimKarteverlauf {
		return nil, errGestolpert
	}

	// **Das Umordnen parkt beide Karten und setzt dann zwei Stellen** (#628).
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE cards SET reihenfolge = -1") &&
		s.stelle == beimStellenparken {
		return nil, errGestolpert
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE cards SET reihenfolge = ?") &&
		s.stelle == beimStellensetzen {
		return nil, errGestolpert
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO history") &&
		s.stelle == beimStellenverlauf {
		return nil, errGestolpert
	}

	// **Das Kopieren schreibt jeden Feldwert einzeln** (#625).
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO card_values") &&
		s.stelle == beimKartefeldeinsetzen {
		return nil, errGestolpert
	}

	// **Und kürzt danach den Verlauf** — dieselbe Anweisung wie beim Löschen,
	// aber ein Lauf, der vorher nicht stolpert.
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM history") &&
		s.stelle == beimKartekuerzen {
		return nil, errGestolpert
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO styling_values") &&
		s.stelle == beimStiloption {
		return nil, errGestolpert
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "UPDATE history SET zurueckgenommen") &&
		s.stelle == beimVerlaufzeiger {
		return nil, errGestolpert
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO wert_herkunft") &&
		s.stelle == beimKartenherkunftSchreiben {
		return nil, errGestolpert
	}

	// **Die Herkunftszeile, an ihrer Anweisung erkannt** (#1025).
	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO wert_herkunft") &&
		s.stelle == beimHerkunftSchreiben {
		return nil, errGestolpert
	}

	if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM wert_herkunft") &&
		s.stelle == beimHerkunftLoeschen {
		return nil, errGestolpert
	}

	switch s.stelle {
	case beimSchreiben, beimFeldschreiben, beimZurueckrollen, beimVerlaufschreiben:
		return nil, errGestolpert

	case beimZweitenSchreiben:
		zweiterSchreibzaehler++
		if zweiterSchreibzaehler >= 2 {
			return nil, errGestolpert
		}

	// **Am Text erkannt und nicht gezählt** (#1025). Hier stand `>= 2` und
	// `>= 3`, gemeint war „der Verlaufseintrag" und „die Kürzung" — und
	// zwischen Feldwert und Verlauf kam eine dritte Anweisung dazu. Ein Zähler
	// zeigt dann auf eine andere Anweisung, ohne dass jemand es sieht; er
	// scheiterte hier nur deshalb sichtbar, weil die erwartete Meldung
	// danebenlag. Wer an der Anweisung festmacht, prüft, was er meint.
	case beimVerlaufseintrag:
		if strings.HasPrefix(strings.TrimSpace(s.abfrage), "INSERT INTO history") {
			return nil, errGestolpert
		}

	case beimVerlaufkuerzen:
		if strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM history") {
			return nil, errGestolpert
		}

	// **Beim zweiten Schreiben der Massenänderung.** Das erste ist der Feldwert,
	// das zweite der Verlaufseintrag; wer beim ersten stolpert, sieht den
	// Verlaufsweg nie.
	case beimMassenverlauf:
		zweiterSchreibzaehler++
		if zweiterSchreibzaehler >= 2 {
			return nil, errGestolpert
		}
	}

	return driver.RowsAffected(0), nil
}
