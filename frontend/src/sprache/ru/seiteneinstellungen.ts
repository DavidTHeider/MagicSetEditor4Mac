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

  titel: 'Настройки',

  namen: {
    ...schalternamen,
    sprache: 'Язык',
    'sprache.eigene': 'Собственный перевод',
    darstellung: 'Региональные форматы',
    zeitzone: 'Часовой пояс',
    'filter.gespeichert': 'Сохранённые фильтры',
    'kuerzel.ziffern': 'Назначение ⌘0 – ⌘9',
    erscheinungsbild: 'Тема',
    schriftgroesse: 'Размер текста',
    aktualisierung: 'Проверять наличие новых версий',
    ersteschritte: 'Введение при запуске',
    seitenleiste: 'Боковая панель',
    leistenlage: 'Положение панели',
    fensterstart: 'Окно при запуске',
    fensterbreite: 'Последняя запомненная ширина',
    fensterhoehe: 'Последняя запомненная высота',
    schriftart: 'Декоративный шрифт',
    feldtiefe: 'Форма карты',
    'tisch.autoanordnung': 'Автоматически раскладывать карты на столе',

    'barrierefreiheit.bewegung': 'Движение',
    'barrierefreiheit.kontrast': 'Контрастность',
    'sicherheit.faktorbeimwechsel': 'Второй фактор при смене профиля',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **A cloud provider's address** — piece 3 of A18 (#1533), labelled with
    // #1540.
    'dienst.anthropic.adresse': 'Адрес Anthropic',
    'dienst.openai.adresse': 'Адрес OpenAI',
    'dienst.xai.adresse': 'Адрес xAI',
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': 'Читать собственную учётную запись на рынке',
    ...kinamen,
    // **The client id per provider** — Werkstück 4 of Phase 4.12 (#1351).
    'anbieter.github.kennung': 'Идентификатор клиента в GitHub',
    'anbieter.google.kennung': 'Идентификатор клиента в Google',
    'anbieter.microsoft.kennung': 'Идентификатор клиента в Microsoft',
    'anbieter.x.kennung': 'Идентификатор клиента в X',

    // **The workspace and the two limits** — Werkstück O3, Phase 5.2 (#1427).
    'arbeitsplatz.navigator': 'Левая колонка',
    'arbeitsplatz.inspektor': 'Правая колонка',
    'arbeitsplatz.navigatorbreite': 'Ширина левой колонки',
    'arbeitsplatz.inspektorbreite': 'Ширина правой колонки',
    'richtlinie.nurlokal': 'Работать только локально',
    'richtlinie.kostengrenze':
      'Предел расходов в наименьшей единице (для EUR — центы)',
  },

  // Origin, default and search — work item U11 of phase 5.1 (#1422).
  herkunftMarke: 'задано {herkunft}',
  wirksamKurz: 'Действует «{wert}».',
  wirksamOffenKurz: 'Действует значение по умолчанию для компьютера.',
  wirksamGehalten: 'Действует «{wert}» — его задаёт политика.',
  herkunftZeile: 'Задано {herkunft}, уровень: {ebene}.',
  herkunftNamen: {
    profil: 'этим профилем',
    nutzer: 'тобой',
    geraet: 'этим устройством',
    vorgabe: 'никем (по умолчанию)',
  },
  ebenenNamen: {
    profil: 'профиль',
    nutzer: 'пользователь',
    geraet: 'устройство',
    arbeitsplatz: 'рабочее пространство',
  },
  ebenenErklaerung:
    'Каждая настройка действует на одном из уровней: устройство, пользователь, профиль или рабочее пространство. **Устройство** описывает этот компьютер — путь, размер окна — и не передаётся при переносе настроек. **Пользователь** действует для тебя в каждом из твоих профилей. **Профиль** действует только здесь и переопределяет то, что задано на уровне пользователя. **Рабочее пространство** — это расположение колонок этого окна. Действует то, что задано в профиле, иначе — то, что задано как пользователь, иначе — значение по умолчанию; для политики — самое строгое значение. Права администратора не требуются ни для одной настройки.',
  fachleute: 'Что ещё могут здесь настроить специалисты',
  fachleuteSatz:
    'Тайм-ауты, доступность через сторонние сети, выход наружу и сервисы для других программ. Раскрытие ничего не меняет — оно лишь показывает больше.',
  zuruecksetzen: 'Сбросить на «{vorgabe}»',
  erbeMitWert: '{erbe} ({wert})',
  zuruecksetzenVon: 'Сбросить «{name}» на значение по умолчанию',
  suche: 'Найти настройку',
  suchePlatzhalter: 'Название, ключ или значение',
  sucheTreffer: 'Совпадений: {zahl}',
  sucheKeine: 'Нет настройки, подходящей к «{frage}».',
  zumGebiet: 'Показать в разделе «{gebiet}»',

  // **Below the appearance settings** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  hinweise: {
    ...protokollhinweise,
    'sprache.eigene':
      'Какой собственный перевод использует окно — пусто означает никакой. Если его больше нет, снова действует выбранный язык.',
    feldtiefe:
      'Сколько показывает форма карты. **Обычная** содержит поля, которые ты заполняешь; **Расширенная** добавляет то, что вычисляет скрипт игрового пакета; **Эксперт** добавляет то, что игровой пакет вообще не объясняет, — `stylesheet` и подобное. Три уровня вложены друг в друга: глубже — значит больше, но не иное. Действует для каждого профиля.',
    sprache:
      'Действует для всего окна и для строки меню. То, что приходит из пакета MSE — названия полей, названия игр, — остаётся так, как записано в файле. Исходный язык — немецкий; остальные языки переведены автоматически и ещё не прошли редактуру.',
    darstellung:
      'Как выглядят числа, суммы и даты — например, **de-CH** или **en-GB**. Пусто — как в системе. Язык текстов это не затрагивает.',
    zeitzone:
      'В каком поясе показывается время — например, **Europe/Berlin**. «Как в системе» — пояс компьютера. Хранится для каждого профиля и не передаётся при переносе настроек.',
    'filter.gespeichert':
      'Сохранённые фильтры всех представлений в виде списка JSON — имя, представление, отметки и то, действует **и** или **или**. Создаются они в навигаторе своего представления; принадлежат профилю и передаются вместе с экспортом настроек.',
    'kuerzel.ziffern':
      'Какую страницу открывает цифра с ⌘. Пусто — разделы боковой панели. Настраивается в таблице сочетаний клавиш в группе «Основные».',
    erscheinungsbild:
      'Светлая, тёмная или как решит компьютер. Выбор действует только для этого приложения.',
    schriftgroesse:
      'Масштабирует весь интерфейс, а не только текст — отступы и границы растут вместе с ним. Предпросмотр карты остаётся прежним: он показывает то, что будет напечатано.',
    aktualisierung:
      'По умолчанию **Выкл.**: тогда в сеть ничего не уходит и ничего не запрашивается. При значении **Только уведомлять** приложение при запуске проверяет, есть ли более новая версия, и сообщает об этом — не более. При значении **Уведомлять и загружать** оно дополнительно предлагает кнопку, которая загружает версию в папку загрузок и сверяет её контрольную сумму; сам по себе ничего не устанавливается.',
    ersteschritte:
      'Три шага к первой карте — в окне при запуске. Тот же переключатель есть в самом окне: если выключить его там, ты найдёшь его снова здесь.',
    seitenleiste:
      'Панель слева с разделами приложения. В свёрнутом виде её ширина достаётся рабочему пространству; ⌥⌘S, кнопка на её краю и пункт в меню **Вид** сворачивают и разворачивают её. Хранится для каждого профиля и не передаётся при переносе настроек.',
    leistenlage:
      'Слева она расположена вертикально рядом с содержимым, сверху — строкой над ним. Те же пункты, сочетания клавиш и клавиши-стрелки — только по горизонтали. Хранится для каждого профиля и не передаётся при переносе настроек.',
    fensterstart:
      'По умолчанию **Заполнить экран** — окно занимает доступное место, но **не полноэкранный режим**: он забрал бы строку меню. **Обычный размер** открывает окно во встроенном размере, **Как в прошлый раз** — таким, каким он остался. Хранится для каждого профиля и не передаётся при переносе настроек.',
    fensterbreite:
      'То, что было измерено при последнем закрытии, — **только для «Как в прошлый раз»**. Пусто означает, что прошлого раза ещё не было, и окно открывается во встроенном размере.',
    fensterhoehe:
      'То же для высоты. Оба значения записываются при закрытии окна; размер вне диапазона от 700 до 8192 пунктов не запоминается — окно, не помещающееся на следующем экране, невозможно было бы ухватить.',
    schriftart:
      'Шрифт заголовков и названий наборов. Основной текст остаётся системным шрифтом — декоративный шрифт над абзацем читается плохо. Все четыре входят в состав macOS; ничего не загружается.',
    'barrierefreiheit.bewegung':
      'Переопределение действует **только в одну сторону**: оно отключает движение, но никогда не включает. Кто включил в системе **уменьшение движения**, всё равно не получит движения здесь — приложение, отменяющее это, работало бы против него.',
    'barrierefreiheit.kontrast':
      'То же направление: **Всегда высокая** убирает волокна фона и повышает цвета, пока любой текст не достигнет 7 : 1. Кто запрашивает у системы больше контраста, получает то же самое.',
    'sicherheit.faktorbeimwechsel':
      '**Единственное** переопределение, которое ослабляет защиту, — и только при **смене** профиля. На вход оно не влияет: там второй фактор делает то, для чего он существует. На компьютере с одним человеком одноразовый код при каждой смене — трение без пользы.',
    'vorgabe.zustand':
      'С чего начинается новая строка в коллекции. **Без значения по умолчанию** означает: поле остаётся пустым, как было всегда.',
    'vorgabe.waehrung':
      'С чего начинается поле валюты новой строки, например **EUR**. Ничего не пересчитывается.',
    'vorgabe.einzelspiel':
      'Состоят ли матчи нового турнира из **одной** партии. Это определяет отображение, а не расчёт: процент выигранных партий и процент выигранных партий соперников ведутся в любом случае.',
    'vorgabe.hochladepfad':
      'Папка, в которой открывается диалог открытия файла, — для изображений, MTG JSON и всего остального, что берётся с диска. Полный путь или путь, начинающийся с **~/**; пусто означает — там, где его открывает операционная система.',
    'vorgabe.ausfuhrpfad':
      'Где открывается диалог сохранения, если у действия нет собственного места. Если при сохранении папки нет — например, отключён накопитель, — приложение сообщает об этом, а не сохраняет в другое место.',
    'vorgabe.sicherungspfad':
      'Где открывается диалог резервной копии твоей библиотеки. Отсутствующая папка останавливает резервное копирование с сообщением.',
    'vorgabe.diagnosepfad':
      'Где сохраняются диагностический текст и диагностические пакеты. Отсутствующая папка останавливает действие с сообщением.',
    'bildausfuhr.ordner':
      'Куда в последний раз записывал вывод изображений. Страница «Изображения» предлагает эту папку и записывает сюда каждую заново выбранную — одну для всех сетов, как в MSE. Пусто — диалог выбора папки открывается в папке для сохраняемых файлов.',
    'ausgabe.ziele':
      'Какую цель «Экспорт» в последний раз показывал для каждого сета. Список в JSON, последний выбор в конце, не больше 200; записывается при выборе в «Куда?». Пусто — ничего не выбрано заранее.',
    'dienst.anthropic.adresse':
      'Куда отправляются запросы. Пусто — собственный интерфейс Anthropic. Другой адрес — это **шлюз**: сервис, передающий запрос дальше; его ключ хранится отдельно.',
    'dienst.openai.adresse':
      'То же для OpenAI. Пусто — `https://api.openai.com`. Всё, что введено здесь, означает, что ты разговариваешь уже не с OpenAI, а с тем, что отвечает по этому адресу.',
    'dienst.xai.adresse': 'То же для xAI. Пусто — `https://api.x.ai`.',
    ...netzhilfe,
    ...handelshilfe,
    ...gestaltungshinweise,
    'konto.schau':
      'Может ли эта программа читать твою **собственную** учётную запись на торговой площадке — твои предложения и твои заказы. **Выкл.** означает: даже не заглядывать. Включения самой торговой площадки для этого недостаточно: та отвечает на то, что запрашивают **чужие** продавцы, эта — на то, что находится **у тебя**. Записывать туда никогда ничего не будут — ни предложения, ни корзины, ни покупки.',
    ...kihinweise,
    'anbieter.github.kennung':
      '`client_id` приложения OAuth, созданного в GitHub. Пусто — вход через GitHub не предлагается.',
    'anbieter.google.kennung':
      '`client_id` приложения OAuth, созданного в Google. Обычно он заканчивается на `apps.googleusercontent.com`.',
    'anbieter.microsoft.kennung':
      'Идентификатор приложения из портала Microsoft Entra — UUID. Microsoft **не** требует секрета для этого способа.',
    'anbieter.x.kennung':
      '`client_id` приложения, созданного на портале разработчиков X.',
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
    'konto.schau.aus': 'Выкл.',
    'konto.schau.an': 'Вкл.',
    ...kiwerte,
    'sprache.erben': 'Как у пользователя',
    'erscheinungsbild.erben': 'Как у пользователя',
    'schriftgroesse.erben': 'Как у пользователя',
    'schriftart.erben': 'Как у пользователя',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': 'Как в системе',
    'erscheinungsbild.hell': 'Светлая',
    'erscheinungsbild.dunkel': 'Тёмная',
    'schriftgroesse.klein': 'Маленький',
    'schriftgroesse.normal': 'Обычный',
    'schriftgroesse.gross': 'Крупный',
    'schriftgroesse.sehrgross': 'Очень крупный',
    'aktualisierung.aus': 'Выкл.',
    'aktualisierung.melden': 'Только уведомлять',
    'aktualisierung.holen': 'Уведомлять и загружать',
    'ersteschritte.an': 'Вкл.',
    'ersteschritte.aus': 'Выкл.',
    'seitenleiste.sichtbar': 'Показана',
    'seitenleiste.verborgen': 'Скрыта',
    'leistenlage.senkrecht': 'Слева',
    'leistenlage.waagerecht': 'Сверху',
    'fensterstart.maximiert': 'Заполнить экран',
    'fensterstart.normal': 'Обычный размер',
    'fensterstart.zuletzt': 'Как в прошлый раз',

    'feldtiefe.normal': 'Обычная',
    'feldtiefe.erweitert': 'Расширенная',
    'feldtiefe.experte': 'Эксперт',
    'tisch.autoanordnung.an': 'Да — стол раскладывает их сам',
    'tisch.autoanordnung.aus': 'Нет — что положено, то остаётся на месте',
    'tisch.beschriftung': 'Подписи на столе',
    'tisch.beschriftung.aus': 'Выкл. — только количество',
    'tisch.beschriftung.klein': 'Мелкие',
    'tisch.beschriftung.normal': 'Средние',
    'tisch.beschriftung.gross': 'Крупные',
    'schriftart.pergament': 'Пергамент — спокойный и читаемый',
    'schriftart.fantasy': 'Фэнтези — игривый',
    'schriftart.antike': 'Античность — римские заглавные буквы',
    'schriftart.kalligrafie': 'Каллиграфия — написанный от руки',
    'barrierefreiheit.bewegung.system': 'Как в системе',
    'barrierefreiheit.bewegung.reduziert': 'Всегда уменьшено',
    'barrierefreiheit.kontrast.system': 'Как в системе',
    'barrierefreiheit.kontrast.hoch': 'Всегда высокая',
    'sicherheit.faktorbeimwechsel.verlangen': 'Также при смене',
    'sicherheit.faktorbeimwechsel.nurAnmelden': 'Только при входе',
    'vorgabe.zustand.ohne': 'Без значения по умолчанию',
    'vorgabe.zustand.mint': 'Mint — идеальное',
    'vorgabe.zustand.nearMint': 'Near Mint — почти идеальное',
    'vorgabe.zustand.excellent': 'Excellent — отличное',
    'vorgabe.zustand.good': 'Good — хорошее',
    'vorgabe.zustand.lightPlayed': 'Light Played — слегка игранное',
    'vorgabe.zustand.played': 'Played — игранное',
    'vorgabe.zustand.poor': 'Poor — плохое',
    'vorgabe.einzelspiel.aus': 'Выкл.',
    'vorgabe.einzelspiel.an': 'Вкл.',

    // **The workspace and the limit** — Werkstück O3 of Phase 5.2 (#1427).
    'arbeitsplatz.navigator.erben': 'Как у пользователя',
    'arbeitsplatz.navigatorbreite.erben': 'Как у пользователя',
    'arbeitsplatz.inspektorbreite.erben': 'Как у пользователя',
    'arbeitsplatz.navigator.sichtbar': 'Показана',
    'arbeitsplatz.navigator.verborgen': 'Скрыта',
    'arbeitsplatz.inspektor.erben': 'Как у пользователя',
    'arbeitsplatz.inspektor.sichtbar': 'Показана',
    'arbeitsplatz.inspektor.verborgen': 'Скрыта',
    'richtlinie.nurlokal.aus': 'Выкл.',
    'richtlinie.nurlokal.an': 'Вкл.',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Where a login comes from in the first place** — follow-up to Werkstück 4
  // of Phase 4.12 (#1351).
  anmeldung: 'Вход через провайдера',
  anmeldungHinweis:
    'Это приложение **не поставляется с учётными данными**. Чтобы входить через GitHub, Google, Microsoft или X, создай у провайдера собственное **приложение OAuth** и введи здесь его `client_id`. В качестве адреса перенаправления укажи адрес, который отображается в браузере при входе, — он указывает на этот компьютер и действует, только пока идёт вход.',

  anmeldungGeheimnis:
    'GitHub дополнительно требует `client_secret` — у Google он необязателен, а Microsoft запрещает его для программ, подобных этой. Он помещается **в связку ключей**, а не сюда: библиотеку передают другим, а связку ключей — нет. Создай его в программе **«Связка ключей»** как **пароль для программ** с именем `MagicSetEditor4Mac-Anmeldung` и учётной записью `github`. В командной строке: `security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`.',

  geheimnis: 'Секрет',
  geheimnisPflicht: 'требуется, но отсутствует',
  geheimnisVerboten:
    'не требует секрета и отклоняет его — здесь достаточно идентификатора клиента.',

  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  wirktNachNeustart: 'Действует после перезапуска.',

  trefferInGruppe: 'Совпадений для «{frage}»: {zahl}',
  gesuchtMarke: 'найдено',
  sucheSprung: 'Перейти к первому совпадению',

  vorgaben: 'Значения по умолчанию',

  // **The pointer to the help** — Werkstück 13 of Phase 4.12 (#1351).

  dienste: 'Сторонние сервисы карт',
  diensteHinweis:
    'Выкл. означает: ни одного запроса к этому сервису не уходит. Приложение продолжает работать без любого сервиса — просто не будет того, что приходит извне.',

  kiAnbieter: 'Языковые модели',
  kiAnbieterHinweis:
    'Куда можно отправлять вопросы. Каждый провайдер вычисляет на своих машинах и видит то, что ты отправляешь, — поэтому каждый включается отдельно, и ни один не включён сам по себе.',
  // See the German catalogue for why these two live here and not under
  // `werte`: they are the same for every service.
  dienstAus: 'Выкл.',
  dienstAn: 'Вкл.',
  dienstUngeprueft: 'Ни разу не запускалось с настоящим сервисом',
  dienstUngeprueftHinweis:
    'Для этого сервиса здесь нет учётной записи. Проверена наша сторона границы — запрос, заголовки, сопоставление. Отвечает ли сервис так, как мы предполагаем, не измерено.',
  fragen: 'Кто на что отвечает',
  fragenHinweis:
    'Если несколько сервисов могут ответить на один и тот же вопрос, здесь решается, какой из них спрашивать. **Запасной вариант не подставляется**: если выбранный молчит, тебе об этом скажут — вместо того чтобы молча показать ответ из другого источника.',
  'frage.suche': 'Поиск карт',

  // The service catalogue — work item A1 of phase 5.3 (#1440).
  ...dienstkatalog,

  dienstSchluesselFehlt: 'Учётные данные не сохранены',
  dienstSchluesselHinweis:
    'Этому сервису нужны учётные данные. Они хранятся в связке ключей macOS, а не в файле библиотеки — его передают другим.',
  zugaenge: 'Учётные данные',
  zugaengeHinweis:
    'Они хранятся в связке ключей macOS и **не** в файле библиотеки — его передают другим. Здесь показано лишь, **сохранено ли** что-то; само значение никогда не возвращается.',
  // See the German catalogue: these name the headers each service asks for
  // (documented in decision 0056).

  kuerzelUeberschrift: 'Сочетания клавиш',
  ...einstellungsbelegung,
  kuerzelHinweis:
    'Они действуют во всём окне. В текстовом поле ⌘Z и ⌘F принадлежат полю.',
  kuerzelZweiteReihe:
    'Восемь сочетаний с ⌥ расположены на клавишах верхнего ряда, слева направо. ' +
    'На клавиатуре с другой раскладкой — например, AZERTY — там напечатан другой ' +
    'символ; клавиша та же.',
  kuerzelTaste: 'Клавиша',
  kuerzelHandlung: 'Действие',

  austausch,
  eigenuebersetzung,
  // **An account's remaining quota at a provider** — A5, #1444.
  kontingentTitel: 'Квота у провайдера',
  kontingentErklaerung:
    'Сколько кредитов входит в тариф и сколько из них осталось. Запрос уходит в сеть и сам стоит один кредит.',
  kontingentHolen: 'Проверить квоту',
  kontingentHolenBei: 'Проверить квоту в {dienst}',
  kontingentLaeuft: 'Проверяется …',
  kontingentZahlen: 'Осталось кредитов: {rest} из {gesamt}.',
  kontingentFrist:
    'Эта цифра у провайдера может отставать до {minuten} мин — это не жёсткий лимит. Расходы здесь по-настоящему ограничивает бюджет этой библиотеки.',
  kontingentUeberzug: 'Кредит сверх тарифа стоит {satz}.',
  // **What is known about AI models** — A6, #1445.
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}
