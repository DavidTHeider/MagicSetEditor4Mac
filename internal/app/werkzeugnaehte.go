// Die Nähte der **Werkzeugseiten** — Schlüsselwörter und Konsole
// (#1062,
// #1070).
//
// **Eine eigene Datei und trotzdem Composition Root.** `app.go` stand an der
// Zeilengrenze aus §2; der Schnitt folgt der Frage: dort **die Anwendung**,
// hier **die beiden Werkzeuge**, die beide dasselbe Zeichenwerk brauchen.

package app

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// werkzeugnaehte verdrahtet die Schlüsselwörter und die Skriptkonsole.
//
// **Beide über dasselbe Zeichenwerk.** Die Schlüsselwörter des **Spiels**
// kommen aus dem Spielpaket, und die Umgebung eines Ausdrucks ist dieselbe, in
// der eine Karte gezeichnet wird — eine zweite liefe auseinander.
func werkzeugnaehte(a *App, db *sql.DB, leser *storage.SetLeser, werk *Zeichenwerk) {
	// **Und das Zufallspack** (#1072) — es braucht denselben Planbauer, weil
	// der Filter jeder Packart ein Ausdruck über `card` ist.
	a.pack = service.NeuerPackbereich(leser, leser, werk.Planbauer).
		MitArbeitern(problemarbeiter())

	// **Die Skriptkonsole** über denselben Planbauer, in dem auch gezeichnet
	// wird: eine zweite Umgebung beantwortete Fragen über ein Programm, das es
	// nicht gibt.
	a.skriptkonsole = &Skriptnaht{
		konsole: service.NeueSkriptkonsole(leser, werk.Planbauer),
	}

	a.schluesselwoerter = service.NeueSchluesselwortpflege(
		SchluesselwortnahtUeber(db), werk.Stile)

	// **Die Paketwerkstatt** (#1164) — sie hängt am Paketverzeichnis und
	// nicht am Bestand.
	//
	// **Ein Fehler hier hält den Start nicht auf.** Ohne Zuhause gibt es kein
	// Paketverzeichnis; die drei Bindungen sagen es dann selbst, statt dass
	// das ganze Fenster leer bleibt (§8). Die Naht bleibt in diesem Fall
	// `nil`, und `MitPaketdateien` wird nicht gerufen — genau der Zweig, den
	// `keinePaketwerkstatt` meldet.
	if werkstatt, fehler := WerkstattnahtUeber(); fehler == nil {
		a.MitPaketdateien(werkstatt)
	}

	// **Und der Plattformöffner** (T8, #1473) — derselbe, den die Anmeldung
	// nimmt.
	//
	// **Hier gesetzt und nicht als Rückfall in der Naht**: ein Rückfall wäre
	// ein Zweig, den im Betrieb nie jemand betritt, und er sähe aus wie eine
	// geprüfte Zusage (§10). Wer ihn austauschen will, ruft
	// [App.MitOeffner].
	a.MitOeffner(browserOeffnen)

	// **Und der Vorlagenzeichner** (#1164) — er zeichnet eine Karte mit einer
	// Vorlage, die noch niemand gespeichert hat.
	//
	// **Über dieselbe Datenbank und dasselbe Paketverzeichnis.** Ein Fehler
	// hält den Start nicht auf; die Bindung sagt es dann selbst (§8).
	if zeichner, fehler := VorlagennahtUeber(db); fehler == nil {
		a.MitVorlagenzeichner(zeichner)
	}

	// **Und die Kartenskripte** (#1337). Sie hängen am Spielpaket und damit am
	// Paketverzeichnis; ein Fehler hält den Start nicht auf, und die Bindung
	// sagt es dann selbst (§8).
	if skripte, fehler := KartenskripteUeber(db); fehler == nil {
		a.MitKartenskripten(skripte)
	}

	// **Und die Auszählungsdimensionen** (#1338) — aus demselben Grund und
	// über denselben Weg.
	if zaehlung, fehler := AuszaehlungUeber(db); fehler == nil {
		a.MitAuszaehlung(zaehlung)
	}

	// **Und die Autoersetzungspflege** (#1341) — aus demselben Grund und über
	// denselben Weg.
	if ersetzung, fehler := AutoersetzungspflegeUeber(db); fehler == nil {
		a.MitAutoersetzungspflege(ersetzung)

		// **Und dieselbe Naht ersetzt beim Tippen** (#1341). Ohne diese Zeile
		// griffen weiter nur die Regeln des Pakets, und eine eigene wäre eine
		// Zeile im Bestand, die nichts tut. `start.go` setzt vorher den
		// Stilwähler ein; hier gewinnt der Weg, der beides kennt.
		a.MitAutoersetzung(ersetzung)
	}
}
