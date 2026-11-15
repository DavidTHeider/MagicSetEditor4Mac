// **Was eine aufgehobene Seite zeigt, sieht niemand** — UX-03.AK-02
// (#1675).
//
// Eine verlassene Seite bleibt in einer `<Activity>` stehen, und React
// verbirgt sie mit `display: none`. Die Rollenabfragen übergehen das von
// selbst; eine Textabfrage nicht — sie fände denselben Satz zweimal, wenn
// die verlassene Seite ihn auch trägt. **`ignore` nimmt ihr die verborgenen
// Bäume**, samt der zwei, die Testing Library ohnehin übergeht.
export const verborgen =
  'script, style, [style*="display: none"], [style*="display: none"] *'
