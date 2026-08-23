export type DictionaryEntry = {
  Dictionary: string
  Headword: string
  HTML: string
  CSS: string
}

export type PackageInfo = {
  Path: string
  Name: string
}

const browserEntries: Record<string, DictionaryEntry> = {
  hello: {
    Dictionary: 'Moyan Sample',
    Headword: 'hello',
    HTML: '<p>a greeting used when meeting someone</p>',
    CSS: '',
  },
}

type WailsApp = {
  LookupWord(word: string): Promise<DictionaryEntry>
  LookupIn(path: string, word: string): Promise<DictionaryEntry>
  SearchCandidates(word: string): Promise<string[]>
  ChooseAndOpenDictionary(): Promise<string>
  Library(): Promise<PackageInfo[]>
  RestoreLibrary(): Promise<void>
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

export async function lookupWord(word: string): Promise<DictionaryEntry | null> {
  const app = window.go?.main?.App
  if (app) {
    try {
      return await app.LookupWord(word)
    } catch {
      return null
    }
  }
  return browserEntries[word.trim()] ?? null
}

export async function lookupIn(path: string, word: string): Promise<DictionaryEntry | null> {
  const app = window.go?.main?.App
  if (app?.LookupIn) {
    try {
      return await app.LookupIn(path, word)
    } catch {
      return null
    }
  }
  return lookupWord(word)
}

export async function listLibrary(): Promise<PackageInfo[]> {
  const app = window.go?.main?.App
  if (!app?.Library) return []
  return app.Library()
}

export async function restoreLibrary(): Promise<void> {
  const app = window.go?.main?.App
  if (!app?.RestoreLibrary) return
  await app.RestoreLibrary()
}
