(function() {
  'use strict';

  const STORAGE_KEY = 'beanAndBrewCart';

  function getCart() {
    try {
      return JSON.parse(localStorage.getItem(STORAGE_KEY)) || [];
    } catch { return []; }
  }

  function saveCart(cart) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(cart));
    updateBadge();
  }

  function updateBadge() {
    const badge = document.getElementById('cart-badge');
    if (!badge) return;
    const cart = getCart();
    const count = cart.reduce((s, i) => s + i.qty, 0);
    badge.textContent = count > 0 ? count : '';
  }

  window.addToCart = function(id, name, price, emoji) {
    const cart = getCart();
    const existing = cart.find(i => i.id === id);
    if (existing) {
      existing.qty++;
    } else {
      cart.push({ id, name, price, emoji, qty: 1 });
    }
    saveCart(cart);
    showToast(name + ' добавлен в корзину');
  };

  window.removeFromCart = function(id) {
    const cart = getCart().filter(i => i.id !== id);
    saveCart(cart);
    renderCartPage();
  };

  window.changeQty = function(id, delta) {
    const cart = getCart();
    const item = cart.find(i => i.id === id);
    if (!item) return;
    item.qty += delta;
    if (item.qty <= 0) {
      saveCart(cart.filter(i => i.id !== id));
    } else {
      saveCart(cart);
    }
    renderCartPage();
  };

  window.showToast = function(msg) {
    const el = document.getElementById('toast');
    if (!el) return;
    el.textContent = msg;
    el.classList.add('show');
    setTimeout(() => el.classList.remove('show'), 2000);
  };

  window.submitCheckout = function(e) {
    e.preventDefault();
    const cart = getCart();
    if (cart.length === 0) {
      showToast('Корзина пуста!');
      return false;
    }
    const name = document.getElementById('name').value || 'guest';
    const items = cart.map(i => ({ product_id: i.id, quantity: i.qty }));

    fetch('/api/v1/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ customer_id: name, items: items })
    })
    .then(r => r.json())
    .then(data => {
      localStorage.removeItem(STORAGE_KEY);
      updateBadge();
      window.location.href = '/order/' + data.id;
    })
    .catch(() => showToast('Ошибка при оформлении заказа'));
    return false;
  };

  function renderCartPage() {
    const tbody = document.getElementById('cart-tbody');
    if (!tbody) return;

    const cart = getCart();
    const emptyEl = document.getElementById('cart-empty');
    const itemsEl = document.getElementById('cart-items');

    if (cart.length === 0) {
      emptyEl.style.display = '';
      itemsEl.style.display = 'none';
      return;
    }
    emptyEl.style.display = 'none';
    itemsEl.style.display = '';

    let total = 0;
    tbody.innerHTML = cart.map(i => {
      const subtotal = i.price * i.qty;
      total += subtotal;
      return `<tr>
        <td style="font-size:2rem;">${i.emoji || '☕'}</td>
        <td class="cart-item-name">${i.name}</td>
        <td>${i.price} ₽</td>
        <td>
          <div class="cart-qty">
            <button onclick="changeQty(${i.id}, -1)">−</button>
            <span>${i.qty}</span>
            <button onclick="changeQty(${i.id}, 1)">+</button>
          </div>
        </td>
        <td><strong>${subtotal} ₽</strong></td>
        <td><button class="btn btn-danger btn-sm" onclick="removeFromCart(${i.id})">✕</button></td>
      </tr>`;
    }).join('');
    document.getElementById('cart-total').textContent = total + ' ₽';
  }

  function renderCheckoutTotal() {
    const el = document.getElementById('checkout-total');
    if (!el) return;
    const cart = getCart();
    const total = cart.reduce((s, i) => s + i.price * i.qty, 0);
    el.textContent = total + ' ₽';
  }

  document.addEventListener('DOMContentLoaded', function() {
    updateBadge();
    renderCartPage();
    renderCheckoutTotal();
  });
})();
