// Ein Treiber, der auf Ansage scheitert.
//
// Zwei Fehlerpfade im Migrator lassen sich mit einer echten SQLite-Datenbank
// nicht herbeiführen: ein `Commit`, der scheitert, obwohl alles davor geglückt
// ist, und ein `Rows.Err()`, das nach einer vollständigen Schleife etwas meldet.
// Beide gehören zu den Fällen, die im Betrieb vorkommen — eine volle Platte,
// eine abgebrochene Verbindung — und die man nie sieht, weil sie selten sind.
//
// Sie einfach ungeprüft zu lassen hieße: eine Behauptung darüber, was im
// Ernstfall geschieht, und im Ernstfall stellt sie sich als falsch heraus, weil
// dort niemand nachgesehen hat. Also gibt es hier einen Treiber, der genau an
// diesen beiden Stellen stolpert.
package storage

import (
	"database/sql"
	"errors"
)

var errGestolpert = errors.New("der Stolpertreiber ist planmäßig gescheitert")

// stolperstelle sagt, wo der Treiber scheitern soll.
type stolperstelle int

const (
	beimCommit stolperstelle = iota
	beimLesen
	beimUmwandeln

	// beimFragen lässt die Abfrage selbst scheitern, nicht erst das Lesen.
	beimFragen

	// **Die drei Abfragen des Paketkatalogs** (#1325). Jede steht für einen
	// Fehlerweg, den ohne sie kein Lauf beträte: der Schreibzähler, die
	// Abhängigkeiten und die Kennung einer Zeile, an der sich nichts geändert
	// hat. Sie unterscheiden sich am **Präfix** der Abfrage, weil sie in
	// derselben Transaktion liegen und keine Reihenfolge sie trennt.
	beimSchreibzaehler
	beimZweitenSchreibzaehler
	beimBedarfslesen
	beimBedarfsscan
	beimBedarfstyp

	// **Wenn `RETURNING` nichts gibt**, weil sich nichts geändert hat: dann
	// liest der Katalog die Kennung nach — und dieser Weg lässt das scheitern.
	beimKennungslesen

	// Die beiden folgenden gehören zum Kartenlöschen (`karteloeschen.go`).
	// Sie beantworten seine zwei Abfragen und scheitern erst danach — sonst
	// käme der Lauf nie bis zu der Stelle, um die es geht
	// (#621).
	beimKarteschreiben
	beimKartecommit
	beimKarteverlauf
	beimKartefeldscan
	beimKartefeldende
	beimKartefeldabfrage

	// Die vier folgenden gehören zum Kartenkopieren (`kartekopieren.go`).
	// Es fragt eine Abfrage mehr als das Löschen — das Set der Karte — und
	// schreibt zwei mehr: die Karte selbst und jeden ihrer Feldwerte
	// (#625).
	beimKartesetid
	beimKarteeinreihen
	beimKartefeldeinsetzen
	beimKartekuerzen

	// Die vier folgenden gehören zum Umordnen (`kartentauschen.go`). Es fragt
	// **eine** Abfrage über beide Karten und schreibt dann Verlauf, Parken und
	// zwei Stellen
	// (#628).
	beimStellenscan
	beimStellenende
	beimStellenverlauf
	beimStellenparken
	beimStellensetzen

	// beimSchreiben lässt `Exec` scheitern und gibt auf Abfragen brauchbare
	// Kennungen zurück — sonst käme eine Massenänderung nie bis zum Schreiben,
	// weil ihre Vorprüfung alle Karten für fehlend hielte.
	beimSchreiben

	// Die drei folgenden gehören zur Feldänderung (`feldaenderung.go`). Sie
	// fragt **eine** Zeile mit **drei** Spalten ab — Kennung, alter Wert und
	// alte Stelle (#160) —, und die bekommt sie von `vorwertzeile`. Ohne diese Stellen bliebe
	// jeder Weg dahinter ungeprüft: das Ändern, das Bestätigen und das
	// Zurückrollen.
	beimFeldschreiben
	beimFeldcommit
	beimZurueckrollen

	// Die beiden folgenden gehören zum Verlauf (`verlaufbewegen.go`). Er fragt
	// **sechs** Spalten ab — Kennung, Karte, Handlung, Hinweg, Rückweg, Zeiger —
	// und geht danach zwei Schreibwege. Ohne diese Stellen bliebe alles hinter
	// der Abfrage ungeprüft.
	beimVerlaufschreiben
	beimVerlaufcommit

	// **Beim zweiten Schreiben.** Der Verlauf schreibt zweimal in einer
	// Transaktion — den Weg und den Zeiger, das Kürzen und den Eintrag. Wer
	// beim ersten stolpert, sieht den zweiten Weg nie.
	beimZweitenSchreiben

	// **Beim dritten Schreiben.** Eine Feldänderung schreibt den Wert, dann den
	// Verlaufseintrag, dann die Kürzung. Wer beim zweiten stolpert, sieht die
	// Kürzung nie.
	beimVerlaufseintrag
	beimVerlaufkuerzen

	// **Beim Lesen des neuen Werts.** Der Verlauf fragt ihn nach der Änderung
	// ab; scheitert diese Abfrage, ist der Eintrag nicht zu bilden.
	beimVerlaufslesen

	// Die beiden folgenden gehören zur Massenänderung (`massenschreiber.go`).
	// Sie liest je Karte den Vorwert und schreibt danach ihren Verlaufseintrag —
	// zwei Wege, die eine funktionierende Datenbank nicht hergibt, und ohne die
	// eine Massenänderung ohne Rückweg dastünde, ohne dass es jemand merkt.
	beimMassenvorwert
	beimMassenverlauf

	// **Beim Bestätigen der Massenänderung mit Erwartung** (U6, #1417). Sie
	// prüft die Kennungen, liest je Karte den Stand, schreibt — und erst der
	// Abschluss scheitert. Ohne diese Stelle bliebe genau der ungeprüft.
	beimErwartungcommit

	// **Beim Zählen der betroffenen Zeilen.** `StilSetzen` (#41) macht aus
	// null betroffenen Zeilen ein `ErrNotFound`; scheitert schon die Zählung,
	// ist die Frage unbeantwortet — und das ist etwas anderes als „null".
	beimZaehlen

	// Die beiden folgenden gehören zur Setänderung (`setwertaenderung.go`,
	// #567).
	// Sie schreibt dreimal — die Angabe, den Verlaufseintrag, die Kürzung —
	// und liest zweimal. Die vorhandenen Stellen sind auf die Aufrufkette der
	// **Feldänderung** eingestellt und treffen hier daneben; ohne diese beiden
	// blieben das Kürzen und das Bestätigen ungeprüft.
	beimSetkuerzen
	beimSetcommit

	// Die beiden folgenden gehören zur Herkunftszeile (`feldaenderung.go`,
	// #1025).
	// Sie wird in derselben Transaktion geführt wie der Wert — angelegt, wenn
	// er von außen kam, gelöscht, wenn er von hier kam. Ohne diese beiden
	// bliebe der Fall ungeprüft, in dem die Zeile und der Wert auseinander-
	// laufen: genau der, den die Tabelle ausschließen soll.
	beimHerkunftSchreiben
	beimHerkunftLoeschen

	// **Beim Schreiben einer Stiloption.** Dieser Fall lieh sich bis #1025 die
	// Stelle des Verlaufskürzens — beide waren „der dritte Schreibvorgang",
	// und das war ihre einzige Gemeinsamkeit. Ein geliehener Zähler prüft, was
	// er zählt, und nicht, was er heißt.
	beimStiloption

	// **Beim Lesen der vorigen Herkunft** (#1036). Sie wird **vor** der
	// Änderung gelesen; danach hat `herkunftFuehren` sie überschrieben, und
	// der Verlaufseintrag könnte sie nicht mehr tragen.
	beimHerkunftLesen

	// **Beim Fortschreiben des Verlaufszeigers** (#1036). Dieser Fall lieh
	// sich bis dahin `beimZweitenSchreiben` — „der zweite Schreibvorgang"
	// war der Zeiger. Seit die Rücknahme auch die Herkunft führt, sind es
	// drei, und der geliehene Zähler zeigte auf die falsche Anweisung. Ein
	// Zähler prüft, was er zählt, und nicht, was er heißt.
	beimVerlaufzeiger

	// **Der Weg, auf dem eine Herkunft zurückkommt** (#1036). Gelesen wird
	// sie über dieselbe Fassung wie beim Leser — deren Wege deckt der Test
	// dort ab; geschrieben wird sie nur hier.
	beimKartenherkunftSchreiben

	// **Beim Bestätigen einer Schlüsselwortänderung**
	// (#1077).
	// Sie schreibt in einer Transaktion: die Änderung, den Verlaufseintrag,
	// die Kürzung — und bestätigt danach. Ein `Commit`, das scheitert, obwohl
	// alles davor geglückt ist, gibt eine echte Datenbank nicht her, und ohne
	// diese Stelle stünde eine Schlüsselwortänderung ohne Rückweg da, ohne
	// dass es jemand merkt.
	beimWortcommit

	// **Die drei Stellen des Setlöschens**
	// (#1218).
	//
	// Es liest den Namen, löscht und bestätigt — drei Wege, an denen eine
	// echte Datenbank nicht scheitert, weil sie funktioniert. Und der Ausgang
	// ist derselbe wie bei den Schlüsselwörtern: was hier halb geschieht,
	// steht ohne Rückweg da. Ein Set hat keinen Verlauf.
	beimSetnamenLesen
	beimSetloeschen
	beimSetloeschcommit
	// **Die zwei Stellen des Arbeitsverlaufs** — Werkstück U11 von Phase 5.1
	// (#1422): ein Fehler mitten im Lesen der Schritte oder des Angehefteten.
	// Alles Übrige läuft gegen SQLite (`arbeitsverlauffehler_test.go`).
	beimArbeitsverlaufende

	// **Die drei Anweisungen des Bereichswechsels** hinter der ersten
	// (Befund 7 des Zweitblicks über Phase 5.1): eine echte Datenbank gibt
	// immer nur die erste zum Scheitern her.
	beimBereichmenge
	beimBereichloeschen
	beimBereichsetzen
	beimAngeheftetende
	// **Die fünf Stellen der Druckablage** — Werkstück A2 von Phase 5.3
	// (#1441).
	// Eine echte Datenbank gibt weder eine scheiternde Zeilennummer noch
	// einen scheiternden Abschluss noch einen Fehler mitten im Lesen her; die
	// übrigen Wege — Tabelle weg, Auslöser, geschlossener Bestand — laufen
	// gegen SQLite selbst (`druckefehler_test.go`).
	beimDruckID
	beimDruckcommit
	beimDruckende
	beimDruckscan
	beimKennungende
	beimKennungscan

	// **Die zwei Stellen der Modellproben** — Werkstück A9 von Phase 5.3
	// (#1448).
	// Dieselbe Begründung wie darüber: ein Fehler **mitten** im Lesen und
	// eine Zeile, die nicht in ihre Felder passt, entstehen im Treiber.
	beimProbenende
	beimProbenscan

	// **Die vier Stellen der Kontaktablage** — Werkstück N1 von Phase 5.4
	// (#1454).
	// Dieselbe Begründung wie darüber, hier für zwei Abfragen: die Kontakte
	// selbst und ihre Rechte.
	beimKontaktende
	beimKontaktscan
	beimRechteende
	beimRechtescan

	// **Und seit `T3` eine fünfte** (#1468): die **erste Begegnung** steht
	// als Zeitpunkt in der Datei, und eine, die keiner ist, gibt kein echter
	// Treiber her.
	beimKontaktfrist
)

// zweiterSchreibzaehler zählt die Schreibvorgänge des Stolpertreibers.
//
// **Als Paketvariable, weil ein Treiber keinen Zustand mitbekommt**: `Open`
// wird von `database/sql` gerufen, nicht von uns. Die Tests, die ihn benutzen,
// setzen ihn zurück.
var zweiterSchreibzaehler int

// verlaufOhneFeld lässt die Verlaufszeile sagen, dass es das Feld vorher nicht
// gab — dann geht `wegGehen` den Löschweg statt den Schreibweg.
var verlaufOhneFeld bool

func init() {
	sql.Register("stolper-karte-schreiben", stolpertreiber{stelle: beimKarteschreiben})
	sql.Register("stolper-karte-verlauf", stolpertreiber{stelle: beimKarteverlauf})
	sql.Register("stolper-karte-feldscan", stolpertreiber{stelle: beimKartefeldscan})
	sql.Register("stolper-karte-feldabfrage", stolpertreiber{stelle: beimKartefeldabfrage})
	sql.Register("stolper-karte-feldende", stolpertreiber{stelle: beimKartefeldende})
	sql.Register("stolper-karte-setid", stolpertreiber{stelle: beimKartesetid})
	sql.Register("stolper-karte-einreihen", stolpertreiber{stelle: beimKarteeinreihen})
	sql.Register("stolper-karte-feldeinsetzen", stolpertreiber{stelle: beimKartefeldeinsetzen})
	sql.Register("stolper-karte-kuerzen", stolpertreiber{stelle: beimKartekuerzen})
	sql.Register("stolper-stellen-scan", stolpertreiber{stelle: beimStellenscan})
	sql.Register("stolper-stellen-ende", stolpertreiber{stelle: beimStellenende})
	sql.Register("stolper-stellen-verlauf", stolpertreiber{stelle: beimStellenverlauf})
	sql.Register("stolper-stellen-parken", stolpertreiber{stelle: beimStellenparken})
	sql.Register("stolper-stellen-setzen", stolpertreiber{stelle: beimStellensetzen})
	sql.Register("stolper-karte-commit", stolpertreiber{stelle: beimKartecommit})
	sql.Register("stolper-commit", stolpertreiber{stelle: beimCommit})
	sql.Register("stolper-lesen", stolpertreiber{stelle: beimLesen})
	sql.Register("stolper-umwandeln", stolpertreiber{stelle: beimUmwandeln})
	sql.Register("stolper-fragen", stolpertreiber{stelle: beimFragen})
	sql.Register("stolper-schreiben", stolpertreiber{stelle: beimSchreiben})
	sql.Register("stolper-feld-schreiben", stolpertreiber{stelle: beimFeldschreiben})
	sql.Register("stolper-feld-commit", stolpertreiber{stelle: beimFeldcommit})
	sql.Register("stolper-zurueckrollen", stolpertreiber{stelle: beimZurueckrollen})
	sql.Register("stolper-verlauf-schreiben", stolpertreiber{stelle: beimVerlaufschreiben})
	sql.Register("stolper-verlauf-commit", stolpertreiber{stelle: beimVerlaufcommit})
	sql.Register("stolper-zweites-schreiben", stolpertreiber{stelle: beimZweitenSchreiben})
	sql.Register("stolper-verlauf-eintrag", stolpertreiber{stelle: beimVerlaufseintrag})
	sql.Register("stolper-verlauf-kuerzen", stolpertreiber{stelle: beimVerlaufkuerzen})
	sql.Register("stolper-verlauf-lesefehler", stolpertreiber{stelle: beimVerlaufslesen})
	sql.Register("stolper-massen-vorwert", stolpertreiber{stelle: beimMassenvorwert})
	sql.Register("stolper-massen-verlauf", stolpertreiber{stelle: beimMassenverlauf})
	sql.Register("stolper-erwartung-commit", stolpertreiber{stelle: beimErwartungcommit})
	sql.Register("stolper-zaehlen", stolpertreiber{stelle: beimZaehlen})
	sql.Register("stolper-set-kuerzen", stolpertreiber{stelle: beimSetkuerzen})
	sql.Register("stolper-set-commit", stolpertreiber{stelle: beimSetcommit})
	sql.Register("stolper-herkunft-schreiben", stolpertreiber{stelle: beimHerkunftSchreiben})
	sql.Register("stolper-herkunft-loeschen", stolpertreiber{stelle: beimHerkunftLoeschen})
	sql.Register("stolper-stiloption", stolpertreiber{stelle: beimStiloption})
	sql.Register("stolper-herkunft-lesen", stolpertreiber{stelle: beimHerkunftLesen})
	sql.Register("stolper-verlauf-zeiger", stolpertreiber{stelle: beimVerlaufzeiger})
	sql.Register("stolper-wort-commit", stolpertreiber{stelle: beimWortcommit})
	sql.Register("stolper-setname-lesen", stolpertreiber{stelle: beimSetnamenLesen})
	sql.Register("stolper-set-loeschen", stolpertreiber{stelle: beimSetloeschen})
	sql.Register("stolper-setloesch-commit", stolpertreiber{stelle: beimSetloeschcommit})
	sql.Register("stolper-kartenherkunft-schreiben",
		stolpertreiber{stelle: beimKartenherkunftSchreiben})

	// **Die vier Wege des Paketkatalogs** (#1325).
	sql.Register("stolper-schreibzaehler", stolpertreiber{stelle: beimSchreibzaehler})
	sql.Register("stolper-bedarf-lesen", stolpertreiber{stelle: beimBedarfslesen})
	sql.Register("stolper-bedarf-scan", stolpertreiber{stelle: beimBedarfsscan})
	sql.Register("stolper-kennung-lesen", stolpertreiber{stelle: beimKennungslesen})
	sql.Register("stolper-zweiter-zaehler",
		stolpertreiber{stelle: beimZweitenSchreibzaehler})
	sql.Register("stolper-bedarf-typ", stolpertreiber{stelle: beimBedarfstyp})
	sql.Register("stolper-arbeitsverlauf-ende", stolpertreiber{stelle: beimArbeitsverlaufende})
	sql.Register("stolper-bereich-menge", stolpertreiber{stelle: beimBereichmenge})
	sql.Register("stolper-bereich-loeschen", stolpertreiber{stelle: beimBereichloeschen})
	sql.Register("stolper-bereich-setzen", stolpertreiber{stelle: beimBereichsetzen})
	sql.Register("stolper-angeheftet-ende", stolpertreiber{stelle: beimAngeheftetende})
	sql.Register("stolper-druck-id", stolpertreiber{stelle: beimDruckID})
	sql.Register("stolper-druck-commit", stolpertreiber{stelle: beimDruckcommit})
	sql.Register("stolper-druck-ende", stolpertreiber{stelle: beimDruckende})
	sql.Register("stolper-druck-scan", stolpertreiber{stelle: beimDruckscan})
	sql.Register("stolper-kennung-ende", stolpertreiber{stelle: beimKennungende})
	sql.Register("stolper-kennung-scan", stolpertreiber{stelle: beimKennungscan})
	sql.Register("stolper-proben-ende", stolpertreiber{stelle: beimProbenende})
	sql.Register("stolper-proben-scan", stolpertreiber{stelle: beimProbenscan})
	sql.Register("stolper-kontakt-ende", stolpertreiber{stelle: beimKontaktende})
	sql.Register("stolper-kontakt-scan", stolpertreiber{stelle: beimKontaktscan})
	sql.Register("stolper-kontakt-frist",
		stolpertreiber{stelle: beimKontaktfrist})
	sql.Register("stolper-rechte-ende", stolpertreiber{stelle: beimRechteende})
	sql.Register("stolper-rechte-scan", stolpertreiber{stelle: beimRechtescan})
}
