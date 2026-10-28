import { navigate } from "../router.js";

const DISCORD_URL = "https://discord.com/invite/gqQFuEK3hk";

export default function render(root) {
  root.innerHTML = `
    <div class="application-view fade-in">
      <div class="mc-container">
        <div class="thankyou-content">
          <h2 class="mc-title">Thank You for Applying!</h2>

          <div style="padding: 1.5rem 0; margin-bottom: 1.5rem;">
            <div class="thankyou-icon">✅</div>
            <p class="thankyou-lead">Your application has been submitted successfully!</p>
            <p class="thankyou-text">Our admins will review your application and contact you via Discord if you are accepted.</p>
            <p class="muted" style="font-size:0.75rem;">Make sure you've joined our Discord server!</p>
          </div>

          <div class="thankyou-actions">
            <button type="button" class="mc-button" id="joinDiscord">Join Discord</button>
            <button type="button" class="mc-button secondary" id="goHome">Back to Form</button>
          </div>
        </div>
      </div>
    </div>
  `;

  const joinButton = root.querySelector("#joinDiscord");
  const homeButton = root.querySelector("#goHome");

  const openDiscord = () => window.open(DISCORD_URL, "_blank");
  const goHome = () => navigate("/");

  joinButton.addEventListener("click", openDiscord);
  homeButton.addEventListener("click", goHome);

  return () => {
    joinButton.removeEventListener("click", openDiscord);
    homeButton.removeEventListener("click", goHome);
  };
}
