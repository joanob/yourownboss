import api from './api'

type Resource = { id: number; name: string }

let cachedPromise: Promise<Resource[] | null> | null = null

export const getResources = (): Promise<Resource[]> => {
  if (!cachedPromise) {
    cachedPromise = api
      .get('/resources')
      .then((res) => (Array.isArray(res.data) ? res.data : []))
      .catch((err) => {
        // clear cache on error so future callers can retry
        cachedPromise = null
        throw err
      })
  }
  // cachedPromise is Promise<Resource[] | null> but callers expect Resource[]
  return cachedPromise as Promise<Resource[]>
}

export const invalidateResources = () => {
  cachedPromise = null
}

export default { getResources, invalidateResources }
