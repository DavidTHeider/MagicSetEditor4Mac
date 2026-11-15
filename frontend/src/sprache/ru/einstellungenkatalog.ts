// **Die Texte des Dienstkatalogs** — die englische Seite zu
// [`de/einstellungenkatalog.ts`](../de/einstellungenkatalog.ts).
export const dienstkatalog = {
  katalog: 'Что умеют сервисы',
  katalogHinweis:
    'Для каждого сервиса — вопросы, для которых здесь построен путь, и открыт ли он **сейчас**. Эта справка не отправляет запрос: она считывает переключатели, связку ключей и роль. Проверка на реальном сервисе действительна для того дня и той версии, в которых она проводилась.',
  katalogFragen: {
    suche: 'Поиск карт',
    kartendaten: 'Данные карт',
    bild: 'Изображения карт',
    regeln: 'Разъяснения правил',
    manakosten: 'Проверка стоимости маны',
    setangaben: 'Сведения о сете',
    preise: 'Цены',
    sammeldaten: 'Сводные данные',
  },
  katalogStatus: {
    bereit: 'готов',
    aus: 'выключен',
    zugangFehlt: 'нет доступа',
    gestoert: 'недоступен',
    gesperrt: 'заблокирован',
    nichtGewaehlt: 'не выбран',
  },
  katalogAngabeFehlt: '{angabe} не задано',
  katalogNachweis:
    'Проверка на реальном сервисе проведена {datum}, версия {fassung} —',
  katalogNachweisAktuell: 'для этой версии.',
  katalogNachweisFremd: 'не для этой версии.',
  katalogDiagnose: 'Диагностика',
  katalogLauf: 'Проверку можно повторить: {lauf}',
  katalogImLauf: 'входит в проверку',
  katalogNichtImLauf: 'не входит в проверку',
  katalogVertrag: 'Договор и условия',
  katalogFelder: {
    basisadresse: 'Адрес',
    endpunkte: 'Конечные точки',
    zugang: 'Доступ',
    tarif: 'Тариф',
    grenze: 'Ограничение',
    frist: 'Срок хранения в кеше',
    wirkung: 'Действие',
    bedingungen: 'Условия',
    abgerufen: 'Проверено',
    beleg: 'Источник',
  },
  katalogWirkung: {
    externSenden: 'отправляет данные стороннему сервису',
    externSchreiben: 'записывает данные в стороннем сервисе',
  },
}
