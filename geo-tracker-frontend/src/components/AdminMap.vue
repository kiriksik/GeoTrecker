<template>
  <div class="layout">
    <header class="header">
      <h1>GeoTracker — Админ</h1>
      <div class="header-right">
        <button class="burger" @click="toggleSidebar">☰</button>
        <span>{{ username }}</span>
        <button @click="logout">Выйти</button>
      </div>
    </header>

    <div class="main">
      <aside class="sidebar" :class="{ open: sidebarOpen }">
        <div class="sidebar-header">
          <h2>Пользователи</h2>
          <button class="close-btn" @click="sidebarOpen = false">✕</button>
          <button v-if="focusedUser" @click="clearFocus" class="clear-button">
            Сбросить фокус
          </button>
        </div>

        <ul>
        <li
            v-for="user in users"
            :key="user.username"
            @click="focusOn(user.username)"
            :class="{ active: user.username === focusedUser }"
        >
            <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
                <strong>{{ user.username }}</strong>
                <span style="font-size: 0.85em; color: #bdc3c7;">({{ user.role }})</span>
            </div>
            <span
                v-if="user.username !== username"
                @click.stop="deleteUser(user.username)"
                style="color: #e74c3c; cursor: pointer;"
                title="Удалить"
            >
                🗑️
            </span>
            </div>
        </li>
        </ul>


      </aside>

      <div id="map" class="map"></div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet.marker.slideto'

const map = ref(null)
const markers = ref({})
const focusedUser = ref(null)
const username = localStorage.getItem('username') || 'Админ'
const router = useRouter()
const sidebarOpen = ref(false)
const users = ref([])
const routeLine = ref(null)

const token = localStorage.getItem('token')
const isAdmin = localStorage.getItem('role') === 'admin'

const toggleSidebar = () => {
  sidebarOpen.value = !sidebarOpen.value
}

const fetchUsers = async () => {
  try {
    const res = await fetch('/api/admin/users', {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }
    })
    if (!res.ok) throw new Error('Ошибка загрузки пользователей')
    users.value = await res.json()
  } catch (err) {
    console.error(err)
  }
}

const deleteUser = async (usernameToDelete) => {
  if (!confirm(`Удалить пользователя ${usernameToDelete}?`)) return
  try {
    const res = await fetch(`/api/admin/users/${usernameToDelete}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!res.ok) throw new Error('Ошибка удаления')
    users.value = users.value.filter(u => u.username !== usernameToDelete)
    if (focusedUser.value === usernameToDelete) clearFocus()
  } catch (err) {
    console.error(err)
  }
}

const focusOn = async (userId) => {
  focusedUser.value = userId
  const marker = markers.value[userId]
  if (marker) {
    map.value.setView(marker.getLatLng(), 14)
    marker.openPopup()
  }

  if (routeLine.value) {
    map.value.removeLayer(routeLine.value)
    routeLine.value = null
  }

  try {
    const res = await fetch(`/api/admin/users/${userId}/history`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    const data = await res.json()
    const points = data.map(p => [p.lat, p.lon])
    if (points.length > 1) {
      routeLine.value = L.polyline(points, {
        color: 'blue',
        weight: 4,
        opacity: 0.7,
      }).addTo(map.value)
      map.value.fitBounds(routeLine.value.getBounds())
    }
  } catch (err) {
    console.error(err)
  }
}

const clearFocus = () => {
  focusedUser.value = null
  if (routeLine.value) {
    map.value.removeLayer(routeLine.value)
    routeLine.value = null
  }
  map.value.setView([59.9343, 30.3351], 12)
}

const logout = () => {
  localStorage.clear()
  router.push('/login')
}

onMounted(() => {
  if (!isAdmin) {
    router.push('/')
    return
  }

  fetchUsers()

  map.value = L.map('map').setView([59.9343, 30.3351], 12)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 18,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map.value)

  const ws = new WebSocket('wss://' + window.location.host + '/ws')

  ws.onmessage = (event) => {
    const { user_id, lat, lon } = JSON.parse(event.data)
    const initials = user_id.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 3)

    if (!markers.value[user_id]) {
      const iconHtml = `
        <div style="display: flex; flex-direction: column; align-items: center; font-weight: bold; color: #2980b9; font-size: 0.75rem;">
          <div style="background: white; padding: 2px 6px; border-radius: 4px; margin-bottom: 2px; box-shadow: 0 0 2px rgba(0,0,0,0.3);">${initials}</div>
          <img src="/icons/red_icon.png"/>
        </div>
      `
      const icon = L.divIcon({ html: iconHtml, className: '', iconSize: [25, 50], iconAnchor: [12, 70] })
      markers.value[user_id] = L.marker([lat, lon], { icon }).addTo(map.value)
    } else {
      markers.value[user_id].slideTo([lat, lon], { duration: 1000, keepAtCenter: false })
    }
  }
})

onUnmounted(() => {

})
</script>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}

.header {
  background-color: #2c3e50;
  color: white;
  padding: 0.5rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 64px;
  min-height: 64px;
  box-shadow: 0 2px 5px rgba(0,0,0,0.3);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header > div {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.95rem;
}

.main {
  display: flex;
  flex: 1;
  min-height: 0;
}

.sidebar {
  width: 320px;
  background-color: #34495e;
  border-right: 2px solid #22313f;
  padding: 1rem;
  overflow-y: auto;
  color: white;
  box-sizing: border-box;
  transition: transform 0.3s ease-in-out;
}

.sidebar h2 {
  margin-top: 0;
  margin-bottom: 1rem;
  font-size: 1.2rem;
  border-bottom: 2px solid #1abc9c;
  padding-bottom: 0.4rem;
}

.sidebar ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.sidebar li {
  padding: 0.5rem 0.75rem;
  margin-bottom: 0.5rem;
  border-radius: 5px;
  background-color: transparent;
  transition: background 0.2s, color 0.2s;
  cursor: pointer;
}

.sidebar li:hover {
  background-color: #1abc9c;
  color: #2c3e50;
}

.sidebar li.active {
  background-color: #16a085;
  font-weight: bold;
  color: white;
  box-shadow: 0 0 6px rgba(0, 0, 0, 0.2);
}

.map {
  flex: 1;
  height: 100%;
  box-shadow: inset 0 0 15px rgba(0,0,0,0.15);
  transition: all 0.3s ease;
  border-radius: 0;
}

button {
  background-color: #e74c3c;
  color: white;
  border: none;
  padding: 0.4em 0.8em;
  border-radius: 20px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9rem;
  transition: background-color 0.3s ease;
}

button:hover {
  background-color: #c0392b;
}

.burger {
  display: none;
  font-size: 1.5rem;
  background: none;
  border: none;
  color: white;
  cursor: pointer;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.close-btn {
  display: none;
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
}

/* Responsive styles */
@media (max-width: 768px) {
  .header {
    flex-direction: column;
    height: auto;
    padding: 0.5rem;
    gap: 0.3rem;
  }

  .burger {
    display: block;
  }
.sidebar.open {
    position:relative;
}
  .sidebar {
    position:absolute;
    top: auto;
    left: 0;
    width: 100%;
    height: auto;
    background: #34495e;
    z-index: 401;
    transform: translateY(-500%);
    transition: transform 0.4s ease;
  }

  .sidebar.open {
    transform: translateY(0);
  }

  .map {
    flex: 1;
    height: auto;
  }

    .close-btn {
        display: block;
    }
}


@media (max-width: 480px) {
  .header {
    flex-direction: column;
    height: auto;
    padding: 0.5rem;
    gap: 0.3rem;
  }

  .header h1 {
    font-size: 1.2rem;
  }

  .header > div {
    flex-direction: column;
    align-items: flex-start;
    font-size: 0.85rem;
  }

  button {
    width: 100%;
    margin-top: 0.5rem;
  }
.sidebar.open {
    position:relative;
}
  .sidebar {
    
    padding: 0.75rem;
  }
}

.clear-button {
  margin-top: 1rem;
  margin-left: 1rem;
  margin-right: 1rem;
  padding-bottom: 0.4rem;
  padding: 0.5rem 1rem;
  background: #e74c3c;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  width: 100%;
}

.clear-button:hover {
  background: #c0392b;
}


</style>
