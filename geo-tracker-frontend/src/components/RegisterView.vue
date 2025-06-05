<template>
  <div class="register-container">
    <form class="register-form" @submit.prevent="register">
      <h2>Регистрация</h2>
      <input v-model="username" placeholder="Имя пользователя" required />
      <input v-model="password" type="password" placeholder="Пароль" required />
      <button type="submit">Зарегистрироваться</button>
      <p v-if="message" class="message">{{ message }}</p>
      <p class="login-link">
        Уже есть аккаунт?
        <router-link to="/login">Войти</router-link>
      </p>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const username = ref('')
const password = ref('')
const message = ref('')
const router = useRouter()

const register = async () => {
  try {
    const res = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
    })

    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Registration failed')

    message.value = 'Успешная регистрация! Переход к входу...'
    setTimeout(() => router.push('/login'), 1500)
  } catch (err) {
    message.value = err.message
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #f5f6fa;
}

.register-form {
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

.register-form h2 {
  margin: 0;
  text-align: center;
  color: #333;
}

.register-form input {
  padding: 0.75rem 1rem;
  font-size: 1rem;
  border: 1px solid #ccc;
  border-radius: 8px;
  outline: none;
  transition: border 0.2s;
}

.register-form input:focus {
  border-color: #646cff;
}

.register-form button {
  padding: 0.75rem 1rem;
  font-size: 1rem;
  background-color: #646cff;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  width: 100%;
}

.register-form button:hover {
  background-color: #535bf2;
}

.message {
  color: #2ecc71;
  text-align: center;
  font-size: 0.9rem;
}

.login-link {
  text-align: center;
  font-size: 0.9rem;
  color: #333;
}

.login-link a {
  color: #646cff;
  text-decoration: none;
  font-weight: 500;
}

.login-link a:hover {
  text-decoration: underline;
}
</style>
