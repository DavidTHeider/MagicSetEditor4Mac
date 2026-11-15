// **Der Hinweis, dass hier nur gelesen wird** — Werkstück 16 von Phase 4.12
// (#1351),
// nach [0073](../../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md).
//
// # Warum er über der ganzen Anwendung steht
//
// **Weil er für alles gilt**, solange der Zeiger dort steht — nicht für ein
// Feld und nicht für eine Seite. Ein Hinweis neben einem einzelnen Knopf sagte
// dasselbe zwanzigmal und an neunzehn falschen Stellen.
//
// **Und weil §6 es verlangt**: eine Oberfläche, die zwanzig Knöpfe zeigt, von
// denen jeder mit derselben Absage endet, erklärt den nächsten Schritt nicht.
// Der Schutz sitzt im Treiber und meldet sich beim ersten Schreibversuch —
// dieser Streifen sagt es vorher.
//
// **Gefragt wird der Arbeitskontext, nicht mehr die Leseart** — Werkstück O1
// von Phase 5.2 (#1425):
// dieselbe Auskunft in einem Zug mit dem Menschen, dem Profil und der Rolle,
// und `nurLesend` gilt dort auch für ein Token, das nur lesen darf.
import { useNurLesend } from '../haken/schreibrecht'
import { useTexte } from '../sprache/kontext'

// **Gefragt wird nicht mehr hier** (G10, #2276): der Raum `Schreibrecht`
// fragt einmal je Profilstand, und dieser Streifen liest ihn wie jeder Knopf,
// der schreibt.
export function Nurlesehinweis() {
  const texte = useTexte()
  const nurLesend = useNurLesend()

  if (!nurLesend) {
    return null
  }

  return (
    <p
      role="status"
      // **Warnung auf betontem Grund, beides Merkmale** — N13 (Phase 6.10,
      // #1671): bis dahin `text-amber-800` auf `bg-slate-100`, ein festes
      // Paar, das keiner Farbwelt folgte. Am Fenster gemessen (`n13`) und in
      // `docs/KONTRASTE.md` eingetragen: 6,10 hell, 6,95 dunkel.
      className="border-b border-kante bg-betont px-4 py-2 text-sm text-fehler"
    >
      {texte.bausteine.nurlesen.hinweis}
    </p>
  )
}
