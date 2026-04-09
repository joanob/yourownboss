import { GameLayout } from '@/components/GameLayout';
import { useCallback, useEffect, useState } from 'react';
import { productionAPI } from '@/lib/api';
import type { ProductionBuilding, ProductionProcess, ProductionProcessResource } from '@/lib/api';
import { formatMoney } from '@/lib/money';

export function EncyclopediaPage() {
  const [buildings, setBuildings] = useState<ProductionBuilding[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [selectedBuilding, setSelectedBuilding] = useState<ProductionBuilding | null>(null);
  const [selectedProcess, setSelectedProcess] = useState<ProductionProcess | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await productionAPI.getProductionBuildings();
      setBuildings(res.data || []);
    } catch (e: any) {
      setError(e?.message || 'Failed to load');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <GameLayout>
      <div style={{ padding: '16px' }}>
        <div
          style={{
            backgroundColor: '#ffffff',
            borderRadius: '8px',
            padding: '24px',
            boxShadow: '0 2px 4px rgba(0,0,0,0.08)',
          }}
        >
          <h2 style={{ marginTop: 0 }}>Enciclopedia</h2>

          {loading && <div>Cargando...</div>}
          {error && (
            <div style={{ color: 'red' }}>
              Error cargando datos. <button onClick={load}>Reintentar</button>
            </div>
          )}

          {!loading && !selectedBuilding && (
            <div>
              <h3>Edificios de producción</h3>
              <div style={{ display: 'grid', gap: 12 }}>
                {buildings.map((b) => (
                  <div
                    key={b.id}
                    onClick={() => setSelectedBuilding(b)}
                    style={{
                      padding: 12,
                      borderRadius: 8,
                      background: '#f8f9fa',
                      cursor: 'pointer',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: 600 }}>{b.name}</div>
                      <div style={{ fontSize: 12, color: '#666' }}>Costo: ${formatMoney(b.cost, false)}</div>
                    </div>
                    <div style={{ fontSize: 12, color: '#007bff' }}>Ver procesos →</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {!loading && selectedBuilding && !selectedProcess && (
            <div>
              <button onClick={() => setSelectedBuilding(null)}>← Volver a edificios</button>
              <h3 style={{ marginTop: 12 }}>Procesos: {selectedBuilding.name}</h3>
              <div style={{ display: 'grid', gap: 12 }}>
                {selectedBuilding.processes.map((p) => (
                  <div
                    key={p.id}
                    onClick={() => setSelectedProcess(p)}
                    style={{
                      padding: 12,
                      borderRadius: 8,
                      background: '#f1f3f5',
                      cursor: 'pointer',
                    }}
                  >
                    <div style={{ fontWeight: 700 }}>{p.name}</div>
                    <div style={{ fontSize: 12, color: '#666' }}>
                      Tiempo: {p.processing_time_ms} ms
                      {p.window_start_hour != null && (
                        <span> • Ventana: {p.window_start_hour}–{p.window_end_hour}h</span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {!loading && selectedProcess && (
            <div>
              <button onClick={() => setSelectedProcess(null)}>← Volver a procesos</button>
              <h3 style={{ marginTop: 12 }}>{selectedProcess.name}</h3>

              <div style={{ display: 'flex', gap: 12, marginTop: 8 }}>
                <div style={{ flex: 1 }}>
                  <h4>Recursos entrada</h4>
                  {selectedProcess.input_resources.length === 0 && <div style={{ color: '#666' }}>—</div>}
                  {selectedProcess.input_resources.map((r: ProductionProcessResource) => (
                    <div key={r.resource_id} style={{ padding: 8, borderRadius: 6, background: '#fff', marginBottom: 8 }}>
                      <div style={{ fontWeight: 600 }}>{r.resource_name}</div>
                      <div style={{ fontSize: 12, color: '#666' }}>Cantidad: {r.quantity}</div>
                    </div>
                  ))}
                </div>

                <div style={{ flex: 1 }}>
                  <h4>Recursos salida</h4>
                  {selectedProcess.output_resources.length === 0 && <div style={{ color: '#666' }}>—</div>}
                  {selectedProcess.output_resources.map((r: ProductionProcessResource) => (
                    <div key={r.resource_id} style={{ padding: 8, borderRadius: 6, background: '#fff', marginBottom: 8 }}>
                      <div style={{ fontWeight: 600 }}>{r.resource_name}</div>
                      <div style={{ fontSize: 12, color: '#666' }}>Cantidad: {r.quantity}</div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </GameLayout>
  );
}
