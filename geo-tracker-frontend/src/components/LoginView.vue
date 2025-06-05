<template>
  <div class="login-container">
    <form class="login-form" @submit.prevent="login">
      <h2>Вход</h2>
      <input v-model="username" placeholder="Имя пользователя" required />
      <input v-model="password" type="password" placeholder="Пароль" required />
      <button type="submit">Войти</button>
      <p v-if="error" class="error">{{ error }}</p>
     <p class="register-link">
  Нет аккаунта?
  <router-link to="/register">Зарегистрироваться</router-link>
</p>

    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const username = ref('')
const password = ref('')
const error = ref('')
const router = useRouter()

const login = async () => {
  try {
    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
    })

    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Login failed')

    localStorage.setItem('token', data.token)

    const resMe = await fetch('/api/me', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
      },
    })

    const dataMe = await resMe.json()
    if (!resMe.ok) throw new Error(dataMe.error || 'Login failed')

    localStorage.setItem('username', dataMe.username)
    router.push('/')
  } catch (err) {
    error.value = err.message
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #f5f6fa;
}

.login-form {
  background: white;
  padding: 2rem 2.5rem;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 360px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.login-form h2 {
  margin: 0;
  text-align: center;
  color: #333;
}

.login-form input {
  padding: 0.75rem 1rem;
  font-size: 1rem;
  border: 1px solid #ccc;
  border-radius: 8px;
  outline: none;
  transition: border 0.2s;
}

.login-form input:focus {
  border-color: #646cff;
}

.login-form button {
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  background-color: #646cff;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  box-sizing: border-box;
  display: block;
  margin: 0 auto;
}

.login-form button:hover {
  background-color: #535bf2;
}


.error {
  color: #e74c3c;
  text-align: center;
  font-size: 0.9rem;
}

.register-link {
  text-align: center;
  font-size: 0.9rem;
  color: #333;
}


.register-link a {
  color: #646cff;
  text-decoration: none;
  font-weight: 500;
}

.register-link a:hover {
  text-decoration: underline;
}
</style>
