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
        {history.length === 0 ? <p>No hay resultados todavía.</p> : (
          <table style={{width:'100%'}}>
            <thead><tr><th>ID</th><th>time_ms</th><th>benefit_per_hour</th></tr></thead>
            <tbody>
              {history.map((h: any) => (
                <tr key={h.id}><td>{h.id}</td><td>{h.time_ms}</td><td>{h.benefit_per_hour}</td></tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

export default SimulationPage
