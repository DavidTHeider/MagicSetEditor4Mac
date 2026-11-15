// **Was beim Start geschah** —
// #962.
//
// # Warum es diese Bindung gibt
//
// `main.lauf` kehrte bei einem Fehler aus `Starten` **vor** `wails.Run`
// zurück. Es gab dann kein Fenster, keine Bindung und keinen Knopf — der
// Rückweg, den #940
// gebaut hat, war im gescheiterten Fall **nicht erreichbar**. Die Abnahme von
// Phase S lautete: „Ein absichtlich abbrechendes Upgrade lässt sich **ohne
// Handarbeit** auf den alten Stand zurückführen" — und der Test dazu fuhr den
// Weg über das `App`-Objekt unmittelbar. Das belegt, dass das Objekt
// wiederverwendbar ist, nicht dass jemand den Weg gehen kann.
//
// **Eine Anwendung, die sich bei einem Datenbankfehler wortlos beendet, lässt
// genau die Menschen allein, für die es die Sicherung gibt.**
//
// Jetzt öffnet das Fenster auch dann, und diese Bindung sagt ihm, was es
// zeigen soll: den Grund im Klartext und den Weg zur Sicherung — über
// **dieselben** Bindungen wie sonst, nicht über eine zweite Fassung derselben
// Frage.

package app

import (
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Startbefund sagt, ob der Bestand offen ist.
type Startbefund struct {
	Offen bool `json:"offen"`

	// Grund ist die Meldung des gescheiterten Starts, oder leer.
	//
	// **Im Klartext und nicht als Code.** Wer hier steht, hat keine
	// Oberfläche, die etwas übersetzen könnte — der Fehler kommt aus dem
	// Öffnen der Datenbank und ist der einzige Anhaltspunkt.
	Grund string `json:"grund"`

	// Fassung ist die laufende Fassung dieser Anwendung (#1168).
	//
	// **Sie steht hier und nicht in einer eigenen Bindung.** Der Startbefund
	// wird beim Aufbau des Fensters **einmal** gelesen, und die Fassung ändert
	// sich zur Laufzeit nicht — eine zweite Bindung wäre ein zweiter Weg zu
	// einer Antwort, die schon unterwegs ist.
	//
	// **Und sie kommt aus Go und nicht aus `package.json`.** Beide sollen
	// dieselbe Zahl tragen, und der Freigabelauf hält sie zusammen
	// (`versionen_pruefen`); die maßgebliche steht in `fassung.go`.
	Fassung string `json:"fassung"`

	// Schluessel ordnet den Grund einer Klasse zu — oder ist leer.
	//
	// **Weil der Grund allein nicht reicht** (Werkstück 3 von Phase 4.13,
	// #1352).
	// Am 2026-09-07 an der laufenden Anwendung gemessen: eine Datei, die kein
	// Bestand ist, gab „journal_mode: sqlite3: file is not a database" — ein
	// roher Treibertext im Fenster, den §8 verbietet. Und der Bestand aus der
	// Zukunft gab einen vollständigen deutschen Satz, der nicht sagt, was zu
	// tun ist.
	//
	// **Der Grund bleibt trotzdem stehen.** Er ist die Einzelheit unter dem
	// Satz — wer auf dieser Seite steht, sucht einen Fehler, und dann ist der
	// rohe Text die einzige Spur. Das ist dieselbe Trennung wie bei
	// [domain.MitGrund], nur zwischen zwei Feldern statt zwischen Meldung und
	// Protokoll.
	Schluessel string `json:"schluessel"`

	// Rueckspielrest sagt, dass ein abgebrochenes Zurückspielen beim Start
	// nicht abgeschlossen werden konnte — oder ist leer.
	//
	// **Im Klartext und nicht als Schlüssel**, wie [Startbefund.Grund] und aus
	// demselben Grund: der Satz nennt einen Pfad und einen Systemfehler, und
	// beides steht in keinem Katalog.
	//
	// **Warum er hier steht und nicht nur im Protokoll**: die Anwendung läuft
	// nach einem solchen Rest mit der **neuen** Datenbank und den **alten**
	// Bildern weiter. Das ist ein Zustand, über den jemand Bescheid wissen
	// muss — angestrichen vom zweiten Zweitblick über die Freigabe 0.3.0
	// (#1170,
	// Befund 2).
	Rueckspielrest string `json:"rueckspielrest"`

	// Rueckweg nennt die Sicherung, mit der diese Fassung wieder arbeiten
	// kann — oder ist `null`.
	//
	// **Weil der Rat sonst ins Leere zeigt** (Werkstück 5 derselben Phase):
	// `bestandNeuereFassung` endet mit „oder spiele eine Sicherung von vorher
	// zurück", und neben dem Bestand liegen bis zu dreißig davon. Welche
	// aufgeht, weiß nur, wer sie öffnet — und das tut `wegzurueck.go`.
	//
	// **Ein Zeiger und kein leerer Verbund**: „es liegt keine da" ist eine
	// eigene Auskunft und soll nicht wie eine Sicherung ohne Namen aussehen.
	Rueckweg *RueckwegDTO `json:"rueckweg"`
}

// Startbefund gibt den Befund des letzten Startversuchs.
func (a *App) Startbefund() Startbefund {
	if a.startfehler != nil {
		return Startbefund{
			Grund:          a.startfehler.Error(),
			Schluessel:     string(startklasse(a.startfehler)),
			Fassung:        Fassung,
			Rueckspielrest: a.rueckspielrest,
			Rueckweg:       a.rueckweg,
		}
	}

	return Startbefund{
		Offen:          a.db != nil,
		Fassung:        Fassung,
		Rueckspielrest: a.rueckspielrest,
	}
}

// startklasse ordnet einen Startfehler einer Klasse zu.
//
// # Warum hier und nicht in `Uebersetze`
//
// **Weil der Start keine Bindungsantwort gibt.** `Uebersetze` macht aus einem
// `error` einen [Fehler] mit Code und Schlüssel; hier gibt es keinen Ruf, den
// man ablehnen könnte — der Start ist gelaufen, und das Fenster fragt danach.
//
// # Die Reihenfolge ist die Aussage
//
// **Der fehlende Schlüssel zuerst**: er wird auf dem Weg nach oben in
// „Datenbank öffnen: …" eingewickelt, und die Zweige darunter erkennen ihn
// sonst nicht mehr — dieselbe Begründung wie in `Uebersetze`.
//
// **Und die neuere Fassung vor dem Unbrauchbaren**: ein Bestand aus der
// Zukunft **ist** ein gültiger Bestand, nur einer, den diese Fassung nicht
// öffnet. Wer beides zusammenlegte, riete zum Ersetzen einer Datei, die in
// Ordnung ist.
//
// **Erkannt an [domain.NeuereFassung] und nicht an `storage.ErrMigrationen`.**
// Der Grund ist die Schichtenregel (§2): `internal/app` kennt `storage` nur in
// den Composition Roots, und diese Datei ist keiner. Der Verbund aus `domain`
// trägt ohnehin mehr — die Namen der unbekannten Migrationen und den Stand
// dieser Fassung —, und `fehlersicherung.go` liest ihn schon so.
func startklasse(fehler error) domain.Schluessel {
	switch {
	case errors.Is(fehler, domain.ErrSchluesselFehlt):
		return domain.SchluesselSchluesselFehlt

	case istNeuereFassung(fehler):
		return domain.SchluesselBestandNeuereFassung

	// **Und die ältere ebenso** (Vermerk 0189): die Datei ist in Ordnung, und
	// der Rat führt über 0.7.0 statt zum Ersetzen.
	case istAeltereFassung(fehler):
		return domain.SchluesselBestandAeltereFassung
	}

	return startlage(fehler)
}

// startlage ordnet die Lagen ein, in denen die Datei **in Ordnung** sein
// kann — F05 (#2344),
// Befunde A-05 und A-21.
//
// **Unbrauchbar ist nur, was der Speicher als beschädigt erkannt hat.** Hier
// stand bis dahin „alles Übrige ist unbrauchbar", und der Rat dazu lautet
// „spiele eine Sicherung zurück": für einen belegten oder vollen Bestand der
// Rat, der Daten kostet. Was niemand erkannt hat, heißt jetzt ungeklärt.
//
// **Platz und Zugriff auch aus dem Dateisystem**: das Anlegen des Ordners
// scheitert mit `ENOSPC` oder `EACCES`, bevor SQLite überhaupt fragt.
func startlage(fehler error) domain.Schluessel {
	switch {
	// **Kaum Platz zuerst** (#2407): es umschließt die Lage, die es erzeugte —
	// einen Fehler des Datenträgers, einen gesperrten oder nur lesbaren Ort.
	case errors.Is(fehler, domain.ErrKaumPlatz):
		return domain.SchluesselKaumPlatz
	case errors.Is(fehler, domain.ErrBelegt):
		return domain.SchluesselBelegt
	case errors.Is(domain.Platzmangel(fehler), domain.ErrVoll):
		return domain.SchluesselVoll
	case errors.Is(fehler, domain.ErrNurLesend):
		return domain.SchluesselBestandNurLesbar
	case errors.Is(domain.Zugriffsmangel(fehler), domain.ErrGesperrt):
		return domain.SchluesselGesperrt
	case errors.Is(fehler, domain.ErrEA):
		return domain.SchluesselBestandEA
	case errors.Is(fehler, domain.ErrBeschaedigt):
		return domain.SchluesselBestandUnbrauchbar
	}

	return domain.SchluesselStartUngeklaert
}
