export type DictionaryEntry = {
  Dictionary: string
  Headword: string
  HTML: string
  CSS: string
}

const browserEntries: Record<string, DictionaryEntry> = {
  '你好': {
    Dictionary: 'CC-CEDICT',
    Headword: '你好',
    HTML: '<p>hello</p><p>hi</p>',
    CSS: '.dictionary-entry { font-family: Georgia, serif; }',
  },
  hello: {
    Dictionary: 'Moyan Sample',
    Headword: 'hello',
    HTML: '<p>a greeting used when meeting someone</p>',
    CSS: '',
  },
}

type WailsApp = {
  LookupWord(word: string): Promise<DictionaryEntry>
  LookupWords(word: string): Promise<DictionaryEntry[]>
  SearchCandidates(word: string): Promise<string[]>
  ChooseAndOpenDictionary(): Promise<string>
}

declare global {
  interface Window {
    go?: { main?: { App?: WailsApp } }
  }
}

export async function chooseAndOpenDictionary(): Promise<string | null> {
  const app = window.go?.main?.App
  if (!app) return null
  return app.ChooseAndOpenDictionary()
}

export async function searchCandidates(word: string): Promise<string[]> {
  const app = window.go?.main?.App
  if (app) return app.SearchCandidates(word)
  return Object.keys(browserEntries).filter((entry) => entry.startsWith(word.trim())).slice(0, 8)
}

export async function lookupWords(word: string): Promise<DictionaryEntry[]> {
  const app = window.go?.main?.App
  if (app) {
    return app.LookupWords(word)
  }
  const entry = browserEntries[word.trim()]
  return entry ? [entry] : []
}

export async function lookupWord(word: string): Promise<DictionaryEntry | null> {
  const entries = await lookupWords(word)
  return entries[0] ?? null
}
