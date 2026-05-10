import React, {useEffect, useState, useRef} from 'react'
import api from '../lib/api'
import { getResources, invalidateResources } from '../lib/resourceCache'

type Resource = { id: number; name: string }

type Proc = { id?: number | null; name: string; resources: Array<{resource_id:number; is_output:boolean}> }
type Building = { id: number; name: string; processes: Proc[] }

const ResourcesPage: React.FC = () => {
  const [resources, setResources] = useState<Resource[]>([])
  const fetchRef = useRef(false)
  const [saving, setSaving] = useState(false)
  const [buildings, setBuildings] = useState<Building[]>([])
  const [expandedResources, setExpandedResources] = useState<number[]>([])
  const [procSims, setProcSims] = useState<Record<number, any[]>>({})
  const [resourcePriceFilter, setResourcePriceFilter] = useState<Record<number, {min:number; max:number}>>({})

  useEffect(() => {
    // fetch once on mount and avoid duplicate fetches in StrictMode (dev)
    if (fetchRef.current) return
    fetchRef.current = true
    getResources()
      .then((data) => {
        setResources(data || [])
      })
      .catch((err) => {
        console.error('GET /resources error:', err)
      })
  }, [])

  const setId = (index: number, value: number) => {
    setResources((prev) => {
      const copy = [...prev]
      copy[index] = {...copy[index], id: value}
      return copy
    })
  }

  const setName = (index: number, value: string) => {
    setResources((prev) => {
      const copy = [...prev]
      copy[index] = {...copy[index], name: value}
      return copy
    })
  }

  const nextFreeId = (list: Resource[]) => {
    const ids = new Set(list.map((r) => Math.max(0, r.id)))
    let i = 1
    while (ids.has(i)) i++
    return i
  }

  const addRow = () => setResources((p) => {
    const id = nextFreeId(p)
    return [...p, {id, name: ''}]
  })

  useEffect(() => {
    let mounted = true
    api.get('/buildings').then((res) => {
      if (!mounted) return
      setBuildings(res.data || [])
    }).catch(() => {}).finally(() => {})
    return () => { mounted = false }
  }, [])

  const removeRow = (index: number) => setResources((p) => p.filter((_, i) => i !== index))

  const saveAll = async () => {
    setSaving(true)
    try {
      await api.post('/resources/upsert', {resources})
      // refresh cache and local state
      invalidateResources()
      try {
        const fresh = await getResources()
        setResources(fresh || [])
      } catch (e) {
        // ignore — we already saved successfully
      }
      alert('Guardado OK')
    } catch (err: any) {
      console.error(err)
      alert('Error guardando. Revisa la consola.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <h2>Resources</h2>
      <p>Lista de recursos — editar los campos y pulsar <strong>Guardar</strong>.</p>

      <table style={{width: '100%', borderCollapse: 'collapse'}}>
        <thead>
          <tr>
            <th style={{textAlign: 'left', padding: 6}}>ID</th>
            <th style={{textAlign: 'left', padding: 6}}>Name</th>
            <th style={{padding: 6}}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {resources.map((r, i) => (
            <React.Fragment key={r.id ?? i}>
              <tr style={{borderTop: '1px solid #eee'}}>
                <td style={{padding: 6}}>
                  <div style={{display:'flex', alignItems:'center', gap:8}}>
                    <button onClick={() => setExpandedResources(prev => prev.includes(r.id) ? prev.filter(x => x !== r.id) : [...prev, r.id])}>
                      {expandedResources.includes(r.id) ? '▾' : '▸'}
                    </button>
                    <input type="number" value={r.id} onChange={(e) => setId(i, parseInt(e.target.value || '0', 10))} />
                  </div>
                </td>
                <td style={{padding: 6}}>
                  <input value={r.name} onChange={(e) => setName(i, e.target.value)} />
                </td>
                <td style={{padding: 6}}>
                  <button onClick={() => removeRow(i)}>Eliminar</button>
                </td>
              </tr>
              {expandedResources.includes(r.id) ? (
                <tr>
                  <td colSpan={3} style={{background:'#fafafa', padding:8}}>
                    <div style={{display:'flex', gap:12, alignItems:'center', marginBottom:8}}>
                      <div>
                        <strong>Filtros precio</strong>
                        <div style={{display:'flex', gap:8}}>
                          <label>Min: <input type="number" value={(resourcePriceFilter[r.id]?.min ?? 0)} onChange={(e)=> setResourcePriceFilter(prev => ({...prev, [r.id]: {min: Number(e.target.value), max: prev[r.id]?.max ?? 0}}))} /></label>
                          <label>Max: <input type="number" value={(resourcePriceFilter[r.id]?.max ?? 0)} onChange={(e)=> setResourcePriceFilter(prev => ({...prev, [r.id]: {min: prev[r.id]?.min ?? 0, max: Number(e.target.value)}}))} /></label>
                        </div>
                      </div>
                      <div>
                        <strong>Procesos que usan este recurso</strong>
                      </div>
                    </div>
                    <div>
                      {buildings.flatMap(b => b.processes.map(p => ({building: b, process: p}))).filter(bp => (bp.process.resources || []).some(pr => Number(pr.resource_id) === r.id)).map(bp => {
                        const p = bp.process
                        const pid = Number(p.id || 0)
                        const isOutput = (p.resources || []).find(pr => Number(pr.resource_id) === r.id)?.is_output
                        const sims = procSims[pid] || null
                        return (
                          <div key={String(bp.building.id)+String(pid)} style={{borderTop:'1px solid #eee', paddingTop:8, marginTop:8}}>
                            <div style={{display:'flex', justifyContent:'space-between', alignItems:'center'}}>
                              <div>{bp.building.name} / {p.name} ({pid}) — {isOutput ? 'Salida' : 'Entrada'}</div>
                              <div style={{display:'flex', gap:8}}>
                                <button onClick={async ()=>{
                                  if (!pid) return
                                  if (procSims[pid]) return
                                  try {
                                    const res = await api.get(`/simulations?process_id=${pid}`)
                                    setProcSims(prev => ({...prev, [pid]: res.data || []}))
                                  } catch (e) {
                                    setProcSims(prev => ({...prev, [pid]: []}))
                                  }
                                }}>Cargar simulaciones</button>
                                <a href={`/simulate/${pid}`}><button>Ir a simulación</button></a>
                              </div>
                            </div>
                            {sims ? (
                              <div style={{marginTop:8}}>
                                <table style={{width:'100%'}}>
                                  <thead><tr><th>ID</th><th>time_ms</th><th>benefit_per_hour</th><th>price (para recurso)</th></tr></thead>
                                  <tbody>
                                    {sims.filter((h:any) => {
                                      const found = (h.resources||[]).find((rr:any)=> Number(rr.resource_id) === r.id)
                                      if (!found) return false
                                      const price = Number(found.price)
                                      const fr = resourcePriceFilter[r.id]
                                      if (fr) {
                                        if (price < fr.min || price > fr.max) return false
                                      }
                                      return true
                                    }).map((h:any) => {
                                      const found = (h.resources||[]).find((rr:any)=> Number(rr.resource_id) === r.id) || {price: 0}
                                      return <tr key={h.id}><td>{h.id}</td><td>{h.time_ms}</td><td>{h.benefit_per_hour}</td><td>{found.price}</td></tr>
                                    })}
                                  </tbody>
                                </table>
                                <div style={{marginTop:6}}>
                                  <strong>Mejor precio (según beneficio):</strong>
                                  {(() => {
                                    const all = (sims || []).map((h:any) => ({h, found: (h.resources||[]).find((rr:any)=> Number(rr.resource_id) === r.id)})).filter(x => x.found)
                                    const fr = resourcePriceFilter[r.id]
                                    const filtered = all.filter(x => {
                                      const price = Number(x.found.price)
                                      if (!fr) return true
                                      return price >= fr.min && price <= fr.max
                                    })
                                    if (filtered.length === 0) return <span> No hay simulaciones cargadas o no hay dentro del rango.</span>
                                    let best = filtered[0]
                                    for (const it of filtered) {
                                      if (Number(it.h.benefit_per_hour) > Number(best.h.benefit_per_hour)) best = it
                                    }
                                    return <span> Precio: {best.found.price} — Benefit/h: {best.h.benefit_per_hour}</span>
                                  })()}
                                </div>
                              </div>
                            ) : null}
                          </div>
                        )
                      })}
                    </div>
                  </td>
                </tr>
              ) : null}
            </React.Fragment>
          ))}
        </tbody>
      </table>

      <div style={{marginTop: 12, display: 'flex', gap: 8}}>
        <button onClick={addRow}>Añadir fila</button>
        <button onClick={saveAll} disabled={saving}>{saving ? 'Guardando...' : 'Guardar'}</button>
      </div>
    </div>
  )
}

export default ResourcesPage
