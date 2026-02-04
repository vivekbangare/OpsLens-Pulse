const loginBtn = document.getElementById("loginBtn")
const usernameInput = document.getElementById("username")
const passwordInput = document.getElementById("password")
const errorBox = document.getElementById("error")

loginBtn.addEventListener("click", () => {
  const username = usernameInput.value.trim()
  const password = passwordInput.value.trim()

  if (username === "admin" && password === "admin@2026") {
    sessionStorage.setItem("loggedIn", "true")
    window.location.href = "/index.html"
  } else {
    errorBox.textContent = "Invalid username or password"
  }
})

// auto-redirect if already logged in
if (sessionStorage.getItem("loggedIn")) {
  window.location.href = "/index.html"
}
