using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using TYBLOG.Data;
using TYBLOG.Models;
using CsvHelper;
using System.Globalization;
using System.IO;
using Microsoft.Extensions.Configuration;

namespace TYBLOG.Controllers;

[ApiController]
[Route("api/[controller]")]
public class ApplicationController : ControllerBase
{
    private readonly ApplicationDbContext _context;
    private readonly ILogger<ApplicationController> _logger;
    private readonly IConfiguration _configuration;

    public class AdminAuthRequest
    {
        public string Password { get; set; } = string.Empty;
    }

    public ApplicationController(ApplicationDbContext context, ILogger<ApplicationController> logger, IConfiguration configuration)
    {
        _context = context;
        _logger = logger;
        _configuration = configuration;
    }

    private bool IsValidAdminPassword(string password)
    {
        var correctPassword = _configuration.GetValue<string>("AdminSettings:Password");
        return !string.IsNullOrEmpty(correctPassword) && password == correctPassword;
    }

    // GET: api/application
    [HttpPost("list")]
    public async Task<ActionResult<IEnumerable<Application>>> GetApplications([FromBody] AdminAuthRequest request)
    {
        if (!IsValidAdminPassword(request.Password))
        {
            return Unauthorized(new { Message = "Invalid admin password" });
        }

        return await _context.Applications.ToListAsync();
    }

    // POST: api/application
    [HttpPost]
    public async Task<ActionResult<Application>> PostApplication(Application application)
    {
        try
        {
            _context.Applications.Add(application);
            await _context.SaveChangesAsync();

            return CreatedAtAction(nameof(GetApplication), new { id = application.Id }, application);
        }
        catch (DbUpdateException ex)
        {
            if (ex.InnerException?.Message.Contains("unique constraint") == true)
            {
                return Conflict("A player with this Minecraft username already applied.");
            }
            
            _logger.LogError(ex, "Error submitting application");
            return StatusCode(500, "An error occurred while submitting your application.");
        }
    }

    // GET: api/application/{id}
    [HttpGet("{id}")]
    public async Task<ActionResult<Application>> GetApplication(int id)
    {
        var application = await _context.Applications.FindAsync(id);

        if (application == null)
        {
            return NotFound();
        }

        return application;
    }

    // GET: api/application/export
    [HttpPost("export")]
    public async Task<IActionResult> ExportToCsv([FromBody] AdminAuthRequest request)
    {
        if (!IsValidAdminPassword(request.Password))
        {
            return Unauthorized(new { Message = "Invalid admin password" });
        }

        var applications = await _context.Applications.ToListAsync();
        
        using var memoryStream = new MemoryStream();
        using (var writer = new StreamWriter(memoryStream))
        using (var csv = new CsvWriter(writer, CultureInfo.InvariantCulture))
        {
            await csv.WriteRecordsAsync(applications);
        }

        return File(memoryStream.ToArray(), "text/csv", "applications.csv");
    }

    // DELETE: api/application/delete
    [HttpPost("delete")]
    public async Task<IActionResult> DeleteApplication([FromBody] DeleteApplicationRequest request)
    {
        if (!IsValidAdminPassword(request.Password))
        {
            return Unauthorized(new { Message = "Invalid admin password" });
        }

        var application = await _context.Applications.FindAsync(request.Id);
        if (application == null)
        {
            return NotFound(new { Message = "Application not found" });
        }

        try
        {
            _context.Applications.Remove(application);
            await _context.SaveChangesAsync();
            
            return Ok(new { Success = true, Message = "Application deleted successfully" });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error deleting application");
            return StatusCode(500, new { Message = "An error occurred while deleting the application" });
        }
    }
    
    public class DeleteApplicationRequest : AdminAuthRequest
    {
        public int Id { get; set; }
    }

    public class ReviewApplicationRequest : AdminAuthRequest
    {
        public int Id { get; set; }
        public string? Notes { get; set; }
        public string? AcceptanceStatus { get; set; }
    }

    // POST: api/application/review
    [HttpPost("review")]
    public async Task<IActionResult> ReviewApplication([FromBody] ReviewApplicationRequest request)
    {
        if (!IsValidAdminPassword(request.Password))
        {
            return Unauthorized(new { Message = "Invalid admin password" });
        }

        var application = await _context.Applications.FindAsync(request.Id);
        if (application == null)
        {
            return NotFound(new { Message = "Application not found" });
        }

        try
        {
            application.IsReviewed = true;
            application.ReviewedAt = DateTime.UtcNow;
            application.ReviewNotes = request.Notes;
            
            // Set acceptance status (defaults to "pending" if not specified)
            application.AcceptanceStatus = string.IsNullOrEmpty(request.AcceptanceStatus) 
                ? "pending" 
                : request.AcceptanceStatus;

            await _context.SaveChangesAsync();
            
            return Ok(new { Success = true, Message = "Application reviewed successfully", Application = application });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error reviewing application");
            return StatusCode(500, new { Message = "An error occurred while reviewing the application" });
        }
    }

    // POST: api/application/unreview
    [HttpPost("unreview")]
    public async Task<IActionResult> UnreviewApplication([FromBody] ReviewApplicationRequest request)
    {
        if (!IsValidAdminPassword(request.Password))
        {
            return Unauthorized(new { Message = "Invalid admin password" });
        }

        var application = await _context.Applications.FindAsync(request.Id);
        if (application == null)
        {
            return NotFound(new { Message = "Application not found" });
        }

        try
        {
            application.IsReviewed = false;
            application.ReviewedAt = null;
            application.ReviewNotes = null;
            application.AcceptanceStatus = "pending"; // Reset to pending when unreviewed

            await _context.SaveChangesAsync();
            
            return Ok(new { Success = true, Message = "Application review status removed successfully", Application = application });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error removing review status");
            return StatusCode(500, new { Message = "An error occurred while removing the review status" });
        }
    }
} 