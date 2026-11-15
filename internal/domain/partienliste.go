package domain

// **Die Liste der eigenen Partien** —
// #1600.
//
// # Wofür es sie gibt
//
// **Weil eine Partie im Bestand liegt und im Fenster nur im Arbeitsspeicher.**
// Bis zum 2026-09-17 lebte sie allein im Zustand von `usePartie`; jeder
// Neustart, jeder Bestandswechsel und jeder **Profilwechsel** warf sie weg
// (`App.tsx` trägt `key={profilstand}`, und §7 Regel 1 verlangt genau das).
// Die Partie stand weiter im Bestand, mit allen Karten, und es führte kein
// Weg dorthin.
//
// **Für den Gast gab es das längst** (G5,
// #1462) —
// für den **Wirt** nicht.
//
// # Warum eine eigene Gestalt und nicht die Partie selbst
//
// **Weil eine Liste keine Sicht ist.** Wer sie liest, will wissen, **ob** es
// eine Partie gibt und welche; was darin liegt, sagt [Partie.Sicht], und die
// rechnet je Sitz. Eine Liste, die Objekte mitgäbe, umginge die Projektion,
// statt sie zu benutzen — und die Bibliothek ist auch für ihren Eigner
// verdeckt (G1).

// Partiezeile ist eine Partie, wie eine Liste sie zeigt.
type Partiezeile struct {
	ID      string
	Fassung int64

	Spiel string
	Stand Partiestand

	// Sitze sind die Plätze am Tisch — Kennung und Name.
	//
	// **Beides, und beides wird gebraucht**: der Name sagt einem Menschen,
	// welche Partie das ist („David und Gegenüber" statt „sitz-1 und
	// sitz-2"), die Kennung sagt dem Kern, **wessen Sicht** gerechnet werden
	// soll, wenn jemand wieder einsteigt.
	//
	// **Eine Sitzkennung ist kein Geheimnis** — sie steht in jeder
	// [Partiesicht], und die Projektion hängt nicht an ihrer Verborgenheit,
	// sondern daran, dass sie je Sitz rechnet (G1).
	Sitze []Zeilensitz

	// Objekte ist, wie viele Dinge in dieser Partie liegen.
	//
	// **Eine Zahl und keine Liste.** Sie sagt einem Menschen, wie weit die
	// Partie ist, und sie verrät nichts über eine einzelne Karte.
	Objekte int

	// Zug ist der wievielte Zug — null heißt: es geht nicht zugweise zu
	// (G9, #1592).
	Zug int

	// EigenerSitz ist die Kennung des Sitzes, der an das fragende Profil
	// gebunden ist — **leer heißt: keiner**, das Profil sieht dieser Partie
	// zu oder hat nie an ihr gesessen (UXR-F16,
	// #2031).
	//
	// **[Partie.Zeile] lässt ihn leer**: sie kennt kein Profil. Befüllt wird
	// er dort, wo es bekannt ist.
	EigenerSitz string

	// Dran sagt, ob die Partie gerade auf den eigenen Sitz wartet.
	//
	// **Falsch ohne eigenen Sitz** — und das ist keine Verneinung, sondern
	// eine fehlende Frage: wer nicht am Tisch sitzt, ist nie „dran“.
	Dran bool
}

// Zeilensitz ist ein Platz, wie eine Liste ihn zeigt.
type Zeilensitz struct {
	Kennung string
	Name    string
}

// Zeile gibt diese Partie, wie eine Liste sie zeigt.
//
// **Eine leere Liste und kein `nil`** — der Satz gilt hier wie an jeder
// anderen Grenze (N1, N4): ein Aufrufer prüft eine Länge und nicht zwei
// Dinge.
func (p Partie) Zeile() Partiezeile {
	plaetze := make([]Zeilensitz, 0, len(p.Sitze))
	for _, sitz := range p.Sitze {
		plaetze = append(plaetze, Zeilensitz{
			Kennung: sitz.Kennung, Name: sitz.Name,
		})
	}

	return Partiezeile{
		ID:      p.ID,
		Fassung: p.Fassung,
		Spiel:   p.Spiel,
		Stand:   p.Stand,
		Sitze:   plaetze,
		Objekte: len(p.Objekte),
		Zug:     p.Zug.Nummer,
	}
}

// SitzeVonProfil gibt die Kennungen der Sitze, die an dieses Profil gebunden
// sind, in der Reihenfolge des Tisches.
//
// # Die Bindung und nichts sonst
//
// **Ein Sitz gehört einem Profil, wenn er an es gebunden ist** (G8,
// #1465).
// Wer die Partie angelegt hat, sitzt nicht dadurch am Wirt — das ist die
// Sicht, die [Partie.SichtFuerProfil] ihm **zeigt**, und keine Auskunft
// darüber, wer er am Tisch ist. Zwei Menschen am Küchentisch tippen zwei
// Namen und binden nichts; für keinen von beiden ist die Partie „dran“.
//
// **Mehrere, weil das Spiel gegen sich selbst es so verlangt**: beide Sitze an
// dasselbe Profil gebunden ist der Normalfall dort (siehe [Sitz.ProfilID]).
//
// **Die Null sitzt nirgends.** Sie heißt „kein Profil“ und steht an jedem
// ungebundenen Sitz — wer mit ihr suchte, fände alle.
//
// **Eine leere Liste und kein `nil`**, wie überall an dieser Grenze.
func (p Partie) SitzeVonProfil(profilID int64) []string {
	sitzt := []string{}

	if profilID == 0 {
		return sitzt
	}

	for _, sitz := range p.Sitze {
		if sitz.ProfilID == profilID {
			sitzt = append(sitzt, sitz.Kennung)
		}
	}

	return sitzt
}
