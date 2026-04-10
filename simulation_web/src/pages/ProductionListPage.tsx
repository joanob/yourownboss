import React, {useEffect, useState, useRef} from 'react'
import { Link } from 'wouter'
import api from '../lib/api'

type Building = { id: number; name: string }

const ProductionListPage: React.FC = () => {
  const [buildings, setBuildings] = useState<Building[]>([])
  const fetchRef = useRef(false)

  useEffect(() => {
    if (fetchRef.current) return
    fetchRef.current = true
    let mounted = true
    api.get('/buildings')
      .then((res) => { if (!mounted) return; setBuildings(res.data || []) })
      .catch(() => {})
    return () => { mounted = false }
  }, [])

  return (
    <div>
      <h2>Production buildings</h2>
      <p>Listado de edificios. Haz click en uno para ver/editar sus procesos.</p>
      <div style={{marginBottom: 12}}>
        <Link href="/production/new"><button>Crear edificio</button></Link>
      </div>
      <table style={{width: '100%', borderCollapse: 'collapse'}}>
        <thead>
          <tr>
            <th style={{textAlign: 'left', padding: 6}}>ID</th>
            <th style={{textAlign: 'left', padding: 6}}>Name</th>
          </tr>
        </thead>
        <tbody>
          {buildings.map((b) => (
            <tr key={b.id} style={{borderTop: '1px solid #eee'}}>
              <td style={{padding: 6}}>
                <Link href={`/production/${b.id}`}>{b.id}</Link>
              </td>
              <td style={{padding: 6}}>
                <Link href={`/production/${b.id}`}>{b.name}</Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default ProductionListPage
