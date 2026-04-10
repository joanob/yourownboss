import React, {useEffect, useState} from 'react'
import { useRoute } from 'wouter'
import api from '../lib/api'

type ResourceRange = { resource_id: number; min_price: number; max_price: number; step: number; is_output: boolean }
type ProcRes = { resource_id: number; is_output: boolean; name: string }
type SimulationResult = any

const SimulationPage: React.FC = () => {
  const [match, params] = useRoute('/simulate/:pid')
  const pid = parseInt(params?.pid || '0', 10)
  const [resources, setResources] = useState<ProcRes[]>([])
  const [ranges, setRanges] = useState<ResourceRange[]>([])
  const [timeMin, setTimeMin] = useState(1000)
  const [timeMax, setTimeMax] = useState(5000)
  const [timeStep, setTimeStep] = useState(1000)
  const [running, setRunning] = useState(false)
  const [history, setHistory] = useState<SimulationResult[]>([])
  const [benefitFilter, setBenefitFilter] = useState<{min:number; max:number}|null>(null)
  const [resourcePriceFilters, setResourcePriceFilters] = useState<Record<number, {min:number; max:number}>>({})
  const [sortDesc, setSortDesc] = useState(true)
  const [expandedSims, setExpandedSims] = useState<number[]>([])

  useEffect(() => {
    if (!pid) return
    let mounted = true
    api.get(`/processes/${pid}/resources`).then((res) => {
      if (!mounted) return
      setResources(res.data || [])
      setRanges((res.data || []).map((r: any) => ({resource_id: r.resource_id, min_price: 0, max_price: 10, step: 1, is_output: r.is_output})))
    }).catch(() => {})

    api.get(`/simulations?process_id=${pid}`).then((res) => { if (!mounted) setHistory(res.data || []) }).catch(()=>{})


    return () => { mounted = false }
  }, [pid])

  // derive default filters from history when it changes
  useEffect(() => {
    if (!history || history.length === 0) {
      setBenefitFilter(null)
      setResourcePriceFilters({})
      return
    }
    let minB = Number.POSITIVE_INFINITY
    let maxB = Number.NEGATIVE_INFINITY
    const rp: Record<number, {min:number; max:number}> = {}
    for (const h of history) {
      const b = Number(h.benefit_per_hour) || 0
      if (b < minB) minB = b
      if (b > maxB) maxB = b
      for (const r of (h.resources || [])) {
        const id = Number(r.resource_id)
        const price = Number(r.price)
        if (!rp[id]) rp[id] = {min: price, max: price}
        else {
          if (price < rp[id].min) rp[id].min = price
          if (price > rp[id].max) rp[id].max = price
        }
      }
    }
    if (minB === Number.POSITIVE_INFINITY) { minB = 0; maxB = 0 }
    setBenefitFilter({min: Math.floor(minB), max: Math.ceil(maxB)})
    setResourcePriceFilters(rp)
  }, [history])

  const applyFilters = (items: SimulationResult[]) => {
    if (!items) return []
    return items.filter((h) => {
      const b = Number(h.benefit_per_hour) || 0
      if (benefitFilter) {
        if (b < benefitFilter.min || b > benefitFilter.max) return false
      }
      for (const [ridStr, fr] of Object.entries(resourcePriceFilters)) {
        const rid = Number(ridStr)
        const found = (h.resources || []).find((rr: any) => Number(rr.resource_id) === rid)
        if (!found) return false
        const price = Number(found.price)
        if (price < fr.min || price > fr.max) return false
      }
      return true
    })
  }

  const filtered = applyFilters(history).sort((a,b) => sortDesc ? (Number(b.benefit_per_hour) - Number(a.benefit_per_hour)) : (Number(a.benefit_per_hour) - Number(b.benefit_per_hour)))

  const setRangeField = (index:number, field: keyof ResourceRange, value:any) => {
    setRanges(rs => rs.map((r,i)=> i===index ? {...r, [field]: value} : r))
  }

  const start = async () => {
    if (!pid) return
    const body = {
      process_id: pid,
      time_min_ms: timeMin,
      time_max_ms: timeMax,
      time_step_ms: timeStep,
      resource_ranges: ranges.map(r => ({resource_id: r.resource_id, min_price: r.min_price, max_price: r.max_price, step: r.step, is_output: r.is_output}))
    }
    setRunning(true)
    try {
      await api.post('/simulations', body)
      alert('Simulación iniciada (202 Accepted).')
    } catch (err) {
      console.error(err)
      alert('Error iniciando simulación')
    } finally { setRunning(false) }
  }

  return (
    <div>
      <h2>Simulación del proceso {pid}</h2>
      <div style={{display:'flex', gap:8}}>
        <label>Time min ms: <input type="number" value={timeMin} onChange={(e)=>setTimeMin(parseInt(e.target.value||'0',10))} /></label>
        <label>Time max ms: <input type="number" value={timeMax} onChange={(e)=>setTimeMax(parseInt(e.target.value||'0',10))} /></label>
        <label>Time step ms: <input type="number" value={timeStep} onChange={(e)=>setTimeStep(parseInt(e.target.value||'0',10))} /></label>
      </div>

      <h3>Rangos de precio por recurso</h3>
      <table>
        <thead><tr><th>Recurso</th><th>Min</th><th>Max</th><th>Step</th><th>Output</th></tr></thead>
        <tbody>
          {ranges.map((r, i) => (
            <tr key={r.resource_id}>
              <td>{resources[i]?.name ?? r.resource_id} ({r.resource_id})</td>
              <td><input type="number" value={r.min_price} onChange={(e)=>setRangeField(i,'min_price', parseInt(e.target.value||'0',10))} /></td>
              <td><input type="number" value={r.max_price} onChange={(e)=>setRangeField(i,'max_price', parseInt(e.target.value||'0',10))} /></td>
              <td><input type="number" value={r.step} onChange={(e)=>setRangeField(i,'step', parseInt(e.target.value||'1',10))} /></td>
              <td>{r.is_output ? 'Sí' : 'No'}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div style={{marginTop:12}}>
        <button onClick={start} disabled={running}>{running ? 'Iniciando...' : 'Iniciar simulación'}</button>
      </div>

      <h3 style={{marginTop:20}}>Simulaciones pasadas</h3>
      <div>
        {(!history || history.length === 0) ? <p>No hay resultados todavía.</p> : (
          <div>
            <div style={{display:'flex', gap:12, alignItems:'center', marginBottom:12}}>
              <div>
                <strong>Beneficio</strong>
                {benefitFilter ? (
                  <div style={{display:'flex', gap:6, alignItems:'center'}}>
                    <label>Min: <input type="number" value={benefitFilter.min} onChange={(e)=> setBenefitFilter({min: Number(e.target.value), max: benefitFilter.max})} /></label>
                    <label>Max: <input type="number" value={benefitFilter.max} onChange={(e)=> setBenefitFilter({min: benefitFilter.min, max: Number(e.target.value)})} /></label>
                  </div>
                ) : null}
              </div>
              <div>
                <strong>Orden</strong>
                <div>
                  <label><input type="checkbox" checked={sortDesc} onChange={(e)=>setSortDesc(e.target.checked)} /> Ordenar por beneficio descendente</label>
                </div>
              </div>
            </div>

            <div style={{marginBottom:12}}>
              <strong>Filtros por recurso</strong>
              <div style={{display:'flex', gap:12, flexWrap:'wrap', marginTop:6}}>
                {resources.map((resrc) => {
                  const rid = resrc.resource_id
                  const fr = resourcePriceFilters[rid] || {min: 0, max: 0}
                  return (
                    <div key={rid} style={{border:'1px solid #eee', padding:8}}>
                      <div>{resrc.name} ({rid})</div>
                      <div style={{display:'flex', gap:6}}>
                        <label>Min: <input type="number" value={fr.min} onChange={(e)=> setResourcePriceFilters(prev => ({...prev, [rid]: {min: Number(e.target.value), max: prev[rid]?.max ?? fr.max}}))} /></label>
                        <label>Max: <input type="number" value={fr.max} onChange={(e)=> setResourcePriceFilters(prev => ({...prev, [rid]: {min: prev[rid]?.min ?? fr.min, max: Number(e.target.value)}}))} /></label>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>

            <table style={{width:'100%'}}>
              <thead><tr><th></th><th>ID</th><th>time_ms</th><th>benefit_per_hour</th></tr></thead>
              <tbody>
                {filtered.map((h: any) => (
                  <React.Fragment key={h.id}>
                    <tr>
                      <td style={{width:40}}>
                        <button onClick={() => {
                          setExpandedSims(prev => prev.includes(h.id) ? prev.filter(x=>x!==h.id) : [...prev, h.id])
                        }}>{expandedSims.includes(h.id) ? '▾' : '▸'}</button>
                      </td>
                      <td>{h.id}</td>
                      <td>{h.time_ms}</td>
                      <td>{h.benefit_per_hour}</td>
                    </tr>
                    {expandedSims.includes(h.id) ? (
                      <tr>
                        <td colSpan={4}>
                          <div style={{padding:8, background:'#fafafa', border:'1px solid #eee'}}>
                            <strong>Recursos</strong>
                            <table style={{width:'100%', marginTop:6}}>
                              <thead><tr><th>Recurso</th><th>is_output</th><th>price</th><th>quantity</th></tr></thead>
                              <tbody>
                                {(h.resources || []).map((rr: any) => {
                                  const name = resources.find(r=>r.resource_id === Number(rr.resource_id))?.name || rr.resource_id
                                  return <tr key={String(rr.resource_id)+String(rr.is_output)}><td>{name}</td><td>{rr.is_output ? 'Sí' : 'No'}</td><td>{rr.price}</td><td>{rr.quantity}</td></tr>
                                })}
                              </tbody>
                            </table>
                          </div>
                        </td>
                      </tr>
                    ) : null}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

export default SimulationPage
