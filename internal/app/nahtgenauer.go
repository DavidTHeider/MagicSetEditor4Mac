package app

// **Was die Hinübersetzung bei vier Codes mehr verlangt als die Klasse** —
// Befund 4 des Nachlaufs über Phase 1 von 0.9.0 (F03, #2342).
//
// [Fehler.AlsFehler] gab bis dahin nur die Klasse zurück. Für `errors.Is`
// genügte das; [Uebersetze] aber erkennt `drossel` an der Kontingentgrenze,
// `abgeschaltet` am Schalter oder Dienst und `netz` am Fall mit seinem Wirt —
// und machte aus allen dreien beim zweiten Hinweg `unbekannt`. Am Dienst kam
// das als 500 an. **Jetzt baut der Rückweg den genauen Fehler aus Schlüssel
// und Werten**, und der zweite Hinweg gibt denselben Satz.
//
// **Und bei `nichtGefunden` die genaueren** (#2491): vier Ausgänge
// umschließen `ErrNotFound` mit einem eigenen Satz, und die Kommandozeile gab
// nach dem Rückweg mit `--json` den allgemeinen Schlüssel aus — „lade die
// Liste neu" für ein Gegenüber, das nicht horcht. **Wer einen fünften
// einführt, trägt ihn hier ein**; der Rundweglauf hält die vier.

import (
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// genauer gibt den Fehler hinter Code, Schlüssel und Werten — oder nil, wo
// die Klasse genügt.
func (f *Fehler) genauer() error {
	switch f.Code {
	case CodeDrossel:
		// **Ohne Zahl die kleinste**: `Sekunden` gibt nie null.
		sekunden, _ := strconv.Atoi(f.Werte["sekunden"]) //nolint:errcheck // siehe darüber

		return &domain.Kontingentgrenze{Warten: time.Duration(sekunden) * time.Second}
	case CodeAbgeschaltet:
		if f.Schluessel == domain.SchluesselDienstAus {
			return &domain.AbgeschalteterDienst{Dienst: domain.Dienst(f.Werte["dienst"])}
		}

		return &domain.Abschaltung{Schalter: f.Werte["schalter"]}
	case CodeNetz:
		return f.gegenstelle()
	case CodeNichtGefunden:
		return nichtGefundenGenauer[f.Schluessel]
	}

	return nil
}

// nichtGefundenGenauer nennt die Ausgänge, die [domain.ErrNotFound] mit einem
// eigenen Schlüssel umschließen — dieselben wie in [Uebersetze].
var nichtGefundenGenauer = map[domain.Schluessel]error{
	domain.SchluesselKeinSetMitStil:      domain.ErrKeinSetMitStil,
	domain.SchluesselOhnePeeridentitaet:  domain.ErrOhnePeeridentitaet,
	domain.SchluesselPeerNichtErreichbar: domain.ErrPeerNichtErreichbar,
	domain.SchluesselPeerOhneAdresse:     domain.ErrPeerOhneAdresse,
}

// gegenstelle baut den Fall einer Gegenstelle aus seinem Schlüssel — ohne
// bekannten Schlüssel bleibt es bei der Klasse.
func (f *Fehler) gegenstelle() error {
	wirt := f.Werte["wirt"]

	if f.Schluessel == domain.SchluesselGedrosseltBis {
		uhrzeit, err := time.ParseInLocation("15:04", f.Werte["uhrzeit"], time.Local)
		if err != nil {
			return domain.NeuerGegenstellenfehler(domain.ErrGedrosselt, nil, wirt, 0)
		}

		jahr, monat, tag := time.Now().Date()

		return domain.NeueDrosselung(nil, wirt,
			time.Date(jahr, monat, tag, uhrzeit.Hour(), uhrzeit.Minute(), 0, 0, time.Local))
	}

	for _, fall := range gegenstellenfaelle {
		if fall.schluessel != f.Schluessel {
			continue
		}

		sekunden, _ := strconv.ParseFloat(f.Werte["sekunden"], 64) //nolint:errcheck // ohne Zahl keine Frist

		return domain.NeuerGegenstellenfehler(fall.klasse, nil, wirt, time.Duration(sekunden*float64(time.Second)))
	}

	return nil
}
