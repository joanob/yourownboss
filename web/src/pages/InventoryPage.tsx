import { useEffect, useState } from 'react';
import { GameLayout } from '@/components/GameLayout';
import { ResourceCard } from '@/components/ResourceCard';
import { useCompany } from '@/contexts/CompanyContext';
import { inventoryAPI, marketAPI } from '@/lib/api';
import type { InventoryItem, Resource } from '@/lib/api';

interface EditingResource {
  targetQuantity: number;
  originalQuantity: number;
}

export function InventoryPage() {
  const { company, refreshCompany } = useCompany();
  const [inventory, setInventory] = useState<InventoryItem[]>([]);
  const [resources, setResources] = useState<Resource[]>([]);
  const [loading, setLoading] = useState(false);
  const [editingMap, setEditingMap] = useState<Record<number, EditingResource>>({});
  const [tradeStatus, setTradeStatus] = useState<
    Record<number, 'idle' | 'loading' | 'success'>
  >({});

  useEffect(() => {
    loadData();
  }, []);

  useEffect(() => {
    if (resources.length === 0) return;

    setEditingMap((prev) => {
      const next = { ...prev };
      let changed = false;

      resources.forEach((resource) => {
        const ownedQty =
          inventory.find((i) => i.resource_id === resource.id)?.quantity || 0;
        const existing = prev[resource.id];
        if (!existing || existing.originalQuantity !== ownedQty) {
          next[resource.id] = {
            originalQuantity: ownedQty,
            targetQuantity: ownedQty,
          };
          changed = true;
        }
      });

      return changed ? next : prev;
    });
  }, [inventory, resources]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [resRes, invRes] = await Promise.all([
        inventoryAPI.getResources(),
        inventoryAPI.getInventory(),
      ]);
      setResources(resRes.data);
      setInventory(invRes.data);
    } catch (err) {
      console.error('Error loading data:', err);
    } finally {
      setLoading(false);
    }
  };

  const getOwnedQuantity = (resourceId: number) =>
    inventory.find((i) => i.resource_id === resourceId)?.quantity || 0;

  const updateQuantity = (resourceId: number, newQuantity: number) => {
    setEditingMap((prev) => {
      const resource = resources.find((r) => r.id === resourceId);
      if (!resource) return prev;

      const existing = prev[resourceId];
      const originalQuantity =
        existing?.originalQuantity ?? getOwnedQuantity(resourceId);

      let nextQuantity = Math.max(0, newQuantity);
      if (company) {
        const maxPacks = Math.floor(company.money / resource.price);
        const maxBuyQuantity =
          originalQuantity + maxPacks * resource.pack_size;
        nextQuantity = Math.min(nextQuantity, maxBuyQuantity);
      }

      return {
        ...prev,
        [resourceId]: {
          originalQuantity,
          targetQuantity: nextQuantity,
        },
      };
    });
  };

  const setSellAll = (resourceId: number) => {
    setEditingMap((prev) => {
      const existing = prev[resourceId];
      const originalQuantity =
        existing?.originalQuantity ?? getOwnedQuantity(resourceId);
      return {
        ...prev,
        [resourceId]: {
          originalQuantity,
          targetQuantity: 0,
        },
      };
    });
  };

  const setMaxBuy = (resourceId: number) => {
    if (!company) return;
    const resource = resources.find((r) => r.id === resourceId);
    if (!resource) return;

    setEditingMap((prev) => {
      const existing = prev[resourceId];
      const originalQuantity =
        existing?.originalQuantity ?? getOwnedQuantity(resourceId);
      const maxPacks = Math.floor(company.money / resource.price);
      const maxUnits = originalQuantity + maxPacks * resource.pack_size;
      return {
        ...prev,
        [resourceId]: {
          originalQuantity,
          targetQuantity: maxUnits,
        },
      };
    });
  };

  const resetQuantity = (resourceId: number) => {
    setEditingMap((prev) => {
      const existing = prev[resourceId];
      const originalQuantity =
        existing?.originalQuantity ?? getOwnedQuantity(resourceId);
      return {
        ...prev,
        [resourceId]: {
          originalQuantity,
          targetQuantity: originalQuantity,
        },
      };
    });
  };

  const handleTrade = async (resourceId: number) => {
    const editingState = editingMap[resourceId];
    if (!editingState) return;

    const resource = resources.find((r) => r.id === resourceId);
    if (!resource) return;

    const diff = editingState.targetQuantity - editingState.originalQuantity;
    if (diff === 0) {
      return;
    }

    const isBuying = diff > 0;
    const unitsToTrade = Math.abs(diff);
    const packCount = Math.ceil(unitsToTrade / resource.pack_size);

    try {
      setTradeStatus((prev) => ({ ...prev, [resourceId]: 'loading' }));
      if (isBuying) {
        await marketAPI.buy(resource.id, packCount);
      } else {
        await marketAPI.sell(resource.id, packCount);
      }
      await loadData();
      await refreshCompany();
      setTradeStatus((prev) => ({ ...prev, [resourceId]: 'success' }));
      window.setTimeout(() => {
        setTradeStatus((prev) => ({ ...prev, [resourceId]: 'idle' }));
      }, 500);
    } catch (err: any) {
      console.error('Error en la operación:', err);
      setTradeStatus((prev) => ({ ...prev, [resourceId]: 'idle' }));
    }
  };

  if (loading) {
    return (
      <GameLayout>
        <div style={{ padding: '16px', textAlign: 'center' }}>
          <p>Cargando...</p>
        </div>
      </GameLayout>
    );
  }

  return (
    <GameLayout>
      <div style={{ padding: '16px 16px 96px' }}>
        <h1 style={{ marginBottom: '16px', fontSize: '24px', fontWeight: 600 }}>
          📦 Inventario
        </h1>

        {/* Resources Grid */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: '16px',
          }}
        >
          {resources.map((resource) => {
            const ownedItem = inventory.find(
              (i) => i.resource_id === resource.id
            );
            const ownedQty = ownedItem?.quantity || 0;
            const editingState =
              editingMap[resource.id] || {
                originalQuantity: ownedQty,
                targetQuantity: ownedQty,
              };

            return (
              <ResourceCard
                key={resource.id}
                resource={resource}
                editingState={editingState}
                company={company}
                tradeStatus={tradeStatus[resource.id] || 'idle'}
                onChangeQuantity={(newQuantity) =>
                  updateQuantity(resource.id, newQuantity)
                }
                onSellAll={() => setSellAll(resource.id)}
                onSetMaxBuy={() => setMaxBuy(resource.id)}
                onTrade={() => handleTrade(resource.id)}
                onCancel={() => resetQuantity(resource.id)}
              />
            );
          })}
        </div>
      </div>
    </GameLayout>
  );
}
