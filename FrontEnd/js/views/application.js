import api from "../api.js";
import { navigate } from "../router.js";


const DISCORD_URL = "https://discord.com/invite/gqQFuEK3hk";

export default function render(root) {
  root.innerHTML = `
    <div class="application-view fade-in">
      <div class="mc-container">
        <h2 class="mc-title text-center">TYSMP Application Form</h2>
        <p class="application-intro">
          Join the <a href="${DISCORD_URL}" target="_blank" rel="noopener" data-discord>discord</a>
        </p>

        <form class="application-form" id="application-form" novalidate>
          <div class="form-group">
            <label class="mc-label" for="discordUsername">What is your discord username? <span style="color:var(--red-500)">*</span></label>
            <input class="mc-input" id="discordUsername" name="discordUsername" type="text" required autocomplete="off">
          </div>

          <div class="form-group">
            <label class="mc-label" for="minecraftUsername">What is your minecraft username? <span style="color:var(--red-500)">*</span></label>
            <input class="mc-input" id="minecraftUsername" name="minecraftUsername" type="text" required autocomplete="off">
          </div>

          <div class="form-group">
            <span class="mc-label">How old are you? <span style="color:var(--red-500)">*</span></span>
            <div class="age-row">
              <div class="age-slider-wrap">
                <div class="age-range-labels"><span>13</span><span>99</span></div>
                <div class="age-slider-holder">
                  <input class="age-slider" id="ageSlider" type="range" min="13" max="99" step="1" value="16">
                  <div class="age-tooltip" id="ageTooltip">16</div>
                </div>
              </div>
              <div class="age-number-wrap">
                <input class="mc-input age-number" id="ageNumber" type="number" min="13" max="99" value="16">
              </div>
            </div>
            <div class="form-hint">Drag the slider or type your age (13-99)</div>
          </div>

          <div class="form-group">
            <label class="mc-label" for="favoriteAboutMinecraft">What do you like most about Minecraft? <span style="color:var(--red-500)">*</span></label>
            <input class="mc-input" id="favoriteAboutMinecraft" name="favoriteAboutMinecraft" type="text" required autocomplete="off">
          </div>

          <div class="form-group">
            <label class="mc-label" for="understandingOfSMP">How do you understand the concept of a private vanilla SMP? <span style="color:var(--red-500)">*</span></label>
            <input class="mc-input" id="understandingOfSMP" name="understandingOfSMP" type="text" required autocomplete="off">
          </div>

          <div class="form-group">
            <span class="mc-label">Have you joined the discord server? (REQUIRED) link in description of form. <span style="color:var(--red-500)">*</span></span>
            <div class="radio-list">
              <label class="radio-option"><input class="mc-radio" type="radio" name="joinedDiscord" value="yes" required> yes</label>
              <label class="radio-option"><input class="mc-radio" type="radio" name="joinedDiscord" value="no" required> no</label>
            </div>
          </div>

          <div class="form-error hidden" id="formError"></div>

          <div class="text-center">
            <button type="submit" class="mc-button" id="submitButton">Submit Application</button>
          </div>
        </form>
      </div>
    </div>
  `;

  const form = root.querySelector("#application-form");
  const slider = root.querySelector("#ageSlider");
  const numberInput = root.querySelector("#ageNumber");
  const tooltip = root.querySelector("#ageTooltip");
  const errorBox = root.querySelector("#formError");
  const submitButton = root.querySelector("#submitButton");

  let tooltipTimer = null;
  let submitting = false;

  function updateTooltip(show) {
    const value = Number(slider.value);
    tooltip.textContent = String(value);
    const percent = (value - 13) / 86;
    const position = percent * Math.max(0, slider.offsetWidth - 20) + 10;
    tooltip.style.left = `${position}px`;
    tooltip.classList.toggle("visible", show);
  }

  function showTooltip() {
    if (tooltipTimer) clearTimeout(tooltipTimer);
    updateTooltip(true);
    tooltipTimer = setTimeout(() => {
      tooltip.classList.remove("visible");
    }, 3000);
  }

  function syncAge(source) {
    if (source === "slider") {
      numberInput.value = slider.value;
    } else {
      let value = Number(numberInput.value);
      if (Number.isNaN(value)) value = 13;
      value = Math.min(99, Math.max(13, value));
      numberInput.value = String(value);
      slider.value = String(value);
    }
    showTooltip();
  }

  slider.addEventListener("input", () => syncAge("slider"));
  slider.addEventListener("mousedown", showTooltip);
  slider.addEventListener("touchstart", showTooltip, { passive: true });
  numberInput.addEventListener("input", () => syncAge("number"));
  numberInput.addEventListener("focus", showTooltip);

  const onResize = () => updateTooltip(tooltip.classList.contains("visible"));
  window.addEventListener("resize", onResize);
  requestAnimationFrame(() => updateTooltip(false));
  showTooltip();

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (submitting) return;

    if (!form.reportValidity()) return;

    const joined = form.querySelector('input[name="joinedDiscord"]:checked');
    if (!joined) {
      errorBox.textContent = "Please answer whether you have joined the discord server.";
      errorBox.classList.remove("hidden");
      return;
    }

    const payload = {
      discordUsername: form.querySelector("#discordUsername").value.trim(),
      minecraftUsername: form.querySelector("#minecraftUsername").value.trim(),
      age: Number(slider.value),
      favoriteAboutMinecraft: form.querySelector("#favoriteAboutMinecraft").value.trim(),
      understandingOfSMP: form.querySelector("#understandingOfSMP").value.trim(),
      joinedDiscord: joined.value === "yes",
    };

    submitting = true;
    errorBox.classList.add("hidden");
    submitButton.disabled = true;
    submitButton.textContent = "Submitting...";

    try {
      await api.submitApplication(payload);
      navigate("/thank-you");
    } catch (error) {
      submitting = false;
      submitButton.disabled = false;
      submitButton.textContent = "Submit Application";
      if (error.status === 409) {
        errorBox.textContent =
          (error.data && error.data.error) || "A player with this Minecraft username already applied.";
      } else {
        errorBox.textContent = "There was an error submitting your application. Please try again.";
      }
      errorBox.classList.remove("hidden");
    }
  });

  return () => {
    if (tooltipTimer) clearTimeout(tooltipTimer);
    window.removeEventListener("resize", onResize);
  };
}
