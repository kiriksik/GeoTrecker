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
    <button v-if="focusedUser" @click="clearFocus" class="clear-button">
        Сбросить фокус
    </button>

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



const routeLine = ref(null)

const focusOn = async (userId) => {
  if (markers.value[userId]) {
    focusedUser.value = userId


    const latlng = markers.value[userId].getLatLng()
    map.value.setView(latlng, 14)
    markers.value[userId].openPopup()

    if (routeLine.value) {
      map.value.removeLayer(routeLine.value)
      routeLine.value = null
    }

    try {
      const token = localStorage.getItem('token')
      const res = await fetch(`/api/history/${userId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
      if (!res.ok) throw new Error('Ошибка при загрузке истории')
      const data = await res.json()

      const points = data.map(p => [p.lat, p.lon])
      if (points.length > 1) {
        routeLine.value = L.polyline(points, {
          color: 'blue',
          weight: 4,
          opacity: 0.7,
          smoothFactor: 1
        }).addTo(map.value)

        map.value.fitBounds(routeLine.value.getBounds())
      }
    } catch (err) {
      console.error('Ошибка загрузки истории:', err)
    }
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
  }, 1000)
}

onUnmounted(() => {
  clearInterval(intervalId)
})

onMounted(() => {
  map.value = L.map('map').setView([59.9343, 30.3351], 12)

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
      const initials = user_id.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 3)

        const iconHtml = `
        <div style="
            position: relative;
            display: flex;
            flex-direction: column;
            align-items: center;
            font-weight: bold;
            color: ${user_id === username ? '#27ae60' : '#2980b9'};
            font-size: 0.75rem;
        ">
            <div style="
            background: white;
            padding: 2px 6px;
            border-radius: 4px;
            margin-bottom: 2px;
            box-shadow: 0 0 2px rgba(0,0,0,0.3);
            ">${initials}</div>
            <img src="${user_id === username ? '/icons/green_icon.png' : '/icons/red_icon.png'}"/>
        </div>
        `

        const icon = L.divIcon({
        html: iconHtml,
        className: '',
        iconSize: [25, 50],
        iconAnchor: [12, 70],
        popupAnchor: [0, -45],
        })

        markers.value[user_id] = L.marker([lat, lon], { icon }).addTo(map.value)

      if (user_id === username)
        focusOn(user_id)
    } else {
      markers.value[user_id].slideTo([lat, lon], {
        duration: 1000,
        keepAtCenter: false,
      })
    }

    
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
