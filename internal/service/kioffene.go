package service

// **Die offenen Aufträge des Auftragswerks** — Werkstück A6 von Phase 5.3
// (#1445),
// getrennt für A7
// (#1446).
//
// # Warum eine eigene Datei
//
// **Weil `kiauftragswerk.go` mit ihr über die Zeilengrenze aus §2 ging.** Der
// Schnitt folgt der Frage: dort steht, **was** ein Auftrag durchläuft, hier,
// **wie lange** er dabei liegt — die Karte, ihr Schloss und die Übergabe an
// ein neu gebautes Werk.

import (
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kibuch hält die offenen Aufträge — **geteilt** zwischen den Werken einer
// Naht, nicht je Werk eines.
//
// # Warum geteilt und nicht kopiert
//
// **Weil zwischen „Werk bauen" und „Auftrag anlegen" eine Lücke liegt.** Die
// Naht baut das Werk, gibt es zurück, und **danach** legt der Aufrufer den
// Auftrag an. Wer in dieser Lücke ein zweites Werk baut, kopierte eine Karte,
// in der der Auftrag noch nicht steht — er landete im ersten Werk, und die
// Übernahme suchte ihn im zweiten. „Den Auftrag gibt es nicht", obwohl er
// eine Sekunde alt ist.
//
// Angestrichen vom Zweitblick über Phase 5.3, Befund 1. Ein Schloss allein
// hätte das Datenrennen behoben und den Verlust stehen gelassen: der
// gefährliche Zustand ist nicht die gleichzeitige Karte, sondern die
// **kopierte**.
type kibuch struct {
	schloss sync.Mutex
	offen   map[string]*kiLauf
}

// neuesBuch gibt ein leeres Auftragsbuch.
func neuesBuch() *kibuch {
	return &kibuch{offen: map[string]*kiLauf{}}
}

// anlegen trägt einen frischen Auftrag ein.
func (b *kibuch) anlegen(kennung string, lauf *kiLauf) {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	b.offen[kennung] = lauf
}

// lies gibt den offenen Auftrag dieser Kennung.
func (b *kibuch) lies(kennung string) (*kiLauf, bool) {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	lauf, gibts := b.offen[kennung]

	return lauf, gibts
}

// vermerke hängt den geprüften Vorschlag an seinen Lauf.
func (b *kibuch) vermerke(lauf *kiLauf, vorschlag *domain.KIVorschlag) {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	lauf.vorschlag = vorschlag
}

// verbrauche nimmt Kennungen aus dem Buch — sie sind eingelöst.
func (b *kibuch) verbrauche(kennungen ...string) {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	for _, kennung := range kennungen {
		delete(b.offen, kennung)
	}
}

// MitOffenen übernimmt die offenen Aufträge eines Vorgängers.
//
// # Wofür es das gibt
//
// **Weil der Katalog bei jedem Ruf neu entsteht und die Aufträge nicht.** Die
// Naht baut das Werk neu, sobald sich die Laufzeit geändert haben könnte; die
// Vorschläge, die jemand vor einer Minute angefordert hat, sollen das
// überleben. Ohne diesen Weg wäre jeder Vorschlag beim Übernehmen „gibt es
// nicht".
//
// **Ein `nil` ist erlaubt** und heißt: es gab keinen Vorgänger.
//
// **Es übernimmt das Buch und kopiert es nicht** — warum, steht bei [kibuch].
func (w *KIAuftragswerk) MitOffenen(vorher *KIAuftragswerk) *KIAuftragswerk {
	if vorher == nil {
		return w
	}

	w.buch = vorher.buch

	return w
}

// lauf gibt den offenen Auftrag dieser Kennung.
func (w *KIAuftragswerk) lauf(kennung string) (*kiLauf, bool) {
	return w.buch.lies(kennung)
}

// ohneBeanstandete gibt die Felder, zu denen kein Befund steht.
func ohneBeanstandete(
	felder map[string][]string, befunde []domain.Befund,
) map[string][]string {
	beanstandet := map[string]bool{}
	for _, befund := range befunde {
		beanstandet[befund.Datei] = true
	}

	gut := map[string][]string{}

	for name, werte := range felder {
		if !beanstandet[name] {
			gut[name] = werte
		}
	}

	return gut
}
