package rest

// AusgabeDerMarke vermerkt eine Ausgabe für die Prozessmarke — **nur für
// Läufe**: die Wege über alle Endpunkte rufen `GET` und `DELETE
// /ausgang/{name}`, ohne die Ausgabe vorher anzulegen, und seit #2002 holt
// sie nur ab, wer sie anstieß.
func AusgabeDerMarke(name string) func(*Dienst) {
	return func(d *Dienst) { d.ausgaenge.vermerke(name, principalMarke) }
}
