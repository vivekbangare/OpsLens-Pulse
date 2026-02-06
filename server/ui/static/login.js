const loginBtn = document.getElementById("loginBtn")
const apiKeyInput = document.getElementById("apiKey")
const errorBox = document.getElementById("error")

loginBtn.addEventListener("click", () => {
  const apiKey = apiKeyInput.value.trim()

  if (!apiKey) {
    errorBox.textContent = "API key is required"
    return
  }

  //save API key in session storage and redirect to hosts page
  sessionStorage.setItem("apiKey", apiKey)
  sessionStorage.setItem("loggedIn", "true")
  window.location.href = "/hosts.html"

})

// auto-redirect if already logged in
if (sessionStorage.getItem("apiKey") && sessionStorage.getItem("loggedIn") === "true") {
  window.location.href = "/hosts.html"
}
