// ESLint 9, Flat Config.
//
// ESLint bleibt bewusst eine Hauptversion zurück. Der Grund ist nicht die
// Technik, sondern ein Peer-Bereich: `eslint-plugin-jsx-a11y@6.10.2` gibt
// `^3 || … || ^9` an und weist ESLint 10 ab. Der Wechsel kostete für jeden
// Installationsweg `--legacy-peer-deps` — und das schaltet die Auflösung von
// Peer-Abhängigkeiten für den **ganzen** Lauf ab. Ein Dauerpreis für nichts.
// Nachgemessen am 2026-08-08; `werkzeuge.test.ts` schlägt an, sobald sich der
// Bereich überschneidet.
import js from '@eslint/js'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import reactHooks from 'eslint-plugin-react-hooks'
import tseslint from 'typescript-eslint'
import globals from 'globals'

export default tseslint.config(
  { ignores: ['dist', 'coverage', 'node_modules', 'wailsjs'] },

  js.configs.recommended,
  tseslint.configs.strictTypeChecked,
  tseslint.configs.stylisticTypeChecked,
  jsxA11y.flatConfigs.recommended,
  reactHooks.configs.flat['recommended-latest'],

  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      globals: globals.browser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      // Ein `any` in der Fassade zum Backend ist der Anfang davon, dass die
      // Typen dort nichts mehr zusagen.
      '@typescript-eslint/no-explicit-any': 'error',
      // Eine nicht abgewartete Zusage schluckt ihren Fehler.
      '@typescript-eslint/no-floating-promises': 'error',

      // **Ein rollender Bereich darf `tabIndex={0}` tragen** — und muss es.
      //
      // WCAG 2.1.1 verlangt, dass erreichbar ist, was rollt: ein Kasten mit
      // überlaufendem Text, den nur die Maus bewegen kann, ist für manche gar
      // nicht da. Die WAI beschreibt dafür `role="region"` mit Namen und
      // `tabindex="0"`; die Regel lässt von Haus aus nur `tabpanel` zu und
      // schlüge hier an — für genau das Muster, das sie schützen soll.
      //
      // **`region` und nichts weiter.** Die Ausnahme ist so schmal wie der
      // Fall, für den es sie gibt (#1166, Werkstück 4); ein `div` ohne Rolle
      // bleibt ein Befund.
      'jsx-a11y/no-noninteractive-tabindex': [
        'error',
        { tags: [], roles: ['tabpanel', 'region'] },
      ],

      // **Eine Tabelle darf ein Baumgitter sein** — UXR-17 (#1925). Die
      // Regel lässt an `table` von Haus aus nur `grid` zu; die Paketliste
      // klappt ihre Gruppen an der Zeile auf, und `aria-expanded` an einer
      // Zeile erlaubt ARIA nur im `treegrid` (axe: `aria-conditional-attr`).
      // Das Muster der WAI für ein Baumgitter ist eine `table`. **Sonst die
      // Vorgaben des Plugins**, Wort für Wort (`eslint-plugin-jsx-a11y`
      // 6.10.2, `lib/index.js`) — eine schmalere Ausnahme gibt es nicht.
      'jsx-a11y/no-noninteractive-element-to-interactive-role': [
        'error',
        {
          ul: [
            'listbox',
            'menu',
            'menubar',
            'radiogroup',
            'tablist',
            'tree',
            'treegrid',
          ],
          ol: [
            'listbox',
            'menu',
            'menubar',
            'radiogroup',
            'tablist',
            'tree',
            'treegrid',
          ],
          li: [
            'menuitem',
            'menuitemradio',
            'menuitemcheckbox',
            'option',
            'row',
            'tab',
            'treeitem',
          ],
          table: ['grid', 'treegrid'],
          td: ['gridcell'],
          fieldset: ['radiogroup', 'presentation'],
        },
      ],
    },
  },

  {
    // Die Konfigurationsdateien laufen unter Node, nicht im Browser, und
    // gehören nicht in das typgeprüfte Projekt — `projectService` fände sie
    // sonst nicht in `tsconfig.json` wieder.
    files: ['*.config.{js,ts}'],
    languageOptions: { globals: globals.node },
    ...tseslint.configs.disableTypeChecked,
  },
)
