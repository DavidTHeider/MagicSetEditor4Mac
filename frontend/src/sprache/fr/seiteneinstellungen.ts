import { dienstkatalog } from './einstellungenkatalog'
import { einstellungsbelegung } from './einstellungsbelegung'
import { einstellungsgruppennamen } from './einstellungsgruppennamen'
import {
  schalternamen,
  schalterwerte,
  seitenschalter,
} from './einstellungenschalter'

// **Die Texte der Einstellungsseite** — ausgelagert für Phase 4.8
// (#1321).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als
// einer ankommt** (§9) — dieselbe Gestalt wie `seitenkarten.ts`. Mit den
// Dienstschaltern ging `seiten.ts` über die Zeilengrenze aus §2.
import { ki } from './seitenki'
import { netzbereich, netzhilfe, netznamen } from './einstellungennetz'
import { handelshilfe, handelsnamen, handelswerte } from './einstellungenhandel'
import {
  gestaltungsbeiwerk,
  gestaltungshinweise,
  gestaltungsnamen,
  gestaltungswerte,
} from './einstellungengestaltung'
import { zugaenge } from './einstellungenzugaenge'
import { ansichtsnamen, ansichtswerte } from './einstellungenansicht'
import { vorgabennamen } from './einstellungenvorgaben'
import {
  protokollhinweise,
  protokollnamen,
  protokollwerte,
} from './einstellungenprotokoll'
import { austausch } from './einstellungenaustausch'
import { eigenuebersetzung } from './eigenuebersetzung'
import { kihinweise, kinamen, kiwerte } from './einstellungenki'

export const einstellungen = {
  ...einstellungsgruppennamen,
  ...seitenschalter,

  titel: 'Réglages',

  namen: {
    ...schalternamen,
    sprache: 'Langue',
    'sprache.eigene': 'Traduction personnelle',
    darstellung: 'Format régional',
    zeitzone: 'Fuseau horaire',
    'filter.gespeichert': 'Filtres enregistrés',
    'kuerzel.ziffern': 'Attribution de ⌘0 à ⌘9',
    erscheinungsbild: 'Thème',
    schriftgroesse: 'Taille du texte',
    aktualisierung: 'Rechercher les nouvelles versions',
    ersteschritte: 'Introduction au démarrage',
    seitenleiste: 'Barre latérale',
    leistenlage: 'Position de la barre latérale',
    fensterstart: 'Fenêtre au démarrage',
    fensterbreite: 'Dernière largeur mémorisée',
    fensterhoehe: 'Dernière hauteur mémorisée',
    schriftart: 'Police d’affichage',
    feldtiefe: 'Formulaire de carte',
    'tisch.autoanordnung': 'Disposer automatiquement les cartes sur la table',

    'barrierefreiheit.bewegung': 'Animations',
    'barrierefreiheit.kontrast': 'Contraste',
    'sicherheit.faktorbeimwechsel':
      'Deuxième facteur lors du changement de profil',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **A cloud provider's address** — piece 3 of A18 (#1533), labelled with
    // #1540.
    'dienst.anthropic.adresse': 'Adresse d’Anthropic',
    'dienst.openai.adresse': 'Adresse d’OpenAI',
    'dienst.xai.adresse': 'Adresse de xAI',
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': 'Lire mon propre compte de marché',
    ...kinamen,
    // **The client id per provider** — Werkstück 4 of Phase 4.12 (#1351).
    'anbieter.github.kennung': 'Identifiant client chez GitHub',
    'anbieter.google.kennung': 'Identifiant client chez Google',
    'anbieter.microsoft.kennung': 'Identifiant client chez Microsoft',
    'anbieter.x.kennung': 'Identifiant client chez X',

    // **The workspace and the two limits** — Werkstück O3, Phase 5.2 (#1427).
    'arbeitsplatz.navigator': 'Colonne de gauche',
    'arbeitsplatz.inspektor': 'Colonne de droite',
    'arbeitsplatz.navigatorbreite': 'Largeur de la colonne de gauche',
    'arbeitsplatz.inspektorbreite': 'Largeur de la colonne de droite',
    'richtlinie.nurlokal': 'Travailler uniquement en local',
    'richtlinie.kostengrenze':
      'Plafond de coûts dans la plus petite unité (centimes pour EUR)',
  },

  // Origin, default and search — work item U11 of phase 5.1 (#1422).
  herkunftMarke: 'défini par {herkunft}',
  wirksamKurz: 'En vigueur : « {wert} ».',
  wirksamOffenKurz: 'La valeur de l’ordinateur s’applique.',
  wirksamGehalten: 'En vigueur : « {wert} » — une politique le maintient.',
  herkunftZeile: 'Défini par {herkunft}, niveau {ebene}.',
  herkunftNamen: {
    profil: 'ce profil',
    nutzer: 'toi',
    geraet: 'cet appareil',
    vorgabe: 'personne (valeur par défaut)',
  },
  ebenenNamen: {
    profil: 'profil',
    nutzer: 'utilisateur',
    geraet: 'appareil',
    arbeitsplatz: 'espace de travail',
  },
  ebenenErklaerung:
    'Chaque réglage s’applique à un niveau : appareil, utilisateur, profil ou espace de travail. **Appareil** décrit cet ordinateur — un chemin, une taille de fenêtre — et n’est pas inclus dans un transfert de réglages. **Utilisateur** vaut pour toi dans chacun de tes profils. **Profil** ne vaut qu’ici et prévaut sur ce que tu as défini comme utilisateur. **Espace de travail** correspond à la disposition des colonnes de cette fenêtre. La valeur en vigueur est celle du profil, sinon celle que tu as définie comme utilisateur, sinon la valeur par défaut — pour une politique, la valeur la plus stricte. Aucun réglage ne demande de droits d’administrateur.',
  fachleute: 'Ce que les experts peuvent aussi régler ici',
  fachleuteSatz:
    'Délais, accessibilité depuis d’autres réseaux, la voie de sortie et les services destinés à d’autres programmes. Déplier ne change rien — cela affiche seulement davantage.',
  zuruecksetzen: 'Rétablir « {vorgabe} »',
  erbeMitWert: '{erbe} ({wert})',
  zuruecksetzenVon: 'Rétablir la valeur par défaut de {name}',
  suche: 'Rechercher un réglage',
  suchePlatzhalter: 'Nom, clé ou valeur',
  sucheTreffer: '{zahl} résultats',
  sucheKeine: 'Aucun réglage ne correspond à « {frage} ».',
  zumGebiet: 'Afficher dans la zone « {gebiet} »',

  // **Below the appearance settings** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  hinweise: {
    ...protokollhinweise,
    'sprache.eigene':
      'Quelle traduction personnelle la fenêtre utilise — vide signifie aucune. Si elle n’existe plus, la langue choisie s’applique de nouveau.',
    feldtiefe:
      'La part du formulaire de carte qui est affichée. **Normal** est ce que tu remplis ; **Avancé** ajoute ce que calcule un script du paquet de jeu ; **Expert** ajoute ce que le paquet de jeu n’explique pas du tout — `stylesheet` et consorts. Les trois s’emboîtent : plus profond signifie plus, jamais autre chose. S’applique par profil.',
    sprache:
      'S’applique à toute la fenêtre et à la barre des menus. Ce qui provient d’un paquet MSE — noms de champs, noms de jeux — reste tel qu’écrit dans le fichier. L’allemand est la langue source ; les autres langues sont traduites automatiquement et pas encore relues.',
    darstellung:
      'La présentation des nombres, des montants et des dates — par exemple **de-CH** ou **en-GB**. Vide signifie : comme le système. La langue des textes n’est pas concernée.',
    zeitzone:
      'Le fuseau dans lequel les heures sont affichées — par exemple **Europe/Berlin**. « Comme le système » désigne celui de l’ordinateur. Ce réglage est enregistré par profil et n’est pas inclus dans un transfert de réglages.',
    'filter.gespeichert':
      'Les filtres enregistrés de chaque vue sous forme de liste JSON — nom, vue, repères et si c’est **et** ou **ou** qui s’applique. Ils sont créés dans le navigateur de leur vue ; ils appartiennent au profil et sont inclus dans l’export des réglages.',
    'kuerzel.ziffern':
      'La page qu’ouvre un chiffre avec ⌘. Vide signifie : les zones de la barre latérale. Cela se règle dans le tableau des raccourcis sous « Général ».',
    erscheinungsbild:
      'Clair, sombre, ou selon ce que dit l’ordinateur. Ce choix ne s’applique qu’à cette application.',
    schriftgroesse:
      'Met à l’échelle toute l’interface, pas seulement le texte — les espacements et les bordures évoluent avec lui. L’aperçu de la carte reste tel quel : il montre ce qui serait imprimé.',
    aktualisierung:
      'Par défaut : **désactivé** — rien ne part sur le réseau et rien n’est demandé. Avec **notifier**, l’application vérifie au démarrage si une version plus récente existe et le signale — rien de plus. Avec **télécharger**, elle propose en plus un bouton qui récupère la version dans le dossier de téléchargements et compare sa somme de contrôle ; rien n’est installé automatiquement.',
    ersteschritte:
      'Les trois étapes jusqu’à ta première carte, sous forme d’une fenêtre au démarrage. Le même interrupteur se trouve dans cette fenêtre — désactive-le là et tu le retrouves ici.',
    seitenleiste:
      'La barre à gauche avec les zones de l’application. Repliée, sa largeur revient aux espaces de travail ; ⌥⌘S, le bouton sur son bord et l’entrée sous **Présentation** la replient et la déplient. Elle est enregistrée par profil et n’est pas incluse dans un transfert de réglages.',
    leistenlage:
      'À gauche, elle se dresse à la verticale à côté du contenu ; en haut, elle forme une rangée au-dessus. Mêmes entrées, mêmes raccourcis et mêmes touches fléchées — seulement en travers. Elle est enregistrée par profil et n’est pas incluse dans un transfert de réglages.',
    fensterstart:
      'Par défaut : **Remplir l’écran** — la fenêtre prend la place dont elle dispose, et **pas le plein écran** : celui-ci emporterait la barre des menus. **Taille normale** ouvre la fenêtre à sa taille prévue, **Comme la dernière fois** dans l’état où tu l’as laissée. Ce réglage est enregistré par profil et n’est pas inclus dans un transfert de réglages.',
    fensterbreite:
      'Ce qui a été mesuré à la dernière fermeture de la fenêtre — **uniquement pour « Comme la dernière fois »**. Vide signifie qu’il n’y a pas encore eu de dernière fois, et la fenêtre s’ouvre à sa taille prévue.',
    fensterhoehe:
      'La même chose pour la hauteur. Les deux valeurs sont écrites à la fermeture de la fenêtre ; une mesure en dehors de 700 à 8192 points n’est pas mémorisée — une fenêtre qui ne tiendrait pas sur l’écran suivant ne pourrait plus être saisie.',
    schriftart:
      'La police des titres et des noms de sets. Le corps du texte reste dans la police du système — une police décorative sur un paragraphe se lit mal. Les quatre sont fournies avec macOS ; rien n’est téléchargé.',
    'barrierefreiheit.bewegung':
      'La dérogation ne va **que dans un sens** : elle désactive les animations, jamais l’inverse. Qui a choisi **réduire les animations** dans le système n’en a de toute façon pas ici — une application qui passerait outre irait à son encontre.',
    'barrierefreiheit.kontrast':
      'Même sens : **élevé** supprime la texture de fibres et renforce les couleurs jusqu’à ce que chaque texte atteigne 7 : 1. Qui demande plus de contraste au système obtient la même chose.',
    'sicherheit.faktorbeimwechsel':
      'La **seule** dérogation qui assouplit — et seulement lors du **changement** de profil. Elle ne change rien à la connexion : là, le deuxième facteur fait ce pour quoi il existe. Sur un ordinateur utilisé par une seule personne, un code à usage unique à chaque changement est une contrainte sans bénéfice.',
    'vorgabe.zustand':
      'Ce avec quoi commence une nouvelle ligne de la collection. **Aucune valeur par défaut** signifie que le champ reste vide, comme toujours.',
    'vorgabe.waehrung':
      'Ce avec quoi commence le champ de devise d’une nouvelle ligne, par exemple **EUR**. Rien n’est converti.',
    'vorgabe.einzelspiel':
      'Indique si un nouveau tournoi se joue en matchs d’**une seule** partie. Cela décide de l’affichage, pas du calcul — le pourcentage de parties gagnées et celui des adversaires sont conservés dans les deux cas.',
    'vorgabe.hochladepfad':
      'Le dossier dans lequel s’ouvre une boîte de dialogue d’ouverture de fichier — images, MTG JSON et tout ce qui vient du disque. Un chemin complet ou commençant par **~/** ; vide signifie là où le système d’exploitation l’ouvre.',
    'vorgabe.ausfuhrpfad':
      'L’endroit où s’ouvre une boîte de dialogue d’enregistrement lorsque l’action n’a pas de lieu propre. Si le dossier est absent au moment de l’enregistrement — un disque débranché, par exemple —, l’application te le signale au lieu d’enregistrer ailleurs.',
    'vorgabe.sicherungspfad':
      'L’endroit où s’ouvre la boîte de dialogue de sauvegarde de ta bibliothèque. Un dossier absent interrompt la sauvegarde avec un message.',
    'vorgabe.diagnosepfad':
      'L’endroit où sont enregistrés les textes et les lots de diagnostic. Un dossier absent interrompt l’opération avec un message.',
    'bildausfuhr.ordner':
      'L’endroit où la sortie d’images a écrit pour la dernière fois. La page Images le propose et inscrit ici chaque dossier nouvellement choisi — un seul pour tous les sets, comme dans MSE. Vide signifie que la boîte de dialogue s’ouvre sur le dossier pour les fichiers enregistrés.',
    'ausgabe.ziele':
      'La cible que « Exporter » a montrée en dernier pour chaque set. Une liste en JSON, le choix le plus récent à la fin, 200 au plus ; elle s’écrit au choix sous « Vers où ? ». Vide signifie : rien n’est présélectionné.',
    'dienst.anthropic.adresse':
      'Où vont les requêtes. Vide signifie l’interface propre d’Anthropic. Une autre adresse est une **passerelle** — un service qui transmet la requête ; sa clé est stockée séparément.',
    'dienst.openai.adresse':
      'La même chose pour OpenAI. Vide signifie `https://api.openai.com`. Toute valeur saisie ici signifie que tu ne parles plus à OpenAI, mais à ce qui répond à cette adresse.',
    'dienst.xai.adresse':
      'La même chose pour xAI. Vide signifie `https://api.x.ai`.',
    ...netzhilfe,
    ...handelshilfe,
    ...gestaltungshinweise,
    'konto.schau':
      'Indique si ce programme peut lire ton **propre** compte sur une place de marché — tes annonces et tes commandes. **Désactivé** signifie : pas même un coup d’œil. Activer la place de marché elle-même ne suffit pas : celle-ci répond à ce que demandent **d’autres** vendeurs, cette fonction à ce qui se trouve **chez toi**. Rien n’y est jamais écrit — ni annonce, ni panier, ni achat.',
    ...kihinweise,
    'anbieter.github.kennung':
      'Le `client_id` de l’application OAuth que tu as créée chez GitHub. Vide signifie : la connexion avec GitHub n’est pas proposée.',
    'anbieter.google.kennung':
      'Le `client_id` de l’application OAuth que tu as créée chez Google. Il se termine en général par `apps.googleusercontent.com`.',
    'anbieter.microsoft.kennung':
      'L’identifiant d’application du portail Microsoft Entra — un UUID. Microsoft n’exige **aucun** secret pour ce flux.',
    'anbieter.x.kennung':
      'Le `client_id` de l’application que tu as créée dans le portail développeur de X.',
  },

  werte: {
    ...schalterwerte,
    ...protokollwerte,
    ...handelswerte,
    ...gestaltungswerte,
    ...ansichtswerte,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **Die Kontoschau** — T11, #1476. Dieselben Worte wie bei der
    // KI-Laufzeit darüber: zwei Schreibweisen für denselben Zustand
    // lernt niemand gern.
    'konto.schau.aus': 'Désactivé',
    'konto.schau.an': 'Activé',
    ...kiwerte,
    'sprache.erben': 'Comme chez l’utilisateur',
    'erscheinungsbild.erben': 'Comme chez l’utilisateur',
    'schriftgroesse.erben': 'Comme chez l’utilisateur',
    'schriftart.erben': 'Comme chez l’utilisateur',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': 'Comme le système',
    'erscheinungsbild.hell': 'Clair',
    'erscheinungsbild.dunkel': 'Sombre',
    'schriftgroesse.klein': 'Petite',
    'schriftgroesse.normal': 'Normale',
    'schriftgroesse.gross': 'Grande',
    'schriftgroesse.sehrgross': 'Très grande',
    'aktualisierung.aus': 'Désactivé',
    'aktualisierung.melden': 'Notifier uniquement',
    'aktualisierung.holen': 'Notifier et télécharger',
    'ersteschritte.an': 'Activé',
    'ersteschritte.aus': 'Désactivé',
    'seitenleiste.sichtbar': 'Visible',
    'seitenleiste.verborgen': 'Masquée',
    'leistenlage.senkrecht': 'Gauche',
    'leistenlage.waagerecht': 'Haut',
    'fensterstart.maximiert': 'Remplir l’écran',
    'fensterstart.normal': 'Taille normale',
    'fensterstart.zuletzt': 'Comme la dernière fois',

    'feldtiefe.normal': 'Normal',
    'feldtiefe.erweitert': 'Avancé',
    'feldtiefe.experte': 'Expert',
    'tisch.autoanordnung.an': 'Oui — la table les dispose',
    'tisch.autoanordnung.aus': 'Non — ce que tu poses reste en place',
    'tisch.beschriftung': 'Étiquettes sur la table',
    'tisch.beschriftung.aus': 'Désactivées — seulement les quantités',
    'tisch.beschriftung.klein': 'Petites',
    'tisch.beschriftung.normal': 'Moyennes',
    'tisch.beschriftung.gross': 'Grandes',
    'schriftart.pergament': 'Parchemin — calme et lisible',
    'schriftart.fantasy': 'Fantasy — ludique',
    'schriftart.antike': 'Antiquité — capitales romaines',
    'schriftart.kalligrafie': 'Calligraphie — écrite à la main',
    'barrierefreiheit.bewegung.system': 'Comme le système',
    'barrierefreiheit.bewegung.reduziert': 'Toujours peu',
    'barrierefreiheit.kontrast.system': 'Comme le système',
    'barrierefreiheit.kontrast.hoch': 'Toujours élevé',
    'sicherheit.faktorbeimwechsel.verlangen': 'Aussi lors du changement',
    'sicherheit.faktorbeimwechsel.nurAnmelden': 'Uniquement à la connexion',
    'vorgabe.zustand.ohne': 'Aucune valeur par défaut',
    'vorgabe.zustand.mint': 'Mint',
    'vorgabe.zustand.nearMint': 'Near Mint',
    'vorgabe.zustand.excellent': 'Excellent',
    'vorgabe.zustand.good': 'Good',
    'vorgabe.zustand.lightPlayed': 'Light Played',
    'vorgabe.zustand.played': 'Played',
    'vorgabe.zustand.poor': 'Poor',
    'vorgabe.einzelspiel.aus': 'Désactivé',
    'vorgabe.einzelspiel.an': 'Activé',

    // **The workspace and the limit** — Werkstück O3 of Phase 5.2 (#1427).
    'arbeitsplatz.navigator.erben': 'Comme chez l’utilisateur',
    'arbeitsplatz.navigatorbreite.erben': 'Comme chez l’utilisateur',
    'arbeitsplatz.inspektorbreite.erben': 'Comme chez l’utilisateur',
    'arbeitsplatz.navigator.sichtbar': 'Visible',
    'arbeitsplatz.navigator.verborgen': 'Masquée',
    'arbeitsplatz.inspektor.erben': 'Comme chez l’utilisateur',
    'arbeitsplatz.inspektor.sichtbar': 'Visible',
    'arbeitsplatz.inspektor.verborgen': 'Masquée',
    'richtlinie.nurlokal.aus': 'Désactivé',
    'richtlinie.nurlokal.an': 'Activé',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Where a login comes from in the first place** — follow-up to Werkstück 4
  // of Phase 4.12 (#1351).
  anmeldung: 'Connexion via un fournisseur',
  anmeldungHinweis:
    'Cette application ne fournit **aucun identifiant**. Pour te connecter avec GitHub, Google, Microsoft ou X, crée toi-même une **application OAuth** chez le fournisseur et saisis ici son `client_id`. Comme URI de redirection, indique l’adresse affichée dans le navigateur pendant la connexion — elle pointe vers cet ordinateur et n’est valable que le temps de la connexion.',

  anmeldungGeheimnis:
    'GitHub exige en plus un `client_secret` — Google le laisse facultatif, Microsoft l’interdit pour des programmes comme celui-ci. Il va **dans le trousseau** et non ici : la bibliothèque se partage, le trousseau non. Crée-le dans **Trousseau d’accès** comme **mot de passe d’app**, avec le nom `MagicSetEditor4Mac-Anmeldung` et le compte `github`. En ligne de commande : `security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`.',

  geheimnis: 'Secret',
  geheimnisPflicht: 'est requis et manquant',
  geheimnisVerboten:
    'n’exige aucun secret et en refuse un — l’identifiant client suffit ici.',

  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  wirktNachNeustart: 'Prend effet après un redémarrage.',

  trefferInGruppe: '{zahl} résultats pour « {frage} »',
  gesuchtMarke: 'trouvé',
  sucheSprung: 'Aller au premier résultat',

  vorgaben: 'Valeurs par défaut',

  // **The pointer to the help** — Werkstück 13 of Phase 4.12 (#1351).

  dienste: 'Services de cartes tiers',
  diensteHinweis:
    'Désactivé signifie : pas une seule requête ne part vers ce service. L’application continue de fonctionner sans aucun service — seul ce qui vient de l’extérieur manque alors.',

  kiAnbieter: 'Modèles de langage',
  kiAnbieterHinweis:
    'Où les questions peuvent être envoyées. Chaque fournisseur calcule sur ses propres machines et voit ce que tu envoies — chacun est donc activé séparément, et aucun ne l’est de lui-même.',
  // See the German catalogue for why these two live here and not under
  // `werte`: they are the same for every service.
  dienstAus: 'Désactivé',
  dienstAn: 'Activé',
  dienstUngeprueft: 'Jamais testé face au vrai service',
  dienstUngeprueftHinweis:
    'Il n’y a ici aucun compte pour ce service. Ce qui est testé, c’est notre côté de la frontière — la requête, les en-têtes, la correspondance. Que le service réponde comme nous le supposons n’est pas mesuré.',
  fragen: 'Qui répond à quoi',
  fragenHinweis:
    'Lorsque plusieurs services peuvent répondre à la même question, ce réglage décide lequel est interrogé. **Il n’y a pas de repli silencieux** : si celui qui est choisi reste muet, on te le dit — au lieu de te montrer discrètement une réponse d’une autre source.',
  'frage.suche': 'Recherche de cartes',

  // The service catalogue — work item A1 of phase 5.3 (#1440).
  ...dienstkatalog,

  dienstSchluesselFehlt: 'Aucun identifiant enregistré',
  dienstSchluesselHinweis:
    'Ce service a besoin d’identifiants. Ils se trouvent dans le trousseau de macOS, et non dans le fichier de la collection — celui-ci circule.',
  zugaenge: 'Identifiants',
  zugaengeHinweis:
    'Ils se trouvent dans le trousseau de macOS et **pas** dans le fichier de la collection — celui-ci circule. Ce que tu vois ici indique seulement **si** quelque chose est enregistré ; la valeur elle-même ne revient jamais.',
  // See the German catalogue: these name the headers each service asks for
  // (documented in decision 0056).

  kuerzelUeberschrift: 'Raccourcis clavier',
  ...einstellungsbelegung,
  kuerzelHinweis:
    'Ils s’appliquent dans toute la fenêtre. Dans un champ de texte, ⌘Z et ⌘F appartiennent au champ.',
  kuerzelZweiteReihe:
    'Les huit raccourcis avec ⌥ se trouvent sur les touches de la rangée du haut, en partant de la gauche. ' +
    'Sur un clavier à disposition différente — AZERTY, par exemple — un autre caractère ' +
    'y est imprimé ; la touche est la même.',
  kuerzelTaste: 'Touche',
  kuerzelHandlung: 'Action',

  austausch,
  eigenuebersetzung,
  // **An account's remaining quota at a provider** — A5, #1444.
  kontingentTitel: 'Quota chez le fournisseur',
  kontingentErklaerung:
    'Le nombre de crédits que comprend le forfait et le nombre restant. L’interrogation passe par le réseau et coûte elle-même un crédit.',
  kontingentHolen: 'Interroger le quota',
  kontingentHolenBei: 'Interroger le quota chez {dienst}',
  kontingentLaeuft: 'Interrogation en cours …',
  kontingentZahlen: 'Il reste {rest} crédits sur {gesamt}.',
  kontingentFrist:
    'Le fournisseur peut communiquer ce chiffre avec jusqu’à {minuten} minutes de retard — ce n’est pas un plafond. Ce qui limite réellement les dépenses ici, c’est le budget de cette bibliothèque.',
  kontingentUeberzug: 'Un crédit au-delà du forfait coûte {satz}.',
  // **What is known about AI models** — A6, #1445.
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}
