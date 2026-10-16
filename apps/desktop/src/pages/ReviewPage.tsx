import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { formatDate } from '../lib/formatDate';
import { formatSignedMoney } from '../lib/formatMoney';

export default function ReviewPage() {
  const [items, setItems] = useState<any[]>(() => {
    try { return JSON.parse(localStorage.getItem('reviewQueue') || '[]'); } catch { return [] }
  });
  const [categories, setCategories] = useState<any[]>([]);
  const [movements, setMovements] = useState<any[]>([]);
  const [selectedCat, setSelectedCat] = useState<Record<string,string>>({});

  useEffect(() => {
    fetch('http://localhost:8081/categories')
      .then(r=>r.json())
      .then(d=> setCategories((d||[]).map((c:any)=>({ id: c.id ?? c.ID, name: c.name ?? c.Name, type: (c.category_type ?? c.CategoryType ?? '').toUpperCase() }))));
    fetch('http://localhost:8081/movements')
      .then(r=>r.json())
      .then(d=> setMovements(d||[]));
  }, []);

  useEffect(() => {
    const handler = () => setItems(JSON.parse(localStorage.getItem('reviewQueue') || '[]'));
    window.addEventListener('storage', handler);
    return () => window.removeEventListener('storage', handler);
  }, []);

  const removeItem = (idx:number) => {
    const q = [...items];
    q.splice(idx,1);
    localStorage.setItem('reviewQueue', JSON.stringify(q));
    setItems(q);
    if (q.length === 0) {
      window.location.href = '/';
    }
  };

  const handleConfirm = async (item:any, idx:number) => {
    const body = {
      amountMinor: Math.round(Number(item.amount) * 100) * (item.type === 'Egreso' ? -1 : 1),
      type: item.type === 'Ingreso' ? 'INCOME' : item.type === 'Egreso' ? 'EXPENSE' : 'TRANSFER',
      accountId: item.account,
      description: item.desc,
      date: item.date,
      categoryId: ''
    };
    const res = await fetch('http://localhost:8081/movements?force=true', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body)});
    if (res.ok) {
      removeItem(idx);
    } else {
      alert('Error al confirmar');
    }
  };

  if (items.length === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Revisión</h1>
        <div className="rounded-[12px] border border-dashed border-line bg-surface/50 p-12 text-center">
          <p className="text-body text-ink-2">No hay ítems pendientes de revisión.</p>
          <p className="mt-1 text-meta text-ink-3">Cuando un documento requiera atención aparecerá aquí.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">Revisión</h1>
      {items.map((it, idx) => {
        const isCategory = !!it.categoryId;
        const mv = movements.find(m=> m.ID === it.movementId);
        return (
        <div key={idx} className="rounded-[12px] border border-line bg-surface p-6">
          {isCategory ? (
            <>
              <p className="text-body font-medium">Categoría eliminada: movimiento requiere revisión</p>
              <ul className="mt-3 space-y-1 text-body">
                <li><b>Fecha:</b> {mv?.MovementDate ? formatDate(mv.MovementDate) : it.date}</li>
                <li><b>Descripción:</b> {mv?.Description || it.description || it.desc || '—'}</li>
                <li><b>Categoría anterior:</b> {it.categoryName || '—'}</li>
                <li><b>Monto:</b> {mv ? formatSignedMoney(mv.AmountMinor) : '—'}</li>
              </ul>
              <div className="mt-4 space-y-3">
                {(() => {
                  const cats = categories.filter(c=> c.type === it.categoryType);
                  return (
                    <>
                      <select
                        value={selectedCat[it.movementId] || ''}
                        onChange={e=> setSelectedCat(s=>({...s, [it.movementId]: e.target.value}))}
                        className="w-full border rounded px-2 py-1 bg-black text-white"
                        disabled={cats.length===0}
                      >
                        <option value="">Seleccionar categoría…</option>
                        {cats.map(c=> (
                          <option key={c.id} value={c.id}>{c.name}</option>
                        ))}
                      </select>
                      <div className="flex gap-2">
                        {cats.length > 0 ? (
                          <button onClick={async ()=>{
                            const catId = selectedCat[it.movementId] || '';
                            await fetch(`http://localhost:8081/movements/${it.movementId}/category`, { method:'PUT', headers:{'Content-Type':'application/json'}, body: JSON.stringify({categoryId: catId}) });
                            removeItem(idx);
                          }} className="px-3 py-1 rounded bg-accent-strong text-white">Reemplazar</button>
                        ) : (
                          <button onClick={()=> window.location.href='/settings'} className="px-3 py-1 rounded border">Agregar categoría</button>
                        )}
                        <button onClick={async ()=>{
                          await fetch(`http://localhost:8081/movements/${it.movementId}/category`, { method:'PUT', headers:{'Content-Type':'application/json'}, body: JSON.stringify({categoryId: null}) });
                          removeItem(idx);
                        }} className="px-3 py-1 rounded border">Guardar sin categoría</button>
                        <button onClick={async ()=>{
                          await fetch(`http://localhost:8081/movements/${it.movementId}`, { method:'DELETE' });
                          removeItem(idx);
                        }} className="px-3 py-1 rounded border text-neg">Eliminar movimiento</button>
                      </div>
                    </>
                  );
                })()}
              </div>
            </>
          ) : (
            <>
              <p className="text-body">Movimiento duplicado detectado:</p>
              <ul className="mt-3 space-y-1 text-body">
                <li><b>Cuenta:</b> {it.account}</li>
                <li><b>Fecha:</b> {it.date}</li>
                <li><b>Tipo:</b> {it.type}</li>
                <li><b>Cantidad:</b> {it.amount}</li>
                <li><b>Descripción:</b> {it.desc}</li>
              </ul>
              <div className="mt-4 flex gap-2">
                <button onClick={()=>removeItem(idx)} className="px-3 py-1 rounded border">Cancelar</button>
                <button onClick={()=>handleConfirm(it, idx)} className="px-3 py-1 rounded bg-accent-strong text-white">Confirmar y agregar</button>
              </div>
            </>
          )}
        </div>
      );
      })}
      <Link to="/" className="px-3 py-1 rounded border inline-block">Volver</Link>
    </div>
  );
}
