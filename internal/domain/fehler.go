// Package domain hält die Entitäten, Wertetypen und Regeln der Fachlichkeit.
//
// Es importiert **nichts** aus dem Projekt und nichts aus `database/sql`
// (AGENTS.md §2); `internal/schichten` weist beides ab.
package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Die drei Fehlerarten, auf die alles abgebildet wird (AGENTS.md §8).
//
// `internal/storage` bildet SQLite-Fehler darauf ab, `internal/app` übersetzt
// sie in eine für das Frontend stabile Struktur. Geprüft wird mit `errors.Is`,
// nie per Zeichenkettenvergleich — ein Fehler, den man nur an seinem Text
// erkennt, überlebt die erste Umformulierung nicht.
var (
	// ErrValidation: die Eingabe ist nicht gültig.
	ErrValidation = errors.New("Validierung fehlgeschlagen")
	// ErrNotFound: der gesuchte Eintrag gibt es nicht.
	ErrNotFound = errors.New("nicht gefunden")
	// ErrConflict: die Änderung widerspricht dem gegenwärtigen Stand.
	ErrConflict = errors.New("Konflikt mit dem gegenwärtigen Stand")

	// ErrSicherung meldet, dass eine Datei nicht als Sicherung taugt
	// (#940).
	//
	// **Warum hier und nicht in `storage`**: `internal/app` muss sie
	// erkennen, um sie zu übersetzen (§8) — und darf `storage` außerhalb der
	// Composition Roots nicht kennen (§2). Dieselbe Richtung wie bei
	// [ErrNotFound]: der Unterbau bildet seine Fehler auf diese ab.
	ErrSicherung = errors.New("die Datei taugt nicht als Sicherung")

	// ErrBelegt meldet, dass jemand anders gerade am Bestand schreibt
	// (#886).
	//
	// **Nicht [ErrConflict], obwohl es nach Konflikt klingt.** Der heißt hier
	// „dein Stand ist überholt", und darauf **lädt** die Oberfläche nach.
	// Nachladen liefe in dieselbe Sperre. Der richtige Rat ist gleich und
	// nochmal, und dafür braucht es einen eigenen Fehler.
	//
	// **Er ist der einzige der fünf, der von allein vergeht.** Wer ihn sieht,
	// hat nichts falsch gemacht — das Fenster oder ein zweiter Lauf hält den
	// Bestand gerade, und in einem Augenblick nicht mehr.
	ErrBelegt = errors.New("der Bestand ist gerade belegt")

	// ErrNurLesend meldet, dass in diesem Profil nicht geschrieben werden darf
	// — Werkstück 16 von Phase 4.12
	// (#1351),
	// nach [0070](../entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md).
	//
	// **Nicht [ErrErlaubnis], obwohl beides von einer Rolle kommt.** Der heißt
	// „dein Nachweis reicht für **diesen Ruf** nicht" und steht an einer
	// einzelnen Handlung. Dieser heißt „**hier** schreibt niemand" — er gilt
	// für alles, solange der Zeiger auf diesem Profil steht, und was zu tun
	// ist, ist etwas anderes: das Profil verlassen oder um eine Rolle bitten.
	//
	// **Und nicht der rohe Treibertext.** SQLite meldet „attempt to write a
	// readonly database"; das erreicht die Oberfläche nie (§8) und sagt
	// niemandem, dass es an einer **Freigabe** liegt und nicht an der Datei.
	ErrNurLesend = errors.New("in diesem Profil darf nur gelesen werden")

	// ErrSchluesselFehlt meldet, dass ein verschlüsselter Bestand ohne seinen
	// Schlüssel dasteht — Werkstück 14 von Phase 4.12
	// (#1351),
	// nach [0072](../entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
	//
	// **Nicht [ErrSicherung] und nicht der rohe SQLite-Fehler**, und das ist
	// hier keine Feinheit, sondern der Unterschied zwischen zwei Ratschlägen,
	// von denen einer Daten kostet: SQLite meldet für einen verschlüsselten
	// Bestand ohne Schlüssel `SQLITE_NOTADB`, und das heißt in dieser
	// Anwendung seit
	// #886
	// „die Datei ist kein Bestand". Wer das liest, ersetzt sie — und die Datei
	// war in Ordnung.
	//
	// **Wiederherstellbar ist er nicht.** Der Schlüssel liegt im
	// Schlüsselbund; ist er dort fort, sind die Daten fort. Genau deshalb sagt
	// diese Meldung, was los ist, statt zum Ersetzen zu raten.
	ErrSchluesselFehlt = errors.New("der Schlüssel des Bestands fehlt")

	// ErrZweiterFaktor meldet, dass hier ein zweiter Faktor verlangt wird —
	// Werkstück 7 von Phase 4.12
	// (#1351).
	//
	// **Nicht [ErrValidation], obwohl er wie eine Abweisung aussieht.** Der
	// heißt „die Eingabe stimmt nicht" und lässt die Fläche dasselbe Formular
	// noch einmal zeigen. Dieser heißt **frag nach dem Code** — die Fläche
	// muss etwas **anderes** tun, und dafür muss sie es unterscheiden können.
	//
	// **Und deshalb ist er auch nicht [ErrConflict]**: es ist kein Streit über
	// einen Stand, sondern ein fehlender Schritt.
	ErrZweiterFaktor = errors.New("hier fehlt der zweite Faktor")

	// ErrErlaubnis meldet, dass ein **gültiger** Nachweis nicht weit genug
	// reicht — Werkstück 10 von Phase 4.12
	// (#1351).
	//
	// **Nicht [ErrValidation], und der Unterschied ist der ganze Punkt.** Der
	// heißt „dein Nachweis stimmt nicht" und darf nichts weiter verraten.
	// Dieser heißt „dein Nachweis stimmt, er darf das nur nicht" — und das
	// **soll** verraten werden: wer ein gültiges Token hält, ist kein
	// Fremder, und ohne diesen Satz suchte er den Fehler bei seiner Anfrage.
	//
	// **Er ist die einzige Stelle, an der eine Ablehnung mehr sagt als
	// „nein".** Deshalb steht er hier und nicht als Zeichenkettenvergleich in
	// einer Fläche: wer ihn setzt, hat entschieden, dass diese Auskunft
	// hinausgehen darf (§8).
	//
	// **Zwei lesen ihn.** Der REST-Zugang macht daraus einen `403`
	// (Werkstück 10), und `internal/app` einen [Fehlercode] `erlaubnis`
	// (Werkstück 11).
	ErrErlaubnis = errors.New("dieser Nachweis reicht nicht so weit")

	// ErrFremderNutzer meldet, dass jemand über einen **anderen** Menschen
	// schreiben wollte, den er nicht werden kann — ID-03
	// (#1702).
	//
	// **Eine Erlaubnisfrage und keine Rollenfrage**: es geht um keinen
	// Zugang zu einem Profil, sondern um einen Menschen. Der Satz zu
	// [ErrErlaubnis] — „wer dieses Profil verwaltet, kann sie dir geben" —
	// schickte jemanden zu einem Verwalter, den es hier nicht gibt. Der
	// nächste Schritt ist ein Nutzerwechsel. Siehe [SchluesselFremderNutzer].
	ErrFremderNutzer = fmt.Errorf(
		"%w: das darf nur dieser Nutzer selbst", ErrErlaubnis)

	// ErrNichtFreigegeben meldet, dass ein Gegenüber diesem Profil zum
	// Handeln nichts freigibt — eine Absage über die Leitung (P11, #2379).
	// Siehe [SchluesselNichtFreigegeben].
	ErrNichtFreigegeben = fmt.Errorf(
		"%w: das Gegenüber gibt nichts frei", ErrErlaubnis)

	// ErrKeinSetMitStil meldet, dass kein offenes Set einen Stil benutzt.
	//
	// **Ein eigener Ausgang und kein blankes [ErrNotFound]**: die
	// Vorlagenvorschau macht daran einen leeren Zustand fest, und der gilt nur
	// für diesen einen Fall. Siehe [SchluesselKeinSetMitStil].
	ErrKeinSetMitStil = fmt.Errorf("%w: kein offenes Set benutzt diesen Stil",
		ErrNotFound)

	// ErrOhnePeeridentitaet meldet, dass dieses Profil keine Identität hat
	// (N2, #1455).
	//
	// **Ein eigener Ausgang und kein blankes [ErrNotFound]**, und der Anlass
	// ist eine Sichtprüfung: wer erreichbar werden wollte, ohne eine
	// Identität zu haben, las „Der gesuchte Eintrag wurde nicht gefunden.
	// Vielleicht hat ihn jemand anderes gelöscht — lade die Liste neu." Das
	// ist der allgemeine Satz zu `nichtGefunden`, und hier war er falsch: es
	// ist nichts gelöscht worden, es gibt nur noch nichts. Siehe
	// [SchluesselOhnePeeridentitaet].
	ErrOhnePeeridentitaet = fmt.Errorf(
		"%w: dieses Profil hat keine brauchbare Identität", ErrNotFound)

	// ErrPeerNichtErreichbar meldet, dass an der Adresse eines Gegenübers
	// niemand annahm — im eigenen Netz wie über ein fremdes
	// (#2491).
	//
	// **Ein eigener Ausgang, aus demselben Grund wie der darüber**: unter
	// „Jetzt übertragen" stand der allgemeine Satz zu `nichtGefunden`, und
	// sein Rat — die Liste neu laden, jemand habe etwas gelöscht — half nicht.
	// **Die Klasse bleibt [ErrNotFound]**: der Horchposten liest sie als „da
	// ist niemand", und Kommandozeile, MCP und REST melden weiter
	// `nichtGefunden` (§16: erweiternd). Siehe [SchluesselPeerNichtErreichbar].
	ErrPeerNichtErreichbar = fmt.Errorf(
		"%w: das Gegenüber ist nicht zu erreichen", ErrNotFound)

	// ErrPeerOhneAdresse meldet, dass für ein Gegenüber keine Adresse
	// eingetragen ist — derselbe Anlass und dieselbe Klasse wie
	// [ErrPeerNichtErreichbar] (#2491), gefunden mit der Hüllenprobe
	// `p10handel09`: der Rat ist hier ein Feld am Kontakt.
	ErrPeerOhneAdresse = fmt.Errorf(
		"%w: für das Gegenüber ist keine Adresse eingetragen", ErrNotFound)

	// ErrOhneNetzerlaubnis meldet, dass die Suche im lokalen Netz nicht
	// hinausging (N2, #1455).
	//
	// **Der Anlass ist eine Sichtprüfung am 2026-09-14.** „Wer in der Nähe
	// ist" antwortete mit „Es ist ein unerwarteter Fehler aufgetreten", und
	// der Grund stand nur im Protokoll:
	//
	//	write udp4 0.0.0.0:52316->224.0.0.251:5353: sendto: no route to host
	//
	// Das ist die Multicastadresse von mDNS, und macOS gibt für sie
	// `EHOSTUNREACH` zurück, solange die Anwendung die Erlaubnis **Lokales
	// Netzwerk** nicht hat. Ein Mensch hat davon nichts erfahren — und die
	// Sache ist mit zwei Klicks erledigt.
	//
	// **Die zweite Lesart steht im Satz mit drin**, denn belegt ist sie
	// nicht: ein Netz ohne Route für Multicast gibt denselben Fehler. Wer
	// eine Ursache behauptet, die er nicht gemessen hat, schickt jemanden an
	// die falsche Stelle (§10). Siehe [SchluesselOhneNetzerlaubnis].
	ErrOhneNetzerlaubnis = fmt.Errorf(
		"%w: ins lokale Netz ging nichts hinaus", ErrValidation)

	// ErrOhneNetzroute meldet, dass es ins lokale Netz **keinen Weg** gibt —
	// `ENETUNREACH`, etwa ohne verbundene Schnittstelle (NET-02.NAT-03,
	// #1706).
	// **Getrennt von [ErrOhneNetzerlaubnis]**, weil der nächste Schritt ein
	// anderer ist: hier helfen keine Systemeinstellungen zur Erlaubnis.
	ErrOhneNetzroute = fmt.Errorf(
		"%w: für das lokale Netz gibt es keine Route", ErrValidation)

	// ErrSchluesselbundAblage meldet, dass sich ein Schlüssel nicht im
	// Schlüsselbund ablegen ließ (S13.NAT-03, #1706) — **ein anderer Fall
	// als ein fehlender** ([ErrOhnePeeridentitaet]): dort ist keiner da, hier
	// ging das Ablegen schief, etwa weil der Schlüsselbund gesperrt ist.
	ErrSchluesselbundAblage = fmt.Errorf(
		"%w: der Schlüssel ließ sich nicht im Schlüsselbund ablegen", ErrValidation)

	// ErrEinladungZuNeu meldet eine Einladung aus einer neueren Fassung
	// (N3, #1456).
	//
	// **Ein eigener Ausgang und kein blankes [ErrValidation]**, weil der
	// nächste Schritt ein anderer ist: hier ist nichts falsch eingegeben,
	// sondern die Gegenseite ist neuer. Wer das nicht sagt, schickt jemanden
	// auf die Suche nach einem Tippfehler. Siehe [SchluesselEinladungZuNeu].
	ErrEinladungZuNeu = fmt.Errorf(
		"%w: diese Einladung stammt aus einer neueren Fassung", ErrValidation)

	// ErrFernhorchenLaeuft meldet, dass schon eine Einladung offen ist
	// (N3, #1456).
	//
	// **Ein eigener Ausgang**, weil der nächste Schritt ein anderer ist als
	// bei einem gewöhnlichen Konflikt: es ist nichts zu ändern, sondern
	// erst zu beenden. Siehe [SchluesselFernhorchenLaeuft].
	ErrFernhorchenLaeuft = fmt.Errorf(
		"%w: dieses Profil ist über fremde Netze schon erreichbar", ErrConflict)

	// ErrAnmeldungAbgelaufen meldet eine Anmeldung, deren Rückruf nicht
	// binnen ihrer Frist kam (BK-07d, #2118). **Ein Ablauf und kein
	// Abbruch**: niemand hat aufgehört, und der Rat ist, neu zu beginnen.
	// Siehe [SchluesselAnmeldungAbgelaufen].
	ErrAnmeldungAbgelaufen = fmt.Errorf(
		"%w: die Anmeldung kam nicht rechtzeitig zurück", context.DeadlineExceeded)

	// ErrUmschlagFassung meldet einen Umschlag aus einer anderen Fassung
	// (N4, #1457).
	//
	// **Eine Abweichung wird abgewiesen und nicht heruntergehandelt** —
	// dieselbe Linie wie beim Handschlag (N1) und bei der Einladung (N3).
	// Siehe [SchluesselUmschlagFassung].
	ErrUmschlagFassung = fmt.Errorf(
		"%w: dieser Umschlag ist aus einer anderen Fassung", ErrValidation)

	// ErrUmschlagUnterschrift meldet eine Unterschrift, die nicht trägt
	// (N4, #1457).
	//
	// **Ein eigener Ausgang**, weil der Rat ein anderer ist: hier ist nichts
	// zu ändern und nichts nachzureichen — der Umschlag ist entweder
	// unterwegs verändert worden oder kommt von jemand anderem, als er sagt.
	// Siehe [SchluesselUmschlagUnterschrift].
	ErrUmschlagUnterschrift = fmt.Errorf(
		"%w: die Unterschrift des Umschlags trägt nicht", ErrValidation)

	// ErrUmschlagInhalt meldet einen Inhalt, der nicht zum Manifest passt
	// (N4, #1457).
	//
	// **Getrennt von der Unterschrift**, obwohl beides „verändert" heißt:
	// eine gebrochene Unterschrift sagt, dass das **Manifest** nicht mehr
	// stimmt; ein abweichender Inhalt sagt, dass ein **Objekt** nicht dazu
	// passt. Zwei verschiedene Stellen, an denen jemand nachsehen würde.
	ErrUmschlagInhalt = fmt.Errorf(
		"%w: ein Objekt passt nicht zum Manifest des Umschlags", ErrValidation)

	// ErrGedrosselt meldet, dass eine Gegenstelle zu viele Anfragen gemeldet
	// hat (#888).
	//
	// **Warum er hier steht und nicht im Transport.** Sein Rat ist eine
	// fachliche Auskunft: nicht „ändere deine Eingabe" und nicht „die
	// Verbindung ist weg", sondern „warte". Wer ihn nur in `internal/netz`
	// führte, könnte ihn in `internal/service` nicht unterscheiden — und dort
	// entscheidet sich, was ein Aufrufer zu sehen bekommt.
	//
	// **Der Anlass** (2026-08-24, Zweitblick über Phase V): der Rückfall auf
	// die Sammeldatei meldete bei einem 429 „Scryfall war nicht erreichbar".
	// Das ist falsch — Scryfall war bestens erreichbar. `internal/service`
	// durfte den Transport nicht kennen (§2) und konnte den Fall deshalb nicht
	// sehen.
	//
	// **Er ist der zweite der Reihe, der von allein vergeht** — wie
	// [ErrBelegt]. Wer ihn sieht, hat nichts falsch gemacht; er hat zu schnell
	// gefragt, und in dreißig Sekunden nicht mehr.
	ErrGedrosselt = errors.New("die Gegenstelle meldet zu viele Anfragen")
)

// NeuereFassung ist die Abweisung einer Sicherung aus einer neueren Fassung.
//
// **In ihren Teilen und nicht als Satz**: die Oberfläche soll sagen können,
// **welchen** Schemastand sie nicht kennt. Aus einem fertigen Satz bekäme sie
// den deutschen Wortlaut oder nichts.
//
// Die Namen sind Bezeichner und werden nicht übersetzt (§9); der Rahmen um sie
// herum schon.
type NeuereFassung struct {
	// Unbekannt sind die Schemaschritte, die diese Fassung nicht kennt —
	// sortiert, damit die Meldung bei jedem Lauf dieselbe ist.
	Unbekannt []string

	// Stand ist der letzte Schritt, den diese Fassung kennt.
	Stand string
}

// Error gibt den Satz, der ohne Katalog gelesen wird.
func (n *NeuereFassung) Error() string {
	return fmt.Sprintf("dieser Bestand führt %s, und diese Fassung kennt nur "+
		"bis %s — er stammt aus einer neueren Fassung von MagicSetEditor4Mac "+
		"und wird von dieser nicht geöffnet, damit kein halber Stand entsteht",
		strings.Join(n.Unbekannt, ", "), n.Stand)
}

// Unwrap ordnet sie den Sicherungsfehlern zu.
func (n *NeuereFassung) Unwrap() error { return ErrSicherung }

// FremdeArchivfassung ist eine Sicherung, deren **Archiv** nicht zu dieser
// Fassung passt (#1162).
//
// **Sie steht neben [NeuereFassung] und nicht in ihr.** Die andere sagt
// etwas über den **Bestand** — welche Schemaschritte er führt. Diese sagt
// etwas über die **Hülle**: welche Fassung sie geschrieben hat, und ob deren
// Gestalt zu dieser hier passt. Eine künftige Fassung kann das Schema
// unverändert lassen und das Archiv ändern.
type FremdeArchivfassung struct {
	// Dort ist die Fassung, die in der Sicherung steht — leer, wenn keine
	// darin steht.
	Dort string

	// Hier ist die Fassung dieser Anwendung.
	Hier string
}

// Error gibt den Satz, der ohne Katalog gelesen wird.
func (f *FremdeArchivfassung) Error() string {
	if f.Dort == "" {
		return fmt.Sprintf("diese Sicherung nennt keine Fassung — jede, die "+
			"MagicSetEditor4Mac schreibt, tut es, und deshalb wird sie nicht "+
			"zurückgespielt; diese Anwendung ist %s", f.Hier)
	}

	return fmt.Sprintf("diese Sicherung nennt die Fassung %q, und diese "+
		"Anwendung ist %s — sie wird nicht zurückgespielt, damit kein halber "+
		"Stand entsteht", f.Dort, f.Hier)
}

// Unwrap ordnet sie den Sicherungsfehlern zu.
func (f *FremdeArchivfassung) Unwrap() error { return ErrSicherung }
