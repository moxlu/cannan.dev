function showReset() {
  document.getElementById("form-login").style.display = "none";
  document.getElementById("form-reset").style.display = "block";
}

function showLogin() {
  document.getElementById("form-login").style.display = "block";
  document.getElementById("form-reset").style.display = "none";
}

function showSection(idToShow) {
  const sections = ["noticeboard", "challenges", "scoreboard"];
  sections.forEach((id) => {
    const section = document.getElementById(id);
    if (section) {
      section.style.display = id === idToShow ? "block" : "none";
    }
  });
}

function openModal() {
  document.getElementById("challenge-modal").style.display = "block";
}

function closeModal() {
  document.getElementById("challenge-modal").style.display = "none";
}

// Close modal on background click
window.onclick = function (event) {
  const modal = document.getElementById("challenge-modal");
  if (event.target === modal) {
    closeModal();
  }
};
