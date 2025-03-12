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
} 