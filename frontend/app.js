const API = 'http://localhost:8080';
let token = localStorage.getItem('token');
let userId = localStorage.getItem('user_id');

// === Router ===
function showPage(page) {
    if (!token && page !== 'login' && page !== 'register') {
        showPage('login');
        return;
    }
    const pages = { login, register, catalog, orders, profile };
    document.getElementById('app').innerHTML = '';
    pages[page]();
    updateNav();
}

function updateNav() {
    document.getElementById('logoutBtn').classList.toggle('hidden', !token);
}

function logout() {
    localStorage.clear();
    token = null;
    userId = null;
    showPage('login');
}

// === API Helper ===
async function api(method, path, body = null) {
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    const opts = { method, headers };
    if (body) opts.body = JSON.stringify(body);
    const res = await fetch(API + path, opts);
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Request failed');
    return data;
}

// === Login ===
function login() {
    document.getElementById('app').innerHTML = `
        <div class="card"><h2>Login</h2>
            <div id="error" class="error"></div>
            <input id="email" placeholder="Email" type="email">
            <input id="password" placeholder="Password" type="password">
            <button class="btn btn-primary" onclick="doLogin()">Login</button>
            <p style="margin-top:10px">No account? <a href="#" onclick="showPage('register')">Register</a></p>
        </div>`;
}

async function doLogin() {
    try {
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const data = await api('POST', '/auth/login', { email, password });
        token = data.token;
        userId = data.user_id;
        localStorage.setItem('token', token);
        localStorage.setItem('user_id', userId);
        showPage('catalog');
    } catch (e) {
        document.getElementById('error').textContent = e.message;
    }
}

// === Register ===
function register() {
    document.getElementById('app').innerHTML = `
        <div class="card"><h2>Register</h2>
            <div id="error" class="error"></div>
            <input id="email" placeholder="Email" type="email">
            <input id="password" placeholder="Password" type="password">
            <input id="first_name" placeholder="First Name">
            <input id="last_name" placeholder="Last Name">
            <button class="btn btn-primary" onclick="doRegister()">Register</button>
            <p style="margin-top:10px">Have account? <a href="#" onclick="showPage('login')">Login</a></p>
        </div>`;
}

async function doRegister() {
    try {
        const body = {
            email: document.getElementById('email').value,
            password: document.getElementById('password').value,
            first_name: document.getElementById('first_name').value,
            last_name: document.getElementById('last_name').value
        };
        const data = await api('POST', '/auth/register', body);
        token = data.token;
        userId = data.user_id;
        localStorage.setItem('token', token);
        localStorage.setItem('user_id', userId);
        showPage('catalog');
    } catch (e) {
        document.getElementById('error').textContent = e.message;
    }
}

// === Catalog ===
async function catalog() {
    document.getElementById('app').innerHTML = '<div class="card"><h2>Catalog</h2><div id="products">Loading...</div></div>';
    try {
        const data = await api('GET', '/products');
        const products = data.products || [];
        let html = '';
        products.forEach(p => {
            html += `<div class="card">
                <strong>${p.name}</strong> — $${p.price.toFixed(2)} | Stock: ${p.stock}
                <p style="color:#666;font-size:13px">${p.description || ''}</p>
                <button class="btn btn-primary btn-sm" onclick="buyProduct('${p.id}')">Buy</button>
            </div>`;
        });
        document.getElementById('products').innerHTML = html || '<p>No products yet.</p>';
    } catch (e) {
        document.getElementById('products').innerHTML = '<p class="error">Failed to load products</p>';
    }
}

async function buyProduct(productId) {
    const qty = prompt('Quantity:');
    if (!qty) return;
    try {
        await api('POST', '/orders', {
            items: [{ product_id: productId, quantity: parseInt(qty) }]
        });
        alert('Order created!');
        showPage('orders');
    } catch (e) {
        alert('Error: ' + e.message);
    }
}

// === Orders ===
async function orders() {
    document.getElementById('app').innerHTML = '<div class="card"><h2>My Orders</h2><div id="orders">Loading...</div></div>';
    try {
        const data = await api('GET', '/orders/user/' + userId);
        // Проверь структуру ответа
        console.log('Orders response:', data);
        const orders = data.orders || [];
        let html = '';
        orders.forEach(o => {
            const badgeClass = 'badge-' + o.status;
            html += `<div class="card">
                <strong>Order #${o.id.substring(0, 8)}</strong>
                <span class="badge ${badgeClass}">${o.status}</span>
                <span style="float:right">$${Number(o.total).toFixed(2)}</span>
                ${o.status === 'confirmed' ? `<button class="btn btn-danger btn-sm" style="margin-top:10px" onclick="cancelOrder('${o.id}')">Cancel</button>` : ''}
            </div>`;
        });
        document.getElementById('orders').innerHTML = html || '<p>No orders yet.</p>';
    } catch (e) {
        document.getElementById('orders').innerHTML = '<p class="error">Failed to load orders: ' + e.message + '</p>';
    }
}

async function cancelOrder(orderId) {
    if (!confirm('Cancel this order?')) return;
    try {
        await api('PUT', '/orders/' + orderId + '/cancel');
        alert('Order cancelled');
        showPage('orders');
    } catch (e) {
        alert('Error: ' + e.message);
    }
}

// === Profile ===
async function profile() {
    document.getElementById('app').innerHTML = '<div class="card"><h2>Profile</h2><div id="profile">Loading...</div></div>';
    try {
        const [profile, balance] = await Promise.all([
            api('GET', '/users/profile'),
            api('GET', '/users/' + userId + '/balance')
        ]);
        document.getElementById('profile').innerHTML = `
            <p><strong>Name:</strong> ${profile.first_name} ${profile.last_name}</p>
            <p><strong>Email:</strong> ${profile.email}</p>
            <p><strong>Role:</strong> ${profile.role}</p>
            <div class="balance" style="margin-top:15px">Balance: $${balance.balance.toFixed(2)}</div>
        `;
    } catch (e) {
        document.getElementById('profile').innerHTML = '<p class="error">Failed to load profile</p>';
    }
}

// === Start ===
showPage(token ? 'catalog' : 'login');