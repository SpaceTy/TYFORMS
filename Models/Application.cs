using System.ComponentModel.DataAnnotations;

namespace TYBLOG.Models;

public class Application
{
    [Key]
    public int Id { get; set; }
    
    [Required]
    public string DiscordUsername { get; set; } = string.Empty;
    
    [Required]
    public string MinecraftUsername { get; set; } = string.Empty;
    
    [Required]
    public int Age { get; set; }
    
    [Required]
    public string FavoriteAboutMinecraft { get; set; } = string.Empty;
    
    [Required]
    public string UnderstandingOfSMP { get; set; } = string.Empty;
    
    [Required]
    public bool JoinedDiscord { get; set; }
    
    public DateTime SubmissionDate { get; set; } = DateTime.UtcNow;
} 