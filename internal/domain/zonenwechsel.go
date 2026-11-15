package domain

// **Ein Zonenwechsel macht ein neues Objekt** — Werkstück M18 der
// Regelmaschine
// (#1619).
//
// # Die Regel
//
// **CR 400.7** (Fassung 2026-08-19): was von einer Zone in eine andere geht,
// ist ein **neues Objekt** ohne Erinnerung an sein voriges Dasein.
//
// # Und warum sie hier ohne Abstrich gilt
//
// **CR 400.7 nennt zwölf Ausnahmen** (400.7a–m), und **jede einzelne**
// spricht von Effekten, Fähigkeiten oder Auslösern: ein Effekt, der die
// Merkmale eines Permanentzaubers ändert (a), eine statische Fähigkeit, die
// eine gewährte Fähigkeit weiterreicht (b), ein Auslöser, der das neue
// Objekt wiederfindet (e, f), ein Effekt, der das Wirken erlaubt (g–i).
// **Keines davon gibt es in diesem Programm** — sie kommen mit `M8`
// (#1610)
// und `M9` (#1611).
//
// # Eine Stelle und nicht zwölf
//
// **Weil es zwölf Stellen gibt, an denen etwas die Zone wechselt** — vom
// Zerstören über das Wirken bis zum Aufräumschritt. Jede einzeln zu räumen
// hieße, die Regel zwölfmal zu haben, und die dreizehnte vergäße sie.

// NeuesObjekt wendet CR 400.7 auf ein Objekt an, das die Zone wechselt.
//
// **Exportiert, weil der Sandkasten dieselbe Regel braucht**: `objekt.schieben`
// legt eine Karte von Hand um, und es ist derselbe Zonenwechsel.
//
// # Was fällt
//
//   - **Marken** (CR 122.2 aus der anderen Richtung: sie hängen am Objekt,
//     und das Objekt ist ein anderes).
//   - **Getappt und verdeckt** — CR 110.5 nennt sie Zustände eines
//     **Permanents**, und außerhalb des Schlachtfelds gibt es keines.
//   - **Markierter Schaden** (CR 120.3), aus demselben Grund — und mit ihm
//     die Todesberührung (CR 704.5h fragt nach Schaden, und den gibt es
//     nicht mehr).
//   - **Was daran hing** — und was es **war**: beides ist eine Beziehung
//     zwischen zwei Objekten, und eines davon gibt es nicht mehr.
//   - **Die Seite** (CR 712.8a): außerhalb von Schlachtfeld und Stapel
//     gelten die Merkmale der Vorderseite.
//   - **Die Ziele** eines Zaubers (CR 115.1) — sie gelten dem Zauber, und
//     der ist mit dem Verrechnen vergangen.
//
// # Was bleibt
//
// **Kennung, Eigner, Art und Karte.** Die Kennung ist die Identität in
// **diesem Programm** und nicht die des Regelwerks: ohne sie ließe sich ein
// Objekt nicht mehr ansprechen, und der Ereignisstrom spräche von etwas,
// das niemand findet. **Das ist eine Abweichung, und sie steht hier**, weil
// CR 400.7 „keine Erinnerung" sagt.
//
// **Und der Beherrscher bleibt nicht** — CR 109.4: außerhalb von
// Schlachtfeld und Stapel hat ein Objekt keinen Beherrscher. Wer das Feld
// trotzdem füllen muss, füllt es mit dem Eigner; das entscheiden die
// Aufrufer, weil es beim Wirken anders läuft als beim Zerstören.
func NeuesObjekt(objekt Spielobjekt, zug int) Spielobjekt {
	// **Auch die Stelle** — sie ist die Lage in der **vorigen** Zone, und
	// wer sie mitnähme, läge in der neuen an einer Stelle, die ihm niemand
	// gegeben hat. Wo eine gilt, setzt sie der Aufrufer.
	objekt.Ordnung = 0

	// **Eine Kopie endet mit der Zone** — CR 707.4 mit CR 400.7: das
	// Objekt ist ein neues, und es kopiert nichts mehr. **Nur bei einer
	// Karte**: bei einem Spielstein **sind** die Felder seine Identität
	// (CR 111.1), und er hört ohnehin auf zu existieren, sobald er das
	// Schlachtfeld verlässt (CR 111.7).
	if objekt.Karte != "" {
		objekt.Felder = nil
	}

	objekt.Zaehler = nil
	objekt.Getappt = false
	objekt.Verdeckt = false
	objekt.Schaden = 0
	objekt.Todesberuehrt = false
	objekt.HaengtAn = ""
	objekt.Haufen = ""
	objekt.Seite = 0
	objekt.Ziele = nil

	// **Und die Wirkungen** (M8, #1610) — sie gehören dem Zauber, und der
	// ist mit dem Verrechnen vergangen. Derselbe Grund wie bei den Zielen.
	objekt.Wirkungen = nil

	// **Der Stempel ist neu** — CR 302.6 fragt, seit wann dieses Objekt
	// unter derselben Kontrolle dasteht, und die Antwort ist: seit jetzt.
	objekt.Seit = zug

	return objekt
}

// Abgeloest nimmt jedes Anhängsel von einem Objekt, das die Zone wechselt.
//
// **Die andere Richtung derselben Regel.** Das Anhängsel bleibt, wo es ist —
// aber woran es hing, gibt es nicht mehr. Was daraus folgt, entscheidet die
// zustandsbasierte Aktion (CR 704.5m): eine Aura an nichts geht in den
// Friedhof, eine Ausrüstung bleibt liegen.
func (p Partie) Abgeloest(kennung string) Partie {
	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	for stelle := range objekte {
		if objekte[stelle].HaengtAn == kennung {
			objekte[stelle].HaengtAn = ""
		}
	}

	p.Objekte = objekte

	return p
}

// zonenwechsel legt ein Objekt in eine andere Zone — nach CR 400.7.
//
// **Der eine Weg.** Wer ein Objekt von Hand umlegt, umgeht die Regel; wer
// diese Funktion ruft, bekommt sie.
func (p Partie) zonenwechsel(stelle int, ziel Zone) Partie {
	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	kennung := objekte[stelle].Kennung

	// **Obenauf, wo die Regel es sagt** — CR 404.1 für den Friedhof und
	// CR 405.2 für den Stapel: was hineinkommt, kommt **oben** darauf.
	//
	// **Null heißt oben** (der Grund steht bei [Spielobjekt.Ordnung]), und
	// deshalb rückt alles andere eine Stelle nach unten, statt dass das neue
	// eine negative Zahl bekäme.
	//
	// **Bis zum 2026-09-17 bekam jede Karte im Friedhof die Null**, und
	// damit lagen alle an derselben Stelle — CR 404.3 (der Eigner ordnet,
	// was gleichzeitig hineinkommt) hatte nichts, worauf es zeigen konnte.
	obenauf(objekte, ziel, objekte[stelle].Eigner)

	objekte[stelle] = NeuesObjekt(objekte[stelle], p.Zug.Nummer)
	objekte[stelle].Zone = ziel

	p.Objekte = objekte

	// **Und die Dauereffekte an ihm fallen weg** — CR 400.7 (M10,
	// #1612):
	// es ist ein neues Objekt, und der Effekt galt dem alten.
	return p.ohneEffekteAn(kennung).Abgeloest(kennung)
}

// obenauf rückt alles in einer Zone eine Stelle nach unten.
//
// **Je Eigner und nicht über alle** — ein Friedhof gehört einem Spieler
// (CR 404.1), und zwei Friedhöfe haben nichts miteinander zu tun. **Der
// Stapel gehört allen** (CR 405.1), und dort zählt jeder Eintrag.
func obenauf(objekte []Spielobjekt, ziel Zone, eigner string) {
	if ziel != ZoneFriedhof && ziel != ZoneStapel {
		return
	}

	for stelle := range objekte {
		if objekte[stelle].Zone != ziel {
			continue
		}

		if ziel == ZoneFriedhof && objekte[stelle].Eigner != eigner {
			continue
		}

		objekte[stelle].Ordnung++
	}
}
