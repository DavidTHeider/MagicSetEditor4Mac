package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`sort_list` ordnet eine Liste nach einem Schlüsselausdruck.**
//
// # Erschlossen, nicht erfunden
//
// Vier Exportvorlagen rufen sie, und alle in derselben Form:
//
//	for each card in sort_list(cards, order_by: {card.card_number}) do …
//
// `magic-forum-spoiler` schreibt `{input.card_number}` statt `{card...}` —
// **dasselbe Glied unter zwei Namen**. Der Schlüsselausdruck bekommt es
// deshalb als beide (#202).
func TestSortListeOrdnetNachDemSchluessel(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list(["b", "c", "a"], order_by: {input})`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if len(wert.Eintraege) != 3 ||
		wert.Eintraege[0].Wert.Text != "a" ||
		wert.Eintraege[2].Wert.Text != "c" {
		t.Errorf("%+v", wert.Eintraege)
	}
}

// **Das Glied heißt auch `card`.**
func TestSortListeKenntDasGliedAlsCard(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list([[n: 2], [n: 1]], order_by: {card.n})`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if len(wert.Eintraege) != 2 || wert.Eintraege[0].Wert.Eintraege[0].Wert.Zahl != 1 {
		t.Errorf("%+v", wert.Eintraege)
	}
}

// **Zahlen werden als Zahlen verglichen, nicht als Text.**
//
// Sonst stünde `10` vor `9`, und eine Kartennummernfolge käme durcheinander —
// genau wofür die Vorlagen sie rufen.
func TestSortListeVergleichtZahlenAlsZahlen(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list([10, 9, 100], order_by: {input})`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if len(wert.Eintraege) != 3 || wert.Eintraege[0].Wert.Zahl != 9 {
		t.Errorf("%+v — 9 gehört nach vorn", wert.Eintraege)
	}
}

// **Die Reihenfolge gleicher Schlüssel bleibt.**
//
// Eine Liste, deren Ordnung sich bei gleichen Schlüsseln von Lauf zu Lauf
// änderte, ergäbe eine Ausgabedatei, die sich ohne Änderung unterscheidet.
func TestSortListeIstStabil(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list([[k: 1, v: "erst"], [k: 1, v: "dann"]], order_by: {input.k})`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Eintraege[0].Wert.Eintraege[1].Wert.Text != "erst" {
		t.Errorf("%+v", wert.Eintraege)
	}
}

// **Ohne `order_by` wird nach dem Glied selbst geordnet.**
func TestSortListeOhneSchluessel(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list(["b", "a"])`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Eintraege[0].Wert.Text != "a" {
		t.Errorf("%+v", wert.Eintraege)
	}
}

// **Ein Fehler im Schlüsselausdruck kommt heraus.**
func TestSortListeMeldetEinenFehlerImSchluessel(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list(["a"], order_by: {gibtesnicht()})`,
		msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler")
	}
}

// **Und etwas, das keine Liste ist, wird benannt abgewiesen.**
func TestSortListeBrauchtEineListe(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list("keine Liste")`, msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler")
	}
}

// **`order_by` erwartet einen Ausdruck in geschweiften Klammern.**
//
// Ein Text dort ist ein Schreibfehler in der Vorlage — und er wird benannt,
// statt still nach dem Text zu ordnen.
func TestSortListeWeistEinFalschesOrderByAb(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list(["a"], order_by: "name")`, msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ohne Stellungsargument ebenfalls.**
func TestSortListeBrauchtEinArgument(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list()`, msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler")
	}
}

// **Der Ordnungsausdruck sieht die Namen des Aufrufers**
// (#651).
//
//	while (ScriptValueP v = it->next()) {
//	  ctx.setVariable(set ? _("card") : _("input"), v);
//	  values.push_back(make_pair(order_by.eval(ctx)->toString(), v));
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 489–520, Fassung f9d9356 vom
// 2021-07-06. **Derselbe `ctx`** — der Rumpf läuft in der Umgebung der
// Rufstelle, und gesetzt wird obendrauf allein das Glied.
//
// `magic-forum-booster` erklärt sein `booster_sort` selbst und ruft es aus
// `order_by`; über einer frischen Umgebung ist der Name fort.
func TestSortListSiehtDieNamenDesAufrufers(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		quelle   string
		erwartet string
	}{
		{
			name: "eine eigene Funktion aus dem Ordnungsausdruck",
			quelle: `rang := { if input == "b" then "1" else "2" }
sort_list(["a", "b"], order_by: { rang() })`,
			erwartet: "ba",
		},
		{
			name: "ein eigener Wert aus dem Ordnungsausdruck",
			quelle: `vorn := "b"
sort_list(["a", "b"], order_by: { if input == vorn then "1" else "2" })`,
			erwartet: "ba",
		},
		// **Das Glied liegt darüber und bleibt nicht liegen.** Ein `input`,
		// das das Ordnen überdauerte, änderte still, was danach danebensteht.
		{
			name: "das Glied überschreibt den Namen des Aufrufers nicht dauerhaft",
			quelle: `input := "aussen"
geordnet := sort_list(["a", "b"], order_by: { input })
input`,
			erwartet: "aussen",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("unerwarteter Fehler: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("Ergebnis = %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
