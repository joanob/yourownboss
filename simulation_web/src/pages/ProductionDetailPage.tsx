import React, {useEffect, useState} from 'react'
import { useRoute, Link, useLocation } from 'wouter'
import api from '../lib/api'

type ProcessResource = { id?: number | null; resource_id: number; is_output: boolean }
type Process = { id?: number | null; name: string; start_hour?: number | null; end_hour?: number | null; resources: ProcessResource[] }
type Building = { id: number; name: string; processes: Process[] }

const emptyBuilding = (): Building => ({ id: 0, name: '', processes: [] })

const ProductionDetailPage: React.FC = () => {
  const [match, params] = useRoute('/production/:id')
  const [loc, setLoc] = useLocation()
  const idParam = params?.id || 'new'
  const [building, setBuilding] = useState<Building>(emptyBuilding())
  const [loading, setLoading] = useState(false)
  const [resources, setResources] = useState<Array<{id:number; name:string}>>([])

  // load resources for select dropdown
  useEffect(() => {
    let mounted = true
    api.get('/resources')
      .then((res) => { if (!mounted) return; setResources(res.data || []) })
      .catch(() => {})
    return () => { mounted = false }
  }, [])

  useEffect(() => {
    if (!match) return
    if (idParam === 'new') {
      setBuilding(emptyBuilding())
      return
    }
    const id = parseInt(idParam, 10)
    if (!id) return
    setLoading(true)
    let mounted = true
    api.get(`/buildings/${id}`).then((res) => {
      if (!mounted) return
      setBuilding(res.data)
    }).catch(() => {}).finally(() => { if (mounted) setLoading(false) })
    return () => { mounted = false }
  }, [match, idParam])

  const setField = (field: keyof Building, value: any) => setBuilding(b => ({...b, [field]: value}))

  const addProcess = () => setBuilding(b => ({...b, processes: [...b.processes, {id: undefined, name: '', start_hour: null, end_hour: null, resources: []}]}))
  const removeProcess = (index: number) => setBuilding(b => ({...b, processes: b.processes.filter((_,i)=>i!==index)}))

  const setProcessField = (index: number, field: keyof Process, value: any) => {
    setBuilding(b => {
      const copy = {...b, processes: b.processes.map((p,i)=> i===index ? {...p, [field]: value} : p)}
      return copy
    })
  }

  const addProcessResource = (pIndex: number) => {
    setBuilding(b => {
      const copy = {...b}
      copy.processes = copy.processes.map((p,i) => i===pIndex ? {...p, resources: [...p.resources, {resource_id: 0, is_output: false}]} : p)
      return copy
    })
  }

  const removeProcessResource = (pIndex:number, rIndex:number) => {
    setBuilding(b => {
      const copy = {...b}
      copy.processes = copy.processes.map((p,i) => i===pIndex ? {...p, resources: p.resources.filter((_,ri)=>ri!==rIndex)} : p)
      return copy
    })
  }

  const setProcessResourceField = (pIndex:number, rIndex:number, field: keyof ProcessResource, value:any) => {
    setBuilding(b => {
      const copy = {...b}
      copy.processes = copy.processes.map((p,i) => {
        if (i!==pIndex) return p
        const resCopy = p.resources.map((r,ri) => ri===rIndex ? {...r, [field]: value} : r)
        return {...p, resources: resCopy}
      })
      return copy
    })
  }

  const save = async () => {
    if (building.id === 0) {
      alert('El id del edificio debe ser un número distinto de 0')
      return
    }
    try {
      await api.post(`/production/${building.id}`, building)
      alert('Guardado OK')
      setLoc('/production')
    } catch (err) {
      console.error(err)
      alert('Error guardando. Revisa la consola.')
    }
  }

  return (
    <div>
      <h2>Building detail</h2>
      {loading ? <p>Cargando...</p> : (
        <div>
          <div style={{display: 'flex', gap: 8, marginBottom: 12}}>
            <label>ID: <input type="number" value={building.id} onChange={(e)=>setField('id', parseInt(e.target.value||'0',10))} /></label>
            <label style={{flex:1}}>Name: <input style={{width:'100%'}} value={building.name} onChange={(e)=>setField('name', e.target.value)} /></label>
          </div>

          <h3>Processes</h3>
          <div>
            {building.processes.map((p, i) => (
              <div key={i} style={{border: '1px solid #eee', padding: 8, marginBottom: 8}}>
                <div style={{display:'flex', gap:8}}>
                  <label style={{width:120}}>Proc ID: <input type="number" value={p.id ?? 0} onChange={(e)=>setProcessField(i,'id', parseInt(e.target.value||'0',10))} /></label>
                  <label style={{flex:1}}>Name: <input value={p.name} onChange={(e)=>setProcessField(i,'name', e.target.value)} /></label>
                  <button onClick={()=>removeProcess(i)}>Eliminar proceso</button>
                </div>
                <div style={{marginTop:8}}>
                  <strong>Resources</strong>
                  <table style={{width:'100%', marginTop:6}}>
                    <thead>
                      <tr><th>Resource ID</th><th>Is Output</th><th></th></tr>
                    </thead>
                    <tbody>
                      {p.resources.map((r, ri) => (
                        <tr key={ri}>
                          <td>
                            <select value={r.resource_id} onChange={(e)=>setProcessResourceField(i,ri,'resource_id', parseInt(e.target.value||'0',10))}>
                              <option value={0}>-- seleccionar recurso --</option>
                              {resources.map((resrc) => (
                                <option key={resrc.id} value={resrc.id}>{resrc.name} ({resrc.id})</option>
                              ))}
                            </select>
                          </td>
                          <td><input type="checkbox" checked={!!r.is_output} onChange={(e)=>setProcessResourceField(i,ri,'is_output', e.target.checked)} /></td>
                          <td><button onClick={()=>removeProcessResource(i,ri)}>Eliminar</button></td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  <div style={{marginTop:6}}><button onClick={()=>addProcessResource(i)}>Añadir recurso</button></div>
                </div>
              </div>
            ))}
          </div>
          <div style={{marginTop:8}}>
            <button onClick={addProcess}>Añadir proceso</button>
          </div>

          <div style={{marginTop:12, display:'flex', gap:8}}>
            <button onClick={save}>Guardar edificio</button>
            <Link href="/production"><button>Volver</button></Link>
          </div>
        </div>
      )}
    </div>
  )
}

export default ProductionDetailPage
