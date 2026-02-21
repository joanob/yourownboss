import type { ChangeEvent } from 'react';
import type { Resource } from '@/lib/api';
import { formatMoney } from '@/lib/money';

interface EditingState {
  targetQuantity: number;
  originalQuantity: number;
}

interface Company {
  money: number;
}

interface ResourceCardProps {
  resource: Resource;
  editingState: EditingState | null;
  company: Company | null;
  tradeStatus: 'idle' | 'loading' | 'success';
  onChangeQuantity: (newQuantity: number) => void;
  onSellAll: () => void;
  onSetMaxBuy: () => void;
  onTrade: () => void;
  onCancel: () => void;
}

export function ResourceCard({
  resource,
  editingState,
  company,
  tradeStatus,
  onChangeQuantity,
  onSellAll,
  onSetMaxBuy,
  onTrade,
  onCancel,
}: ResourceCardProps) {
  if (!editingState) return null;

  const diff = editingState.targetQuantity - editingState.originalQuantity;
  const isBuying = diff > 0;
  const unitsToTrade = Math.abs(diff);
  const packCount = Math.ceil(unitsToTrade / resource.pack_size);
  const totalCost = packCount * resource.price;

  const canDecrease = editingState.targetQuantity > 0;

  const nextTarget = editingState.targetQuantity + resource.pack_size;
  const nextDiff = nextTarget - editingState.originalQuantity;
  let canIncrease = true;
  if (company && nextDiff > 0) {
    const nextPackCount = Math.ceil(nextDiff / resource.pack_size);
    const nextTotalCost = nextPackCount * resource.price;
    canIncrease = nextTotalCost <= company.money;
  }

  const maxPacks = company ? Math.floor(company.money / resource.price) : null;
  const maxBuyQuantity =
    company && maxPacks !== null
      ? editingState.originalQuantity + maxPacks * resource.pack_size
      : null;
  const canBuyMax =
    !!company &&
    maxBuyQuantity !== null &&
    editingState.targetQuantity < maxBuyQuantity;

  const handleDecreaseClick = () => {
    const newQty = Math.max(0, editingState.targetQuantity - resource.pack_size);
    onChangeQuantity(newQty);
  };

  const handleIncreaseClick = () => {
    const newQty = editingState.targetQuantity + resource.pack_size;
    onChangeQuantity(newQty);
  };

  const handleDirectInput = (e: ChangeEvent<HTMLInputElement>) => {
    const newValue = parseInt(e.target.value) || 0;
    onChangeQuantity(Math.max(0, newValue));
  };

  const showSuccess = tradeStatus === 'success';
  const particles = [
    { left: '10%', top: '18%', delay: '0ms', size: 18 },
    { left: '22%', top: '42%', delay: '40ms', size: 20 },
    { left: '36%', top: '48%', delay: '80ms', size: 22 },
    { left: '50%', top: '44%', delay: '20ms', size: 24 },
    { left: '64%', top: '48%', delay: '100ms', size: 22 },
    { left: '78%', top: '42%', delay: '60ms', size: 20 },
    { left: '88%', top: '20%', delay: '120ms', size: 18 },
    { left: '14%', top: '70%', delay: '90ms', size: 18 },
    { left: '84%', top: '70%', delay: '30ms', size: 18 },
  ];

  return (
    <div
      style={{
        padding: '20px',
        backgroundColor: '#ffffff',
        borderRadius: '12px',
        boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
        border: '2px solid #007bff',
        transition: 'all 0.2s',
        position: 'relative',
      }}
    >
      {showSuccess && (
        <div
          style={{
            position: 'absolute',
            inset: 0,
            pointerEvents: 'none',
            overflow: 'hidden',
          }}
        >
          <style>{`
            @keyframes floatDollar {
              0% { transform: translateY(0); opacity: 0; }
              20% { opacity: 1; }
              100% { transform: translateY(-32px); opacity: 0; }
            }
          `}</style>
          {particles.map((particle, index) => (
            <span
              key={index}
              style={{
                position: 'absolute',
                left: particle.left,
                top: particle.top,
                fontSize: `${particle.size}px`,
                animation: `floatDollar 0.5s ease-out ${particle.delay} forwards`,
                opacity: 0,
                color: '#28a745',
              }}
            >
              $
            </span>
          ))}
        </div>
      )}
      {/* Row 1: Name & Price */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '16px',
        }}
      >
        <h3
          style={{
            margin: 0,
            fontSize: '16px',
            fontWeight: 600,
          }}
        >
          {resource.name}
        </h3>
        <div style={{ fontSize: '13px', color: '#666' }}>
          ${formatMoney(resource.price)}
          {resource.pack_size > 1 && ` / ${resource.pack_size}u`}
        </div>
      </div>

      {/* Row 2: Quantity Controls */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          gap: '12px',
          marginBottom: '12px',
        }}
      >
        <button
          onClick={handleDecreaseClick}
          disabled={!canDecrease}
          style={{
            width: '32px',
            height: '32px',
            fontSize: '18px',
            fontWeight: 600,
            backgroundColor: !canDecrease ? '#ccc' : '#dc3545',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: !canDecrease ? 'not-allowed' : 'pointer',
          }}
        >
          −
        </button>
        <input
          type="number"
          min="0"
          value={editingState.targetQuantity}
          onChange={handleDirectInput}
          style={{
            minWidth: '80px',
            padding: '8px',
            fontSize: '18px',
            fontWeight: 600,
            textAlign: 'center',
            border: '2px solid #ddd',
            borderRadius: '6px',
            boxSizing: 'border-box',
          }}
        />
        <button
          onClick={handleIncreaseClick}
          disabled={!canIncrease}
          style={{
            width: '32px',
            height: '32px',
            fontSize: '18px',
            fontWeight: 600,
            backgroundColor: !canIncrease ? '#ccc' : '#28a745',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: !canIncrease ? 'not-allowed' : 'pointer',
          }}
        >
          +
        </button>
      </div>

      {/* Row 3: Sell All / Price Display / Buy Max */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          gap: '8px',
          marginBottom: '12px',
        }}
      >
        <button
          onClick={onSellAll}
          disabled={editingState.originalQuantity === 0}
          style={{
            flex: 1,
            padding: '6px 8px',
            fontSize: '12px',
            fontWeight: 600,
            backgroundColor:
              editingState.originalQuantity === 0 ? '#ddd' : '#dc3545',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor:
              editingState.originalQuantity === 0
                ? 'not-allowed'
                : 'pointer',
          }}
        >
          Vender todo
        </button>
        <div
          style={{
            flex: 1,
            fontSize: '12px',
            fontWeight: 600,
            textAlign: 'center',
          }}
        >
          {diff !== 0 ? (
            <>
              {isBuying ? '-' : '+'}${formatMoney(totalCost)}
            </>
          ) : (
            <span style={{ color: '#999' }}>—</span>
          )}
        </div>
        <button
          onClick={onSetMaxBuy}
          disabled={!canBuyMax}
          style={{
            flex: 1,
            padding: '6px 8px',
            fontSize: '12px',
            fontWeight: 600,
            backgroundColor: !canBuyMax ? '#ccc' : '#007bff',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: !canBuyMax ? 'not-allowed' : 'pointer',
          }}
        >
          Comprar máx
        </button>
      </div>

      {/* Row 4: Confirm / Cancel Buttons */}
      <div
        style={{
          display: 'flex',
          gap: '8px',
        }}
      >
        <button
          onClick={onCancel}
          style={{
            flex: 1,
            padding: '10px',
            fontSize: '14px',
            fontWeight: 600,
            backgroundColor: '#6c757d',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
          }}
        >
          Cancelar
        </button>
        <button
          onClick={onTrade}
          disabled={diff === 0 || tradeStatus === 'loading'}
          style={{
            flex: 1,
            padding: '10px',
            fontSize: '14px',
            fontWeight: 600,
            backgroundColor:
              tradeStatus === 'success'
                ? '#20c997'
                : tradeStatus === 'loading'
                ? '#ffc107'
                : diff === 0
                ? '#ccc'
                : isBuying
                ? '#28a745'
                : '#007bff',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor:
              diff === 0 || tradeStatus === 'loading'
                ? 'not-allowed'
                : 'pointer',
            transition: 'background-color 0.25s ease, transform 0.2s ease',
            transform: tradeStatus === 'success' ? 'scale(1.02)' : 'scale(1)',
          }}
        >
          {tradeStatus === 'loading'
            ? 'Procesando...'
            : tradeStatus === 'success'
            ? 'Listo!'
            : 'Confirmar'}
        </button>
      </div>
    </div>
  );
}
