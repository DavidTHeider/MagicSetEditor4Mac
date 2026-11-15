package msescript

// aufruf ruft eine Funktion oder bindet Argumente an sie.
//
// # Die Umgebung kommt vom Aufrufer
//
// Der Rumpf läuft in einer Ebene **über der Umgebung des Aufrufs**, nicht über
// der, in der er geschrieben wurde. Der Grund steht in `umgebung.go`:
// `mana_sort` benutzt `input`, das in keiner Zuweisung und in keiner
// Argumentliste steht. Plausibel, nicht belegt — siehe dort.
//
// Die Argumente legen sich als Namen in diese Ebene. **Das erste
// Stellungsargument heißt `input`** (#216): vier der neunzehn Exportvorlagen
// von `M15-Magic-Pack` lesen `input` in einem Block.
//
// Was daraus für `rgb(156, 130, 90)` folgt, entscheidet weiterhin die
// Standardbibliothek (#24): eine eingebaute Funktion bekommt ihre
// Stellungsargumente unverändert und liest kein `input`.
func (a *auswerter) aufruf(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	ziel, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	ruf, err := a.argumentwerte(knoten, umgebung)
	if err != nil {
		return Nichts, err
	}

	if knoten.Form == FormBindung {
		return a.binden(knoten, ziel, gebundeneArgumente(ruf))
	}

	// **`nil` gerufen gibt sein `input` zurück.**
	//
	// Herkunft: `ScriptNil::eval` in `src/script/value.cpp:456` von MSE2,
	// samt dem Kommentar, der dort steht — `nil(input) == input`.
	//
	// # Warum das kein Nachsehen ist
	//
	// In den Stilen steht `image: { module_card() }`. `module_card` entsteht in
	// einem **anderen** Paket; fehlt es, ist der Name `nil`. MSE zeichnet dann
	// nichts und malt die übrigen Ebenen der Karte trotzdem — dieses Programm
	// brach die Karte ab. Gemessen am 2026-08-16: 168 Feldstile scheiterten
	// genau hier, und keiner davon wegen eines Fehlers in der Vorlage
	// (#240).
	//
	// Der Rückgabewert ist `input` und nicht `nil`, weil `nil` in MSE damit die
	// **Einheit der Verkettung** ist: in `f + nil + g` verhält sich das mittlere
	// Glied wie eine Funktion, die durchreicht.
	//
	// Fassung f9d9356 vom 2021-07-06.
	if ziel.Art == ArtNichts {
		if len(ruf.Stellung) > 0 {
			return ruf.Stellung[0], nil
		}

		if eingabe, da := umgebung.Suche("input"); da {
			return eingabe, nil
		}

		return Nichts, nil
	}

	// **Eine eingebaute Funktion bekommt auch die Stellungsargumente.** Ein
	// Block kann sie nicht lesen — er hat keine Parameterliste —, aber `rgb`,
	// `max` und `length` werden ausschließlich so gerufen.
	// Ein Wert ohne Funktionsteil ist keine Funktion — die Prüfung darunter
	// fängt ihn, aber der Zugriff hier käme davor.
	// **Ein gerufener Fehlerwert bleibt einer.**
	//
	//	ScriptValueP eval(Context&, bool openScope) const override {
	//	  return delay_error(error);
	//	}
	//
	// `ScriptDelayedError`, `src/script/value.cpp`, Zeile 163, Fassung f9d9356
	// vom 2021-07-06. Die Vorlagen schreiben `f(x) or else "…"`, wo `f` selbst
	// aus einem Zugriff kommt — bräche der Aufruf ab, käme `or else` nie zum
	// Zug (siehe [Fehlerwert]).
	if ziel.IstFehler() {
		return ziel, nil
	}

	teil := ziel.Funktion
	if teil == nil {
		return Nichts, a.fehler(knoten, SchluesselKeineFunktion,
			string(ziel.Art))
	}

	// Ein anderes Skript desselben Netzes: der Umweg über das Netz macht die
	// Abhängigkeit sichtbar und fängt den Kreis.
	if teil.Netzname != "" && a.netz != nil {
		a.merke(teil.Netzname)

		return a.netz.Wert(a.ctx, teil.Netzname)
	}

	if teil.Eingebaut != nil {
		ruf.Benannt = verbunden(teil.Gebunden, ruf.Benannt)

		// **Ohne Stellungsargument gilt das `input` der Umgebung.**
		//
		// Herkunft: MSE hat gar keine Stellungsargumente. Jede eingebaute
		// Funktion liest die Kontextvariable `input` —
		// `SCRIPT_PARAM_C(String,input)` in `src/script/functions/editor.cpp:240`
		// für `chosen`, und ebenso überall sonst. Was wie ein
		// Stellungsargument aussieht, ist bei MSE eine Schreibweise, die
		// `input` setzt; genau deshalb steht in den Vorlagen
		// `chosen@(choice: "white")` ohne Klammerfüllung.
		//
		// Gemessen am 2026-08-16: 172 Feldstile scheiterten an „chosen
		// erwartet ein Argument", 162 weitere am unbekannten Namen `input`
		// (#240).
		//
		// **Das Stellungsargument schlägt die Umgebung**, nicht umgekehrt:
		// sonst gäbe `to_upper("x")` je nach Umgebung etwas anderes.
		//
		// Fassung f9d9356 vom 2021-07-06.
		if len(ruf.Stellung) == 0 {
			if eingabe, da := umgebung.Suche("input"); da {
				ruf.Stellung = []Wert{eingabe}
			}
		}

		return teil.Eingebaut(a.ctx, ruf.mitEingabe())
	}

	// Eine zweite Prüfung auf `ArtFunktion` träfe nie zu: nur eine Funktion hat
	// einen Funktionsteil, und ohne ihn ist der Aufruf schon oben abgewiesen.
	// Ein Zweig, den kein Test betreten kann, ist eine Behauptung darüber, was
	// im Ernstfall geschieht.
	//
	// Ein Funktionswert ohne Rumpf und ohne Go-Funktion ist ein Fehler dieses
	// Pakets, kein Fehler der Vorlage — gemeldet wird er trotzdem mit Stelle,
	// weil ein Absturz über eine Wails-Bindung die Anwendung beendet.
	if teil.Rumpf == nil {
		return Nichts, a.fehler(knoten, SchluesselOhneRumpf)
	}

	innen := umgebung.Darueber()

	// **Die Reihenfolge steht an einer Stelle** — siehe [belegeAufruf].
	belegeAufruf(innen, teil, ruf)

	return a.werte(teil.Rumpf, innen)
}

// binden legt Argumente an eine Funktion, ohne sie zu rufen.
func (a *auswerter) binden(knoten *Knoten, ziel Wert, gebunden []Eintrag) (Wert, error) {
	if ziel.Art != ArtFunktion || ziel.Funktion == nil {
		return Nichts, a.fehler(knoten,
			SchluesselBindenFunktion, string(ziel.Art))
	}

	teil := ziel.Funktion

	// **`Eingebaut` reist mit.** Ohne dieses Feld ergab `to_lower@()` einen
	// Funktionswert ohne Rumpf und ohne Go-Funktion; der Aufruf danach lief in
	// einen Absturz. Gefunden hat es der Test, der die Bindung an einer
	// eingebauten Funktion probiert.
	neu := &Funktionsteil{
		Rumpf:     teil.Rumpf,
		Eingebaut: teil.Eingebaut,
		Netzname:  teil.Netzname,
		Gebunden:  make([]Eintrag, 0, len(teil.Gebunden)+len(gebunden)),
	}
	neu.Gebunden = append(neu.Gebunden, teil.Gebunden...)
	neu.Gebunden = append(neu.Gebunden, gebunden...)

	return Wert{Art: ArtFunktion, Funktion: neu}, nil
}

// argumentwerte rechnet die Argumente aus, nach Stellung und Namen getrennt.
//
// Gerechnet wird **jedes**, auch ein Stellungsargument an einem Block, der es
// nicht lesen kann. Ein Argument, dessen Auswertung scheitert, soll den Fehler
// melden und nicht davon abhängen, wer gerufen wird.
func (a *auswerter) argumentwerte(knoten *Knoten, umgebung *Umgebung) (Ruf, error) {
	ruf := Ruf{Zeile: knoten.Zeile, Spalte: knoten.Spalte, Umgebung: umgebung}

	for i := 1; i < len(knoten.Kinder); i++ {
		wert, err := a.werte(knoten.Kinder[i], umgebung)
		if err != nil {
			return Ruf{}, err
		}

		// **Ein Fehlerwert als Argument wirft.** MSE reicht ihn an die
		// eingebaute Funktion weiter, und dort wirft die erste Umwandlung —
		// dieselbe Wirkung, eine Stelle früher und mit derselben Meldung
		// (siehe [Fehlerwert]).
		if wert.IstFehler() {
			return Ruf{}, a.fehler(knoten.Kinder[i], SchluesselDurchgereicht, wert.Text)
		}

		if knoten.Namen[i] == "" {
			ruf.Stellung = append(ruf.Stellung, wert)

			continue
		}

		ruf.Benannt = append(ruf.Benannt, Eintrag{Name: knoten.Namen[i], Wert: wert})
	}

	return ruf, nil
}

// gebundeneArgumente macht aus einem Ruf die Liste, die `@` festhält.
//
// **Das Stellungsargument gehört dazu, als `input`.** MSE kennt gar kein
// Stellungsargument: `f(x)` setzt `x` als Kontextvariable `input`
// (`src/script/parser.cpp`, Zeilen 881–893), und `Context::makeClosure` nimmt
// die Namen der innersten Ebene mit — `input` darunter
// (`src/script/context.cpp`, Zeilen 318–333). Zwischen `f@(x)` und
// `f@(input: x)` ist dort kein Unterschied.
//
// Fassung f9d9356 vom 2021-07-06.
//
// **Die Rechnung steht an einer Stelle.** Zwei Wege, die dasselbe tun sollen
// und es getrennt tun, laufen wieder auseinander.
func gebundeneArgumente(ruf Ruf) []Eintrag {
	gebunden := make([]Eintrag, 0, len(ruf.Benannt)+1)

	// **Zuerst**, damit ein ausdrückliches `input:` es schlägt — dieselbe
	// Reihenfolge wie beim Aufruf selbst.
	if len(ruf.Stellung) > 0 {
		gebunden = append(gebunden, Eintrag{Name: "input", Wert: ruf.Stellung[0]})
	}

	return append(gebunden, ruf.Benannt...)
}

// belegeAufruf legt die Argumente eines Aufrufs in eine Umgebung.
//
// # Die Reihenfolge ist gemessen, nicht gesetzt
//
//	zuerst  die mit `@` vorab gebundenen Werte
//	dann    die **benannten** Argumente des Aufrufs
//	zuletzt das **Stellungsargument**, als `input`
//
// Am 2026-08-20 am Orakel abgefragt, sechs Fälle
// (`TestWerSchlaegtWenBeimAufruf`). Zwei Belege tragen sie:
//
//   - **Eine Closure-Bindung greift nur, wenn die Variable nicht unter den
//     Argumenten des Aufrufs steht** (`src/script/value.cpp`, Zeile 608).
//     Deshalb steht `Gebunden` zuerst — `@` heißt vorbelegen, nicht
//     festnageln.
//   - **MSE kennt kein Stellungsargument.** Ein bloßes Argument wird beim
//     Zerteilen zu `input` (`src/script/parser.cpp`, Zeilen 881–893), und
//     `f(x, input: y)` ist dort `f(input: x, input: y)` — der erste gewinnt.
//     Deshalb steht die Stellung zuletzt.
//
// Fassung f9d9356 vom 2021-07-06.
//
// # Eine Stelle für beide Wege
//
// Die Reihenfolge steht an **einer** Stelle, und [auswerter_aufruf.go] wie
// [auswerter_ruf.go] nehmen sie von hier.
func belegeAufruf(innen *Umgebung, teil *Funktionsteil, ruf Ruf) {
	// Über den Index statt über den Wert: ein `Eintrag` wiegt inzwischen 176
	// Byte, und `gocritic` zählt jede Kopie in einer Schleife nach.
	for i := range teil.Gebunden {
		innen.Setze(teil.Gebunden[i].Name, teil.Gebunden[i].Wert)
	}

	for i := range ruf.Benannt {
		innen.Setze(ruf.Benannt[i].Name, ruf.Benannt[i].Wert)
	}

	// **Ohne Stellungsargument wird nichts gesetzt** — dann bleibt `input` aus
	// der äußeren Umgebung sichtbar, und wo es auch dort fehlt, ist es ein
	// unbekannter Name. Ein leerer Text an dieser Stelle machte aus einem
	// Tippfehler ein Ergebnis.
	if len(ruf.Stellung) > 0 {
		innen.Setze("input", ruf.Stellung[0])
	}
}
