<template>
  <div class="layout">
<header class="header">
  <h1>GeoTracker</h1>
  <div class="header-right">
    <button class="burger" @click="toggleSidebar">☰</button>
    <span>{{ username }}</span>
    <button @click="logout">Выйти</button>
  </div>
</header>


    <div class="main">
<aside class="sidebar" :class="{ open: sidebarOpen }">
  <div class="sidebar-header">
    <h2>Активные пользователи</h2>
    <button class="close-btn" @click="sidebarOpen = false">✕</button>
  </div>
  <ul>
    <li 
      v-for="(pos, user) in markers"
      :key="user"
      @click="focusOn(user)"
      :class="{ active: user === focusedUser }"
    >
      {{ user }}
    </li>
  </ul>
</aside>


      <div id="map" class="map"></div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet.marker.slideto'

const map = ref(null)
const markers = ref({})
const focusedUser = ref(null)
const username = localStorage.getItem('username') || 'Гость'
const router = useRouter()
const sidebarOpen = ref(false)
let intervalId = null

const toggleSidebar = () => {
  sidebarOpen.value = !sidebarOpen.value
}


const defaultIcon = L.icon({
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
})

const greenIcon = L.icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-green.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
})

const focusOn = (userId) => {
  if (markers.value[userId]) {
    focusedUser.value = userId
    const latlng = markers.value[userId].getLatLng()
    map.value.setView(latlng, 14)
    markers.value[userId].openPopup()


    if (window.innerWidth <= 768) {
      sidebarOpen.value = false
    }
  }
}


const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  router.push('/login')
}

const sendLocation = (lat, lon) => {
  const token = localStorage.getItem('token')
  if (!token) return

  fetch('/api/location', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      user_id: username,
      lat,
      lon,
    }),
  }).catch(err => console.error('Location send error:', err))
}

const startLocationUpdates = () => {
  if (!navigator.geolocation) {
    console.error('Geolocation is not supported')
    return
  }

  navigator.geolocation.getCurrentPosition(
    (pos) => {
      const { latitude, longitude } = pos.coords
      sendLocation(latitude, longitude)
    },
    (err) => {
      console.error('Geolocation error:', err)
    }
  )

  intervalId = setInterval(() => {
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const { latitude, longitude } = pos.coords
        sendLocation(latitude, longitude)
      },
      (err) => {
        console.error('Geolocation error:', err)
      }
    )
  }, 10000)
}

onUnmounted(() => {
  clearInterval(intervalId)
})

onMounted(() => {
  map.value = L.map('map').setView([55.75, 37.61], 10)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 18,
    attribution: '&copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors',
  }).addTo(map.value)

  const ws = new WebSocket('wss://' + window.location.host + '/ws')

  ws.onopen = () => {
    console.log('WebSocket connected')
  }

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    const { user_id, lat, lon } = data

    if (!markers.value[user_id]) {
      const icon = user_id === username ? greenIcon : defaultIcon
      markers.value[user_id] = L.marker([lat, lon], { icon }).addTo(map.value).bindPopup(user_id)
    } else {
      markers.value[user_id].slideTo([lat, lon], {
        duration: 1000,
        keepAtCenter: false,
      })
    }

    const group = L.featureGroup(Object.values(markers.value))
    map.value.fitBounds(group.getBounds().pad(0.5))
  }

  startLocationUpdates()

  ws.onclose = () => {
    console.log('WebSocket disconnected')
  }
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
  width: 280px;
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

/* Responsive styles */
@media (max-width: 768px) {
  .burger {
    display: block;
  }

  .sidebar {
    position: absolute;
    top: 56px;
    left: 0;
    width: 100%;
    height: 200px;
    background: #34495e;
    z-index: 10;
    transform: translateY(-130%);
    transition: transform 0.3s ease;
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

  .sidebar {
    padding: 0.75rem;
  }
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



</style>
