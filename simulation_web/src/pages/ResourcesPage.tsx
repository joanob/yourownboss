import React, {useEffect, useState, useRef} from 'react'
import api from '../lib/api'
import { getResources, invalidateResources } from '../lib/resourceCache'

type Resource = { id: number; name: string }

const ResourcesPage: React.FC = () => {
  const [resources, setResources] = useState<Resource[]>([])
  const fetchRef = useRef(false)
  const [saving, setSaving] = useState(false)

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

  const addRow = () => setResources((p) => [...p, {id: 0, name: ''}])

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

  console.log(resources)

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
            <tr key={r.id ?? i} style={{borderTop: '1px solid #eee'}}>
              <td style={{padding: 6}}>
                <input type="number" value={r.id} onChange={(e) => setId(i, parseInt(e.target.value || '0', 10))} />
              </td>
              <td style={{padding: 6}}>
                <input value={r.name} onChange={(e) => setName(i, e.target.value)} />
              </td>
              <td style={{padding: 6}}>
                <button onClick={() => removeRow(i)}>Eliminar</button>
              </td>
            </tr>
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
