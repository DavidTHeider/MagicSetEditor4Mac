// The research texts — work package A16 of phase 5.3
// (#1515).
export const recherche = {
  rechercheTitel: '查询',
  rechercheErklaerung:
    '在这台电脑上查询卡牌文字、裁决和规则——全文与向量合并检索。嵌入模型会参与，且在本地运行；不会调用任何写作文本的模型。没有任何内容会离开这台电脑，也不会更改任何内容。',

  rechercheFall: '查询目的',
  rechercheFallKarten: '卡牌',
  rechercheFallRegeln: '规则',
  rechercheFallRulings: '裁决',

  rechercheFrage: '问题',
  rechercheSuchen: '查询',
  rechercheSprache: '语言',
  rechercheAlleSprachen: '任意',

  rechercheModus: '检索方式',
  rechercheModusHybrid: '全文与向量合并',
  rechercheModusLexikalisch: '仅全文——没有嵌入模型',
  rechercheModusOhneIndex: '未检索——没有搜索索引',

  rechercheLeereNichtsGefunden: '换个词试试？',
  rechercheLeereOhneIndex: '可在“设置 › 诊断”下建立索引。',
  rechercheLeereOhneModell:
    '没有嵌入模型时，搜索只能找到词语，找不到相似内容。',
  rechercheLeereOhneQuelle: '这个问题需要索引中没有的来源。',
  rechercheLeereGesperrt: '尚未获准访问此来源。',
  rechercheLeereBudget: '本次运行已达到上限。',

  rechercheHerkunft: '修订版 {revision}，有效期 {datum}，获取于 {abgerufen}',
  rechercheWeitere: '同一张卡牌的另外 {anzahl} 个部分',
  rechercheOeffnen: '打开卡牌',
  rechercheEinschraenkung: '未完成',
  rechercheStaende: '修订版',
  rechercheIndex: '搜索索引',
  rechercheMehr: '载入更多',

  aehnlichTitel: '相似卡牌',
  aehnlichErklaerung:
    '按这张卡牌的文字而非名称搜索。这张卡牌本身不会出现在列表中。',
  aehnlichSuchen: '查找相似卡牌',
  aehnlichLeer: '尚未搜索。',

  freigabeTitel: '模型将获得的内容',
  freigabeErklaerung:
    '整理出模型将会看到的内容——但不会发送。之后再生成的人，就知道自己看过了什么。',
  freigabeAufgabe: '任务',
  freigabeAnbieter: '提供方',
  freigabeModell: '模型',
  freigabeFenster: '上下文窗口',
  freigabeReserve: '输出预留',
  freigabeZeigen: '整理',
  freigabeOrtLokal: '在这台电脑上计算。',
  freigabeOrtCloud: '在外部提供方处计算。',
  freigabeBudget: '{verbraucht} / {grenze} 个词元，预留 {reserve}',
  freigabeZaehlweise: '按 {zaehlweise} 估算',
  freigabeZieladresse: '目标地址',
  freigabeKategorien: '数据类别',
  freigabePrivat: '包含你自己的内容。',
  freigabeOeffentlich: '仅限公开来源。',
  freigabeBelege: '引用依据',
  freigabeKeineBelege: '没有引用依据——如此整理后，模型将看不到任何内容。',
  freigabeVeraltet: '不是当前生效的修订版',
  freigabeLaeuftAb: '有效期至 {zeit}',

  // **Releasing your own collection** — A13.4
  // (#1452).
  freigabeEigeneKarten: '同时搜索我自己的卡牌',
  freigabeBestandErklaerung:
    '不勾选此项时，你自己的卡牌不会参与——搜索只涵盖公开来源，回答中也会注明这一点。如果模型在别处运行，随之发送的内容会离开这台电脑。',
}
