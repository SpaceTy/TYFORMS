using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;

namespace TYBLOG.Controllers
{
    [ApiController]
    [Route("api/[controller]")]
    public class AuthController : ControllerBase
    {
        private readonly IConfiguration _configuration;

        public class LoginRequest
        {
            public string Password { get; set; } = string.Empty;
        }

        public class LoginResponse
        {
            public bool Success { get; set; }
            public string Message { get; set; } = string.Empty;
        }

        public AuthController(IConfiguration configuration)
        {
            _configuration = configuration;
        }

        [HttpPost("verify")]
        public ActionResult<LoginResponse> Verify([FromBody] LoginRequest request)
        {
            var correctPassword = _configuration.GetValue<string>("AdminSettings:Password");
            
            if (string.IsNullOrEmpty(correctPassword))
            {
                return BadRequest(new LoginResponse 
                { 
                    Success = false, 
                    Message = "Server configuration error" 
                });
            }

            if (request.Password == correctPassword)
            {
                return Ok(new LoginResponse
                {
                    Success = true,
                    Message = "Authentication successful"
                });
            }
            
            return Unauthorized(new LoginResponse
            {
                Success = false,
                Message = "Incorrect password"
            });
        }
    }
}