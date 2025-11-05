<template>
  <div id="app">
    <nav>
      <router-link to="/broadcaster">主播端</router-link>
      <router-link to="/viewer">觀眾端</router-link>
    </nav>
    <router-view />
  </div>
</template>

<style scoped>
nav {
  padding: 20px;
  background: #333;
  display: flex;
  gap: 20px;
}

nav a {
  color: white;
  text-decoration: none;
  padding: 10px 20px;
  border-radius: 4px;
  background: #555;
}

nav a.router-link-active {
  background: #007bff;
}
</style>

